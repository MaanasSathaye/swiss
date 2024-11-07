package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/MaanasSathaye/swiss/leastconnections"
	"github.com/MaanasSathaye/swiss/roundrobin"
	"github.com/MaanasSathaye/swiss/server"
	"github.com/gofrs/uuid/v5"
)

func rr(wg *sync.WaitGroup) {
	var (
		s              *server.Server
		err            error
		backendServers []*server.Server
		h              string
		p              int
		requestCount   int32
		ctx            context.Context
		cancel         context.CancelFunc
		start          time.Time
		duration       time.Duration
		conn           net.Conn
	)

	defer wg.Done()

	ctx, cancel = context.WithDeadline(context.Background(), time.Now().Add(10*time.Second))
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

	h, p = server.GetHostAndPort()
	go func() {
		err = lb.Start(ctx, h, p)
		if err != nil {
			log.Fatalf("Failed to start load balancer: %v", err)
		}
	}()

	time.Sleep(1 * time.Second)
	start = time.Now()
	done := make(chan struct{})
	maxRequests := 5000
	connLimit := make(chan struct{}, 100)

	go func() {
		for atomic.LoadInt32(&requestCount) < int32(maxRequests) {
			connLimit <- struct{}{}
			go func() {
				defer func() { <-connLimit }()
				for attempt := 0; attempt < 3; attempt++ {
					conn, err = net.DialTimeout("tcp", fmt.Sprintf("%s:%d", h, p), 5*time.Second)
					if err != nil {
						log.Printf("Unable to dial to server on attempt %d", attempt+1)
						time.Sleep(time.Duration(50*(1<<attempt)) * time.Millisecond) // exponential backoff
						continue
					}
					defer conn.Close()
					_, err = fmt.Fprintln(conn, "Hello from client")
					if err == nil {
						atomic.AddInt32(&requestCount, 1)
					}
					return
				}
			}()
		}
		for i := 0; i < cap(connLimit); i++ {
			connLimit <- struct{}{}
		}
		close(done)
	}()

	<-done
	duration = time.Since(start)

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
		h              string
		p              int
		requestCount   int32
		ctx            context.Context
		cancel         context.CancelFunc
		start          time.Time
		duration       time.Duration
		conn           net.Conn
	)

	defer wg.Done()

	ctx, cancel = context.WithDeadline(context.Background(), time.Now().Add(10*time.Second))
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

	h, p = server.GetHostAndPort()
	go func() {
		err = lb.Start(ctx, h, p)
		if err != nil {
			log.Fatalf("Failed to start load balancer: %v", err)
		}
	}()

	time.Sleep(1 * time.Second)
	start = time.Now()
	done := make(chan struct{})
	maxRequests := 5000
	connLimit := make(chan struct{}, 100)

	go func() {
		for atomic.LoadInt32(&requestCount) < int32(maxRequests) {
			connLimit <- struct{}{}
			go func() {
				defer func() { <-connLimit }()
				for attempt := 0; attempt < 3; attempt++ {
					conn, err = net.DialTimeout("tcp", fmt.Sprintf("%s:%d", h, p), 5*time.Second)
					if err != nil {
						log.Printf("Unable to dial to server on attempt %d", attempt+1)
						time.Sleep(time.Duration(50*(1<<attempt)) * time.Millisecond) // exponential backoff
						continue
					}
					defer conn.Close()
					_, err = fmt.Fprintln(conn, "Hello from client")
					if err == nil {
						atomic.AddInt32(&requestCount, 1)
					}
					return
				}
			}()
		}
		for i := 0; i < cap(connLimit); i++ {
			connLimit <- struct{}{}
		}
		close(done)
	}()

	<-done
	duration = time.Since(start)

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
		h              string
		p              int
		requestCount   int32
		ctx            context.Context
		cancel         context.CancelFunc
		start          time.Time
		duration       time.Duration
		conn           net.Conn
		key            uuid.UUID
	)

	defer wg.Done()

	ctx, cancel = context.WithDeadline(context.Background(), time.Now().Add(10*time.Second))
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

	h, p = server.GetHostAndPort()
	go func() {
		err = lb.Start(ctx, h, p)
		if err != nil {
			log.Fatalf("Failed to start load balancer: %v", err)
		}
	}()

	time.Sleep(1 * time.Second)
	start = time.Now()
	done := make(chan struct{})
	maxRequests := 5000
	connLimit := make(chan struct{}, 100)

	go func() {
		for atomic.LoadInt32(&requestCount) < int32(maxRequests) {
			connLimit <- struct{}{}
			go func() {
				defer func() { <-connLimit }()
				for attempt := 0; attempt < 3; attempt++ {
					conn, err = net.DialTimeout("tcp", fmt.Sprintf("%s:%d", h, p), 5*time.Second)
					if err != nil {
						log.Printf("Unable to dial to server on attempt %d", attempt+1)
						time.Sleep(time.Duration(50*(1<<attempt)) * time.Millisecond) // exponential backoff
						continue
					}
					defer conn.Close()
					if key, err = uuid.NewV4(); err != nil {
						log.Printf("unable to generate UUID")
						return
					}
					_, err = fmt.Fprintln(conn, "Hello from client", key)
					if err == nil {
						atomic.AddInt32(&requestCount, 1)
					}
					return
				}
			}()
		}
		for i := 0; i < cap(connLimit); i++ {
			connLimit <- struct{}{}
		}
		close(done)
	}()

	<-done
	duration = time.Since(start)

	log.Printf("Finished serving %d requests in %v", maxRequests, duration)
	for _, ss := range backendServers {
		fstats := ss.Stats
		log.Printf("Final stats (rendezvous hashing) %s:%d - Connections: %d, Added: %d, Removed: %d",
			ss.Host, ss.Port, fstats.Connections, fstats.ConnectionsAdded, fstats.ConnectionsRemoved)
	}
}

// func rs(wg *sync.WaitGroup) {
// 	var (
// 		s              *server.Server
// 		err            error
// 		backendServers []*server.Server
// 		conn           net.Conn
// 	)

// 	defer wg.Done()

// 	ctx, cancel := context.WithCancel(context.Background())
// 	defer cancel()

// 	startServer := func() (*server.Server, error) {
// 		s, err = server.NewServer(ctx)
// 		if err != nil {
// 			return nil, err
// 		}
// 		go s.Start(ctx)
// 		return s, nil
// 	}
// 	for i := 0; i < 3; i++ {
// 		s, err = startServer()
// 		backendServers = append(backendServers, s)
// 	}

// 	lb := reservoir.NewReservoirLoadBalancer(ctx, 3)
// 	for _, s = range backendServers {
// 		lb.AddServer(s.Host, s.Port, s.Stats.Connections)
// 	}

// 	h, p := server.GetHostAndPort()
// 	go func() {
// 		if err = lb.Start(ctx, h, p); err != nil {
// 			log.Fatalf("Failed to start load balancer: %v", err)
// 		}
// 	}()

// 	requestCount := 0
// 	maxRequests := 10000
// 	start := time.Now()
// 	done := make(chan struct{})

// 	go func() {
// 		if requestCount >= maxRequests {
// 			done <- struct{}{}
// 			return
// 		}

// 		go func() {
// 			var (
// 				attempt = 1
// 			)

// 			for {
// 				// Attempt to connect with retry logic
// 				conn, err = net.Dial("tcp", fmt.Sprintf("%s:%d", h, p))
// 				if err == nil {
// 					break
// 				}
// 				defer conn.Close()
// 				if attempt > 5 { // Max retry attempts
// 					log.Printf("Failed to connect after %d attempts: %v", attempt, err)
// 					return
// 				}
// 				backoff := time.Duration(1<<uint(attempt-1)) * time.Second
// 				log.Printf("Connection attempt %d failed, retrying in %s: %v", attempt, backoff, err)
// 				time.Sleep(backoff)
// 				attempt++
// 			}

// 			if _, err := fmt.Fprintln(conn, "Hello from client"); err != nil {
// 				log.Printf("Failed to send message: %v", err)
// 				return
// 			}
// 			requestCount++
// 		}()
// 	}()

// 	<-done
// 	duration := time.Since(start)
// 	log.Printf("Finished serving %d requests in %v", maxRequests, duration)
// 	for _, ss := range backendServers {
// 		fstats := ss.Stats
// 		log.Printf("Final stats (reservoir sampling) %s:%d - Connections: %d, Added: %d, Removed: %d",
// 			ss.Host, ss.Port, fstats.Connections, fstats.ConnectionsAdded, fstats.ConnectionsRemoved)
// 	}
// }

func main() {
	var (
		wg sync.WaitGroup
	)

	fmt.Println("Waiting for goroutines to finish...")

	wg.Add(3)
	go rr(&wg)
	go lc(&wg)
	go rh(&wg)
	// go rs(&wg)
	wg.Wait()
	fmt.Println("Done!")
}
