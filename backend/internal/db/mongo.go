package db

import (
	"go.mongodb.org/mongo-driver/v2/mongo"
)

// MongoCollections holds the MongoDB collection names
const (
	CitizenProfilesColl = "citizen_profiles"
	ScoringRulesColl    = "scoring_rules"
)

// GetCitizenProfiles returns the citizen profiles collection
func (c *Client) GetCitizenProfiles() *mongo.Collection {
	return c.doc.Database("social_credit").Collection(CitizenProfilesColl)
}

// GetScoringRules returns the scoring rules collection
func (c *Client) GetScoringRules() *mongo.Collection {
	return c.doc.Database("social_credit").Collection(ScoringRulesColl)
}
