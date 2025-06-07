package sinks

import (
	"context"
	"log"
	"time"

	"github.com/gocql/gocql"
	"github.com/kxrxh/social-rating-system/stream/internal/config"
	"github.com/kxrxh/social-rating-system/stream/internal/models"
)

// ScyllaSink handles writing to ScyllaDB
type ScyllaSink struct {
	session *gocql.Session
	cfg     *config.Config
}

// NewScyllaSink creates a new ScyllaDB sink
func NewScyllaSink(session *gocql.Session, cfg *config.Config) *ScyllaSink {
	return &ScyllaSink{
		session: session,
		cfg:     cfg,
	}
}

// Process handles writing to ScyllaDB
func (ss *ScyllaSink) Process(ctx context.Context, in <-chan interface{}) {
	// Prepare CQL statement for events archive
	insertStmt := `INSERT INTO events_archive (
		citizen_id, event_id, event_type, event_subtype, source_system,
		timestamp, location_city, location_region, location_country,
		confidence, score_change, new_score, previous_score, applied_rules
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	// Prepare the statement once for better performance
	prepared := ss.session.Query(insertStmt)

	// Use batch processing for better throughput
	eventBatch := make([]models.ProcessedEvent, 0, ss.cfg.BatchSize)

	// Flush ticker for periodic writes
	ticker := time.NewTicker(ss.cfg.FlushInterval)
	defer ticker.Stop()

	flushBatch := func() {
		if len(eventBatch) == 0 {
			return
		}

		// Process batch
		batch := ss.session.NewBatch(gocql.LoggedBatch)
		for _, processedEvent := range eventBatch {
			batch.Query(insertStmt,
				processedEvent.CitizenID,
				processedEvent.EventID,
				processedEvent.EventType,
				processedEvent.EventSubtype,
				processedEvent.SourceSystem,
				processedEvent.Timestamp,
				processedEvent.Location.City,
				processedEvent.Location.Region,
				processedEvent.Location.Country,
				processedEvent.Confidence,
				processedEvent.ScoreChange,
				processedEvent.NewScore,
				processedEvent.PreviousScore,
				processedEvent.AppliedRules,
			)
		}

		if err := ss.session.ExecuteBatch(batch); err != nil {
			log.Printf("Error writing batch to ScyllaDB: %v", err)
			// Fallback to individual inserts on batch failure
			for _, processedEvent := range eventBatch {
				if err := prepared.Bind(
					processedEvent.CitizenID,
					processedEvent.EventID,
					processedEvent.EventType,
					processedEvent.EventSubtype,
					processedEvent.SourceSystem,
					processedEvent.Timestamp,
					processedEvent.Location.City,
					processedEvent.Location.Region,
					processedEvent.Location.Country,
					processedEvent.Confidence,
					processedEvent.ScoreChange,
					processedEvent.NewScore,
					processedEvent.PreviousScore,
					processedEvent.AppliedRules,
				).Exec(); err != nil {
					log.Printf("Error writing single event to ScyllaDB: %v", err)
				}
			}
		} else {
			log.Printf("ScyllaDB: Wrote batch of %d events", len(eventBatch))
		}

		eventBatch = eventBatch[:0]
	}

	for {
		select {
		case event, ok := <-in:
			if !ok {
				// Channel closed, flush remaining batch and return
				flushBatch()
				return
			}

			processedEvent, ok := event.(models.ProcessedEvent)
			if !ok {
				continue
			}

			eventBatch = append(eventBatch, processedEvent)

			// Flush batch if it reaches the batch size
			if len(eventBatch) >= ss.cfg.BatchSize {
				flushBatch()
			}

		case <-ticker.C:
			// Periodic flush
			flushBatch()

		case <-ctx.Done():
			// Context cancelled, flush remaining batch and return
			flushBatch()
			return
		}
	}
}
