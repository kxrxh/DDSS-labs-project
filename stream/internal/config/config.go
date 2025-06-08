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
	InfluxRawEventsBucket          string
	InfluxDerivedMetricsBucket     string
	InfluxAnalyticalDatasetsBucket string

	// ScyllaDB configuration
	ScyllaHosts    []string
	ScyllaKeyspace string

	// Dgraph configuration
	DgraphHosts []string

	// MinIO configuration
	MinIOEndpoint  string
	MinIOAccessKey string
	MinIOSecretKey string
	MinIOBucket    string
	MinIOUseSSL    bool

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
		InfluxURL:                      "http://influxdb.influxdb.svc.cluster.local:8086",
		InfluxToken:                    "SJlclSgz9_7kcda4525xgHreow5f64j7GWZtzVM79LQ35KAtjy3y306Djomc5hgD25WvERI0qqvsaGZ0_VAx7w==",
		InfluxOrg:                      "social-rating",
		InfluxRawEventsBucket:          "raw_events",
		InfluxDerivedMetricsBucket:     "derived_metrics",
		InfluxAnalyticalDatasetsBucket: "analytical_datasets",

		// ScyllaDB defaults
		ScyllaHosts:    []string{"scylladb.scylladb.svc.cluster.local"},
		ScyllaKeyspace: "social_rating",

		// Dgraph defaults (based on Helm chart deployment: <release-name>-dgraph-alpha.<namespace>.svc.cluster.local)
		DgraphHosts: []string{"dgraph-dgraph-alpha.dgraph.svc.cluster.local:9080"},

		// MinIO defaults
		MinIOEndpoint:  "minio-service.minio.svc.cluster.local:9000",
		MinIOAccessKey: "admin",
		MinIOSecretKey: "password123",
		MinIOBucket:    "stream-backups",
		MinIOUseSSL:    false,

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

	if influxAnalyticalDatasetsBucket := os.Getenv("INFLUX_ANALYTICAL_DATASETS_BUCKET"); influxAnalyticalDatasetsBucket != "" {
		cfg.InfluxAnalyticalDatasetsBucket = influxAnalyticalDatasetsBucket
	}

	if scyllaHosts := os.Getenv("SCYLLA_HOSTS"); scyllaHosts != "" {
		cfg.ScyllaHosts = strings.Split(scyllaHosts, ",")
	}

	if scyllaKeyspace := os.Getenv("SCYLLA_KEYSPACE"); scyllaKeyspace != "" {
		cfg.ScyllaKeyspace = scyllaKeyspace
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

	// MinIO environment variables
	if minioEndpoint := os.Getenv("MINIO_ENDPOINT"); minioEndpoint != "" {
		cfg.MinIOEndpoint = minioEndpoint
	}

	if minioAccessKey := os.Getenv("MINIO_ACCESS_KEY"); minioAccessKey != "" {
		cfg.MinIOAccessKey = minioAccessKey
	}

	if minioSecretKey := os.Getenv("MINIO_SECRET_KEY"); minioSecretKey != "" {
		cfg.MinIOSecretKey = minioSecretKey
	}

	if minioBucket := os.Getenv("MINIO_BUCKET"); minioBucket != "" {
		cfg.MinIOBucket = minioBucket
	}

	if minioUseSSLStr := os.Getenv("MINIO_USE_SSL"); minioUseSSLStr != "" {
		if minioUseSSL, err := strconv.ParseBool(minioUseSSLStr); err == nil {
			cfg.MinIOUseSSL = minioUseSSL
		}
	}

	// Validate required configuration
	if cfg.InfluxToken == "" {
		return nil, fmt.Errorf("INFLUX_TOKEN is required")
	}

	return cfg, nil
}
