package rendezvous

import (
	"fmt"
	"hash/fnv"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"

	"github.com/MaanasSathaye/swiss/server"
)

type RendezvousLoadBalancer struct {
	servers []*server.Server
	mutex   sync.Mutex
}

func NewRendezvousLoadBalancer() *RendezvousLoadBalancer {
	return &RendezvousLoadBalancer{
		servers: []*server.Server{},
	}
}

func (lb *RendezvousLoadBalancer) AddServer(srv *server.Server) error {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()
	lb.servers = append(lb.servers, srv)
	return nil
}

func (lb *RendezvousLoadBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	lb.mutex.Lock()
	defer lb.mutex.Unlock()

	var (
		chosenServer *server.Server
		maxWeight    = uint32(0)
		requestKey   = r.URL.Path
	)

	for _, srv := range lb.servers {
		weight := lb.computeWeight(srv, requestKey)
		if weight > maxWeight {
			maxWeight = weight
			chosenServer = srv
		}
	}

	serverURL, err := url.Parse(fmt.Sprintf("http://%s:%d", chosenServer.Stats.Host, chosenServer.Stats.Port))
	if err != nil {
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(serverURL)
	log.Printf("Forwarding request to server: %s", serverURL.String())
	proxy.ServeHTTP(w, r)
}

func (lb *RendezvousLoadBalancer) computeWeight(srv *server.Server, key string) uint32 {
	hash := fnv.New32a()
	hash.Write([]byte(srv.Stats.Host))
	hash.Write([]byte(fmt.Sprintf("%d", srv.Stats.Port)))
	hash.Write([]byte(key))
	return hash.Sum32()
}

func (lb *RendezvousLoadBalancer) StartBalancer(host string, port int) error {
	lbAddr := fmt.Sprintf("%s:%d", host, port)
	return http.ListenAndServe(lbAddr, lb)
}
