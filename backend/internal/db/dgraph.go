package db

import (
	"context"
	"encoding/json"
	"fmt"

	dgo "github.com/dgraph-io/dgo/v230"
	api "github.com/dgraph-io/dgo/v230/protos/api"
	"google.golang.org/grpc"
)

type DgraphClient struct {
	client *dgo.Dgraph
}

type Citizen struct {
	UID    string  `json:"uid,omitempty"`
	Name   string  `json:"name,omitempty"`
	Score  float64 `json:"score,omitempty"`
	Region string  `json:"region,omitempty"`
	Events []Event `json:"events,omitempty"`
}

type Event struct {
	UID         string  `json:"uid,omitempty"`
	Type        string  `json:"type,omitempty"`
	ScoreChange float64 `json:"score_change,omitempty"`
	Description string  `json:"description,omitempty"`
	Timestamp   string  `json:"timestamp,omitempty"`
}

func NewDgraphClient(addr string) (*DgraphClient, error) {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	client := dgo.NewDgraphClient(api.NewDgraphClient(conn))
	return &DgraphClient{client: client}, nil
}

func (d *DgraphClient) GetCitizen(ctx context.Context, id string) (*Citizen, error) {
	const q = `
        query citizen($id: string) {
            citizen(func: uid($id)) {
                uid
                name
                score
                region
                events {
                    uid
                    type
                    score_change
                    description
                    timestamp
                }
            }
        }
    `

	vars := map[string]string{"$id": id}
	resp, err := d.client.NewTxn().QueryWithVars(ctx, q, vars)
	if err != nil {
		return nil, err
	}

	type result struct {
		Citizen []Citizen `json:"citizen"`
	}

	var r result
	if err := json.Unmarshal(resp.Json, &r); err != nil {
		return nil, err
	}

	if len(r.Citizen) == 0 {
		return nil, fmt.Errorf("citizen not found")
	}

	return &r.Citizen[0], nil
}

func (d *DgraphClient) UpdateScore(ctx context.Context, id string, newScore float64) error {
	mu := &api.Mutation{
		SetNquads: []byte(fmt.Sprintf(`<%s> <score> "%f" .`, id, newScore)),
	}

	_, err := d.client.NewTxn().Mutate(ctx, mu)
	return err
}

func (d *DgraphClient) AddEvent(ctx context.Context, citizenID string, event Event) error {
	txn := d.client.NewTxn()
	defer txn.Discard(ctx)

	pb, err := json.Marshal(event)
	if err != nil {
		return err
	}

	mu := &api.Mutation{
		SetJson: pb,
	}

	_, err = txn.Mutate(ctx, mu)
	if err != nil {
		return err
	}

	return txn.Commit(ctx)
} 