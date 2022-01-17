package attribution

import (
	"backend/cmd/cliutil"
	"backend/db"
	"backend/external"
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/dgo/v210/protos/api"
	"time"
)

// AddAttributions adds the given attributions to the database
func AddAttributions(c external.Database, attributions []Attribution) error {
	// validate data
	for _, a := range attributions {
		if a.Address.Uid == "" || a.Tag == "" || a.Timestamp == "" || a.Uid == "" {
			return fmt.Errorf("attribution invalid: %v", a)
		}
	}

	pb, err := json.Marshal(attributions)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return err
	}

	req := &api.Request{
		Mutations: []*api.Mutation{{
			SetJson: pb,
		}},
		CommitNow: true,
	}
	err = db.TxWithRetry(c, time.Minute*5, req)
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	return err
}
