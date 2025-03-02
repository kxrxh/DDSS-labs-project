package db

import (
	"context"
	"embed"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"github.com/kxrxh/ddss/pkg/models"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// go:embed internal/db/schema.dgraph
var embeded embed.FS

// DgraphClient wraps the Dgraph client
type DgraphClient struct {
	*dgo.Dgraph
}

// NewDgraphClient initializes a Dgraph client
func NewDgraphClient(addr string) (*DgraphClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to dial dgraph at %s: %w", addr, err)
	}
	return &DgraphClient{dgo.NewDgraphClient(api.NewDgraphClient(conn))}, nil
}

// SetupDgraphSchema sets up the Dgraph schema for the Social Credit System
func (c *Client) SetupDgraphSchema() error {
	conn, err := grpc.Dial(c.dgraphAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to Dgraph: %w", err)
	}
	defer conn.Close()

	dc := dgo.NewDgraphClient(api.NewDgraphClient(conn))
	schema, err := embeded.ReadFile("internal/db/schema.dgraph")
	if err != nil {
		return fmt.Errorf("failed to read schema: %w", err)
	}
	op := &api.Operation{
		Schema: string(schema),
	}

	if err := dc.Alter(context.Background(), op); err != nil {
		return fmt.Errorf("failed to setup Dgraph schema: %w", err)
	}

	return nil
}

// CreateSocialRelationship creates a relationship between two citizens in the graph
func (c *Client) CreateSocialRelationship(relationship models.GraphRelationship) error {
	conn, err := grpc.Dial(c.dgraphAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to Dgraph: %w", err)
	}
	defer conn.Close()

	dc := dgo.NewDgraphClient(api.NewDgraphClient(conn))
	txn := dc.NewTxn()
	defer txn.Discard(context.Background())

	// First, find the UIDs of both citizens
	q := fmt.Sprintf(`
		{
			fromCitizen(func: eq(simulatedID, "%s")) { uid }
			toCitizen(func: eq(simulatedID, "%s")) { uid }
		}
	`, relationship.FromID, relationship.ToID)

	resp, err := txn.Query(context.Background(), q)
	if err != nil {
		return fmt.Errorf("failed to query citizens: %w", err)
	}

	var result struct {
		FromCitizen []struct {
			UID string `json:"uid"`
		}
		ToCitizen []struct {
			UID string `json:"uid"`
		}
	}

	if err := json.Unmarshal(resp.Json, &result); err != nil {
		return fmt.Errorf("failed to parse query response: %w", err)
	}

	if len(result.FromCitizen) == 0 || len(result.ToCitizen) == 0 {
		return fmt.Errorf("one or both citizens not found")
	}

	fromUID := result.FromCitizen[0].UID
	toUID := result.ToCitizen[0].UID

	// Step 2: Create the relationship edge
	relationshipData := map[string]interface{}{
		"uid":         "_:relationship",
		"dgraph.type": "Relationship",
		"from": map[string]string{
			"uid": fromUID,
		},
		"to": map[string]string{
			"uid": toUID,
		},
		"relationshipType":       relationship.RelationshipType,
		"trust":                  relationship.Trust,
		"influence":              relationship.MutualInfluence,
		"startDate":              relationship.StartDate,
		"interactions":           relationship.Interactions,
		"lastInteraction":        relationship.LastInteraction,
		"strengthScore":          relationship.StrengthScore,
		"communicationFrequency": relationship.CommunicationFrequency,
		"metadata":               relationship.Metadata,
		"isPrimary":              relationship.IsPrimary,
		"isReported":             relationship.IsReported,
		"monitoringLevel":        relationship.MonitoringLevel,
		"flagged":                relationship.Flagged,
		"reportCount":            relationship.ReportCount,
	}

	if !relationship.EndDate.IsZero() {
		relationshipData["endDate"] = relationship.EndDate
	}

	mu := &api.Mutation{
		SetJson: mustMarshal(relationshipData),
	}

	_, err = txn.Mutate(context.Background(), mu)
	if err != nil {
		return fmt.Errorf("failed to create relationship: %w", err)
	}

	return txn.Commit(context.Background())
}

// RecordActivity records a citizen activity in the graph
func (c *Client) RecordActivity(activity models.GraphActivity) error {
	conn, err := grpc.Dial(c.dgraphAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to Dgraph: %w", err)
	}
	defer conn.Close()

	dc := dgo.NewDgraphClient(api.NewDgraphClient(conn))
	txn := dc.NewTxn()
	defer txn.Discard(context.Background())

	// Find the citizen UID
	q := fmt.Sprintf(`
		{
			citizen(func: eq(simulatedID, "%s")) { uid }
		}
	`, activity.CitizenID)

	resp, err := txn.Query(context.Background(), q)
	if err != nil {
		return fmt.Errorf("failed to query citizen: %w", err)
	}

	var result struct {
		Citizen []struct {
			UID string `json:"uid"`
		} `json:"citizen"`
	}

	if err := json.Unmarshal(resp.Json, &result); err != nil {
		return fmt.Errorf("failed to unmarshal query result: %w", err)
	}

	if len(result.Citizen) == 0 {
		return fmt.Errorf("citizen not found: %s", activity.CitizenID)
	}

	citizenUID := result.Citizen[0].UID

	// Create the activity
	activityData := map[string]interface{}{
		"dgraph.type": "Activity",
		"citizen": map[string]string{
			"uid": citizenUID,
		},
		"timestamp":          activity.Timestamp,
		"type":               activity.Type,
		"description":        activity.Description,
		"pointsImpact":       activity.PointsImpact,
		"evidenceHash":       activity.EvidenceHash,
		"verificationStatus": activity.VerificationStatus,
	}

	// Add location if provided
	if activity.LocationID != "" {
		// First, find the location UID
		locQuery := fmt.Sprintf(`
			{
				location(func: eq(locationID, "%s")) { uid }
			}
		`, activity.LocationID)

		locResp, err := txn.Query(context.Background(), locQuery)
		if err != nil {
			return fmt.Errorf("failed to query location: %w", err)
		}

		var locResult struct {
			Location []struct {
				UID string `json:"uid"`
			} `json:"location"`
		}

		if err := json.Unmarshal(locResp.Json, &locResult); err != nil {
			return fmt.Errorf("failed to unmarshal location query result: %w", err)
		}

		if len(locResult.Location) > 0 {
			activityData["location"] = map[string]string{
				"uid": locResult.Location[0].UID,
			}
		}
	}

	// Add associated citizens if provided
	if len(activity.AssociatedCitizenIDs) > 0 {
		associatedUIDs := make([]map[string]string, 0, len(activity.AssociatedCitizenIDs))
		for _, id := range activity.AssociatedCitizenIDs {
			// Find the associated citizen UID
			assocQuery := fmt.Sprintf(`
				{
					citizen(func: eq(simulatedID, "%s")) { uid }
				}
			`, id)

			assocResp, err := txn.Query(context.Background(), assocQuery)
			if err != nil {
				return fmt.Errorf("failed to query associated citizen: %w", err)
			}

			var assocResult struct {
				Citizen []struct {
					UID string `json:"uid"`
				} `json:"citizen"`
			}

			if err := json.Unmarshal(assocResp.Json, &assocResult); err != nil {
				return fmt.Errorf("failed to unmarshal associated citizen query result: %w", err)
			}

			if len(assocResult.Citizen) > 0 {
				associatedUIDs = append(associatedUIDs, map[string]string{
					"uid": assocResult.Citizen[0].UID,
				})
			}
		}

		if len(associatedUIDs) > 0 {
			activityData["associatedCitizens"] = associatedUIDs
		}
	}

	// Add flags if provided
	if len(activity.Flags) > 0 {
		activityData["flags"] = activity.Flags
	}

	mu := &api.Mutation{
		SetJson: mustMarshal(activityData),
	}

	if _, err := txn.Mutate(context.Background(), mu); err != nil {
		return fmt.Errorf("failed to record activity: %w", err)
	}

	return txn.Commit(context.Background())
}

// GetCitizenRelationships retrieves all relationships for a given citizen
func (c *Client) GetCitizenRelationships(citizenID string) ([]models.GraphRelationshipView, error) {
	conn, err := grpc.Dial(c.dgraphAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Dgraph: %w", err)
	}
	defer conn.Close()

	dc := dgo.NewDgraphClient(api.NewDgraphClient(conn))
	txn := dc.NewTxn()
	defer txn.Discard(context.Background())

	// Query both outgoing and incoming relationships
	q := fmt.Sprintf(`
		{
			outgoing(func: eq(simulatedID, "%s")) {
				uid
				simulatedID
				name
				outRelations: ~from @facets(type, strength, trust, influence, startDate, endDate, interactions, lastInteraction) {
					uid
					simulatedID
					name
				}
			}
			
			incoming(func: eq(simulatedID, "%s")) {
				uid
				simulatedID
				name
				inRelations: ~to @facets(type, strength, trust, influence, startDate, endDate, interactions, lastInteraction) {
					uid
					simulatedID
					name
				}
			}
		}
	`, citizenID, citizenID)

	resp, err := txn.Query(context.Background(), q)
	if err != nil {
		return nil, fmt.Errorf("failed to query relationships: %w", err)
	}

	var result struct {
		Outgoing []struct {
			UID          string                           `json:"uid"`
			SimulatedID  string                           `json:"simulatedID"`
			Name         string                           `json:"name"`
			OutRelations []models.GraphRelationshipPerson `json:"outRelations"`
		}
		Incoming []struct {
			UID         string                           `json:"uid"`
			SimulatedID string                           `json:"simulatedID"`
			Name        string                           `json:"name"`
			InRelations []models.GraphRelationshipPerson `json:"inRelations"`
		}
	}

	if err := json.Unmarshal(resp.Json, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal relationships: %w", err)
	}

	relationships := []models.GraphRelationshipView{}

	// Convert outgoing relationships
	if len(result.Outgoing) > 0 && len(result.Outgoing[0].OutRelations) > 0 {
		for _, rel := range result.Outgoing[0].OutRelations {
			relationships = append(relationships, models.GraphRelationshipView{
				CitizenID:          citizenID,
				RelatedCitizenID:   rel.SimulatedID,
				RelatedCitizenName: rel.Name,
				RelationshipType:   rel.Type,
				Strength:           rel.Strength,
				Trust:              rel.Trust,
				Influence:          rel.Influence,
				StartDate:          rel.StartDate,
				EndDate:            rel.EndDate,
				Interactions:       rel.Interactions,
				LastInteraction:    rel.LastInteraction,
				Direction:          "outgoing",
			})
		}
	}

	// Convert incoming relationships
	if len(result.Incoming) > 0 && len(result.Incoming[0].InRelations) > 0 {
		for _, rel := range result.Incoming[0].InRelations {
			relationships = append(relationships, models.GraphRelationshipView{
				CitizenID:          citizenID,
				RelatedCitizenID:   rel.SimulatedID,
				RelatedCitizenName: rel.Name,
				RelationshipType:   rel.Type,
				Strength:           rel.Strength,
				Trust:              rel.Trust,
				Influence:          rel.Influence,
				StartDate:          rel.StartDate,
				EndDate:            rel.EndDate,
				Interactions:       rel.Interactions,
				LastInteraction:    rel.LastInteraction,
				Direction:          "incoming",
			})
		}
	}

	return relationships, nil
}

// GetSocialConnectionPath finds connection paths between two citizens in the social graph
func (c *Client) GetSocialConnectionPath(fromID, toID string, maxDepth int) ([]models.GraphSocialPath, error) {
	conn, err := grpc.Dial(c.dgraphAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Dgraph: %w", err)
	}
	defer conn.Close()

	dc := dgo.NewDgraphClient(api.NewDgraphClient(conn))
	txn := dc.NewTxn()
	defer txn.Discard(context.Background())

	// Limit the depth for performance reasons
	if maxDepth <= 0 || maxDepth > 6 {
		maxDepth = 4 // Default max depth
	}

	// Query to find the shortest path
	q := fmt.Sprintf(`
		{
			shortest(from: %s, to: %s) @facets {
				path: as(shortest_path(from: %s, to: %s, numpaths: 1)) {
					uid
					simulatedID
					name
					from @facets(type, strength, trust, influence, startDate)
					to @facets(type, strength, trust, influence, startDate)
				}
			}
		}
	`,
		// We need placeholders for the from/to UIDs
		"uid(from)", "uid(to)",
		"uid(from)", "uid(to)")

	// Add the parameter map
	vars := map[string]string{
		"$from_id": fromID,
		"$to_id":   toID,
	}

	// Add variables definitions for the UIDs
	q = fmt.Sprintf(`
		query shortest($from_id: string, $to_id: string) {
			from as var(func: eq(simulatedID, $from_id))
			to as var(func: eq(simulatedID, $to_id))
			%s
		}
	`, q)

	resp, err := txn.QueryWithVars(context.Background(), q, vars)
	if err != nil {
		return nil, fmt.Errorf("failed to query social path: %w", err)
	}

	var result struct {
		Shortest []struct {
			Path []struct {
				UID         string  `json:"uid"`
				SimulatedID string  `json:"simulatedID"`
				Name        string  `json:"name"`
				Type        string  `json:"type"`
				Strength    float64 `json:"strength"`
				Trust       float64 `json:"trust"`
				Influence   float64 `json:"influence"`
				StartDate   string  `json:"startDate"`
			} `json:"path"`
		} `json:"shortest"`
	}

	if err := json.Unmarshal(resp.Json, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal social path result: %w", err)
	}

	// Convert the result to SocialPath objects
	path := result.Shortest[0].Path
	socialPaths := make([]models.GraphSocialPath, 0)

	for i := 0; i < len(path)-1; i++ {
		startDate, _ := time.Parse(time.RFC3339, path[i].StartDate)
		if startDate.IsZero() {
			startDate = time.Now() // Default to now if not provided
		}

		socialPaths = append(socialPaths, models.GraphSocialPath{
			FromID:           path[i].SimulatedID,
			FromName:         path[i].Name,
			ToID:             path[i+1].SimulatedID,
			ToName:           path[i+1].Name,
			RelationshipType: path[i].Type,
			Strength:         path[i].Strength,
			Trust:            path[i].Trust,
			Influence:        path[i].Influence,
			StartDate:        startDate,
		})
	}

	return socialPaths, nil
}

// IdentifyInfluentialNodes identifies the most influential citizens in the social graph
func (c *Client) IdentifyInfluentialNodes(limit int) ([]models.GraphInfluentialCitizen, error) {
	conn, err := grpc.Dial(c.dgraphAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Dgraph: %w", err)
	}
	defer conn.Close()

	dc := dgo.NewDgraphClient(api.NewDgraphClient(conn))
	txn := dc.NewTxn()
	defer txn.Discard(context.Background())

	// Use a reasonable default limit
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	// Query to find influential citizens based on relationship metrics
	// This uses a custom score combining number of relationships, trust, influence
	q := fmt.Sprintf(`
		{
			var(func: type(Citizen)) {
				relationships as count(~from)
				incomingInfluence as sum(~from) @facets(influence)
				socialScore as math(relationships * 0.5 + incomingInfluence * 0.5)
			}
			
			influencers(func: uid(socialScore), orderdesc: val(socialScore), first: %d) {
				uid
				simulatedID
				name
				socialInfluenceScore: val(socialScore)
				relationshipCount: val(relationships)
				incomingTrustSum: sum(~from) @facets(trust)
				outgoingCount: count(from)
				
				topConnections: from (orderdesc: influence, first: 5) {
					uid
					simulatedID
					name
					influence @facets
					trust @facets
				}
			}
		}
	`, limit)

	resp, err := txn.Query(context.Background(), q)
	if err != nil {
		return nil, fmt.Errorf("failed to query influential nodes: %w", err)
	}

	var result struct {
		Influencers []struct {
			UID                  string  `json:"uid"`
			SimulatedID          string  `json:"simulatedID"`
			Name                 string  `json:"name"`
			SocialInfluenceScore float64 `json:"socialInfluenceScore"`
			RelationshipCount    int     `json:"relationshipCount"`
			IncomingTrustSum     float64 `json:"incomingTrustSum"`
			OutgoingCount        int     `json:"outgoingCount"`
			TopConnections       []struct {
				UID         string  `json:"uid"`
				SimulatedID string  `json:"simulatedID"`
				Name        string  `json:"name"`
				Influence   float64 `json:"influence"`
				Trust       float64 `json:"trust"`
			} `json:"topConnections"`
		} `json:"influencers"`
	}

	if err := json.Unmarshal(resp.Json, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal influential nodes result: %w", err)
	}

	// Convert the result to InfluentialCitizen objects
	influencers := make([]models.GraphInfluentialCitizen, 0, len(result.Influencers))
	for _, inf := range result.Influencers {
		influencers = append(influencers, models.GraphInfluentialCitizen{
			CitizenID:            inf.SimulatedID,
			Name:                 inf.Name,
			InfluenceScore:       inf.SocialInfluenceScore,
			RelationshipCount:    inf.RelationshipCount,
			IncomingTrustSum:     inf.IncomingTrustSum,
			OutgoingCount:        inf.OutgoingCount,
			TopConnectedCitizens: mapTopConnections(inf.TopConnections),
		})
	}

	return influencers, nil
}

// mapTopConnections converts the TopConnections data to the expected format
func mapTopConnections(connections []struct {
	UID         string  `json:"uid"`
	SimulatedID string  `json:"simulatedID"`
	Name        string  `json:"name"`
	Influence   float64 `json:"influence"`
	Trust       float64 `json:"trust"`
}) []models.GraphConnectedCitizen {
	result := make([]models.GraphConnectedCitizen, 0, len(connections))
	for _, conn := range connections {
		result = append(result, models.GraphConnectedCitizen{
			CitizenID: conn.SimulatedID,
			Name:      conn.Name,
			Influence: conn.Influence,
			Trust:     conn.Trust,
		})
	}
	return result
}

// Helper function to marshal JSON and handle errors
func mustMarshal(v interface{}) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("failed to marshal JSON: %v", err))
	}
	return b
}
