package events

import (
	"encoding/json"
	"time"
)

// Event is the interface that all event types must implement.
type Event interface {
	GetBaseEvent() BaseEvent
}

// EventLocation represents the geographic location of an event.
type EventLocation struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Region    string  `json:"region,omitempty"`
	City      string  `json:"city,omitempty"`
	District  string  `json:"district,omitempty"`
	Country   string  `json:"country,omitempty"`
}

// MarshalJSON ensures the EventLocation is properly serialized as a complex object
func (l EventLocation) MarshalJSON() ([]byte, error) {
	type Alias EventLocation
	return json.Marshal(&struct {
		Alias
	}{
		Alias: Alias(l),
	})
}

// UnmarshalJSON ensures the EventLocation is properly deserialized from a complex object
func (l *EventLocation) UnmarshalJSON(data []byte) error {
	type Alias EventLocation
	aux := &struct {
		*Alias
	}{
		Alias: (*Alias)(l),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	return nil
}

// BaseEvent contains common fields for all events, designed for a flat JSON structure.
type BaseEvent struct {
	EventID      string                 `json:"event_id"`
	CitizenID    string                 `json:"citizen_id"`
	EventType    string                 `json:"event_type"`
	EventSubtype string                 `json:"event_subtype,omitempty"`
	SourceSystem string                 `json:"source_system"`
	Timestamp    time.Time              `json:"timestamp"`
	Location     EventLocation          `json:"location"`
	Confidence   float64                `json:"confidence"`
	Payload      map[string]interface{} `json:"payload,omitempty"` // Use interface{} to allow different payload structures
	Version      int                    `json:"version"`
	EventName    string                 `json:"event_name,omitempty"`
}

// MarshalJSON ensures the BaseEvent is properly serialized with all fields and correct timestamp format.
func (b BaseEvent) MarshalJSON() ([]byte, error) {
	type Alias BaseEvent
	return json.Marshal(&struct {
		Alias
		Timestamp string `json:"timestamp"`
	}{
		Alias:     Alias(b),
		Timestamp: b.Timestamp.Format(time.RFC3339),
	})
}

// GetBaseEvent returns the base event fields.
func (b BaseEvent) GetBaseEvent() BaseEvent {
	return b
}

// NewCitizen represents a new citizen joining the system.
type NewCitizen struct {
	BaseEvent
}

// SocialMediaPost represents a social media activity.
type SocialMediaPost struct {
	BaseEvent
}

// CommunityService represents community service activities.
type CommunityService struct {
	BaseEvent
}

// TaxPayment represents tax payment events.
type TaxPayment struct {
	BaseEvent
}

// CriminalActivity represents criminal activities.
type CriminalActivity struct {
	BaseEvent
}

// EducationalAchievement represents educational achievements.
type EducationalAchievement struct {
	BaseEvent
}

// HealthContribution represents health-related contributions.
type HealthContribution struct {
	BaseEvent
}

// EnvironmentalImpact represents environmental impact events.
type EnvironmentalImpact struct {
	BaseEvent
}
