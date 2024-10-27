package rendezvous

import (
	"context"
	"crypto/md5"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"sync"

	"github.com/MaanasSathaye/swiss/server"
)

type RendezvousHashingLoadBalancer struct {
	servers  []*server.DummyServer
	listener *net.TCPListener
	mu       sync.Mutex
	stopChan chan struct{}
	wg       sync.WaitGroup
}

func NewRendezvousHashingLoadBalancer(ctx context.Context) *RendezvousHashingLoadBalancer {
	return &RendezvousHashingLoadBalancer{
		servers:  []*server.DummyServer{},
		stopChan: make(chan struct{}),
	}
}

// AddServer adds a backend server to the load balancer's rotation
func (lb *RendezvousHashingLoadBalancer) AddServer(host string, port int) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.servers = append(lb.servers, &server.DummyServer{Host: host, Port: port})
}

// Start begins accepting TCP connections for load balancing
func (lb *RendezvousHashingLoadBalancer) Start(ctx context.Context, host string, port int) error {
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

	log.Printf("Rendezvous Hashing Load balancer started on %s:%d\n", host, port)
	return nil
}

// handleConnection manages the connection between client and backend server
func (lb *RendezvousHashingLoadBalancer) handleConnection(clientConn *net.TCPConn) {
	defer clientConn.Close()

	// Assume we read from the connection here
	key := make([]byte, 16)
	if _, err := clientConn.Read(key); err != nil {
		log.Printf("Failed to read key from client connection: %v", err)
		return
	}

	chosenServer, err := lb.getServer(key)
	if err != nil {
		log.Printf("Failed to select server: %v", err)
		return
	}

	backendAddr := fmt.Sprintf("%s:%d", chosenServer.Host, chosenServer.Port)
	backendConn, err := net.Dial("tcp", backendAddr)
	if err != nil {
		log.Printf("Failed to connect to backend server %s: %v", backendAddr, err)
		return
	}
	defer backendConn.Close()

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

// getServer uses rendezvous hashing to choose a backend server
func (lb *RendezvousHashingLoadBalancer) getServer(key []byte) (*server.DummyServer, error) {
	var (
		chosenServer *server.DummyServer
		err          error
	)
	lb.mu.Lock()
	defer lb.mu.Unlock()

	if len(lb.servers) == 0 {
		return nil, fmt.Errorf("no servers available")
	}
	maxWeight := big.NewInt(0)

	for _, srv := range lb.servers {
		weight := lb.computeWeight(srv, key)
		if weight.Cmp(maxWeight) == 1 {
			maxWeight = weight
			chosenServer = srv
		}
	}

	return chosenServer, err
}

// computeWeight calculates the weight for a server using its host, port, and the request key
func (lb *RendezvousHashingLoadBalancer) computeWeight(srv *server.DummyServer, key []byte) *big.Int {
	hash := md5.New()
	hash.Write([]byte(srv.Host))
	hash.Write([]byte(fmt.Sprintf("%d", srv.Port)))
	hash.Write(key)
	hashBytes := hash.Sum(nil)

	weight := big.NewInt(0)
	weight.SetBytes(hashBytes)
	return weight
}

// Stop shuts down the load balancer gracefully
func (lb *RendezvousHashingLoadBalancer) Stop() {
	close(lb.stopChan)
	if lb.listener != nil {
		lb.listener.Close()
	}
	lb.wg.Wait()
	log.Println("Rendezvous Hashing Load balancer stopped.")
}
