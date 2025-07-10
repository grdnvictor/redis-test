package storage

// PushElementsToList ajoute des éléments à une liste (gauche ou droite)
func (redisStorage *RedisInMemoryStorage) PushElementsToList(listKey string, newElements []string, pushToLeft bool) int {
	redisStorage.storageMutex.Lock()
	defer redisStorage.storageMutex.Unlock()

	storageValue, keyExists := redisStorage.storageData[listKey]
	var redisListStructure *RedisListStructure

	if !keyExists {
		// Créer une nouvelle liste
		redisListStructure = &RedisListStructure{ListElements: make([]string, 0)}
		redisStorage.storageData[listKey] = &RedisStorageValue{
			StoredData: redisListStructure,
			DataType:   RedisListType,
		}
	} else {
		// Vérifier que c'est bien une liste
		if storageValue.DataType != RedisListType {
			return -1 // Erreur de type
		}
		redisListStructure = storageValue.StoredData.(*RedisListStructure)
	}

	// Ajouter les éléments
	if pushToLeft {
		// LPUSH - ajouter à gauche (début)
		updatedElements := make([]string, len(newElements)+len(redisListStructure.ListElements))
		copy(updatedElements, newElements)
		copy(updatedElements[len(newElements):], redisListStructure.ListElements)
		redisListStructure.ListElements = updatedElements
	} else {
		// RPUSH - ajouter à droite (fin)
		redisListStructure.ListElements = append(redisListStructure.ListElements, newElements...)
	}

	return len(redisListStructure.ListElements)
}

// PopElementFromList supprime et retourne un élément de la liste
func (redisStorage *RedisInMemoryStorage) PopElementFromList(listKey string, popFromLeft bool) (string, bool) {
	redisStorage.storageMutex.Lock()
	defer redisStorage.storageMutex.Unlock()

	storageValue, keyExists := redisStorage.storageData[listKey]
	if !keyExists {
		return "", false
	}

	if storageValue.DataType != RedisListType {
		return "", false
	}

	redisListStructure := storageValue.StoredData.(*RedisListStructure)
	if len(redisListStructure.ListElements) == 0 {
		return "", false
	}

	var poppedElement string
	if popFromLeft {
		// LPOP - supprimer à gauche
		poppedElement = redisListStructure.ListElements[0]
		redisListStructure.ListElements = redisListStructure.ListElements[1:]
	} else {
		// RPOP - supprimer à droite
		poppedElement = redisListStructure.ListElements[len(redisListStructure.ListElements)-1]
		redisListStructure.ListElements = redisListStructure.ListElements[:len(redisListStructure.ListElements)-1]
	}

	// Supprimer la clé si la liste est vide
	if len(redisListStructure.ListElements) == 0 {
		delete(redisStorage.storageData, listKey)
	}

	return poppedElement, true
}

// GetListLength retourne la longueur d'une liste
func (redisStorage *RedisInMemoryStorage) GetListLength(listKey string) int {
	redisStorage.storageMutex.RLock()
	defer redisStorage.storageMutex.RUnlock()

	storageValue, keyExists := redisStorage.storageData[listKey]
	if !keyExists {
		return 0
	}

	if storageValue.DataType != RedisListType {
		return -1 // Erreur de type
	}

	redisListStructure := storageValue.StoredData.(*RedisListStructure)
	return len(redisListStructure.ListElements)
}

// GetListElementsInRange retourne une partie de la liste
func (redisStorage *RedisInMemoryStorage) GetListElementsInRange(listKey string, startIndex, stopIndex int) []string {
	redisStorage.storageMutex.RLock()
	defer redisStorage.storageMutex.RUnlock()

	storageValue, keyExists := redisStorage.storageData[listKey]
	if !keyExists {
		return []string{}
	}

	if storageValue.DataType != RedisListType {
		return nil // Erreur de type
	}

	redisListStructure := storageValue.StoredData.(*RedisListStructure)
	listLength := len(redisListStructure.ListElements)

	if listLength == 0 {
		return []string{}
	}

	// Gérer les indices négatifs (comme Redis)
	if startIndex < 0 {
		startIndex = listLength + startIndex
	}
	if stopIndex < 0 {
		stopIndex = listLength + stopIndex
	}

	// Limiter aux bornes
	if startIndex < 0 {
		startIndex = 0
	}
	if stopIndex >= listLength {
		stopIndex = listLength - 1
	}
	if startIndex > stopIndex {
		return []string{}
	}

	return redisListStructure.ListElements[startIndex : stopIndex+1]
}
