package roundrobin_test

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/MaanasSathaye/swiss/roundrobin"
	"github.com/MaanasSathaye/swiss/server"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("roundrobin.RRLoadBalancer", func() {
	var (
		lb             *roundrobin.RRLoadBalancer
		srv            *server.Server
		backendServers []*server.Server
		err            error
		s              *server.Server
	)
	ctx, done := context.WithTimeout(context.Background(), 15*time.Second)
	defer done()

	// Helper function to start a server on a dynamic port
	startServer := func() (*server.Server, error) {
		s, err = server.NewServer(ctx)
		if err != nil {
			return nil, err
		}
		go s.Start(ctx)
		time.Sleep(100 * time.Millisecond) // Allow server to start up
		return s, nil
	}

	AfterEach(func() {
		for _, s := range backendServers {
			s.Stop()
		}
		if lb != nil {
			lb.Stop()
		}
	})

	It("should start and stop the load balancer successfully", func() {
		lb = roundrobin.NewRRLoadBalancer(ctx)
		h, p := server.GetHostAndPort()
		err := lb.Start(ctx, h, p)
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

		lb = roundrobin.NewRRLoadBalancer(ctx)
		lb.AddServer(srv.Host, srv.Port)
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

	FIt("should round-robin connections across multiple servers", func() {
		var (
			err  error
			s    *server.Server
			conn net.Conn
			n    int
			wg   sync.WaitGroup
		)

		numConnections := 6
		wg.Add(numConnections)

		for i := 0; i < 3; i++ {
			s, err = startServer()
			Expect(err).To(BeNil())
			backendServers = append(backendServers, s)
		}

		lb = roundrobin.NewRRLoadBalancer(ctx)
		for _, s = range backendServers {
			lb.AddServer(s.Host, s.Port)
		}

		h, p := server.GetHostAndPort()
		err = lb.Start(ctx, h, p)
		Expect(err).To(BeNil())

		for i := 0; i < numConnections; i++ {
			go func(i int) {
				time.Sleep(5 * time.Second)
				defer wg.Done()

				log.Printf("Dialing connection %d\n", i)
				conn, err = net.DialTimeout("tcp", fmt.Sprintf("%s:%d", h, p), 5*time.Second)
				Expect(err).To(BeNil())
				defer conn.Close()

				log.Printf("Sending message from connection %d\n", i)
				_, err = conn.Write([]byte("Hello"))
				Expect(err).To(BeNil())

				// Set read deadline to avoid hanging
				conn.SetReadDeadline(time.Now().Add(5 * time.Second))

				buff := make([]byte, 1024)
				if n, err = conn.Read(buff); err != nil {
					log.Printf("Error on connection %d: %v\n", i, err)
				}
				Expect(err).To(BeNil())
				Expect(string(buff[:n])).To(ContainSubstring("Acknowledged"))

				log.Printf("Connection %d processed\n", i)
			}(i)
		}
		time.Sleep(10 * time.Second)
		wg.Wait()
		log.Println("All connections processed, now stopping servers and load balancer.")

		for _, s := range backendServers {
			log.Printf("Server %s:%d stats - Connections: %d, Added: %d, Removed: %d",
				s.Host, s.Port, s.Stats.Connections, s.Stats.ConnectionsAdded, s.Stats.ConnectionsRemoved)

			Expect(s.Stats.ConnectionsAdded).To(BeNumerically(">", 0))
		}
	})

})
