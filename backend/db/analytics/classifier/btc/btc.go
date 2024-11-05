package btc

import (
	"backend/constants"
	"backend/db"
	"backend/external"
	"context"
	"encoding/json"
	"github.com/dgraph-io/dgo/v240/protos/api"
	"github.com/qrest/gomisc/serror"
	"strconv"
)

// ClassifyDestinationAndOriginsByBlock sets the transaction type for destination transactions in the given block range and
// the origin transaction type for all transactions which are connected to mixing
// transactions in this block.
// Destination transactions are transactions which are connected to outputs of mixing transactions and at the
// same time are not mixing transactions themselves. Origin transactions are transactions which are connected to
// inputs of mixing transactions and at the same time are not mixing transactions themselves.
func ClassifyDestinationAndOriginsByBlock(ctx context.Context, c external.Database,
	fromBlockID int64, toBlockID int64) ([]db.Transaction, error) {
	query := `query Q($from:int,$to:int) {
				b as var(func: between(id, $from, $to))
				var(func: uid(b))@cascade{
					wasabi2Destinations as transactions@filter(not has(Transaction.type)){
						tx_inputs{
							~tx_outputs@filter(eq(Transaction.type,"` + constants.TypeWasabi2Mixing + `"))
						}
					}
				}

				var(func: uid(b))@cascade{
					whirlpoolDstTemp as transactions@filter(not has(Transaction.type)){
						tx_inputs{
							~tx_outputs@filter(eq(Transaction.type,"` + constants.TypeWhirlpoolMixing + `"))
						}
					}
				}

				# give wasabi transactions priority
				whirlpoolDestinations as var(func: uid(whirlpoolDstTemp))@filter(not uid(wasabi2Destinations))

				var(func: uid(b)){
					transactions@filter(eq(Transaction.type,"` + constants.TypeWasabi2Mixing + `")){
						tx_inputs{
							wasabi2Origins as ~tx_outputs@filter(not has(Transaction.type))
						}
					}
				}

				var(func: uid(b))@cascade{
					whirlpoolMixing as transactions@filter(not has(Transaction.type)){
						tx_inputs{
							~tx_outputs@filter(eq(Transaction.type,"` + constants.TypeWhirlpoolOrigin + `"))
						}
					}
				}

				q(func: uid(whirlpoolMixing)){
					uid
					txhash
					fee
					Transaction.type
					tx_inputs{
						uid
						amount
						inputindex
						outputindex
					}
					tx_outputs{
						uid
						amount
						inputindex
						outputindex
					}
				}
			  }`

	req := &api.Request{
		Query: query,
		Vars: map[string]string{"$from": strconv.FormatInt(fromBlockID, 10),
			"$to": strconv.FormatInt(toBlockID, 10)},
		Mutations: []*api.Mutation{
			{
				Cond:      "@if(gt(len(whirlpoolDestinations), 0))",
				SetNquads: []byte("uid(whirlpoolDestinations) <Transaction.type> \"" + constants.TypeWhirlpoolDestination + "\" ."),
			},
			{
				Cond:      "@if(gt(len(wasabi2Origins), 0))",
				SetNquads: []byte("uid(wasabi2Origins) <Transaction.type> \"" + constants.TypeWasabi2Origin + "\" ."),
			},
			{
				Cond:      "@if(gt(len(wasabi2Destinations), 0))",
				SetNquads: []byte("uid(wasabi2Destinations) <Transaction.type> \"" + constants.TypeWasabi2Destination + "\" ."),
			},
		},
		CommitNow: true,
	}

	resp, err := db.MutationWithRetryAndResponse(ctx, c, req)
	if err != nil {
		return nil, err
	}
	// json struct
	var r struct {
		Mixing []db.Transaction `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		return nil, serror.New(err)
	}

	return r.Mixing, nil
}
