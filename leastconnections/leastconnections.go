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
	Servers []*server.Server
	Mutex   sync.Mutex
}

func NewLeastConnectionsLoadBalancer() *LeastConnectionsLoadBalancer {
	return &LeastConnectionsLoadBalancer{
		Servers: []*server.Server{},
	}
}

func (lb *LeastConnectionsLoadBalancer) AddServer(srv *server.Server) error {
	lb.Mutex.Lock()
	defer lb.Mutex.Unlock()
	lb.Servers = append(lb.Servers, srv)
	return nil
}

func (lb *LeastConnectionsLoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var (
		leastConnServers []*server.Server
		chosenServer     *server.Server
		minConnections   = -1
	)

	lb.Mutex.Lock()
	defer lb.Mutex.Unlock()
	for i, srv := range lb.Servers {
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
