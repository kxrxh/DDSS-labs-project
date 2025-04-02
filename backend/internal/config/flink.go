package config

// FlinkConfig holds the configuration for connecting to Flink REST API.
type FlinkConfig struct {
	JobManagerHost string `mapstructure:"jobmanager_host" yaml:"jobmanager_host"`
	RestPort       int    `mapstructure:"rest_port" yaml:"rest_port"`
	// Add other necessary fields like API version, timeout, auth if needed
}
