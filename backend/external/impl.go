package external

import (
	"context"
	"github.com/dgraph-io/dgo/v240"
	"github.com/dgraph-io/dgo/v240/protos/api"
	"github.com/qrest/gomisc/serror"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"time"
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
//  1. Modify the schema.
//  2. Drop a predicate.
//  3. Drop the database.
func (g *GraphDB) Alter(ctx context.Context, op *api.Operation) error {
	return g.Dgraph.Alter(ctx, op)
}

// NewTxn creates a new transaction.
func (g *GraphDB) NewTxn() *dgo.Txn {
	return g.Dgraph.NewTxn()
}

// Close shutdown down all the connections to the Dgraph Cluster.
func (g *GraphDB) Close() {
	g.Dgraph.Close()
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
