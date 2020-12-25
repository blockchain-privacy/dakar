package main

import (
	dbaddr "dashrpc/db/address"
	dbblk "dashrpc/db/block"
	dbtx "dashrpc/db/transaction"
	"errors"
	"github.com/dgraph-io/dgo/v2"
)

// isLikelyBlock returns true if the given query string is likely a block hash
func isLikelyBlock(query string) bool {
	return query[0:1] == "0"
}

// isLikelyAddress returns true if the given query string is likely a address hash
func isLikelyAddress(query string) bool {
	return query[0:1] == "X"
}

type SearchResult struct {
	resultType string
	result     interface{}
}

// GetBlock searches for the hash specified in query. If a block is found the returned bool is true
func GetBlock(dgraph *dgo.Dgraph, query string) (SearchResult, bool, error) {
	block, err := dbblk.GetFrontendBlock(dgraph, query)
	if err != nil {

		// only print error if it is not expected
		if !errors.Is(err, dbblk.ErrorBlockNotFound) {
			return SearchResult{}, false, err
		}
		return SearchResult{}, false, nil
	}

	return SearchResult{resultType: "block", result: block}, true, nil
}

// GetTransaction searches for the hash specified in query. If a transaction is found the returned bool is true
func GetTransaction(dgraph *dgo.Dgraph, query string) (SearchResult, bool, error) {
	tx, err := dbtx.GetFrontendTransaction(dgraph, query)
	if err != nil {

		// only print error if it is not expected
		if !errors.Is(err, dbtx.ErrorTransactionNotFound) {
			return SearchResult{}, false, err
		}
		return SearchResult{}, false, nil
	}

	return SearchResult{resultType: "tx", result: tx}, true, nil
}

// GetAddress searches for the hash specified in query. If an address is found the returned bool is true.
// A maximum of 20 elements is returned.
func GetAddress(dgraph *dgo.Dgraph, query string) (SearchResult, bool, error) {
	return GetAddressWithOptions(dgraph, query, dbaddr.SortAscendingByOutputTime, 0)
}

// GetAddressWithOptions searches for the hash specified in query. If an address is found the returned bool is true.
// It supports sorting and setting an offset. For sorting use the constants defined in the db address module.
// A maximum of 20 elements is returned.
func GetAddressWithOptions(dgraph *dgo.Dgraph, query string, sortOrder int, offset int) (SearchResult, bool, error) {
	addr, err := dbaddr.GetFrontendAddress(dgraph, query, sortOrder, offset)
	if err != nil {

		// only print error if it is not expected
		if !errors.Is(err, dbaddr.ErrorAddressNotFound) {
			return SearchResult{}, false, err
		}
		return SearchResult{}, false, nil
	}

	return SearchResult{resultType: "addr", result: addr}, true, nil
}
