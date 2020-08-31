package analytics

import (
	"dashrpc/cmd/cliutil"
	"dashrpc/db"
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/dgo/v2"
)

// gets block information from the database
func GetOrigins(c *dgo.Dgraph, transactionHash string, depth uint) (origins []string, err error) {
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

	resp, err := c.NewReadOnlyTxn().QueryWithVars(db.GetBackendContext(),
		query, map[string]string{"$hash": transactionHash})

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
