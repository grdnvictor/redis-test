package storage

// AddMembersToSet ajoute des membres à un set
func (redisStorage *RedisInMemoryStorage) AddMembersToSet(setKey string, newMembers []string) int {
	redisStorage.storageMutex.Lock()
	defer redisStorage.storageMutex.Unlock()

	storageValue, keyExists := redisStorage.storageData[setKey]
	var redisSetStructure *RedisSetStructure

	if !keyExists {
		redisSetStructure = &RedisSetStructure{SetElements: make(map[string]bool)}
		redisStorage.storageData[setKey] = &RedisStorageValue{
			StoredData: redisSetStructure,
			DataType:   RedisSetType,
		}
	} else {
		if storageValue.DataType != RedisSetType {
			return -1
		}
		redisSetStructure = storageValue.StoredData.(*RedisSetStructure)
	}

	addedMemberCount := 0
	for _, newMember := range newMembers {
		if !redisSetStructure.SetElements[newMember] {
			redisSetStructure.SetElements[newMember] = true
			addedMemberCount++
		}
	}

	return addedMemberCount
}

// GetAllSetMembers retourne tous les membres d'un set
func (redisStorage *RedisInMemoryStorage) GetAllSetMembers(setKey string) []string {
	redisStorage.storageMutex.RLock()
	defer redisStorage.storageMutex.RUnlock()

	storageValue, keyExists := redisStorage.storageData[setKey]
	if !keyExists {
		return []string{}
	}

	if storageValue.DataType != RedisSetType {
		return nil
	}

	redisSetStructure := storageValue.StoredData.(*RedisSetStructure)
	setMembers := make([]string, 0, len(redisSetStructure.SetElements))
	for setMember := range redisSetStructure.SetElements {
		setMembers = append(setMembers, setMember)
	}

	return setMembers
}

// CheckSetMemberExists vérifie si un membre est dans un set
func (redisStorage *RedisInMemoryStorage) CheckSetMemberExists(setKey string, memberToCheck string) bool {
	redisStorage.storageMutex.RLock()
	defer redisStorage.storageMutex.RUnlock()

	storageValue, keyExists := redisStorage.storageData[setKey]
	if !keyExists {
		return false
	}

	if storageValue.DataType != RedisSetType {
		return false
	}

	redisSetStructure := storageValue.StoredData.(*RedisSetStructure)
	return redisSetStructure.SetElements[memberToCheck]
}
