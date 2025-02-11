package db

import (
	"context"
	"database/sql"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Client represents a composite database client that manages connections
// to multiple databases
type Client struct {
	graph *DgraphClient
	doc   *mongo.Client
	sql   *sql.DB
}

// Config holds database connection configuration
type Config struct {
	DgraphAddr  string
	MongoURI    string
	PostgresURI string
}

// NewClient creates a new composite database client
func NewClient(cfg Config) (*Client, error) {
	// Initialize Dgraph client
	dgraph, err := NewDgraphClient(cfg.DgraphAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create dgraph client: %w", err)
	}

	// Initialize MongoDB client
	mongoClient, err := mongo.Connect(options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		return nil, fmt.Errorf("failed to create mongo client: %w", err)
	}

	// Initialize Postgres client
	pgClient, err := sql.Open("postgres", cfg.PostgresURI)
	if err != nil {
		return nil, fmt.Errorf("failed to create postgres client: %w", err)
	}

	return &Client{
		graph: dgraph,
		doc:   mongoClient,
		sql:   pgClient,
	}, nil
}

// Close closes all database connections
func (c *Client) Close() error {
	var errs []error
	if err := c.doc.Disconnect(context.Background()); err != nil {
		errs = append(errs, err)
	}
	if err := c.sql.Close(); err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return fmt.Errorf("errors closing connections: %v", errs)
	}
	return nil
}
