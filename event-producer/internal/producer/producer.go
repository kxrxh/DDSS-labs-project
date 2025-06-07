package producer

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/dgraph-io/dgo/v240"
	"github.com/dgraph-io/dgo/v240/protos/api"
	"github.com/google/uuid"
	"github.com/kxrxh/social-rating-system/event-producer/internal/config"
	"github.com/kxrxh/social-rating-system/event-producer/internal/generator"
	"github.com/twmb/franz-go/pkg/kgo"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Producer manages the Kafka client and event generation/sending loop.
type Producer struct {
	client       *kgo.Client
	cfg          *config.Config
	generator    *generator.Generator
	citizenIDs   []string
	mongoClient  *mongo.Client
	dgraphClient *dgo.Dgraph
	grpcConn     *grpc.ClientConn
}

// NewProducer creates and initializes a new Producer.
func NewProducer(cfg *config.Config, gen *generator.Generator) (*Producer, error) {
	log.Printf("Initializing Kafka client with brokers: %v", cfg.KafkaBrokers)
	opts := []kgo.Opt{
		kgo.SeedBrokers(cfg.KafkaBrokers...),
		kgo.DefaultProduceTopic(cfg.KafkaTopic),
		kgo.RequiredAcks(kgo.AllISRAcks()),
		kgo.ProducerBatchCompression(kgo.SnappyCompression()),
		kgo.AllowAutoTopicCreation(),
		kgo.WithLogger(kgo.BasicLogger(log.Writer(), kgo.LogLevelInfo, func() string { return "KAFKA: " })),
	}

	client, err := kgo.NewClient(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kafka client: %w", err)
	}

	// Ping to check connectivity early
	ctxPing, cancelPing := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelPing()
	if err = client.Ping(ctxPing); err != nil {
		log.Printf("Warning: Failed to ping Kafka broker(s): %v. Continuing...", err)
	} else {
		log.Println("Kafka client initialized and connection ping successful.")
	}

	// Initialize MongoDB client
	mongoClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI(cfg.MongoURI))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	// Ping MongoDB to verify connection
	if err = mongoClient.Ping(context.Background(), nil); err != nil {
		return nil, fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	// Initialize Dgraph gRPC client
	conn, err := grpc.Dial(cfg.DgraphURL, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Printf("Warning: Failed to connect to Dgraph: %v", err)
	}

	var dgraphClient *dgo.Dgraph
	if conn != nil {
		dgraphClient = dgo.NewDgraphClient(api.NewDgraphClient(conn))
	}

	// Initialize base citizens if none exist
	collection := mongoClient.Database("social_rating").Collection("citizens")
	count, err := collection.CountDocuments(context.Background(), bson.M{})
	if err != nil {
		return nil, fmt.Errorf("failed to count citizens in MongoDB: %w", err)
	}

	if count < int64(cfg.NumCitizens) {
		log.Println("No citizens found in MongoDB. Creating base citizens...")
		baseCitizens := createBaseCitizens(cfg.NumCitizens)
		_, err = collection.InsertMany(context.Background(), baseCitizens)
		if err != nil {
			return nil, fmt.Errorf("failed to create base citizens: %w", err)
		}
		log.Printf("Created %d base citizens", len(baseCitizens))

		// Extract citizen IDs for relationship creation
		newCitizenIDs := make([]string, len(baseCitizens))
		for i, citizen := range baseCitizens {
			if bsonData, ok := citizen.(bson.M); ok {
				if id, exists := bsonData["_id"]; exists {
					newCitizenIDs[i] = id.(string)
				}
			}
		}

		// Create relationships in Dgraph
		if dgraphClient != nil {
			producer := &Producer{dgraphClient: dgraphClient, cfg: cfg}
			if err := producer.createCitizenRelationships(newCitizenIDs); err != nil {
				log.Printf("Warning: Failed to create citizen relationships in Dgraph: %v", err)
			}
		}
	}

	// Fetch all citizen IDs from MongoDB
	cursor, err := collection.Find(context.Background(), bson.M{}, options.Find().SetProjection(bson.M{"_id": 1}))
	if err != nil {
		return nil, fmt.Errorf("failed to fetch citizens from MongoDB: %w", err)
	}
	defer cursor.Close(context.Background())

	var citizens []struct {
		ID string `bson:"_id"`
	}
	if err = cursor.All(context.Background(), &citizens); err != nil {
		return nil, fmt.Errorf("failed to decode citizens from MongoDB: %w", err)
	}

	// Extract citizen IDs
	citizenIDs := make([]string, len(citizens))
	for i, citizen := range citizens {
		citizenIDs[i] = citizen.ID
	}

	log.Printf("Found %d citizens in MongoDB. Will generate events for these citizens.", len(citizenIDs))

	producer := &Producer{
		client:       client,
		cfg:          cfg,
		generator:    gen,
		citizenIDs:   citizenIDs,
		mongoClient:  mongoClient,
		dgraphClient: dgraphClient,
		grpcConn:     conn,
	}

	// Initialize Dgraph schema if needed
	if dgraphClient != nil {
		if err := producer.initializeDgraphSchema(); err != nil {
			log.Printf("Warning: Failed to initialize Dgraph schema: %v", err)
		}
	}

	return producer, nil
}

// createBaseCitizens creates a set of initial citizens for the system.
func createBaseCitizens(count int) []interface{} {
	citizens := make([]interface{}, count)
	now := time.Now()
	baselineScore := 500.0 // From system configuration

	for i := 0; i < count; i++ {
		citizenID := uuid.New().String()

		// Calculate initial tier based on baseline score
		tier := calculateTier(baselineScore)

		citizens[i] = bson.M{
			"_id":           citizenID,
			"age":           rand.Intn(65) + 18, // Random age between 18 and 82
			"score":         baselineScore,      // Start with baseline score from system config
			"type":          generateRandomCitizenType(),
			"status":        "active",
			"tier":          tier,
			"created_at":    now,
			"last_updated":  now,
			"last_score_at": now,
		}
	}

	return citizens
}

// calculateTier determines the tier based on score (matches system configuration)
func calculateTier(score float64) string {
	if score >= 800 {
		return "excellent"
	} else if score >= 600 {
		return "good"
	} else if score >= 400 {
		return "average"
	} else if score >= 200 {
		return "poor"
	} else {
		return "very_poor"
	}
}

// Run starts the event production loop.
// It blocks until the context is cancelled.
func (p *Producer) Run(ctx context.Context) {
	if len(p.citizenIDs) == 0 {
		log.Println("No citizens available. Event production loop will not start.")
		return
	}

	ticker := time.NewTicker(p.cfg.SendInterval)
	defer ticker.Stop()

	log.Printf("Starting event production loop (interval: %s)...", p.cfg.SendInterval)

	for {
		select {
		case <-ticker.C:
			// Generate a random event
			event := p.generator.GenerateRandomEvent(p.citizenIDs[rand.Intn(len(p.citizenIDs))])

			// Handle new citizen events (now with event type "system" and subtype "new_citizen")
			if event.EventType == "system" && event.EventSubtype == "new_citizen" {
				// Add new citizen to MongoDB
				collection := p.mongoClient.Database("social_rating").Collection("citizens")

				// Extract data from the payload
				payload := event.Payload

				age, ok := payload["initial_age"].(int)
				if !ok {
					log.Printf("Error asserting initial_age to int for new citizen %s", event.CitizenID)
					continue
				}

				score, ok := payload["initial_score"].(float64)
				if !ok {
					log.Printf("Error asserting initial_score to float64 for new citizen %s", event.CitizenID)
					continue
				}

				cType, ok := payload["citizen_type"].(string)
				if !ok {
					log.Printf("Error asserting citizen_type to string for new citizen %s", event.CitizenID)
					continue
				}

				status, ok := payload["initial_status"].(string)
				if !ok {
					log.Printf("Error asserting initial_status to string for new citizen %s", event.CitizenID)
					continue
				}

				// Calculate tier for the new citizen
				tier := calculateTier(score)

				_, err := collection.InsertOne(context.Background(), bson.M{
					"_id":           event.CitizenID,
					"age":           age,
					"score":         score,
					"type":          cType,
					"status":        status,
					"tier":          tier,
					"created_at":    time.Now(),
					"last_updated":  time.Now(),
					"last_score_at": time.Now(),
				})
				if err != nil {
					log.Printf("Error adding new citizen to MongoDB: %v", err)
					continue
				}

				// Add new citizen ID to the list
				p.citizenIDs = append(p.citizenIDs, event.CitizenID)
				log.Printf("Added new citizen %s to the system (age: %d, score: %.1f, tier: %s)",
					event.CitizenID, age, score, tier)
			}

			// Send event to Kafka
			eventJSON, err := event.ToJSON()
			if err != nil {
				log.Printf("Error serializing event to JSON: %v", err)
				continue
			}

			record := &kgo.Record{
				Topic: p.cfg.KafkaTopic,
				Key:   []byte(event.CitizenID), // Use CitizenID as key for partitioning
				Value: eventJSON,
				Headers: []kgo.RecordHeader{
					{Key: "event_type", Value: []byte(event.EventType)},
					{Key: "event_subtype", Value: []byte(event.EventSubtype)},
					{Key: "source_system", Value: []byte(event.SourceSystem)},
				},
			}

			// Produce the record
			p.client.Produce(ctx, record, func(r *kgo.Record, err error) {
				if err != nil {
					log.Printf("Error producing event %s: %v", event.EventID, err)
				} else {
					log.Printf("Successfully produced event %s (type: %s, subtype: %s, citizen: %s)",
						event.EventID, event.EventType, event.EventSubtype, event.CitizenID)
				}
			})

		case <-ctx.Done():
			log.Println("Context cancelled, stopping event production loop.")
			return
		}
	}
}

// Close gracefully shuts down the producer.
func (p *Producer) Close() {
	log.Println("Closing Kafka producer...")
	if p.client != nil {
		p.client.Close()
	}

	if p.mongoClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		p.mongoClient.Disconnect(ctx)
	}

	if p.grpcConn != nil {
		p.grpcConn.Close()
	}

	log.Println("Producer closed.")
}

func generateRandomCitizenType() string {
	types := []string{"regular", "student", "senior", "veteran", "public_servant"}
	return types[rand.Intn(len(types))]
}

// initializeDgraphSchema sets up the schema for citizen relationships in Dgraph
func (p *Producer) initializeDgraphSchema() error {
	if p.dgraphClient == nil {
		return fmt.Errorf("dgraph client not available")
	}

	ctx := context.Background()

	// Set up the schema
	schema := `
		citizen_id: string @index(exact) .
		citizen_name: string .
		age: int .
		type: [uid] @reverse .
		knows: [uid] @reverse .
		family_member: [uid] @reverse .
		colleague: [uid] @reverse .
		neighbor: [uid] @reverse .
		friend: [uid] @reverse .
		
		type Citizen {
			citizen_id
			citizen_name
			age
			knows
			family_member
			colleague
			neighbor
			friend
		}
	`

	op := &api.Operation{Schema: schema}
	return p.dgraphClient.Alter(ctx, op)
}

// createCitizenRelationships creates relationships between citizens in Dgraph
func (p *Producer) createCitizenRelationships(citizenIDs []string) error {
	if p.dgraphClient == nil {
		return fmt.Errorf("dgraph client not available")
	}

	log.Println("Creating citizen relationships in Dgraph...")

	ctx := context.Background()
	txn := p.dgraphClient.NewTxn()
	defer txn.Discard(ctx)

	// Create citizen nodes first
	mu := &api.Mutation{}
	nquads := ""

	for _, citizenID := range citizenIDs {
		nquads += fmt.Sprintf(`_:%s <citizen_id> "%s" .`+"\n", citizenID, citizenID)
		nquads += fmt.Sprintf(`_:%s <dgraph.type> "Citizen" .`+"\n", citizenID)
	}

	mu.SetNquads = []byte(nquads)
	if _, err := txn.Mutate(ctx, mu); err != nil {
		return fmt.Errorf("failed to create citizen nodes: %w", err)
	}

	if err := txn.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit citizen nodes: %w", err)
	}

	// Create relationships in separate transaction
	txn = p.dgraphClient.NewTxn()
	defer txn.Discard(ctx)

	relationshipNquads := ""
	for i, citizenID := range citizenIDs {
		// Create family relationships (2-4 family members per citizen)
		familySize := rand.Intn(3) + 2
		for j := 0; j < familySize && j < len(citizenIDs)-1; j++ {
			relatedCitizenIdx := (i + j + 1) % len(citizenIDs)
			if relatedCitizenIdx != i {
				relationshipNquads += fmt.Sprintf(`_:%s <family_member> _:%s .`+"\n", citizenID, citizenIDs[relatedCitizenIdx])
			}
		}

		// Create friend relationships (3-8 friends per citizen)
		friendCount := rand.Intn(6) + 3
		for j := 0; j < friendCount && j < len(citizenIDs)-1; j++ {
			friendIdx := rand.Intn(len(citizenIDs))
			if friendIdx != i {
				relationshipNquads += fmt.Sprintf(`_:%s <friend> _:%s .`+"\n", citizenID, citizenIDs[friendIdx])
			}
		}

		// Create colleague relationships (5-15 colleagues per citizen)
		colleagueCount := rand.Intn(11) + 5
		for j := 0; j < colleagueCount && j < len(citizenIDs)-1; j++ {
			colleagueIdx := rand.Intn(len(citizenIDs))
			if colleagueIdx != i {
				relationshipNquads += fmt.Sprintf(`_:%s <colleague> _:%s .`+"\n", citizenID, citizenIDs[colleagueIdx])
			}
		}
	}

	mu = &api.Mutation{SetNquads: []byte(relationshipNquads)}
	if _, err := txn.Mutate(ctx, mu); err != nil {
		log.Printf("Warning: Failed to create relationship batch: %v", err)
		return err
	}

	if err := txn.Commit(ctx); err != nil {
		return fmt.Errorf("failed to commit relationships: %w", err)
	}

	log.Printf("Successfully created relationships for %d citizens", len(citizenIDs))
	return nil
}
