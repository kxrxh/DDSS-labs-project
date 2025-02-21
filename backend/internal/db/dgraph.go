package db

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// DgraphClient wraps the Dgraph client
type DgraphClient struct {
	*dgo.Dgraph
}

// NewDgraphClient initializes a Dgraph client
func NewDgraphClient(addr string) (*DgraphClient, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to dial dgraph at %s: %w", addr, err)
	}
	return &DgraphClient{dgo.NewDgraphClient(api.NewDgraphClient(conn))}, nil
}

// SetupDgraphSchema defines the schema for citizens and relationships
func (c *Client) SetupDgraphSchema() error {
	schema := `
		type Citizen {
			simulated_id: string
			name: string
			relations: [uid] @reverse
		}

		simulated_id: string @index(exact) .
		name: string .
		relations: [uid] .
	`

	err := c.graph.Alter(context.Background(), &api.Operation{
		Schema: schema,
	})
	if err != nil {
		return fmt.Errorf("failed to set dgraph schema: %w", err)
	}
	return nil
}

// AddCitizen adds a citizen node with optional relations
func (c *Client) AddCitizen(simulatedID, name string, relationIDs []string) (string, error) {
	ctx := context.Background()
	txn := c.graph.NewTxn()
	defer txn.Discard(ctx)

	citizen := map[string]interface{}{
		"simulated_id": simulatedID,
		"name":         name,
		"dgraph.type":  "Citizen",
	}
	if len(relationIDs) > 0 {
		relations := make([]map[string]string, len(relationIDs))
		for i, id := range relationIDs {
			relations[i] = map[string]string{"simulated_id": id}
		}
		citizen["relations"] = relations
	}

	jsonData, err := json.Marshal(citizen)
	if err != nil {
		return "", fmt.Errorf("failed to marshal citizen: %w", err)
	}

	mu := &api.Mutation{SetJson: jsonData}
	resp, err := txn.Mutate(ctx, mu)
	if err != nil {
		return "", fmt.Errorf("failed to mutate citizen: %w", err)
	}

	err = txn.Commit(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Return the UID of the created citizen
	return resp.Uids["blank-0"], nil
}

// GetCitizenRelations retrieves relations for a given citizen
func (c *Client) GetCitizenRelations(simulatedID string) ([]string, error) {
	ctx := context.Background()
	query := `
		query Citizen($id: string) {
			citizen(func: eq(simulated_id, $id)) {
				simulated_id
				name
				relations {	
					simulated_id
				}
			}
		}
	`
	vars := map[string]string{"$id": simulatedID}
	resp, err := c.graph.NewTxn().QueryWithVars(ctx, query, vars)
	if err != nil {
		return nil, fmt.Errorf("query failed: %w", err)
	}

	type relation struct {
		SimulatedID string `json:"simulated_id"`
	}
	type citizen struct {
		Relations []relation `json:"relations"`
	}
	type response struct {
		Citizen []citizen `json:"citizen"`
	}

	var result response
	if err := json.Unmarshal(resp.Json, &result); err != nil {
		return nil, fmt.Errorf("unmarshal failed: %w", err)
	}

	if len(result.Citizen) == 0 {
		return nil, nil // No citizen found
	}

	relationIDs := make([]string, len(result.Citizen[0].Relations))
	for i, rel := range result.Citizen[0].Relations {
		relationIDs[i] = rel.SimulatedID
	}
	return relationIDs, nil
}
