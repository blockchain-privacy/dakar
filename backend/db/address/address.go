package address

import (
	"backend/cmd/cliutil"
	"backend/db"
	"backend/external"

	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/dgraph-io/dgo/v210/protos/api"
)

// GetFrontendAddress returns address information for the frontend sorted as specified by sortOrder.
// Use one of the constants like SortAscendingByInputTime to set the sortOrder
func GetFrontendAddress(c external.Database, addrHash string, sortOrder int,
	offset int, filters []int) (addr FrontendAddress, err error) {
	const maxOutputsPerQuery = 20
	sortDirection := "asc"
	sortBy := "val(ots)"

	switch sortOrder {
	case SortAscendingByInputTime:
		sortBy = "val(its)"
	case SortDescendingByInputTime:
		sortDirection = "desc"
		sortBy = "val(its)"
	case SortAscendingByOutputTime:
		// do nothing, values are already correctly set
	case SortDescendingByOutputTime:
		sortDirection = "desc"
	case SortAscendingByAmount:
		sortBy = "amount"
	case SortDescendingByAmount:
		sortDirection = "desc"
		sortBy = "amount"
	default:
		err = errors.New("error unrecognized sort order")
		return
	}
	var filter string
	for i, f := range filters {
		switch f {
		case FilterByCoinbase:
			filter += "eq(iscoinbase, true)"
		case FilterByUnspent:
			filter += " NOT has(~tx_inputs)"
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
	resp, err := c.Query(ctx, query, vars)
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
			Sum int64 `json:"sum"`
		} `json:"input_sum"`
		OutputSum []struct {
			Sum int64 `json:"sum"`
		} `json:"output_sum"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return addr, err
	}

	if len(r.QueryMaxCount) != 1 || len(r.CoinbaseCount) != 1 ||
		len(r.InputSum) != 1 || len(r.OutputSum) != 1 ||
		len(r.InputCount) != 1 || len(r.OutputCount) != 1 {
		err = ErrInvalidResult
		return
	}

	// not checking the length of r.Outputs, as for certain filters the number of outputs can be 0
	// instead check for the calculated output count
	if r.OutputCount[0].Count == 0 {
		err = ErrAddressNotFound
		return
	}

	addr = FrontendAddress{
		Hash:          addrHash,
		QueryMaxCount: r.QueryMaxCount[0].Count,
		InputCount:    r.InputCount[0].Count,
		OutputCount:   r.OutputCount[0].Count,
		InputSum:      r.InputSum[0].Sum,
		OutputSum:     r.OutputSum[0].Sum,
		Outputs:       r.Outputs,
		CoinbaseCount: r.CoinbaseCount[0].Count,
	}

	return
}

// UpsertAddresses upserts addresses
func UpsertAddresses(c external.Database, addresses []Address) error {
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

		addresses[i].UID = fmt.Sprintf("uid(a%d)", i)
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
	return db.TxWithRetry(c, time.Minute*15, req)
}

// GetAddressUIDs returns all requested address nodes
func GetAddressUIDs(c external.Database, addressHashes []string) (addresses []Address, err error) {
	query := `{
				q(func: eq(addresshash,` + db.CreateCommaArray(addressHashes) + `)){
					uid
					addresshash
				}
			  }`

	resp, err := db.ReadOnlyTxWithRetry(c, time.Minute*10, query)

	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}
	var r struct {
		Addresses []Address `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	addresses = r.Addresses

	return
}
