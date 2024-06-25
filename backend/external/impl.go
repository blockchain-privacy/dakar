package external

import (
	"context"
	"github.com/dgraph-io/dgo/v230"
	"github.com/dgraph-io/dgo/v230/protos/api"
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

// CreateClient create a new dgraph client connecting to the specified host and port
func CreateClient(endpoint string) (Database, *grpc.ClientConn, error) {
	conn, err := grpc.NewClient(endpoint, grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024*1024*1024)))
	if err != nil {
		err = serror.NewStackError(err)
		return nil, conn, err
	}

	return &GraphDB{Dgraph: dgo.NewDgraphClient(api.NewDgraphClient(conn))}, conn, nil
}

// WaitForDatabase waits until the database is ready to receive requests
func WaitForDatabase(c Database) bool {
	for range 20 {
		if IsConnectionEstablished(c) {
			return true
		}

		time.Sleep(time.Second * 5)
	}

	return false
}

// IsConnectionEstablished test the database connection
func IsConnectionEstablished(c Database) bool {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	_, err := c.Query(ctx, "{q(func: uid(0x1)){uid}}", nil)

	return err == nil
}
