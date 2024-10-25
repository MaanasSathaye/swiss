package roundrobin

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

type BackendServer struct {
	Host string
	Port int
}

type RRLoadBalancer struct {
	servers      []*BackendServer
	currentIndex int
	mu           sync.Mutex
	listener     *net.TCPListener
	stopChan     chan struct{}
	wg           sync.WaitGroup
}

// NewRRLoadBalancer initializes a new round-robin load balancer
func NewRRLoadBalancer(ctx context.Context) *RRLoadBalancer {
	return &RRLoadBalancer{
		servers:  []*BackendServer{},
		stopChan: make(chan struct{}),
	}
}

// AddServer adds a backend server to the load balancer's rotation
func (lb *RRLoadBalancer) AddServer(host string, port int) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.servers = append(lb.servers, &BackendServer{Host: host, Port: port})
}

// getNextServer returns the next server in round-robin order
func (lb *RRLoadBalancer) getNextServer() *BackendServer {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	server := lb.servers[lb.currentIndex]
	lb.currentIndex = (lb.currentIndex + 1) % len(lb.servers)
	return server
}

// Start begins accepting TCP connections for load balancing
func (lb *RRLoadBalancer) Start(ctx context.Context, host string, port int) error {
	var (
		err  error
		addr *net.TCPAddr
		conn *net.TCPConn
	)
	addr, err = net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return err
	}

	lb.listener, err = net.ListenTCP("tcp", addr)
	if err != nil {
		return err
	}

	lb.wg.Add(1)
	go func() {
		defer lb.wg.Done()
		for {
			select {
			case <-lb.stopChan:
				return
			default:
				conn, err = lb.listener.AcceptTCP()
				if err != nil {
					continue
				}
				go lb.handleConnection(conn)
			}
		}
	}()

	log.Printf("Load balancer started on %s:%d\n", host, port)
	return nil
}

// handleConnection manages the connection between client and backend server
func (lb *RRLoadBalancer) handleConnection(clientConn *net.TCPConn) {
	defer clientConn.Close()

	server := lb.getNextServer()
	backendAddr := fmt.Sprintf("%s:%d", server.Host, server.Port)
	backendConn, err := net.Dial("tcp", backendAddr)
	if err != nil {
		log.Printf("Failed to connect to backend server %s: %v", backendAddr, err)
		return
	}
	// Bi-directional copy
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func() {
		defer wg.Done()
		io.Copy(backendConn, clientConn)
	}()

	go func() {
		defer wg.Done()
		io.Copy(clientConn, backendConn)
	}()

	wg.Wait()
}

// Stop shuts down the load balancer gracefully
func (lb *RRLoadBalancer) Stop() {
	close(lb.stopChan)
	if lb.listener != nil {
		lb.listener.Close()
	}
	lb.wg.Wait()
	log.Printf("Load balancer stopped.")
}
