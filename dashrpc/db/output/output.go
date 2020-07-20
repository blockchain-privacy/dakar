package output

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgraph-io/dgo/v2"
	"strconv"
)

// gets output information from the database
func GetOutput(c *dgo.Dgraph, txHash string, index uint32, output *Output, isInput bool) error {
	tx := c.NewReadOnlyTxn()
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
	resp, err := tx.QueryWithVars(context.Background(), query, vars)
	if err != nil {
		return err
	}
	var r outputQuery
	err = json.Unmarshal(resp.Json, &r)

	if err != nil {
		return err
	}

	lenQ := len(r.GetOutput)
	if lenQ == 0 {
		return errors.New("no output found")
	}

	lenTx := len(r.GetOutput[0].Transaction)
	if lenTx == 0 {
		return errors.New("no output found")
	}

	lenO := len(r.GetOutput[0].Transaction[0].Outputs)
	if lenO == 0 {
		return errors.New("no output found")
	}

	*output = r.GetOutput[0].Transaction[0].Outputs[0]
	if lenQ > 1 || lenTx > 1 || lenO > 1 {
		// found more than one output, which should not be possible
		return errors.New("found more than one output")
	}

	return nil
}
