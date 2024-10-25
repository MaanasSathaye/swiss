package leastconnections_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"sync"
	"time"

	"github.com/MaanasSathaye/swiss/leastconnections"
	"github.com/MaanasSathaye/swiss/server"
	"github.com/MaanasSathaye/swiss/stats"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

//This test and the one below are poorly written tests. I just haven't figured out how to pass and update a pool of servers with equal connections (yet).

// var _ = Describe("LeastConnectionsLoadBalancer Unequal Connections", func() {
// 	var (
// 		lb      *leastconnections.LeastConnectionsLoadBalancer
// 		servers []*server.Server
// 		lbAddr  string
// 	)

// 	BeforeEach(func() {
// 		lb = leastconnections.NewLeastConnectionsLoadBalancer()

// 		servers = []*server.Server{}
// 		for i := 0; i < 2; i++ {
// 			host, port := server.GetHostAndPort()
// 			mockStats := stats.ServerConfig{
// 				Id:          strconv.Itoa(i),
// 				Host:        host,
// 				Port:        port,
// 				UpdatedAt:   time.Now(),
// 				Connections: int(i * 8), // Server 0 has 0, Server 1 has 8
// 			}

// 			srv, err := server.NewServer(mockStats, 5)
// 			Expect(err).NotTo(HaveOccurred())
// 			Expect(srv.Start()).To(Succeed())
// 			servers = append(servers, srv)
// 			Expect(lb.AddServer(srv)).To(Succeed())
// 		}

// 		_, lbPort := server.GetHostAndPort()
// 		lbAddr = fmt.Sprintf("localhost:%d", lbPort)

// 		go func() {
// 			Expect(lb.StartBalancer("localhost", lbPort)).To(Succeed())
// 		}()

// 		time.Sleep(500 * time.Millisecond) // Allow time for server to start
// 	})

// 	AfterEach(func() {
// 		for _, srv := range servers {
// 			srv.Stop()
// 		}
// 	})

// 	It("should distribute requests based on least connections", func() {
// 		totalRequests := 10
// 		wg := &sync.WaitGroup{}
// 		wg.Add(totalRequests)

// 		for i := 0; i < totalRequests; i++ {
// 			go func() {
// 				defer wg.Done()
// 				conn, err := net.Dial("tcp", lbAddr)
// 				Expect(err).NotTo(HaveOccurred())
// 				defer conn.Close()

// 				// Since we're not using http.Get, we'll simulate a request by writing to and reading from the connection
// 				_, err = conn.Write([]byte("GET / HTTP/1.0\r\n\r\n"))
// 				Expect(err).NotTo(HaveOccurred())

// 				buf := make([]byte, 1024)
// 				_, err = conn.Read(buf)
// 				Expect(err).NotTo(HaveOccurred())
// 			}()
// 		}

// 		wg.Wait()

// 		// After all requests, server with initially 0 connections should have handled more requests
// 		Expect(servers[0].Stats.ConnectionsAdded).To(BeNumerically(">", 0))
// 		Expect(servers[1].Stats.ConnectionsAdded).To(BeNumerically("==", 0)) // Assuming no new connections added due to high initial load
// 	})
// })

var _ = Describe("LeastConnectionsLoadBalancer", func() {
	var (
		servers []*server.Server
		lb      *leastconnections.LeastConnectionsLoadBalancer
	)

	BeforeEach(func() {
		lb = leastconnections.NewLeastConnectionsLoadBalancer()

		// Create and start multiple servers
		for i := 0; i < 2; i++ {
			host, port := server.GetHostAndPort()
			mockStats := stats.ServerConfig{
				Id:          strconv.Itoa(i),
				Host:        host,
				Port:        port,
				UpdatedAt:   time.Now(),
				Connections: int(i * 8), // Server 0 has 0, Server 1 has 8
			}
			srv, err := server.NewServer(mockStats, 2) // 2 workers per server
			Expect(err).NotTo(HaveOccurred())

			// Set simulated request duration to 500ms for testing
			srv.SimulateRequestDuration = 500 * time.Millisecond

			err = srv.Start()
			Expect(err).NotTo(HaveOccurred())

			servers = append(servers, srv)
			lb.AddServer(srv) // Add each server to the load balancer
		}
	})

	AfterEach(func() {
		for _, srv := range servers {
			srv.Stop()
		}
	})

	It("should distribute requests based on least connections", func() {
		totalRequests := 50
		wg := &sync.WaitGroup{}
		wg.Add(totalRequests)

		// Send multiple requests concurrently
		for i := 0; i < totalRequests; i++ {
			go func() {
				defer wg.Done()

				recorder := httptest.NewRecorder()
				req, err := http.NewRequest("GET", "/", nil)
				Expect(err).NotTo(HaveOccurred())

				lb.ServeHTTP(recorder, req)

				Expect(recorder.Code).To(Equal(http.StatusOK))

				serverID := recorder.Body.String()
				Expect(serverID).NotTo(BeEmpty())
			}()
		}

		wg.Wait()

		// Check that the load was balanced across servers
		// Using a tolerance of ±2 to account for randomness
		for _, srv := range servers {
			fmt.Printf("Server %s has %d connections added\n", srv.Stats.Id, srv.Stats.ConnectionsAdded)
			Expect(srv.Stats.ConnectionsAdded).To(BeNumerically("~", totalRequests/len(servers), 2))
		}
	})
})
