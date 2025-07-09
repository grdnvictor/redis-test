package storage

import (
	"fmt"
	"sync"
	"time"
)

// Storage est le stockage principal en mémoire avec gestion de la concurrence
type Storage struct {
	data   map[string]string
	expiry map[string]time.Time
	mutex  sync.RWMutex
}

// NewStorage crée une nouvelle instance de stockage
func NewStorage() *Storage {
	return &Storage{
		data:   make(map[string]string),
		expiry: make(map[string]time.Time),
	}
}

// Set stocke une valeur string
func (s *Storage) Set(key, value string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.data[key] = value
	delete(s.expiry, key) // Supprime l'expiration si une nouvelle valeur est définie
}

// SetWithExpiry stocke une valeur string avec TTL
func (s *Storage) SetWithExpiry(key, value string, ttl time.Duration) {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.data[key] = value
	s.expiry[key] = time.Now().Add(ttl)
}

// Get récupère une valeur, retourne false si la clé n'existe pas ou a expiré
func (s *Storage) Get(key string) (string, bool) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Vérifie si la clé a expiré
	if expTime, exists := s.expiry[key]; exists && time.Now().After(expTime) {
		delete(s.data, key)
		delete(s.expiry, key)
		return "", false
	}

	value, exists := s.data[key]
	return value, exists
}

// Delete supprime une clé
func (s *Storage) Delete(key string) bool {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	_, exists := s.data[key]
	if exists {
		delete(s.data, key)
		delete(s.expiry, key)
	}
	return exists
}

// Exists vérifie si une clé existe (et n'a pas expiré)
func (s *Storage) Exists(key string) bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Vérifie si la clé a expiré
	if expTime, exists := s.expiry[key]; exists && time.Now().After(expTime) {
		delete(s.data, key)
		delete(s.expiry, key)
		return false
	}

	_, exists := s.data[key]
	return exists
}

// matchRedisPattern implémente le matching de pattern glob style Redis
func matchRedisPattern(pattern, str string) bool {
	return matchPattern(pattern, str, 0, 0)
}

// matchPattern est la fonction récursive qui fait le vrai travail
func matchPattern(pattern, str string, pIdx, sIdx int) bool {
	// Si on a parcouru tout le pattern et toute la string, c'est un match
	if pIdx == len(pattern) && sIdx == len(str) {
		return true
	}

	// Si on a fini le pattern mais pas la string, pas de match
	if pIdx == len(pattern) {
		return false
	}

	// Gestion du caractère d'échappement
	if pattern[pIdx] == '\\' && pIdx+1 < len(pattern) {
		// Le caractère suivant doit matcher exactement
		if sIdx < len(str) && pattern[pIdx+1] == str[sIdx] {
			return matchPattern(pattern, str, pIdx+2, sIdx+1)
		}
		return false
	}

	// Gestion du wildcard *
	if pattern[pIdx] == '*' {
		// * peut matcher 0 ou plusieurs caractères
		// Essayer de matcher 0 caractère
		if matchPattern(pattern, str, pIdx+1, sIdx) {
			return true
		}
		// Essayer de matcher 1 ou plusieurs caractères
		for i := sIdx; i < len(str); i++ {
			if matchPattern(pattern, str, pIdx+1, i+1) {
				return true
			}
		}
		return false
	}

	// Gestion du wildcard ?
	if pattern[pIdx] == '?' {
		// ? doit matcher exactement 1 caractère
		if sIdx < len(str) {
			return matchPattern(pattern, str, pIdx+1, sIdx+1)
		}
		return false
	}

	// Gestion des classes de caractères [...]
	if pattern[pIdx] == '[' {
		// Trouver la fin de la classe
		endIdx := pIdx + 1
		for endIdx < len(pattern) && pattern[endIdx] != ']' {
			if pattern[endIdx] == '\\' && endIdx+1 < len(pattern) {
				endIdx += 2
			} else {
				endIdx++
			}
		}

		if endIdx >= len(pattern) {
			// Pas de ] fermant, traiter [ comme un caractère normal
			if sIdx < len(str) && pattern[pIdx] == str[sIdx] {
				return matchPattern(pattern, str, pIdx+1, sIdx+1)
			}
			return false
		}

		// Vérifier si le caractère actuel match la classe
		if sIdx < len(str) && matchCharClass(pattern[pIdx+1:endIdx], str[sIdx]) {
			return matchPattern(pattern, str, endIdx+1, sIdx+1)
		}
		return false
	}

	// Caractère normal, doit matcher exactement
	if sIdx < len(str) && pattern[pIdx] == str[sIdx] {
		return matchPattern(pattern, str, pIdx+1, sIdx+1)
	}

	return false
}

// matchCharClass vérifie si un caractère correspond à une classe de caractères
func matchCharClass(class string, ch byte) bool {
	negated := false
	idx := 0

	// Vérifier si c'est une classe négative
	if len(class) > 0 && class[0] == '^' {
		negated = true
		idx = 1
	}

	matched := false

	for idx < len(class) {
		if class[idx] == '\\' && idx+1 < len(class) {
			// Caractère échappé
			if class[idx+1] == ch {
				matched = true
				break
			}
			idx += 2
		} else if idx+2 < len(class) && class[idx+1] == '-' {
			// Intervalle de caractères
			if ch >= class[idx] && ch <= class[idx+2] {
				matched = true
				break
			}
			idx += 3
		} else {
			// Caractère simple
			if class[idx] == ch {
				matched = true
				break
			}
			idx++
		}
	}

	if negated {
		return !matched
	}
	return matched
}

// FlushAll vide le stockage
func (s *Storage) FlushAll() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	s.data = make(map[string]string)
	s.expiry = make(map[string]time.Time)
}

// Keys retourne toutes les clés correspondant à un motif (pattern) comme Redis
func (s *Storage) Keys(pattern string) []string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	fmt.Printf("=== DEBUG Keys ===\n")
	fmt.Printf("Pattern reçu: '%s' (longueur: %d)\n", pattern, len(pattern))
	fmt.Printf("Pattern bytes: %v\n", []byte(pattern))

	var result []string
	now := time.Now()

	// Créer une liste des clés valides (non expirées)
	validKeys := make([]string, 0)
	fmt.Printf("Clés dans le storage:\n")
	for key, _ := range s.data {
		fmt.Printf("  - '%s'\n", key)
		// Vérifier si la clé a expiré
		if expTime, exists := s.expiry[key]; exists && now.After(expTime) {
			fmt.Printf("    (expirée)\n")
			continue // Ignorer les clés expirées
		}
		validKeys = append(validKeys, key)
	}

	fmt.Printf("Clés valides: %v\n", validKeys)

	// Si le motif est "*", retourner toutes les clés valides
	if pattern == "*" {
		fmt.Printf("Pattern est '*', retourne toutes les clés\n")
		return validKeys
	}

	// Tester le pattern matching
	fmt.Printf("\nTest du pattern matching:\n")
	for _, key := range validKeys {
		matches := matchRedisPattern(pattern, key)
		fmt.Printf("  matchRedisPattern('%s', '%s') = %v\n", pattern, key, matches)
		if matches {
			result = append(result, key)
		}
	}

	fmt.Printf("Résultat final: %v\n", result)
	fmt.Printf("=================\n")
	return result
}

// Pour tester directement le pattern matching
func TestPatternMatch() {
	fmt.Println("=== TEST DIRECT ===")
	testCases := []struct {
		pattern string
		str     string
	}{
		{"d*", "dog"},
		{"d*", "door"},
		{"d*", "uu"},
		{"*", "anything"},
		{"?oo*", "door"},
		{"[dD]*", "dog"},
	}

	for _, tc := range testCases {
		result := matchRedisPattern(tc.pattern, tc.str)
		fmt.Printf("matchRedisPattern('%s', '%s') = %v\n", tc.pattern, tc.str, result)
	}
	fmt.Println("==================")
}
