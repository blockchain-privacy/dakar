package analytics

import (
	"dashrpc/cmd/cliutil"
	"dashrpc/db"
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

// gets transaction information from the database including uids of origins
func GetOrigins(c *dgo.Dgraph, txHash string) (origins []Origin, err error) {
	query := `query Q($hash: string) {
				q(func: eq(txhash, $hash)){
					origins{
						uid
						txhash
						privacytype
						fee
						block: ~transactions {
							blockhash
							ts
							id
						}
					}
				}
			  }`

	resp, err := c.NewReadOnlyTxn().QueryWithVars(db.GetFrontendContext(), query, map[string]string{"$hash": txHash})

	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r struct {
		Q []struct {
			Origins []struct {
				Uid   string `json:"uid,omitempty"`
				Hash  string `json:"txhash,omitempty"`
				Block []struct {
					Hash string `json:"blockhash,omitempty"`
					Ts   string `json:"ts,omitempty"`
					Id   uint64 `json:"id,omitempty"`
				} `json:"block,omitempty"`
			} `json:"origins,omitempty"`
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

	for _, o := range r.Q[0].Origins {
		if len(o.Block) == 0 {
			err = errors.New("invalid response from database")
			return
		}

		origins = append(origins, Origin{
			Uid:            o.Uid,
			Hash:           o.Hash,
			BlockHash:      o.Block[0].Hash,
			BlockId:        o.Block[0].Id,
			BlockTimestamp: o.Block[0].Ts,
		})
	}

	return
}

// gets the shortest path between uidFrom and uidTo
func GetShortestPath(c *dgo.Dgraph, uidFrom string, uidTo string) (p []PathElement, err error) {
	query := fmt.Sprintf(`{
				shortest(from: %s, to: %s){
					tx_inputs
					~tx_outputs@filter(eq(privacytype, ["mixing","origin"]))
				}
			  }`, uidFrom, uidTo)

	resp, err := db.ReadOnlyTxWithRetry(c, db.GetBackendContext(), query)

	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r struct {
		Path []transaction `json:"_path_,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if len(r.Path) == 0 || r.Path[0].Uid == "" || r.Path[0].Input == nil {
		err = errors.New("error invalid path")
		return
	}

	p = convertPath(&r.Path[0])

	return
}

// gets the weight of the shortest path between uidFrom and uidTo
func GetShortestPathWeight(c *dgo.Dgraph, uidFrom string, uidTo string) (weight float64, err error) {
	query := fmt.Sprintf(`{
				shortest(from: %s, to: %s){
					tx_inputs
					~tx_outputs@filter(eq(privacytype, ["mixing","origin"]))
				}
			  }`, uidFrom, uidTo)

	resp, err := db.ReadOnlyTxWithRetry(c, db.GetBackendContext(), query)

	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r struct {
		Path []struct {
			Weight float64 `json:"_weight_,omitempty"`
		} `json:"_path_,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if len(r.Path) == 0 {
		err = errors.New("error invalid path")
		return
	}

	weight = r.Path[0].Weight

	return
}

// gets up to numPath paths between uidFrom and uidTo up to depth
func GetPaths(c *dgo.Dgraph, uidFrom string, uidTo string, numPaths uint32, depth uint32) (paths [][]PathElement, err error) {
	query := fmt.Sprintf(`{
				shortest(from: %s, to: %s, numpaths: %d, depth: %d){
					tx_inputs
					~tx_outputs@filter(eq(privacytype, ["mixing","origin"]))
				}
			  }`, uidFrom, uidTo, numPaths, depth)

	resp, err := db.ReadOnlyTxWithRetry(c, db.GetBackendContext(), query)

	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r struct {
		Path []transaction `json:"_path_,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if len(r.Path) == 0 {
		err = errors.New("error invalid path")
		return
	}

	// || r.Path[0].Uid == "" || r.Path[0].Input == nil

	for _, p := range r.Path {
		paths = append(paths, convertPath(&p))
	}

	return
}

func convertPath(firstTransaction *transaction) (p []PathElement) {
	var nextInput *input
	var nextTransaction *transaction

	// set start element
	nextTransaction = firstTransaction

	// add path elements
	for {
		if nextInput != nil {
			p = append(p, PathElement{
				uid:           nextInput.Uid,
				isTransaction: false,
			})
			nextTransaction = nextInput.Transaction
			nextInput = nil
		} else if nextTransaction != nil {
			p = append(p, PathElement{
				uid:           nextTransaction.Uid,
				isTransaction: true,
			})
			nextInput = nextTransaction.Input
			nextTransaction = nil
		} else {
			break
		}
	}
	return
}
