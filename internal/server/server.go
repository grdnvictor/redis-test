package server

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"redis-go/internal/commands"
	"redis-go/internal/config"
	"redis-go/internal/protocol"
	"redis-go/internal/storage"
)

// Server représente le serveur Redis
type Server struct {
	config       *config.Config
	storage      *storage.Storage
	commands     *commands.Registry
	listener     net.Listener
	clients      map[net.Conn]bool
	clientsMutex sync.RWMutex
	shutdown     chan struct{}
	wg           sync.WaitGroup
}

// New crée une nouvelle instance de serveur
func New(cfg *config.Config) *Server {
	server := &Server{
		config:   cfg,
		storage:  storage.New(),
		commands: commands.NewRegistry(),
		clients:  make(map[net.Conn]bool),
		shutdown: make(chan struct{}),
	}

	// Démarrage du garbage collector pour les clés expirées
	server.startExpirationGC()

	return server
}

// Start démarre le serveur TCP
func (s *Server) Start() error {
	address := fmt.Sprintf("%s:%d", s.config.Host, s.config.Port)

	listener, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("impossible d'écouter sur %s: %v", address, err)
	}

	s.listener = listener
	log.Printf("🚀 Serveur Redis-Go en écoute sur %s", address)

	// Boucle d'acceptation des connexions
	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-s.shutdown:
				// Arrêt normal du serveur
				return nil
			default:
				log.Printf("❌ Erreur lors de l'acceptation de connexion: %v", err)
				continue
			}
		}

		log.Printf("🔗 Nouvelle connexion depuis %s", conn.RemoteAddr())

		// Vérification du nombre maximum de connexions
		s.clientsMutex.Lock()
		if len(s.clients) >= s.config.MaxConnections {
			s.clientsMutex.Unlock()
			conn.Close()
			log.Printf("🚫 Connexion refusée: limite atteinte (%d connexions max)", s.config.MaxConnections)
			continue
		}

		s.clients[conn] = true
		s.clientsMutex.Unlock()

		// Gestion du client dans une goroutine séparée
		s.wg.Add(1)
		go s.handleClient(conn)
	}
}

// Stop arrête le serveur proprement
func (s *Server) Stop() error {
	log.Printf("⏹️  Arrêt du serveur en cours...")
	close(s.shutdown)

	if s.listener != nil {
		s.listener.Close()
	}

	// Fermeture de toutes les connexions clients
	s.clientsMutex.Lock()
	clientCount := len(s.clients)
	for conn := range s.clients {
		conn.Close()
	}
	s.clientsMutex.Unlock()

	if clientCount > 0 {
		log.Printf("🔌 Fermeture de %d connexions clients...", clientCount)
	}

	// Attente de la fin de toutes les goroutines
	s.wg.Wait()

	return nil
}

// handleClient gère une connexion client
func (s *Server) handleClient(conn net.Conn) {
	defer s.wg.Done()
	defer func() {
		log.Printf("🔌 Connexion fermée depuis %s", conn.RemoteAddr())
		conn.Close()
		s.clientsMutex.Lock()
		delete(s.clients, conn)
		s.clientsMutex.Unlock()
	}()

	parser := protocol.NewParser(conn)
	encoder := protocol.NewEncoder(conn)

	// Boucle de traitement des commandes
	for {
		select {
		case <-s.shutdown:
			return
		default:
			// Définir un timeout pour éviter les blocages
			conn.SetReadDeadline(time.Now().Add(30 * time.Second))

			// Parsing de la commande
			args, err := parser.ParseCommand()
			if err != nil {
				// Log différencié selon le type d'erreur
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					log.Printf("⏰ Timeout de connexion pour %s", conn.RemoteAddr())
				} else {
					log.Printf("⚠️  Erreur de parsing depuis %s: %v", conn.RemoteAddr(), err)
				}
				return
			}

			if len(args) == 0 {
				continue
			}

			// Extraction de la commande et des arguments
			command := args[0]
			commandArgs := args[1:]

			// Log des commandes (optionnel, peut être verbeux)
			// log.Printf("📝 Commande reçue de %s: %s %v", conn.RemoteAddr(), command, commandArgs)

			// Exécution de la commande
			if err := s.commands.Execute(command, commandArgs, s.storage, encoder); err != nil {
				log.Printf("❌ Erreur d'exécution de commande pour %s: %v", conn.RemoteAddr(), err)
				encoder.WriteError("ERREUR : erreur interne du serveur")
			}
		}
	}
}

// startExpirationGC démarre le garbage collector pour les clés expirées
func (s *Server) startExpirationGC() {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		ticker := time.NewTicker(s.config.ExpirationCheckInterval)
		defer ticker.Stop()

		log.Printf("🧹 Garbage collector démarré (intervalle: %v)", s.config.ExpirationCheckInterval)

		for {
			select {
			case <-s.shutdown:
				log.Printf("🧹 Arrêt du garbage collector")
				return
			case <-ticker.C:
				// Nettoyage des clés expirées
				cleaned := s.storage.CleanupExpired()
				if cleaned > 0 {
					log.Printf("🧹 Nettoyage: %d clés expirées supprimées", cleaned)
				}
			}
		}
	}()
}
