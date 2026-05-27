// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package dash

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

// ClassifyDestinationAndOriginsByBlock sets the transaction type for destination transactions in the given block range and
// the origin transaction type for all transactions which are connected to mixing
// transactions in this block. Additionally, it returns all transactions connected to newly
// classified origin transaction which have no transaction type set yet.
// Destination transactions are transactions which are connected to outputs of mixing transactions and at the
// same time are not mixing transactions themselves. Origin transactions are transactions which are connected to
// inputs of mixing transactions and at the same time are not mixing transactions themselves.
func ClassifyDestinationAndOriginsByBlock(ctx context.Context, c external.Database,
	fromBlockID int64, toBlockID int64) (toClassify []db.Transaction,
	origins []db.Transaction, err error) {
	query := `query Q($from:int,$to:int) {
				b as var(func: between(id, $from, $to))
				var(func: uid(b))@cascade{
					dest as transactions@filter(not has(Transaction.type)){
						tx_inputs{
							~tx_outputs@filter(eq(Transaction.type,"` + constants.TypeDashMixing + `"))
						}
					}
				}
				var(func: uid(b)){
					transactions@filter(eq(Transaction.type,"` + constants.TypeDashMixing + `")){
						tx_inputs{
							orig as ~tx_outputs@filter(not has(Transaction.type))
						}
					}
				}
	
				var(func: uid(orig)){
					tx_outputs{
						# do not limit by number of inputs as there could be multiple with the same address
						to_classify as ~tx_inputs@filter(not has(Transaction.type) and le(count(tx_outputs),2))@cascade{
							~transactions@filter(le(id,$to))
						}
					}
				}
	
				q(func: uid(to_classify)){
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

				o(func: uid(orig)){
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
				Cond:      "@if(gt(len(dest), 0))",
				SetNquads: []byte("uid(dest) <Transaction.type> \"" + constants.TypeDashDestination + "\" ."),
			},
			{
				// only insert origins if there are no transactions to classify
				Cond:      "@if(gt(len(orig), 0) and eq(len(to_classify),0))",
				SetNquads: []byte("uid(orig) <Transaction.type> \"" + constants.TypeDashOrigin + "\" ."),
			}},
		CommitNow: true,
	}
	resp, err := db.MutationWithRetryAndResponse(ctx, c, req)
	if err != nil {
		return
	}

	// json struct
	var r struct {
		Collaterals []db.Transaction `json:"q,omitempty"`
		Origins     []db.Transaction `json:"o,omitempty"`
	}

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		err = serror.New(err)
		return
	}

	toClassify = r.Collaterals
	origins = r.Origins

	return
}

// SetCollateralCreation sets the collateral creation transaction type, if its input transaction are
// either of the type origin, mixing or collateral creation. Returns the number of newly
// classified transactions.
func SetCollateralCreation(ctx context.Context, c external.Database, txUids []string) (insertCount int, err error) {
	const query = `query Q($uids: string) {
				cc as var(func: uid($uids))@filter(not has(Transaction.type) or eq(Transaction.type,"` + constants.TypeDashDestination + `"))@cascade{	
					tx_inputs{
						~tx_outputs@filter(eq(Transaction.type,"` + constants.TypeDashMixing +
		`") or eq(Transaction.type,"` + constants.TypeDashOrigin + `") or eq(Transaction.type,"` + constants.TypeDashCC + `"))}
				}
				q(func: uid(cc)){count(uid)}
			  }`

	const nQuad = "uid(cc) <Transaction.type> \"" + constants.TypeDashCC + "\" ."

	req := &api.Request{
		Query: query,
		Vars:  map[string]string{"$uids": db.CreateCommaArray(txUids)},
		Mutations: []*api.Mutation{{
			Cond:      "@if(gt(len(cc), 0))",
			SetNquads: []byte(nQuad),
		}},
		CommitNow: true,
	}

	resp, err := db.MutationWithRetryAndResponse(ctx, c, req)
	if err != nil {
		return
	}

	// json struct
	var r struct {
		Query []struct {
			Count int `json:"count,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		err = serror.New(err)
		return
	}

	if len(r.Query) != 1 {
		err = serror.FromStr("wrong number of returned query counts")
		return
	}

	insertCount = r.Query[0].Count

	return
}

// SetCollateralPayment sets the collateral payment transaction type, if its input transaction are
// either of the type origin, collateral creation or collateral payment. Returns the number
// of newly classified transactions.
func SetCollateralPayment(ctx context.Context, c external.Database, txUids []string) (insertCount int, err error) {
	// collateral payments + collateral creations + origins
	const filter = "@filter(eq(Transaction.type,\"" + constants.TypeDashOrigin + "\") or eq(Transaction.type,\"" + constants.TypeDashCC + "\")or eq(Transaction.type,\"" + constants.TypeDashCP + "\"))"

	const query = `query Q($uids: string) {
				cp as var(func: uid($uids))@filter(not has(Transaction.type) or eq(Transaction.type,"` + constants.TypeDashDestination + `"))@cascade{	
					tx_inputs{
						~tx_outputs` + filter + `}
				}
				q(func: uid(cp)){count(uid)}
			  }`

	req := &api.Request{
		Query: query,
		Vars:  map[string]string{"$uids": db.CreateCommaArray(txUids)},
		Mutations: []*api.Mutation{{
			Cond:      "@if(gt(len(cp), 0))",
			SetNquads: []byte("uid(cp) <Transaction.type> \"" + constants.TypeDashCP + "\" ."),
		}},
		CommitNow: true,
	}

	resp, err := db.MutationWithRetryAndResponse(ctx, c, req)
	if err != nil {
		return
	}

	// json struct
	var r struct {
		Query []struct {
			Count int `json:"count,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		err = serror.New(err)
		return
	}

	if len(r.Query) != 1 {
		err = serror.FromStr("wrong number of returned query counts")
		return
	}

	insertCount = r.Query[0].Count

	return
}

// GetCollateralInputTransactions returns the input transactions of
// the provided transactions until the given block height
func GetCollateralInputTransactions(ctx context.Context, c external.Database, txUids []string,
	blockHeight int64) (outputTransactions []db.Transaction, err error) {
	const query = `query Q($uids:string,$bid:int){
				var (func: uid($uids)){
					tx_outputs{
						v as ~tx_inputs@filter(le(count(tx_outputs),2))@cascade{
							~transactions@filter(le(id,$bid))
						}
					}
				}

				q(func: uid(v)){
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

	resp, err := db.QueryVarWithRetry(ctx, c, query,
		map[string]string{"$uids": db.CreateCommaArray(txUids), "$bid": strconv.FormatInt(blockHeight, 10)})
	if err != nil {
		return
	}

	var r struct {
		Q []db.Transaction `json:"q"`
	}

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		err = serror.New(err)
		return
	}

	outputTransactions = r.Q

	return
}
