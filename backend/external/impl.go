// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package external

import (
	"context"
	"time"

	"github.com/dgraph-io/dgo/v250"
	"github.com/dgraph-io/dgo/v250/protos/api"
	"gitlab.com/blockchain-privacy/gomisc/serror"
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

// DropAll drops all data and the schema of the database
func (g *GraphDB) DropAll(ctx context.Context) error {
	if err := g.Dgraph.DropAll(ctx); err != nil {
		return serror.New(err)
	}

	return nil
}

// DropData drops all data of the database
func (g *GraphDB) DropData(ctx context.Context) error {
	if err := g.Dgraph.DropData(ctx); err != nil {
		return serror.New(err)
	}

	return nil
}

// DropNamespace drops all data of the namespace
func (g *GraphDB) DropNamespace(ctx context.Context, nsID uint64) error {
	if err := g.Dgraph.DropNamespace(ctx, nsID); err != nil {
		return serror.New(err)
	}

	return nil
}

// DropPredicate drops the predicate of the specified namespace
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

func (g *GraphDB) CreateNamespace(ctx context.Context) (uint64, error) {
	nsID, err := g.Dgraph.CreateNamespace(ctx)
	if err != nil {
		return 0, serror.New(err)
	}

	return nsID, nil
}

// CreateClient creates a new dgraph client connecting to the specified endpoint
func CreateClient(endpoint string) (Database, error) {
	for range 20 {
		dgraphClient, err := dgo.NewClient(endpoint,
			dgo.WithGrpcOption(grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024*1024*1024))),
			dgo.WithGrpcOption(grpc.WithTransportCredentials(insecure.NewCredentials())))
		if err == nil {
			return &GraphDB{Dgraph: dgraphClient}, nil
		}

		time.Sleep(time.Second * 5)
	}

	return nil, serror.FromStr("could not connect to database")
}

// CreateClientWithNamespace create a new dgraph client connecting to the specified endpoint and namespace
func CreateClientWithNamespace(ctx context.Context, endpoint, user, password string,
	namespaceID uint64) (Database, error) {
	for range 20 {
		dgraphClient, err := dgo.NewClient(endpoint,
			dgo.WithGrpcOption(grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(1024*1024*1024))),
			dgo.WithGrpcOption(grpc.WithTransportCredentials(insecure.NewCredentials())))
		if err == nil {
			err = dgraphClient.LoginIntoNamespace(ctx, user, password, namespaceID)
			if err == nil {
				return &GraphDB{Dgraph: dgraphClient}, nil
			}
		}

		time.Sleep(time.Second * 5)
	}

	return nil, serror.FromStr("could not connect to database")
}
