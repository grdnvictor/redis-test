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

	// Enregistrement des commandes de base
	registry.registerBasicCommands()

	return registry
}

// registerBasicCommands enregistre toutes les commandes de base
func (r *Registry) registerBasicCommands() {
	// Commandes String
	r.commands["SET"] = r.handleSet
	r.commands["GET"] = r.handleGet
	r.commands["DEL"] = r.handleDel
	r.commands["EXISTS"] = r.handleExists
	r.commands["KEYS"] = r.handleKeys

	// Commandes utilitaires
	r.commands["PING"] = r.handlePing
	r.commands["ECHO"] = r.handleEcho
	r.commands["DBSIZE"] = r.handleDbSize

	// TODO: Ajouter commandes LIST, SET, HASH, ZSET
}

// Execute exécute une commande donnée
func (r *Registry) Execute(command string, args []string, store *storage.Storage, encoder *protocol.Encoder) error {
	handler, exists := r.commands[strings.ToUpper(command)]
	if !exists {
		return encoder.WriteError(fmt.Sprintf("ERR unknown command '%s'", command))
	}

	return handler(args, store, encoder)
}

// handleSet implémente SET key value [EX seconds]
func (r *Registry) handleSet(args []string, store *storage.Storage, encoder *protocol.Encoder) error {
	if len(args) < 2 {
		return encoder.WriteError("ERR wrong number of arguments for 'set' command")
	}

	key := args[0]
	value := args[1]
	var ttl *time.Duration

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
			ttl = &duration
			i++ // Skip next argument (seconds value)
		}
	}

	store.Set(key, value, storage.TypeString, ttl)
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

	// Vérification du type
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

// handleKeys implémente KEYS <pattern> (pattern matching comme Redis)
func (r *Registry) handleKeys(args []string, store *storage.Storage, encoder *protocol.Encoder) error {
	if len(args) != 1 {
		return encoder.WriteError("ERR wrong number of arguments for 'keys' command")
	}

	pattern := args[0]
	fmt.Printf("DEBUG handleKeys: pattern='%s'\n", pattern) // Ajoutez cette ligne
	keys := store.Keys(pattern)
	return encoder.WriteArray(keys)
}

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

	return encoder.WriteInteger(int64(len(store.Keys("*"))))
}
