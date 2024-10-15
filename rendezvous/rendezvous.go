package rendezvous

import (
	"crypto/md5"
	"fmt"
	"math/big"
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
		maxWeight    = big.NewInt(0)
		requestKey   = []byte(r.URL.Path)
	)

	for _, srv := range lb.servers {
		weight := lb.computeWeight(srv, requestKey)
		if weight.Cmp(maxWeight) == 1 {
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
	proxy.ServeHTTP(w, r)
}

func (lb *RendezvousLoadBalancer) computeWeight(srv *server.Server, key []byte) *big.Int {
	hash := md5.New()
	hash.Write([]byte(srv.Stats.Host))
	hash.Write([]byte(fmt.Sprintf("%d", srv.Stats.Port)))
	hash.Write(key)
	hashBytes := hash.Sum(nil)

	weight := big.NewInt(0)
	weight.SetBytes(hashBytes)
	return weight
}

func (lb *RendezvousLoadBalancer) StartBalancer(host string, port int) error {
	lbAddr := fmt.Sprintf("%s:%d", host, port)
	return http.ListenAndServe(lbAddr, lb)
}
