package storage

import (
	"sync"
	"time"
)

// RedisInMemoryStorage est le stockage principal en mémoire avec gestion de la concurrence
type RedisInMemoryStorage struct {
	storageData  map[string]*RedisStorageValue
	storageMutex sync.RWMutex
}

// NewRedisInMemoryStorage crée une nouvelle instance de stockage
func NewRedisInMemoryStorage() *RedisInMemoryStorage {
	return &RedisInMemoryStorage{
		storageData: make(map[string]*RedisStorageValue),
	}
}

// SetKeyValue stocke une valeur avec type et TTL optionnel
func (redisStorage *RedisInMemoryStorage) SetKeyValue(storageKey string, keyData interface{}, dataType RedisDataType, timeToLive *time.Duration) {
	redisStorage.storageMutex.Lock()
	defer redisStorage.storageMutex.Unlock()

	var expirationTime *time.Time
	if timeToLive != nil {
		calculatedExpiry := time.Now().Add(*timeToLive)
		expirationTime = &calculatedExpiry
	}

	redisStorage.storageData[storageKey] = &RedisStorageValue{
		StoredData:     keyData,
		DataType:       dataType,
		ExpirationTime: expirationTime,
	}
}

// GetKeyValue récupère une valeur, retourne nil si la clé n'existe pas ou a expiré
func (redisStorage *RedisInMemoryStorage) GetKeyValue(storageKey string) *RedisStorageValue {
	redisStorage.storageMutex.RLock()
	defer redisStorage.storageMutex.RUnlock()

	storageValue, keyExists := redisStorage.storageData[storageKey]
	if !keyExists {
		return nil
	}

	// Vérifier l'expiration
	if storageValue.ExpirationTime != nil && time.Now().After(*storageValue.ExpirationTime) {
		// Clé expirée - suppression lazy
		delete(redisStorage.storageData, storageKey)
		return nil
	}

	return storageValue
}

// DeleteKeyValue supprime une clé et retourne true si elle existait
func (redisStorage *RedisInMemoryStorage) DeleteKeyValue(storageKey string) bool {
	redisStorage.storageMutex.Lock()
	defer redisStorage.storageMutex.Unlock()

	_, keyExists := redisStorage.storageData[storageKey]
	if keyExists {
		delete(redisStorage.storageData, storageKey)
	}
	return keyExists
}

// CheckKeyExists vérifie si une clé existe et n'a pas expiré
func (redisStorage *RedisInMemoryStorage) CheckKeyExists(storageKey string) bool {
	redisStorage.storageMutex.RLock()
	defer redisStorage.storageMutex.RUnlock()

	storageValue, keyExists := redisStorage.storageData[storageKey]
	if !keyExists {
		return false
	}

	// Vérifier l'expiration
	if storageValue.ExpirationTime != nil && time.Now().After(*storageValue.ExpirationTime) {
		delete(redisStorage.storageData, storageKey)
		return false
	}

	return true
}

// GetStorageSize retourne le nombre de clés valides (non expirées)
func (redisStorage *RedisInMemoryStorage) GetStorageSize() int {
	redisStorage.storageMutex.RLock()
	defer redisStorage.storageMutex.RUnlock()

	validKeyCount := 0
	currentTime := time.Now()

	for _, storageValue := range redisStorage.storageData {
		if storageValue.ExpirationTime == nil || currentTime.Before(*storageValue.ExpirationTime) {
			validKeyCount++
		}
	}

	return validKeyCount
}

// CleanupExpiredKeys supprime activement les clés expirées
func (redisStorage *RedisInMemoryStorage) CleanupExpiredKeys() int {
	redisStorage.storageMutex.Lock()
	defer redisStorage.storageMutex.Unlock()

	currentTime := time.Now()
	cleanedKeyCount := 0

	for storageKey, storageValue := range redisStorage.storageData {
		if storageValue.ExpirationTime != nil && currentTime.After(*storageValue.ExpirationTime) {
			delete(redisStorage.storageData, storageKey)
			cleanedKeyCount++
		}
	}

	return cleanedKeyCount
}

// FlushAllKeys vide tout le stockage
func (redisStorage *RedisInMemoryStorage) FlushAllKeys() {
	redisStorage.storageMutex.Lock()
	defer redisStorage.storageMutex.Unlock()
	redisStorage.storageData = make(map[string]*RedisStorageValue)
}

// GetKeyDataType retourne le type d'une clé
func (redisStorage *RedisInMemoryStorage) GetKeyDataType(storageKey string) RedisDataType {
	redisStorage.storageMutex.RLock()
	defer redisStorage.storageMutex.RUnlock()

	storageValue, keyExists := redisStorage.storageData[storageKey]
	if !keyExists {
		return -1 // Clé inexistante
	}

	// Vérifier l'expiration
	if storageValue.ExpirationTime != nil && time.Now().After(*storageValue.ExpirationTime) {
		delete(redisStorage.storageData, storageKey)
		return -1
	}

	return storageValue.DataType
}
