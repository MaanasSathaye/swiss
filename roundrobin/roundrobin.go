package roundrobin

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"

	"github.com/MaanasSathaye/swiss/server"
)

type RRLoadBalancer struct {
	current int
	servers []*server.Server
	mutex   sync.Mutex
}

func NewRRLoadBalancer() *RRLoadBalancer {
	return &RRLoadBalancer{
		servers: []*server.Server{},
		current: 0,
	}
}

func (lb *RRLoadBalancer) AddServer(srv *server.Server) error {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()
	lb.servers = append(lb.servers, srv)
	return nil
}

func (lb *RRLoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	lb.mutex.Lock()
	serverURL := lb.servers[lb.current]
	lb.current = (lb.current + 1) % len(lb.servers)
	lb.mutex.Unlock()

	serverAddr := fmt.Sprintf("%s:%d", serverURL.Stats.Host, serverURL.Stats.Port)
	targetURL := &url.URL{
		Scheme: "http",
		Host:   serverAddr,
	}
	proxy := httputil.NewSingleHostReverseProxy(targetURL)
	log.Printf("Forwarding request to server: %s", targetURL)
	proxy.ServeHTTP(w, r)
}

func (lb *RRLoadBalancer) StartBalancer(host string, port int) error {
	lbAddr := fmt.Sprintf("%s:%d", host, port)
	log.Printf("Load balancer is listening on %s", lbAddr)
	return http.ListenAndServe(lbAddr, lb)
}
