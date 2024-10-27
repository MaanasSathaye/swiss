package rendezvous_test

import (
	"context"
	"crypto/rand"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/MaanasSathaye/swiss/rendezvous"
	"github.com/MaanasSathaye/swiss/server"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("rendezvous.RendezvousLoadBalancer", func() {
	var (
		lb             *rendezvous.RendezvousHashingLoadBalancer
		srv            *server.Server
		backendServers []*server.Server
		err            error
		s              *server.Server
	)

	ctx, done := context.WithTimeout(context.Background(), 15*time.Second)
	defer done()

	// Helper function to start a server with a dynamic port
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
		lb = rendezvous.NewRendezvousHashingLoadBalancer(ctx)
		h, p := server.GetHostAndPort()
		err = lb.Start(ctx, h, p)
		Expect(err).To(BeNil())

		time.Sleep(1 * time.Second)
	})

	It("should start a server and handle a connection via load balancer", func() {
		var (
			err  error
			conn net.Conn
			n    int
		)
		srv, err = startServer()
		Expect(err).To(BeNil())

		lb = rendezvous.NewRendezvousHashingLoadBalancer(ctx)
		lb.AddServer(srv.Host, srv.Port)
		h, p := server.GetHostAndPort()
		err = lb.Start(ctx, h, p)
		Expect(err).To(BeNil())

		conn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", h, p))
		Expect(err).To(BeNil())
		defer conn.Close()

		key := make([]byte, 16)
		_, err = conn.Write(key)
		Expect(err).To(BeNil())

		_, err = conn.Write([]byte("Hello"))
		Expect(err).To(BeNil())

		conn.SetReadDeadline(time.Now().Add(5 * time.Second))
		buff := make([]byte, 1024)
		n, err = conn.Read(buff)
		Expect(err).To(BeNil())
		Expect(string(buff[:n])).To(ContainSubstring("Acknowledged"))
	})

	FIt("should distribute connections based on Rendezvous hashing across multiple servers", func() {
		var (
			err  error
			conn net.Conn
			wg   sync.WaitGroup
		)

		lb = rendezvous.NewRendezvousHashingLoadBalancer(ctx)

		for i := 0; i < 3; i++ {
			s, err = startServer()
			Expect(err).To(BeNil())
			backendServers = append(backendServers, s)
			lb.AddServer(s.Host, s.Port)
		}

		h, p := server.GetHostAndPort()
		err = lb.Start(ctx, h, p)
		Expect(err).To(BeNil())

		stopTime := time.Now().Add(30 * time.Second)
		connectionsSent := 0
		for time.Now().Before(stopTime) {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()

				conn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", h, p))
				Expect(err).To(BeNil())
				defer conn.Close()

				key := make([]byte, 16)
				_, err = rand.Read(key) // Generate a unique key
				Expect(err).To(BeNil())
				_, err = conn.Write(key)
				Expect(err).To(BeNil())

				_, err = conn.Write([]byte("Hello"))
				Expect(err).To(BeNil())

				buff := make([]byte, 1024)
				conn.SetReadDeadline(time.Now().Add(5 * time.Second))
				_, err = conn.Read(buff)
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

		avgConnections := totalConnections / len(backendServers)
		for _, s := range backendServers {
			log.Printf("Server %s:%d stats - Connections: %d, Added: %d, Removed: %d",
				s.Host, s.Port, s.Stats.Connections, s.Stats.ConnectionsAdded, s.Stats.ConnectionsRemoved)

			Expect(s.Stats.ConnectionsAdded).To(BeNumerically("~", avgConnections, 10))
		}
	})
})
