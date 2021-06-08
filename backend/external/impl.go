package external

import (
	"context"
	"github.com/dgraph-io/dgo/v210"
	"github.com/dgraph-io/dgo/v210/protos/api"
)

type GraphDB struct {
	Dgraph *dgo.Dgraph
}

func (g *GraphDB) Mutate(ctx context.Context, req *api.Request) (*api.Response, error) {
	return g.Dgraph.NewTxn().Do(ctx, req)
}

func (g *GraphDB) Query(ctx context.Context, q string, vars map[string]string) (*api.Response, error) {
	return g.Dgraph.NewReadOnlyTxn().QueryWithVars(ctx, q, vars)
}

func (g *GraphDB) Alter(ctx context.Context, op *api.Operation) error {
	return g.Dgraph.Alter(ctx, op)
}
