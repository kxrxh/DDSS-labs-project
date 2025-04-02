package config

// RedPandaConfig holds the configuration for RedPanda connection.
type RedPandaConfig struct {
	Brokers []string `yaml:"brokers"`
	Topic   string   `yaml:"topic"`
	GroupID string   `yaml:"group_id"`
	// Add other necessary configuration fields like SASL, TLS etc. if needed
}
