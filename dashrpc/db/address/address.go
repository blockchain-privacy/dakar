package address

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"strconv"
)

// gets address information from the database
func GetAddress(c *dgo.Dgraph, addrHash string) (addr Address, err error) {

	tx := c.NewReadOnlyTxn()
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
	resp, err := tx.QueryWithVars(context.Background(), query, vars)
	if err != nil {
		return addr, err
	}
	var r addressQuery
	err = json.Unmarshal(resp.Json, &r)

	if err != nil {
		return addr, err
	}

	lenQ := len(r.Q)

	if lenQ == 0 {
		err = errors.New("no addresses found")
		return addr, err
	}

	addr = r.Q[0]
	if lenQ > 1 {
		// found more than one address, which should not be possible
		err = errors.New("found more than one address")
		return addr, err
	}

	return addr, err
}

// gets address information from the database and checks if it is complete
func GetCompleteAddress(c *dgo.Dgraph, addressHash string) (addr Address, err error) {
	addr, err = GetAddress(c, addressHash)
	if err != nil {
		return addr, err
	}

	if !addr.isComplete() {
		err = errors.New("address not complete")
		return addr, err
	}

	return addr, err
}

// upserts an address
func UpsertAddress(c *dgo.Dgraph, address *Address) (*api.Response, error) {
	(*address).Uid = "uid(v)"
	(*address).DType = []string{"Address"}
	pb, err := json.Marshal(address)
	if err != nil {
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
	mu := &api.Mutation{
		SetJson: pb,
	}
	req := &api.Request{
		Query:     query,
		Vars:      vars,
		Mutations: []*api.Mutation{mu},
		CommitNow: true,
	}

	res, err := c.NewTxn().Do(context.Background(), req)
	return res, err
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
		addresses[i].DType = []string{"Address"}
		query += fmt.Sprintf("a%d as var(func: eq(addresshash, $h%d))\n", i, i)
		vars["$h"+strconv.Itoa(i)] = addresses[i].Hash
	}

	queryPrefix += ") {\n"

	pb, err := json.Marshal(addresses)
	if err != nil {
		return nil, err
	}

	mu := &api.Mutation{
		SetJson: pb,
	}
	req := &api.Request{
		Query:     queryPrefix + query + "}",
		Vars:      vars,
		Mutations: []*api.Mutation{mu},
		CommitNow: true,
	}

	res, err := c.NewTxn().Do(context.Background(), req)

	if err != nil {
		return res, err
	}

	return res, err
}
