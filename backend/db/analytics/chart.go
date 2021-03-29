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
func GetPrivacyTypeData(c *dgo.Dgraph, privacyType string) (ts []time.Time, err error) {
	query := `query Q($pt: string){
				q(func: eq(privacytype,$pt))@normalize{
					~transactions{
						ts:ts
					}
			  	}
			  }`

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*5, query,
		map[string]string{"$pt": privacyType})
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
