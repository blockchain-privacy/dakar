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

// gets address information from the database
func GetAddress(c *dgo.Dgraph, addrHash string) (addr Address, err error) {
	query := `query Q($hash: string) {
				q(func: eq(addresshash, $hash)){
					uid
					addresshash
					addr_outputs{
						uid
						amount
						index
						iscoinbase
						txtype
					}
				}
			  }
				`
	vars := make(map[string]string)
	vars["$hash"] = addrHash
	resp, err := c.NewReadOnlyTxn().QueryWithVars(db.GetContext(), query, vars)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}
	var r addressQuery

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	return r.payload()
}

// gets address information for the frontend
func GetFrontendAddress(c *dgo.Dgraph, addrHash string) (addr FrontendAddress, err error) {
	// todo remove first: 200 limit
	query := `query Q($hash: string) {
				q(func: eq(addresshash, $hash)){
					addresshash
					addr_outputs(first: 200)@normalize{
						amount:amount
						index:index
						iscoinbase:iscoinbase
						~tx_outputs{
							output_transaction: txhash
						}
						~tx_inputs{
							input_transaction: txhash
						}
					}
				}
			  }`

	vars := make(map[string]string)
	vars["$hash"] = addrHash
	resp, err := c.NewReadOnlyTxn().QueryWithVars(db.GetContext(), query, vars)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r struct {
		Address []FrontendAddress `json:"q"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return addr, err
	}

	if len(r.Address) == 0 || len(r.Address[0].Outputs) == 0 {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), ErrorAddressNotFound)
		return
	} else if len(r.Address) != 1 {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), ErrorInvalidResult)
		return
	}

	addr = r.Address[0]

	return
}

// gets all input addresses of the transaction specified by uid
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

	resp, err := c.NewReadOnlyTxn().QueryWithVars(db.GetContext(), query, map[string]string{"$uid": uid})
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

// gets address information from the database and checks if it is complete
func GetCompleteAddress(c *dgo.Dgraph, addressHash string) (addr Address, err error) {
	addr, err = GetAddress(c, addressHash)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if !addr.isComplete() {
		err = errors.New("address is not complete")
		return
	}

	return
}

// upserts an address
func UpsertAddress(c *dgo.Dgraph, address Address) (*api.Response, error) {
	address.Uid = "uid(v)"
	address.SetDType()
	pb, err := json.Marshal(address)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return nil, err
	}

	query := `
		query Q($hash: string) {
			q(func: eq(addresshash, $hash)) {
				v as uid
			}
		}
	`

	vars := make(map[string]string)
	vars["$hash"] = address.Hash

	req := &api.Request{
		Query: query,
		Vars:  vars,
		Mutations: []*api.Mutation{{
			SetJson: pb,
		}},
		CommitNow: true,
	}

	return c.NewTxn().Do(db.GetContext(), req)
}

// upserts addresses
func UpsertAddresses(c *dgo.Dgraph, addresses []Address) (*api.Response, error) {
	if addresses == nil {
		return nil, errors.New("got null pointer for addresses")
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
		return nil, err
	}

	req := &api.Request{
		Query: queryPrefix + query + "}",
		Vars:  vars,
		Mutations: []*api.Mutation{{
			SetJson: pb,
		}},
		CommitNow: true,
	}

	return c.NewTxn().Do(db.GetContext(), req)
}

// gets the number of addresses in the database
func GetCount(c *dgo.Dgraph) (uint64, error) {
	return db.GetCount(c, DType)
}
