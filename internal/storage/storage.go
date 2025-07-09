package storage

import (
	"sync"
	"time"
)

// DataType représente le type de données stocké
type DataType int

const (
	TypeString DataType = iota
	TypeList
	TypeSet
	TypeHash
	TypeZSet
)

// Value représente une valeur stockée avec son type et TTL
type Value struct {
	Data      interface{}
	Type      DataType
	ExpiresAt *time.Time
}

// RedisList représente une liste Redis
type RedisList struct {
	elements []string
}

// RedisSet représente un set Redis
type RedisSet struct {
	elements map[string]bool
}

// RedisHash représente un hash Redis
type RedisHash struct {
	fields map[string]string
}

// Storage est le stockage principal en mémoire avec gestion de la concurrence
type Storage struct {
	data  map[string]*Value
	mutex sync.RWMutex
}

// New crée une nouvelle instance de stockage
func New() *Storage {
	return &Storage{
		data: make(map[string]*Value),
	}
}

// Set stocke une valeur avec type et TTL optionnel
func (s *Storage) Set(key string, data interface{}, dataType DataType, ttl *time.Duration) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	var expiresAt *time.Time
	if ttl != nil {
		expiry := time.Now().Add(*ttl)
		expiresAt = &expiry
	}

	s.data[key] = &Value{
		Data:      data,
		Type:      dataType,
		ExpiresAt: expiresAt,
	}
}

// Get récupère une valeur, retourne nil si la clé n'existe pas ou a expiré
func (s *Storage) Get(key string) *Value {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	value, exists := s.data[key]
	if !exists {
		return nil
	}

	// Vérifier l'expiration
	if value.ExpiresAt != nil && time.Now().After(*value.ExpiresAt) {
		// Clé expirée - on la supprime de façon lazy
		delete(s.data, key)
		return nil
	}

	return value
}

// Delete supprime une clé et retourne true si elle existait
func (s *Storage) Delete(key string) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, exists := s.data[key]
	if exists {
		delete(s.data, key)
	}
	return exists
}

// Exists vérifie si une clé existe et n'a pas expiré
func (s *Storage) Exists(key string) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	value, exists := s.data[key]
	if !exists {
		return false
	}

	// Vérifier l'expiration
	if value.ExpiresAt != nil && time.Now().After(*value.ExpiresAt) {
		delete(s.data, key)
		return false
	}

	return true
}

// Keys retourne toutes les clés correspondant au pattern (style Redis glob)
func (s *Storage) Keys(pattern string) []string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var result []string
	now := time.Now()

	for key, value := range s.data {
		// Ignorer les clés expirées
		if value.ExpiresAt != nil && now.After(*value.ExpiresAt) {
			continue
		}

		// Vérifier si la clé correspond au pattern
		if matchGlobPattern(pattern, key) {
			result = append(result, key)
		}
	}

	return result
}

// Size retourne le nombre de clés valides (non expirées)
func (s *Storage) Size() int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	count := 0
	now := time.Now()

	for _, value := range s.data {
		if value.ExpiresAt == nil || now.Before(*value.ExpiresAt) {
			count++
		}
	}

	return count
}

// CleanupExpired supprime activement les clés expirées
func (s *Storage) CleanupExpired() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	now := time.Now()
	cleaned := 0

	for key, value := range s.data {
		if value.ExpiresAt != nil && now.After(*value.ExpiresAt) {
			delete(s.data, key)
			cleaned++
		}
	}

	return cleaned
}

// === MÉTHODES POUR LES LISTES ===

// ListPush ajoute des éléments à une liste (gauche ou droite)
func (s *Storage) ListPush(key string, elements []string, left bool) int {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	value, exists := s.data[key]
	var list *RedisList

	if !exists {
		// Créer une nouvelle liste
		list = &RedisList{elements: make([]string, 0)}
		s.data[key] = &Value{
			Data: list,
			Type: TypeList,
		}
	} else {
		// Vérifier que c'est bien une liste
		if value.Type != TypeList {
			return -1 // Erreur de type
		}
		list = value.Data.(*RedisList)
	}

	// Ajouter les éléments
	if left {
		// LPUSH - ajouter à gauche (début)
		newElements := make([]string, len(elements)+len(list.elements))
		copy(newElements, elements)
		copy(newElements[len(elements):], list.elements)
		list.elements = newElements
	} else {
		// RPUSH - ajouter à droite (fin)
		list.elements = append(list.elements, elements...)
	}

	return len(list.elements)
}

// ListPop supprime et retourne un élément de la liste
func (s *Storage) ListPop(key string, left bool) (string, bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	value, exists := s.data[key]
	if !exists {
		return "", false
	}

	if value.Type != TypeList {
		return "", false
	}

	list := value.Data.(*RedisList)
	if len(list.elements) == 0 {
		return "", false
	}

	var element string
	if left {
		// LPOP - supprimer à gauche
		element = list.elements[0]
		list.elements = list.elements[1:]
	} else {
		// RPOP - supprimer à droite
		element = list.elements[len(list.elements)-1]
		list.elements = list.elements[:len(list.elements)-1]
	}

	// Supprimer la clé si la liste est vide
	if len(list.elements) == 0 {
		delete(s.data, key)
	}

	return element, true
}

// ListLen retourne la longueur d'une liste
func (s *Storage) ListLen(key string) int {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	value, exists := s.data[key]
	if !exists {
		return 0
	}

	if value.Type != TypeList {
		return -1 // Erreur de type
	}

	list := value.Data.(*RedisList)
	return len(list.elements)
}

// ListRange retourne une partie de la liste
func (s *Storage) ListRange(key string, start, stop int) []string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	value, exists := s.data[key]
	if !exists {
		return []string{}
	}

	if value.Type != TypeList {
		return nil // Erreur de type
	}

	list := value.Data.(*RedisList)
	length := len(list.elements)

	if length == 0 {
		return []string{}
	}

	// Gérer les indices négatifs (comme Redis)
	if start < 0 {
		start = length + start
	}
	if stop < 0 {
		stop = length + stop
	}

	// Limiter aux bornes
	if start < 0 {
		start = 0
	}
	if stop >= length {
		stop = length - 1
	}
	if start > stop {
		return []string{}
	}

	return list.elements[start : stop+1]
}

// === MÉTHODES POUR LES SETS ===

// SetAdd ajoute des membres à un set
func (s *Storage) SetAdd(key string, members []string) int {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	value, exists := s.data[key]
	var set *RedisSet

	if !exists {
		set = &RedisSet{elements: make(map[string]bool)}
		s.data[key] = &Value{
			Data: set,
			Type: TypeSet,
		}
	} else {
		if value.Type != TypeSet {
			return -1
		}
		set = value.Data.(*RedisSet)
	}

	added := 0
	for _, member := range members {
		if !set.elements[member] {
			set.elements[member] = true
			added++
		}
	}

	return added
}

// SetMembers retourne tous les membres d'un set
func (s *Storage) SetMembers(key string) []string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	value, exists := s.data[key]
	if !exists {
		return []string{}
	}

	if value.Type != TypeSet {
		return nil
	}

	set := value.Data.(*RedisSet)
	members := make([]string, 0, len(set.elements))
	for member := range set.elements {
		members = append(members, member)
	}

	return members
}

// SetIsMember vérifie si un membre est dans un set
func (s *Storage) SetIsMember(key string, member string) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	value, exists := s.data[key]
	if !exists {
		return false
	}

	if value.Type != TypeSet {
		return false
	}

	set := value.Data.(*RedisSet)
	return set.elements[member]
}

// === MÉTHODES POUR LES HASHES ===

// HashSet définit un field dans un hash
func (s *Storage) HashSet(key string, field string, value string) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	val, exists := s.data[key]
	var hash *RedisHash

	if !exists {
		hash = &RedisHash{fields: make(map[string]string)}
		s.data[key] = &Value{
			Data: hash,
			Type: TypeHash,
		}
	} else {
		if val.Type != TypeHash {
			return false
		}
		hash = val.Data.(*RedisHash)
	}

	_, existed := hash.fields[field]
	hash.fields[field] = value
	return !existed // true si nouveau field
}

// HashGet récupère un field d'un hash
func (s *Storage) HashGet(key string, field string) (string, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	value, exists := s.data[key]
	if !exists {
		return "", false
	}

	if value.Type != TypeHash {
		return "", false
	}

	hash := value.Data.(*RedisHash)
	val, exists := hash.fields[field]
	return val, exists
}

// HashGetAll retourne tous les fields et valeurs d'un hash
func (s *Storage) HashGetAll(key string) map[string]string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	value, exists := s.data[key]
	if !exists {
		return map[string]string{}
	}

	if value.Type != TypeHash {
		return nil
	}

	hash := value.Data.(*RedisHash)
	result := make(map[string]string)
	for k, v := range hash.fields {
		result[k] = v
	}
	return result
}

// === PATTERN MATCHING ===

// matchGlobPattern implémente le pattern matching style Redis avec *, ?, et [...]
func matchGlobPattern(pattern, str string) bool {
	return matchGlob(pattern, str, 0, 0)
}

// matchGlob fonction récursive pour le pattern matching
func matchGlob(pattern, str string, patternIdx, strIdx int) bool {
	// Fin du pattern et de la string = match
	if patternIdx == len(pattern) && strIdx == len(str) {
		return true
	}

	// Fin du pattern mais pas de la string = pas de match
	if patternIdx == len(pattern) {
		return false
	}

	// Fin de la string mais pas du pattern = seulement OK si le reste du pattern est que des *
	if strIdx == len(str) {
		for i := patternIdx; i < len(pattern); i++ {
			if pattern[i] != '*' {
				return false
			}
		}
		return true
	}

	// Caractère actuel du pattern
	patternChar := pattern[patternIdx]

	switch patternChar {
	case '*':
		// * peut matcher 0 ou plusieurs caractères
		// Essayer de matcher 0 caractère (avancer dans le pattern seulement)
		if matchGlob(pattern, str, patternIdx+1, strIdx) {
			return true
		}
		// Essayer de matcher 1 caractère à la fois
		if strIdx < len(str) {
			return matchGlob(pattern, str, patternIdx, strIdx+1)
		}
		return false

	case '?':
		// ? doit matcher exactement 1 caractère
		if strIdx < len(str) {
			return matchGlob(pattern, str, patternIdx+1, strIdx+1)
		}
		return false

	case '[':
		// Classe de caractères [abc] ou [a-z] ou [^abc]
		if strIdx >= len(str) {
			return false
		}

		// Trouver la fin de la classe
		classEnd := patternIdx + 1
		for classEnd < len(pattern) && pattern[classEnd] != ']' {
			classEnd++
		}

		if classEnd >= len(pattern) {
			// Pas de ] fermant, traiter [ comme caractère littéral
			if pattern[patternIdx] == str[strIdx] {
				return matchGlob(pattern, str, patternIdx+1, strIdx+1)
			}
			return false
		}

		// Extraire la classe sans [ et ]
		class := pattern[patternIdx+1 : classEnd]
		matched := matchCharacterClass(class, str[strIdx])

		if matched {
			return matchGlob(pattern, str, classEnd+1, strIdx+1)
		}
		return false

	case '\\':
		// Caractère d'échappement
		if patternIdx+1 < len(pattern) && strIdx < len(str) {
			if pattern[patternIdx+1] == str[strIdx] {
				return matchGlob(pattern, str, patternIdx+2, strIdx+1)
			}
		}
		return false

	default:
		// Caractère littéral
		if strIdx < len(str) && patternChar == str[strIdx] {
			return matchGlob(pattern, str, patternIdx+1, strIdx+1)
		}
		return false
	}
}

// matchCharacterClass vérifie si un caractère correspond à une classe [abc], [a-z], [^abc]
func matchCharacterClass(class string, char byte) bool {
	if len(class) == 0 {
		return false
	}

	negated := false
	idx := 0

	// Vérifier si c'est une classe négative [^...]
	if class[0] == '^' {
		negated = true
		idx = 1
	}

	matched := false

	for idx < len(class) {
		if idx+2 < len(class) && class[idx+1] == '-' {
			// Intervalle de caractères [a-z]
			if char >= class[idx] && char <= class[idx+2] {
				matched = true
				break
			}
			idx += 3
		} else {
			// Caractère simple
			if class[idx] == char {
				matched = true
				break
			}
			idx++
		}
	}

	// Appliquer la négation si nécessaire
	if negated {
		return !matched
	}
	return matched
}

// === MÉTHODES UTILITAIRES ===

// FlushAll vide tout le stockage
func (s *Storage) FlushAll() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.data = make(map[string]*Value)
}

// Type retourne le type d'une clé
func (s *Storage) Type(key string) DataType {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	value, exists := s.data[key]
	if !exists {
		return -1 // Clé inexistante
	}

	// Vérifier l'expiration
	if value.ExpiresAt != nil && time.Now().After(*value.ExpiresAt) {
		delete(s.data, key)
		return -1
	}

	return value.Type
}
