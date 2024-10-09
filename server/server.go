package server

import (
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/MaanasSathaye/swiss/stats"
)

type Server struct {
	alive      bool
	statusChan chan struct{}
	helpers    stats.ServerFuncs
	Stats      stats.ServerConfig
	// listener   net.Listener
	mutex sync.Mutex
}

func NewServer(helpers stats.ServerFuncs, stats stats.ServerConfig) (ns *Server, err error) {
	return &Server{
		Stats:      stats,
		helpers:    helpers,
		alive:      false,
		statusChan: make(chan struct{}, 1),
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

	go func() {
		lis, err := net.Listen("tcp", s.Stats.Addr())
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		defer lis.Close()

		for s.alive {
			conn, err := lis.Accept()
			if err != nil {
				log.Printf("failed to accept connection: %v", err)
				continue
			}
			s.mutex.Lock()
			s.Stats.Connections++
			s.Stats.ConnectionsAdded++
			s.mutex.Unlock()
			go s.handleConnection(conn)
		}
	}()

	return nil
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()
	buffer := make([]byte, 1024)
	for {
		_, err := conn.Read(buffer)
		if err != nil {
			s.mutex.Lock()
			s.Stats.Connections--
			s.Stats.ConnectionsRemoved++
			s.mutex.Unlock()
			return
		}
	}
}

func (s *Server) Stop() {
	s.mutex.Lock()
	s.alive = false
	close(s.statusChan)
	s.mutex.Unlock()
}
