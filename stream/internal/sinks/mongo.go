package sinks

import (
	"context"
	"log"

	"github.com/kxrxh/social-rating-system/stream/internal/models"
)

// MongoSink handles processed events for MongoDB
type MongoSink struct{}

// NewMongoSink creates a new MongoDB sink
func NewMongoSink() *MongoSink {
	return &MongoSink{}
}

// Process processes events for MongoDB storage
func (ms *MongoSink) Process(ctx context.Context, in <-chan interface{}) {
	for {
		select {
		case event, ok := <-in:
			if !ok {
				return
			}
			// MongoDB updates are already handled by the scoring engine
			if processedEvent, ok := event.(models.ProcessedEvent); ok {
				log.Printf("MongoDB: Processed event %s with score change %+.2f",
					processedEvent.EventID, processedEvent.ScoreChange)
			}
		case <-ctx.Done():
			return
		}
	}
}
