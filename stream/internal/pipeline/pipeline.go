package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gocql/gocql"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/kxrxh/social-rating-system/stream/internal/config"
	"github.com/kxrxh/social-rating-system/stream/internal/models"
	"github.com/kxrxh/social-rating-system/stream/internal/scoring"
	"github.com/kxrxh/social-rating-system/stream/internal/sinks"
	"github.com/twmb/franz-go/pkg/kgo"
)

// Pipeline represents the complete streaming pipeline
type Pipeline struct {
	kafkaClient   *kgo.Client
	scoringEngine *scoring.Engine
	influxClient  influxdb2.Client
	scyllaSession *gocql.Session
	cfg           *config.Config
	eventSource   chan interface{}
	stopChan      chan struct{}
	running       bool
}

// New creates a new streaming pipeline
func New(kafkaClient *kgo.Client, scoringEngine *scoring.Engine,
	influxClient influxdb2.Client, scyllaSession *gocql.Session, cfg *config.Config) *Pipeline {

	// Increase buffer size significantly for high throughput
	eventSourceBufferSize := cfg.BatchSize * 20 // Increased from 10 to 20

	return &Pipeline{
		kafkaClient:   kafkaClient,
		scoringEngine: scoringEngine,
		influxClient:  influxClient,
		scyllaSession: scyllaSession,
		cfg:           cfg,
		eventSource:   make(chan interface{}, eventSourceBufferSize),
		stopChan:      make(chan struct{}),
		running:       false,
	}
}

// Start starts the streaming pipeline
func (p *Pipeline) Start(ctx context.Context) error {
	log.Printf("Starting stream processing pipeline with %d workers...", p.cfg.Workers)
	log.Printf("Event source buffer size: %d", cap(p.eventSource))
	log.Printf("Sink buffer size: %d", p.cfg.BatchSize*6)

	p.running = true

	// Start Kafka consumption goroutine
	go p.consumeKafkaEvents(ctx)

	// Create larger buffered channels for different sinks to prevent blocking
	bufferSize := p.cfg.BatchSize * 10 // Increased from 6 to 10
	mongoChannel := make(chan interface{}, bufferSize)
	influxChannel := make(chan interface{}, bufferSize)
	scyllaChannel := make(chan interface{}, bufferSize)

	// Start event processing workers (direct processing, no go-streams)
	log.Printf("Starting %d event processing workers...", p.cfg.Workers)
	for i := 0; i < p.cfg.Workers; i++ {
		go p.processEventsWorker(ctx, mongoChannel, influxChannel, scyllaChannel)
	}

	// Create sink processors
	mongoSinkProcessor := sinks.NewMongoSink()
	influxSinkProcessor := sinks.NewInfluxSink(p.influxClient, p.cfg)
	scyllaSinkProcessor := sinks.NewScyllaSink(p.scyllaSession, p.cfg)

	// Start multiple sink processors for better throughput
	// Increase MongoDB processors for high throughput
	for i := 0; i < p.cfg.Workers*2; i++ {
		go mongoSinkProcessor.Process(ctx, mongoChannel)
	}
	// Multiple InfluxDB processors for better throughput
	log.Printf("Starting %d InfluxDB sink processors...", p.cfg.Workers)
	for i := 0; i < p.cfg.Workers; i++ {
		go influxSinkProcessor.Process(ctx, influxChannel)
	}
	// Multiple ScyllaDB processors for better throughput
	for i := 0; i < p.cfg.Workers*2; i++ {
		go scyllaSinkProcessor.Process(ctx, scyllaChannel)
	}

	// Start pipeline health monitoring
	go p.monitorPipelineHealth(ctx)

	return nil
}

// Stop stops the streaming pipeline
func (p *Pipeline) Stop(ctx context.Context) error {
	log.Println("Stopping stream processing pipeline...")
	p.running = false
	close(p.stopChan)
	close(p.eventSource)
	return nil
}

// IsRunning returns whether the pipeline is currently running
func (p *Pipeline) IsRunning() bool {
	return p.running
}

// processEventsWorker processes events from eventSource and distributes to all sinks
func (p *Pipeline) processEventsWorker(ctx context.Context, mongoChannel, influxChannel, scyllaChannel chan interface{}) {
	log.Println("Event processing worker started")
	var processedCount int64

	for {
		select {
		case event, ok := <-p.eventSource:
			if !ok {
				log.Printf("Event processing worker stopping. Processed %d events", processedCount)
				return
			}

			// Process event through scoring engine
			processedEvent := p.processEvent(event)
			if processedEvent == nil {
				continue
			}

			processedCount++

			// Send to all sinks (non-blocking)
			select {
			case mongoChannel <- processedEvent:
			default:
				log.Printf("MongoDB channel full, dropping event")
			}

			select {
			case influxChannel <- processedEvent:
			default:
				log.Printf("InfluxDB channel full, dropping event")
			}

			select {
			case scyllaChannel <- processedEvent:
			default:
				log.Printf("ScyllaDB channel full, dropping event")
			}

		case <-ctx.Done():
			log.Printf("Event processing worker stopped by context. Processed %d events", processedCount)
			return
		}
	}
}

// consumeKafkaEvents consumes events from Kafka and feeds them to the pipeline
func (p *Pipeline) consumeKafkaEvents(ctx context.Context) {
	log.Printf("Starting Kafka consumption with single optimized consumer")

	var totalProcessed int64
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ticker.C:
				log.Printf("Kafka consumption metrics: %d total events processed", totalProcessed)
			case <-ctx.Done():
				return
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			log.Printf("Kafka consumer stopping, total events processed: %d", totalProcessed)
			return
		case <-p.stopChan:
			log.Printf("Kafka consumer stopping, total events processed: %d", totalProcessed)
			return
		default:
			// Poll without timeout to avoid constant deadline exceeded errors
			fetches := p.kafkaClient.PollFetches(ctx)

			if errs := fetches.Errors(); len(errs) > 0 {
				for _, err := range errs {
					log.Printf("Kafka fetch error: %v", err)
				}
				continue
			}

			var batchProcessed int
			fetches.EachPartition(func(partition kgo.FetchTopicPartition) {
				if len(partition.Records) == 0 {
					return
				}

				log.Printf("Processing batch of %d records from partition %d", len(partition.Records), partition.Partition)

				// Process records in parallel batches
				batchSize := 50 // Process in smaller batches
				for i := 0; i < len(partition.Records); i += batchSize {
					end := i + batchSize
					if end > len(partition.Records) {
						end = len(partition.Records)
					}

					batch := partition.Records[i:end]
					events := make([]models.Event, 0, len(batch))

					// Parse batch
					for _, record := range batch {
						var event models.Event
						if err := json.Unmarshal(record.Value, &event); err != nil {
							log.Printf("Error unmarshalling event: %v", err)
							continue
						}
						events = append(events, event)
					}

					// Send events to pipeline with non-blocking approach
					for _, event := range events {
						select {
						case p.eventSource <- event:
							batchProcessed++
						case <-ctx.Done():
							return
						default:
							// Pipeline is full, log warning but continue
							log.Printf("Pipeline congested, dropping event: %s", event.EventID)
						}
					}
				}
			})

			if batchProcessed > 0 {
				totalProcessed += int64(batchProcessed)
				log.Printf("Processed batch of %d events (total: %d)", batchProcessed, totalProcessed)
			}

			// Commit offsets after processing
			if err := p.kafkaClient.CommitUncommittedOffsets(ctx); err != nil {
				log.Printf("Error committing offsets: %v", err)
			}
		}
	}
}

// processEvent processes a single event through the scoring engine
func (p *Pipeline) processEvent(event interface{}) interface{} {
	evt, ok := event.(models.Event)
	if !ok {
		log.Printf("Invalid event type: %T", event)
		return nil
	}

	log.Printf("Pipeline processEvent: Processing event %s through scoring engine", evt.EventID)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	processedEvent, err := p.scoringEngine.ProcessEvent(ctx, evt)
	if err != nil {
		log.Printf("Error processing event %s: %v", evt.EventID, err)
		// Return the original event with error information
		errorEvent := models.ProcessedEvent{
			Event:           evt,
			ProcessedAt:     time.Now(),
			ProcessingError: err.Error(),
			ProcessingID:    fmt.Sprintf("error_%d", time.Now().UnixNano()),
		}
		log.Printf("Pipeline processEvent: Returning error event %s", errorEvent.EventID)
		return errorEvent
	}

	log.Printf("Pipeline processEvent: Successfully processed event %s, sending to sinks", processedEvent.EventID)
	return *processedEvent
}

// monitorPipelineHealth monitors and logs pipeline performance metrics
func (p *Pipeline) monitorPipelineHealth(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Log channel utilization
			eventSourceUsage := float64(len(p.eventSource)) / float64(cap(p.eventSource)) * 100
			log.Printf("Pipeline Health - Event source buffer usage: %.1f%% (%d/%d)",
				eventSourceUsage, len(p.eventSource), cap(p.eventSource))

			if eventSourceUsage > 80 {
				log.Printf("WARNING: Event source buffer is %.1f%% full - possible bottleneck", eventSourceUsage)
			}

		case <-ctx.Done():
			return
		}
	}
}
