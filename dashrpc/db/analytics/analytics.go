package analytics

import (
	"dashrpc/cmd/cliutil"
	"dashrpc/db"
	dbtx "dashrpc/db/transaction"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgraph-io/dgo/v2"
)

// Searches for all potential origins. The returned string slice contains the uids of the found transactions
func AnalyzeOrigins(c *dgo.Dgraph, transactionHash string) (origins []string, err error) {
	query := `query Q($hash: string) {
				tx as var(func: eq(txhash, $hash))
	
				var(func: uid(tx))@recurse{
					tx_inputs
					v as ~tx_outputs@filter(eq(privacytype, ["mixing","origin"]))
				}

				q(func: uid(v))@filter(eq(privacytype,"origin")){
					uid
				}
			  }`

	resp, err := db.ReadOnlyTxVarWithRetry(c, db.GetBackendContext(), query, map[string]string{"$hash": transactionHash})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r struct {
		Transaction []struct {
			Uid string `json:"uid,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	for _, uid := range r.Transaction {
		origins = append(origins, uid.Uid)
	}

	return
}

// gets the number of origins of a transaction
func GetOriginCount(c *dgo.Dgraph, txHash string) (originCount int, err error) {
	query := `query Q($hash: string) {
				q(func: eq(txhash, $hash)){
					count(origins)
				}
			  }`

	resp, err := db.ReadOnlyTxVarWithRetry(c, db.GetBackendContext(), query, map[string]string{"$hash": txHash})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r struct {
		Q []struct {
			OriginCount int `json:"count(origins),omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if len(r.Q) != 1 {
		err = errors.New("invalid response from database")
		return
	}
	originCount = r.Q[0].OriginCount
	return
}

// Searches for all potential origins. The returned string slice contains the uids of the found transactions
func GetPaths(c *dgo.Dgraph, transactionHash string) (paths []TransactionPath, err error) {
	query := `query Q($hash: string) {
				tx as var(func: eq(txhash, $hash))
	
				q(func: uid(tx))@recurse{
					tx_inputs
					~tx_outputs@filter(eq(privacytype, ["mixing","origin"]))
					txhash
					privacytype
				}
			  }`

	resp, err := c.NewReadOnlyTxn().QueryWithVars(db.GetBackendContext(), query, map[string]string{"$hash": transactionHash})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r struct {
		Transaction []transaction `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if len(r.Transaction) != 1 || len(r.Transaction[0].Inputs) == 0 {
		err = errors.New("invalid response from database")
		return
	}

	paths = getTransactionsPaths(r.Transaction[0].Inputs)

	return
}

// starting at inputs get all possible paths to PrivateSend origins
func getTransactionsPaths(inputs []input) (filteredPaths []TransactionPath) {
	var paths []TransactionPath

	findPaths(inputs, &paths, nil)

	// post processing of paths
	pathMap := make(map[string]bool)
	for _, p := range paths {
		cutPath := p.cutTail()
		if len(cutPath) == 0 {
			continue
		}

		pathHash := cutPath.hash()

		// if path is already in map then continue
		if pathMap[pathHash] {
			continue
		}
		pathMap[pathHash] = true
		filteredPaths = append(filteredPaths, cutPath)
	}

	return
}

// saves all paths in inputs to paths; path holds the path of the curren recursion
func findPaths(inputs []input, paths *[]TransactionPath, path TransactionPath) {
	for _, i := range inputs {
		// transaction slice ALWAYS exists and has ALWAYS one element
		tx := i.Transaction[0]

		// copy path
		newPath := make(TransactionPath, len(path))
		copy(newPath, path)

		// add new PathElement
		isOrigin := false
		if tx.PrivacyType == dbtx.PrivacyOrigin {
			isOrigin = true
		}
		newPath = append(newPath, PathElement{
			IsOrigin: isOrigin,
			Hash:     tx.Hash,
		})

		// end of path
		if tx.Inputs == nil {
			*paths = append(*paths, newPath)
			continue
		}

		findPaths(tx.Inputs, paths, newPath)
	}
}
