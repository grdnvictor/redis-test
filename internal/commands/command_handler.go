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
	commands := map[string]RedisCommandHandler{
		// Commandes String
		"SET":    commandRegistry.handleSetCommand,
		"GET":    commandRegistry.handleGetCommand,
		"DEL":    commandRegistry.handleDeleteCommand,
		"EXISTS": commandRegistry.handleExistsCommand,
		"KEYS":   commandRegistry.handleKeysCommand,
		"TYPE":   commandRegistry.handleTypeCommand,
		"INCR":   commandRegistry.handleIncrementCommand,
		"DECR":   commandRegistry.handleDecrementCommand,
		"INCRBY": commandRegistry.handleIncrementByCommand,
		"DECRBY": commandRegistry.handleDecrementByCommand,

		// Commandes List
		"LPUSH":  commandRegistry.handleLeftPushCommand,
		"RPUSH":  commandRegistry.handleRightPushCommand,
		"LPOP":   commandRegistry.handleLeftPopCommand,
		"RPOP":   commandRegistry.handleRightPopCommand,
		"LLEN":   commandRegistry.handleListLengthCommand,
		"LRANGE": commandRegistry.handleListRangeCommand,

		// Commandes Set
		"SADD":      commandRegistry.handleSetAddCommand,
		"SMEMBERS":  commandRegistry.handleSetMembersCommand,
		"SISMEMBER": commandRegistry.handleSetIsMemberCommand,

		// Commandes Hash
		"HSET":    commandRegistry.handleHashSetCommand,
		"HGET":    commandRegistry.handleHashGetCommand,
		"HGETALL": commandRegistry.handleHashGetAllCommand,

		// Commandes utilitaires
		"PING":     commandRegistry.handlePingCommand,
		"ECHO":     commandRegistry.handleEchoCommand,
		"DBSIZE":   commandRegistry.handleDatabaseSizeCommand,
		"FLUSHALL": commandRegistry.handleFlushAllCommand,
		"ALAIDE":   commandRegistry.handleHelpCommand,
	}

	for commandName, handler := range commands {
		commandRegistry.registeredCommands[commandName] = handler
	}
}

// ExecuteCommand exécute une commande donnée
func (commandRegistry *RedisCommandRegistry) ExecuteCommand(commandName string, commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	upperCommandName := strings.ToUpper(commandName)
	commandHandler, commandExists := commandRegistry.registeredCommands[upperCommandName]

	if !commandExists {
		suggestion := commandRegistry.findSimilarCommand(upperCommandName)
		if suggestion != "" {
			return protocolEncoder.WriteErrorResponse(fmt.Sprintf("ERREUR : commande inconnue '%s'. Vouliez-vous dire '%s' ?", commandName, suggestion))
		}
		return protocolEncoder.WriteErrorResponse(fmt.Sprintf("ERREUR : commande inconnue '%s'", commandName))
	}

	return commandHandler(commandArguments, redisStorage, protocolEncoder)
}

// findSimilarCommand trouve la commande la plus similaire en utilisant la distance de Levenshtein
func (commandRegistry *RedisCommandRegistry) findSimilarCommand(input string) string {
	minDistance := 3 // Seuil de similarité
	bestMatch := ""

	for commandName := range commandRegistry.registeredCommands {
		distance := levenshteinDistance(input, commandName)
		if distance < minDistance {
			minDistance = distance
			bestMatch = commandName
		}
	}

	return bestMatch
}

// levenshteinDistance calcule la distance de Levenshtein entre deux chaînes
func levenshteinDistance(a, b string) int {
	if len(a) == 0 {
		return len(b)
	}
	if len(b) == 0 {
		return len(a)
	}

	matrix := make([][]int, len(a)+1)
	for i := range matrix {
		matrix[i] = make([]int, len(b)+1)
		matrix[i][0] = i
	}
	for j := 0; j <= len(b); j++ {
		matrix[0][j] = j
	}

	for i := 1; i <= len(a); i++ {
		for j := 1; j <= len(b); j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			matrix[i][j] = minimum(
				matrix[i-1][j]+1,      // suppression
				matrix[i][j-1]+1,      // insertion
				matrix[i-1][j-1]+cost, // substitution
			)
		}
	}

	return matrix[len(a)][len(b)]
}

func minimum(a, b, c int) int {
	if a < b && a < c {
		return a
	}
	if b < c {
		return b
	}
	return c
}
