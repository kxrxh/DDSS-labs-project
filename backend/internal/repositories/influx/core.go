package influx

import (
	"context"
	"fmt"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"

	"github.com/kxrxh/ddss/internal/repositories"
)

var _ repositories.Repository = (*Repository)(nil)

// Repository holds the InfluxDB client and provides access to its APIs.
type Repository struct {
	client       influxdb2.Client
	queryAPI     api.QueryAPI
	writeAPI     api.WriteAPIBlocking
	organization string
	bucket       string
}

// Config holds the necessary configuration for connecting to InfluxDB.
type Config struct {
	URL    string // InfluxDB server URL (e.g., "http://localhost:8086")
	Token  string // API token
	Org    string // Organization name
	Bucket string // Bucket name
}

// New creates and returns a new Repository instance connected to InfluxDB.
func New(ctx context.Context, cfg Config) (*Repository, error) {
	if cfg.URL == "" || cfg.Token == "" || cfg.Org == "" || cfg.Bucket == "" {
		return nil, fmt.Errorf("influxdb config is incomplete (URL, Token, Org, Bucket are required)")
	}

	// Create a new client using an InfluxDB server base URL and an authentication token
	client := influxdb2.NewClient(cfg.URL, cfg.Token)

	// Use default options for this example
	// client := influxdb2.NewClientWithOptions(cfg.URL, cfg.Token, influxdb2.DefaultOptions().SetLogLevel(3))

	// Check connectivity
	health, err := client.Health(ctx)
	if err != nil {
		client.Close() // Close client if health check fails
		return nil, fmt.Errorf("failed to check influxdb health: %w", err)
	}
	if health.Status != "pass" {
		client.Close()
		return nil, fmt.Errorf("influxdb health check failed: status %s, message %s", health.Status, *health.Message)
	}

	// Get query and write APIs
	queryAPI := client.QueryAPI(cfg.Org)
	writeAPI := client.WriteAPIBlocking(cfg.Org, cfg.Bucket)

	return &Repository{
		client:       client,
		queryAPI:     queryAPI,
		writeAPI:     writeAPI,
		organization: cfg.Org,
		bucket:       cfg.Bucket,
	}, nil
}

// Close closes the InfluxDB client connection.
func (r *Repository) Close() error {
	if r.client != nil {
		r.client.Close()
	}
	return nil
}

// QueryAPI returns the InfluxDB QueryAPI interface.
func (r *Repository) QueryAPI() api.QueryAPI {
	return r.queryAPI
}

// WriteAPI returns the InfluxDB WriteAPIBlocking interface.
func (r *Repository) WriteAPI() api.WriteAPIBlocking {
	return r.writeAPI
}

// Organization returns the configured organization name.
func (r *Repository) Organization() string {
	return r.organization
}

// Bucket returns the configured bucket name.
func (r *Repository) Bucket() string {
	return r.bucket
}
