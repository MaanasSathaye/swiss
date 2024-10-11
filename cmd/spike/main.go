package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/MaanasSathaye/swiss/server"
	"github.com/MaanasSathaye/swiss/stats"
)

func main() {
	host, port := server.GetHostAndPort()
	mockStats := stats.ServerConfig{
		Id:        "test-server",
		Host:      host,
		Port:      port,
		UpdatedAt: time.Now(),
		Load:      0.0,
	}

	server, err := server.NewServer(nil, mockStats)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	if err := server.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	log.Println("Server started successfully", server.Stats, "Initial server connections count:", server.Stats.Connections)

	go func() {
		time.Sleep(1 * time.Second)

		url := fmt.Sprintf("http://%s:%d", host, port)
		resp, err := http.Get(url)
		if err != nil {
			log.Fatalf("Failed to send request: %v", err)
		}
		defer resp.Body.Close()

		log.Println("Response from server:", resp.Status)
	}()
	time.Sleep(3 * time.Second)
	server.Stop()
	log.Printf("Server stopped. Final server stats: %v", server.Stats)
}
