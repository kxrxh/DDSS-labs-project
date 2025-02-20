package db

import (
	"context"
	"fmt"

	"github.com/tikv/client-go/v2/txnkv"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// Client represents a composite database client that manages connections
// to multiple databases
type Client struct {
	graph *DgraphClient
	doc   *mongo.Client
	tikv  *txnkv.Client
}

// Config holds database connection configuration
type Config struct {
	DgraphAddr string
	MongoURI   string
	TiKVAddr   []string // TiKV accepts a list of PD addresses
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

	// Initialize TiKV client
	tikvClient, err := txnkv.NewClient(cfg.TiKVAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to create tikv client: %w", err)
	}

	return &Client{
		graph: dgraph,
		doc:   mongoClient,
		tikv:  tikvClient,
	}, nil
}

// Close closes all database connections
func (c *Client) Close() error {
	var errs []error
	if err := c.doc.Disconnect(context.Background()); err != nil {
		errs = append(errs, err)
	}
	if err := c.tikv.Close(); err != nil {
		errs = append(errs, err)
	}
	if len(errs) > 0 {
		return fmt.Errorf("errors closing connections: %v", errs)
	}
	return nil
}