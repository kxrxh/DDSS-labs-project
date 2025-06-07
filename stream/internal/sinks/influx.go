package sinks

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/kxrxh/social-rating-system/stream/internal/config"
	"github.com/kxrxh/social-rating-system/stream/internal/models"
)

// InfluxSink handles writing to InfluxDB with optimized schema
type InfluxSink struct {
	client          influxdb2.Client
	cfg             *config.Config
	eventsWriteAPI  api.WriteAPI
	metricsWriteAPI api.WriteAPI
	eventsBatch     []*write.Point
	metricsBatch    []*write.Point
	lastFlush       time.Time
}

// NewInfluxSink creates a new InfluxDB sink with improved batching
func NewInfluxSink(client influxdb2.Client, cfg *config.Config) *InfluxSink {
	sink := &InfluxSink{
		client:       client,
		cfg:          cfg,
		eventsBatch:  make([]*write.Point, 0, cfg.BatchSize),
		metricsBatch: make([]*write.Point, 0, cfg.BatchSize),
		lastFlush:    time.Now(),
	}

	// Initialize write APIs
	sink.eventsWriteAPI = client.WriteAPI(cfg.InfluxOrg, cfg.InfluxRawEventsBucket)
	sink.metricsWriteAPI = client.WriteAPI(cfg.InfluxOrg, cfg.InfluxDerivedMetricsBucket)

	// Handle write errors
	sink.setupErrorHandling()

	return sink
}

// setupErrorHandling configures error handling for write APIs
func (is *InfluxSink) setupErrorHandling() {
	go func() {
		for err := range is.eventsWriteAPI.Errors() {
			log.Printf("InfluxDB events write error: %v", err)
		}
	}()

	go func() {
		for err := range is.metricsWriteAPI.Errors() {
			log.Printf("InfluxDB metrics write error: %v", err)
		}
	}()
}

// Process handles writing to InfluxDB with new optimized format
func (is *InfluxSink) Process(ctx context.Context, in <-chan interface{}) {
	log.Printf("Starting InfluxDB sink with buckets: events=%s, metrics=%s",
		is.cfg.InfluxRawEventsBucket, is.cfg.InfluxDerivedMetricsBucket)

	ticker := time.NewTicker(is.cfg.FlushInterval)
	defer ticker.Stop()
	defer is.forceFlush()

	for {
		select {
		case event, ok := <-in:
			if !ok {
				log.Println("InfluxDB sink: input channel closed")
				return
			}
			is.processEvent(event)

		case <-ticker.C:
			is.flushIfNeeded(false)

		case <-ctx.Done():
			log.Println("InfluxDB sink: context cancelled")
			return
		}
	}
}

// processEvent converts and batches a single event
func (is *InfluxSink) processEvent(event interface{}) {
	processedEvent, ok := event.(models.ProcessedEvent)
	if !ok {
		log.Printf("InfluxDB sink: invalid event type %T", event)
		return
	}

	log.Printf("Processing event %s for citizen %s",
		processedEvent.EventID, processedEvent.CitizenID)

	// Create and add points to batches
	eventPoint := is.createEventPoint(processedEvent)
	is.eventsBatch = append(is.eventsBatch, eventPoint)

	// Create citizen score snapshot
	scorePoint := is.createScorePoint(processedEvent)
	is.metricsBatch = append(is.metricsBatch, scorePoint)

	// Create rule execution metrics if rules were applied
	if len(processedEvent.AppliedRules) > 0 {
		rulePoints := is.createRuleMetrics(processedEvent)
		is.metricsBatch = append(is.metricsBatch, rulePoints...)
	}

	// Create location-based aggregation point
	locationPoint := is.createLocationMetric(processedEvent)
	is.metricsBatch = append(is.metricsBatch, locationPoint)

	// Flush if batches are full
	is.flushIfNeeded(true)
}

// createEventPoint creates the main event record with event_type as measurement
func (is *InfluxSink) createEventPoint(event models.ProcessedEvent) *write.Point {
	// Use event_type as measurement name for better querying
	measurement := fmt.Sprintf("event_%s", strings.ToLower(event.EventType))

	// Tags (low cardinality, used for filtering)
	tags := map[string]string{
		"citizen_id":    event.CitizenID,
		"event_subtype": event.EventSubtype,
		"source_system": event.SourceSystem,
		"city":          event.Location.City,
		"region":        event.Location.Region,
		"country":       event.Location.Country,
		"processing_id": event.ProcessingID,
	}

	// Add categorized payload as tags (only string values with low cardinality)
	for key, value := range event.Payload {
		if strVal, ok := value.(string); ok && len(strVal) < 50 {
			tags[fmt.Sprintf("p_%s", key)] = strVal
		}
	}

	// Fields (high cardinality, actual data)
	fields := map[string]interface{}{
		"event_id":           event.EventID,
		"event_name":         event.EventName,
		"confidence":         event.Confidence,
		"latitude":           event.Location.Latitude,
		"longitude":          event.Location.Longitude,
		"version":            event.Version,
		"previous_score":     event.PreviousScore,
		"new_score":          event.NewScore,
		"score_change":       event.ScoreChange,
		"processing_time_ms": event.ProcessedAt.Sub(event.Timestamp).Milliseconds(),
	}

	// Add payload fields (numeric and complex data)
	for key, value := range event.Payload {
		switch v := value.(type) {
		case int, int32, int64, float32, float64, bool:
			fields[fmt.Sprintf("payload_%s", key)] = v
		case string:
			if len(v) >= 50 { // Long strings as fields
				fields[fmt.Sprintf("payload_%s", key)] = v
			}
		default:
			// Complex objects as JSON strings
			if jsonData, err := json.Marshal(v); err == nil {
				fields[fmt.Sprintf("payload_%s", key)] = string(jsonData)
			}
		}
	}

	// Add error information if present
	if event.ProcessingError != "" {
		fields["processing_error"] = event.ProcessingError
		tags["has_error"] = "true"
	} else {
		tags["has_error"] = "false"
	}

	// Full payload as compressed JSON for backup/debugging
	if payloadJSON, err := json.Marshal(event.Payload); err == nil {
		fields["payload_full"] = string(payloadJSON)
	}

	return influxdb2.NewPoint(measurement, tags, fields, event.Timestamp)
}

// createScorePoint creates a citizen score snapshot
func (is *InfluxSink) createScorePoint(event models.ProcessedEvent) *write.Point {
	tags := map[string]string{
		"citizen_id":    event.CitizenID,
		"trigger_event": event.EventType,
		"city":          event.Location.City,
		"region":        event.Location.Region,
		"country":       event.Location.Country,
		"processing_id": event.ProcessingID,
	}

	fields := map[string]interface{}{
		"score":          event.NewScore,
		"score_change":   event.ScoreChange,
		"previous_score": event.PreviousScore,
		"confidence":     event.Confidence,
		"rules_applied":  len(event.AppliedRules),
		"event_id":       event.EventID,
	}

	if event.ProcessingError != "" {
		fields["error"] = event.ProcessingError
		tags["status"] = "error"
	} else {
		tags["status"] = "success"
	}

	return influxdb2.NewPoint("citizen_score", tags, fields, event.ProcessedAt)
}

// createRuleMetrics creates metrics for each applied rule
func (is *InfluxSink) createRuleMetrics(event models.ProcessedEvent) []*write.Point {
	points := make([]*write.Point, 0, len(event.AppliedRules))

	for _, rule := range event.AppliedRules {
		tags := map[string]string{
			"rule_name":     rule,
			"citizen_id":    event.CitizenID,
			"event_type":    event.EventType,
			"event_subtype": event.EventSubtype,
			"city":          event.Location.City,
			"region":        event.Location.Region,
			"country":       event.Location.Country,
			"processing_id": event.ProcessingID,
		}

		fields := map[string]interface{}{
			"execution_count": 1,
			"score_impact":    event.ScoreChange, // This could be refined per rule
			"event_id":        event.EventID,
			"citizen_score":   event.NewScore,
		}

		point := influxdb2.NewPoint("rule_execution", tags, fields, event.ProcessedAt)
		points = append(points, point)
	}

	return points
}

// createLocationMetric creates location-based aggregation metrics
func (is *InfluxSink) createLocationMetric(event models.ProcessedEvent) *write.Point {
	tags := map[string]string{
		"event_type":    event.EventType,
		"city":          event.Location.City,
		"region":        event.Location.Region,
		"country":       event.Location.Country,
		"processing_id": event.ProcessingID,
	}

	fields := map[string]interface{}{
		"event_count":        1,
		"total_score_change": event.ScoreChange,
		"avg_confidence":     event.Confidence,
		"citizen_count":      1, // Will be aggregated
	}

	// Add processing status
	if event.ProcessingError != "" {
		tags["status"] = "error"
		fields["error_count"] = 1
		fields["success_count"] = 0
	} else {
		tags["status"] = "success"
		fields["error_count"] = 0
		fields["success_count"] = 1
	}

	return influxdb2.NewPoint("location_events", tags, fields, event.ProcessedAt)
}

// flushIfNeeded flushes batches if they exceed size limits or time threshold
func (is *InfluxSink) flushIfNeeded(checkSize bool) {
	shouldFlush := false

	if checkSize {
		shouldFlush = len(is.eventsBatch) >= is.cfg.BatchSize ||
			len(is.metricsBatch) >= is.cfg.BatchSize
	}

	// Also flush based on time
	if time.Since(is.lastFlush) >= is.cfg.FlushInterval {
		shouldFlush = true
	}

	if shouldFlush {
		is.flush()
	}
}

// flush writes all batched points to InfluxDB
func (is *InfluxSink) flush() {
	eventCount := len(is.eventsBatch)
	metricCount := len(is.metricsBatch)

	if eventCount == 0 && metricCount == 0 {
		return
	}

	// Write events batch
	if eventCount > 0 {
		for _, point := range is.eventsBatch {
			is.eventsWriteAPI.WritePoint(point)
		}
		is.eventsBatch = is.eventsBatch[:0]
	}

	// Write metrics batch
	if metricCount > 0 {
		for _, point := range is.metricsBatch {
			is.metricsWriteAPI.WritePoint(point)
		}
		is.metricsBatch = is.metricsBatch[:0]
	}

	is.lastFlush = time.Now()
	log.Printf("InfluxDB: Flushed %d events, %d metrics", eventCount, metricCount)
}

// forceFlush ensures all remaining data is written before shutdown
func (is *InfluxSink) forceFlush() {
	is.flush()
	is.eventsWriteAPI.Flush()
	is.metricsWriteAPI.Flush()
	log.Println("InfluxDB sink: Force flush completed")
}
