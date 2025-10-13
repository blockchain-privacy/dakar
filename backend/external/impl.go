package external

import (
	"context"
	"time"

	"github.com/dgraph-io/dgo/v250"
	"github.com/dgraph-io/dgo/v250/protos/api"
	"github.com/qrest/gomisc/serror"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// GraphDB is a wrapper for Dgraph
type GraphDB struct {
	Dgraph *dgo.Dgraph
}

// Mutate executes a query followed by one or more than one mutation.
func (g *GraphDB) Mutate(ctx context.Context, req *api.Request) (*api.Response, error) {
	resp, err := g.Dgraph.NewTxn().Do(ctx, req)
	if err != nil {
		return nil, serror.New(err)
	}

	return resp, nil
}

// Query but allows a variable map to be used.
// This can provide safety against injection attacks.
func (g *GraphDB) Query(ctx context.Context, q string, vars map[string]string) (*api.Response, error) {
	resp, err := g.Dgraph.NewReadOnlyTxn().QueryWithVars(ctx, q, vars)
	if err != nil {
		return nil, serror.New(err)
	}

	return resp, nil
}

// NewTxn creates a new transaction.
func (g *GraphDB) NewTxn() *dgo.Txn {
	return g.Dgraph.NewTxn()
}

// Close shutdown down all the connections to the Dgraph Cluster.
func (g *GraphDB) Close() {
	g.Dgraph.Close()
}

// DropAll resets the database
func (g *GraphDB) DropAll(ctx context.Context) error {
	if err := g.Dgraph.DropAll(ctx); err != nil {
		return serror.New(err)
	}

	return nil
}

// DropPredicate dops the predicate of the specified namespace
func (g *GraphDB) DropPredicate(ctx context.Context, predicate string) error {
	if err := g.Dgraph.DropPredicate(ctx, predicate); err != nil {
		return serror.New(err)
	}

	return nil
}

// SetSchema sets the schema of the specified namespace
func (g *GraphDB) SetSchema(ctx context.Context, schema string) error {
	if err := g.Dgraph.SetSchema(ctx, schema); err != nil {
		return serror.New(err)
	}

	return nil
}

// CreateClient create a new dgraph client connecting to the specified host and port
func CreateClient(endpoint string) (Database, error) {
	dgraphClient, err := dgo.NewClient(endpoint,
		dgo.WithGrpcOption(grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024*1024*1024))),
		dgo.WithGrpcOption(grpc.WithTransportCredentials(insecure.NewCredentials())))
	if err != nil {
		err = serror.New(err)
		return nil, err
	}

	return &GraphDB{Dgraph: dgraphClient}, nil
}

// WaitForDatabase waits until the database is ready to receive requests
func WaitForDatabase(c Database) bool {
	for range 20 {
		if isConnectionEstablished(c) {
			return true
		}

		time.Sleep(time.Second * 5)
	}

	return false
}

// isConnectionEstablished test the database connection
func isConnectionEstablished(c Database) bool {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*20)
	defer cancel()
	_, err := c.Query(ctx, "{q(func: uid(0x1)){uid}}", nil)

	return err == nil
}
