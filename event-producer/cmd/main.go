package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kxrxh/social-rating-system/event-producer/internal/config"
	"github.com/kxrxh/social-rating-system/event-producer/internal/generator"
	"github.com/kxrxh/social-rating-system/event-producer/internal/producer"
)

func main() {
	// Create a context that will be cancelled on SIGINT/SIGTERM
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Received shutdown signal, initiating graceful shutdown...")
		cancel()
	}()

	// Initialize configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Configuration loaded: %+v", cfg)

	// Create event generator
	gen := generator.NewGenerator()
	log.Println("Event generator created.")

	// Create and initialize producer
	prod, err := producer.NewProducer(cfg, gen)
	if err != nil {
		log.Fatalf("Failed to create producer: %v", err)
	}
	defer prod.Close()

	// Start event production loop
	prod.Run(ctx)

	// Wait for context cancellation (from signal handler)
	<-ctx.Done()

	// Give some time for in-flight messages to be processed
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	<-shutdownCtx.Done()
	log.Println("Shutdown complete.")
}
