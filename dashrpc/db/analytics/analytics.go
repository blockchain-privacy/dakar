package analytics

import (
	"dashrpc/cmd/cliutil"
	"dashrpc/db"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgraph-io/dgo/v2"
)

// Searches for all potential origins up to depth. The returned string slice contains the uids of
// the found transactions
func AnalyzeOrigins(c *dgo.Dgraph, transactionHash string, depth uint) (origins []string, err error) {
	if depth == 0 || depth > 30 {
		err = fmt.Errorf("invalid depth")
		return
	}

	queryStart := `query Q($hash: string){
		var(func: eq(txhash, $hash)){
			tx_inputs{
			  inputs1 as uid
			}
		}
`

	var queryMiddle string
	var txUids string

	for i := uint(0); i < depth; i++ {
		txUids += fmt.Sprintf("tx%d", i+1)
		if i+1 < depth {
			txUids += ","
			queryMiddle += fmt.Sprintf(`
		var(func: uid(inputs%d)){
			~tx_outputs@filter(eq(privacytype, "mixing")){
				tx_inputs{
					inputs%d as uid
				}
			}
		}`, i+1, i+2)
		}

		queryMiddle += fmt.Sprintf(`
		var(func: uid(inputs%d)){
			tx%d as ~tx_outputs@filter(eq(privacytype, "origin")){}
		}`, i+1, i+1)
	}

	queryEnd := fmt.Sprintf(`
		filtertx(func: uid(%s)){
			uid
		}
}`, txUids)

	query := queryStart + queryMiddle + queryEnd

	resp, err := db.ReadOnlyTxVarWithRetry(c, db.GetBackendContext(), query, map[string]string{"$hash": transactionHash})

	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r struct {
		Transaction []struct {
			Uid string `json:"uid,omitempty"`
		} `json:"filtertx,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	for _, uid := range r.Transaction {
		origins = append(origins, uid.Uid)
	}

	return
}

// gets transaction information from the database including uids of origins
func GetOrigins(c *dgo.Dgraph, txHash string) (origins []Origin, err error) {
	query := `query Q($hash: string) {
				q(func: eq(txhash, $hash)){
					origins{
						txhash
						privacytype
						fee
						block: ~transactions {
							blockhash
							ts
							id
						}
					}
				}
			  }
				`

	resp, err := c.NewReadOnlyTxn().QueryWithVars(db.GetFrontendContext(), query, map[string]string{"$hash": txHash})

	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r struct {
		Q []struct {
			Origins []struct {
				Hash  string `json:"txhash,omitempty"`
				Block []struct {
					Hash string `json:"blockhash,omitempty"`
					Ts   string `json:"ts,omitempty"`
					Id   uint64 `json:"id,omitempty"`
				} `json:"block,omitempty"`
			} `json:"origins,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if len(r.Q) != 1 {
		err = errors.New("invalid response from database")
		return
	}

	for _, o := range r.Q[0].Origins {
		if len(o.Block) == 0 {
			err = errors.New("invalid response from database")
			return
		}

		origins = append(origins, Origin{
			Hash:           o.Hash,
			BlockHash:      o.Block[0].Hash,
			BlockId:        o.Block[0].Id,
			BlockTimestamp: o.Block[0].Ts,
		})
	}

	return
}
