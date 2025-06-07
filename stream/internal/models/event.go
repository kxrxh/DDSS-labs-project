package models

import (
	"encoding/json"
	"time"
)

// Event represents a social credit system event (matches event-producer structure)
type Event struct {
	EventID      string                 `json:"event_id" bson:"event_id"`
	CitizenID    string                 `json:"citizen_id" bson:"citizen_id"`
	EventType    string                 `json:"event_type" bson:"event_type"`
	EventSubtype string                 `json:"event_subtype" bson:"event_subtype"`
	SourceSystem string                 `json:"source_system" bson:"source_system"`
	Timestamp    time.Time              `json:"timestamp" bson:"timestamp"`
	Location     EventLocation          `json:"location" bson:"location"`
	Confidence   float64                `json:"confidence" bson:"confidence"`
	Payload      map[string]interface{} `json:"payload" bson:"payload"`
	Version      int                    `json:"version" bson:"version"`
	EventName    string                 `json:"event_name" bson:"event_name"`
}

// EventLocation represents the location where an event occurred
type EventLocation struct {
	City      string  `json:"city" bson:"city"`
	Region    string  `json:"region" bson:"region"`
	Country   string  `json:"country" bson:"country"`
	Latitude  float64 `json:"latitude" bson:"latitude"`
	Longitude float64 `json:"longitude" bson:"longitude"`
}

// ProcessedEvent represents an event after processing with scoring information
type ProcessedEvent struct {
	Event           `bson:",inline"`
	ProcessedAt     time.Time `json:"processed_at" bson:"processed_at"`
	AppliedRules    []string  `json:"applied_rules" bson:"applied_rules"`
	ScoreChange     float64   `json:"score_change" bson:"score_change"`
	NewScore        float64   `json:"new_score" bson:"new_score"`
	PreviousScore   float64   `json:"previous_score" bson:"previous_score"`
	ProcessingID    string    `json:"processing_id" bson:"processing_id"`
	ProcessingError string    `json:"processing_error,omitempty" bson:"processing_error,omitempty"`
}

// Citizen represents a citizen profile in MongoDB
type Citizen struct {
	ID          string                 `json:"_id" bson:"_id"`
	Age         int                    `json:"age" bson:"age"`
	Score       float64                `json:"score" bson:"score"`
	Type        string                 `json:"type" bson:"type"`
	Status      string                 `json:"status" bson:"status"`
	CreatedAt   time.Time              `json:"created_at" bson:"created_at"`
	LastUpdated time.Time              `json:"last_updated" bson:"last_updated"`
	Tier        string                 `json:"tier,omitempty" bson:"tier,omitempty"`
	LastScoreAt *time.Time             `json:"last_score_at,omitempty" bson:"last_score_at,omitempty"`
	Metadata    map[string]interface{} `json:"metadata,omitempty" bson:"metadata,omitempty"`
}

// ScoringRule represents a scoring rule definition from MongoDB
type ScoringRule struct {
	ID          string                 `json:"_id" bson:"_id"`
	Name        string                 `json:"name" bson:"name"`
	EventType   string                 `json:"event_type" bson:"event_type"`
	Conditions  map[string]interface{} `json:"conditions" bson:"conditions"`
	Points      float64                `json:"points" bson:"points"`
	Multiplier  float64                `json:"multiplier" bson:"multiplier"`
	Cooldown    *time.Duration         `json:"cooldown,omitempty" bson:"cooldown,omitempty"`
	Status      string                 `json:"status" bson:"status"`
	ValidFrom   time.Time              `json:"valid_from" bson:"valid_from"`
	ValidTo     *time.Time             `json:"valid_to,omitempty" bson:"valid_to,omitempty"`
	CreatedAt   time.Time              `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at" bson:"updated_at"`
	Description string                 `json:"description,omitempty" bson:"description,omitempty"`
}

// SystemConfiguration represents the global system configuration
type SystemConfiguration struct {
	ID               string                 `json:"_id" bson:"_id"`
	BaselineScore    float64                `json:"baseline_score" bson:"baseline_score"`
	TierDefinitions  map[string]TierDef     `json:"tier_definitions" bson:"tier_definitions"`
	ActiveRuleSet    string                 `json:"active_rule_set" bson:"active_rule_set"`
	ScoreDecayPolicy map[string]interface{} `json:"score_decay_policy" bson:"score_decay_policy"`
	MaxScore         float64                `json:"max_score" bson:"max_score"`
	MinScore         float64                `json:"min_score" bson:"min_score"`
	UpdatedAt        time.Time              `json:"updated_at" bson:"updated_at"`
	Version          int                    `json:"version" bson:"version"`
}

// TierDef represents a scoring tier definition
type TierDef struct {
	MinScore    float64 `json:"min_score" bson:"min_score"`
	MaxScore    float64 `json:"max_score" bson:"max_score"`
	Name        string  `json:"name" bson:"name"`
	Color       string  `json:"color,omitempty" bson:"color,omitempty"`
	Description string  `json:"description,omitempty" bson:"description,omitempty"`
}

// ScoreChange represents a score change record for InfluxDB
type ScoreChange struct {
	CitizenID     string    `json:"citizen_id"`
	EventID       string    `json:"event_id"`
	EventType     string    `json:"event_type"`
	PreviousScore float64   `json:"previous_score"`
	NewScore      float64   `json:"new_score"`
	ScoreChange   float64   `json:"score_change"`
	AppliedRules  []string  `json:"applied_rules"`
	ProcessingID  string    `json:"processing_id"`
	Timestamp     time.Time `json:"timestamp"`
	PreviousTier  string    `json:"previous_tier,omitempty"`
	NewTier       string    `json:"new_tier,omitempty"`
	Location      string    `json:"location,omitempty"`
}

// RuleApplicationState represents cooldown state in ScyllaDB
type RuleApplicationState struct {
	CitizenID        string     `json:"citizen_id"`
	RuleID           string     `json:"rule_id"`
	LastApplied      time.Time  `json:"last_applied"`
	ApplicationCount int        `json:"application_count"`
	CooldownUntil    *time.Time `json:"cooldown_until,omitempty"`
}

// InfluxEventRecord represents the raw event structure for InfluxDB
type InfluxEventRecord struct {
	CitizenID    string                 `json:"citizen_id"`
	EventID      string                 `json:"event_id"`
	EventType    string                 `json:"event_type"`
	EventSubtype string                 `json:"event_subtype"`
	SourceSystem string                 `json:"source_system"`
	Confidence   float64                `json:"confidence"`
	City         string                 `json:"city"`
	Region       string                 `json:"region"`
	Country      string                 `json:"country"`
	Latitude     float64                `json:"latitude"`
	Longitude    float64                `json:"longitude"`
	Payload      map[string]interface{} `json:"payload"`
	Timestamp    time.Time              `json:"timestamp"`
}

// CitizenActivitySummary represents aggregated citizen activity for InfluxDB
type CitizenActivitySummary struct {
	CitizenID      string    `json:"citizen_id"`
	TimeWindow     time.Time `json:"time_window"`
	EventCount     int       `json:"event_count"`
	ScoreChange    float64   `json:"total_score_change"`
	AvgConfidence  float64   `json:"avg_confidence"`
	EventTypes     []string  `json:"event_types"`
	LocationCity   string    `json:"location_city"`
	LocationRegion string    `json:"location_region"`
}

// SystemActivitySummary represents system-wide activity aggregates
type SystemActivitySummary struct {
	TimeWindow        time.Time `json:"time_window"`
	TotalEvents       int       `json:"total_events"`
	TotalCitizens     int       `json:"total_citizens"`
	AvgScoreChange    float64   `json:"avg_score_change"`
	TopEventTypes     []string  `json:"top_event_types"`
	ActiveRegions     []string  `json:"active_regions"`
	ProcessingLatency float64   `json:"avg_processing_latency_ms"`
}

// ToJSON converts any model to JSON bytes
func (e *Event) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

func (pe *ProcessedEvent) ToJSON() ([]byte, error) {
	return json.Marshal(pe)
}

func (sc *ScoreChange) ToJSON() ([]byte, error) {
	return json.Marshal(sc)
}

// FromJSON parses JSON bytes into Event
func (e *Event) FromJSON(data []byte) error {
	return json.Unmarshal(data, e)
}

// CalculateTier determines the citizen's tier based on score and tier definitions
func (c *Citizen) CalculateTier(config *SystemConfiguration) string {
	for tierName, tierDef := range config.TierDefinitions {
		if c.Score >= tierDef.MinScore && c.Score <= tierDef.MaxScore {
			return tierName
		}
	}
	return "unknown"
}
