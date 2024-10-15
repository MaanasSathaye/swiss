package leastconnections

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"

	"github.com/MaanasSathaye/swiss/server"
)

type LeastConnectionsLoadBalancer struct {
	servers []*server.Server
	mutex   sync.Mutex
}

func NewLeastConnectionsLoadBalancer() *LeastConnectionsLoadBalancer {
	return &LeastConnectionsLoadBalancer{
		servers: []*server.Server{},
	}
}

func (lb *LeastConnectionsLoadBalancer) AddServer(srv *server.Server) error {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()
	lb.servers = append(lb.servers, srv)
	return nil
}

func (lb *LeastConnectionsLoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		leastConnServers []*server.Server
		chosenServer     *server.Server
		minConnections   = -1
	)

	lb.mutex.Lock()
	defer lb.mutex.Unlock()
	for i, srv := range lb.servers {
		if i == 0 || srv.Stats.Connections < minConnections {
			leastConnServers = []*server.Server{srv}
			minConnections = srv.Stats.Connections
		} else if srv.Stats.Connections == minConnections {
			leastConnServers = append(leastConnServers, srv)
		}
	}

	chosenServer = leastConnServers[rand.Intn(len(leastConnServers))]
	chosenServer.Stats.Connections++

	defer func() {
		chosenServer.Stats.Connections--
	}()

	serverAddr := fmt.Sprintf("%s:%d", chosenServer.Stats.Host, chosenServer.Stats.Port)
	targetURL := &url.URL{
		Scheme: "http",
		Host:   serverAddr,
	}

	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	proxy.ServeHTTP(w, r)
}

func (lb *LeastConnectionsLoadBalancer) StartBalancer(host string, port int) error {
	lbAddr := fmt.Sprintf("%s:%d", host, port)
	log.Printf("Load balancer is listening on %s", lbAddr)
	return http.ListenAndServe(lbAddr, lb)
}
