package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/MaanasSathaye/swiss/leastconnections"
	"github.com/MaanasSathaye/swiss/server"
	"github.com/MaanasSathaye/swiss/stats"
)

const (
	numRequests  = 100
	numServers   = 3
	downtimeProb = 0.1 // 10% chance of server going down
)

func main() {
	// Create and start servers
	servers := make([]*server.Server, numServers)
	for i := 0; i < numServers; i++ {
		host, port := server.GetHostAndPort()
		mockStats := stats.ServerConfig{
			Id:        fmt.Sprintf("server-%d", i),
			Host:      host,
			Port:      port,
			UpdatedAt: time.Now(),
		}
		srv, err := server.NewServer(mockStats)
		if err != nil {
			log.Fatalf("Failed to create server: %v", err)
		}
		err = srv.Start()
		if err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
		servers[i] = srv
	}

	// Initialize load balancer and add servers
	lb := leastconnections.NewLeastConnectionsLoadBalancer()
	for _, srv := range servers {
		err := lb.AddServer(srv)
		if err != nil {
			log.Fatalf("Failed to add server to load balancer: %v", err)
		}
	}

	// Start the load balancer
	lbHost, lbPort := server.GetHostAndPort()
	go func() {
		err := lb.StartBalancer(lbHost, lbPort)
		if err != nil {
			log.Fatalf("Failed to start load balancer: %v", err)
		}
	}()

	// Give some time for the load balancer to start
	time.Sleep(1 * time.Second)

	// Create a WaitGroup to handle concurrent requests
	var wg sync.WaitGroup

	// Function to send requests
	sendRequest := func() {
		defer wg.Done()
		// Simulate sending a request to the load balancer
		resp, err := http.Get(fmt.Sprintf("http://%s:%d/", lbHost, lbPort))
		if err != nil {
			log.Printf("Request failed: %v", err)
			return
		}
		defer resp.Body.Close()

		// Log the response status
		log.Printf("Response status: %s", resp.Status)

		// Check if server goes down
		if rand.Float64() < downtimeProb {
			// Randomly choose one of the servers to simulate going down
			srvToShutdown := servers[rand.Intn(numServers)]
			srvToShutdown.Stop()
			log.Printf("Server %s:%d has gone down", srvToShutdown.Stats.Host, srvToShutdown.Stats.Port)

			// Update the server list in the load balancer if necessary
			// You might need to handle this in your LoadBalancer struct
		}

		// Log server states
		for _, srv := range servers {
			srv.Mutex.Lock()
			if srv.Alive {
				log.Printf("Server %s:%d - Connections: %d, ConnectionsAdded: %d, ConnectionsRemoved: %d",
					srv.Stats.Host, srv.Stats.Port,
					srv.Stats.Connections, srv.Stats.ConnectionsAdded, srv.Stats.ConnectionsRemoved)
			}
			srv.Mutex.Unlock()
		}
	}

	// Simulate sending requests
	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go sendRequest()
	}

	// Wait for all requests to complete
	wg.Wait()
	log.Println("All requests have been processed.")
}
