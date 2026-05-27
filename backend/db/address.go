// SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package db

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/dgraph-io/dgo/v250/protos/api"
	"gitlab.com/blockchain-privacy/dakar/external"
	"gitlab.com/blockchain-privacy/gomisc/serror"
)

// ErrAddressNotFound is returned if no address has been found
var ErrAddressNotFound = errors.New("no address found")

// AddressDType is the dgraph database type for the Address type
const AddressDType = "Address"

// if an address has more outputs than maximumSortAndFilterOutputs sorting and filtering will be ignored
const maximumSortAndFilterOutputs = 10000

const (
	// SortAscendingByOutputTime sort outputs ascending by the output transaction timestamp
	SortAscendingByOutputTime int = iota
	// SortDescendingByOutputTime sort outputs descending by the output transaction timestamp
	SortDescendingByOutputTime
	// SortAscendingByInputTime sort outputs ascending by the input transaction timestamp
	SortAscendingByInputTime
	// SortDescendingByInputTime sort outputs descending by the input transaction timestamp
	SortDescendingByInputTime
	// SortAscendingByAmount sort outputs ascending by the output amount
	SortAscendingByAmount
	// SortDescendingByAmount sort outputs ascending by the output amount
	SortDescendingByAmount
)

const (
	// FilterByCoinbase filters outputs if they are a coinbase output
	FilterByCoinbase int = iota
	// FilterByUnspent filters outputs if they are unspent
	FilterByUnspent
)

// IsValidSortOrder returns true if sortOrder has a valid sort order value
func IsValidSortOrder(sortOrder int) bool {
	return sortOrder == SortAscendingByInputTime || sortOrder == SortDescendingByInputTime ||
		sortOrder == SortAscendingByOutputTime || sortOrder == SortDescendingByOutputTime ||
		sortOrder == SortAscendingByAmount || sortOrder == SortDescendingByAmount
}

// IsValidFilter returns true if filters had a valid value
func IsValidFilter(filters []int) bool {
	for _, f := range filters {
		if f != FilterByUnspent && f != FilterByCoinbase {
			return false
		}
	}

	return true
}

// Address holds data for the database address type
type Address struct {
	UID     string   `json:"uid,omitempty"`
	Hash    string   `json:"addresshash,omitempty"`
	Outputs []Output `json:"addr_outputs,omitempty"`
	DType   []string `json:"dgraph.type,omitempty"`
}

func (a *Address) String() string {
	output := fmt.Sprintf("UID: %s, Hash: %s", a.UID, a.Hash)

	if a.Outputs != nil {
		output += fmt.Sprintf(", OutputCount: %d", len(a.Outputs))
	}

	return output
}

// SetDType sets the DType for dgraph type recognition
func (a *Address) SetDType() {
	a.DType = []string{AddressDType}
}

// FrontendOutput is the representation for the frontend of an output
type FrontendOutput struct {
	Amount                int64  `json:"amount"`
	InputTransactionHash  string `json:"inputTransactionHash"`
	InputTimestamp        string `json:"inputTimestamp"`
	OutputTransactionHash string `json:"outputTransactionHash"`
	OutputTimestamp       string `json:"outputTimestamp"`
}

func (o FrontendOutput) String() string {
	return fmt.Sprintf("Amount: %d", o.Amount)
}

// FrontendAddress is the representation for the frontend of an address
type FrontendAddress struct {
	Hash string `json:"addresshash"`
	// QueryMaxCount is the number results for the given filter.
	// If IsOutputManipulationSupported is true, this number will always be zero and should be ignored.
	QueryMaxCount int64            `json:"queryMaxCount"`
	OutputCount   int64            `json:"outputCount"`
	InputCount    int64            `json:"inputCount"`
	InputSum      int64            `json:"inputSum"`
	OutputSum     int64            `json:"outputSum"`
	Outputs       []FrontendOutput `json:"outputs"`
	// IsOutputManipulationSupported is true if the address has too many outputs to allow performant sorting and filtering
	IsOutputManipulationSupported bool `json:"isOutputManipulationSupported"`
}

func (f FrontendAddress) String() string {
	return fmt.Sprintf("Hash: %s, OutputCount: %d", f.Hash, len(f.Outputs))
}

// GetFrontendAddress returns the address with outputs filter and sorted as specified.
// If the address has too many outputs the filter and sort order will be ignored.
func GetFrontendAddress(ctx context.Context, c external.Database, addrHash string, sortOrder int,
	offset int, filters []int) (*FrontendAddress, error) {
	addr, err := GetFrontendAddressHeader(ctx, c, addrHash)
	if err != nil {
		return nil, err
	}

	if addr.OutputCount > maximumSortAndFilterOutputs {
		outputs, err := getFrontendAddressOutputs(ctx, c, addrHash, offset)
		if err != nil {
			return nil, err
		}

		addr.QueryMaxCount = 0
		addr.IsOutputManipulationSupported = false
		addr.Outputs = outputs
	} else {
		outputs, queryMaxCount, err := GetFrontendAddressOutputsSortAndFilter(ctx, c, addrHash, sortOrder, offset, filters)
		if err != nil {
			return nil, err
		}

		addr.QueryMaxCount = queryMaxCount
		addr.IsOutputManipulationSupported = true
		addr.Outputs = outputs
	}

	return &addr, nil
}

// GetFrontendAddressOutputsSortAndFilter returns outputs for the address sorted as specified by sortOrder.
// Use one of the constants like SortAscendingByInputTime to set the sortOrder
func GetFrontendAddressOutputsSortAndFilter(ctx context.Context, c external.Database, addrHash string, sortOrder int,
	offset int, filters []int) (outputs []FrontendOutput, queryMaxCount int64, err error) {
	if addrHash == "" {
		err = serror.New(ErrEmptyRequestArgument)
		return
	}

	const maxOutputsPerQuery = 20
	sortDirection := "asc"
	sortBy := "val(ots)"

	const sortDescending = "desc"

	switch sortOrder {
	case SortAscendingByInputTime:
		sortBy = "val(its)"
	case SortDescendingByInputTime:
		sortDirection = sortDescending
		sortBy = "val(its)"
	case SortAscendingByOutputTime:
		// do nothing, values are already correctly set
	case SortDescendingByOutputTime:
		sortDirection = sortDescending
	case SortAscendingByAmount:
		sortBy = "amount"
	case SortDescendingByAmount:
		sortDirection = sortDescending
		sortBy = "amount"
	default:
		err = serror.FromStr("error unrecognized sort order")
		return
	}

	filterBuilder := strings.Builder{}
	for i, f := range filters {
		switch f {
		case FilterByCoinbase:
			filterBuilder.WriteString("eq(iscoinbase, true)")
		case FilterByUnspent:
			filterBuilder.WriteString(" NOT has(~tx_inputs)")
		default:
			err = serror.FromStr("error unrecognized filter")
			return
		}

		if i+1 < len(filters) {
			filterBuilder.WriteString(" AND ")
		}
	}

	filter := filterBuilder.String()
	if len(filters) > 0 {
		filter = fmt.Sprintf("@filter(%s)", filter)
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

		c(func:uid(a), orderdesc: ` + sortBy + ")" + filter + `{
			count(uid)
        }

		q(func: uid(a), order` + sortDirection + ":" + sortBy + ", first:" +
		strconv.Itoa(maxOutputsPerQuery) + ",offset:" + strconv.Itoa(offset) + ")" + filter + `@normalize{
			amount:amount
			outputTimestamp:val(ots)
			inputTimestamp:val(its)
			~tx_outputs{
				outputTransactionHash: txhash
			}
			~tx_inputs{
				inputTransactionHash: txhash
			}
		}
	}`

	vars := make(map[string]string)
	vars["$hash"] = addrHash
	resp, err := c.Query(ctx, query, vars)
	if err != nil {
		err = serror.New(err)
		return
	}

	var r struct {
		Outputs       []FrontendOutput `json:"q"`
		QueryMaxCount []struct {
			Count int64 `json:"count"`
		} `json:"c"`
	}

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		err = serror.New(err)
		return
	}

	if len(r.QueryMaxCount) != 1 {
		err = serror.New(errInvalidResult)
		return
	}

	queryMaxCount = r.QueryMaxCount[0].Count
	outputs = r.Outputs

	return
}

// getFrontendAddressOutputs returns outputs for the address with the given offset
func getFrontendAddressOutputs(ctx context.Context, c external.Database, addrHash string,
	offset int) (outputs []FrontendOutput, err error) {
	if addrHash == "" {
		err = serror.New(ErrEmptyRequestArgument)
		return
	}

	const maxOutputsPerQuery = 20

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

		q(func: uid(a), first:` + strconv.Itoa(maxOutputsPerQuery) + ",offset:" + strconv.Itoa(offset) + `)@normalize{
			amount:amount
			outputTimestamp:val(ots)
			inputTimestamp:val(its)
			~tx_outputs{
				outputTransactionHash: txhash
			}
			~tx_inputs{
				inputTransactionHash: txhash
			}
		}
	}`

	resp, err := c.Query(ctx, query, map[string]string{"$hash": addrHash})
	if err != nil {
		err = serror.New(err)
		return
	}

	var r struct {
		Outputs []FrontendOutput `json:"q"`
	}

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		err = serror.New(err)
		return
	}

	outputs = r.Outputs

	return
}

// UpsertAddresses upserts the given addresses. The list must not contain duplicate addresses.
func UpsertAddresses(ctx context.Context, c external.Database, addresses []Address) error {
	if addresses == nil {
		return serror.FromStr("got null pointer for addresses")
	}

	// The following block creates the query for 4 addresses:
	//		query Q($h0: string,$h1: string,$h2: string,$h3: string) {
	//		a0 as var(func: eq(addresshash, $h0))
	//		a1 as var(func: eq(addresshash, $h1))
	//		a2 as var(func: eq(addresshash, $h2))
	//		a3 as var(func: eq(addresshash, $h3))
	//		}

	vars := make(map[string]string)
	query := strings.Builder{}
	queryPrefix := strings.Builder{}

	queryPrefix.WriteString("query Q(")
	// set uid for all addresses and build query
	for i := range addresses {
		queryPrefix.WriteString("$h" + strconv.Itoa(i) + ": string")

		if i+1 < len(addresses) {
			queryPrefix.WriteRune(',')
		}

		addresses[i].UID = fmt.Sprintf("uid(a%d)", i)
		addresses[i].SetDType()
		_, err := fmt.Fprintf(&query, "a%d as var(func: eq(addresshash, $h%d))\n", i, i)
		if err != nil {
			return serror.New(err)
		}
		vars["$h"+strconv.Itoa(i)] = addresses[i].Hash
	}

	queryPrefix.WriteString(") {\n")

	pb, err := json.Marshal(addresses)
	if err != nil {
		return serror.New(err)
	}

	return MutationWithRetry(ctx, c, &api.Request{
		Query: queryPrefix.String() + query.String() + "}",
		Vars:  vars,
		Mutations: []*api.Mutation{{
			SetJson: pb,
		}},
		CommitNow: true,
	})
}

// GetAddressUIDs returns all requested address nodes.
func GetAddressUIDs(ctx context.Context, c external.Database, addressHashes []string) ([]Address, error) {
	if len(addressHashes) == 0 {
		return nil, serror.New(ErrEmptyRequestArgument)
	}

	for _, a := range addressHashes {
		if !isValidQueryInput(a) {
			return nil, serror.FromStr("invalid address hash")
		}
	}

	query := `{
				q(func: eq(addresshash,` + CreateCommaArray(addressHashes) + `)){
					uid
					addresshash
				}
			  }`

	resp, err := c.Query(ctx, query, nil)
	if err != nil {
		return nil, serror.New(err)
	}
	var r struct {
		Addresses []Address `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		return nil, serror.New(err)
	}

	return r.Addresses, nil
}

// GetAddressUIDsFilterPublicAttribution returns all requested address nodes.
// If an address has a public attribution attached it will be excluded.
func GetAddressUIDsFilterPublicAttribution(ctx context.Context, c external.Database, addressHashes []string) ([]Address, error) {
	if len(addressHashes) == 0 {
		return nil, serror.New(ErrEmptyRequestArgument)
	}

	for _, a := range addressHashes {
		if !isValidQueryInput(a) {
			return nil, serror.FromStr("invalid address hash")
		}
	}

	query := `{
				a as var(func: eq(addresshash,` + CreateCommaArray(addressHashes) + `))

				with_attribution as var(func: uid(a))@filter(has(~Attribution.address))
				without_attribution as var(func: uid(a))@filter(not has(~Attribution.address))
				  
				non_public_attr as var(func: uid(with_attribution)) @cascade{
					~Attribution.address@filter(eq(Attribution.isPublic, false))
				}
				
				q(func: uid(without_attribution, non_public_attr)) {
					uid
					addresshash
				}
			  }`

	resp, err := c.Query(ctx, query, nil)
	if err != nil {
		return nil, serror.New(err)
	}
	var r struct {
		Addresses []Address `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		return nil, serror.New(err)
	}

	return r.Addresses, nil
}

// GetAddressesByBlockRange returns all address-output mappings of the given block range
func GetAddressesByBlockRange(ctx context.Context, c external.Database, blockHeightStart int, blockHeightEnd int,
	convertUIDs bool) (addresses []Address, err error) {
	const query = `query Q($start: string,$end: string) {
				var(func: between(id,$start,$end)) {
					transactions {
						o as tx_outputs
						i as tx_inputs
					}
				}
				
				var(func: uid(o,i)){
					a as ~addr_outputs
				}
				
				q(func: uid(a)){
					uid 
					addresshash
					dgraph.type
					addr_outputs@filter(uid(o,i)){
						uid
					}
				}
			  }`

	resp, err := QueryVarWithRetry(ctx, c, query,
		map[string]string{"$start": strconv.Itoa(blockHeightStart), "$end": strconv.Itoa(blockHeightEnd)})
	if err != nil {
		return
	}

	var r struct {
		Addresses []Address `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		err = serror.New(err)
		return
	}

	if convertUIDs {
		for i := range r.Addresses {
			r.Addresses[i].UID = "_:" + r.Addresses[i].UID

			for y := range r.Addresses[i].Outputs {
				r.Addresses[i].Outputs[y].UID = "_:" + r.Addresses[i].Outputs[y].UID
			}
		}
	}

	addresses = r.Addresses

	return
}

// GetFrontendAddressHeader return address information
func GetFrontendAddressHeader(ctx context.Context, c external.Database, addrHash string) (addr FrontendAddress, err error) {
	if addrHash == "" {
		err = serror.New(ErrEmptyRequestArgument)
		return
	}

	// fill variables
	const query = `query Q($hash: string){
		var(func: eq(addresshash, $hash)){
			addr_outputs{
				a as uid
				oamt as amount
			}
		}
		var(func: uid(a))@filter(has(~tx_inputs)){
    		iamt as amount
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
	}`

	resp, err := c.Query(ctx, query, map[string]string{"$hash": addrHash})
	if err != nil {
		err = serror.New(err)
		return
	}

	var r struct {
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

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		err = serror.New(err)
		return addr, err
	}

	if len(r.InputSum) != 1 || len(r.OutputSum) != 1 ||
		len(r.InputCount) != 1 || len(r.OutputCount) != 1 {
		err = serror.New(errInvalidResult)
		return
	}

	// not checking the length of r.Outputs, as for certain filters the number of outputs can be 0
	// instead check for the calculated output count
	if r.OutputCount[0].Count == 0 {
		err = serror.New(ErrAddressNotFound)
		return
	}

	addr = FrontendAddress{
		Hash:        addrHash,
		InputCount:  r.InputCount[0].Count,
		OutputCount: r.OutputCount[0].Count,
		InputSum:    r.InputSum[0].Sum,
		OutputSum:   r.OutputSum[0].Sum,
	}

	return
}
