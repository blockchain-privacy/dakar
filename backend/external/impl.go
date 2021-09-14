package external

import (
	"context"
	"github.com/dgraph-io/dgo/v210"
	"github.com/dgraph-io/dgo/v210/protos/api"
)

// GraphDB is a wrapper for Dgraph
type GraphDB struct {
	Dgraph *dgo.Dgraph
}

// Mutate executes a query followed by one or more than one mutation.
func (g *GraphDB) Mutate(ctx context.Context, req *api.Request) (*api.Response, error) {
	return g.Dgraph.NewTxn().Do(ctx, req)
}

// Query but allows a variable map to be used.
// This can provide safety against injection attacks.
func (g *GraphDB) Query(ctx context.Context, q string, vars map[string]string) (*api.Response, error) {
	return g.Dgraph.NewReadOnlyTxn().QueryWithVars(ctx, q, vars)
}

// Alter can be used to do the following by setting various fields of api.Operation:
//   1. Modify the schema.
//   2. Drop a predicate.
//   3. Drop the database.
func (g *GraphDB) Alter(ctx context.Context, op *api.Operation) error {
	return g.Dgraph.Alter(ctx, op)
}

// NewTxn creates a new transaction.
func (g *GraphDB) NewTxn() *dgo.Txn {
	return g.Dgraph.NewTxn()
}
