package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/MaanasSathaye/swiss/leastconnections"
	"github.com/MaanasSathaye/swiss/rendezvous"
	"github.com/MaanasSathaye/swiss/reservoir"
	"github.com/MaanasSathaye/swiss/roundrobin"
	"github.com/MaanasSathaye/swiss/server"
	"github.com/gofrs/uuid/v5"
)

func rr(wg *sync.WaitGroup) {
	var (
		s              *server.Server
		err            error
		backendServers []*server.Server
		conn           net.Conn
	)

	defer wg.Done()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	startServer := func() (*server.Server, error) {
		s, err = server.NewServer(ctx)
		if err != nil {
			return nil, err
		}
		go s.Start(ctx)
		return s, nil
	}
	for i := 0; i < 3; i++ {
		s, err = startServer()
		backendServers = append(backendServers, s)
	}

	lb := roundrobin.NewRRLoadBalancer(ctx)
	for _, s = range backendServers {
		lb.AddServer(s.Host, s.Port)
	}

	h, p := server.GetHostAndPort()
	go func() {
		if err = lb.Start(ctx, h, p); err != nil {
			log.Fatalf("Failed to start load balancer: %v", err)
		}
	}()

	requestCount := 0
	requestRate := 10
	maxRequests := 1000
	start := time.Now()
	tick := time.Tick(time.Second / time.Duration(requestRate))
	minuteTicker := time.NewTicker(time.Minute)
	defer minuteTicker.Stop()
	done := make(chan struct{})

	log.Println("Starting load test with 10 requests/second.")

	go func() {
		for {
			select {
			case <-tick:
				if requestCount >= maxRequests {
					done <- struct{}{}
					return
				}

				go func() {
					conn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", h, p))
					if err != nil {
						log.Printf("Connection error: %v", err)
						return
					}
					defer conn.Close()

					fmt.Fprintln(conn, "Hello from client")
					requestCount++
				}()

			case <-minuteTicker.C:
				log.Printf("Minute elapsed - requests served: %d", requestCount)
				for _, server := range backendServers {
					tstats := server.Stats
					log.Printf("Server %s:%d - Connections: %d, Added: %d, Removed: %d",
						server.Host, server.Port, tstats.Connections, tstats.ConnectionsAdded, tstats.ConnectionsRemoved)
				}
			}
		}
	}()

	<-done
	duration := time.Since(start)
	log.Printf("Finished serving %d requests in %v", maxRequests, duration)
	for _, ss := range backendServers {
		fstats := ss.Stats
		log.Printf("Final stats (round robin) %s:%d - Connections: %d, Added: %d, Removed: %d",
			ss.Host, ss.Port, fstats.Connections, fstats.ConnectionsAdded, fstats.ConnectionsRemoved)
	}
}

func lc(wg *sync.WaitGroup) {
	var (
		s              *server.Server
		err            error
		backendServers []*server.Server
		conn           net.Conn
	)

	defer wg.Done()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	startServer := func() (*server.Server, error) {
		s, err = server.NewServer(ctx)
		if err != nil {
			return nil, err
		}
		go s.Start(ctx)
		return s, nil
	}
	for i := 0; i < 3; i++ {
		s, err = startServer()
		backendServers = append(backendServers, s)
	}

	lb := leastconnections.NewLeastConnectionsLoadBalancer(ctx)
	for _, s = range backendServers {
		lb.AddServer(s.Host, s.Port, s.Stats.Connections)
	}

	h, p := server.GetHostAndPort()
	go func() {
		if err = lb.Start(ctx, h, p); err != nil {
			log.Fatalf("Failed to start load balancer: %v", err)
		}
	}()

	requestCount := 0
	requestRate := 10
	maxRequests := 1000
	start := time.Now()
	tick := time.Tick(time.Second / time.Duration(requestRate))
	minuteTicker := time.NewTicker(time.Minute)
	defer minuteTicker.Stop()
	done := make(chan struct{})

	log.Println("Starting load test with 10 requests/second.")

	go func() {
		for {
			select {
			case <-tick:
				if requestCount >= maxRequests {
					done <- struct{}{}
					return
				}

				go func() {
					conn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", h, p))
					if err != nil {
						log.Printf("Connection error: %v", err)
						return
					}
					defer conn.Close()

					fmt.Fprintln(conn, "Hello from client")
					requestCount++
				}()

			case <-minuteTicker.C:
				log.Printf("Minute elapsed - requests served: %d", requestCount)
				for _, server := range backendServers {
					tstats := server.Stats
					log.Printf("Server %s:%d - Connections: %d, Added: %d, Removed: %d",
						server.Host, server.Port, tstats.Connections, tstats.ConnectionsAdded, tstats.ConnectionsRemoved)
				}
			}
		}
	}()

	<-done
	duration := time.Since(start)
	log.Printf("Finished serving %d requests in %v", maxRequests, duration)
	for _, ss := range backendServers {
		fstats := ss.Stats
		log.Printf("Final stats (least connections) %s:%d - Connections: %d, Added: %d, Removed: %d",
			ss.Host, ss.Port, fstats.Connections, fstats.ConnectionsAdded, fstats.ConnectionsRemoved)
	}
}

func rh(wg *sync.WaitGroup) {
	var (
		s              *server.Server
		err            error
		backendServers []*server.Server
		conn           net.Conn
		key            uuid.UUID
	)

	defer wg.Done()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	startServer := func() (*server.Server, error) {
		s, err = server.NewServer(ctx)
		if err != nil {
			return nil, err
		}
		go s.Start(ctx)
		return s, nil
	}
	for i := 0; i < 3; i++ {
		s, err = startServer()
		backendServers = append(backendServers, s)
	}

	lb := rendezvous.NewRendezvousHashingLoadBalancer(ctx)
	for _, s = range backendServers {
		lb.AddServer(s.Host, s.Port)
	}

	h, p := server.GetHostAndPort()
	go func() {
		if err = lb.Start(ctx, h, p); err != nil {
			log.Fatalf("Failed to start load balancer: %v", err)
		}
	}()

	requestCount := 0
	requestRate := 10
	maxRequests := 1000
	start := time.Now()
	tick := time.Tick(time.Second / time.Duration(requestRate))
	minuteTicker := time.NewTicker(time.Minute)
	defer minuteTicker.Stop()
	done := make(chan struct{})

	log.Println("Starting load test with 10 requests/second.")

	go func() {
		for {
			select {
			case <-tick:
				if requestCount >= maxRequests {
					done <- struct{}{}
					return
				}

				go func() {
					conn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", h, p))
					if err != nil {
						log.Printf("Connection error: %v", err)
						return
					}
					defer conn.Close()

					if key, err = uuid.NewV4(); err != nil {
						log.Printf("Keygen error: %v", err)
						return
					}
					fmt.Fprintln(conn, key) // Send the unique key to the load balancer

					fmt.Fprintln(conn, "Hello from client")
					requestCount++
				}()

			case <-minuteTicker.C:
				log.Printf("Minute elapsed - requests served: %d", requestCount)
				for _, server := range backendServers {
					tstats := server.Stats
					log.Printf("Server %s:%d - Connections: %d, Added: %d, Removed: %d",
						server.Host, server.Port, tstats.Connections, tstats.ConnectionsAdded, tstats.ConnectionsRemoved)
				}
			}
		}
	}()

	<-done
	duration := time.Since(start)
	log.Printf("Finished serving %d requests in %v", maxRequests, duration)
	for _, ss := range backendServers {
		fstats := ss.Stats
		log.Printf("Final stats (rendezvous hashing) %s:%d - Connections: %d, Added: %d, Removed: %d",
			ss.Host, ss.Port, fstats.Connections, fstats.ConnectionsAdded, fstats.ConnectionsRemoved)
	}
}

func rs(wg *sync.WaitGroup) {
	var (
		s              *server.Server
		err            error
		backendServers []*server.Server
		conn           net.Conn
	)

	defer wg.Done()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	startServer := func() (*server.Server, error) {
		s, err = server.NewServer(ctx)
		if err != nil {
			return nil, err
		}
		go s.Start(ctx)
		return s, nil
	}
	for i := 0; i < 3; i++ {
		s, err = startServer()
		backendServers = append(backendServers, s)
	}

	lb := reservoir.NewReservoirLoadBalancer(ctx)
	for _, s = range backendServers {
		lb.AddServer(s.Host, s.Port, s.Stats.Connections)
	}

	h, p := server.GetHostAndPort()
	go func() {
		if err = lb.Start(ctx, h, p); err != nil {
			log.Fatalf("Failed to start load balancer: %v", err)
		}
	}()

	requestCount := 0
	requestRate := 10
	maxRequests := 1000
	start := time.Now()
	tick := time.Tick(time.Second / time.Duration(requestRate))
	minuteTicker := time.NewTicker(time.Minute)
	defer minuteTicker.Stop()
	done := make(chan struct{})

	log.Println("Starting load test with 10 requests/second.")

	go func() {
		for {
			select {
			case <-tick:
				if requestCount >= maxRequests {
					done <- struct{}{}
					return
				}

				go func() {
					conn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", h, p))
					if err != nil {
						log.Printf("Connection error: %v", err)
						return
					}
					defer conn.Close()

					fmt.Fprintln(conn, "Hello from client")
					requestCount++
				}()

			case <-minuteTicker.C:
				log.Printf("Minute elapsed - requests served: %d", requestCount)
				for _, server := range backendServers {
					tstats := server.Stats
					log.Printf("Server %s:%d - Connections: %d, Added: %d, Removed: %d",
						server.Host, server.Port, tstats.Connections, tstats.ConnectionsAdded, tstats.ConnectionsRemoved)
				}
			}
		}
	}()

	<-done
	duration := time.Since(start)
	log.Printf("Finished serving %d requests in %v", maxRequests, duration)
	for _, ss := range backendServers {
		fstats := ss.Stats
		log.Printf("Final stats (reservoir sampling) %s:%d - Connections: %d, Added: %d, Removed: %d",
			ss.Host, ss.Port, fstats.Connections, fstats.ConnectionsAdded, fstats.ConnectionsRemoved)
	}
}

func main() {
	var (
		wg sync.WaitGroup
	)

	wg.Add(4)
	go rr(&wg)
	go lc(&wg)
	go rh(&wg)
	go rs(&wg)

	fmt.Println("Waiting for goroutines to finish...")
	wg.Wait()
	fmt.Println("Done!")
}
