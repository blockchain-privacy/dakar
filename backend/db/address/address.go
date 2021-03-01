package address

import (
	"backend/cmd/cliutil"
	"backend/db"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"strconv"
	"time"
)

// GetFrontendAddress returns address information for the frontend sorted as specified by sortOrder.
// Use one of the constants like SortAscendingByInputTime to set the sortOrder
func GetFrontendAddress(c *dgo.Dgraph, addrHash string, sortOrder int, offset int, filters []int) (addr FrontendAddress,
	err error) {
	const maxOutputsPerQuery = 20
	sortDirection := "asc"
	sortBy := "val(ots)"

	switch sortOrder {
	case SortAscendingByInputTime:
		sortBy = "val(its)"
		break
	case SortDescendingByInputTime:
		sortDirection = "desc"
		sortBy = "val(its)"
		break
	case SortAscendingByOutputTime:
		// do nothing, values are already correctly set
		break
	case SortDescendingByOutputTime:
		sortDirection = "desc"
		break
	case SortAscendingByAmount:
		sortBy = "amount"
		break
	case SortDescendingByAmount:
		sortDirection = "desc"
		sortBy = "amount"
		break
	default:
		err = errors.New("error unrecognized sort order")
		return
	}
	var filter string
	for i, f := range filters {
		switch f {
		case FilterByCoinbase:
			filter += "eq(iscoinbase, true)"
			break
		case FilterByUnspent:
			filter += " NOT has(~tx_inputs)"
			break
		default:
			err = errors.New("error unrecognized filter")
			return
		}

		if i+1 < len(filters) {
			filter += " AND "
		}
	}

	if len(filters) > 0 {
		filter = fmt.Sprintf("@filter(%s)", filter)
	}

	// fill variables
	query := `query Q($hash: string){
		var(func: eq(addresshash, $hash)){
			addr_outputs{
				a as uid
				oamt as amount
				~tx_outputs{
					~transactions{
						obts as ts
					}
					otts as min(val(obts))
				}
				ots as min(val(otts))
				~tx_inputs{
					~transactions{
						ibts as ts
					}
					itts as min(val(ibts))
				}
				its as min(val(itts))
			}
		}
		var(func: uid(a))@filter(has(~tx_inputs)){
    		iamt as amount
  		}
		coinbase(func: uid(a))@filter(eq(iscoinbase, true)){
			count(uid)
		}
		c(func:uid(a), orderdesc: ` + sortBy + ")" + filter + `{
			count(uid)
        }
		ci(func: uid(iamt)){
			count(uid)
		}
		co(func: uid(oamt)){
			count(uid)
		}
		input_sum(){
			sum:sum(val(iamt))
		}
		output_sum(){
			sum:sum(val(oamt))
		}
		q(func: uid(a), order` + sortDirection + ":" + sortBy + ", first:" +
		strconv.Itoa(maxOutputsPerQuery) + ",offset:" + strconv.Itoa(offset) + ")" + filter + `@normalize{
			amount:amount
			is_coinbase:iscoinbase
			output_ts:val(ots)
			input_ts:val(its)
			input_index:inputindex
			output_index:outputindex
			~tx_outputs{
				output_transaction: txhash
			}
			~tx_inputs{
				input_transaction: txhash
			}
		}
	}`

	vars := make(map[string]string)
	vars["$hash"] = addrHash
	ctx, cancel := db.GetFrontendContext()
	defer cancel()
	resp, err := c.NewReadOnlyTxn().QueryWithVars(ctx, query, vars)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r struct {
		Outputs       []FrontendOutput `json:"q"`
		QueryMaxCount []struct {
			Count int64 `json:"count"`
		} `json:"c"`
		CoinbaseCount []struct {
			Count int64 `json:"count"`
		} `json:"coinbase"`
		InputCount []struct {
			Count int64 `json:"count"`
		} `json:"ci"`
		OutputCount []struct {
			Count int64 `json:"count"`
		} `json:"co"`
		InputSum []struct {
			// if the input sum is 0 it may be returned as a float, e.g. "0.00000".
			// Because of this we have to first save it as a string and after that convert it to an int64.
			// Can be reversed if https://github.com/dgraph-io/dgraph/pull/7176 is merged
			Sum json.Number `json:"sum"`
		} `json:"input_sum"`
		OutputSum []struct {
			Sum json.Number `json:"sum"`
		} `json:"output_sum"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return addr, err
	}

	if len(r.QueryMaxCount) != 1 || len(r.CoinbaseCount) != 1 ||
		len(r.InputSum) != 1 || len(r.OutputSum) != 1 ||
		len(r.InputCount) != 1 || len(r.OutputCount) != 1 {
		err = ErrorInvalidResult
		return
	}

	// not checking the length of r.Outputs, as for certain filters the number of outputs can be 0
	// instead check for the calculated output count
	if r.OutputCount[0].Count == 0 {
		err = ErrorAddressNotFound
		return
	}

	// try to convert input sum to int64
	inputSum, conversionErr := convertJsonNumber(r.InputSum[0].Sum)
	if conversionErr != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), conversionErr)
		return
	}

	// try to convert output sum to int64
	outputSum, conversionErr := convertJsonNumber(r.OutputSum[0].Sum)
	if conversionErr != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), conversionErr)
		return
	}

	addr = FrontendAddress{
		Hash:          addrHash,
		QueryMaxCount: r.QueryMaxCount[0].Count,
		InputCount:    r.InputCount[0].Count,
		OutputCount:   r.OutputCount[0].Count,
		InputSum:      inputSum,
		OutputSum:     outputSum,
		Outputs:       r.Outputs,
		CoinbaseCount: r.CoinbaseCount[0].Count,
	}

	return
}

// convertJsonNumber tries to convert num to an int64
func convertJsonNumber(num json.Number) (number int64, err error) {
	number, intErr := num.Int64()
	if intErr != nil {
		// also accept float 0.000
		floatInputSum, floatErr := num.Float64()
		if floatErr != nil || floatInputSum != 0 {
			err = errors.New("could not convert json.Number to int64")
			return
		}

		number = int64(floatInputSum)
	}

	return
}

// GetInputAddressesOfTransaction gets all input addresses of the transaction specified by uid
func GetInputAddressesOfTransaction(c *dgo.Dgraph, uid string) (addresses []Address, err error) {
	query := `query Q($uid: string){
				q(func: uid($uid)){
					inputs: tx_inputs @normalize{
						~addr_outputs{
							addresshash: addresshash
						}
					}
			  	}
			   }`
	ctx, cancel := db.GetBackendContext()
	defer cancel()
	resp, err := db.ReadOnlyTxVarWithRetry(c, ctx, query, map[string]string{"$uid": uid})

	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	// json struct
	var r struct {
		Transaction []struct {
			Inputs []struct {
				AddressHash string `json:"addresshash"`
			} `json:"inputs,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if len(r.Transaction) == 0 {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), ErrorAddressNotFound)
		return
	}

	if len(r.Transaction) > 1 {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), ErrorInvalidResult)
		return
	}

	for _, e := range r.Transaction[0].Inputs {
		addresses = append(addresses, Address{Hash: e.AddressHash})
	}

	return
}

// upserts addresses
func UpsertAddresses(c *dgo.Dgraph, addresses []Address) error {
	if addresses == nil {
		return errors.New("got null pointer for addresses")
	}

	// the following block creates the query for 4 addresses the query looks like this:
	//		query Q($h0: string,$h1: string,$h2: string,$h3: string) {
	//		a0 as var(func: eq(addresshash, $h0))
	//		a1 as var(func: eq(addresshash, $h1))
	//		a2 as var(func: eq(addresshash, $h2))
	//		a3 as var(func: eq(addresshash, $h3))
	//		}
	// $h0 ... $h4 are needed to be later replaced. This prevents string injection

	vars := make(map[string]string)
	queryPrefix := "query Q("
	var query string
	// set uid for all addresses and build query
	for i := range addresses {
		queryPrefix += "$h" + strconv.Itoa(i) + ": string"

		if i+1 < len(addresses) {
			queryPrefix += ","
		}

		addresses[i].Uid = fmt.Sprintf("uid(a%d)", i)
		addresses[i].SetDType()
		query += fmt.Sprintf("a%d as var(func: eq(addresshash, $h%d))\n", i, i)
		vars["$h"+strconv.Itoa(i)] = addresses[i].Hash
	}

	queryPrefix += ") {\n"

	pb, err := json.Marshal(addresses)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return err
	}

	req := &api.Request{
		Query: queryPrefix + query + "}",
		Vars:  vars,
		Mutations: []*api.Mutation{{
			SetJson: pb,
		}},
		CommitNow: true,
	}
	return db.TxWithRetry(c, time.Minute*3, req)
}

// gets the number of addresses in the database
func GetCount(c *dgo.Dgraph) (uint64, error) {
	return db.GetCount(c, DType)
}
