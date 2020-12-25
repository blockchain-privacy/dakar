package address

import (
	"dashrpc/cmd/cliutil"
	"dashrpc/db"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"strconv"
)

// GetFrontendAddress returns address information for the frontend sorted as specified by sortOrder.
// Use one of the constants like SortAscendingByInputTime to set the sortOrder
func GetFrontendAddress(c *dgo.Dgraph, addrHash string, sortOrder int, offset int) (addr FrontendAddress,
	err error) {
	const maxOutputsPerQuery = 20
	sortDirection := "asc"
	sortBy := "ots"

	switch sortOrder {
	case SortAscendingByInputTime:
		sortBy = "its"
		break
	case SortDescendingByInputTime:
		sortDirection = "desc"
		sortBy = "its"
		break
	case SortAscendingByOutputTime:
		// do nothing, values are already correctly set
		break
	case SortDescendingByOutputTime:
		sortDirection = "desc"
		break
	default:
		err = errors.New("error unrecognized sort order")
		return
	}

	// fill variables
	query := `query Q($hash: string){
		var(func: eq(addresshash, $hash)){
			addr_outputs{
				a as uid
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
		c(func:uid(a), orderdesc: val(` + sortBy + `)){
          count(uid)
        }
		q(func: uid(a), order` + sortDirection + ": val(" + sortBy + "), first:" +
		strconv.Itoa(maxOutputsPerQuery) + ",offset:" + strconv.Itoa(offset) + `)@normalize{
			amount:amount
			iscoinbase:iscoinbase
			output_ts:val(ots)
			input_ts:val(its)
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
	resp, err := c.NewReadOnlyTxn().QueryWithVars(db.GetFrontendContext(), query, vars)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r struct {
		Outputs []FrontendOutput `json:"q"`
		Counts  []struct {
			Count int64 `json:"count"`
		} `json:"c"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return addr, err
	}

	if len(r.Counts) != 1 {
		err = ErrorInvalidResult
		return
	}

	if len(r.Outputs) == 0 {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), ErrorAddressNotFound)
		return
	}

	addr = FrontendAddress{
		Hash:       addrHash,
		NumOutputs: r.Counts[0].Count,
		Outputs:    r.Outputs,
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

	resp, err := db.ReadOnlyTxVarWithRetry(c, db.GetBackendContext(),
		query, map[string]string{"$uid": uid})

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

	return db.TxWithRetry(c, db.GetBackendContext(), req)
}

// gets the number of addresses in the database
func GetCount(c *dgo.Dgraph) (uint64, error) {
	return db.GetCount(c, DType)
}
