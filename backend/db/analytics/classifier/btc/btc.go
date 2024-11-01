package btc

import (
	"backend/constants"
	"backend/db"
	"backend/external"
	"context"
	"github.com/dgraph-io/dgo/v240/protos/api"
	"strconv"
)

// ClassifyDestinationAndOriginsByBlock sets the transaction type for destination transactions in the given block range and
// the origin transaction type for all transactions which are connected to mixing
// transactions in this block.
// Destination transactions are transactions which are connected to outputs of mixing transactions and at the
// same time are not mixing transactions themselves. Origin transactions are transactions which are connected to
// inputs of mixing transactions and at the same time are not mixing transactions themselves.
func ClassifyDestinationAndOriginsByBlock(ctx context.Context, c external.Database,
	fromBlockID int64, toBlockID int64) error {
	query := `query Q($from:int,$to:int) {
				b as var(func: between(id, $from, $to))
				var(func: uid(b))@cascade{
					dest as transactions@filter(not has(Transaction.type)){
						tx_inputs{
							~tx_outputs@filter(eq(Transaction.type,"` + constants.TypeWasabi2Mixing + `"))
						}
					}
				}
				var(func: uid(b)){
					transactions@filter(eq(Transaction.type,"` + constants.TypeWasabi2Mixing + `")){
						tx_inputs{
							orig as ~tx_outputs@filter(not has(Transaction.type))
						}
					}
				}
			  }`

	req := &api.Request{
		Query: query,
		Vars: map[string]string{"$from": strconv.FormatInt(fromBlockID, 10),
			"$to": strconv.FormatInt(toBlockID, 10)},
		Mutations: []*api.Mutation{
			{
				Cond:      "@if(gt(len(dest), 0))",
				SetNquads: []byte("uid(dest) <Transaction.type> \"" + constants.TypeDashDestination + "\" ."),
			},
			{
				// only insert origins if there are no transactions to classify
				Cond:      "@if(gt(len(orig), 0))",
				SetNquads: []byte("uid(orig) <Transaction.type> \"" + constants.TypeDashOrigin + "\" ."),
			}},
		CommitNow: true,
	}

	return db.MutationWithRetry(ctx, c, req)
}
