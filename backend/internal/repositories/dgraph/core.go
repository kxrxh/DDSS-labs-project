package dgraph

import (
	"context"
	"fmt"

	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/kxrxh/ddss/internal/repositories"
)

var _ repositories.Repository = (*Repository)(nil)

type Repository struct {
	dg   *dgo.Dgraph
	conn *grpc.ClientConn
}

// New creates and returns a new Repository instance connected to Dgraph.
// It takes one or more target addresses (e.g., "localhost:9080") for the Dgraph alpha nodes.
func New(ctx context.Context, addresses ...string) (*Repository, error) {
	if len(addresses) == 0 {
		return nil, fmt.Errorf("at least one dgraph alpha address must be provided")
	}

	// Dial a gRPC connection.
	// Consider using secure credentials in production.
	conn, err := grpc.NewClient(addresses[0], grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to dial dgraph grpc connection: %w", err)
	}

	// Create a new Dgraph client.
	dgraphClient := dgo.NewDgraphClient(api.NewDgraphClient(conn))

	// Optional: You might want to verify the connection here if possible,
	// e.g., by checking cluster health or performing a simple query.

	return &Repository{dg: dgraphClient, conn: conn}, nil
}

// Close closes the underlying gRPC connection to Dgraph.
func (r *Repository) Close() error {
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}

// Client returns the underlying Dgraph client.
// Use this for performing Dgraph-specific operations.
func (r *Repository) Client() *dgo.Dgraph {
	return r.dg
}
