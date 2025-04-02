// Package generator provides functionality for creating mock events.
package generator

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/google/uuid"
	// Use the correct path based on your go.mod file
	"github.com/kxrxh/social-rating-system/event-producer/pkg/events"
)

// Generator generates events.
type Generator struct {
	randSource *rand.Rand
}

// NewGenerator creates a new event generator.
func NewGenerator() *Generator {
	return &Generator{
		randSource: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// GenerateRandomEvent creates a random event for the given citizen ID.
func (g *Generator) GenerateRandomEvent(citizenID string) events.EventWithBase {
	eventTypeIndex := g.randSource.Intn(4) // 0 to 3

	switch eventTypeIndex {
	case 0:
		amount := (g.randSource.Float64() - 0.5) * 2000 // -1000 to +1000
		return events.FinancialEvent{
			BaseEvent: events.BaseEvent{
				CitizenID: citizenID,
				EventType: events.FinancialTransaction,
				Timestamp: time.Now().UTC(),
			},
			Amount:      amount,
			Description: fmt.Sprintf("Transaction %s", uuid.New().String()[:8]),
		}
	case 1:
		eventNames := []string{"City Cleanup", "Charity Run", "Local Festival", "Blood Donation"}
		roles := []string{"participant", "volunteer", "organizer"}
		return events.PublicEvent{
			BaseEvent: events.BaseEvent{
				CitizenID: citizenID,
				EventType: events.PublicEventParticipation,
				Timestamp: time.Now().UTC(),
			},
			EventName: eventNames[g.randSource.Intn(len(eventNames))],
			Role:      roles[g.randSource.Intn(len(roles))],
		}
	case 2:
		violations := []string{"Speeding", "Illegal Parking", "Public Disturbance"}
		details := []string{"Minor infraction", "Warning issued", "Fine applied"}
		return events.ViolationEvent{
			BaseEvent: events.BaseEvent{
				CitizenID: citizenID,
				EventType: events.TrafficViolation,
				Timestamp: time.Now().UTC(),
			},
			ViolationType: violations[g.randSource.Intn(len(violations))],
			Details:       details[g.randSource.Intn(len(details))],
			DemeritPoints: g.randSource.Intn(5), // 0 to 4 points
		}
	case 3:
		platforms := []string{"Twitter", "Facebook", "Local Forum"}
		actions := []string{"positive_post", "constructive_comment", "sharing_news", "negative_comment", "spreading_misinformation"}
		return events.SocialMediaEvent{
			BaseEvent: events.BaseEvent{
				CitizenID: citizenID,
				EventType: events.SocialMediaActivity,
				Timestamp: time.Now().UTC(),
			},
			Platform: platforms[g.randSource.Intn(len(platforms))],
			Action:   actions[g.randSource.Intn(len(actions))],
		}
	default:
		// Fallback, should not happen with Intn(4)
		log.Println("Warning: Reached default case in event generation")
		return events.FinancialEvent{ // Return a default valid event
			BaseEvent: events.BaseEvent{
				CitizenID: citizenID,
				EventType: events.FinancialTransaction,
				Timestamp: time.Now().UTC(),
			},
			Amount:      1.0,
			Description: "Default fallback event",
		}
	}
}
