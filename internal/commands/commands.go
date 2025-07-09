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
}

// Execute exécute une commande donnée
func (r *Registry) Execute(command string, args []string, store *storage.Storage, encoder *protocol.Encoder) error {
	handler, exists := r.commands[strings.ToUpper(command)]
	if !exists {
		return encoder.WriteError(fmt.Sprintf("ERR unknown command '%s'", command))
	}

	return handler(args, store, encoder)
}

// === COMMANDES STRING ===

// handleSet implémente SET key value [EX seconds]
func (r *Registry) handleSet(args []string, store *storage.Storage, encoder *protocol.Encoder) error {
	if len(args) < 2 {
		return encoder.WriteError("ERR wrong number of arguments for 'set' command")
	}

	key := args[0]
	value := args[1]

	// Parsing des options (EX pour TTL)
	for i := 2; i < len(args); i++ {
		switch strings.ToUpper(args[i]) {
		case "EX":
			if i+1 >= len(args) {
				return encoder.WriteError("ERR syntax error")
			}
			seconds, err := strconv.Atoi(args[i+1])
			if err != nil {
				return encoder.WriteError("ERR value is not an integer or out of range")
			}
			duration := time.Duration(seconds) * time.Second
			store.Set(key, value, storage.TypeString, &duration)
			return encoder.WriteSimpleString("OK")
		}
	}

	// SET sans TTL
	store.Set(key, value, storage.TypeString, nil)
	return encoder.WriteSimpleString("OK")
}

// handleGet implémente GET key
func (r *Registry) handleGet(args []string, store *storage.Storage, encoder *protocol.Encoder) error {
	if len(args) != 1 {
		return encoder.WriteError("ERR wrong number of arguments for 'get' command")
	}

	key := args[0]
	value := store.Get(key)

	if value == nil {
		return encoder.WriteNullBulkString()
	}

	if value.Type != storage.TypeString {
		return encoder.WriteError("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	return encoder.WriteBulkString(value.Data.(string))
}

// handleDel implémente DEL key [key ...]
func (r *Registry) handleDel(args []string, store *storage.Storage, encoder *protocol.Encoder) error {
	if len(args) == 0 {
		return encoder.WriteError("ERR wrong number of arguments for 'del' command")
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
		return encoder.WriteError("ERR wrong number of arguments for 'exists' command")
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
		return encoder.WriteError("ERR wrong number of arguments for 'keys' command")
	}

	pattern := args[0]
	keys := store.Keys(pattern)
	return encoder.WriteArray(keys)
}

// handleType implémente TYPE key
func (r *Registry) handleType(args []string, store *storage.Storage, encoder *protocol.Encoder) error {
	if len(args) != 1 {
		return encoder.WriteError("ERR wrong number of arguments for 'type' command")
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
		return encoder.WriteError("ERR wrong number of arguments for 'incr' command")
	}

	key := args[0]
	value := store.Get(key)

	var intValue int64 = 0
	if value != nil {
		if value.Type != storage.TypeString {
			return encoder.WriteError("WRONGTYPE Operation against a key holding the wrong kind of value")
		}

		var err error
		intValue, err = strconv.ParseInt(value.Data.(string), 10, 64)
		if err != nil {
			return encoder.WriteError("ERR value is not an integer or out of range")
		}
	}

	intValue++
	store.Set(key, strconv.FormatInt(intValue, 10), storage.TypeString, nil)
	return encoder.WriteInteger(intValue)
}

// handleDecr implémente DECR key
func (r *Registry) handleDecr(args []string, store *storage.Storage, encoder *protocol.Encoder) error {
	if len(args) != 1 {
		return encoder.WriteError("ERR wrong number of arguments for 'decr' command")
	}

	key := args[0]
	value := store.Get(key)

	var intValue int64 = 0
	if value != nil {
		if value.Type != storage.TypeString {
			return encoder.WriteError("WRONGTYPE Operation against a key holding the wrong kind of value")
		}

		var err error
		intValue, err = strconv.ParseInt(value.Data.(string), 10, 64)
		if err != nil {
			return encoder.WriteError("ERR value is not an integer or out of range")
		}
	}

	intValue--
	store.Set(key, strconv.FormatInt(intValue, 10), storage.TypeString, nil)
	return encoder.WriteInteger(intValue)
}

// handleIncrBy implémente INCRBY key increment
func (r *Registry) handleIncrBy(args []string, store *storage.Storage, encoder *protocol.Encoder) error {
	if len(args) != 2 {
		return encoder.WriteError("ERR wrong number of arguments for 'incrby' command")
	}

	key := args[0]
	increment, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return encoder.WriteError("ERR value is not an integer or out of range")
	}

	value := store.Get(key)
	var intValue int64 = 0
	if value != nil {
		if value.Type != storage.TypeString {
			return encoder.WriteError("WRONGTYPE Operation against a key holding the wrong kind of value")
		}

		intValue, err = strconv.ParseInt(value.Data.(string), 10, 64)
		if err != nil {
			return encoder.WriteError("ERR value is not an integer or out of range")
		}
	}

	intValue += increment
	store.Set(key, strconv.FormatInt(intValue, 10), storage.TypeString, nil)
	return encoder.WriteInteger(intValue)
}

// handleDecrBy implémente DECRBY key decrement
func (r *Registry) handleDecrBy(args []string, store *storage.Storage, encoder *protocol.Encoder) error {
	if len(args) != 2 {
		return encoder.WriteError("ERR wrong number of arguments for 'decrby' command")
	}

	key := args[0]
	decrement, err := strconv.ParseInt(args[1], 10, 64)
	if err != nil {
		return encoder.WriteError("ERR value is not an integer or out of range")
	}

	value := store.Get(key)
	var intValue int64 = 0
	if value != nil {
		if value.Type != storage.TypeString {
			return encoder.WriteError("WRONGTYPE Operation against a key holding the wrong kind of value")
		}

		intValue, err = strconv.ParseInt(value.Data.(string), 10, 64)
		if err != nil {
			return encoder.WriteError("ERR value is not an integer or out of range")
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
		return encoder.WriteError("ERR wrong number of arguments for 'lpush' command")
	}

	key := args[0]
	elements := args[1:]

	length := store.ListPush(key, elements, true) // true = left
	if length == -1 {
		return encoder.WriteError("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	return encoder.WriteInteger(int64(length))
}

// handleRPush implémente RPUSH key element [element ...]
func (r *Registry) handleRPush(args []string, store *storage.Storage, encoder *protocol.Encoder) error {
	if len(args) < 2 {
		return encoder.WriteError("ERR wrong number of arguments for 'rpush' command")
	}

	key := args[0]
	elements := args[1:]

	length := store.ListPush(key, elements, false) // false = right
	if length == -1 {
		return encoder.WriteError("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	return encoder.WriteInteger(int64(length))
}

// handleLPop implémente LPOP key
func (r *Registry) handleLPop(args []string, store *storage.Storage, encoder *protocol.Encoder) error {
	if len(args) != 1 {
		return encoder.WriteError("ERR wrong number of arguments for 'lpop' command")
	}

	key := args[0]
	element, exists := store.ListPop(key, true) // true = left
	if !exists {
		return encoder.WriteNullBulkString()
	}

	return encoder.WriteBulkString(element)
}

// handleRPop implémente RPOP key
func (r *Registry) handleRPop(args []string, store *storage.Storage, encoder *protocol.Encoder) error {
	if len(args) != 1 {
		return encoder.WriteError("ERR wrong number of arguments for 'rpop' command")
	}

	key := args[0]
	element, exists := store.ListPop(key, false) // false = right
	if !exists {
		return encoder.WriteNullBulkString()
	}

	return encoder.WriteBulkString(element)
}

// handleLLen implémente LLEN key
func (r *Registry) handleLLen(args []string, store *storage.Storage, encoder *protocol.Encoder) error {
	if len(args) != 1 {
		return encoder.WriteError("ERR wrong number of arguments for 'llen' command")
	}

	key := args[0]
	length := store.ListLen(key)
	if length == -1 {
		return encoder.WriteError("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	return encoder.WriteInteger(int64(length))
}

// handleLRange implémente LRANGE key start stop
func (r *Registry) handleLRange(args []string, store *storage.Storage, encoder *protocol.Encoder) error {
	if len(args) != 3 {
		return encoder.WriteError("ERR wrong number of arguments for 'lrange' command")
	}

	key := args[0]
	start, err := strconv.Atoi(args[1])
	if err != nil {
		return encoder.WriteError("ERR value is not an integer or out of range")
	}

	stop, err := strconv.Atoi(args[2])
	if err != nil {
		return encoder.WriteError("ERR value is not an integer or out of range")
	}

	elements := store.ListRange(key, start, stop)
	if elements == nil {
		return encoder.WriteError("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	return encoder.WriteArray(elements)
}

// === COMMANDES SET ===

// handleSAdd implémente SADD key member [member ...]
func (r *Registry) handleSAdd(args []string, store *storage.Storage, encoder *protocol.Encoder) error {
	if len(args) < 2 {
		return encoder.WriteError("ERR wrong number of arguments for 'sadd' command")
	}

	key := args[0]
	members := args[1:]

	added := store.SetAdd(key, members)
	if added == -1 {
		return encoder.WriteError("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	return encoder.WriteInteger(int64(added))
}

// handleSMembers implémente SMEMBERS key
func (r *Registry) handleSMembers(args []string, store *storage.Storage, encoder *protocol.Encoder) error {
	if len(args) != 1 {
		return encoder.WriteError("ERR wrong number of arguments for 'smembers' command")
	}

	key := args[0]
	members := store.SetMembers(key)
	if members == nil {
		return encoder.WriteError("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	return encoder.WriteArray(members)
}

// handleSIsMember implémente SISMEMBER key member
func (r *Registry) handleSIsMember(args []string, store *storage.Storage, encoder *protocol.Encoder) error {
	if len(args) != 2 {
		return encoder.WriteError("ERR wrong number of arguments for 'sismember' command")
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
		return encoder.WriteError("ERR wrong number of arguments for 'hset' command")
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
		return encoder.WriteError("ERR wrong number of arguments for 'hget' command")
	}

	key := args[0]
	field := args[1]

	value, exists := store.HashGet(key, field)
	if !exists {
		return encoder.WriteNullBulkString()
	}

	return encoder.WriteBulkString(value)
}

// handleHGetAll implémente HGETALL key
func (r *Registry) handleHGetAll(args []string, store *storage.Storage, encoder *protocol.Encoder) error {
	if len(args) != 1 {
		return encoder.WriteError("ERR wrong number of arguments for 'hgetall' command")
	}

	key := args[0]
	fields := store.HashGetAll(key)
	if fields == nil {
		return encoder.WriteError("WRONGTYPE Operation against a key holding the wrong kind of value")
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
		return encoder.WriteError("ERR wrong number of arguments for 'echo' command")
	}

	return encoder.WriteBulkString(args[0])
}

// handleDbSize implémente DBSIZE
func (r *Registry) handleDbSize(args []string, store *storage.Storage, encoder *protocol.Encoder) error {
	if len(args) != 0 {
		return encoder.WriteError("ERR wrong number of arguments for 'dbsize' command")
	}

	return encoder.WriteInteger(int64(store.Size()))
}

// handleFlushAll implémente FLUSHALL
func (r *Registry) handleFlushAll(args []string, store *storage.Storage, encoder *protocol.Encoder) error {
	if len(args) != 0 {
		return encoder.WriteError("ERR wrong number of arguments for 'flushall' command")
	}

	store.FlushAll()
	return encoder.WriteSimpleString("OK")
}
