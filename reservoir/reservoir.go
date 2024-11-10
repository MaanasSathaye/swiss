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

func NewReservoirLoadBalancer(ctx context.Context) *ReservoirLoadBalancer {
	return &ReservoirLoadBalancer{
		servers:  []*server.DummyServer{},
		stopChan: make(chan struct{}),
	}
}

func (lb *ReservoirLoadBalancer) AddServer(host string, port int, connections int32) {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	lb.servers = append(lb.servers, &server.DummyServer{Host: host, Port: port, Connections: connections})
	lb.reservoir = New[*server.DummyServer](len(lb.servers))
}

func (lb *ReservoirLoadBalancer) getServer() *server.DummyServer {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	return lb.reservoir.Sample(lb.servers...)
}

func (lb *ReservoirLoadBalancer) Start(ctx context.Context, host string, port int) error {
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

	log.Printf("Reservoir load balancer started on %s:%d\n", host, port)
	return nil
}

func (lb *ReservoirLoadBalancer) handleConnection(clientConn *net.TCPConn) {
	defer clientConn.Close()

	server := lb.getServer()
	backendAddr := fmt.Sprintf("%s:%d", server.Host, server.Port)
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

func (lb *ReservoirLoadBalancer) Stop() {
	close(lb.stopChan)
	if lb.listener != nil {
		lb.listener.Close()
	}
	lb.wg.Wait()
	log.Printf("Reservoir load balancer stopped.")
}

func New[T any](k int) Reservoir[T] {
	return Reservoir[T]{
		sampled: make([]T, 0, k),
		k:       k,
		gen:     rand.New(rand.NewSource(time.Now().Unix())),
	}
}

type Reservoir[T any] struct {
	sampled []T
	k       int
	gen     *rand.Rand
}

func (r *Reservoir[T]) Sample(items ...T) T {
	r.sampled = r.sampled[:0]
	for idx, i := range items {
		if idx < r.k {
			r.sampled = append(r.sampled, i)
			continue
		}
		j := r.gen.Intn(idx + 1)
		if j < r.k {
			r.sampled[j] = i
		}
	}
	if len(r.sampled) > 0 {
		return r.sampled[r.gen.Intn(len(r.sampled))]
	}
	return items[r.gen.Intn(len(items))]
}
