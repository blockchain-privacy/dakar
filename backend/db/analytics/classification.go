package analytics

import (
	"backend/cmd/cliutil"
	"backend/constants"
	"backend/db"
	"fmt"
	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"strconv"
	"time"
)

// DoClassification sets the privacy type for destination transactions in the given block and
// the origin privacy type for all transactions which are connected to mixing transactions in this block
func DoClassification(c *dgo.Dgraph, blockId uint64) (err error) {
	query := `query Q($bid: string) {
				b as var(func: eq(id,$bid))
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

	if err = db.TxWithRetry(c, time.Minute*10, req); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	return
}

// SetCollateralCreation sets the collateral creation privacy type, if its input transaction are
// either of the type origin, mixing or collateral creation
func SetCollateralCreation(c *dgo.Dgraph, txUids []string) (err error) {
	// build uid list in this form: [uid1,uid2]
	uidList := "["
	for i, uid := range txUids {
		uidList += uid
		if i+1 < len(txUids) {
			uidList += ","
		}
	}
	uidList += "]"

	// @filter(eq(privacytype, ["mixing", "origin", "cc"]))
	const filter = "@filter(eq(privacytype,[" + constants.PrivacyCollateralCreation + "," +
		constants.PrivacyMixing + "," + constants.PrivacyOrigin + "]))"

	query := `query Q($uids: string) {
				cc as var(func: uid($uids))@cascade{	
					tx_inputs{
						~tx_outputs` + filter + `}}}`

	req := &api.Request{
		Query: query,
		Vars:  map[string]string{"$uids": uidList},
		Mutations: []*api.Mutation{{
			Cond:      "@if(gt(len(cc), 0))",
			SetNquads: []byte("uid(cc) <privacytype> \"" + constants.PrivacyCollateralCreation + "\" ."),
		}},
		CommitNow: true,
	}

	if err = db.TxWithRetry(c, time.Minute*10, req); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	return
}

// SetCollateralPayment sets the collateral payment privacy type, if its input transaction are
// either of the type origin, collateral creation or collateral payment
func SetCollateralPayment(c *dgo.Dgraph, txUids []string) (err error) {
	// build uid list in this form: [uid1,uid2]
	uidList := "["
	for i, uid := range txUids {
		uidList += uid
		if i+1 < len(txUids) {
			uidList += ","
		}
	}
	uidList += "]"

	// @filter(eq(privacytype, ["origin", "cc"]))
	const filter = "@filter(eq(privacytype,[" + constants.PrivacyCollateralCreation + "," +
		constants.PrivacyCollateralPayment + "," + constants.PrivacyOrigin + "]))"

	query := `query Q($uids: string) {
				cp as var(func: uid($uids))@cascade{	
					tx_inputs{
						~tx_outputs` + filter + `}}}`

	req := &api.Request{
		Query: query,
		Vars:  map[string]string{"$uids": uidList},
		Mutations: []*api.Mutation{{
			Cond:      "@if(gt(len(cp), 0))",
			SetNquads: []byte("uid(cp) <privacytype> \"" + constants.PrivacyCollateralPayment + "\" ."),
		}},
		CommitNow: true,
	}

	if err = db.TxWithRetry(c, time.Minute*10, req); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	return
}
