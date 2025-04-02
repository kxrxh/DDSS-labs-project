package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds the application configuration.
type Config struct {
	KafkaBrokers []string
	KafkaTopic   string
	NumCitizens  int
	SendInterval time.Duration
}

// LoadConfig loads configuration from environment variables or uses defaults.
func LoadConfig() (*Config, error) {
	cfg := &Config{
		KafkaBrokers: []string{"localhost:19092"}, // Default Redpanda port
		KafkaTopic:   "citizen_events",
		NumCitizens:  10,
		SendInterval: 2 * time.Second,
	}

	// Override with environment variables if set
	if brokers := os.Getenv("KAFKA_BROKERS"); brokers != "" {
		cfg.KafkaBrokers = strings.Split(brokers, ",")
	}

	if topic := os.Getenv("KAFKA_TOPIC"); topic != "" {
		cfg.KafkaTopic = topic
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

// --- Helper (kept private as only used here) ---

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

// getEnvSlice splits a comma-separated env var or returns fallback.
func getEnvSlice(key string, fallback []string) []string {
	if value, ok := os.LookupEnv(key); ok {
		return strings.Split(value, ",")
	}
	return fallback
}
