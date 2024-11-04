package reservoir

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/MaanasSathaye/swiss/server"
)

type ReservoirLoadBalancer struct {
	servers   []*server.DummyServer
	reservoir Reservoir[*server.DummyServer]
	mu        sync.Mutex
	listener  *net.TCPListener
	stopChan  chan struct{}
	wg        sync.WaitGroup
}

// NewReservoirLoadBalancer initializes a new reservoir sampling-based load balancer
func NewReservoirLoadBalancer(ctx context.Context, k int) *ReservoirLoadBalancer {
	return &ReservoirLoadBalancer{
		servers:   []*server.DummyServer{},
		reservoir: New[*server.DummyServer](k),
		stopChan:  make(chan struct{}),
	}
}

// AddServer adds a backend server to the reservoir
func (lb *ReservoirLoadBalancer) AddServer(host string, port, connections int) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.servers = append(lb.servers, &server.DummyServer{Host: host, Port: port, Connections: connections})
}

// getServer selects a server using reservoir sampling
func (lb *ReservoirLoadBalancer) getServer() *server.DummyServer {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	sampled := lb.reservoir.Sample(lb.servers...)
	if len(sampled) > 0 {
		return sampled[0]
	}
	return lb.servers[rand.Intn(len(lb.servers))]
}

// Start begins accepting TCP connections for load balancing with reservoir sampling
func (lb *ReservoirLoadBalancer) Start(ctx context.Context, host string, port int) error {
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

	log.Printf("Reservoir load balancer started on %s:%d\n", host, port)
	return nil
}

// handleConnection manages the connection between client and backend server using reservoir sampling
func (lb *ReservoirLoadBalancer) handleConnection(clientConn *net.TCPConn) {
	defer clientConn.Close()

	server := lb.getServer()
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

// Stop shuts down the reservoir load balancer gracefully
func (lb *ReservoirLoadBalancer) Stop() {
	close(lb.stopChan)
	if lb.listener != nil {
		lb.listener.Close()
	}
	lb.wg.Wait()
	log.Printf("Reservoir load balancer stopped.")
}

type Option[T any] func(*Reservoir[T])

func OptionTest[T any](r *Reservoir[T]) {
	r.gen = rand.New(rand.NewSource(0))
}

func New[T any](k int, options ...Option[T]) Reservoir[T] {
	r := Reservoir[T]{
		position: 0,
		sampled:  make([]T, 0, k),
		k:        k,
		gen:      rand.New(rand.NewSource(time.Now().Unix())),
	}

	for _, opt := range options {
		opt(&r)
	}

	return r
}

type Reservoir[T any] struct {
	sampled  []T
	position int
	k        int
	gen      *rand.Rand
}

func (t *Reservoir[T]) Sample(items ...T) []T {
	for idx, i := range items {
		pos := t.position + idx
		if pos < t.k {
			t.sampled = append(t.sampled, i)
			continue
		}

		j := t.gen.Intn(pos + 1)
		if j < t.k {
			t.sampled[j] = i
		}
	}
	t.position += len(items)

	return t.sampled
}
