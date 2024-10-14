package server_test

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/MaanasSathaye/swiss/server"
	"github.com/MaanasSathaye/swiss/stats"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("GetFreePort", func() {
	It("should return a free port", func() {
		port, err := server.GetFreePort()
		Expect(err).To(BeNil())
		Expect(port).Should(BeNumerically(">", 0))

		// Try to listen on the port
		l, err := net.Listen("tcp", "localhost"+":"+strconv.Itoa(port))
		Expect(err).To(BeNil())
		defer l.Close()
	})
})

var _ = Describe("Server", func() {
	var (
		srv  *server.Server
		err  error
		host string
		port int
	)

	BeforeEach(func() {
		host, port = server.GetHostAndPort()
		mockStats := stats.ServerConfig{
			Id:        "test-server",
			Host:      host,
			Port:      port,
			UpdatedAt: time.Now(),
			Load:      0.0,
		}

		srv, err = server.NewServer(mockStats)
		Expect(err).NotTo(HaveOccurred())

		err = srv.Start()
		Expect(err).NotTo(HaveOccurred())

		go func() {
			// Wait for the server to start listening
			time.Sleep(1 * time.Second)

			url := fmt.Sprintf("http://%s:%d", host, port)
			resp, err := http.Get(url)
			Expect(err).NotTo(HaveOccurred())
			defer resp.Body.Close()

			Expect(resp.StatusCode).To(Equal(http.StatusOK))
		}()
	})

	AfterEach(func() {
		srv.Stop()
	})

	It("should start the server and handle an HTTP request", func() {
		Expect(srv.Stats.Connections).To(Equal(0))

		// Wait to give the request time to process
		time.Sleep(3 * time.Second)
		Expect(srv.Stats.ConnectionsAdded).To(Equal(1))
		Expect(srv.Stats.ConnectionsRemoved).To(Equal(1))
	})
})

var _ = Describe("Multiple Servers", func() {
	var (
		servers []*server.Server
	)

	BeforeEach(func() {
		servers = []*server.Server{}
	})
	AfterEach(func() {
		for _, srv := range servers {
			srv.Stop()
		}
	})

	It("should start multiple servers and handle HTTP requests on different ports", func() {
		servercount := 5
		addresses := []string{}

		for i := 0; i < servercount; i++ {
			host, port := server.GetHostAndPort()
			mockStats := stats.ServerConfig{
				Id:        fmt.Sprintf("test-server-%d", i),
				Host:      host,
				Port:      port,
				UpdatedAt: time.Now(),
				Load:      0.0,
			}

			srv, err := server.NewServer(mockStats)
			Expect(err).NotTo(HaveOccurred())

			err = srv.Start()
			Expect(err).NotTo(HaveOccurred())

			servers = append(servers, srv)
			addresses = append(addresses, fmt.Sprintf("http://%s:%d", host, port))
		}

		// Send HTTP requests to each server and verify they respond
		for i, address := range addresses {
			go func(addr string, index int) {
				time.Sleep(1 * time.Second)

				resp, err := http.Get(addr)
				Expect(err).NotTo(HaveOccurred())
				defer resp.Body.Close()

				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				Expect(servers[index].Stats.ConnectionsAdded).To(Equal(1))
				Expect(servers[index].Stats.ConnectionsRemoved).To(Equal(1))
			}(address, i)
		}

		// Give enough time for requests to be processed
		time.Sleep(3 * time.Second)

		for _, srv := range servers {
			Expect(srv.Stats.ConnectionsAdded).To(Equal(1))
			Expect(srv.Stats.ConnectionsRemoved).To(Equal(1))
		}
	})
})
