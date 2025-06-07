// Package generator provides functionality for creating mock events.
package generator

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

// Event represents a social credit system event with proper JSON tags
type Event struct {
	EventID      string                 `json:"event_id"`
	CitizenID    string                 `json:"citizen_id"`
	EventType    string                 `json:"event_type"`
	EventSubtype string                 `json:"event_subtype"`
	SourceSystem string                 `json:"source_system"`
	Timestamp    time.Time              `json:"timestamp"`
	Location     EventLocation          `json:"location"`
	Confidence   float64                `json:"confidence"`
	Payload      map[string]interface{} `json:"payload"`
	Version      int                    `json:"version"`
	EventName    string                 `json:"event_name"`
}

// EventLocation represents the location where an event occurred
type EventLocation struct {
	City      string  `json:"city"`
	Region    string  `json:"region"`
	Country   string  `json:"country"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// Generator creates random events for the social rating system.
type Generator struct {
	rand *rand.Rand
}

// NewGenerator creates a new event generator.
func NewGenerator() *Generator {
	return &Generator{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GenerateRandomEvent creates a random event for the given citizen.
func (g *Generator) GenerateRandomEvent(citizenID string) Event {
	// 3% chance to generate a new citizen event
	if g.rand.Float32() < 0.03 {
		return g.generateNewCitizenEvent()
	}

	// Generate events that match our scoring rules
	// Financial: 30%, Social: 35%, Legal: 20%, Relationship: 15%
	randVal := g.rand.Float32()

	if randVal < 0.30 {
		return g.generateFinancialEvent(citizenID)
	} else if randVal < 0.65 {
		return g.generateSocialEvent(citizenID)
	} else if randVal < 0.85 {
		return g.generateLegalEvent(citizenID)
	} else {
		return g.generateRelationshipEvent(citizenID)
	}
}

// generateFinancialEvent creates a financial event (payment_completed or payment_failed)
func (g *Generator) generateFinancialEvent(citizenID string) Event {
	// 70% chance of successful payment, 30% chance of failed payment
	isSuccess := g.rand.Float32() < 0.70

	var eventSubtype string
	var eventName string
	if isSuccess {
		eventSubtype = "payment_completed"
		eventName = "Payment Completed"
	} else {
		eventSubtype = "payment_failed"
		eventName = "Payment Failed"
	}

	event := Event{
		EventID:      uuid.New().String(),
		CitizenID:    citizenID,
		EventType:    "financial",
		EventSubtype: eventSubtype,
		SourceSystem: "financial_system",
		Timestamp:    time.Now(),
		Location:     generateRandomLocation(),
		Confidence:   0.8 + g.rand.Float64()*0.2, // High confidence for financial events
		Payload:      make(map[string]interface{}),
		Version:      1,
		EventName:    eventName,
	}

	// Add financial-specific payload
	amount := float64(g.rand.Intn(5000) + 100) // $100 - $5100
	event.Payload["amount"] = amount
	event.Payload["payment_method"] = generateRandomPaymentMethod()
	event.Payload["transaction_id"] = uuid.New().String()

	if isSuccess {
		event.Payload["processing_time"] = g.rand.Intn(10) + 1 // 1-10 seconds
		event.Payload["bank_response"] = "approved"
	} else {
		event.Payload["failure_reason"] = generateRandomFailureReason()
		event.Payload["bank_response"] = "declined"
	}

	return event
}

// generateSocialEvent creates a social event (volunteering)
func (g *Generator) generateSocialEvent(citizenID string) Event {
	event := Event{
		EventID:      uuid.New().String(),
		CitizenID:    citizenID,
		EventType:    "social",
		EventSubtype: "volunteering",
		SourceSystem: "community_service",
		Timestamp:    time.Now(),
		Location:     generateRandomLocation(),
		Confidence:   0.9 + g.rand.Float64()*0.1, // Very high confidence for volunteering
		Payload:      make(map[string]interface{}),
		Version:      1,
		EventName:    "Volunteer Activity",
	}

	// Add social-specific payload
	event.Payload["activity_type"] = generateRandomVolunteerActivity()
	event.Payload["duration_hours"] = g.rand.Intn(8) + 1 // 1-8 hours
	event.Payload["organization"] = generateRandomOrganization()
	event.Payload["participants_count"] = g.rand.Intn(50) + 5 // 5-55 people
	event.Payload["verified_by"] = "community_coordinator"

	return event
}

// generateLegalEvent creates a legal event (violation)
func (g *Generator) generateLegalEvent(citizenID string) Event {
	event := Event{
		EventID:      uuid.New().String(),
		CitizenID:    citizenID,
		EventType:    "legal",
		EventSubtype: "violation",
		SourceSystem: "legal_system",
		Timestamp:    time.Now(),
		Location:     generateRandomLocation(),
		Confidence:   0.95 + g.rand.Float64()*0.05, // Very high confidence for legal events
		Payload:      make(map[string]interface{}),
		Version:      1,
		EventName:    "Legal Violation",
	}

	// Add legal-specific payload
	event.Payload["violation_type"] = generateRandomViolationType()
	event.Payload["severity"] = generateRandomSeverity()
	event.Payload["fine_amount"] = float64(g.rand.Intn(1000) + 50) // $50 - $1050
	event.Payload["case_number"] = generateRandomCaseNumber()
	event.Payload["court_jurisdiction"] = generateRandomJurisdiction()

	return event
}

// generateNewCitizenEvent creates an event for a new citizen joining the system.
func (g *Generator) generateNewCitizenEvent() Event {
	newCitizenID := uuid.New().String()
	event := Event{
		EventID:      uuid.New().String(),
		CitizenID:    newCitizenID,
		EventType:    "system",
		EventSubtype: "new_citizen",
		SourceSystem: "citizen_registry",
		Timestamp:    time.Now(),
		Location:     generateRandomLocation(),
		Confidence:   1.0, // Perfect confidence for system events
		Payload:      make(map[string]interface{}),
		Version:      1,
		EventName:    "New Citizen Registration",
	}

	// Populate the payload with new citizen data
	event.Payload["initial_age"] = g.rand.Intn(65) + 18 // 18-82 years old
	event.Payload["initial_score"] = 500.0              // Baseline score from system config
	event.Payload["citizen_type"] = generateRandomCitizenType()
	event.Payload["initial_status"] = "active"
	event.Payload["registration_source"] = generateRandomRegistrationSource()

	return event
}

// ToJSON converts the event to JSON bytes
func (e *Event) ToJSON() ([]byte, error) {
	return json.Marshal(e)
}

// Helper functions for generating random data
func generateRandomLocation() EventLocation {
	locations := []struct {
		city      string
		region    string
		country   string
		latitude  float64
		longitude float64
	}{
		{"New York", "New York", "USA", 40.7128, -74.0060},
		{"Los Angeles", "California", "USA", 34.0522, -118.2437},
		{"Chicago", "Illinois", "USA", 41.8781, -87.6298},
		{"Houston", "Texas", "USA", 29.7604, -95.3698},
		{"Phoenix", "Arizona", "USA", 33.4484, -112.0740},
		{"Philadelphia", "Pennsylvania", "USA", 39.9526, -75.1652},
		{"San Antonio", "Texas", "USA", 29.4241, -98.4936},
		{"San Diego", "California", "USA", 32.7157, -117.1611},
		{"Dallas", "Texas", "USA", 32.7767, -96.7970},
		{"San Jose", "California", "USA", 37.3382, -121.8863},
	}

	loc := locations[rand.Intn(len(locations))]
	return EventLocation{
		City:      loc.city,
		Region:    loc.region,
		Country:   loc.country,
		Latitude:  loc.latitude,
		Longitude: loc.longitude,
	}
}

func generateRandomPaymentMethod() string {
	methods := []string{"credit_card", "debit_card", "bank_transfer", "digital_wallet", "check"}
	return methods[rand.Intn(len(methods))]
}

func generateRandomFailureReason() string {
	reasons := []string{"insufficient_funds", "expired_card", "invalid_account", "network_error", "fraud_detected"}
	return reasons[rand.Intn(len(reasons))]
}

func generateRandomVolunteerActivity() string {
	activities := []string{"food_bank", "elderly_care", "environmental_cleanup", "education_support", "disaster_relief", "animal_shelter", "community_garden"}
	return activities[rand.Intn(len(activities))]
}

func generateRandomOrganization() string {
	orgs := []string{"Red Cross", "Habitat for Humanity", "Local Food Bank", "Community Center", "Environmental Alliance", "Youth Mentorship Program"}
	return orgs[rand.Intn(len(orgs))]
}

func generateRandomViolationType() string {
	violations := []string{"traffic_violation", "noise_complaint", "littering", "parking_violation", "public_disturbance", "minor_assault", "property_damage"}
	return violations[rand.Intn(len(violations))]
}

func generateRandomSeverity() string {
	severities := []string{"minor", "moderate", "serious"}
	return severities[rand.Intn(len(severities))]
}

func generateRandomCaseNumber() string {
	return fmt.Sprintf("CASE-%d-%04d", time.Now().Year(), rand.Intn(9999)+1)
}

func generateRandomJurisdiction() string {
	jurisdictions := []string{"Municipal Court", "County Court", "District Court", "Traffic Court"}
	return jurisdictions[rand.Intn(len(jurisdictions))]
}

func generateRandomCitizenType() string {
	types := []string{"regular", "student", "senior", "veteran", "public_servant"}
	return types[rand.Intn(len(types))]
}

func generateRandomRegistrationSource() string {
	sources := []string{"online_portal", "government_office", "immigration_office", "birth_certificate", "naturalization"}
	return sources[rand.Intn(len(sources))]
}

// generateRelationshipEvent creates relationship-based events that affect multiple citizens
func (g *Generator) generateRelationshipEvent(citizenID string) Event {
	relationshipTypes := []string{"family_event", "workplace_event", "community_event"}
	relationshipType := relationshipTypes[g.rand.Intn(len(relationshipTypes))]

	var eventSubtype, eventName string
	var isPositive bool

	switch relationshipType {
	case "family_event":
		// 60% positive family events, 40% negative
		isPositive = g.rand.Float32() < 0.60
		if isPositive {
			subtypes := []string{"family_gathering", "family_support", "family_celebration"}
			eventSubtype = subtypes[g.rand.Intn(len(subtypes))]
			eventName = "Family Positive Event"
		} else {
			subtypes := []string{"family_conflict", "family_neglect", "family_dispute"}
			eventSubtype = subtypes[g.rand.Intn(len(subtypes))]
			eventName = "Family Negative Event"
		}

	case "workplace_event":
		// 70% positive workplace events, 30% negative
		isPositive = g.rand.Float32() < 0.70
		if isPositive {
			subtypes := []string{"team_collaboration", "workplace_achievement", "mentoring"}
			eventSubtype = subtypes[g.rand.Intn(len(subtypes))]
			eventName = "Workplace Positive Event"
		} else {
			subtypes := []string{"workplace_conflict", "poor_teamwork", "misconduct"}
			eventSubtype = subtypes[g.rand.Intn(len(subtypes))]
			eventName = "Workplace Negative Event"
		}

	case "community_event":
		// 80% positive community events, 20% negative
		isPositive = g.rand.Float32() < 0.80
		if isPositive {
			subtypes := []string{"community_organizing", "neighborhood_help", "local_leadership"}
			eventSubtype = subtypes[g.rand.Intn(len(subtypes))]
			eventName = "Community Positive Event"
		} else {
			subtypes := []string{"community_disruption", "neighbor_complaint", "public_disturbance"}
			eventSubtype = subtypes[g.rand.Intn(len(subtypes))]
			eventName = "Community Negative Event"
		}
	}

	event := Event{
		EventID:      uuid.New().String(),
		CitizenID:    citizenID,
		EventType:    "relationship",
		EventSubtype: eventSubtype,
		SourceSystem: "social_network_system",
		Timestamp:    time.Now(),
		Location:     generateRandomLocation(),
		Confidence:   0.75 + g.rand.Float64()*0.20, // Medium to high confidence
		Payload:      make(map[string]interface{}),
		Version:      1,
		EventName:    eventName,
	}

	// Add relationship-specific payload
	event.Payload["relationship_type"] = relationshipType
	event.Payload["is_positive"] = isPositive
	event.Payload["impact_radius"] = g.rand.Intn(3) + 1   // 1-3 degrees of separation
	event.Payload["intensity"] = g.rand.Float64()         // 0.0-1.0 intensity
	event.Payload["duration_hours"] = g.rand.Intn(24) + 1 // 1-24 hours

	// Add specific context based on event type
	switch relationshipType {
	case "family_event":
		event.Payload["family_size_affected"] = g.rand.Intn(5) + 2 // 2-6 family members
		event.Payload["event_location"] = generateRandomFamilyLocation()
	case "workplace_event":
		event.Payload["team_size_affected"] = g.rand.Intn(10) + 3 // 3-12 colleagues
		event.Payload["department"] = generateRandomDepartment()
	case "community_event":
		event.Payload["community_size_affected"] = g.rand.Intn(20) + 5 // 5-24 community members
		event.Payload["event_scope"] = generateRandomCommunityScope()
	}

	// Mark event as requiring relationship processing
	event.Payload["requires_relationship_processing"] = true

	return event
}

func generateRandomFamilyLocation() string {
	locations := []string{"home", "family_restaurant", "park", "family_event_venue", "relative_house"}
	return locations[rand.Intn(len(locations))]
}

func generateRandomDepartment() string {
	departments := []string{"engineering", "marketing", "sales", "hr", "finance", "operations", "customer_service"}
	return departments[rand.Intn(len(departments))]
}

func generateRandomCommunityScope() string {
	scopes := []string{"neighborhood", "local_district", "community_center", "local_business", "public_space"}
	return scopes[rand.Intn(len(scopes))]
}
