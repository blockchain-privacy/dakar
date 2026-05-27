// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package analytics

import (
	"context"
	"github.com/dgraph-io/dgo/v250/protos/api"
	"gitlab.com/blockchain-privacy/dakar/db"
	"gitlab.com/blockchain-privacy/dakar/external"
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
