package sources

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/kxrxh/social-rating-system/stream/internal/models"
	"github.com/reugn/go-streams"
	"github.com/twmb/franz-go/pkg/kgo"
)

// KafkaSource implements streams.Source interface for consuming from Kafka/Redpanda
type KafkaSource struct {
	client  *kgo.Client
	out     chan interface{}
	ctx     context.Context
	cancel  context.CancelFunc
	topic   string
	groupID string
}

// NewKafkaSource creates a new Kafka source for consuming events
func NewKafkaSource(brokers []string, topic, groupID string) (*KafkaSource, error) {
	opts := []kgo.Opt{
		kgo.SeedBrokers(brokers...),
		kgo.ConsumeTopics(topic),
		kgo.ConsumerGroup(groupID),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtStart()), // Start from beginning if no offset
		kgo.WithLogger(kgo.BasicLogger(log.Writer(), kgo.LogLevelInfo, func() string {
			return "KAFKA_SOURCE: "
		})),
	}

	client, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &KafkaSource{
		client:  client,
		out:     make(chan interface{}, 100), // Buffered channel for better throughput
		ctx:     ctx,
		cancel:  cancel,
		topic:   topic,
		groupID: groupID,
	}, nil
}

// Out returns the output channel
func (k *KafkaSource) Out() <-chan interface{} {
	return k.out
}

// Via connects this source to a flow
func (k *KafkaSource) Via(flow streams.Flow) streams.Flow {
	go func() {
		for item := range k.Out() {
			flow.In() <- item
		}
		close(flow.In())
	}()
	return flow
}

// To connects this source to a sink
func (k *KafkaSource) To(sink streams.Sink) {
	go func() {
		for item := range k.Out() {
			sink.In() <- item
		}
		close(sink.In())
	}()
}

// Start begins consuming messages from Kafka
func (k *KafkaSource) Start() {
	go k.consume()
}

// consume is the main consumption loop
func (k *KafkaSource) consume() {
	defer close(k.out)
	defer k.client.Close()

	log.Printf("Starting Kafka consumption from topic: %s, group: %s", k.topic, k.groupID)

	for {
		select {
		case <-k.ctx.Done():
			log.Println("Kafka source context cancelled, stopping consumption")
			return
		default:
			// Poll for messages with timeout
			fetches := k.client.PollFetches(k.ctx)

			// Handle any errors
			if errs := fetches.Errors(); len(errs) > 0 {
				for _, err := range errs {
					log.Printf("Kafka fetch error: %v", err)
				}
				continue
			}

			// Process each message
			fetches.EachPartition(func(p kgo.FetchTopicPartition) {
				for _, record := range p.Records {
					if err := k.processRecord(record); err != nil {
						log.Printf("Error processing record: %v", err)
						continue
					}
				}
			})

			// Commit offsets
			if err := k.client.CommitUncommittedOffsets(k.ctx); err != nil {
				log.Printf("Error committing offsets: %v", err)
			}
		}
	}
}

// processRecord processes a single Kafka record
func (k *KafkaSource) processRecord(record *kgo.Record) error {
	// Parse the JSON message into an Event
	var event models.Event
	if err := json.Unmarshal(record.Value, &event); err != nil {
		log.Printf("Error unmarshalling event JSON: %v, raw message: %s", err, string(record.Value))
		return err
	}

	// Validate the event
	if event.EventID == "" || event.CitizenID == "" || event.EventType == "" {
		log.Printf("Invalid event: missing required fields - EventID: %s, CitizenID: %s, EventType: %s",
			event.EventID, event.CitizenID, event.EventType)
		return nil // Skip invalid events rather than error
	}

	// Add some metadata from Kafka
	if event.Payload == nil {
		event.Payload = make(map[string]interface{})
	}
	event.Payload["kafka_partition"] = record.Partition
	event.Payload["kafka_offset"] = record.Offset
	event.Payload["kafka_timestamp"] = record.Timestamp

	log.Printf("Processing event: ID=%s, Type=%s, Citizen=%s",
		event.EventID, event.EventType, event.CitizenID)

	// Send the event to the output channel
	select {
	case k.out <- event:
		return nil
	case <-k.ctx.Done():
		return k.ctx.Err()
	case <-time.After(5 * time.Second):
		log.Printf("Timeout sending event to output channel, event: %s", event.EventID)
		return nil
	}
}

// Stop gracefully stops the Kafka source
func (k *KafkaSource) Stop() {
	log.Println("Stopping Kafka source...")
	k.cancel()
}

// KafkaSourceBuilder helps build KafkaSource with configuration
type KafkaSourceBuilder struct {
	brokers []string
	topic   string
	groupID string
}

// NewKafkaSourceBuilder creates a new builder
func NewKafkaSourceBuilder() *KafkaSourceBuilder {
	return &KafkaSourceBuilder{}
}

// WithBrokers sets the Kafka brokers
func (b *KafkaSourceBuilder) WithBrokers(brokers []string) *KafkaSourceBuilder {
	b.brokers = brokers
	return b
}

// WithTopic sets the Kafka topic
func (b *KafkaSourceBuilder) WithTopic(topic string) *KafkaSourceBuilder {
	b.topic = topic
	return b
}

// WithGroupID sets the consumer group ID
func (b *KafkaSourceBuilder) WithGroupID(groupID string) *KafkaSourceBuilder {
	b.groupID = groupID
	return b
}

// Build creates the KafkaSource
func (b *KafkaSourceBuilder) Build() (*KafkaSource, error) {
	if len(b.brokers) == 0 {
		b.brokers = []string{"localhost:9092"}
	}
	if b.topic == "" {
		b.topic = "citizen_events"
	}
	if b.groupID == "" {
		b.groupID = "social-credit-processor"
	}

	return NewKafkaSource(b.brokers, b.topic, b.groupID)
}
