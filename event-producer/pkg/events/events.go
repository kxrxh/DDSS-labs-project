// Package events defines the shared event structures and types.
package events

import (
	"time"
)

// EventType defines the type of event.
type EventType string

// Constants for different event types.
const (
	FinancialTransaction     EventType = "financial_transaction"
	PublicEventParticipation EventType = "public_event_participation"
	TrafficViolation         EventType = "traffic_violation"
	SocialMediaActivity      EventType = "social_media_activity"
)

// BaseEvent contains common fields for all events.
type BaseEvent struct {
	CitizenID string    `json:"citizen_id"`
	EventType EventType `json:"event_type"`
	Timestamp time.Time `json:"timestamp"`
}

// FinancialEvent represents a financial transaction.
type FinancialEvent struct {
	BaseEvent
	Amount      float64 `json:"amount"` // Positive for income, negative for expense
	Description string  `json:"description"`
}

// PublicEvent represents participation in a public event.
type PublicEvent struct {
	BaseEvent
	EventName string `json:"event_name"`
	Role      string `json:"role"` // e.g., participant, organizer
}

// ViolationEvent represents some kind of violation.
type ViolationEvent struct {
	BaseEvent
	ViolationType string `json:"violation_type"`
	Details       string `json:"details"`
	DemeritPoints int    `json:"demerit_points,omitempty"` // Optional points deduction
}

// SocialMediaEvent represents activity on social media.
type SocialMediaEvent struct {
	BaseEvent
	Platform string `json:"platform"`
	Action   string `json:"action"` // e.g., positive_post, negative_comment, fake_news_share
	Content  string `json:"content_summary,omitempty"`
}

// EventWithBase is an interface to ensure an event can provide its BaseEvent.
// This is useful for generic handling where you need common fields.
type EventWithBase interface {
	GetBaseEvent() BaseEvent
}

// Implement the interface for each event type.
func (e FinancialEvent) GetBaseEvent() BaseEvent   { return e.BaseEvent }
func (e PublicEvent) GetBaseEvent() BaseEvent      { return e.BaseEvent }
func (e ViolationEvent) GetBaseEvent() BaseEvent   { return e.BaseEvent }
func (e SocialMediaEvent) GetBaseEvent() BaseEvent { return e.BaseEvent }
