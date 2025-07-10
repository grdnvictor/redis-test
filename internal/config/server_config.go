package config

import (
	"os"
	"strconv"
	"time"
)

// ServerConfiguration contient toute la configuration du serveur Redis
type ServerConfiguration struct {
	NetworkConfiguration     NetworkConfiguration
	PerformanceConfiguration PerformanceConfiguration
	MaintenanceConfiguration MaintenanceConfiguration
}

// NetworkConfiguration gère les paramètres réseau
type NetworkConfiguration struct {
	HostAddress string
	PortNumber  int
}

// PerformanceConfiguration gère les paramètres de performance
type PerformanceConfiguration struct {
	MaximumConnections int
}

// MaintenanceConfiguration gère les paramètres de maintenance
type MaintenanceConfiguration struct {
	ExpirationCheckInterval time.Duration
}

// LoadServerConfiguration charge la configuration depuis les variables d'environnement
// avec des valeurs par défaut raisonnables
func LoadServerConfiguration() *ServerConfiguration {
	configuration := &ServerConfiguration{
		NetworkConfiguration: NetworkConfiguration{
			HostAddress: getEnvironmentString("REDIS_HOST", "localhost"),
			PortNumber:  getEnvironmentInteger("REDIS_PORT", 6379),
		},
		PerformanceConfiguration: PerformanceConfiguration{
			MaximumConnections: getEnvironmentInteger("REDIS_MAX_CONNECTIONS", 1000),
		},
		MaintenanceConfiguration: MaintenanceConfiguration{
			ExpirationCheckInterval: time.Duration(getEnvironmentInteger("REDIS_EXPIRATION_CHECK_INTERVAL", 1)) * time.Second,
		},
	}

	return configuration
}

// getEnvironmentString récupère une variable d'environnement string avec valeur par défaut
func getEnvironmentString(environmentKey, defaultValue string) string {
	if environmentValue := os.Getenv(environmentKey); environmentValue != "" {
		return environmentValue
	}
	return defaultValue
}

// getEnvironmentInteger récupère une variable d'environnement int avec valeur par défaut
func getEnvironmentInteger(environmentKey string, defaultValue int) int {
	if environmentValue := os.Getenv(environmentKey); environmentValue != "" {
		if integerValue, parseError := strconv.Atoi(environmentValue); parseError == nil {
			return integerValue
		}
	}
	return defaultValue
}
