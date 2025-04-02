// Package producer handles the Kafka client connection and event production loop.
package producer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
	"github.com/kxrxh/social-rating-system/event-producer/internal/config"
	"github.com/kxrxh/social-rating-system/event-producer/internal/generator"
	"github.com/twmb/franz-go/pkg/kgo"
)

// Producer manages the Kafka client and event generation/sending loop.
type Producer struct {
	client     *kgo.Client
	cfg        *config.Config
	generator  *generator.Generator
	citizenIDs []string
}

// NewProducer creates and initializes a new Producer.
func NewProducer(cfg *config.Config, gen *generator.Generator) (*Producer, error) {
	log.Printf("Initializing Kafka client with brokers: %v", cfg.KafkaBrokers)
	opts := []kgo.Opt{
		kgo.SeedBrokers(cfg.KafkaBrokers...),
		kgo.DefaultProduceTopic(cfg.KafkaTopic),
		kgo.RequiredAcks(kgo.LeaderAck()),
		kgo.ProducerBatchCompression(kgo.SnappyCompression()),
		kgo.AllowAutoTopicCreation(),
		// You might want to add other options like authentication here
	}

	client, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka client: %w", err)
	}

	// Ping to check connectivity early
	ctxPing, cancelPing := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelPing()
	if err = client.Ping(ctxPing); err != nil {
		log.Printf("Warning: Failed to ping Kafka broker(s): %v. Continuing...", err)
		// Depending on requirements, you might want to Fatalf or return error here
	} else {
		log.Println("Kafka client initialized and connection ping successful.")
	}

	// Prepare citizen IDs
	cIDs := make([]string, cfg.NumCitizens)
	for i := 0; i < cfg.NumCitizens; i++ {
		cIDs[i] = uuid.NewString()
	}
	log.Printf("Simulating events for %d citizens: %v...", len(cIDs), cIDs)

	return &Producer{
		client:     client,
		cfg:        cfg,
		generator:  gen,
		citizenIDs: cIDs,
	}, nil
}

// Run starts the event production loop.
// It blocks until the context is cancelled.
func (p *Producer) Run(ctx context.Context) {
	ticker := time.NewTicker(p.cfg.SendInterval)
	defer ticker.Stop()

	log.Printf("Starting event production loop (interval: %s)...", p.cfg.SendInterval)

	for {
		select {
		case <-ticker.C:
			// Select a random citizen
			citizenID := p.citizenIDs[rand.Intn(len(p.citizenIDs))] // Note: Uses global rand, consider passing generator's rand

			// Generate a random event
			event := p.generator.GenerateRandomEvent(citizenID)

			baseEvent := event.GetBaseEvent()

			// Marshal event to JSON
			jsonData, err := json.Marshal(event)
			if err != nil {
				log.Printf("Error marshalling event (%T) to JSON for citizen %s: %v", event, baseEvent.CitizenID, err)
				continue // Skip this event
			}

			// Create Kafka record
			record := &kgo.Record{
				Key:       []byte(baseEvent.CitizenID),
				Value:     jsonData,
				Headers:   []kgo.RecordHeader{{Key: "event_type", Value: []byte(baseEvent.EventType)}},
				Timestamp: baseEvent.Timestamp,
			}

			// Produce record asynchronously
			p.client.Produce(ctx, record, func(r *kgo.Record, err error) {
				if err != nil {
					if ctx.Err() != nil {
						log.Printf("Failed to produce message for key %s due to shutdown: %v", string(r.Key), err)
					} else {
						log.Printf("Error producing message for key %s: %v", string(r.Key), err)
					}
				} else {
					// log.Printf("Produced %s for %s => partition %d offset %d", baseEvent.EventType, string(r.Key), r.Partition, r.Offset)
				}
			})

		case <-ctx.Done():
			log.Println("Stopping event production loop due to context cancellation.")
			return
		}
	}
}

// Close shuts down the producer client gracefully.
func (p *Producer) Close() {
	log.Println("Shutting down Kafka client...")
	// Close() blocks until the client has cleanly flushed and shut down.
	p.client.Close()
	log.Println("Kafka client shut down.")
}
