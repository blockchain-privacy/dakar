package analytics

import (
	"backend/cmd/cliutil"
	"backend/constants"
	"backend/db"
	dbtx "backend/db/transaction"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"strconv"
	"time"
)

// DoClassification sets the privacy type for destination transactions in the given block and
// the origin privacy type for all transactions which are connected to mixing
// transactions in this block. Additionally, it returns all transactions connected to newly
// classified origin transaction which have no privacy type set yet.
func DoClassification(c *dgo.Dgraph, blockId uint64) (toClassify []dbtx.Transaction,
	origins []dbtx.Transaction, err error) {
	const query = `query Q($bid: string) {
				b as var(func: eq(id,$bid)){t as ts}
				var(func: uid(b))@cascade{
					dest as transactions@filter(not has(privacytype)){
						tx_inputs{
							~tx_outputs@filter(le(privacytype,` + constants.StrPrivacyMixingLast + `))
						}
					}
				}
				var(func: uid(b)){
					transactions@filter(le(privacytype,` + constants.StrPrivacyMixingLast + `)){
						tx_inputs{
							orig as ~tx_outputs@filter(not has(privacytype))
						}
					}
				}
	
				var(func: uid(orig)){
					tx_outputs{
						# do not limit by number of inputs as there could be multiple with the same address
						to_classify as ~tx_inputs@filter(not has(privacytype) and le(count(tx_outputs),2))@cascade{
							~transactions@filter(le(ts,val(t)))
						}
					}
				}
	
				q(func: uid(to_classify)){
					uid
					txhash
					fee
					privacytype
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
					privacytype
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
		Vars:  map[string]string{"$bid": strconv.FormatUint(blockId, 10)},
		Mutations: []*api.Mutation{
			{
				Cond:      "@if(gt(len(dest), 0))",
				SetNquads: []byte("uid(dest) <privacytype> \"" + constants.StrPrivacyDestination + "\" ."),
			},
			{
				// only insert origins if there are not transactions to classify
				Cond:      "@if(gt(len(orig), 0) and eq(len(to_classify),0))",
				SetNquads: []byte("uid(orig) <privacytype> \"" + constants.StrPrivacyOrigin + "\" ."),
			}},
		CommitNow: true,
	}
	resp, err := db.TxWithRetryAndResponse(c, time.Minute*10, req)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	// json struct
	var r struct {
		Collaterals []dbtx.Transaction `json:"q,omitempty"`
		Origins     []dbtx.Transaction `json:"o,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	toClassify = r.Collaterals
	origins = r.Origins

	return
}

// SetCollateralCreation sets the collateral creation privacy type, if its input transaction are
// either of the type origin, mixing or collateral creation. Returns the number of newly
// classified transactions.
func SetCollateralCreation(c *dgo.Dgraph, txUids []string) (insertCount uint64, err error) {
	uidList := db.CreateUidList(txUids)

	// @filter(le(privacytype, 299))
	const filter = "@filter(le(privacytype," + constants.StrPrivacyOriginLast + "))"

	const query = `query Q($uids: string) {
				cc as var(func: uid($uids))@filter(not has(privacytype) or (ge(privacytype,` +
		constants.StrPrivacyDestinationFirst + `) and le(privacytype,` + constants.StrPrivacyDestinationLast + `)))@cascade{	
					tx_inputs{
						~tx_outputs` + filter + `}
				}
				q(func: uid(cc)){count(uid)}
			  }`

	const nQuad = "uid(cc) <privacytype> \"" + constants.StrPrivacyCollateralCreation + "\" ."

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
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	// json struct
	var r struct {
		Query []struct {
			Count uint64 `json:"count,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if len(r.Query) != 1 {
		err = errors.New("wrong number of returned query counts")
		return
	}

	insertCount = r.Query[0].Count

	return
}

// SetCollateralPayment sets the collateral payment privacy type, if its input transaction are
// either of the type origin, collateral creation or collateral payment. Returns the number
// of newly classified transactions.
func SetCollateralPayment(c *dgo.Dgraph, txUids []string) (insertCount uint64, err error) {
	uidList := db.CreateUidList(txUids)

	// collateral payments + collateral creations + origins
	const filter = "@filter(ge(privacytype," + constants.StrPrivacyOriginFirst +
		") and le(privacytype," + constants.StrPrivacyCollateralPaymentLast + "))"

	query := `query Q($uids: string) {
				cp as var(func: uid($uids))@filter(not has(privacytype) or (ge(privacytype,` + constants.StrPrivacyDestinationFirst + `) 
													and le(privacytype,` + constants.StrPrivacyDestinationLast + `)))@cascade{	
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
			SetNquads: []byte("uid(cp) <privacytype> \"" + constants.StrPrivacyCollateralPayment + "\" ."),
		}},
		CommitNow: true,
	}

	resp, err := db.TxWithRetryAndResponse(c, time.Minute*10, req)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	// json struct
	var r struct {
		Query []struct {
			Count uint64 `json:"count,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if len(r.Query) != 1 {
		err = errors.New("wrong number of returned query counts")
		return
	}

	insertCount = r.Query[0].Count

	return
}
