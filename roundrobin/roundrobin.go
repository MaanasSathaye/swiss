package roundrobin

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
)

type RRLoadBalancer struct {
	servers []*url.URL
	current int
	mutex   sync.Mutex
}

func NewRRLoadBalancer() *RRLoadBalancer {
	return &RRLoadBalancer{
		servers: []*url.URL{},
		current: 0,
	}
}

func (lb *RRLoadBalancer) AddServer(host string, port int) error {
	serverURL, err := url.Parse(fmt.Sprintf("http://%s:%d", host, port))
	if err != nil {
		return err
	}
	lb.mutex.Lock()
	defer lb.mutex.Unlock()
	lb.servers = append(lb.servers, serverURL)
	return nil
}

func (lb *RRLoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	lb.mutex.Lock()
	serverURL := lb.servers[lb.current]
	lb.current = (lb.current + 1) % len(lb.servers)
	lb.mutex.Unlock()

	proxy := httputil.NewSingleHostReverseProxy(serverURL)
	log.Printf("Forwarding request to server: %s", serverURL)
	proxy.ServeHTTP(w, r)
}

func (lb *RRLoadBalancer) StartBalancer(host string, port int) error {
	lbAddr := fmt.Sprintf("%s:%d", host, port)
	log.Printf("Load balancer is listening on %s", lbAddr)
	return http.ListenAndServe(lbAddr, lb)
}
