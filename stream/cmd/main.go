package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/kxrxh/social-rating-system/stream/internal/backup"
	"github.com/kxrxh/social-rating-system/stream/internal/config"
	"github.com/kxrxh/social-rating-system/stream/internal/database"
	"github.com/kxrxh/social-rating-system/stream/internal/health"
	"github.com/kxrxh/social-rating-system/stream/internal/pipeline"
	"github.com/kxrxh/social-rating-system/stream/internal/scoring"
	"github.com/kxrxh/social-rating-system/stream/internal/sinks"
)

func main() {
	log.Println("Starting Social Credit Stream Processor...")

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Setup signal handling for graceful shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Initialize database connections
	mongoClient, err := database.InitMongoDB(ctx, cfg)
	if err != nil {
		log.Fatalf("Failed to initialize MongoDB: %v", err)
	}
	defer mongoClient.Disconnect(ctx)

	influxClient := database.InitInfluxDB(cfg)
	defer influxClient.Close()

	scyllaSession, err := database.InitScyllaDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize ScyllaDB: %v", err)
	}
	defer scyllaSession.Close()

	// Initialize MinIO sink for backup operations
	minioSink, err := sinks.NewMinIOSink(
		cfg.MinIOEndpoint,
		cfg.MinIOAccessKey,
		cfg.MinIOSecretKey,
		cfg.MinIOBucket,
		cfg.MinIOUseSSL,
		log.Default(),
	)
	if err != nil {
		log.Fatalf("Failed to initialize MinIO sink: %v", err)
	}

	// Initialize backup manager
	backupManager := backup.NewManager(minioSink, log.Default())
	log.Println("Starting periodic backup scheduler...")
	backupManager.SchedulePeriodicBackups(ctx)

	// Perform initial system backup
	log.Println("Creating initial system backup...")
	if err := backupManager.BackupSystemSnapshot(ctx); err != nil {
		log.Printf("Warning: Initial system backup failed: %v", err)
	} else {
		log.Println("Initial system backup completed successfully")
	}

	// Initialize scoring engine with Dgraph URL from config
	dgraphURL := cfg.DgraphHosts[0] // Use first Dgraph host from config
	scoringEngine := scoring.NewEngine(mongoClient, cfg.MongoDatabase, dgraphURL)
	if err := scoringEngine.LoadConfiguration(ctx); err != nil {
		log.Fatalf("Failed to load scoring configuration: %v", err)
	}

	// Create Kafka consumer
	kafkaClient, err := database.CreateKafkaClient(cfg)
	if err != nil {
		log.Fatalf("Failed to create Kafka client: %v", err)
	}
	defer kafkaClient.Close()

	// Create and start stream processing pipeline
	streamPipeline := pipeline.New(kafkaClient, scoringEngine, influxClient, scyllaSession, minioSink, cfg)

	// Add backup manager to pipeline for metrics reporting
	streamPipeline.SetBackupManager(backupManager)

	// Start health check server
	healthServer := health.NewServer(":8080", streamPipeline)
	healthServer.SetBackupManager(backupManager)
	go healthServer.Start()

	log.Println("Stream processor running. Press Ctrl+C to stop.")

	// Start the pipeline
	log.Println("Starting pipeline...")
	go func() {
		log.Println("Pipeline Start() called")
		if err := streamPipeline.Start(ctx); err != nil {
			log.Printf("Pipeline error: %v", err)
		} else {
			log.Println("Pipeline Start() completed successfully")
		}
	}()

	// Wait for shutdown signal
	<-ctx.Done()
	log.Println("Shutdown signal received, stopping stream processor...")

	// Graceful shutdown with timeout
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := streamPipeline.Stop(shutdownCtx); err != nil {
		log.Printf("Error during shutdown: %v", err)
	}

	if err := healthServer.Stop(); err != nil {
		log.Printf("Error stopping health server: %v", err)
	}

	log.Println("Stream processor shut down gracefully.")
}
