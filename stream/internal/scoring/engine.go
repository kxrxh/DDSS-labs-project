package scoring

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/kxrxh/social-rating-system/stream/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// CitizenCache represents a cached citizen record
type CitizenCache struct {
	citizen     *models.Citizen
	lastUpdated time.Time
	dirty       bool
}

// Engine handles the scoring logic for events
type Engine struct {
	mongoClient *mongo.Client
	database    string
	config      *models.SystemConfiguration
	rules       []models.ScoringRule

	// Dgraph client for relationship queries
	httpClient *http.Client
	dgraphURL  string

	// Performance optimizations
	citizenCache  map[string]*CitizenCache
	cacheMutex    sync.RWMutex
	updateBatch   map[string]*models.Citizen
	batchMutex    sync.Mutex
	batchInterval time.Duration
}

// NewEngine creates a new scoring engine
func NewEngine(mongoClient *mongo.Client, database string, dgraphURL string) *Engine {
	engine := &Engine{
		mongoClient:   mongoClient,
		database:      database,
		dgraphURL:     dgraphURL,
		httpClient:    &http.Client{Timeout: 5 * time.Second},
		citizenCache:  make(map[string]*CitizenCache),
		updateBatch:   make(map[string]*models.Citizen),
		batchInterval: 1 * time.Second, // Batch updates every second
	}

	// Start batch update processor
	go engine.processBatchUpdates()

	return engine
}

// LoadConfiguration loads the system configuration and rules from MongoDB
func (e *Engine) LoadConfiguration(ctx context.Context) error {
	db := e.mongoClient.Database(e.database)

	// Load system configuration
	configCollection := db.Collection("system_configuration")
	var config models.SystemConfiguration
	err := configCollection.FindOne(ctx, bson.M{"_id": "default"}).Decode(&config)
	if err != nil {
		return fmt.Errorf("failed to load system configuration: %w", err)
	}
	e.config = &config

	// Load active scoring rules
	rulesCollection := db.Collection("scoring_rule_definitions")
	cursor, err := rulesCollection.Find(ctx, bson.M{
		"status":     "active",
		"valid_from": bson.M{"$lte": time.Now()},
		"$or": []bson.M{
			{"valid_to": bson.M{"$gte": time.Now()}},
			{"valid_to": nil},
		},
	})
	if err != nil {
		return fmt.Errorf("failed to load scoring rules: %w", err)
	}
	defer cursor.Close(ctx)

	var rules []models.ScoringRule
	if err := cursor.All(ctx, &rules); err != nil {
		return fmt.Errorf("failed to decode scoring rules: %w", err)
	}
	e.rules = rules

	log.Printf("Loaded %d active scoring rules and system configuration", len(e.rules))
	return nil
}

// ProcessEvent processes an event and calculates score changes
func (e *Engine) ProcessEvent(ctx context.Context, event models.Event) (*models.ProcessedEvent, error) {
	// Get current citizen data (now with caching)
	citizen, err := e.getCitizenCached(ctx, event.CitizenID)
	if err != nil {
		return nil, fmt.Errorf("failed to get citizen %s: %w", event.CitizenID, err)
	}

	// Find applicable rules for this event
	applicableRules := e.findApplicableRules(event)
	if len(applicableRules) == 0 {
		log.Printf("No applicable rules found for event type: %s", event.EventType)
		// Return processed event with no score change
		return &models.ProcessedEvent{
			Event:         event,
			ProcessedAt:   time.Now(),
			AppliedRules:  []string{},
			ScoreChange:   0,
			NewScore:      citizen.Score,
			PreviousScore: citizen.Score,
			ProcessingID:  fmt.Sprintf("proc_%d", time.Now().UnixNano()),
		}, nil
	}

	// Calculate total score change
	totalScoreChange := 0.0
	appliedRuleIDs := make([]string, 0, len(applicableRules))

	for _, rule := range applicableRules {
		// Check if rule conditions are met
		if !e.evaluateRuleConditions(rule, event) {
			continue
		}

		// TODO: Check cooldown state in ScyllaDB if needed
		// For now, we'll skip cooldown implementation

		// Calculate score change for this rule
		scoreChange := e.calculateRuleScore(rule, event, citizen)
		totalScoreChange += scoreChange
		appliedRuleIDs = append(appliedRuleIDs, rule.ID)

		log.Printf("Applied rule %s (%s) to event %s: %+.2f points",
			rule.ID, rule.Name, event.EventID, scoreChange)
	}

	// Calculate new score with bounds checking
	newScore := citizen.Score + totalScoreChange
	if newScore > e.config.MaxScore {
		newScore = e.config.MaxScore
	}
	if newScore < e.config.MinScore {
		newScore = e.config.MinScore
	}

	// Update citizen score using batch updates instead of immediate DB write
	updatedCitizen := *citizen
	updatedCitizen.Score = newScore
	now := time.Now()
	updatedCitizen.LastUpdated = now
	updatedCitizen.LastScoreAt = &now
	e.scheduleBatchUpdate(citizen.ID, &updatedCitizen)

	// Create processed event
	processedEvent := &models.ProcessedEvent{
		Event:         event,
		ProcessedAt:   time.Now(),
		AppliedRules:  appliedRuleIDs,
		ScoreChange:   totalScoreChange,
		NewScore:      newScore,
		PreviousScore: citizen.Score,
		ProcessingID:  fmt.Sprintf("proc_%d", time.Now().UnixNano()),
	}

	log.Printf("Processed event %s for citizen %s: score %+.2f → %.2f (change: %+.2f)",
		event.EventID, event.CitizenID, citizen.Score, newScore, totalScoreChange)

	// Process relationship events - affect related citizens
	if event.EventType == "relationship" {
		go e.processRelationshipImpact(ctx, event, totalScoreChange)
	}

	return processedEvent, nil
}

// getCitizenCached retrieves citizen data with caching
func (e *Engine) getCitizenCached(ctx context.Context, citizenID string) (*models.Citizen, error) {
	e.cacheMutex.RLock()
	if cached, exists := e.citizenCache[citizenID]; exists {
		// Check if cache is still valid (5 minutes)
		if time.Since(cached.lastUpdated) < 5*time.Minute {
			e.cacheMutex.RUnlock()
			return cached.citizen, nil
		}
	}
	e.cacheMutex.RUnlock()

	// Cache miss or expired, fetch from database
	citizen, err := e.getCitizen(ctx, citizenID)
	if err != nil {
		return nil, err
	}

	// Update cache
	e.cacheMutex.Lock()
	e.citizenCache[citizenID] = &CitizenCache{
		citizen:     citizen,
		lastUpdated: time.Now(),
		dirty:       false,
	}
	e.cacheMutex.Unlock()

	return citizen, nil
}

// scheduleBatchUpdate schedules a citizen update for batch processing
func (e *Engine) scheduleBatchUpdate(citizenID string, citizen *models.Citizen) {
	e.batchMutex.Lock()
	e.updateBatch[citizenID] = citizen
	e.batchMutex.Unlock()

	// Update cache immediately for consistency
	e.cacheMutex.Lock()
	if cached, exists := e.citizenCache[citizenID]; exists {
		cached.citizen = citizen
		cached.dirty = true
		cached.lastUpdated = time.Now()
	}
	e.cacheMutex.Unlock()
}

// processBatchUpdates processes citizen updates in batches
func (e *Engine) processBatchUpdates() {
	ticker := time.NewTicker(e.batchInterval)
	defer ticker.Stop()

	for range ticker.C {
		e.batchMutex.Lock()
		if len(e.updateBatch) == 0 {
			e.batchMutex.Unlock()
			continue
		}

		// Copy batch and clear for next iteration
		batch := make(map[string]*models.Citizen)
		for k, v := range e.updateBatch {
			batch[k] = v
		}
		e.updateBatch = make(map[string]*models.Citizen)
		e.batchMutex.Unlock()

		// Process batch updates
		if err := e.processBatchUpdatesDB(context.Background(), batch); err != nil {
			log.Printf("Error processing batch updates: %v", err)
		} else {
			log.Printf("Successfully processed batch of %d citizen updates", len(batch))
		}
	}
}

// processBatchUpdatesDB performs batch database updates
func (e *Engine) processBatchUpdatesDB(ctx context.Context, batch map[string]*models.Citizen) error {
	if len(batch) == 0 {
		return nil
	}

	collection := e.mongoClient.Database(e.database).Collection("citizens")

	// Prepare bulk write operations
	var operations []mongo.WriteModel
	for citizenID, citizen := range batch {
		// Calculate new tier
		newTier := citizen.CalculateTier(e.config)

		update := bson.M{
			"$set": bson.M{
				"score":         citizen.Score,
				"tier":          newTier,
				"last_updated":  citizen.LastUpdated,
				"last_score_at": citizen.LastScoreAt,
			},
		}

		operation := mongo.NewUpdateOneModel().
			SetFilter(bson.M{"_id": citizenID}).
			SetUpdate(update)
		operations = append(operations, operation)
	}

	// Execute bulk write
	opts := options.BulkWrite().SetOrdered(false) // Allow parallel execution
	result, err := collection.BulkWrite(ctx, operations, opts)
	if err != nil {
		return fmt.Errorf("bulk write failed: %w", err)
	}

	log.Printf("Bulk update completed: %d matched, %d modified",
		result.MatchedCount, result.ModifiedCount)

	return nil
}

// getCitizen retrieves citizen data from MongoDB
func (e *Engine) getCitizen(ctx context.Context, citizenID string) (*models.Citizen, error) {
	collection := e.mongoClient.Database(e.database).Collection("citizens")
	var citizen models.Citizen
	err := collection.FindOne(ctx, bson.M{"_id": citizenID}).Decode(&citizen)
	if err != nil {
		return nil, err
	}
	return &citizen, nil
}

// findApplicableRules finds rules that match the event type
func (e *Engine) findApplicableRules(event models.Event) []models.ScoringRule {
	var applicable []models.ScoringRule
	for _, rule := range e.rules {
		if rule.EventType == event.EventType || rule.EventType == "*" {
			applicable = append(applicable, rule)
		}
	}
	return applicable
}

// evaluateRuleConditions checks if the rule conditions are satisfied by the event
func (e *Engine) evaluateRuleConditions(rule models.ScoringRule, event models.Event) bool {
	// Simple condition evaluation - in a real system this would be more sophisticated
	for key, expectedValue := range rule.Conditions {
		if key == "event_subtype" {
			if event.EventSubtype != expectedValue {
				return false
			}
		}
		// Check payload conditions
		if payloadValue, exists := event.Payload[key]; exists {
			if payloadValue != expectedValue {
				// Try type-specific comparisons
				switch v := expectedValue.(type) {
				case string:
					if fmt.Sprintf("%v", payloadValue) != v {
						return false
					}
				case float64:
					if payloadFloat, ok := payloadValue.(float64); !ok || payloadFloat != v {
						return false
					}
				case bool:
					if payloadBool, ok := payloadValue.(bool); !ok || payloadBool != v {
						return false
					}
				default:
					return false
				}
			}
		} else if key != "event_subtype" {
			// Required condition not found in payload
			return false
		}
	}
	return true
}

// calculateRuleScore calculates the score change for a specific rule
func (e *Engine) calculateRuleScore(rule models.ScoringRule, event models.Event, citizen *models.Citizen) float64 {
	baseScore := rule.Points
	multiplier := rule.Multiplier

	// Apply multipliers based on citizen type or other factors
	if citizen.Type == "vip" {
		multiplier *= 1.1 // VIP citizens get 10% bonus
	}

	// Apply confidence factor
	multiplier *= event.Confidence

	return baseScore * multiplier
}

// GetConfiguration returns the current system configuration
func (e *Engine) GetConfiguration() *models.SystemConfiguration {
	return e.config
}

// GetRules returns the currently loaded rules
func (e *Engine) GetRules() []models.ScoringRule {
	return e.rules
}

// processRelationshipImpact processes relationship events to affect related citizens
func (e *Engine) processRelationshipImpact(ctx context.Context, event models.Event, scoreChange float64) {
	// Check if this event requires relationship processing
	requiresProcessing, ok := event.Payload["requires_relationship_processing"].(bool)
	if !ok || !requiresProcessing {
		return
	}

	log.Printf("Processing relationship impact for event %s (score change: %+.2f)", event.EventID, scoreChange)

	// Get relationship type and impact details
	relationshipType, _ := event.Payload["relationship_type"].(string)
	isPositive, _ := event.Payload["is_positive"].(bool)
	intensity, _ := event.Payload["intensity"].(float64)

	// Query related citizens from Dgraph based on relationship type
	relatedCitizens, err := e.queryRelatedCitizens(ctx, event.CitizenID, relationshipType)
	if err != nil {
		log.Printf("Failed to query related citizens for %s: %v", event.CitizenID, err)
		return
	}

	if len(relatedCitizens) == 0 {
		log.Printf("No related citizens found for %s", event.CitizenID)
		return
	}

	// Calculate impact factor based on event characteristics
	impactFactor := e.calculateRelationshipImpactFactor(relationshipType, isPositive, intensity)
	secondaryScoreChange := scoreChange * impactFactor

	log.Printf("Applying relationship impact to %d related citizens (factor: %.3f, secondary change: %+.2f)",
		len(relatedCitizens), impactFactor, secondaryScoreChange)

	// Apply secondary effects to related citizens
	for _, relatedCitizenID := range relatedCitizens {
		go e.applySecondaryScoreEffect(ctx, relatedCitizenID, secondaryScoreChange, event)
	}
}

// queryRelatedCitizens queries Dgraph for citizens related to the given citizen
func (e *Engine) queryRelatedCitizens(ctx context.Context, citizenID string, relationshipType string) ([]string, error) {
	// Build GraphQL++ query based on relationship type
	var query string
	switch relationshipType {
	case "family_event":
		query = fmt.Sprintf(`{
			citizen(func: eq(citizen_id, "%s")) {
				family_member @filter(not eq(citizen_id, "%s")) {
					citizen_id
				}
			}
		}`, citizenID, citizenID)
	case "workplace_event":
		query = fmt.Sprintf(`{
			citizen(func: eq(citizen_id, "%s")) {
				colleague @filter(not eq(citizen_id, "%s")) {
					citizen_id
				}
			}
		}`, citizenID, citizenID)
	case "community_event":
		// Query multiple relationship types for community events
		query = fmt.Sprintf(`{
			citizen(func: eq(citizen_id, "%s")) {
				neighbor @filter(not eq(citizen_id, "%s")) {
					citizen_id
				}
				friend @filter(not eq(citizen_id, "%s")) {
					citizen_id
				}
			}
		}`, citizenID, citizenID, citizenID)
	default:
		// Default to friends and general relationships
		query = fmt.Sprintf(`{
			citizen(func: eq(citizen_id, "%s")) {
				friend @filter(not eq(citizen_id, "%s")) {
					citizen_id
				}
			}
		}`, citizenID, citizenID)
	}

	// Execute query against Dgraph
	return e.executeDgraphQuery(ctx, query)
}

// executeDgraphQuery executes a query against Dgraph and returns citizen IDs
func (e *Engine) executeDgraphQuery(ctx context.Context, query string) ([]string, error) {
	url := fmt.Sprintf("http://%s/query", e.dgraphURL)

	queryData := map[string]interface{}{
		"query": query,
	}

	jsonData, err := json.Marshal(queryData)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("query failed with status %d", resp.StatusCode)
	}

	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Extract citizen IDs from the response
	return e.extractCitizenIDsFromResponse(result)
}

// extractCitizenIDsFromResponse extracts citizen IDs from Dgraph query response
func (e *Engine) extractCitizenIDsFromResponse(result map[string]interface{}) ([]string, error) {
	citizenIDs := make([]string, 0)

	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return citizenIDs, nil
	}

	citizens, ok := data["citizen"].([]interface{})
	if !ok || len(citizens) == 0 {
		return citizenIDs, nil
	}

	// Process first citizen (the one we queried for)
	if len(citizens) > 0 {
		citizen := citizens[0].(map[string]interface{})

		// Extract different relationship types
		relationshipTypes := []string{"family_member", "colleague", "neighbor", "friend"}
		for _, relType := range relationshipTypes {
			if relations, exists := citizen[relType].([]interface{}); exists {
				for _, relation := range relations {
					if relMap, ok := relation.(map[string]interface{}); ok {
						if citizenID, exists := relMap["citizen_id"].(string); exists {
							citizenIDs = append(citizenIDs, citizenID)
						}
					}
				}
			}
		}
	}

	return citizenIDs, nil
}

// calculateRelationshipImpactFactor calculates how much a relationship event affects related citizens
func (e *Engine) calculateRelationshipImpactFactor(relationshipType string, isPositive bool, intensity float64) float64 {
	baseFactor := 0.1 // Base 10% impact

	// Adjust factor based on relationship type
	switch relationshipType {
	case "family_event":
		baseFactor = 0.3 // Family events have higher impact (30%)
	case "workplace_event":
		baseFactor = 0.2 // Workplace events moderate impact (20%)
	case "community_event":
		baseFactor = 0.15 // Community events lower impact (15%)
	}

	// Adjust for intensity
	adjustedFactor := baseFactor * intensity

	// Reduce impact for negative events (they spread less effectively)
	if !isPositive {
		adjustedFactor *= 0.7
	}

	return adjustedFactor
}

// applySecondaryScoreEffect applies secondary score effects to a related citizen
func (e *Engine) applySecondaryScoreEffect(ctx context.Context, citizenID string, scoreChange float64, originalEvent models.Event) {
	// Get citizen data
	citizen, err := e.getCitizenCached(ctx, citizenID)
	if err != nil {
		log.Printf("Failed to get related citizen %s: %v", citizenID, err)
		return
	}

	// Calculate new score with bounds checking
	newScore := citizen.Score + scoreChange
	if newScore > e.config.MaxScore {
		newScore = e.config.MaxScore
	}
	if newScore < e.config.MinScore {
		newScore = e.config.MinScore
	}

	// Update citizen score using batch updates
	updatedCitizen := *citizen
	updatedCitizen.Score = newScore
	now := time.Now()
	updatedCitizen.LastUpdated = now
	updatedCitizen.LastScoreAt = &now
	e.scheduleBatchUpdate(citizen.ID, &updatedCitizen)

	log.Printf("Applied secondary effect to citizen %s from event %s: score %+.2f → %.2f (change: %+.2f)",
		citizenID, originalEvent.EventID, citizen.Score, newScore, scoreChange)
}
