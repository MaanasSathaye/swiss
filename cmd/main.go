package main

// func main() {
// 	lb := leastconnections.NewLeastConnectionsLoadBalancer()

// 	serverCount := 3
// 	var wg sync.WaitGroup

// 	for i := 0; i < serverCount; i++ {
// 		host, port := server.GetHostAndPort()
// 		srvConfig := stats.ServerConfig{
// 			Host: host,
// 			Port: port,
// 		}
// 		srv, err := server.NewServer(srvConfig)
// 		if err != nil {
// 			log.Fatalf("Failed to create server: %v", err)
// 		}

// 		if err := lb.AddServer(srv); err != nil {
// 			log.Fatalf("Failed to add server: %v", err)
// 		}

// 		wg.Add(1)
// 		go func(s *server.Server) {
// 			defer wg.Done()
// 			if err := s.Start(); err != nil {
// 				log.Fatalf("Server start failed: %v", err)
// 			}
// 		}(srv)
// 	}

// 	go func() {
// 		if err := lb.StartBalancer("0.0.0.0", 8080); err != nil {
// 			log.Fatalf("Load balancer start failed: %v", err)
// 		}
// 	}()

// 	client := &http.Client{}

// 	for i := 0; i < 1000; i++ {
// 		simulateRequest(client, lb)
// 	}

// 	wg.Wait()
// }

// func simulateRequest(client *http.Client, lb *leastconnections.LeastConnectionsLoadBalancer) {
// 	req, err := requests.NewVariableRequest(context.Background(), "http://localhost:8080")
// 	if err != nil {
// 		log.Fatalf("Failed to create request: %v", err)
// 	}

// 	resp, err := client.Do(req)
// 	if err != nil {
// 		log.Fatalf("Request failed: %v", err)
// 	}
// 	resp.Body.Close()

// 	lb.Mutex.Lock()
// 	for _, srv := range lb.Servers {
// 		if rand.Float64() < 0.1 {
// 			srv.Stop()
// 			log.Printf("Server %s:%d went down", srv.Stats.Host, srv.Stats.Port)
// 		}
// 		// log.Printf("Server %s:%d - Connections: %d, ConnectionsAdded: %d, ConnectionsRemoved: %d", srv.Stats.Host, srv.Stats.Port, srv.Stats.Connections, srv.Stats.ConnectionsAdded, srv.Stats.ConnectionsRemoved)
// 		log.Printf("Server %s:%d - Connections: %d", srv.Stats.Host, srv.Stats.Port, srv.Stats.Connections)

// 	}
// 	lb.Mutex.Unlock()
// }
