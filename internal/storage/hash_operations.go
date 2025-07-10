package storage

// SetHashField définit un field dans un hash
func (redisStorage *RedisInMemoryStorage) SetHashField(hashKey string, fieldName string, fieldValue string) bool {
	redisStorage.storageMutex.Lock()
	defer redisStorage.storageMutex.Unlock()

	storageValue, keyExists := redisStorage.storageData[hashKey]
	var redisHashStructure *RedisHashStructure

	if !keyExists {
		redisHashStructure = &RedisHashStructure{HashFields: make(map[string]string)}
		redisStorage.storageData[hashKey] = &RedisStorageValue{
			StoredData: redisHashStructure,
			DataType:   RedisHashType,
		}
	} else {
		if storageValue.DataType != RedisHashType {
			return false
		}
		redisHashStructure = storageValue.StoredData.(*RedisHashStructure)
	}

	_, fieldAlreadyExists := redisHashStructure.HashFields[fieldName]
	redisHashStructure.HashFields[fieldName] = fieldValue
	return !fieldAlreadyExists // true si nouveau field
}

// GetHashField récupère un field d'un hash
func (redisStorage *RedisInMemoryStorage) GetHashField(hashKey string, fieldName string) (string, bool) {
	redisStorage.storageMutex.RLock()
	defer redisStorage.storageMutex.RUnlock()

	storageValue, keyExists := redisStorage.storageData[hashKey]
	if !keyExists {
		return "", false
	}

	if storageValue.DataType != RedisHashType {
		return "", false
	}

	redisHashStructure := storageValue.StoredData.(*RedisHashStructure)
	fieldValue, fieldExists := redisHashStructure.HashFields[fieldName]
	return fieldValue, fieldExists
}

// GetAllHashFields retourne tous les fields et valeurs d'un hash
func (redisStorage *RedisInMemoryStorage) GetAllHashFields(hashKey string) map[string]string {
	redisStorage.storageMutex.RLock()
	defer redisStorage.storageMutex.RUnlock()

	storageValue, keyExists := redisStorage.storageData[hashKey]
	if !keyExists {
		return map[string]string{}
	}

	if storageValue.DataType != RedisHashType {
		return nil
	}

	redisHashStructure := storageValue.StoredData.(*RedisHashStructure)
	hashFieldsCopy := make(map[string]string)
	for fieldName, fieldValue := range redisHashStructure.HashFields {
		hashFieldsCopy[fieldName] = fieldValue
	}
	return hashFieldsCopy
}
