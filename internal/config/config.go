package config

import (
	"os"
	"strconv"
	"time"
)

// Config contient toute la configuration du serveur Redis
type Config struct {
	Host                    string
	Port                    int
	MaxConnections          int
	ExpirationCheckInterval time.Duration
	// TODO: Ajouter config pour persistence (RDB/AOF)
}

// Load charge la configuration depuis les variables d'environnement
// ou utilise des valeurs par défaut raisonnables
func Load() *Config {
	cfg := &Config{
		Host:                    getEnvString("REDIS_HOST", "localhost"),
		Port:                    getEnvInt("REDIS_PORT", 6379),
		MaxConnections:          getEnvInt("REDIS_MAX_CONNECTIONS", 1000),
		ExpirationCheckInterval: time.Duration(getEnvInt("REDIS_EXPIRATION_CHECK_INTERVAL", 1)) * time.Second,
	}

	return cfg
}

// getEnvString récupère une variable d'environnement string avec une valeur par défaut
func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvInt récupère une variable d'environnement int avec une valeur par défaut
func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
