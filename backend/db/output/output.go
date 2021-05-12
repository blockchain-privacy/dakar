package output

import (
	"backend/cmd/cliutil"
	"backend/db"
	"time"

	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/dgraph-io/dgo/v2"
)

const NumDenominations = 5

const (
	// MinCollateral is 1/10 of the smallest denomination: round(100001/10).
	MinCollateral = 10000

	// OldMinCollateral is the minimum collateral before the 5th denomination was added in protocol version 70213 it was round(1000010/10): 100000
	// OldMinCollateral = 100000
	MaxCollateral    = 40000  // 4*MinCollateral
	OldMaxCollateral = 400000 // 4*OldMinCollateral
)

var DenominationsTypes = [NumDenominations]int64{1000010000, 100001000, 10000100, 1000010, 100001}

// GetOutput gets output information from the database
func GetOutput(c *dgo.Dgraph, txHash string, index uint32, isInput bool) (op Output, err error) {
	// build query
	relationship := "tx_outputs"
	indextype := "outputindex"
	if isInput {
		relationship = "tx_inputs"
		indextype = "inputindex"
	}

	query := fmt.Sprintf(`query Q($hash: string, $idx: string) {
				getOutput(func: eq(txhash, $hash)) {
					%s @filter(eq(%s, $idx)) {
						uid
						outputindex
						inputindex
						amount
						txtype
						iscoinbase
						dgraph.type
					}
				}
			  }
				`, relationship, indextype)

	vars := make(map[string]string)
	vars["$hash"] = txHash
	vars["$idx"] = strconv.FormatUint(uint64(index), 10)

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*10, query, vars)

	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}
	var r outputQuery

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	return r.payload()
}

// GetCount gets the number of outputs in the database
func GetCount(c *dgo.Dgraph) (uint64, error) {
	return db.GetCount(c, DType)
}

func CountOutputDenominations(outputs []Output) [NumDenominations]int {

	var amounts []int64

	for _, o := range outputs {
		if o.Amount == nil {
			log.Println("error amount not set")
			return [NumDenominations]int{}
		}
		amounts = append(amounts, *o.Amount)
	}

	return CountAmountDenominations(amounts)
}

func CountAmountDenominations(amounts []int64) (denominations [NumDenominations]int) {
	for _, amt := range amounts {
	inner:
		for i, v := range DenominationsTypes {
			if amt == v {
				denominations[i]++
				break inner
			}
		}
	}

	return
}
