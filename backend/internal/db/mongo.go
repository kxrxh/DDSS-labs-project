package db

import (
	"context"
	"fmt"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

// SetupMongoCollections initializes MongoDB collections and indexes
func (c *Client) SetupMongoCollections() error {
	ctx := context.Background()
	db := c.doc.Database("social_credit_db")

	// Citizen collection
	citizenCollection := db.Collection("citizens")
	citizenIndex := mongo.IndexModel{
		Keys:    bson.M{"simulated_id": 1},
		Options: options.Index().SetUnique(true),
	}
	_, err := citizenCollection.Indexes().CreateOne(ctx, citizenIndex)
	if err != nil {
		return fmt.Errorf("failed to create citizen index: %w", err)
	}

	// Config collection
	configCollection := db.Collection("configs")
	configIndex := mongo.IndexModel{
		Keys: bson.M{"rule_id": 1},
	}
	_, err = configCollection.Indexes().CreateOne(ctx, configIndex)
	if err != nil {
		return fmt.Errorf("failed to create config index: %w", err)
	}

	// Insert default scoring rule
	defaultRule := bson.M{
		"rule_id": "late_payment",
		"points":  -10,
		"type":    "deduction",
	}
	_, err = configCollection.InsertOne(ctx, defaultRule)
	return err
}

// AddCitizenProfile adds a citizen profile to MongoDB
func (c *Client) AddCitizenProfile(simulatedID, name string, age int, score int) error {
	ctx := context.Background()
	collection := c.doc.Database("social_credit_db").Collection("citizens")

	doc := bson.M{
		"simulated_id": simulatedID,
		"name":         name,
		"age":          age,
		"score":        score,
	}
	_, err := collection.InsertOne(ctx, doc)
	if err != nil {
		return fmt.Errorf("failed to insert citizen profile: %w", err)
	}
	return nil
}

// GetCitizenScore retrieves a citizen's score
func (c *Client) GetCitizenScore(simulatedID string) (int, error) {
	ctx := context.Background()
	collection := c.doc.Database("social_credit_db").Collection("citizens")

	var result struct {
		Score int `bson:"score"`
	}
	err := collection.FindOne(ctx, bson.M{"simulated_id": simulatedID}).Decode(&result)
	if err != nil {
		return 0, fmt.Errorf("failed to find citizen: %w", err)
	}
	return result.Score, nil
}