package commands

import (
	"fmt"
	"strings"

	"redis-go/internal/protocol"
	"redis-go/internal/storage"
)

// RedisCommandHandler représente une fonction qui traite une commande Redis
type RedisCommandHandler func(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error

// RedisCommandRegistry contient toutes les commandes supportées
type RedisCommandRegistry struct {
	registeredCommands map[string]RedisCommandHandler
}

// NewRedisCommandRegistry crée un nouveau registre de commandes
func NewRedisCommandRegistry() *RedisCommandRegistry {
	commandRegistry := &RedisCommandRegistry{
		registeredCommands: make(map[string]RedisCommandHandler),
	}

	// Enregistrement des commandes
	commandRegistry.registerAllCommands()

	return commandRegistry
}

// registerAllCommands enregistre toutes les commandes supportées
func (commandRegistry *RedisCommandRegistry) registerAllCommands() {
	// Commandes String
	commandRegistry.registeredCommands["SET"] = commandRegistry.handleSetCommand
	commandRegistry.registeredCommands["GET"] = commandRegistry.handleGetCommand
	commandRegistry.registeredCommands["DEL"] = commandRegistry.handleDeleteCommand
	commandRegistry.registeredCommands["EXISTS"] = commandRegistry.handleExistsCommand
	commandRegistry.registeredCommands["KEYS"] = commandRegistry.handleKeysCommand
	commandRegistry.registeredCommands["TYPE"] = commandRegistry.handleTypeCommand
	commandRegistry.registeredCommands["INCR"] = commandRegistry.handleIncrementCommand
	commandRegistry.registeredCommands["DECR"] = commandRegistry.handleDecrementCommand
	commandRegistry.registeredCommands["INCRBY"] = commandRegistry.handleIncrementByCommand
	commandRegistry.registeredCommands["DECRBY"] = commandRegistry.handleDecrementByCommand

	// Commandes List
	commandRegistry.registeredCommands["LPUSH"] = commandRegistry.handleLeftPushCommand
	commandRegistry.registeredCommands["RPUSH"] = commandRegistry.handleRightPushCommand
	commandRegistry.registeredCommands["LPOP"] = commandRegistry.handleLeftPopCommand
	commandRegistry.registeredCommands["RPOP"] = commandRegistry.handleRightPopCommand
	commandRegistry.registeredCommands["LLEN"] = commandRegistry.handleListLengthCommand
	commandRegistry.registeredCommands["LRANGE"] = commandRegistry.handleListRangeCommand

	// Commandes Set
	commandRegistry.registeredCommands["SADD"] = commandRegistry.handleSetAddCommand
	commandRegistry.registeredCommands["SMEMBERS"] = commandRegistry.handleSetMembersCommand
	commandRegistry.registeredCommands["SISMEMBER"] = commandRegistry.handleSetIsMemberCommand

	// Commandes Hash
	commandRegistry.registeredCommands["HSET"] = commandRegistry.handleHashSetCommand
	commandRegistry.registeredCommands["HGET"] = commandRegistry.handleHashGetCommand
	commandRegistry.registeredCommands["HGETALL"] = commandRegistry.handleHashGetAllCommand

	// Commandes utilitaires
	commandRegistry.registeredCommands["PING"] = commandRegistry.handlePingCommand
	commandRegistry.registeredCommands["ECHO"] = commandRegistry.handleEchoCommand
	commandRegistry.registeredCommands["DBSIZE"] = commandRegistry.handleDatabaseSizeCommand
	commandRegistry.registeredCommands["FLUSHALL"] = commandRegistry.handleFlushAllCommand
	commandRegistry.registeredCommands["ALAIDE"] = commandRegistry.handleHelpCommand
}

// ExecuteCommand exécute une commande donnée
func (commandRegistry *RedisCommandRegistry) ExecuteCommand(commandName string, commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	commandHandler, commandExists := commandRegistry.registeredCommands[strings.ToUpper(commandName)]
	if !commandExists {
		return protocolEncoder.WriteErrorResponse(fmt.Sprintf("ERREUR : commande inconnue '%s'", commandName))
	}

	return commandHandler(commandArguments, redisStorage, protocolEncoder)
}
