package leastconnections_test

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/MaanasSathaye/swiss/leastconnections"
	"github.com/MaanasSathaye/swiss/server"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("LeastConnectionsLoadBalancer", func() {
	var (
		lb             *leastconnections.LeastConnectionsLoadBalancer
		srv            *server.Server
		backendServers []*server.Server
		err            error
		s              *server.Server
	)
	ctx, done := context.WithTimeout(context.Background(), 60*time.Second)
	defer done()

	// Helper function to start a server on a dynamic port
	startServer := func() (*server.Server, error) {
		s, err = server.NewServer(ctx)
		if err != nil {
			return nil, err
		}
		go s.Start(ctx)
		return s, nil
	}

	AfterEach(func() {
		for _, s := range backendServers {
			s.Stop()
		}

		if lb != nil {
			lb.Stop()
		}
		time.Sleep(1 * time.Second)
	})

	It("should start and stop the load balancer successfully", func() {
		lb = leastconnections.NewLeastConnectionsLoadBalancer(ctx)
		h, p := server.GetHostAndPort()
		err := lb.Start(ctx, h, p)
		Expect(err).To(BeNil())
	})

	It("should start a server and handle a connection via load balancer", func() {
		var (
			conn net.Conn
			n    int
		)
		srv, err = startServer()
		Expect(err).To(BeNil())

		lb = leastconnections.NewLeastConnectionsLoadBalancer(ctx)
		lb.AddServer(srv.Host, srv.Port, srv.Stats.Connections)
		h, p := server.GetHostAndPort()
		err = lb.Start(ctx, h, p)
		Expect(err).To(BeNil())

		conn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", h, p))
		Expect(err).To(BeNil())
		defer conn.Close()

		_, err = conn.Write([]byte("Hello"))
		Expect(err).To(BeNil())

		buff := make([]byte, 1024)
		n, err = conn.Read(buff)
		Expect(err).To(BeNil())
		Expect(string(buff[:n])).To(ContainSubstring("Acknowledged"))
	})

	It("should distribute connections based on least connections", func() {
		var (
			wg          sync.WaitGroup
			connections = make(map[string]int)
			s0          *server.Server
			conn        net.Conn
		)

		lb = leastconnections.NewLeastConnectionsLoadBalancer(ctx)

		s0, err = startServer()
		Expect(err).To(BeNil())
		s0.Stats.Connections = 100
		backendServers = append(backendServers, s0)
		lb.AddServer(s0.Host, s0.Port, s0.Stats.Connections)

		for i := 0; i < 2; i++ {
			s, err := startServer()
			Expect(err).To(BeNil())
			backendServers = append(backendServers, s)
			lb.AddServer(s.Host, s.Port, s.Stats.Connections)
		}

		h, p := server.GetHostAndPort()
		err := lb.Start(ctx, h, p)
		Expect(err).To(BeNil())

		stopTime := time.Now().Add(10 * time.Second)
		connectionsSent := 0
		for time.Now().Before(stopTime) {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()

				conn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", h, p))
				if err != nil {
					log.Printf("Failed to establish connection: %v", err)
					return
				}
				defer conn.Close()

				_, err = conn.Write([]byte("Hello\n"))
				if err != nil {
					log.Printf("Failed to write to connection: %v", err)
					return
				}

				buff := make([]byte, 1024)
				conn.SetReadDeadline(time.Now().Add(15 * time.Second))
				n, err := conn.Read(buff)
				if err != nil {
					log.Printf("Failed to read from connection: %v", err)
					return
				}
				if n > 0 {
					response := strings.TrimSpace(string(buff[:n]))
					if !strings.Contains(response, "Acknowledged") {
						log.Printf("Connection didn't acknowledge: %s", response)
					}
				}

				for _, s := range backendServers {
					if s.Stats.Connections > 0 {
						connections[fmt.Sprintf("%s:%d", s.Host, s.Port)]++
						break
					}
				}
			}(connectionsSent)
			connectionsSent++
			time.Sleep(100 * time.Millisecond)
		}

		wg.Wait()
		log.Println("All connections processed, now stopping servers and load balancer.")

		for _, s := range backendServers {
			log.Printf("Server %s:%d stats - Connections: %d, Added: %d, Removed: %d",
				s.Host, s.Port, s.Stats.Connections, s.Stats.ConnectionsAdded, s.Stats.ConnectionsRemoved)
		}

		for _, s := range backendServers {
			if s.Stats.Connections == 100 {
				Expect(s.Stats.ConnectionsAdded).To(Equal(int32(0)), fmt.Sprintf("Server %s:%d should not have receveived any connections", s.Host, s.Port))
			} else {
				Expect(s.Stats.ConnectionsAdded).To(BeNumerically("~", int32(50), 1), fmt.Sprintf("Server %s:%d should not have significantly more connections than the least loaded server", s.Host, s.Port))
			}
		}
	})

})
