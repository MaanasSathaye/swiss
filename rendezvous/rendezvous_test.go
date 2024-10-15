package rendezvous_test

import (
	"fmt"
	"net/http"
	"time"

	"github.com/MaanasSathaye/swiss/rendezvous"
	"github.com/MaanasSathaye/swiss/server"
	"github.com/MaanasSathaye/swiss/stats"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("RendezvousLoadBalancer", func() {
	var (
		lb      *rendezvous.RendezvousLoadBalancer
		lbHost  string
		lbPort  int
		servers []*server.Server
		err     error
		srv     *server.Server
	)

	BeforeEach(func() {
		lb = rendezvous.NewRendezvousLoadBalancer()

		// Create and start 3 mock servers
		for i := 0; i < 3; i++ {
			host, port := server.GetHostAndPort()

			mockStats := stats.ServerConfig{
				Id:        fmt.Sprintf("server-%d", i),
				Host:      host,
				Port:      port,
				UpdatedAt: time.Now(),
			}

			srv, err = server.NewServer(mockStats)
			Expect(err).NotTo(HaveOccurred())

			err = srv.Start()
			Expect(err).NotTo(HaveOccurred())

			servers = append(servers, srv)
			err = lb.AddServer(srv)
			Expect(err).NotTo(HaveOccurred())
		}

		lbHost, lbPort = server.GetHostAndPort()

		go func() {
			err = lb.StartBalancer(lbHost, lbPort)
			Expect(err).NotTo(HaveOccurred())
		}()

		time.Sleep(1 * time.Second)
	})

	AfterEach(func() {
		for _, srv := range servers {
			srv.Stop()
		}
	})

	It("should distribute requests across servers based on hash weight", func() {
		requestCount := 100
		serverRequestCounts := make([]int, len(servers))

		for i := 0; i < requestCount; i++ {
			url := fmt.Sprintf("http://%s:%d/some-path-%d", lbHost, lbPort, i)
			resp, err := http.Get(url)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			for idx, srv := range servers {
				if srv.Alive {
					serverRequestCounts[idx]++
				}
			}
		}
		Expect(serverRequestCounts[0]).To(BeNumerically("~", serverRequestCounts[1]))
		Expect(serverRequestCounts[1]).To(BeNumerically("~", serverRequestCounts[2]))
	})
})
