package analytics

import (
	"backend/cmd/cliutil"
	"backend/db"
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/dgo/v2"
	"time"
)

// GetPrivacyTypeData returns timestamps when the transactions of the specified privacyType occur.
// If the string is empty then all privacy transactions are considered.
func GetPrivacyTypeData(c *dgo.Dgraph, startRange string, stopRange string) (ts []time.Time, err error) {
	const query = `query Q($ge:string,$le:string){
				q(func:between(privacytype,$ge,$le))@normalize{
					~transactions{
						ts:ts
					}
			  	}
			  }`

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*5, query,
		map[string]string{"$ge": startRange, "$le": stopRange})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	// json struct
	var r struct {
		Query []struct {
			Timestamp time.Time `json:"ts,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	for _, q := range r.Query {
		ts = append(ts, q.Timestamp)
	}

	return
}
