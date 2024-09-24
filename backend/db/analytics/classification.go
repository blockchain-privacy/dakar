package analytics

import (
	"backend/constants"
	"backend/db"
	"backend/external"
	"context"
	"github.com/qrest/gomisc/serror"

	"encoding/json"
	"strconv"
	"time"

	"github.com/dgraph-io/dgo/v240/protos/api"
)

// ClassifyDestinationAndOriginsByBlock sets the privacy type for destination transactions in the given block range and
// the origin privacy type for all transactions which are connected to mixing
// transactions in this block. Additionally, it returns all transactions connected to newly
// classified origin transaction which have no privacy type set yet.
// Destination transactions are transactions which are connected to outputs of mixing transactions and at the
// same time are not mixing transactions themselves. Origin transactions are transactions which are connected to
// inputs of mixing transactions and at the same time are not mixing transactions themselves.
func ClassifyDestinationAndOriginsByBlock(c external.Database, fromBlockID int64, toBlockID int64) (toClassify []db.Transaction,
	origins []db.Transaction, err error) {
	query := `query Q($from:int,$to:int) {
				b as var(func: between(id, $from, $to))
				var(func: uid(b))@cascade{
					dest as transactions@filter(not has(Transaction.type)){
						tx_inputs{
							~tx_outputs@filter(eq(Transaction.type,"` + constants.TypeMixing + `"))
						}
					}
				}
				var(func: uid(b)){
					transactions@filter(eq(Transaction.type,"` + constants.TypeMixing + `")){
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
				SetNquads: []byte("uid(dest) <Transaction.type> \"" + constants.TypeDestination + "\" ."),
			},
			{
				// only insert origins if there are no transactions to classify
				Cond:      "@if(gt(len(orig), 0) and eq(len(to_classify),0))",
				SetNquads: []byte("uid(orig) <Transaction.type> \"" + constants.TypeOrigin + "\" ."),
			}},
		CommitNow: true,
	}
	resp, err := db.TxWithRetryAndResponse(c, time.Minute*10, req)
	if err != nil {
		return
	}

	// json struct
	var r struct {
		Collaterals []db.Transaction `json:"q,omitempty"`
		Origins     []db.Transaction `json:"o,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = serror.New(err)
		return
	}

	toClassify = r.Collaterals
	origins = r.Origins

	return
}

// SetCollateralCreation sets the collateral creation privacy type, if its input transaction are
// either of the type origin, mixing or collateral creation. Returns the number of newly
// classified transactions.
func SetCollateralCreation(c external.Database, txUids []string) (insertCount int, err error) {
	uidList := db.CreateCommaArray(txUids)

	const query = `query Q($uids: string) {
				cc as var(func: uid($uids))@filter(not has(Transaction.type) or eq(Transaction.type,"` + constants.TypeDestination + `"))@cascade{	
					tx_inputs{
						~tx_outputs@filter(eq(Transaction.type,"` + constants.TypeMixing +
		`") or eq(Transaction.type,"` + constants.TypeOrigin + `") or eq(Transaction.type,"` + constants.TypeCC + `"))}
				}
				q(func: uid(cc)){count(uid)}
			  }`

	const nQuad = "uid(cc) <Transaction.type> \"" + constants.TypeCC + "\" ."

	req := &api.Request{
		Query: query,
		Vars:  map[string]string{"$uids": uidList},
		Mutations: []*api.Mutation{{
			Cond:      "@if(gt(len(cc), 0))",
			SetNquads: []byte(nQuad),
		}},
		CommitNow: true,
	}

	resp, err := db.TxWithRetryAndResponse(c, time.Minute*10, req)
	if err != nil {
		return
	}

	// json struct
	var r struct {
		Query []struct {
			Count int `json:"count,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
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

// SetCollateralPayment sets the collateral payment privacy type, if its input transaction are
// either of the type origin, collateral creation or collateral payment. Returns the number
// of newly classified transactions.
func SetCollateralPayment(c external.Database, txUids []string) (insertCount int, err error) {
	uidList := db.CreateCommaArray(txUids)

	// collateral payments + collateral creations + origins
	const filter = "@filter(eq(Transaction.type,\"" + constants.TypeOrigin + "\") or eq(Transaction.type,\"" + constants.TypeCC + "\")or eq(Transaction.type,\"" + constants.TypeCP + "\"))"

	const query = `query Q($uids: string) {
				cp as var(func: uid($uids))@filter(not has(Transaction.type) or eq(Transaction.type,"` + constants.TypeDestination + `"))@cascade{	
					tx_inputs{
						~tx_outputs` + filter + `}
				}
				q(func: uid(cp)){count(uid)}
			  }`

	req := &api.Request{
		Query: query,
		Vars:  map[string]string{"$uids": uidList},
		Mutations: []*api.Mutation{{
			Cond:      "@if(gt(len(cp), 0))",
			SetNquads: []byte("uid(cp) <Transaction.type> \"" + constants.TypeCP + "\" ."),
		}},
		CommitNow: true,
	}

	resp, err := db.TxWithRetryAndResponse(c, time.Minute*10, req)
	if err != nil {
		return
	}

	// json struct
	var r struct {
		Query []struct {
			Count int `json:"count,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
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
func GetCollateralInputTransactions(c external.Database, txUids []string,
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

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*5, query,
		map[string]string{"$uids": db.CreateCommaArray(txUids), "$bid": strconv.FormatInt(blockHeight, 10)})
	if err != nil {
		return
	}

	var r struct {
		Q []db.Transaction `json:"q"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = serror.New(err)
		return
	}

	outputTransactions = r.Q

	return
}

// RemoveTransactionTypeOfAllTransactions removes the transaction type of all transactions
func RemoveTransactionTypeOfAllTransactions(ctx context.Context, c external.Database) (err error) {
	req := &api.Request{
		Query:     "{t as var(func: has(txhash))}",
		Mutations: []*api.Mutation{{DelNquads: []byte("uid(t) <Transaction.type> * .")}},
		CommitNow: true,
	}

	return db.MutationWithRetry(ctx, c, req)
}
