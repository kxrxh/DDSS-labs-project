package redpanda

import (
	"context"

	"github.com/kxrxh/social-rating-system/internal/config"
	"github.com/kxrxh/social-rating-system/internal/logger"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
)

// Producer handles sending messages to Kafka/RedPanda using franz-go.
type Producer struct {
	client *kgo.Client
	topic  string
}

// NewProducer creates a new franz-go Kafka producer instance.
func NewProducer(cfg config.RedPandaConfig) (*Producer, error) {
	opts := []kgo.Opt{
		kgo.SeedBrokers(cfg.Brokers...),
		kgo.DefaultProduceTopic(cfg.Topic),
		// Add other client options if needed (e.g., SASL, TLS)
		// kgo.SASL(scram.Auth{User: "user", Pass: "pass"}.AsSha256Mechanism()),
		// kgo.Dialer(tlsDialer),
	}

	client, err := kgo.NewClient(opts...)
	if err != nil {
		logger.Log.Error("Failed to create franz-go client", zap.Error(err))
		return nil, err
	}

	// Optional: Ping the cluster to ensure connectivity
	if err := client.Ping(context.Background()); err != nil {
		logger.Log.Error("Failed to ping Kafka cluster", zap.Error(err))
		client.Close() // Close the client if ping fails
		return nil, err
	}

	logger.Log.Info("franz-go Producer initialized",
		zap.String("topic", cfg.Topic),
		zap.Strings("brokers", cfg.Brokers),
	)

	return &Producer{client: client, topic: cfg.Topic}, nil
}

// SendMessage sends a message synchronously to the configured Kafka topic.
// For higher throughput, consider using the asynchronous Produce method.
func (p *Producer) SendMessage(ctx context.Context, key, value []byte) error {
	record := &kgo.Record{Key: key, Value: value, Topic: p.topic}

	// ProduceSync waits for the record to be produced and acknowledged.
	results := p.client.ProduceSync(ctx, record)

	// ProduceSync returns a slice of ProduceResult, check the error for the first (and only) result.
	if err := results[0].Err; err != nil {
		logger.Log.Error("Failed to send message to Kafka with franz-go",
			zap.String("topic", record.Topic),
			zap.Error(err),
		)
		return err
	}

	return nil
}

// Close closes the franz-go client connection.
func (p *Producer) Close() {
	if p.client != nil {
		logger.Log.Info("Closing franz-go Producer...")
		p.client.Close() // Close is synchronous
	}
}
