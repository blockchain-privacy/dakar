// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package db

import (
	"context"
	"encoding/json"
	"gitlab.com/blockchain-privacy/dakar/external"
	"gitlab.com/blockchain-privacy/gomisc/serror"
	"strconv"
)

// Search returns the type of entities the query matches
func Search(ctx context.Context, c external.Database, queryString string) (string, error) {
	if queryString == "" {
		return "", serror.New(ErrEmptyRequestArgument)
	}

	var query string
	_, err := strconv.Atoi(queryString)
	if err == nil {
		query = `query Q($q:string){
				tx(func: eq(txhash,$q)){uid}
				blockID(func: eq(id,$q)){uid}
				blockHash(func: eq(blockhash,$q)){uid}
				address(func: eq(addresshash,$q)){uid}
			  }`
	} else {
		// don't query for block ID because query is not a number
		query = `query Q($q:string){
				tx(func: eq(txhash,$q)){uid}
				blockHash(func: eq(blockhash,$q)){uid}
				address(func: eq(addresshash,$q)){uid}
			  }`
	}

	resp, err := c.Query(ctx, query, map[string]string{"$q": queryString})
	if err != nil {
		return "", serror.New(err)
	}

	// json struct
	var r struct {
		Transactions []UIDNode `json:"tx,omitempty"`
		BlocksByID   []UIDNode `json:"blockID,omitempty"`
		BlocksByHash []UIDNode `json:"blockHash,omitempty"`
		Address      []UIDNode `json:"address,omitempty"`
	}

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		return "", serror.New(err)
	}

	if len(r.Transactions) > 0 {
		return "tx", nil
	}

	if len(r.BlocksByID) > 0 {
		return "block", nil
	}

	if len(r.BlocksByHash) > 0 {
		return "block", nil
	}

	if len(r.Address) > 0 {
		return "addr", nil
	}

	return "", nil
}
