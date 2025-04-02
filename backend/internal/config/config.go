package config

import (
	"time"

	"github.com/kxrxh/ddss/internal/logger"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// Config holds all configuration for the application
type Config struct {
	Clickhouse struct {
		DSN string `mapstructure:"dsn"`
	} `mapstructure:"clickhouse"`
	Dgraph struct {
		Addresses []string `mapstructure:"addresses"`
	} `mapstructure:"dgraph"`
	Influx struct {
		URL    string `mapstructure:"url"`
		Token  string `mapstructure:"token"`
		Org    string `mapstructure:"org"`
		Bucket string `mapstructure:"bucket"`
	} `mapstructure:"influx"`
	Mongo struct {
		URI      string        `mapstructure:"uri"`
		Database string        `mapstructure:"database"`
		Timeout  time.Duration `mapstructure:"timeout"`
	} `mapstructure:"mongo"`
	JWT struct {
		Secret string `mapstructure:"secret"`
	} `mapstructure:"jwt"`
	Server struct {
		Port string `mapstructure:"port"`
	} `mapstructure:"server"`
	Logger struct {
		Level      string `mapstructure:"level"`
		Encoding   string `mapstructure:"encoding"`
		OutputPath string `mapstructure:"output_path"`
	} `mapstructure:"logger"`
	RedPanda RedPandaConfig `mapstructure:"redpanda"`
	Flink    FlinkConfig    `mapstructure:"flink"`
}

// Load reads the configuration from files and environment variables
func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// Load environment variables
	viper.AutomaticEnv()
	viper.SetEnvPrefix("DDSS")

	// Set defaults
	setDefaults()

	// Read config file if it exists
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		} else {
			logger.Log.Warn("No config file found, using defaults and environment variables", zap.Error(err))
		}
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

// setDefaults sets the default values for configuration
func setDefaults() {
	viper.SetDefault("clickhouse.dsn", "tcp://localhost:9000?database=default&username=default&read_timeout=10&write_timeout=20")
	viper.SetDefault("dgraph.addresses", []string{"localhost:9080"})
	viper.SetDefault("influx.url", "http://localhost:8086")
	viper.SetDefault("mongo.uri", "mongodb://localhost:27017")
	viper.SetDefault("mongo.database", "mydatabase")
	viper.SetDefault("mongo.timeout", "10s")
	viper.SetDefault("server.port", ":3000")

	// Logger defaults
	viper.SetDefault("logger.level", "info")
	viper.SetDefault("logger.encoding", "json")
	viper.SetDefault("logger.output_path", "stdout")

	// RedPanda defaults
	viper.SetDefault("redpanda.brokers", []string{"localhost:9092"})
	viper.SetDefault("redpanda.topic", "default-topic")
	viper.SetDefault("redpanda.group_id", "default-group")

	// Flink defaults
	viper.SetDefault("flink.jobmanager_host", "localhost")
	viper.SetDefault("flink.rest_port", 8081)
}
