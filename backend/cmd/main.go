package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/kxrxh/ddss/internal/api"
	"github.com/kxrxh/ddss/internal/config"
	"github.com/kxrxh/ddss/internal/flink"
	"github.com/kxrxh/ddss/internal/logger"
	"github.com/kxrxh/ddss/internal/redpanda"
	"github.com/kxrxh/ddss/internal/repositories"
	"github.com/kxrxh/ddss/internal/repositories/clickhouse"
	"github.com/kxrxh/ddss/internal/repositories/dgraph"
	"github.com/kxrxh/ddss/internal/repositories/influx"
	"github.com/kxrxh/ddss/internal/repositories/mongo"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
)

func main() {
	log.Println("Starting application...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	if err := logger.Initialize(&logger.Config{
		Level:      cfg.Logger.Level,
		Encoding:   cfg.Logger.Encoding,
		OutputPath: cfg.Logger.OutputPath,
	}); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	defer logger.Sync()

	// Switch to zap logger
	log := logger.Log

	log.Info("Application starting...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel() // Ensure context is cancelled on exit

	// --- Initialize RedPanda Consumer ---
	log.Info("Initializing RedPanda consumer...")
	consumer, err := redpanda.NewConsumer(cfg.RedPanda)
	if err != nil {
		log.Fatal("Failed to initialize RedPanda consumer", zap.Error(err))
	}
	defer consumer.Close() // Ensure consumer is closed on shutdown
	log.Info("RedPanda consumer initialized")

	initializedRepos := make(map[string]repositories.Repository)
	var initErr error

	// --- Initialize Repositories ---
	log.Info("Initializing databases...")

	// ClickHouse
	initializedRepos["clickhouse"], initErr = clickhouse.New(cfg.Clickhouse.DSN)
	if initErr != nil {
		log.Fatal("Failed to initialize ClickHouse", zap.Error(initErr))
	}
	log.Info("ClickHouse connected")

	// Dgraph
	initializedRepos["dgraph"], initErr = dgraph.New(ctx, cfg.Dgraph.Addresses...)
	if initErr != nil {
		log.Fatal("Failed to initialize Dgraph", zap.Error(initErr))
	}
	log.Info("Dgraph connected")

	// InfluxDB
	initializedRepos["influx"], initErr = influx.New(ctx, influx.Config{
		URL:    cfg.Influx.URL,
		Token:  cfg.Influx.Token,
		Org:    cfg.Influx.Org,
		Bucket: cfg.Influx.Bucket,
	})
	if initErr != nil {
		log.Fatal("Failed to initialize InfluxDB", zap.Error(initErr))
	}
	log.Info("InfluxDB connected")

	// MongoDB
	initializedRepos["mongo"], initErr = mongo.New(ctx, mongo.Config{
		URI:      cfg.Mongo.URI,
		Database: cfg.Mongo.Database,
		Timeout:  cfg.Mongo.Timeout,
	})
	if initErr != nil {
		log.Fatal("Failed to initialize MongoDB", zap.Error(initErr))
	}
	log.Info("MongoDB connected")

	log.Info("All databases initialized successfully")

	// --- Initialize Flink Client ---
	log.Info("Initializing Flink client...")
	flinkClient, err := flink.NewClient(cfg.Flink)
	if err != nil {
		log.Fatal("Failed to initialize Flink client", zap.Error(err))
	}
	log.Info("Flink client initialized")

	// --- Setup and Start API Server ---
	app := api.NewServer(initializedRepos, flinkClient)

	go func() {
		if err := app.Listen(cfg.Server.Port); err != nil {
			log.Error("Error starting API server", zap.Error(err))
			cancel() // Trigger shutdown if server fails to start
		}
	}()
	log.Info("API server started", zap.String("port", cfg.Server.Port))

	// --- Start RedPanda Consumer ---
	// Define the message handler using the log
	handler := func(ctx context.Context, record *kgo.Record) error {
		log.Info("Received message from RedPanda",
			zap.String("topic", record.Topic),
			zap.Int32("partition", record.Partition),
			zap.Int64("offset", record.Offset),
			zap.ByteString("key", record.Key),
			zap.ByteString("value", record.Value),
		)
		// TODO: Implement actual message processing logic here
		return nil
	}

	// Start consuming in a separate goroutine
	go func() {
		log.Info("Starting RedPanda consumer loop...")
		err := consumer.StartConsuming(ctx, handler)
		if err != nil && !errors.Is(err, context.Canceled) {
			log.Error("RedPanda consumer stopped with error", zap.Error(err))
			cancel() // Trigger shutdown if consumer fails unexpectedly
		}
		log.Info("RedPanda consumer loop stopped.")
	}()

	// --- Graceful Shutdown ---
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Info("Application started successfully. Waiting for shutdown signal...")

	select {
	case sig := <-sigChan:
		log.Info("Received signal, shutting down...", zap.String("signal", sig.String()))
		cancel() // Trigger shutdown via context cancellation
	case <-ctx.Done(): // Handle shutdown triggered by server/consumer error or signal
		log.Info("Context cancelled, shutting down...")
	}

	// Wait for graceful shutdown procedures (API server, DBs)
	// The deferred consumer.Close() will be called automatically

	log.Info("Shutting down API server...")
	if err := app.Shutdown(); err != nil {
		log.Error("Error shutting down API server", zap.Error(err))
	}
	log.Info("API server shutdown complete")

	// Close database connections
	log.Info("Closing database connections...")
	for name, repo := range initializedRepos {
		if err := repo.Close(); err != nil {
			log.Error("Error closing connection",
				zap.String("database", name),
				zap.Error(err))
		} else {
			log.Info("Connection closed", zap.String("database", name))
		}
	}

	// Consumer is closed via defer
	log.Info("Shutdown complete")
}
