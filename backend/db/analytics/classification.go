package analytics

import (
	"backend/db"
	"backend/external"
	"context"
	"github.com/dgraph-io/dgo/v250/protos/api"
)

// RemoveTransactionTypeOfAllTransactions removes the transaction type of all transactions.
// Used for testing.
func RemoveTransactionTypeOfAllTransactions(ctx context.Context, c external.Database) (err error) {
	req := &api.Request{
		Query:     "{t as var(func: has(txhash))}",
		Mutations: []*api.Mutation{{DelNquads: []byte("uid(t) <Transaction.type> * .")}},
		CommitNow: true,
	}

	return db.MutationWithRetry(ctx, c, req)
}
