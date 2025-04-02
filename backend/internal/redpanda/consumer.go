package redpanda

import (
	"context"
	"errors"

	"github.com/kxrxh/social-rating-system/internal/config"
	"github.com/kxrxh/social-rating-system/internal/logger"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.uber.org/zap"
)

// MessageHandler is a function type that processes a single Kafka record.
type MessageHandler func(ctx context.Context, record *kgo.Record) error

// Consumer handles consuming messages from Kafka/RedPanda using franz-go.
type Consumer struct {
	client *kgo.Client
}

// NewConsumer creates a new franz-go Kafka consumer instance.
func NewConsumer(cfg config.RedPandaConfig) (*Consumer, error) {
	opts := []kgo.Opt{
		kgo.SeedBrokers(cfg.Brokers...),
		kgo.ConsumerGroup(cfg.GroupID),
		kgo.ConsumeTopics(cfg.Topic),
		// Add other client options if needed (e.g., SASL, TLS, FetchMaxBytes)
		// kgo.SASL(scram.Auth{User: "user", Pass: "pass"}.AsSha256Mechanism()),
		// kgo.Dialer(tlsDialer),
	}

	client, err := kgo.NewClient(opts...)
	if err != nil {
		logger.Log.Error("Failed to create franz-go consumer client", zap.Error(err))
		return nil, err
	}

	// Optional: Ping the cluster to ensure connectivity
	if err := client.Ping(context.Background()); err != nil {
		logger.Log.Error("Failed to ping Kafka cluster", zap.Error(err))
		client.Close() // Close the client if ping fails
		return nil, err
	}

	logger.Log.Info("franz-go Consumer initialized",
		zap.String("topic", cfg.Topic),
		zap.String("group_id", cfg.GroupID),
		zap.Strings("brokers", cfg.Brokers),
	)

	return &Consumer{client: client}, nil
}

// StartConsuming starts the message consumption loop.
// It blocks until the context is canceled or a fatal error occurs.
func (c *Consumer) StartConsuming(ctx context.Context, handler MessageHandler) error {
	if c.client == nil {
		return errors.New("consumer client is not initialized")
	}

	logger.Log.Info("Starting consumer loop...")

	for {
		fetches := c.client.PollFetches(ctx)
		if err := fetches.Err(); err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				logger.Log.Info("Consumer context canceled, shutting down.")
				return nil // Expected error on shutdown
			}
			logger.Log.Error("Fatal error during fetch", zap.Error(err))
			return err // Return other fatal errors
		}

		fetches.EachPartition(func(p kgo.FetchTopicPartition) {
			for _, record := range p.Records {
				err := handler(ctx, record) // Process the message
				if err != nil {
					// Decide how to handle processing errors (log, skip, retry, dead-letter queue)
					logger.Log.Error("Error processing record",
						zap.Int64("offset", record.Offset),
						zap.Int32("partition", record.Partition),
						zap.Error(err),
					)
					// Potentially commit anyway or implement specific error handling
				}
			}
			// Errors per partition can be checked here using p.Err(), if needed.
		})

		// Optional: Commit offsets explicitly if AutoCommit is disabled
		// if err := c.client.CommitUncommittedOffsets(ctx); err != nil {
		// 	 logger.Log.Error("Failed to commit offsets", zap.Error(err))
		// 	 // Handle commit failure (retry?)
		// }
	}
}

// Close gracefully shuts down the consumer client.
func (c *Consumer) Close() {
	if c.client != nil {
		logger.Log.Info("Closing franz-go Consumer...")
		c.client.Close() // Close is synchronous
	}
}

// Example usage (can be placed in main.go or similar)
/*
func main() {
    cfg, err := config.Load()
    if err != nil {
        logger.Log.Fatal("Failed to load config", zap.Error(err))
    }

    consumer, err := redpanda.NewConsumer(cfg.RedPanda)
    if err != nil {
        logger.Log.Fatal("Failed to create consumer", zap.Error(err))
    }
    defer consumer.Close()

    ctx, cancel := context.WithCancel(context.Background())

    // Handle OS signals for graceful shutdown
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
    go func() {
        <-sigChan
        logger.Log.Info("Received shutdown signal, canceling context...")
        cancel()
    }()

    // Define your message handler logic
    handler := func(ctx context.Context, record *kgo.Record) error {
        logger.Log.Info("Received message",
            zap.ByteString("key", record.Key),
            zap.ByteString("value", record.Value),
            zap.Int32("partition", record.Partition),
            zap.Int64("offset", record.Offset),
        )
        // Add your message processing logic here
        return nil
    }

    // Start consuming in a separate goroutine
    go func() {
        err := consumer.StartConsuming(ctx, handler)
        if err != nil && !errors.Is(err, context.Canceled) {
            logger.Log.Error("Consumer stopped with error", zap.Error(err))
            // Potentially trigger application shutdown or restart logic
            cancel() // Ensure main exits if consumer fails unexpectedly
        }
    }()

    logger.Log.Info("Consumer started. Press Ctrl+C to exit.")
    <-ctx.Done() // Wait until context is canceled
    logger.Log.Info("Shutting down application.")
}
*/
