// Command producer is the main entry point for the event producer service.
package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/kxrxh/social-rating-system/event-producer/internal/config"
	"github.com/kxrxh/social-rating-system/event-producer/internal/generator"
	"github.com/kxrxh/social-rating-system/event-producer/internal/producer"
)

func main() {
	log.Println("Starting event producer service...")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Configuration loaded: %+v", cfg)

	// Create event generator
	gen := generator.NewGenerator()
	log.Println("Event generator created.")

	// Create Kafka producer
	prod, err := producer.NewProducer(cfg, gen)
	if err != nil {
		log.Fatalf("Failed to create producer: %v", err)
	}
	defer prod.Close() // Ensure graceful shutdown of the producer

	// Setup signal handling for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop() // Ensure the context is cancelled eventually

	log.Println("Event producer running. Press Ctrl+C to stop.")

	// Run the producer loop
	prod.Run(ctx)

	// Context was cancelled (signal received), wait for cleanup
	log.Println("Shutdown signal received, initiating graceful shutdown...")

	// The defer prod.Close() will handle the Kafka client shutdown
	log.Println("Event producer shut down gracefully.")
}
