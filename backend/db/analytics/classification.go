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
func DoClassification(c *dgo.Dgraph, blockId uint64) (toClassify []dbtx.Transaction, err error) {
	query := `query Q($bid: string) {
				b as var(func: eq(id,$bid)){t as ts}
				var(func: uid(b))@cascade{
					dest as transactions@filter(not has(privacytype)){
						tx_inputs{
							~tx_outputs@filter(eq(privacytype,"mixing"))
						}
					}
				}
				var(func: uid(b)){
					transactions@filter(eq(privacytype,"mixing")){
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
			  }`

	req := &api.Request{
		Query: query,
		Vars:  map[string]string{"$bid": strconv.FormatUint(blockId, 10)},
		Mutations: []*api.Mutation{{
			Cond:      "@if(gt(len(dest), 0))",
			SetNquads: []byte("uid(dest) <privacytype> \"" + constants.PrivacyDestination + "\" ."),
		},
			{
				Cond:      "@if(gt(len(orig), 0))",
				SetNquads: []byte("uid(orig) <privacytype> \"" + constants.PrivacyOrigin + "\" ."),
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
		Transaction []dbtx.Transaction `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	toClassify = r.Transaction

	return
}

// SetCollateralCreation sets the collateral creation privacy type, if its input transaction are
// either of the type origin, mixing or collateral creation. Returns the number of newly
// classified transactions.
func SetCollateralCreation(c *dgo.Dgraph, txUids []string) (insertCount uint64, err error) {
	uidList := db.CreateUidList(txUids)

	// @filter(eq(privacytype, ["mixing", "origin", "cc"]))
	const filter = "@filter(eq(privacytype,[" + constants.PrivacyCollateralCreation + "," +
		constants.PrivacyMixing + "," + constants.PrivacyOrigin + "]))"

	query := `query Q($uids: string) {
				cc as var(func: uid($uids))@filter(not has(privacytype) or eq(privacytype,"destination"))@cascade{	
					tx_inputs{
						~tx_outputs` + filter + `}
				}
				q(func: uid(cc)){count(uid)}
			  }`

	req := &api.Request{
		Query: query,
		Vars:  map[string]string{"$uids": uidList},
		Mutations: []*api.Mutation{{
			Cond:      "@if(gt(len(cc), 0))",
			SetNquads: []byte("uid(cc) <privacytype> \"" + constants.PrivacyCollateralCreation + "\" ."),
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

	// @filter(eq(privacytype, ["origin", "cc"]))
	const filter = "@filter(eq(privacytype,[" + constants.PrivacyCollateralCreation + "," +
		constants.PrivacyCollateralPayment + "," + constants.PrivacyOrigin + "]))"

	query := `query Q($uids: string) {
				cp as var(func: uid($uids))@filter(not has(privacytype) or eq(privacytype,"destination"))@cascade{	
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
			SetNquads: []byte("uid(cp) <privacytype> \"" + constants.PrivacyCollateralPayment + "\" ."),
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
