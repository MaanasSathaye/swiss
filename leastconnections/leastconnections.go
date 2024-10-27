package leastconnections

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"sync"

	"github.com/MaanasSathaye/swiss/server"
)

type LeastConnectionsLoadBalancer struct {
	servers  []*server.DummyServer
	mu       sync.Mutex
	listener *net.TCPListener
	stopChan chan struct{}
	wg       sync.WaitGroup
}

// NewLeastConnectionsLoadBalancer initializes a new least Connections load balancer
func NewLeastConnectionsLoadBalancer(ctx context.Context) *LeastConnectionsLoadBalancer {
	return &LeastConnectionsLoadBalancer{
		servers:  []*server.DummyServer{},
		stopChan: make(chan struct{}),
	}
}

// AddServer adds a backend server to the load balancer's rotation
func (lb *LeastConnectionsLoadBalancer) AddServer(host string, port, connections int) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.servers = append(lb.servers, &server.DummyServer{Host: host, Port: port, Connections: connections})
}

// getLeastConnectionServer returns the server with the least Connections
func (lb *LeastConnectionsLoadBalancer) getLeastConnectionServer() *server.DummyServer {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	var leastConnServers []*server.DummyServer
	minConnections := -1

	for _, srv := range lb.servers {
		if minConnections < 0 || srv.Connections < minConnections {
			leastConnServers = []*server.DummyServer{srv}
			minConnections = srv.Connections
		} else if srv.Connections == minConnections {
			leastConnServers = append(leastConnServers, srv)
		}
	}

	return leastConnServers[rand.Intn(len(leastConnServers))]
}

// Start begins accepting TCP Connections for load balancing
func (lb *LeastConnectionsLoadBalancer) Start(ctx context.Context, host string, port int) error {
	addr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", host, port))
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
				conn, err := lb.listener.AcceptTCP()
				if err != nil {
					continue
				}
				go lb.handleConnection(conn)
			}
		}
	}()

	log.Printf("Least Connections Load balancer started on %s:%d\n", host, port)
	return nil
}

// handleConnection manages the connection between client and backend server
func (lb *LeastConnectionsLoadBalancer) handleConnection(clientConn *net.TCPConn) {
	defer clientConn.Close()

	server := lb.getLeastConnectionServer()
	backendAddr := fmt.Sprintf("%s:%d", server.Host, server.Port)
	backendConn, err := net.Dial("tcp", backendAddr)
	if err != nil {
		log.Printf("Failed to connect to backend server %s: %v", backendAddr, err)
		return
	}
	defer backendConn.Close()

	server.Connections++

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
	server.Connections--

}

// Stop shuts down the load balancer gracefully
func (lb *LeastConnectionsLoadBalancer) Stop() {
	close(lb.stopChan)
	if lb.listener != nil {
		lb.listener.Close()
	}
	lb.wg.Wait()
	log.Println("Least Connections Load balancer stopped.")
}
