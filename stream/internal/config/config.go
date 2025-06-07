package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all configuration for the stream processor
type Config struct {
	// Kafka/Redpanda configuration
	KafkaBrokers []string
	KafkaTopic   string
	KafkaGroupID string

	// MongoDB configuration
	MongoURI      string
	MongoDatabase string

	// InfluxDB configuration
	InfluxURL   string
	InfluxToken string
	InfluxOrg   string
	// Specific buckets for different data types
	InfluxRawEventsBucket      string
	InfluxDerivedMetricsBucket string

	// ScyllaDB configuration
	ScyllaHosts    []string
	ScyllaKeyspace string

	// DuckDB configuration
	DuckDBPath string

	// Dgraph configuration
	DgraphHosts []string

	// Processing configuration
	BatchSize     int
	FlushInterval time.Duration
	Workers       int
}

// LoadConfig loads configuration from environment variables with sensible defaults
func LoadConfig() (*Config, error) {
	cfg := &Config{
		// Kafka defaults
		KafkaBrokers: []string{"localhost:19092"},
		KafkaTopic:   "citizen_events",
		KafkaGroupID: "social-credit-processor",

		// MongoDB defaults
		MongoURI:      "mongodb://mongodb-service.mongodb.svc.cluster.local:27017",
		MongoDatabase: "social_rating",

		// InfluxDB defaults
		InfluxURL:                  "http://influxdb.influxdb.svc.cluster.local:8086",
		InfluxToken:                "SJlclSgz9_7kcda4525xgHreow5f64j7GWZtzVM79LQ35KAtjy3y306Djomc5hgD25WvERI0qqvsaGZ0_VAx7w==",
		InfluxOrg:                  "social-rating",
		InfluxRawEventsBucket:      "raw_events",
		InfluxDerivedMetricsBucket: "derived_metrics",

		// ScyllaDB defaults
		ScyllaHosts:    []string{"scylladb.scylladb.svc.cluster.local"},
		ScyllaKeyspace: "social_rating",

		// DuckDB defaults
		DuckDBPath: "/tmp/social_rating_analytics.db",

		// Dgraph defaults (based on Helm chart deployment: <release-name>-dgraph-alpha.<namespace>.svc.cluster.local)
		DgraphHosts: []string{"dgraph-dgraph-alpha.dgraph.svc.cluster.local:9080"},

		// Processing defaults
		BatchSize:     1000,
		FlushInterval: 1 * time.Second,
		Workers:       12,
	}

	// Override with environment variables
	if brokers := os.Getenv("KAFKA_BROKERS"); brokers != "" {
		cfg.KafkaBrokers = strings.Split(brokers, ",")
	}

	if topic := os.Getenv("KAFKA_TOPIC"); topic != "" {
		cfg.KafkaTopic = topic
	}

	if groupID := os.Getenv("KAFKA_GROUP_ID"); groupID != "" {
		cfg.KafkaGroupID = groupID
	}

	if mongoURI := os.Getenv("MONGO_URI"); mongoURI != "" {
		cfg.MongoURI = mongoURI
	}

	if mongoDB := os.Getenv("MONGO_DATABASE"); mongoDB != "" {
		cfg.MongoDatabase = mongoDB
	}

	if influxURL := os.Getenv("INFLUX_URL"); influxURL != "" {
		cfg.InfluxURL = influxURL
	}

	if influxToken := os.Getenv("INFLUX_TOKEN"); influxToken != "" {
		cfg.InfluxToken = influxToken
	}

	if influxOrg := os.Getenv("INFLUX_ORG"); influxOrg != "" {
		cfg.InfluxOrg = influxOrg
	}

	if influxRawEventsBucket := os.Getenv("INFLUX_RAW_EVENTS_BUCKET"); influxRawEventsBucket != "" {
		cfg.InfluxRawEventsBucket = influxRawEventsBucket
	}

	if influxDerivedMetricsBucket := os.Getenv("INFLUX_DERIVED_METRICS_BUCKET"); influxDerivedMetricsBucket != "" {
		cfg.InfluxDerivedMetricsBucket = influxDerivedMetricsBucket
	}

	if scyllaHosts := os.Getenv("SCYLLA_HOSTS"); scyllaHosts != "" {
		cfg.ScyllaHosts = strings.Split(scyllaHosts, ",")
	}

	if scyllaKeyspace := os.Getenv("SCYLLA_KEYSPACE"); scyllaKeyspace != "" {
		cfg.ScyllaKeyspace = scyllaKeyspace
	}

	if duckDBPath := os.Getenv("DUCKDB_PATH"); duckDBPath != "" {
		cfg.DuckDBPath = duckDBPath
	}

	if dgraphHosts := os.Getenv("DGRAPH_HOSTS"); dgraphHosts != "" {
		cfg.DgraphHosts = strings.Split(dgraphHosts, ",")
	}

	if batchSizeStr := os.Getenv("BATCH_SIZE"); batchSizeStr != "" {
		if batchSize, err := strconv.Atoi(batchSizeStr); err == nil && batchSize > 0 {
			cfg.BatchSize = batchSize
		}
	}

	if flushIntervalStr := os.Getenv("FLUSH_INTERVAL"); flushIntervalStr != "" {
		if flushInterval, err := time.ParseDuration(flushIntervalStr); err == nil {
			cfg.FlushInterval = flushInterval
		}
	}

	if workersStr := os.Getenv("WORKERS"); workersStr != "" {
		if workers, err := strconv.Atoi(workersStr); err == nil && workers > 0 {
			cfg.Workers = workers
		}
	}

	// Validate required configuration
	if cfg.InfluxToken == "" {
		return nil, fmt.Errorf("INFLUX_TOKEN is required")
	}

	return cfg, nil
}
