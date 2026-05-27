// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package btc

import (
	"context"
	"encoding/json"
	"github.com/dgraph-io/dgo/v250/protos/api"
	"gitlab.com/blockchain-privacy/dakar/constants"
	"gitlab.com/blockchain-privacy/dakar/db"
	"gitlab.com/blockchain-privacy/dakar/external"
	"gitlab.com/blockchain-privacy/gomisc/serror"
	"strconv"
)

// ClassifyDestinationAndOriginsByBlock sets the transaction type for
// - wasabi 2.0 origin transactions
// - wasabi 2.0 destination transactions
// - whirlpool destination transactions
// in the given block range.
// Destination transactions are transactions which are connected to outputs of mixing transactions and at the
// same time are not mixing transactions themselves. Origin transactions are transactions which are connected to
// inputs of mixing transactions and at the same time are not mixing transactions themselves.
func ClassifyDestinationAndOriginsByBlock(ctx context.Context, c external.Database,
	fromBlockID int64, toBlockID int64) error {
	query := `query Q($from:int,$to:int,$whirlpoolMixingUIDs:string) {
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
			  }`

	req := &api.Request{
		Query: query,
		Vars:  map[string]string{"$from": strconv.FormatInt(fromBlockID, 10), "$to": strconv.FormatInt(toBlockID, 10)},
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

	return db.MutationWithRetry(ctx, c, req)
}

func GetPotentialWhirlpoolMixingTransactions(ctx context.Context, c external.Database,
	whirlpoolMixingUIDs []string) ([]db.Transaction, map[string][]string, error) {
	query := `query Q($whirlpoolMixingUIDs:string) {
				q(func: uid($whirlpoolMixingUIDs)){
					uid
					tx_inputs{
						~tx_outputs@filter(not eq(Transaction.type,"` + constants.TypeWasabi2Mixing + `")){
							uid
							Transaction.type
							tx_inputs{
								amount
							}
							tx_outputs{
								amount
							}
						}
					}
				}
			  }`

	req := &api.Request{
		Query:     query,
		Vars:      map[string]string{"$whirlpoolMixingUIDs": db.CreateCommaArray(whirlpoolMixingUIDs)},
		CommitNow: true,
	}

	resp, err := db.MutationWithRetryAndResponse(ctx, c, req)
	if err != nil {
		return nil, nil, err
	}
	// json struct
	var r struct {
		Mixing []struct {
			UID    string `json:"uid,omitempty"`
			Inputs []struct {
				Origins []db.Transaction `json:"~tx_outputs,omitempty"`
			} `json:"tx_inputs,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		return nil, nil, serror.New(err)
	}

	originToMixing := map[string][]string{}
	var origins []db.Transaction
	for _, m := range r.Mixing {
		for _, input := range m.Inputs {
			for _, origin := range input.Origins {
				originToMixing[origin.UID] = append(originToMixing[origin.UID], m.UID)
				origins = append(origins, origin)
			}
		}
	}

	return origins, originToMixing, nil
}
