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
	client            influxdb2.Client
	cfg               *config.Config
	eventsWriteAPI    api.WriteAPI
	metricsWriteAPI   api.WriteAPI
	analyticsWriteAPI api.WriteAPI
	eventsBatch       []*write.Point
	metricsBatch      []*write.Point
	analyticsBatch    []*write.Point
	lastFlush         time.Time
}

// NewInfluxSink creates a new InfluxDB sink with improved batching
func NewInfluxSink(client influxdb2.Client, cfg *config.Config) *InfluxSink {
	sink := &InfluxSink{
		client:         client,
		cfg:            cfg,
		eventsBatch:    make([]*write.Point, 0, cfg.BatchSize),
		metricsBatch:   make([]*write.Point, 0, cfg.BatchSize),
		analyticsBatch: make([]*write.Point, 0, cfg.BatchSize),
		lastFlush:      time.Now(),
	}

	// Initialize write APIs
	sink.eventsWriteAPI = client.WriteAPI(cfg.InfluxOrg, cfg.InfluxRawEventsBucket)
	sink.metricsWriteAPI = client.WriteAPI(cfg.InfluxOrg, cfg.InfluxDerivedMetricsBucket)
	sink.analyticsWriteAPI = client.WriteAPI(cfg.InfluxOrg, cfg.InfluxAnalyticalDatasetsBucket)

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

	go func() {
		for err := range is.analyticsWriteAPI.Errors() {
			log.Printf("InfluxDB analytics write error: %v", err)
		}
	}()
}

// Process handles writing to InfluxDB with new optimized format
func (is *InfluxSink) Process(ctx context.Context, in <-chan interface{}) {
	log.Printf("Starting InfluxDB sink with buckets: events=%s, metrics=%s, analytics=%s",
		is.cfg.InfluxRawEventsBucket, is.cfg.InfluxDerivedMetricsBucket, is.cfg.InfluxAnalyticalDatasetsBucket)

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

	// Create analytical datasets
	analyticalPoints := is.createAnalyticalPoints(processedEvent)
	is.analyticsBatch = append(is.analyticsBatch, analyticalPoints...)

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

// createAnalyticalPoints creates comprehensive analytical datasets for OLAP and business intelligence
func (is *InfluxSink) createAnalyticalPoints(event models.ProcessedEvent) []*write.Point {
	points := make([]*write.Point, 0, 8)

	// 1. Citizen Behavior Pattern Analysis
	behaviorPoint := is.createBehaviorAnalysisPoint(event)
	points = append(points, behaviorPoint)

	// 2. Cross-Citizen Comparative Analysis
	comparativePoint := is.createComparativeAnalysisPoint(event)
	points = append(points, comparativePoint)

	// 3. Time-based Trend Analysis
	trendPoint := is.createTrendAnalysisPoint(event)
	points = append(points, trendPoint)

	// 4. Statistical Aggregations for BI
	statsPoint := is.createStatisticalAnalysisPoint(event)
	points = append(points, statsPoint)

	// 5. Geographic Analytics
	geoPoint := is.createGeographicAnalysisPoint(event)
	points = append(points, geoPoint)

	// 6. Risk Assessment Analytics
	riskPoint := is.createRiskAnalysisPoint(event)
	points = append(points, riskPoint)

	// 7. System Performance Analytics
	perfPoint := is.createPerformanceAnalysisPoint(event)
	points = append(points, perfPoint)

	return points
}

// createBehaviorAnalysisPoint creates points for citizen behavioral pattern analysis
func (is *InfluxSink) createBehaviorAnalysisPoint(event models.ProcessedEvent) *write.Point {
	tags := map[string]string{
		"citizen_id":       event.CitizenID,
		"event_type":       event.EventType,
		"event_subtype":    event.EventSubtype,
		"behavior_pattern": is.categorizeBehaviorPattern(event),
		"time_of_day":      is.getTimeOfDayCategory(event.Timestamp),
		"day_of_week":      event.Timestamp.Weekday().String(),
		"month":            event.Timestamp.Month().String(),
		"processing_id":    event.ProcessingID,
	}

	fields := map[string]interface{}{
		"score_impact":          event.ScoreChange,
		"cumulative_score":      event.NewScore,
		"confidence_level":      event.Confidence,
		"rules_triggered_count": len(event.AppliedRules),
		"frequency_this_hour":   1, // Will be aggregated
		"frequency_this_day":    1, // Will be aggregated
		"behavior_consistency":  is.calculateBehaviorConsistency(event),
		"location_mobility":     is.calculateLocationMobility(event),
		"event_sequence_number": 1, // For sequential analysis
	}

	return influxdb2.NewPoint("citizen_behavior_analysis", tags, fields, event.Timestamp)
}

// createComparativeAnalysisPoint creates points for cross-citizen comparative analysis
func (is *InfluxSink) createComparativeAnalysisPoint(event models.ProcessedEvent) *write.Point {
	tags := map[string]string{
		"event_type":       event.EventType,
		"score_tier":       is.getScoreTier(event.NewScore),
		"score_change_dir": is.getScoreChangeDirection(event.ScoreChange),
		"region":           event.Location.Region,
		"country":          event.Location.Country,
		"time_bucket":      is.getTimeBucket(event.Timestamp),
		"processing_id":    event.ProcessingID,
	}

	fields := map[string]interface{}{
		"citizen_count":      1, // For unique citizen counting
		"total_score_change": event.ScoreChange,
		"avg_score_change":   event.ScoreChange, // Will be aggregated
		"max_score_change":   event.ScoreChange,
		"min_score_change":   event.ScoreChange,
		"score_variance":     0.0, // For statistical analysis
		"percentile_rank":    is.calculatePercentileRank(event.NewScore),
		"z_score":            is.calculateZScore(event.ScoreChange),
		"deviation_from_avg": 0.0,            // Calculated during aggregation
		"regional_score_avg": event.NewScore, // Will be aggregated by region
	}

	return influxdb2.NewPoint("cross_citizen_comparison", tags, fields, event.Timestamp)
}

// createTrendAnalysisPoint creates points for time-based trend analysis
func (is *InfluxSink) createTrendAnalysisPoint(event models.ProcessedEvent) *write.Point {
	tags := map[string]string{
		"trend_category":     is.getTrendCategory(event),
		"time_window":        is.getTimeWindow(event.Timestamp),
		"event_type":         event.EventType,
		"geographic_segment": fmt.Sprintf("%s_%s", event.Location.Country, event.Location.Region),
		"score_tier":         is.getScoreTier(event.NewScore),
		"processing_id":      event.ProcessingID,
	}

	fields := map[string]interface{}{
		"trend_direction":       is.getTrendDirection(event.ScoreChange),
		"trend_magnitude":       abs(event.ScoreChange),
		"trend_velocity":        is.calculateTrendVelocity(event),
		"trend_acceleration":    0.0,            // Calculated from historical data
		"moving_average_7d":     event.NewScore, // Will be calculated
		"moving_average_30d":    event.NewScore, // Will be calculated
		"seasonal_adjustment":   1.0,            // Seasonal factor
		"trend_stability_index": is.calculateTrendStability(event),
		"forecast_confidence":   0.85, // Model confidence for forecasting
	}

	return influxdb2.NewPoint("trend_analysis", tags, fields, event.Timestamp)
}

// createStatisticalAnalysisPoint creates points for advanced statistical analysis
func (is *InfluxSink) createStatisticalAnalysisPoint(event models.ProcessedEvent) *write.Point {
	tags := map[string]string{
		"statistical_model": "social_credit_analytics",
		"data_segment":      is.getDataSegment(event),
		"analysis_type":     "real_time_stats",
		"event_type":        event.EventType,
		"region":            event.Location.Region,
		"processing_id":     event.ProcessingID,
	}

	fields := map[string]interface{}{
		"sample_size":            1,
		"mean_score":             float64(event.NewScore),
		"median_score":           float64(event.NewScore),
		"mode_score":             float64(event.NewScore),
		"std_deviation":          0.0, // Calculated during aggregation
		"variance":               0.0,
		"skewness":               0.0, // Distribution skewness
		"kurtosis":               0.0, // Distribution kurtosis
		"confidence_interval":    0.95,
		"p_value":                0.05, // For statistical significance
		"correlation_score":      0.0,  // Correlation with other metrics
		"regression_coefficient": 0.0,  // For predictive modeling
	}

	return influxdb2.NewPoint("statistical_analysis", tags, fields, event.Timestamp)
}

// createGeographicAnalysisPoint creates points for geographic and spatial analytics
func (is *InfluxSink) createGeographicAnalysisPoint(event models.ProcessedEvent) *write.Point {
	tags := map[string]string{
		"country":         event.Location.Country,
		"region":          event.Location.Region,
		"city":            event.Location.City,
		"geographic_tier": is.getGeographicTier(event.Location),
		"urban_rural":     is.getUrbanRuralClassification(event.Location),
		"processing_id":   event.ProcessingID,
	}

	fields := map[string]interface{}{
		"latitude":               event.Location.Latitude,
		"longitude":              event.Location.Longitude,
		"geographic_density":     1, // Events per geographic area
		"regional_score_avg":     float64(event.NewScore),
		"spatial_correlation":    0.0, // Correlation with neighboring areas
		"geographic_mobility":    is.calculateGeographicMobility(event),
		"distance_from_center":   is.calculateDistanceFromCenter(event.Location),
		"population_density_est": is.estimatePopulationDensity(event.Location),
		"economic_indicator":     is.getEconomicIndicator(event.Location),
	}

	return influxdb2.NewPoint("geographic_analysis", tags, fields, event.Timestamp)
}

// createRiskAnalysisPoint creates points for risk assessment and prediction
func (is *InfluxSink) createRiskAnalysisPoint(event models.ProcessedEvent) *write.Point {
	tags := map[string]string{
		"risk_category":   is.getRiskCategory(event),
		"risk_level":      is.getRiskLevel(event.NewScore),
		"event_type":      event.EventType,
		"citizen_segment": is.getCitizenSegment(event),
		"alert_level":     is.getAlertLevel(event),
		"processing_id":   event.ProcessingID,
	}

	fields := map[string]interface{}{
		"risk_score":              is.calculateRiskScore(event),
		"probability_default":     is.calculateDefaultProbability(event),
		"early_warning_signal":    is.getEarlyWarningSignal(event),
		"risk_mitigation_factor":  is.getRiskMitigationFactor(event),
		"fraud_detection_score":   is.calculateFraudScore(event),
		"anomaly_detection_score": is.calculateAnomalyScore(event),
		"predictive_risk_30d":     is.calculatePredictiveRisk(event, 30),
		"predictive_risk_90d":     is.calculatePredictiveRisk(event, 90),
		"volatility_index":        is.calculateVolatilityIndex(event),
	}

	return influxdb2.NewPoint("risk_analysis", tags, fields, event.Timestamp)
}

// createPerformanceAnalysisPoint creates points for system performance analytics
func (is *InfluxSink) createPerformanceAnalysisPoint(event models.ProcessedEvent) *write.Point {
	tags := map[string]string{
		"system_component": "stream_processor",
		"processing_stage": "event_analysis",
		"data_source":      event.SourceSystem,
		"processing_id":    event.ProcessingID,
	}

	processingTime := event.ProcessedAt.Sub(event.Timestamp)

	fields := map[string]interface{}{
		"processing_latency_ms": processingTime.Milliseconds(),
		"processing_latency_us": processingTime.Microseconds(),
		"throughput_events_sec": 1.0,   // Will be aggregated
		"memory_usage_mb":       0.0,   // System metrics
		"cpu_usage_percent":     0.0,   // System metrics
		"queue_depth":           0,     // Processing queue metrics
		"error_rate":            0.0,   // Error percentage
		"success_rate":          100.0, // Success percentage
		"data_quality_score":    is.calculateDataQualityScore(event),
		"processing_efficiency": is.calculateProcessingEfficiency(event),
	}

	if event.ProcessingError != "" {
		fields["error_rate"] = 100.0
		fields["success_rate"] = 0.0
		tags["error_type"] = "processing_error"
	}

	return influxdb2.NewPoint("system_performance", tags, fields, event.ProcessedAt)
}

// Helper methods for analytical calculations

func (is *InfluxSink) categorizeBehaviorPattern(event models.ProcessedEvent) string {
	// Categorize behavior patterns based on event type and score change
	if event.ScoreChange > 0 {
		return "positive_behavior"
	} else if event.ScoreChange < 0 {
		return "negative_behavior"
	}
	return "neutral_behavior"
}

func (is *InfluxSink) getTimeOfDayCategory(t time.Time) string {
	hour := t.Hour()
	switch {
	case hour >= 6 && hour < 12:
		return "morning"
	case hour >= 12 && hour < 18:
		return "afternoon"
	case hour >= 18 && hour < 22:
		return "evening"
	default:
		return "night"
	}
}

func (is *InfluxSink) calculateBehaviorConsistency(event models.ProcessedEvent) float64 {
	// Calculate behavior consistency based on event patterns and score changes
	consistency := 0.5 // Base consistency

	// Higher consistency for smaller score changes (more predictable behavior)
	if abs(event.ScoreChange) <= 10 {
		consistency += 0.3
	} else if abs(event.ScoreChange) <= 50 {
		consistency += 0.2
	} else if abs(event.ScoreChange) <= 100 {
		consistency += 0.1
	}

	// Event type patterns affect consistency
	switch event.EventType {
	case "financial":
		consistency += 0.1 // Financial events are more consistent
	case "social":
		consistency -= 0.1 // Social events are less predictable
	case "legal":
		consistency -= 0.2 // Legal events indicate inconsistency
	case "civic":
		consistency += 0.15 // Civic engagement shows consistency
	}

	// High confidence events indicate consistent behavior
	if event.Confidence > 0.9 {
		consistency += 0.1
	} else if event.Confidence < 0.7 {
		consistency -= 0.1
	}

	// Clamp between 0 and 1
	if consistency > 1.0 {
		consistency = 1.0
	} else if consistency < 0.0 {
		consistency = 0.0
	}

	return consistency
}

func (is *InfluxSink) calculateLocationMobility(event models.ProcessedEvent) float64 {
	// Calculate location mobility based on geographic patterns
	mobility := 0.3 // Base mobility

	// Time of day affects mobility patterns
	hour := event.Timestamp.Hour()
	switch {
	case hour >= 7 && hour <= 9: // Morning commute
		mobility += 0.3
	case hour >= 17 && hour <= 19: // Evening commute
		mobility += 0.3
	case hour >= 12 && hour <= 14: // Lunch time
		mobility += 0.2
	case hour >= 22 || hour <= 6: // Night/early morning
		mobility -= 0.1
	}

	// Weekend vs weekday patterns
	if event.Timestamp.Weekday() == 0 || event.Timestamp.Weekday() == 6 { // Weekend
		mobility += 0.2
	}

	// Event types that indicate mobility
	switch event.EventType {
	case "transportation":
		mobility += 0.4
	case "financial":
		if subtype, ok := event.Payload["transaction_type"].(string); ok {
			if subtype == "atm_withdrawal" || subtype == "card_payment" {
				mobility += 0.3
			}
		}
	case "social":
		mobility += 0.2
	case "civic":
		mobility += 0.1
	}

	// Clamp between 0 and 1
	if mobility > 1.0 {
		mobility = 1.0
	} else if mobility < 0.0 {
		mobility = 0.0
	}

	return mobility
}

func (is *InfluxSink) getScoreTier(score float64) string {
	switch {
	case score >= 900:
		return "excellent"
	case score >= 800:
		return "very_good"
	case score >= 700:
		return "good"
	case score >= 600:
		return "fair"
	case score >= 500:
		return "poor"
	default:
		return "very_poor"
	}
}

func (is *InfluxSink) getScoreChangeDirection(change float64) string {
	if change > 0 {
		return "increase"
	} else if change < 0 {
		return "decrease"
	}
	return "unchanged"
}

func (is *InfluxSink) getTimeBucket(t time.Time) string {
	return fmt.Sprintf("%04d-%02d-%02d_%02d", t.Year(), t.Month(), t.Day(), t.Hour())
}

func (is *InfluxSink) calculatePercentileRank(score float64) float64 {
	// Simplified percentile calculation - would use historical data in production
	return score / 1000.0 * 100.0
}

func (is *InfluxSink) calculateZScore(scoreChange float64) float64 {
	// Simplified Z-score calculation
	return scoreChange / 50.0 // Assuming std dev of 50
}

func (is *InfluxSink) getTrendCategory(event models.ProcessedEvent) string {
	if abs(event.ScoreChange) >= 100 {
		return "significant_change"
	} else if abs(event.ScoreChange) >= 50 {
		return "moderate_change"
	}
	return "minor_change"
}

func (is *InfluxSink) getTimeWindow(t time.Time) string {
	return fmt.Sprintf("%04d-W%02d", t.Year(), getWeekNumber(t))
}

func (is *InfluxSink) getTrendDirection(scoreChange float64) float64 {
	if scoreChange > 0 {
		return 1.0
	} else if scoreChange < 0 {
		return -1.0
	}
	return 0.0
}

func (is *InfluxSink) calculateTrendVelocity(event models.ProcessedEvent) float64 {
	// Calculate trend velocity based on score change magnitude and time factors
	velocity := float64(abs(event.ScoreChange)) / 100.0 // Base velocity from score change

	// Adjust velocity based on event confidence
	if event.Confidence > 0.9 {
		velocity *= 1.2 // High confidence events have more reliable velocity
	} else if event.Confidence < 0.7 {
		velocity *= 0.8 // Low confidence events have dampened velocity
	}

	// Event type affects trend velocity
	switch event.EventType {
	case "legal":
		velocity *= 1.5 // Legal events create rapid trend changes
	case "financial":
		if transactionType, ok := event.Payload["transaction_type"].(string); ok {
			switch transactionType {
			case "bankruptcy", "default":
				velocity *= 2.0 // Major financial events have high velocity
			case "large_transfer", "major_purchase":
				velocity *= 1.3 // Significant transactions have higher velocity
			case "regular_payment":
				velocity *= 0.7 // Regular payments have lower velocity
			}
		}
	case "civic":
		velocity *= 1.1 // Civic engagement creates moderate velocity
	case "social":
		velocity *= 0.9 // Social events have slightly lower velocity
	case "system":
		if event.EventSubtype == "new_citizen" {
			velocity *= 0.5 // New citizen events have lower initial velocity
		}
	}

	// Time of day affects velocity (business hours vs off-hours)
	hour := event.Timestamp.Hour()
	if hour >= 9 && hour <= 17 {
		velocity *= 1.1 // Business hours have higher velocity
	} else if hour >= 22 || hour <= 6 {
		velocity *= 0.8 // Late night/early morning have lower velocity
	}

	// Score tier affects velocity (higher scores have more stable trends)
	tier := is.getScoreTier(event.NewScore)
	switch tier {
	case "excellent":
		velocity *= 0.8 // Excellent scores have more stable velocity
	case "very_good", "good":
		velocity *= 0.9 // Good scores have slightly more stable velocity
	case "poor", "very_poor":
		velocity *= 1.3 // Poor scores have more volatile velocity
	}

	// Cap velocity at reasonable maximum
	if velocity > 5.0 {
		velocity = 5.0
	}

	return velocity
}

func (is *InfluxSink) calculateTrendStability(event models.ProcessedEvent) float64 {
	// Calculate trend stability based on score patterns and behavioral consistency
	stability := 0.7 // Base stability

	// Score change magnitude affects stability (smaller changes = more stable)
	changeAbs := abs(event.ScoreChange)
	if changeAbs <= 10 {
		stability += 0.2 // Very small changes indicate high stability
	} else if changeAbs <= 25 {
		stability += 0.1 // Small changes indicate good stability
	} else if changeAbs <= 50 {
		stability -= 0.1 // Medium changes reduce stability
	} else if changeAbs <= 100 {
		stability -= 0.2 // Large changes significantly reduce stability
	} else {
		stability -= 0.3 // Very large changes indicate instability
	}

	// Event confidence affects stability
	if event.Confidence > 0.9 {
		stability += 0.15 // High confidence indicates stable data
	} else if event.Confidence < 0.7 {
		stability -= 0.15 // Low confidence indicates unstable data
	}

	// Current score tier affects stability
	tier := is.getScoreTier(event.NewScore)
	switch tier {
	case "excellent":
		stability += 0.2 // Excellent citizens have very stable trends
	case "very_good":
		stability += 0.15 // Very good citizens have stable trends
	case "good":
		stability += 0.1 // Good citizens have fairly stable trends
	case "fair":
		stability -= 0.05 // Fair citizens have slightly less stable trends
	case "poor":
		stability -= 0.15 // Poor citizens have unstable trends
	case "very_poor":
		stability -= 0.25 // Very poor citizens have very unstable trends
	}

	// Event type patterns affect stability
	switch event.EventType {
	case "civic":
		stability += 0.1 // Civic engagement indicates stable behavior
	case "financial":
		if transactionType, ok := event.Payload["transaction_type"].(string); ok {
			switch transactionType {
			case "regular_payment", "salary_deposit", "bill_payment":
				stability += 0.15 // Regular financial behavior is stable
			case "savings_deposit":
				stability += 0.1 // Saving money indicates stability
			case "large_withdrawal", "cash_advance":
				stability -= 0.2 // Large withdrawals indicate instability
			case "loan_default", "bankruptcy":
				stability -= 0.4 // Defaults indicate severe instability
			}
		}
	case "legal":
		stability -= 0.3 // Legal issues indicate instability
	case "social":
		if event.ScoreChange > 0 {
			stability += 0.05 // Positive social events add stability
		} else {
			stability -= 0.1 // Negative social events reduce stability
		}
	case "system":
		if event.EventSubtype == "data_correction" {
			stability -= 0.1 // Data corrections indicate some instability
		}
	}

	// Time patterns affect stability
	dayOfWeek := event.Timestamp.Weekday()
	if dayOfWeek == 0 || dayOfWeek == 6 { // Weekend
		stability -= 0.05 // Weekend events are slightly less stable
	}

	// Behavioral consistency from earlier calculation
	behaviorConsistency := is.calculateBehaviorConsistency(event)
	stability += (behaviorConsistency - 0.5) * 0.3 // Adjust based on consistency

	// Processing errors indicate instability
	if event.ProcessingError != "" {
		stability -= 0.2
	}

	// Clamp between 0 and 1
	if stability > 1.0 {
		stability = 1.0
	} else if stability < 0.0 {
		stability = 0.0
	}

	return stability
}

func (is *InfluxSink) getDataSegment(event models.ProcessedEvent) string {
	return fmt.Sprintf("%s_%s", event.EventType, is.getScoreTier(event.NewScore))
}

func (is *InfluxSink) getGeographicTier(location models.EventLocation) string {
	// Categorize geographic areas by size/importance based on city names and population patterns
	city := strings.ToLower(location.City)

	// Major metropolitan areas (Tier 1)
	majorCities := []string{"new york", "london", "paris", "tokyo", "berlin", "moscow", "beijing", "shanghai", "mumbai", "delhi", "los angeles", "chicago", "toronto", "sydney", "melbourne"}
	for _, majorCity := range majorCities {
		if strings.Contains(city, majorCity) {
			return "tier_1"
		}
	}

	// Regional centers and large cities (Tier 2)
	regionalCities := []string{"manchester", "birmingham", "glasgow", "lyon", "marseille", "frankfurt", "hamburg", "munich", "barcelona", "madrid", "milan", "rome", "amsterdam", "rotterdam"}
	for _, regionalCity := range regionalCities {
		if strings.Contains(city, regionalCity) {
			return "tier_2"
		}
	}

	// Medium cities and towns (Tier 3)
	if location.Country == "United States" || location.Country == "United Kingdom" || location.Country == "Germany" || location.Country == "France" {
		return "tier_3"
	}

	// Small towns and rural areas (Tier 4)
	return "tier_4"
}

func (is *InfluxSink) getUrbanRuralClassification(location models.EventLocation) string {
	// Classify as urban or rural based on location characteristics
	city := strings.ToLower(location.City)

	// Major urban indicators
	urbanKeywords := []string{"city", "york", "london", "paris", "berlin", "tokyo", "center", "downtown", "metro", "district"}
	for _, keyword := range urbanKeywords {
		if strings.Contains(city, keyword) {
			return "urban"
		}
	}

	// Rural indicators
	ruralKeywords := []string{"village", "farm", "county", "rural", "countryside", "valley", "hill", "mountain", "lake"}
	for _, keyword := range ruralKeywords {
		if strings.Contains(city, keyword) {
			return "rural"
		}
	}

	// Suburban indicators
	suburbanKeywords := []string{"suburb", "gardens", "heights", "park", "grove", "meadow", "ridge", "estates"}
	for _, keyword := range suburbanKeywords {
		if strings.Contains(city, keyword) {
			return "suburban"
		}
	}

	// Default classification based on population density estimation
	// If we have specific coordinates, we could do more sophisticated classification
	if location.Latitude != 0 && location.Longitude != 0 {
		// Areas near major coordinate clusters are likely urban
		return "urban"
	}

	return "suburban" // Default
}

func (is *InfluxSink) calculateGeographicMobility(event models.ProcessedEvent) float64 {
	// Calculate geographic mobility based on event patterns and location changes
	mobility := 0.2 // Base mobility

	// Transportation events indicate high mobility
	if event.EventType == "transportation" {
		mobility += 0.6
		// Different transportation types have different mobility levels
		if transportType, ok := event.Payload["transport_type"].(string); ok {
			switch transportType {
			case "flight", "train_long_distance":
				mobility += 0.3 // High mobility
			case "bus", "train_local", "metro":
				mobility += 0.2 // Medium mobility
			case "taxi", "rideshare":
				mobility += 0.1 // Local mobility
			}
		}
	}

	// Financial events at different locations indicate mobility
	if event.EventType == "financial" {
		if transactionType, ok := event.Payload["transaction_type"].(string); ok {
			switch transactionType {
			case "atm_withdrawal", "card_payment":
				mobility += 0.3 // Person is moving around
			case "online_purchase":
				mobility -= 0.1 // Staying at home
			}
		}
	}

	// Social events indicate movement to social gatherings
	if event.EventType == "social" {
		mobility += 0.2
	}

	// Civic events may require travel to specific locations
	if event.EventType == "civic" {
		mobility += 0.15
	}

	// Time patterns affect mobility
	hour := event.Timestamp.Hour()
	if hour >= 7 && hour <= 9 || hour >= 17 && hour <= 19 {
		mobility += 0.2 // Commute times
	}

	// Clamp between 0 and 1
	if mobility > 1.0 {
		mobility = 1.0
	} else if mobility < 0.0 {
		mobility = 0.0
	}

	return mobility
}

func (is *InfluxSink) calculateDistanceFromCenter(location models.EventLocation) float64 {
	// Calculate approximate distance from major city centers
	// Using simplified geographic calculations

	var centerLat, centerLon float64

	// Define major city centers based on country
	switch location.Country {
	case "United States":
		centerLat, centerLon = 39.8283, -98.5795 // Geographic center of US
	case "United Kingdom":
		centerLat, centerLon = 54.5973, -2.7081 // Geographic center of UK
	case "Germany":
		centerLat, centerLon = 51.1657, 10.4515 // Geographic center of Germany
	case "France":
		centerLat, centerLon = 46.6034, 1.8883 // Geographic center of France
	case "Canada":
		centerLat, centerLon = 56.1304, -106.3468 // Geographic center of Canada
	default:
		// Use location coordinates as reference if available
		if location.Latitude != 0 && location.Longitude != 0 {
			return 50.0 // Default distance for unknown countries
		}
		return 100.0
	}

	// Calculate approximate distance using simplified formula
	if location.Latitude == 0 && location.Longitude == 0 {
		return 150.0 // Default for missing coordinates
	}

	// Simplified distance calculation (Haversine approximation)
	deltaLat := location.Latitude - centerLat
	deltaLon := location.Longitude - centerLon
	distance := 111.32 * (deltaLat*deltaLat + deltaLon*deltaLon*0.5) // Rough km calculation

	if distance < 0 {
		distance = -distance
	}

	return distance
}

func (is *InfluxSink) estimatePopulationDensity(location models.EventLocation) float64 {
	// Estimate population density based on geographic tier and urban classification
	tier := is.getGeographicTier(location)
	classification := is.getUrbanRuralClassification(location)

	density := 100.0 // Base density per km²

	// Adjust by geographic tier
	switch tier {
	case "tier_1": // Major metropolitan areas
		density = 8000.0
	case "tier_2": // Regional centers
		density = 3000.0
	case "tier_3": // Medium cities
		density = 1200.0
	case "tier_4": // Small towns
		density = 400.0
	}

	// Adjust by urban/rural classification
	switch classification {
	case "urban":
		density *= 1.5 // Urban areas are denser
	case "suburban":
		density *= 0.7 // Suburban areas are less dense
	case "rural":
		density *= 0.2 // Rural areas are much less dense
	}

	// Country-specific adjustments
	switch location.Country {
	case "Japan", "South Korea", "Singapore":
		density *= 2.0 // Very high density countries
	case "Netherlands", "Belgium":
		density *= 1.5 // High density countries
	case "Canada", "Australia", "Russia":
		density *= 0.3 // Low density countries
	case "United States":
		density *= 0.8 // Medium-low density
	}

	return density
}

func (is *InfluxSink) getEconomicIndicator(location models.EventLocation) float64 {
	// Get economic indicator based on country and region
	indicator := 50.0 // Base economic index (0-100 scale)

	// Country-level economic indicators (simplified)
	switch location.Country {
	case "United States", "Canada":
		indicator = 85.0
	case "Germany", "United Kingdom", "France", "Netherlands":
		indicator = 80.0
	case "Japan", "South Korea", "Australia":
		indicator = 78.0
	case "Italy", "Spain":
		indicator = 70.0
	case "Russia", "Brazil":
		indicator = 60.0
	case "China", "India":
		indicator = 65.0
	default:
		indicator = 55.0
	}

	// Regional adjustments within countries
	region := strings.ToLower(location.Region)
	if strings.Contains(region, "capital") || strings.Contains(region, "central") {
		indicator += 10.0 // Capital regions are usually more prosperous
	}

	// Urban areas typically have higher economic indicators
	classification := is.getUrbanRuralClassification(location)
	switch classification {
	case "urban":
		indicator += 5.0
	case "rural":
		indicator -= 10.0
	}

	// Geographic tier affects economic activity
	tier := is.getGeographicTier(location)
	switch tier {
	case "tier_1":
		indicator += 8.0
	case "tier_2":
		indicator += 3.0
	case "tier_4":
		indicator -= 5.0
	}

	// Clamp between 0 and 100
	if indicator > 100.0 {
		indicator = 100.0
	} else if indicator < 0.0 {
		indicator = 0.0
	}

	return indicator
}

func (is *InfluxSink) getRiskCategory(event models.ProcessedEvent) string {
	if event.ScoreChange < -50 {
		return "high_risk"
	} else if event.ScoreChange < -20 {
		return "medium_risk"
	}
	return "low_risk"
}

func (is *InfluxSink) getRiskLevel(score float64) string {
	return is.getScoreTier(score) // Reuse score tier logic
}

func (is *InfluxSink) getCitizenSegment(event models.ProcessedEvent) string {
	return fmt.Sprintf("segment_%s", is.getScoreTier(event.NewScore))
}

func (is *InfluxSink) getAlertLevel(event models.ProcessedEvent) string {
	if abs(event.ScoreChange) >= 100 {
		return "critical"
	} else if abs(event.ScoreChange) >= 50 {
		return "warning"
	}
	return "normal"
}

func (is *InfluxSink) calculateRiskScore(event models.ProcessedEvent) float64 {
	// Calculate composite risk score
	baseRisk := float64(1000-event.NewScore) / 1000.0
	changeRisk := float64(abs(event.ScoreChange)) / 100.0
	return (baseRisk + changeRisk) / 2.0
}

func (is *InfluxSink) calculateDefaultProbability(event models.ProcessedEvent) float64 {
	// Calculate probability of default based on score
	return (1000.0 - float64(event.NewScore)) / 1000.0
}

func (is *InfluxSink) getEarlyWarningSignal(event models.ProcessedEvent) float64 {
	if event.ScoreChange < -75 {
		return 1.0 // High warning
	} else if event.ScoreChange < -25 {
		return 0.5 // Medium warning
	}
	return 0.0 // No warning
}

func (is *InfluxSink) getRiskMitigationFactor(event models.ProcessedEvent) float64 {
	// Calculate risk mitigation factor based on citizen behavior and event characteristics
	mitigation := 0.5 // Base mitigation factor

	// Higher scores indicate better risk mitigation
	if event.NewScore >= 800 {
		mitigation += 0.3 // Excellent citizens have good risk mitigation
	} else if event.NewScore >= 600 {
		mitigation += 0.2 // Good citizens have decent mitigation
	} else if event.NewScore >= 400 {
		mitigation += 0.1 // Average citizens have some mitigation
	} else {
		mitigation -= 0.1 // Poor scores indicate less mitigation
	}

	// Positive score changes indicate improving behavior (better mitigation)
	if event.ScoreChange > 0 {
		mitigation += 0.15
	} else if event.ScoreChange < -50 {
		mitigation -= 0.2 // Large negative changes reduce mitigation
	}

	// High confidence events indicate reliable data (better mitigation)
	if event.Confidence > 0.9 {
		mitigation += 0.1
	} else if event.Confidence < 0.7 {
		mitigation -= 0.1
	}

	// Event types that indicate responsible behavior
	switch event.EventType {
	case "civic":
		mitigation += 0.2 // Civic engagement indicates responsibility
	case "financial":
		if transactionType, ok := event.Payload["transaction_type"].(string); ok {
			switch transactionType {
			case "loan_payment", "bill_payment":
				mitigation += 0.15 // Paying bills indicates responsibility
			case "savings_deposit":
				mitigation += 0.1 // Saving money is good behavior
			}
		}
	case "legal":
		mitigation -= 0.3 // Legal issues reduce mitigation
	case "social":
		mitigation += 0.05 // Social engagement can be positive
	}

	// Clamp between 0 and 1
	if mitigation > 1.0 {
		mitigation = 1.0
	} else if mitigation < 0.0 {
		mitigation = 0.0
	}

	return mitigation
}

func (is *InfluxSink) calculateFraudScore(event models.ProcessedEvent) float64 {
	// Calculate fraud detection score based on event patterns and anomalies
	fraudScore := 0.05 // Base fraud probability (low)

	// Large score changes can indicate fraudulent activity
	if abs(event.ScoreChange) >= 200 {
		fraudScore += 0.4 // Very large changes are suspicious
	} else if abs(event.ScoreChange) >= 100 {
		fraudScore += 0.2 // Large changes are somewhat suspicious
	}

	// Low confidence events might indicate data issues or fraud
	if event.Confidence < 0.5 {
		fraudScore += 0.3
	} else if event.Confidence < 0.7 {
		fraudScore += 0.1
	}

	// Certain event types are more prone to fraud
	switch event.EventType {
	case "financial":
		if transactionType, ok := event.Payload["transaction_type"].(string); ok {
			switch transactionType {
			case "large_transfer", "cash_advance":
				fraudScore += 0.2 // High-risk transaction types
			case "atm_withdrawal":
				if amount, ok := event.Payload["amount"].(float64); ok && amount > 1000 {
					fraudScore += 0.15 // Large cash withdrawals
				}
			case "online_purchase":
				fraudScore += 0.05 // Online purchases have some fraud risk
			}
		}
	case "identity":
		fraudScore += 0.25 // Identity-related events are high risk
	case "system":
		if event.EventSubtype == "data_correction" {
			fraudScore += 0.15 // Data corrections might indicate manipulation
		}
	}

	// Time-based anomalies
	hour := event.Timestamp.Hour()
	if hour >= 2 && hour <= 5 {
		fraudScore += 0.1 // Unusual hours increase fraud risk
	}

	// Geographic anomalies (simplified)
	if event.Location.Country == "" || event.Location.City == "" {
		fraudScore += 0.1 // Missing location data is suspicious
	}

	// Processing errors might indicate system manipulation
	if event.ProcessingError != "" {
		fraudScore += 0.2
	}

	// Clamp between 0 and 1
	if fraudScore > 1.0 {
		fraudScore = 1.0
	} else if fraudScore < 0.0 {
		fraudScore = 0.0
	}

	return fraudScore
}

func (is *InfluxSink) calculateAnomalyScore(event models.ProcessedEvent) float64 {
	// Calculate anomaly detection score
	if abs(event.ScoreChange) >= 200 {
		return 0.9 // High anomaly
	} else if abs(event.ScoreChange) >= 100 {
		return 0.5 // Medium anomaly
	}
	return 0.1 // Low anomaly
}

func (is *InfluxSink) calculatePredictiveRisk(event models.ProcessedEvent, days int) float64 {
	// Calculate predictive risk for future period
	currentRisk := is.calculateRiskScore(event)
	timeDecay := 1.0 - (float64(days) / 365.0 * 0.2) // Risk decreases over time
	return currentRisk * timeDecay
}

func (is *InfluxSink) calculateVolatilityIndex(event models.ProcessedEvent) float64 {
	// Calculate score volatility index
	return float64(abs(event.ScoreChange)) / 50.0 // Normalized volatility
}

func (is *InfluxSink) calculateDataQualityScore(event models.ProcessedEvent) float64 {
	score := 100.0
	if event.Confidence < 0.8 {
		score -= 20.0
	}
	if event.ProcessingError != "" {
		score -= 50.0
	}
	if len(event.Payload) == 0 {
		score -= 10.0
	}
	return score
}

func (is *InfluxSink) calculateProcessingEfficiency(event models.ProcessedEvent) float64 {
	// Calculate processing efficiency based on time and resources
	processingTime := event.ProcessedAt.Sub(event.Timestamp)
	if processingTime.Milliseconds() < 100 {
		return 100.0
	} else if processingTime.Milliseconds() < 500 {
		return 80.0
	}
	return 60.0
}

// Utility functions
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

func getWeekNumber(t time.Time) int {
	_, week := t.ISOWeek()
	return week
}

// flushIfNeeded flushes batches if they exceed size limits or time threshold
func (is *InfluxSink) flushIfNeeded(checkSize bool) {
	shouldFlush := false

	if checkSize {
		shouldFlush = len(is.eventsBatch) >= is.cfg.BatchSize ||
			len(is.metricsBatch) >= is.cfg.BatchSize ||
			len(is.analyticsBatch) >= is.cfg.BatchSize
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
	analyticsCount := len(is.analyticsBatch)

	if eventCount == 0 && metricCount == 0 && analyticsCount == 0 {
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

	// Write analytics batch
	if analyticsCount > 0 {
		for _, point := range is.analyticsBatch {
			is.analyticsWriteAPI.WritePoint(point)
		}
		is.analyticsBatch = is.analyticsBatch[:0]
	}

	is.lastFlush = time.Now()
	log.Printf("InfluxDB: Flushed %d events, %d metrics, %d analytics", eventCount, metricCount, analyticsCount)
}

// forceFlush ensures all remaining data is written before shutdown
func (is *InfluxSink) forceFlush() {
	is.flush()
	is.eventsWriteAPI.Flush()
	is.metricsWriteAPI.Flush()
	is.analyticsWriteAPI.Flush()
	log.Println("InfluxDB sink: Force flush completed")
}
