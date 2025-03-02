package db

import (
	"context"
	"fmt"
	"time"

	"github.com/kxrxh/ddss/pkg/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	dbName          = "social_credit_db"
	citizenCollName = "citizens"
	rulesCollName   = "rules"
	transCollName   = "transactions"
)

// SetupMongoCollections initializes MongoDB collections and indexes
func (c *Client) SetupMongoCollections() error {
	ctx := context.Background()
	db := c.doc.Database(dbName)

	// Citizen collection
	citizenCollection := db.Collection(citizenCollName)
	citizenIndex := mongo.IndexModel{
		Keys:    bson.M{"simulated_id": 1},
		Options: options.Index().SetUnique(true),
	}
	_, err := citizenCollection.Indexes().CreateOne(ctx, citizenIndex)
	if err != nil {
		return fmt.Errorf("failed to create citizen index: %w", err)
	}

	// Rules collection
	rulesCollection := db.Collection(rulesCollName)
	ruleIndex := mongo.IndexModel{
		Keys:    bson.M{"rule_id": 1},
		Options: options.Index().SetUnique(true),
	}
	_, err = rulesCollection.Indexes().CreateOne(ctx, ruleIndex)
	if err != nil {
		return fmt.Errorf("failed to create rule index: %w", err)
	}

	// Transactions collection
	txCollection := db.Collection(transCollName)
	txIndex := mongo.IndexModel{
		Keys:    bson.M{"transaction_id": 1},
		Options: options.Index().SetUnique(true),
	}
	_, err = txCollection.Indexes().CreateOne(ctx, txIndex)
	if err != nil {
		return fmt.Errorf("failed to create transaction index: %w", err)
	}

	// Add citizen_id index to transactions for quick lookups
	citizenIdIndex := mongo.IndexModel{
		Keys: bson.M{"citizen_id": 1},
	}
	_, err = txCollection.Indexes().CreateOne(ctx, citizenIdIndex)
	if err != nil {
		return fmt.Errorf("failed to create citizen_id index: %w", err)
	}

	// Insert default scoring rules if they don't exist
	filter := bson.M{"rule_id": "late_payment"}
	count, err := rulesCollection.CountDocuments(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to check for default rule: %w", err)
	}

	if count == 0 {
		defaultRule := models.Rule{
			RuleID:      "late_payment",
			Name:        "Late Payment Penalty",
			Description: "Penalty for late payment of bills or taxes",
			Category:    "financial",
			Subcategory: "payment",
			PointImpact: models.PointImpact{
				BaseValue: -10,
				MultiplierFactors: []models.MultiplierFactor{
					{
						FactorName: "repeat_offense",
						Threshold:  2.0,
						Multiplier: 1.5,
					},
				},
				MaxImpact: -30,
			},
			Conditions: models.RuleConditions{
				Prerequisites: []string{"has_active_account"},
				Exclusions:    []string{"hardship_exemption"},
			},
			Enforcement: models.RuleEnforcement{
				Automatic:      true,
				RequiresReview: false,
				ReviewLevel:    0,
			},
			Metadata: models.RuleMetadata{
				CreatedAt:  time.Now(),
				ModifiedAt: time.Now(),
				Version:    "1.0",
				Active:     true,
			},
		}
		_, err = rulesCollection.InsertOne(ctx, defaultRule)
		if err != nil {
			return fmt.Errorf("failed to insert default rule: %w", err)
		}
	}

	return nil
}

// AddCitizenProfile adds a comprehensive citizen profile to MongoDB
func (c *Client) AddCitizenProfile(citizen models.Citizen) error {
	ctx := context.Background()
	collection := c.doc.Database(dbName).Collection(citizenCollName)

	_, err := collection.InsertOne(ctx, citizen)
	if err != nil {
		return fmt.Errorf("failed to insert citizen profile: %w", err)
	}
	return nil
}

// UpdateCitizenProfile updates an existing citizen profile
func (c *Client) UpdateCitizenProfile(citizen models.Citizen) error {
	ctx := context.Background()
	collection := c.doc.Database(dbName).Collection(citizenCollName)

	filter := bson.M{"simulated_id": citizen.SimulatedID}
	update := bson.M{"$set": citizen}

	result, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update citizen profile: %w", err)
	}

	if result.MatchedCount == 0 {
		return fmt.Errorf("no citizen found with ID %s", citizen.SimulatedID)
	}

	return nil
}

// GetCitizenByID retrieves a citizen by their ID
func (c *Client) GetCitizenByID(simulatedID string) (*models.Citizen, error) {
	ctx := context.Background()
	collection := c.doc.Database(dbName).Collection(citizenCollName)

	var citizen models.Citizen
	err := collection.FindOne(ctx, bson.M{"simulated_id": simulatedID}).Decode(&citizen)
	if err != nil {
		return nil, fmt.Errorf("failed to find citizen: %w", err)
	}

	return &citizen, nil
}

// GetCitizenScore retrieves a citizen's score
func (c *Client) GetCitizenScore(simulatedID string) (int, error) {
	ctx := context.Background()
	collection := c.doc.Database(dbName).Collection(citizenCollName)

	var result struct {
		CreditProfile struct {
			CurrentScore int `bson:"current_score"`
		} `bson:"credit_profile"`
	}

	err := collection.FindOne(ctx, bson.M{"simulated_id": simulatedID}).Decode(&result)
	if err != nil {
		return 0, fmt.Errorf("failed to find citizen: %w", err)
	}

	return result.CreditProfile.CurrentScore, nil
}

// RecordTransaction records a credit score transaction
func (c *Client) RecordTransaction(transaction models.Transaction) error {
	ctx := context.Background()
	collection := c.doc.Database(dbName).Collection(transCollName)

	_, err := collection.InsertOne(ctx, transaction)
	if err != nil {
		return fmt.Errorf("failed to insert transaction: %w", err)
	}

	// Update citizen's score
	citizenCollection := c.doc.Database(dbName).Collection(citizenCollName)
	filter := bson.M{"simulated_id": transaction.CitizenID}

	// Get current timestamp
	now := time.Now()

	// Add score history entry and update current score
	update := bson.M{
		"$set": bson.M{
			"credit_profile.current_score": transaction.NewScore,
			"credit_profile.last_updated":  now,
		},
		"$push": bson.M{
			"credit_profile.score_history": bson.M{
				"date":   now,
				"score":  transaction.NewScore,
				"change": transaction.PointsChange,
			},
		},
	}

	_, err = citizenCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update citizen score: %w", err)
	}

	return nil
}

// AddRule adds a new rule to the system
func (c *Client) AddRule(rule models.Rule) error {
	ctx := context.Background()
	collection := c.doc.Database(dbName).Collection(rulesCollName)

	_, err := collection.InsertOne(ctx, rule)
	if err != nil {
		return fmt.Errorf("failed to insert rule: %w", err)
	}

	return nil
}

// GetRuleByID retrieves a rule by its ID
func (c *Client) GetRuleByID(ruleID string) (*models.Rule, error) {
	ctx := context.Background()
	collection := c.doc.Database(dbName).Collection(rulesCollName)

	var rule models.Rule
	err := collection.FindOne(ctx, bson.M{"rule_id": ruleID}).Decode(&rule)
	if err != nil {
		return nil, fmt.Errorf("failed to find rule: %w", err)
	}

	return &rule, nil
}

// GetAllRules retrieves all rules with optional filtering
func (c *Client) GetAllRules(query bson.M, limit int) ([]models.Rule, error) {
	ctx := context.Background()
	collection := c.doc.Database(dbName).Collection(rulesCollName)

	findOptions := options.Find()
	findOptions.SetLimit(int64(limit))

	cursor, err := collection.Find(ctx, query, findOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to find rules: %w", err)
	}
	defer cursor.Close(ctx)

	var rules []models.Rule
	if err = cursor.All(ctx, &rules); err != nil {
		return nil, fmt.Errorf("failed to decode rules: %w", err)
	}

	return rules, nil
}

// SearchCitizens searches for citizens based on various criteria
func (c *Client) SearchCitizens(query bson.M, limit int) ([]models.Citizen, error) {
	ctx := context.Background()
	collection := c.doc.Database(dbName).Collection(citizenCollName)

	opts := options.Find()
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := collection.Find(ctx, query, opts)
	if err != nil {
		return nil, fmt.Errorf("search failed: %w", err)
	}
	defer cursor.Close(ctx)

	var citizens []models.Citizen
	if err := cursor.All(ctx, &citizens); err != nil {
		return nil, fmt.Errorf("failed to decode search results: %w", err)
	}

	return citizens, nil
}

// GetCitizenTransactions retrieves a citizen's transaction history
func (c *Client) GetCitizenTransactions(simulatedID string, limit int) ([]models.Transaction, error) {
	ctx := context.Background()
	collection := c.doc.Database(dbName).Collection(transCollName)

	opts := options.Find().SetSort(bson.M{"timestamp": -1}) // Sort by most recent
	if limit > 0 {
		opts.SetLimit(int64(limit))
	}

	cursor, err := collection.Find(ctx, bson.M{"citizen_id": simulatedID}, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to find transactions: %w", err)
	}
	defer cursor.Close(ctx)

	var transactions []models.Transaction
	if err := cursor.All(ctx, &transactions); err != nil {
		return nil, fmt.Errorf("failed to decode transactions: %w", err)
	}

	return transactions, nil
}

// DeleteCitizen removes a citizen from the database by ID
func (c *Client) DeleteCitizen(citizenID string) error {
	ctx := context.Background()
	collection := c.doc.Database(dbName).Collection(citizenCollName)

	filter := bson.M{"simulated_id": citizenID}
	result, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return fmt.Errorf("failed to delete citizen: %w", err)
	}

	if result.DeletedCount == 0 {
		return fmt.Errorf("no citizen found with ID %s", citizenID)
	}

	return nil
}
