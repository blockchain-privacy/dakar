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

var DenominationsTypes = [NumDenominations]int64{1000010000, 100001000, 10000100, 1000010, 100001}

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

	resp, err := db.ReadOnlyTxVarWithRetryAndTimeout(c, time.Second*20, query, vars)

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
	ctx, cancel := db.GetBackendContext()
	defer cancel()
	resp, err := db.ReadOnlyTxVarWithRetry(c, ctx, query, vars)

	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	// struct for json parsing
	type verboseOutputQuery struct {
		GetOutput []struct {
			Outputs []struct {
				Uid                string   `json:"uid,omitempty"`
				OutputIndex        *uint32  `json:"outputindex,omitempty"`
				InputIndex         *uint32  `json:"inputindex,omitempty"`
				TxType             string   `json:"txtype,omitempty"`
				Amount             *int64   `json:"amount,omitempty"`
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
		err = ErrorNotFound
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
			  }`

	ctx, cancel := db.GetBackendContext()
	defer cancel()
	resp, err := db.ReadOnlyTxVarWithRetry(c, ctx, query, map[string]string{"$id": uid})

	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	// struct for json parsing
	type queryOutput struct {
		GetOutput []struct {
			Uid                string   `json:"uid,omitempty"`
			OutputIndex        *uint32  `json:"outputindex,omitempty"`
			InputIndex         *uint32  `json:"inputindex,omitempty"`
			TxType             string   `json:"txtype,omitempty"`
			Amount             *int64   `json:"amount,omitempty"`
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
		err = ErrorNotFound
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
