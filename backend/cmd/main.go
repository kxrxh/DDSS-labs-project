package main

import (
	"log"

	"github.com/kxrxh/ddss/internal/db"
)

func main() {
	// Create database client with local connection strings
	_, err := db.NewClient(db.Config{
		// DGraph typically uses port 9080 for gRPC
		DgraphAddr: "localhost:9080",
		// Standard MongoDB connection string
		MongoURI: "mongodb://localhost:27017",
		// TiKV PD (Placement Driver) endpoint
		TiKVAddr: []string{"localhost:2379"},
	})
	if err != nil {
		log.Fatalf("Failed to create database clients: %v", err)
	}

	log.Println("Successfully connected to databases")

	// inf wait here
	select {}
}
