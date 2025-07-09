package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"redis-go/internal/storage"
	"syscall"

	"redis-go/internal/config"
	"redis-go/internal/server"
)

func main() {
	// Chargement de la configuration depuis les variables d'environnement ou valeurs par défaut
	cfg := config.Load()

	// Création du serveur Redis
	redisServer := server.New(cfg)

	// Canal pour intercepter les signaux système (CTRL+C, SIGTERM)
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	// Démarrage du serveur dans une goroutine séparée
	go func() {
		log.Printf("Starting Redis server on %s:%d", cfg.Host, cfg.Port)
		if err := redisServer.Start(); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Attente du signal d'arrêt
	<-interrupt
	fmt.Println("\nShutting down server...")

	// Arrêt propre du serveur
	if err := redisServer.Stop(); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}

	log.Println("Server stopped")
	storage.TestPatternMatch()
}
