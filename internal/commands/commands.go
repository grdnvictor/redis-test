package commands

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"redis-go/internal/protocol"
	"redis-go/internal/storage"
)

// Handler représente une fonction qui traite une commande Redis
type Handler func(args []string, store *storage.Storage, encoder *protocol.Encoder) error

// Registry contient toutes les commandes supportées
type Registry struct {
	commands map[string]Handler
}

// NewRegistry crée un nouveau registre de commandes
func NewRegistry() *Registry {
	registry := &Registry{
		commands: make(map[string]Handler),
	}

	// Enregistrement des commandes
	registry.registerCommands()

	return registry
}

// registerCommands enregistre toutes les commandes supportées
func (r *Registry) registerCommands() {
	// Commandes String
	r.commands["SET"] = r.handleSet
	r.commands["GET"] = r.handleGet
	r.commands["DEL"] = r.handleDel
	r.commands["EXISTS"] = r.handleExists
	r.commands["KEYS"] = r.handleKeys
	r.commands["TYPE"] = r.handleType
	r.commands["INCR"] = r.handleIncr
	r.commands["DECR"] = r.handleDecr
	r.commands["INCRBY"] = r.handleIncrBy
	r.commands["DECRBY"] = r.handleDecrBy

	// Commandes List
	r.commands["LPUSH"] = r.handleLPush
	r.commands["RPUSH"] = r.handleRPush
	r.commands["LPOP"] = r.handleLPop
	r.commands["RPOP"] = r.handleRPop
	r.commands["LLEN"] = r.handleLLen
	r.commands["LRANGE"] = r.handleLRange

	// Commandes Set
	r.commands["SADD"] = r.handleSAdd
	r.commands["SMEMBERS"] = r.handleSMembers
	r.commands["SISMEMBER"] = r.handleSIsMember

	// Commandes Hash
	r.commands["HSET"] = r.handleHSet
	r.commands["HGET"] = r.handleHGet
	r.commands["HGETALL"] = r.handleHGetAll

	// Commandes utilitaires
	r.commands["PING"] = r.handlePing
	r.commands["ECHO"] = r.handleEcho
	r.commands["DBSIZE"] = r.handleDbSize
	r.commands["FLUSHALL"] = r.handleFlushAll
	r.commands["ALAIDE"] = r.handleHelp // À l'aide ! 😄
}

// Execute exécute une commande donnée
func (r *Registry) Execute(command string, args []string, store *storage.Storage, encoder *protocol.Encoder) error {
	handler, exists := r.commands[strings.ToUpper(command)]
	if !exists {
		return encoder.WriteError(fmt.Sprintf("ERREUR : commande inconnue '%s'", command))
	}

	return handler(args, store, encoder)
}

// === COMMANDES STRING ===

// handleSet implémente SET key value [EX seconds]
func (r *Registry) handleSet(args []string, store *storage.Storage, encoder *protocol.Encoder) error {
	if len(args) < 2 {
		return encoder.WriteError("ERREUR : nombre d'arguments incorrect pour 'SET' (attendu: SET clé valeur [EX secondes])")
	}

	key := args[0]
	value := args[1]

	// Parsing des options (EX pour TTL)
	for i := 2; i < len(args); i++ {
		switch strings.ToUpper(args[i]) {
		case "EX":
			if i+1 >= len(args) {
				return encoder.WriteError("ERREUR : valeur manquante après 'EX'")
			}
			seconds, err := strconv.Atoi(args[i+1])
			if err != nil {
				return encoder.WriteError("ERREUR : la valeur après 'EX' doit être un nombre entier")
			}
			if seconds <= 0 {
				return encoder.WriteError("ERREUR : le délai d'expiration doit être positif")
			}
			duration := time.Duration(seconds) * time.Second
			store.Set(key, value, storage.TypeString, &duration)
			return encoder.WriteSimpleString("OK")
		default:
			return encoder.WriteError(fmt.Sprintf("ERREUR : option inconnue '%s' pour SET", args[i]))
		}
	}

	// SET sans TTL
	store.Set(key, value, storage.TypeString, nil)
	return encoder.WriteSimpleString("OK")
}

// handleGet implémente GET key
func (r *Registry) handleGet(args []string, store *storage.Storage, encoder *protocol.Encoder) error {
	if len(args) != 1 {
		return encoder.WriteError("ERREUR : nombre d'arguments incorrect pour 'GET' (attendu: GET clé)")
	}

	key := args[0]
	value := store.Get(key)

	if value == nil {
		return encoder.WriteBulkString("(nil)")
	}

	if value.Type != storage.TypeString {
		return encoder.WriteError("ERREUR : cette clé ne contient pas une chaîne de caractères")
	}

	return encoder.WriteBulkString(value.Data.(string))
}

// handleDel implémente DEL key [key ...]
func (r *Registry) handleDel(args []string, store *storage.Storage, encoder *protocol.Encoder) error {
	if len(args) == 0 {
		return encoder.WriteError("ERREUR : nombre d'arguments incorrect pour 'DEL' (attendu: DEL clé [clé ...])")
	}

	deletedCount := int64(0)
	for _, key := range args {
		if store.Delete(key) {
			deletedCount++
		}
	}

	return encoder.WriteInteger(deletedCount)
}

// handleExists implémente EXISTS key [key ...]
func (r *Registry) handleExists(args []string, store *storage.Storage, encoder *protocol.Encoder) error {
	if len(args) == 0 {
		return encoder.WriteError("ERREUR : nombre d'arguments incorrect pour 'EXISTS' (attendu: EXISTS clé [clé ...])")
	}

	existsCount := int64(0)
	for _, key := range args {
		if store.Exists(key) {
			existsCount++
		}
	}

	return encoder.WriteInteger(existsCount)
}

// handleKeys implémente KEYS <pattern>
func (r *Registry) handleKeys(args []string, store *storage.Storage, encoder *protocol.Encoder) error {
	if len(args) != 1 {
		return encoder.WriteError("ERREUR : nombre d'arguments incorrect pour 'KEYS' (attendu: KEYS motif)")
	}

	pattern := args[0]
	keys := store.Keys(pattern)

	// Si aucune clé trouvée, afficher un message explicite
	if len(keys) == 0 {
		return encoder.WriteBulkString("(empty list or set)")
	}

	// Si des clés trouvées, utiliser WriteArray pour avoir les numéros 1), 2), etc.
	return encoder.WriteArray(keys)
}

// handleType implémente TYPE key
func (r *Registry) handleType(args []string, store *storage.Storage, encoder *protocol.Encoder) error {
	if len(args) != 1 {
		return encoder.WriteError("ERREUR : nombre d'arguments incorrect pour 'TYPE' (attendu: TYPE clé)")
	}

	key := args[0]
	dataType := store.Type(key)

	var typeStr string
	switch dataType {
	case storage.TypeString:
		typeStr = "string"
	case storage.TypeList:
		typeStr = "list"
	case storage.TypeSet:
		typeStr = "set"
	case storage.TypeHash:
		typeStr = "hash"
	case storage.TypeZSet:
		typeStr = "zset"
	default:
		typeStr = "none"
	}

	return encoder.WriteSimpleString(typeStr)
}

// handleIncr implémente INCR key
func (r *Registry) handleIncr(args []string, store *storage.Storage, encoder *protocol.Encoder) error {
	if len(args) != 1 {
		return encoder.WriteError("ERREUR : nombre d'arguments incorrect pour 'INCR' (attendu: INCR clé)")
	}

	key := args[0]
	value := store.Get(key)

	var intValue int64 = 0
	if value != nil {
		if value.Type != storage.TypeString {
			return encoder.WriteError("ERREUR : cette clé ne contient pas une chaîne de caractères")
		}

		var err error
		intValue, err = strconv.ParseInt(value.Data.(string), 10, 64)
		if err != nil {
			return encoder.WriteError("ERREUR : la valeur n'est pas un nombre entier")
		}
	}

	intValue++
	store.Set(key, strconv.FormatInt(intValue, 10), storage.TypeString, nil)
	return encoder.WriteInteger(intValue)
}

// handleDecr implémente DECR key
func (r *Registry) handleDecr(args []string, store *storage.Storage, encoder *protocol.Encoder) error {
	if len(args) != 1 {
		return encoder.WriteError("ERREUR : nombre d'arguments incorrect pour 'DECR' (attendu: DECR clé)")
	}

	key := args[0]
	value := store.Get(key)

	var intValue int64 = 0
	if value != nil {
		if value.Type != storage.TypeString {
			return encoder.WriteError("ERREUR : cette clé ne contient pas une chaîne de caractères")
		}

		var err error
		intValue, err = strconv.ParseInt(value.Data.(string), 10, 64)
		if err != nil {
			return encoder.WriteError("ERREUR : la valeur n'est pas un nombre entier")
		}
	}

	intValue--
	store.Set(key, strconv.FormatInt(intValue, 10), storage.TypeString, nil)
	return encoder.WriteInteger(intValue)
}

// handleIncrBy implémente INCRBY key increment
func (r *Registry) handleIncrBy(args []string, store *storage.Storage, encoder *protocol.Encoder) error {
	if len(args) != 2 {
		return encoder.WriteError("ERREUR : nombre d'arguments incorrect pour 'INCRBY' (attendu: INCRBY clé incrément)")
	}

	key := args[0]
	increment, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return encoder.WriteError("ERREUR : l'incrément doit être un nombre entier")
	}

	value := store.Get(key)
	var intValue int64 = 0
	if value != nil {
		if value.Type != storage.TypeString {
			return encoder.WriteError("ERREUR : cette clé ne contient pas une chaîne de caractères")
		}

		intValue, err = strconv.ParseInt(value.Data.(string), 10, 64)
		if err != nil {
			return encoder.WriteError("ERREUR : la valeur n'est pas un nombre entier")
		}
	}

	intValue += increment
	store.Set(key, strconv.FormatInt(intValue, 10), storage.TypeString, nil)
	return encoder.WriteInteger(intValue)
}

// handleDecrBy implémente DECRBY key decrement
func (r *Registry) handleDecrBy(args []string, store *storage.Storage, encoder *protocol.Encoder) error {
	if len(args) != 2 {
		return encoder.WriteError("ERREUR : nombre d'arguments incorrect pour 'DECRBY' (attendu: DECRBY clé décrément)")
	}

	key := args[0]
	decrement, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return encoder.WriteError("ERREUR : le décrément doit être un nombre entier")
	}

	value := store.Get(key)
	var intValue int64 = 0
	if value != nil {
		if value.Type != storage.TypeString {
			return encoder.WriteError("ERREUR : cette clé ne contient pas une chaîne de caractères")
		}

		intValue, err = strconv.ParseInt(value.Data.(string), 10, 64)
		if err != nil {
			return encoder.WriteError("ERREUR : la valeur n'est pas un nombre entier")
		}
	}

	intValue -= decrement
	store.Set(key, strconv.FormatInt(intValue, 10), storage.TypeString, nil)
	return encoder.WriteInteger(intValue)
}

// === COMMANDES LIST ===

// handleLPush implémente LPUSH key element [element ...]
func (r *Registry) handleLPush(args []string, store *storage.Storage, encoder *protocol.Encoder) error {
	if len(args) < 2 {
		return encoder.WriteError("ERREUR : nombre d'arguments incorrect pour 'LPUSH' (attendu: LPUSH clé élément [élément ...])")
	}

	key := args[0]
	elements := args[1:]

	length := store.ListPush(key, elements, true) // true = left
	if length == -1 {
		return encoder.WriteError("ERREUR : cette clé ne contient pas une liste")
	}

	return encoder.WriteInteger(int64(length))
}

// handleRPush implémente RPUSH key element [element ...]
func (r *Registry) handleRPush(args []string, store *storage.Storage, encoder *protocol.Encoder) error {
	if len(args) < 2 {
		return encoder.WriteError("ERREUR : nombre d'arguments incorrect pour 'RPUSH' (attendu: RPUSH clé élément [élément ...])")
	}

	key := args[0]
	elements := args[1:]

	length := store.ListPush(key, elements, false) // false = right
	if length == -1 {
		return encoder.WriteError("ERREUR : cette clé ne contient pas une liste")
	}

	return encoder.WriteInteger(int64(length))
}

// handleLPop implémente LPOP key
func (r *Registry) handleLPop(args []string, store *storage.Storage, encoder *protocol.Encoder) error {
	if len(args) != 1 {
		return encoder.WriteError("ERREUR : nombre d'arguments incorrect pour 'LPOP' (attendu: LPOP clé)")
	}

	key := args[0]
	element, exists := store.ListPop(key, true) // true = left
	if !exists {
		return encoder.WriteBulkString("(nil)")
	}

	return encoder.WriteBulkString(element)
}

// handleRPop implémente RPOP key
func (r *Registry) handleRPop(args []string, store *storage.Storage, encoder *protocol.Encoder) error {
	if len(args) != 1 {
		return encoder.WriteError("ERREUR : nombre d'arguments incorrect pour 'RPOP' (attendu: RPOP clé)")
	}

	key := args[0]
	element, exists := store.ListPop(key, false) // false = right
	if !exists {
		return encoder.WriteBulkString("(nil)")
	}

	return encoder.WriteBulkString(element)
}

// handleLLen implémente LLEN key
func (r *Registry) handleLLen(args []string, store *storage.Storage, encoder *protocol.Encoder) error {
	if len(args) != 1 {
		return encoder.WriteError("ERREUR : nombre d'arguments incorrect pour 'LLEN' (attendu: LLEN clé)")
	}

	key := args[0]
	length := store.ListLen(key)
	if length == -1 {
		return encoder.WriteError("ERREUR : cette clé ne contient pas une liste")
	}

	return encoder.WriteInteger(int64(length))
}

// handleLRange implémente LRANGE key start stop
func (r *Registry) handleLRange(args []string, store *storage.Storage, encoder *protocol.Encoder) error {
	if len(args) != 3 {
		return encoder.WriteError("ERREUR : nombre d'arguments incorrect pour 'LRANGE' (attendu: LRANGE clé début fin)")
	}

	key := args[0]
	start, err := strconv.Atoi(args[1])
	if err != nil {
		return encoder.WriteError("ERREUR : l'index de début doit être un nombre entier")
	}

	stop, err := strconv.Atoi(args[2])
	if err != nil {
		return encoder.WriteError("ERREUR : l'index de fin doit être un nombre entier")
	}

	elements := store.ListRange(key, start, stop)
	if elements == nil {
		return encoder.WriteError("ERREUR : cette clé ne contient pas une liste")
	}

	return encoder.WriteArray(elements)
}

// === COMMANDES SET ===

// handleSAdd implémente SADD key member [member ...]
func (r *Registry) handleSAdd(args []string, store *storage.Storage, encoder *protocol.Encoder) error {
	if len(args) < 2 {
		return encoder.WriteError("ERREUR : nombre d'arguments incorrect pour 'SADD' (attendu: SADD clé membre [membre ...])")
	}

	key := args[0]
	members := args[1:]

	added := store.SetAdd(key, members)
	if added == -1 {
		return encoder.WriteError("ERREUR : cette clé ne contient pas un ensemble")
	}

	return encoder.WriteInteger(int64(added))
}

// handleSMembers implémente SMEMBERS key
func (r *Registry) handleSMembers(args []string, store *storage.Storage, encoder *protocol.Encoder) error {
	if len(args) != 1 {
		return encoder.WriteError("ERREUR : nombre d'arguments incorrect pour 'SMEMBERS' (attendu: SMEMBERS clé)")
	}

	key := args[0]
	members := store.SetMembers(key)
	if members == nil {
		return encoder.WriteError("ERREUR : cette clé ne contient pas un ensemble")
	}

	return encoder.WriteArray(members)
}

// handleSIsMember implémente SISMEMBER key member
func (r *Registry) handleSIsMember(args []string, store *storage.Storage, encoder *protocol.Encoder) error {
	if len(args) != 2 {
		return encoder.WriteError("ERREUR : nombre d'arguments incorrect pour 'SISMEMBER' (attendu: SISMEMBER clé membre)")
	}

	key := args[0]
	member := args[1]

	isMember := store.SetIsMember(key, member)
	if isMember {
		return encoder.WriteInteger(1)
	}
	return encoder.WriteInteger(0)
}

// === COMMANDES HASH ===

// handleHSet implémente HSET key field value [field value ...]
func (r *Registry) handleHSet(args []string, store *storage.Storage, encoder *protocol.Encoder) error {
	if len(args) < 3 || len(args)%2 == 0 {
		return encoder.WriteError("ERREUR : nombre d'arguments incorrect pour 'HSET' (attendu: HSET clé champ valeur [champ valeur ...])")
	}

	key := args[0]
	fieldsSet := int64(0)

	// Traiter les paires field/value
	for i := 1; i < len(args); i += 2 {
		field := args[i]
		value := args[i+1]

		isNew := store.HashSet(key, field, value)
		if isNew {
			fieldsSet++
		}
	}

	return encoder.WriteInteger(fieldsSet)
}

// handleHGet implémente HGET key field
func (r *Registry) handleHGet(args []string, store *storage.Storage, encoder *protocol.Encoder) error {
	if len(args) != 2 {
		return encoder.WriteError("ERREUR : nombre d'arguments incorrect pour 'HGET' (attendu: HGET clé champ)")
	}

	key := args[0]
	field := args[1]

	value, exists := store.HashGet(key, field)
	if !exists {
		return encoder.WriteBulkString("(nil)")
	}

	return encoder.WriteBulkString(value)
}

// handleHGetAll implémente HGETALL key
func (r *Registry) handleHGetAll(args []string, store *storage.Storage, encoder *protocol.Encoder) error {
	if len(args) != 1 {
		return encoder.WriteError("ERREUR : nombre d'arguments incorrect pour 'HGETALL' (attendu: HGETALL clé)")
	}

	key := args[0]
	fields := store.HashGetAll(key)
	if fields == nil {
		return encoder.WriteError("ERREUR : cette clé ne contient pas un hash")
	}

	// Convertir en array alternant field/value
	result := make([]string, 0, len(fields)*2)
	for field, value := range fields {
		result = append(result, field, value)
	}

	return encoder.WriteArray(result)
}

// === COMMANDES UTILITAIRES ===

// handlePing implémente PING [message]
func (r *Registry) handlePing(args []string, store *storage.Storage, encoder *protocol.Encoder) error {
	if len(args) == 0 {
		return encoder.WriteSimpleString("PONG")
	}

	return encoder.WriteBulkString(args[0])
}

// handleEcho implémente ECHO message
func (r *Registry) handleEcho(args []string, store *storage.Storage, encoder *protocol.Encoder) error {
	if len(args) != 1 {
		return encoder.WriteError("ERREUR : nombre d'arguments incorrect pour 'ECHO' (attendu: ECHO message)")
	}

	return encoder.WriteBulkString(args[0])
}

// handleDbSize implémente DBSIZE
func (r *Registry) handleDbSize(args []string, store *storage.Storage, encoder *protocol.Encoder) error {
	if len(args) != 0 {
		return encoder.WriteError("ERREUR : DBSIZE ne prend aucun argument")
	}

	return encoder.WriteInteger(int64(store.Size()))
}

// handleFlushAll implémente FLUSHALL
func (r *Registry) handleFlushAll(args []string, store *storage.Storage, encoder *protocol.Encoder) error {
	if len(args) != 0 {
		return encoder.WriteError("ERREUR : FLUSHALL ne prend aucun argument")
	}

	store.FlushAll()
	return encoder.WriteSimpleString("OK")
}

// handleHelp implémente ALAIDE [commande] - Version simple et efficace
func (r *Registry) handleHelp(args []string, store *storage.Storage, encoder *protocol.Encoder) error {
	if len(args) == 0 {
		// Liste toutes les commandes séparées par des virgules
		return encoder.WriteSimpleString("ALAIDE Redis-Go: SET, GET, DEL, EXISTS, TYPE, INCR, DECR, INCRBY, DECRBY, LPUSH, RPUSH, LPOP, RPOP, LLEN, LRANGE, SADD, SMEMBERS, SISMEMBER, HSET, HGET, HGETALL, PING, ECHO, KEYS, DBSIZE, FLUSHALL - Tapez ALAIDE <commande> pour details")
	}

	// Aide détaillée pour une commande spécifique
	command := strings.ToUpper(args[0])

	switch command {
	case "SET":
		return encoder.WriteSimpleString("SET key value [EX seconds] - Stocke une valeur avec TTL optionnel en secondes")
	case "GET":
		return encoder.WriteSimpleString("GET key - Recupere une valeur. Retourne (nil) si la cle n'existe pas")
	case "DEL":
		return encoder.WriteSimpleString("DEL key [key ...] - Supprime une ou plusieurs cles")
	case "EXISTS":
		return encoder.WriteSimpleString("EXISTS key [key ...] - Verifie l'existence de cles")
	case "TYPE":
		return encoder.WriteSimpleString("TYPE key - Retourne le type de donnees (string, list, set, hash, none)")
	case "INCR":
		return encoder.WriteSimpleString("INCR key - Incremente un compteur de 1")
	case "DECR":
		return encoder.WriteSimpleString("DECR key - Decremente un compteur de 1")
	case "INCRBY":
		return encoder.WriteSimpleString("INCRBY key increment - Incremente un compteur par la valeur donnee")
	case "DECRBY":
		return encoder.WriteSimpleString("DECRBY key decrement - Decremente un compteur par la valeur donnee")
	case "LPUSH":
		return encoder.WriteSimpleString("LPUSH key element [element ...] - Ajoute des elements au debut de la liste")
	case "RPUSH":
		return encoder.WriteSimpleString("RPUSH key element [element ...] - Ajoute des elements a la fin de la liste")
	case "LPOP":
		return encoder.WriteSimpleString("LPOP key - Retire et retourne le premier element de la liste")
	case "RPOP":
		return encoder.WriteSimpleString("RPOP key - Retire et retourne le dernier element de la liste")
	case "LLEN":
		return encoder.WriteSimpleString("LLEN key - Retourne la longueur de la liste")
	case "LRANGE":
		return encoder.WriteSimpleString("LRANGE key start stop - Retourne une partie de la liste (indices, -1 = dernier)")
	case "SADD":
		return encoder.WriteSimpleString("SADD key member [member ...] - Ajoute des membres uniques a un set")
	case "SMEMBERS":
		return encoder.WriteSimpleString("SMEMBERS key - Retourne tous les membres d'un set")
	case "SISMEMBER":
		return encoder.WriteSimpleString("SISMEMBER key member - Teste si un membre appartient au set (retourne 1 ou 0)")
	case "HSET":
		return encoder.WriteSimpleString("HSET key field value [field value ...] - Definit des champs dans un hash")
	case "HGET":
		return encoder.WriteSimpleString("HGET key field - Recupere la valeur d'un champ dans un hash")
	case "HGETALL":
		return encoder.WriteSimpleString("HGETALL key - Retourne tous les champs et valeurs d'un hash")
	case "PING":
		return encoder.WriteSimpleString("PING [message] - Test de connexion. Retourne PONG ou le message")
	case "ECHO":
		return encoder.WriteSimpleString("ECHO message - Retourne le message tel quel")
	case "KEYS":
		return encoder.WriteSimpleString("KEYS pattern - Recherche des cles par motif (* = tout, ? = 1 char, [abc] = choix)")
	case "DBSIZE":
		return encoder.WriteSimpleString("DBSIZE - Retourne le nombre total de cles dans la base")
	case "FLUSHALL":
		return encoder.WriteSimpleString("FLUSHALL - Vide completement la base de donnees")
	default:
		return encoder.WriteSimpleString("Commande inconnue. Tapez ALAIDE pour voir toutes les commandes disponibles")
	}
}
