package server

import (
	"net"
	"sync"

	"redis-go/internal/commands"
	"redis-go/internal/config"
	"redis-go/internal/storage"
)

// RedisServerInstance représente le serveur Redis
type RedisServerInstance struct {
	serverConfiguration *config.ServerConfiguration
	redisStorage        *storage.RedisInMemoryStorage
	commandRegistry     *commands.RedisCommandRegistry
	networkListener     net.Listener
	connectedClients    map[net.Conn]bool
	clientsMutex        sync.RWMutex
	shutdownSignal      chan struct{}
	activeGoroutines    sync.WaitGroup
}

// NewRedisServerInstance crée une nouvelle instance de serveur
func NewRedisServerInstance(serverConfiguration *config.ServerConfiguration) *RedisServerInstance {
	redisServerInstance := &RedisServerInstance{
		serverConfiguration: serverConfiguration,
		redisStorage:        storage.NewRedisInMemoryStorage(),
		commandRegistry:     commands.NewRedisCommandRegistry(),
		connectedClients:    make(map[net.Conn]bool),
		shutdownSignal:      make(chan struct{}),
	}

	// Démarrage du garbage collector pour les clés expirées
	redisServerInstance.startExpirationGarbageCollector()

	return redisServerInstance
}
