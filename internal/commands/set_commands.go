package commands

import (
	"redis-go/internal/protocol"
	"redis-go/internal/storage"
)

// handleSetAddCommand implémente SADD key member [member ...]
func (commandRegistry *RedisCommandRegistry) handleSetAddCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) < 2 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'SADD' (attendu: SADD clé membre [membre ...])")
	}

	setKey := commandArguments[0]
	membersToAdd := commandArguments[1:]

	addedMemberCount := redisStorage.AddMembersToSet(setKey, membersToAdd)
	if addedMemberCount == -1 {
		return protocolEncoder.WriteErrorResponse("ERREUR : cette clé ne contient pas un ensemble")
	}

	return protocolEncoder.WriteIntegerResponse(int64(addedMemberCount))
}

// handleSetMembersCommand implémente SMEMBERS key
func (commandRegistry *RedisCommandRegistry) handleSetMembersCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) != 1 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'SMEMBERS' (attendu: SMEMBERS clé)")
	}

	setKey := commandArguments[0]
	setMembers := redisStorage.GetAllSetMembers(setKey)
	if setMembers == nil {
		return protocolEncoder.WriteErrorResponse("ERREUR : cette clé ne contient pas un ensemble")
	}

	return protocolEncoder.WriteArrayResponse(setMembers)
}

// handleSetIsMemberCommand implémente SISMEMBER key member
func (commandRegistry *RedisCommandRegistry) handleSetIsMemberCommand(commandArguments []string, redisStorage *storage.RedisInMemoryStorage, protocolEncoder *protocol.RedisSerializationProtocolEncoder) error {
	if len(commandArguments) != 2 {
		return protocolEncoder.WriteErrorResponse("ERREUR : nombre d'arguments incorrect pour 'SISMEMBER' (attendu: SISMEMBER clé membre)")
	}

	setKey := commandArguments[0]
	memberToCheck := commandArguments[1]

	memberExists := redisStorage.CheckSetMemberExists(setKey, memberToCheck)
	if memberExists {
		return protocolEncoder.WriteIntegerResponse(1)
	}
	return protocolEncoder.WriteIntegerResponse(0)
}
