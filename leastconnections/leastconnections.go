package leastconnections

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sort"
	"sync"

	"github.com/MaanasSathaye/swiss/stats"
)

type LeastConnectionsLoadBalancer struct {
	servers []stats.ServerConfig
	mutex   sync.Mutex
}

func NewLeastConnectionsLoadBalancer() *LeastConnectionsLoadBalancer {
	return &LeastConnectionsLoadBalancer{
		servers: []stats.ServerConfig{},
	}
}

func (lb *LeastConnectionsLoadBalancer) AddServer(host string, port int) error {
	serverID, err := url.Parse(fmt.Sprintf("http://%s:%d", host, port))
	if err != nil {
		return err
	}
	serverConfig := stats.ServerConfig{
		Id:   serverID.String(),
		Host: host,
		Port: port,
	}
	lb.mutex.Lock()
	defer lb.mutex.Unlock()
	lb.servers = append(lb.servers, serverConfig)
	return nil
}

func (lb *LeastConnectionsLoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		leastConnectionsServers []*stats.ServerConfig
		chosenServer            *stats.ServerConfig
	)

	lb.mutex.Lock()
	minConnections := -1

	// Sort servers by connections (ascending order)
	sort.Slice(lb.servers, func(i, j int) bool {
		return lb.servers[i].Connections < lb.servers[j].Connections
	})

	// Find servers with the least connections
	for _, server := range lb.servers {
		if minConnections == -1 || server.Connections == minConnections {
			leastConnectionsServers = append(leastConnectionsServers, &server)
			minConnections = server.Connections
		} else {
			break // All remaining servers have more connections
		}
	}

	if leastConnectionsServers == nil {
		http.Error(w, "No available servers", http.StatusServiceUnavailable)
		lb.mutex.Unlock()
		return
	}

	// Choose a random server from servers with the least connections
	randomIndex := rand.Intn(len(leastConnectionsServers))
	chosenServer = leastConnectionsServers[randomIndex]

	chosenServer.Connections++
	defer func() {
		chosenServer.Connections--
	}()

	serverAddr := fmt.Sprintf("%s:%d", chosenServer.Host, chosenServer.Port)
	targetURL := &url.URL{
		Scheme: "http",
		Host:   serverAddr,
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	log.Printf("Forwarding request to server: %s", targetURL.String())
	proxy.ServeHTTP(w, r)
	lb.mutex.Unlock()
}

func (lb *LeastConnectionsLoadBalancer) StartBalancer(host string, port int) error {
	lbAddr := fmt.Sprintf("%s:%d", host, port)
	log.Printf("Load balancer is listening on %s", lbAddr)
	return http.ListenAndServe(lbAddr, lb)
}
