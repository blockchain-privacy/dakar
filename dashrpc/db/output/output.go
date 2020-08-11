package output

import (
	"dashrpc/cmd/cliutil"
	"dashrpc/db"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgraph-io/dgo/v2"
	"log"
	"math"
	"strconv"
)

// gets output information from the database
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

	resp, err := c.NewReadOnlyTxn().QueryWithVars(db.GetContext(), query, vars)
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

// same as 'GetOutput', but also includes the connected transactions
func GetVerboseOutput(c *dgo.Dgraph, txHash string, index uint32, isInput bool) (op VerboseOutput, err error) {
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
						~tx_inputs{
							txhash
						}
						~tx_outputs{
							txhash
						}
						~addr_outputs{
							addresshash
						}
					}
				}
			  }
				`, relationship, indextype)

	vars := make(map[string]string)
	vars["$hash"] = txHash
	vars["$idx"] = strconv.FormatUint(uint64(index), 10)

	resp, err := c.NewReadOnlyTxn().QueryWithVars(db.GetContext(), query, vars)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	// struct for json parsing
	type verboseOutputQuery struct {
		GetOutput []struct {
			Outputs []struct {
				Uid                string   `json:"uid,omitempty"`
				OutputIndex        *uint64  `json:"outputindex,omitempty"`
				InputIndex         *uint64  `json:"inputindex,omitempty"`
				TxType             string   `json:"txtype,omitempty"`
				Amount             string   `json:"amount,omitempty"`
				IsCoinbase         *bool    `json:"iscoinbase,omitempty"`
				DType              []string `json:"dgraph.type,omitempty"`
				OutputTransactions []struct {
					Hash string `json:"txhash,omitempty"`
				} `json:"~tx_outputs,omitempty"`
				InputTransactions []struct {
					Hash string `json:"txhash,omitempty"`
				} `json:"~tx_inputs,omitempty"`
				Addresses []struct {
					Hash string `json:"addresshash,omitempty"`
				} `json:"~addr_outputs,omitempty"`
			} `json:"tx_outputs"`
		} `json:"getOutput"`
	}

	var r verboseOutputQuery

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		return op, err
	}

	// check result of parsed json
	if len(r.GetOutput) != 1 ||
		len(r.GetOutput[0].Outputs) != 1 ||
		len(r.GetOutput[0].Outputs[0].Addresses) == 0 ||
		(len(r.GetOutput[0].Outputs[0].OutputTransactions) != 1 &&
			len(r.GetOutput[0].Outputs[0].InputTransactions) != 1) {
		// todo compare errors with error.Is or error.As
		err = errors.New(ErrorNotFound)
		return
	}

	// convert to return type
	out := r.GetOutput[0].Outputs[0]
	op = VerboseOutput{
		Uid:         out.Uid,
		OutputIndex: out.OutputIndex,
		InputIndex:  out.InputIndex,
		TxType:      out.TxType,
		Amount:      out.Amount,
		IsCoinbase:  out.IsCoinbase,
	}

	if out.OutputTransactions != nil {
		op.OutputTransaction = out.OutputTransactions[0].Hash
	}

	if out.InputTransactions != nil {
		op.InputTransaction = out.InputTransactions[0].Hash
	}

	for _, v := range out.Addresses {
		op.Addresses = append(op.Addresses, v.Hash)
	}

	return
}

// same as 'GetOutputVerbose', but search by uid
func GetVerboseOutputByUid(c *dgo.Dgraph, uid string) (op VerboseOutput, err error) {
	// build query
	query := `query Q($id: string) {
				q(func: uid($id)) {
						uid
						outputindex
						inputindex
						amount
						txtype
						iscoinbase
						dgraph.type
						~tx_inputs{
							txhash
						}
						~tx_outputs{
							txhash
						}
						~addr_outputs{
							addresshash
						}
				}
			  }
				`

	resp, err := c.NewReadOnlyTxn().QueryWithVars(db.GetContext(), query, map[string]string{"$id": uid})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	// struct for json parsing
	type queryOutput struct {
		GetOutput []struct {
			Uid                string   `json:"uid,omitempty"`
			OutputIndex        *uint64  `json:"outputindex,omitempty"`
			InputIndex         *uint64  `json:"inputindex,omitempty"`
			TxType             string   `json:"txtype,omitempty"`
			Amount             string   `json:"amount,omitempty"`
			IsCoinbase         *bool    `json:"iscoinbase,omitempty"`
			DType              []string `json:"dgraph.type,omitempty"`
			OutputTransactions []struct {
				Hash string `json:"txhash,omitempty"`
			} `json:"~tx_outputs,omitempty"`
			InputTransactions []struct {
				Hash string `json:"txhash,omitempty"`
			} `json:"~tx_inputs,omitempty"`
			Addresses []struct {
				Hash string `json:"addresshash,omitempty"`
			} `json:"~addr_outputs,omitempty"`
		} `json:"q"`
	}

	var r queryOutput
	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	// check result of parsed json
	if len(r.GetOutput) != 1 ||
		len(r.GetOutput[0].Addresses) == 0 ||
		(len(r.GetOutput[0].OutputTransactions) != 1 &&
			len(r.GetOutput[0].InputTransactions) != 1) {
		// todo compare errors with error.Is or error.As
		err = errors.New(ErrorNotFound)
		return
	}

	// convert to return type
	out := r.GetOutput[0]
	op = VerboseOutput{
		Uid:         out.Uid,
		OutputIndex: out.OutputIndex,
		InputIndex:  out.InputIndex,
		TxType:      out.TxType,
		Amount:      out.Amount,
		IsCoinbase:  out.IsCoinbase,
	}

	if out.OutputTransactions != nil {
		op.OutputTransaction = out.OutputTransactions[0].Hash
	}

	if out.InputTransactions != nil {
		op.InputTransaction = out.InputTransactions[0].Hash
	}

	for _, v := range out.Addresses {
		op.Addresses = append(op.Addresses, v.Hash)
	}

	return
}

// gets the number of outputs in the database
func GetCount(c *dgo.Dgraph) (uint64, error) {
	return db.GetCount(c, DType)
}

func almostEqual(a, b float64) bool {
	var delta float64
	delta = 0.00001
	return math.Abs(a-b) <= delta
}

func CountDenominations(outputs []Output) []int {
	denominationsTypes := []float64{1.00001, 0.100001, 0.0100001, 0.00100001}
	denominations := make([]int, len(denominationsTypes))

	for _, o := range outputs {
	inner:
		for i, v := range denominationsTypes {
			amt, err := strconv.ParseFloat(o.Amount, 64)
			if err != nil {
				log.Println("Error converting", o.Amount, "to string")
				return nil
			}
			if almostEqual(amt, v) {
				denominations[i]++
				break inner
			}
		}
	}

	return denominations
}
