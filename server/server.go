package server

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/MaanasSathaye/swiss/stats"
)

type DummyServer struct {
	Host        string
	Port        int
	Connections int
}
type Server struct {
	Alive    bool
	Stats    stats.ConnectionStats
	Host     string
	Port     int
	listener *net.TCPListener
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// NewServer initializes a new Server with dynamically allocated Host and Port.
func NewServer(ctx context.Context) (*Server, error) {
	Host, Port := GetHostAndPort()
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", Host, Port))
	if err != nil {
		return nil, fmt.Errorf("error resolving TCP address: %w", err)
	}

	listener, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return nil, fmt.Errorf("error starting TCP listener: %w", err)
	}

	server := &Server{
		Alive:    true,
		Stats:    stats.ConnectionStats{},
		Host:     Host,
		Port:     Port,
		listener: listener,
		stopChan: make(chan struct{}),
	}

	return server, nil
}

// Start begins accepting TCP connections on the server.
func (s *Server) Start(ctx context.Context) {
	var (
		err  error
		conn *net.TCPConn
	)
	log.Printf("Server starting on %s:%d\n", s.Host, s.Port)
	s.wg.Add(1)

	go func() {
		defer s.wg.Done()
		for {
			select {
			case <-s.stopChan:
				log.Printf("Stopping server %s%d...\n", s.Host, s.Port)
				return
			default:
				s.listener.SetDeadline(time.Now().Add(1 * time.Second))
				conn, err = s.listener.AcceptTCP()
				if err != nil {
					if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
						continue
					}
					continue
				}
				s.wg.Add(1)
				go s.HandleConnection(ctx, conn)
			}
		}
	}()
}

func (s *Server) HandleConnection(ctx context.Context, conn *net.TCPConn) {
	var (
		err error
		n   int
	)
	defer s.wg.Done()
	defer conn.Close()

	log.Printf("New connection from %s\n", conn.RemoteAddr())
	s.Stats.Connections++
	s.Stats.ConnectionsAdded++

	// time.Sleep(time.Duration(rand.Intn(10)) * time.Second)
	time.Sleep(5 * time.Second)

	buffer := make([]byte, 1024)
	s.wg.Add(1)
	for {
		defer s.wg.Done()
		n, err = conn.Read(buffer)
		if err != nil {
			if err == io.EOF {
				log.Printf("Connection from %s closed\n", conn.RemoteAddr())
				s.Stats.UpdatedAt = time.Now()
				return
			}
			log.Printf("Error reading from connection: %v\n", err)
			return
		}
		log.Printf("Received: %s\n", string(buffer[:n]))

		if _, err = conn.Write([]byte("Acknowledged\n")); err != nil {
			log.Printf("Error writing to connection: %v\n", err)
			return
		}
		s.Stats.Connections--
		s.Stats.ConnectionsRemoved++
		s.Stats.UpdatedAt = time.Now()
	}
}

// Stop gracefully shuts down the server.
func (s *Server) Stop() {
	if s.Alive {
		close(s.stopChan)
		s.listener.Close()
		log.Printf("Server %s%d stopped.\n", s.Host, s.Port)
		s.Alive = false
	}
}

func (s *Server) Address() string {
	return fmt.Sprint(s.Host, ":", s.Port)
}

// https://github.com/phayes/freeport/blob/master/freeport.go
// GetFreePort asks the kernel for a free open Port that is ready to use.
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

	Port, err := GetFreePort()
	if err != nil {
		Port = 0
	}
	return "0.0.0.0", Port
}
