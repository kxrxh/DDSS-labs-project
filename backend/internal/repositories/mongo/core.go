package mongo

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/kxrxh/social-rating-system/internal/repositories"
)

var _ repositories.Repository = (*Repository)(nil)

// Repository holds the MongoDB client and database instance.
type Repository struct {
	client *mongo.Client
	db     *mongo.Database
}

// Config holds the necessary configuration for connecting to MongoDB.
type Config struct {
	URI      string        // MongoDB connection URI (e.g., "mongodb://localhost:27017")
	Database string        // Database name
	Timeout  time.Duration // Connection and operation timeout
}

// New creates and returns a new Repository instance connected to MongoDB.
func New(ctx context.Context, cfg Config) (*Repository, error) {
	if cfg.URI == "" || cfg.Database == "" {
		return nil, fmt.Errorf("mongo config is incomplete (URI and Database are required)")
	}

	timeout := cfg.Timeout
	if timeout == 0 {
		timeout = 10 * time.Second // Default timeout
	}

	// Set client options
	clientOptions := options.Client().ApplyURI(cfg.URI)
	clientOptions.SetServerSelectionTimeout(timeout) // Timeout for server selection
	clientOptions.SetConnectTimeout(timeout)         // Timeout for establishing connection
	clientOptions.SetTimeout(timeout)                // Timeout for operations

	// Connect to MongoDB
	ctxConnect, cancelConnect := context.WithTimeout(ctx, timeout)
	defer cancelConnect()
	client, err := mongo.Connect(ctxConnect, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mongo: %w", err)
	}

	// Ping the primary server to verify the connection
	ctxPing, cancelPing := context.WithTimeout(ctx, timeout)
	defer cancelPing()
	if err := client.Ping(ctxPing, readpref.Primary()); err != nil {
		client.Disconnect(ctx) // Disconnect if ping fails
		return nil, fmt.Errorf("failed to ping mongo: %w", err)
	}

	db := client.Database(cfg.Database)

	return &Repository{client: client, db: db}, nil
}

// Close disconnects the MongoDB client.
func (r *Repository) Close() error {
	if r.client != nil {
		// Use a background context for disconnection if the original context might be cancelled
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		return r.client.Disconnect(ctx)
	}
	return nil
}

// DB returns the underlying MongoDB database instance.
func (r *Repository) DB() *mongo.Database {
	return r.db
}

// Client returns the underlying MongoDB client instance.
func (r *Repository) Client() *mongo.Client {
	return r.client
}
