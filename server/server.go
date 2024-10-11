package server

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/MaanasSathaye/swiss/stats"
)

type Server struct {
	alive      bool
	statusChan chan struct{}
	helpers    stats.ServerFuncs
	Stats      stats.ServerConfig
	mux        *http.ServeMux
	mutex      sync.Mutex
}

func NewServer(helpers stats.ServerFuncs, stats stats.ServerConfig) (ns *Server, err error) {
	return &Server{
		Stats:      stats,
		helpers:    helpers,
		alive:      false,
		statusChan: make(chan struct{}, 1),
		mux:        http.NewServeMux(),
		mutex:      sync.Mutex{},
	}, nil
}

func (s *Server) Start() error {
	s.mutex.Lock()
	if s.alive {
		s.mutex.Unlock()
		return fmt.Errorf("server already running")
	}
	s.alive = true
	s.mutex.Unlock()

	lis, err := net.Listen("tcp", s.Stats.Addr())
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
		return err
	}

	s.mux.HandleFunc("/", s.handleConnections)

	go func() {
		log.Printf("HTTP server is listening on %s\n", s.Stats.Addr())
		if err := http.Serve(lis, s.mux); err != nil && s.alive {
			log.Printf("HTTP server error: %v", err)
		}
	}()

	return nil
}

func (s *Server) handleConnections(w http.ResponseWriter, r *http.Request) {
	log.Printf("Received request at %s from %s\n", s.Stats.Addr(), r.RemoteAddr)
	s.mutex.Lock()
	s.Stats.Connections++
	s.Stats.ConnectionsAdded++
	s.mutex.Unlock()

	defer func() {
		s.mutex.Lock()
		s.Stats.Connections--
		s.Stats.ConnectionsRemoved++
		s.mutex.Unlock()
	}()
}

func (s *Server) Stop() {
	s.mutex.Lock()
	s.alive = false
	close(s.statusChan)
	s.mutex.Unlock()
}
