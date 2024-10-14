package leastconnections_test

import (
	"fmt"
	"net/http"
	"time"

	"github.com/MaanasSathaye/swiss/leastconnections"
	"github.com/MaanasSathaye/swiss/server"
	"github.com/MaanasSathaye/swiss/stats"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

//This test and the one below are poorly written tests. I just haven't figured out how to pass and update a pool of servers with equal connections (yet).

var _ = Describe("LeastConnectionsLoadBalancer Unequal Connections", func() {

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

		host1, port1 := server.GetHostAndPort()
		mockStats1 := stats.ServerConfig{
			Host:        host1,
			Port:        port1,
			UpdatedAt:   time.Now(),
			Connections: 2,
		}

		srv1, err = server.NewServer(mockStats1)
		Expect(err).NotTo(HaveOccurred())
		err = srv1.Start()
		Expect(err).NotTo(HaveOccurred())

		servers = append(servers, srv1)

		err = lb.AddServer(srv1)
		Expect(err).NotTo(HaveOccurred())

		hosts = append(hosts, host1)
		ports = append(ports, port1)

		host2, port2 := server.GetHostAndPort()
		mockStats2 := stats.ServerConfig{
			Host:        host2,
			Port:        port2,
			UpdatedAt:   time.Now(),
			Connections: 10,
		}

		srv2, err = server.NewServer(mockStats2)
		Expect(err).NotTo(HaveOccurred())

		err = srv2.Start()
		Expect(err).NotTo(HaveOccurred())

		servers = append(servers, srv2)
		err = lb.AddServer(srv2)
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
		requestCount := 5

		for i := 0; i < requestCount; i++ {
			url := fmt.Sprintf("http://%s:%d", lbHost, lbPort)
			resp, err := http.Get(url)
			Expect(err).NotTo(HaveOccurred())
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
		}
		Expect(srv1.Stats.ConnectionsAdded).To(BeNumerically(">", srv2.Stats.ConnectionsAdded))
	})
})

// var _ = Describe("LeastConnectionsLoadBalancer Equal Connections", func() {

// 	var (
// 		lb      *leastconnections.LeastConnectionsLoadBalancer
// 		hosts   []string
// 		ports   []int
// 		lbHost  string
// 		lbPort  int
// 		servers []*server.Server
// 		srv1    *server.Server
// 		srv2    *server.Server
// 		err     error
// 	)

// 	BeforeEach(func() {
// 		lb = leastconnections.NewLeastConnectionsLoadBalancer()

// 		host1, port1 := server.GetHostAndPort()
// 		mockStats1 := stats.ServerConfig{
// 			Host:        host1,
// 			Port:        port1,
// 			UpdatedAt:   time.Now(),
// 			Connections: 0,
// 		}

// 		srv1, err = server.NewServer(mockStats1)
// 		Expect(err).NotTo(HaveOccurred())
// 		err = srv1.Start()
// 		Expect(err).NotTo(HaveOccurred())

// 		servers = append(servers, srv1)

// 		err = lb.AddServer(srv1)
// 		Expect(err).NotTo(HaveOccurred())

// 		hosts = append(hosts, host1)
// 		ports = append(ports, port1)

// 		host2, port2 := server.GetHostAndPort()
// 		mockStats2 := stats.ServerConfig{
// 			Host:        host2,
// 			Port:        port2,
// 			UpdatedAt:   time.Now(),
// 			Connections: 0,
// 		}

// 		srv2, err = server.NewServer(mockStats2)
// 		Expect(err).NotTo(HaveOccurred())

// 		err = srv2.Start()
// 		Expect(err).NotTo(HaveOccurred())

// 		servers = append(servers, srv2)
// 		err = lb.AddServer(srv2)
// 		Expect(err).NotTo(HaveOccurred())

// 		hosts = append(hosts, host2)
// 		ports = append(ports, port2)

// 		lbHost, lbPort = server.GetHostAndPort()

// 		go func() {
// 			err = lb.StartBalancer(lbHost, lbPort)
// 			Expect(err).NotTo(HaveOccurred())
// 		}()

// 		time.Sleep(1 * time.Second)
// 	})

// 	AfterEach(func() {
// 		for _, srv := range servers {
// 			srv.Stop()
// 		}
// 	})

// 	It("should distribute requests based on least connections", func() {
// 		requestCount := 10

// 		for i := 0; i < requestCount; i++ {
// 			url := fmt.Sprintf("http://%s:%d", lbHost, lbPort)
// 			resp, err := http.Get(url)
// 			Expect(err).NotTo(HaveOccurred())
// 			Expect(resp.StatusCode).To(Equal(http.StatusOK))
// 		}
// 		Expect(srv1.Stats.ConnectionsAdded).To(BeNumerically("~", srv2.Stats.ConnectionsAdded))
// 	})
// })
