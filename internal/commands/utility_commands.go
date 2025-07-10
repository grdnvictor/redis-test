package commands

import (
	"strings"

	"redis-go/internal/protocol"
	"redis-go/internal/storage"
)

// handlePingCommand implémente PING [message]
func (commandRegistry *RedisCommandRegistry) handlePingCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) == 0 {
		return protocolEncoder.WriteSimpleStringResponse("PONG")
	}

	return protocolEncoder.WriteBulkStringResponse(commandArguments[0])
}

// handleEchoCommand implémente ECHO message
func (commandRegistry *RedisCommandRegistry) handleEchoCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) != 1 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'ECHO' (attendu: ECHO message)")
	}

	return protocolEncoder.WriteBulkStringResponse(commandArguments[0])
}

// handleDatabaseSizeCommand implémente DBSIZE
func (commandRegistry *RedisCommandRegistry) handleDatabaseSizeCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) != 0 {
		return protocolEncoder.WriteErrorResponse("ERREUR : DBSIZE ne prend aucun argument")
	}

	return protocolEncoder.WriteIntegerResponse(int64(redisStorage.GetStorageSize()))
}

// handleFlushAllCommand implémente FLUSHALL
func (commandRegistry *RedisCommandRegistry) handleFlushAllCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) != 0 {
		return protocolEncoder.WriteErrorResponse("ERREUR : FLUSHALL ne prend aucun argument")
	}

	redisStorage.FlushAllKeys()
	return protocolEncoder.WriteSimpleStringResponse("OK")
}

// handleHelpCommand implémente ALAIDE [commande] - Version simple et efficace
func (commandRegistry *RedisCommandRegistry) handleHelpCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) == 0 {
		// Liste toutes les commandes séparées par des virgules
		return protocolEncoder.WriteSimpleStringResponse("ALAIDE Redis-Go: SET, GET, DEL, EXISTS, TYPE, INCR, DECR, INCRBY, DECRBY, LPUSH, RPUSH, LPOP, RPOP, LLEN, LRANGE, SADD, SMEMBERS, SISMEMBER, HSET, HGET, HGETALL, PING, ECHO, KEYS, DBSIZE, FLUSHALL - Tapez ALAIDE <commande> pour details")
	}

	// Aide détaillée pour une commande spécifique
	requestedCommand := strings.ToUpper(commandArguments[0])

	switch requestedCommand {
	case "SET":
		return protocolEncoder.WriteSimpleStringResponse("SET key value [EX seconds] - Stocke une valeur avec TTL optionnel en secondes")
	case "GET":
		return protocolEncoder.WriteSimpleStringResponse("GET key - Recupere une valeur. Retourne (nil) si la cle n'existe pas")
	case "DEL":
		return protocolEncoder.WriteSimpleStringResponse("DEL key [key ...] - Supprime une ou plusieurs cles")
	case "EXISTS":
		return protocolEncoder.WriteSimpleStringResponse("EXISTS key [key ...] - Verifie l'existence de cles")
	case "TYPE":
		return protocolEncoder.WriteSimpleStringResponse("TYPE key - Retourne le type de donnees (string, list, set, hash, none)")
	case "INCR":
		return protocolEncoder.WriteSimpleStringResponse("INCR key - Incremente un compteur de 1")
	case "DECR":
		return protocolEncoder.WriteSimpleStringResponse("DECR key - Decremente un compteur de 1")
	case "INCRBY":
		return protocolEncoder.WriteSimpleStringResponse("INCRBY key increment - Incremente un compteur par la valeur donnee")
	case "DECRBY":
		return protocolEncoder.WriteSimpleStringResponse("DECRBY key decrement - Decremente un compteur par la valeur donnee")
	case "LPUSH":
		return protocolEncoder.WriteSimpleStringResponse("LPUSH key element [element ...] - Ajoute des elements au debut de la liste")
	case "RPUSH":
		return protocolEncoder.WriteSimpleStringResponse("RPUSH key element [element ...] - Ajoute des elements a la fin de la liste")
	case "LPOP":
		return protocolEncoder.WriteSimpleStringResponse("LPOP key - Retire et retourne le premier element de la liste")
	case "RPOP":
		return protocolEncoder.WriteSimpleStringResponse("RPOP key - Retire et retourne le dernier element de la liste")
	case "LLEN":
		return protocolEncoder.WriteSimpleStringResponse("LLEN key - Retourne la longueur de la liste")
	case "LRANGE":
		return protocolEncoder.WriteSimpleStringResponse("LRANGE key start stop - Retourne une partie de la liste (indices, -1 = dernier)")
	case "SADD":
		return protocolEncoder.WriteSimpleStringResponse("SADD key member [member ...] - Ajoute des membres uniques a un set")
	case "SMEMBERS":
		return protocolEncoder.WriteSimpleStringResponse("SMEMBERS key - Retourne tous les membres d'un set")
	case "SISMEMBER":
		return protocolEncoder.WriteSimpleStringResponse("SISMEMBER key member - Teste si un membre appartient au set (retourne 1 ou 0)")
	case "HSET":
		return protocolEncoder.WriteSimpleStringResponse("HSET key field value [field value ...] - Definit des champs dans un hash")
	case "HGET":
		return protocolEncoder.WriteSimpleStringResponse("HGET key field - Recupere la valeur d'un champ dans un hash")
	case "HGETALL":
		return protocolEncoder.WriteSimpleStringResponse("HGETALL key - Retourne tous les champs et valeurs d'un hash")
	case "PING":
		return protocolEncoder.WriteSimpleStringResponse("PING [message] - Test de connexion. Retourne PONG ou le message")
	case "ECHO":
		return protocolEncoder.WriteSimpleStringResponse("ECHO message - Retourne le message tel quel")
	case "KEYS":
		return protocolEncoder.WriteSimpleStringResponse("KEYS pattern - Recherche des cles par motif (* = tout, ? = 1 char, [abc] = choix)")
	case "DBSIZE":
		return protocolEncoder.WriteSimpleStringResponse("DBSIZE - Retourne le nombre total de cles dans la base")
	case "FLUSHALL":
		return protocolEncoder.WriteSimpleStringResponse("FLUSHALL - Vide completement la base de donnees")
	default:
		return protocolEncoder.WriteSimpleStringResponse("Commande inconnue. Tapez ALAIDE pour voir toutes les commandes disponibles")
	}
}
