package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/kxrxh/ddss/internal/api"
	"github.com/kxrxh/ddss/internal/db"
)

func main() {
	// Database configuration
	cfg := db.Config{
		DgraphAddr: "localhost:9080",
		MongoURI:   "mongodb://localhost:27017",
		TiKVAddr:   []string{"localhost:2379"},
	}

	// Initialize database client
	client, err := db.NewClient(cfg)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
	defer client.Close()

	// Create the REST API server
	server := api.NewServer(":8080", client)

	// Start the server in a goroutine
	go func() {
		if err := server.Start(); err != nil {
			log.Printf("Server error: %v\n", err)
		}
	}()

	// Handle graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Gracefully shutdown the server with a timeout of 5 seconds
	if err := server.Shutdown(); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	log.Println("Server exited properly")
}
