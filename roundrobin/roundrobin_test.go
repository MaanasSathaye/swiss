package roundrobin_test

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/MaanasSathaye/swiss/roundrobin"
	"github.com/MaanasSathaye/swiss/server"
	"github.com/MaanasSathaye/swiss/stats"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("RoundRobinLoadBalancer", func() {
	var (
		lb      *roundrobin.RRLoadBalancer
		lbHost  string
		lbPort  int
		servers []*server.Server
		srv     *server.Server
		err     error
	)

	BeforeEach(func() {
		lb = roundrobin.NewRRLoadBalancer()

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

	It("should distribute requests across multiple servers", func() {
		requestCount := 20
		responses := map[string]int{}

		for i := 0; i < requestCount; i++ {
			url := fmt.Sprintf("http://%s:%d", lbHost, lbPort)
			resp, err := http.Get(url)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))

			body, err := io.ReadAll(resp.Body)
			Expect(err).NotTo(HaveOccurred())
			resp.Body.Close()

			response := string(body)
			responses[response]++
		}

		Expect(len(responses)).To(Equal(requestCount))

	})
})
