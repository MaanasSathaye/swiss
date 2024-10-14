package leastconnections_test

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/MaanasSathaye/swiss/leastconnections"
	"github.com/MaanasSathaye/swiss/server"
	"github.com/MaanasSathaye/swiss/stats"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("LeastConnectionsLoadBalancer", func() {
	var (
		lb      *leastconnections.LeastConnectionsLoadBalancer
		hosts   []string
		ports   []int
		lbHost  string
		lbPort  int
		servers []*server.Server
		srv1    *server.Server
		srv2    *server.Server
		err     error
	)

	BeforeEach(func() {
		lb = leastconnections.NewLeastConnectionsLoadBalancer()
		hosts = []string{}
		ports = []int{}

		// Create the first server with 5 active connections
		host1, port1 := server.GetHostAndPort()
		mockStats1 := stats.ServerConfig{
			Host:        host1,
			Port:        port1,
			UpdatedAt:   time.Now(),
			Connections: 5,
		}

		srv1, err = server.NewServer(mockStats1)
		Expect(err).NotTo(HaveOccurred())

		err = srv1.Start()
		Expect(err).NotTo(HaveOccurred())

		servers = append(servers, srv1)
		err = lb.AddServer(host1, port1)
		Expect(err).NotTo(HaveOccurred())

		hosts = append(hosts, host1)
		ports = append(ports, port1)

		// Create the second server with 4 active connections
		host2, port2 := server.GetHostAndPort()
		mockStats2 := stats.ServerConfig{
			Host:        host2,
			Port:        port2,
			UpdatedAt:   time.Now(),
			Connections: 4,
		}

		srv2, err = server.NewServer(mockStats2)
		Expect(err).NotTo(HaveOccurred())

		err = srv2.Start()
		Expect(err).NotTo(HaveOccurred())

		servers = append(servers, srv2)
		err = lb.AddServer(host2, port2)
		Expect(err).NotTo(HaveOccurred())

		hosts = append(hosts, host2)
		ports = append(ports, port2)

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

	It("should distribute requests based on least connections", func() {
		requestCount := 20 // Increased request count for better distribution

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
		// TODO THIS TEST DOESN'T ACTUALLY WORK SO I DON'T KNOW IF LEAST CONNECTIONS IS WORKING PROPERLY
		Expect(srv1.Stats.Connections).To(BeNumerically("~", srv2.Stats.Connections))
	})
})
