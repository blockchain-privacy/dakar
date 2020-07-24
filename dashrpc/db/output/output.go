package output

import (
	"dashrpc/db"
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/dgo/v2"
	"strconv"
)

// gets output information from the database
func GetOutput(c *dgo.Dgraph, txHash string, index uint32, isInput bool) (op Output, err error) {
	// build query
	relationship := "tx_outputs"
	if isInput {
		relationship = "tx_inputs"
	}

	query := fmt.Sprintf(`query Q($hash: string, $idx: string) {
				getOutput(func: eq(txhash, $hash)) {
					%s @filter(eq(index, $idx)) {
						uid
						index
						amount
						txtype
						iscoinbase
						dgraph.type
					}
				}
			  }
				`, relationship)

	vars := make(map[string]string)
	vars["$hash"] = txHash
	vars["$idx"] = strconv.FormatUint(uint64(index), 10)

	resp, err := c.NewReadOnlyTxn().QueryWithVars(db.GetContext(), query, vars)
	if err != nil {
		return op, err
	}
	var r outputQuery

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		return op, err
	}

	return r.payload()
}
