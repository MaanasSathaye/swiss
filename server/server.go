package server

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"sync"

	"github.com/MaanasSathaye/swiss/stats"
	"github.com/gofrs/uuid/v5"
)

type Server struct {
	alive      bool
	statusChan chan struct{}
	// helpers    stats.ServerFuncs
	Stats stats.ServerConfig
	mux   *http.ServeMux
	mutex sync.Mutex
}

func NewServer(stats stats.ServerConfig) (ns *Server, err error) {
	return &Server{
		Stats:      stats,
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

	resp, err := uuid.NewV4()
	if err != nil {
		http.Error(w, "Error generating UUID", http.StatusInternalServerError)
		return
	}
	w.Write([]byte(resp.String()))
}

func (s *Server) Stop() {
	s.mutex.Lock()
	s.alive = false
	close(s.statusChan)
	s.mutex.Unlock()
}

// https://github.com/phayes/freeport/blob/master/freeport.go
// GetFreePort asks the kernel for a free open port that is ready to use.
func GetFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func GetHostAndPort() (string, int) {

	port, err := GetFreePort()
	if err != nil {
		port = 0
	}
	return "0.0.0.0", port
}
