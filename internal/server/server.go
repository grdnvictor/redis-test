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
		return fmt.Errorf("failed to listen on %s: %v", address, err)
	}

	s.listener = listener
	log.Printf("Redis server listening on %s", address)

	// Boucle d'acceptation des connexions
	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-s.shutdown:
				// Arrêt normal du serveur
				return nil
			default:
				log.Printf("Failed to accept connection: %v", err)
				continue
			}
		}

		log.Printf("New connection from %s", conn.RemoteAddr())

		// Vérification du nombre maximum de connexions
		s.clientsMutex.Lock()
		if len(s.clients) >= s.config.MaxConnections {
			s.clientsMutex.Unlock()
			conn.Close()
			log.Printf("Connection rejected: max connections reached (%d)", s.config.MaxConnections)
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
	close(s.shutdown)

	if s.listener != nil {
		s.listener.Close()
	}

	// Fermeture de toutes les connexions clients
	s.clientsMutex.Lock()
	for conn := range s.clients {
		conn.Close()
	}
	s.clientsMutex.Unlock()

	// Attente de la fin de toutes les goroutines
	s.wg.Wait()

	return nil
}

// handleClient gère une connexion client
func (s *Server) handleClient(conn net.Conn) {
	defer s.wg.Done()
	defer func() {
		log.Printf("Connection closed from %s", conn.RemoteAddr())
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
				log.Printf("Parse error from %s: %v", conn.RemoteAddr(), err)
				return
			}

			if len(args) == 0 {
				continue
			}

			// Extraction de la commande et des arguments
			command := args[0]
			commandArgs := args[1:]

			// Exécution de la commande
			if err := s.commands.Execute(command, commandArgs, s.storage, encoder); err != nil {
				log.Printf("Command execution error for %s: %v", conn.RemoteAddr(), err)
				encoder.WriteError("ERR internal server error")
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

		for {
			select {
			case <-s.shutdown:
				return
			case <-ticker.C:
				// Nettoyage des clés expirées
				cleaned := s.storage.CleanupExpired()
				if cleaned > 0 {
					log.Printf("Cleaned %d expired keys", cleaned)
				}
			}
		}
	}()
}
