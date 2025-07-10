package commands

import (
	"redis-go/internal/protocol"
	"redis-go/internal/storage"
)

// handleHashSetCommand implémente HSET key field value [field value ...]
func (commandRegistry *RedisCommandRegistry) handleHashSetCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) < 3 || len(commandArguments)%2 == 0 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'HSET' (attendu: HSET clé champ valeur [champ valeur ...])")
	}

	hashKey := commandArguments[0]
	newFieldCount := int64(0)

	// Traiter les paires field/value
	for argumentIndex := 1; argumentIndex < len(commandArguments); argumentIndex += 2 {
		fieldName := commandArguments[argumentIndex]
		fieldValue := commandArguments[argumentIndex+1]

		isNewField := redisStorage.SetHashField(hashKey, fieldName, fieldValue)
		if isNewField {
			newFieldCount++
		}
	}

	return protocolEncoder.WriteIntegerResponse(newFieldCount)
}

// handleHashGetCommand implémente HGET key field
func (commandRegistry *RedisCommandRegistry) handleHashGetCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) != 2 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'HGET' (attendu: HGET clé champ)")
	}

	hashKey := commandArguments[0]
	fieldName := commandArguments[1]

	fieldValue, fieldExists := redisStorage.GetHashField(hashKey, fieldName)
	if !fieldExists {
		return protocolEncoder.WriteBulkStringResponse("(nil)")
	}

	return protocolEncoder.WriteBulkStringResponse(fieldValue)
}

// handleHashGetAllCommand implémente HGETALL key
func (commandRegistry *RedisCommandRegistry) handleHashGetAllCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) != 1 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'HGETALL' (attendu: HGETALL clé)")
	}

	hashKey := commandArguments[0]
	hashFields := redisStorage.GetAllHashFields(hashKey)
	if hashFields == nil {
		return protocolEncoder.WriteErrorResponse("ERREUR : cette clé ne contient pas un hash")
	}

	// Convertir en array alternant field/value
	responseArray := make([]string, 0, len(hashFields)*2)
	for fieldName, fieldValue := range hashFields {
		responseArray = append(responseArray, fieldName, fieldValue)
	}

	return protocolEncoder.WriteArrayResponse(responseArray)
}
