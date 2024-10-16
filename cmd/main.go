package main

import (
	"context"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/MaanasSathaye/swiss/leastconnections"
	"github.com/MaanasSathaye/swiss/requests"
	"github.com/MaanasSathaye/swiss/server"
	"github.com/MaanasSathaye/swiss/stats"
)

func main() {
	// t := rand.NewSource(time.Now().UnixNano())

	lb := leastconnections.NewLeastConnectionsLoadBalancer()

	serverCount := 3
	var wg sync.WaitGroup

	for i := 0; i < serverCount; i++ {
		host, port := server.GetHostAndPort()
		srvConfig := stats.ServerConfig{
			Host: host,
			Port: port,
		}
		srv, err := server.NewServer(srvConfig)
		if err != nil {
			log.Fatalf("Failed to create server: %v", err)
		}

		if err := lb.AddServer(srv); err != nil {
			log.Fatalf("Failed to add server: %v", err)
		}

		wg.Add(1)
		go func(s *server.Server) {
			defer wg.Done()
			if err := s.Start(); err != nil {
				log.Fatalf("Server start failed: %v", err)
			}
		}(srv)
	}

	go func() {
		if err := lb.StartBalancer("0.0.0.0", 8080); err != nil {
			log.Fatalf("Load balancer start failed: %v", err)
		}
	}()

	client := &http.Client{}

	for i := 0; i < 1000; i++ {
		simulateRequest(client, lb)
	}

	wg.Wait()
}

func simulateRequest(client *http.Client, lb *leastconnections.LeastConnectionsLoadBalancer) {
	// req, err := http.NewRequest("GET", "http://localhost:8080", nil)
	// if err != nil {
	// 	log.Fatalf("Failed to create request: %v", err)
	// }

	req, err := requests.NewVariableRequest(context.Background(), "http://localhost:8080")
	if err != nil {
		log.Fatalf("Failed to create request: %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Request failed: %v", err)
	}
	resp.Body.Close()

	lb.Mutex.Lock()
	for _, srv := range lb.Servers {
		if rand.Float64() < 0.1 {
			srv.Stop()
			log.Printf("Server %s:%d went down", srv.Stats.Host, srv.Stats.Port)
		}
		log.Printf("Server %s:%d - Connections: %d, ConnectionsAdded: %d, ConnectionsRemoved: %d", srv.Stats.Host, srv.Stats.Port, srv.Stats.Connections, srv.Stats.ConnectionsAdded, srv.Stats.ConnectionsRemoved)
	}
	lb.Mutex.Unlock()

	time.Sleep(10 * time.Millisecond)
}
