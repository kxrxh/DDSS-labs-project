package db

import (
	"context"
	"fmt"
	"time"

	"github.com/tikv/client-go/v2/txnkv"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Client represents a database client with connections to multiple databases
type Client struct {
	doc        *mongo.Client // MongoDB client for document storage
	tikv       *txnkv.Client // TiKV client for key-value storage
	graph      *DgraphClient // Dgraph client for graph operations
	dgraphAddr string        // Address of the Dgraph server
}

// Config holds database connection configuration
type Config struct {
	DgraphAddr string
	MongoURI   string
	TiKVAddr   []string // TiKV accepts a list of PD addresses
}

// ConnectMongoDB establishes a connection to MongoDB
func ConnectMongoDB(uri string) (*mongo.Client, error) {
	if uri == "" {
		return nil, fmt.Errorf("MongoDB URI is required")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to create MongoDB client: %w", err)
	}

	// Ping the database to verify connection
	if err := client.Ping(ctx, readpref.Primary()); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	return client, nil
}

// ConnectTiKV establishes a connection to TiKV
func ConnectTiKV(pdAddr string) (*txnkv.Client, error) {
	if pdAddr == "" {
		return nil, fmt.Errorf("TiKV PD address is required")
	}

	client, err := txnkv.NewClient([]string{pdAddr})
	if err != nil {
		return nil, fmt.Errorf("failed to create TiKV client: %w", err)
	}

	return client, nil
}

// NewClient creates a new client with connections to all databases
func NewClient(config Config) (*Client, error) {
	doc, err := ConnectMongoDB(config.MongoURI)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %v", err)
	}

	// Use first TiKV address if available
	tikvPD := ""
	if len(config.TiKVAddr) > 0 {
		tikvPD = config.TiKVAddr[0]
	}

	tikv, err := ConnectTiKV(tikvPD)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to TiKV: %v", err)
	}

	graph, err := NewDgraphClient(config.DgraphAddr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Dgraph: %v", err)
	}

	return &Client{
		doc:        doc,
		tikv:       tikv,
		graph:      graph,
		dgraphAddr: config.DgraphAddr,
	}, nil
}

// Close closes all database connections
func (c *Client) Close() error {
	var errs []error

	if c.doc != nil {
		if err := c.doc.Disconnect(context.Background()); err != nil {
			errs = append(errs, fmt.Errorf("failed to disconnect MongoDB: %v", err))
		}
	}

	if c.tikv != nil {
		if err := c.tikv.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close TiKV connection: %v", err))
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors when closing database connections: %v", errs)
	}

	return nil
}
