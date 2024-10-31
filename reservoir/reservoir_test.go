package reservoir_test

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/MaanasSathaye/swiss/reservoir"
	"github.com/MaanasSathaye/swiss/server"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("reservoir.ReservoirLoadBalancer", func() {
	var (
		lb             *reservoir.ReservoirLoadBalancer
		srv            *server.Server
		backendServers []*server.Server
		err            error
		s              *server.Server
	)
	ctx, done := context.WithTimeout(context.Background(), 15*time.Second)
	defer done()

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
		lb = reservoir.NewReservoirLoadBalancer(ctx, 3)
		h, p := server.GetHostAndPort()
		err := lb.Start(ctx, h, p)
		Expect(err).To(BeNil())

		time.Sleep(1 * time.Second)
	})

	It("should start a server and handle a connection via load balancer", func() {
		srv, err = startServer()
		Expect(err).To(BeNil())

		lb = reservoir.NewReservoirLoadBalancer(ctx, 1)
		lb.AddServer(srv.Host, srv.Port, s.Stats.Connections)
		h, p := server.GetHostAndPort()
		err = lb.Start(ctx, h, p)
		Expect(err).To(BeNil())

		conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", h, p))
		Expect(err).To(BeNil())
		defer conn.Close()

		_, err = conn.Write([]byte("Hello"))
		Expect(err).To(BeNil())

		buff := make([]byte, 1024)
		n, err := conn.Read(buff)
		Expect(err).To(BeNil())
		Expect(string(buff[:n])).To(ContainSubstring("Acknowledged"))
	})

	It("should distribute connections randomly across multiple servers", func() {
		for i := 0; i < 3; i++ {
			s, err = startServer()
			Expect(err).To(BeNil())
			backendServers = append(backendServers, s)
		}

		lb = reservoir.NewReservoirLoadBalancer(ctx, 3)
		for _, s = range backendServers {
			lb.AddServer(s.Host, s.Port, s.Stats.Connections)
		}

		h, p := server.GetHostAndPort()
		err = lb.Start(ctx, h, p)
		Expect(err).To(BeNil())

		stopTime := time.Now().Add(10 * time.Second)
		connectionsSent := 0
		wg := sync.WaitGroup{}
		for time.Now().Before(stopTime) {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()

				conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", h, p))
				Expect(err).To(BeNil())
				defer conn.Close()

				_, err = conn.Write([]byte("Hello"))
				Expect(err).To(BeNil())

				buff := make([]byte, 1024)
				conn.SetReadDeadline(time.Now().Add(10 * time.Second))
				_, err = conn.Read(buff)
				log.Printf("Server %s:%d stats - Connections: %d, Added: %d, Removed: %d",
					s.Host, s.Port, s.Stats.Connections, s.Stats.ConnectionsAdded, s.Stats.ConnectionsRemoved)
				Expect(err).To(BeNil())
				Expect(string(buff)).To(ContainSubstring("Acknowledged"))
			}(connectionsSent)
			connectionsSent++
			time.Sleep(100 * time.Millisecond)
		}

		wg.Wait()
		log.Println("All connections processed, now stopping servers and load balancer.")

		totalConnections := 0
		for _, s := range backendServers {
			totalConnections += s.Stats.ConnectionsAdded
		}

		for _, s := range backendServers {
			log.Printf("Server %s:%d stats - Connections: %d, Added: %d, Removed: %d",
				s.Host, s.Port, s.Stats.Connections, s.Stats.ConnectionsAdded, s.Stats.ConnectionsRemoved)

			Expect(s.Stats.ConnectionsAdded).To(BeNumerically(">", 0))
		}
	})

})
