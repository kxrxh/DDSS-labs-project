package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds all configuration for the event producer.
type Config struct {
	// Kafka configuration
	KafkaBrokers []string
	KafkaTopic   string

	// MongoDB configuration
	MongoURI string

	// Dgraph configuration
	DgraphURL string

	// Event generation configuration
	SendInterval time.Duration
	NumCitizens  int
}

// NewDefaultConfig creates a new Config with default values.
func NewDefaultConfig() *Config {
	return &Config{
		KafkaBrokers: []string{"localhost:9092"},
		KafkaTopic:   "social-rating-events",
		MongoURI:     "mongodb://mongodb-service.mongodb.svc.cluster.local:27017",
		SendInterval: time.Second * 2,
		NumCitizens:  25,
	}
}

// LoadConfig loads configuration from environment variables or uses defaults.
func LoadConfig() (*Config, error) {
	cfg := &Config{
		KafkaBrokers: []string{"localhost:19092"}, // Default Redpanda port
		KafkaTopic:   "citizen_events",
		MongoURI:     "mongodb://mongodb-service.mongodb.svc.cluster.local:27017", // Default MongoDB URI
		DgraphURL:    "dgraph-dgraph-alpha.dgraph.svc.cluster.local:9080",         // Default Dgraph URL (Helm chart)
		NumCitizens:  25,
		SendInterval: 2 * time.Second,
	}

	// Override with environment variables if set
	if brokers := os.Getenv("KAFKA_BROKERS"); brokers != "" {
		cfg.KafkaBrokers = strings.Split(brokers, ",")
	}

	if topic := os.Getenv("KAFKA_TOPIC"); topic != "" {
		cfg.KafkaTopic = topic
	}

	if mongoURI := os.Getenv("MONGO_URI"); mongoURI != "" {
		cfg.MongoURI = mongoURI
	}

	if dgraphURL := os.Getenv("DGRAPH_URL"); dgraphURL != "" {
		cfg.DgraphURL = dgraphURL
	}

	if numCitizensStr := os.Getenv("NUM_CITIZENS"); numCitizensStr != "" {
		num, err := strconv.Atoi(numCitizensStr)
		if err != nil {
			return nil, fmt.Errorf("invalid NUM_CITIZENS value: %w", err)
		}
		if num > 0 {
			cfg.NumCitizens = num
		}
	}

	if intervalStr := os.Getenv("SEND_INTERVAL"); intervalStr != "" {
		duration, err := time.ParseDuration(intervalStr) // e.g., "5s", "100ms"
		if err != nil {
			return nil, fmt.Errorf("invalid SEND_INTERVAL value: %w", err)
		}
		if duration > 0 {
			cfg.SendInterval = duration
		}
	}

	return cfg, nil
}