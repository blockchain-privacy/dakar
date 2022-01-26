package attribution

import (
	"backend/cmd/cliutil"
	"backend/db"
	"backend/external"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgraph-io/dgo/v210/protos/api"
	"time"
)

// AddAttributions adds the given attributions to the database
func AddAttributions(c external.Database, attributions []Attribution) error {
	// validate data
	for _, a := range attributions {
		if a.Address.Uid == "" || a.Tag == "" || a.Timestamp == "" {
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

// GetUserAttributions returns all attributions of a user
func GetUserAttributions(c external.Database, userID string) (attributions []FrontendAttribution, err error) {
	const query = `query Q($user:string) {
				var(func:uid($user))@filter(type(User)){
					a as ~attribution_user
				}

				q(func: uid(a)){
					uid
					attribution_ts
					attribution_tag
					attribution_address{
						addresshash
					}
				}
			  }`

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*3, query, map[string]string{"$user": userID})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r struct {
		Attributions []struct {
			Uid       string `json:"uid,omitempty"`
			Timestamp string `json:"attribution_ts,omitempty"`
			Tag       string `json:"attribution_tag,omitempty"`
			Address   struct {
				Hash string `json:"addresshash,omitempty"`
			} `json:"attribution_address,omitempty"`
		} `json:"q,omitempty"`
	}
	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	for _, attribution := range r.Attributions {
		attributions = append(attributions, FrontendAttribution{
			Uid:       attribution.Uid,
			Timestamp: attribution.Timestamp,
			Address:   attribution.Address.Hash,
			Tag:       attribution.Tag,
		})
	}

	return
}

// DeleteAttribution deletes the given attribution
func DeleteAttribution(c external.Database, userID string, attributionUID string) (err error) {
	req := &api.Request{
		Query: `query Q($user:string,$attribution:string) {
				var(func:uid($user))@filter(type(User)){
					a as ~attribution_user@filter(uid($attribution))
				}
			  }`,
		Vars: map[string]string{"$user": userID, "$attribution": attributionUID},
		Mutations: []*api.Mutation{{
			DelNquads: []byte("uid(a) * * ."),
		}},
		CommitNow: true,
	}
	resp, txErr := db.TxWithRetryAndResponse(c, time.Minute*5, req)
	if txErr != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), txErr)
		return
	}

	if resp.GetMetrics().NumUids["mutation_cost"] == 0 {
		return errors.New("nothing was deleted")
	}

	return
}

// DeleteAllAttributions deletes all attributions of a given user
func DeleteAllAttributions(c external.Database, userID string) (err error) {
	req := &api.Request{
		Query: `query Q($user:string) {
				var(func:uid($user))@filter(type(User)){
					a as ~attribution_user
				}
			  }`,
		Vars: map[string]string{"$user": userID},
		Mutations: []*api.Mutation{{
			DelNquads: []byte("uid(a) * * ."),
		}},
		CommitNow: true,
	}
	resp, txErr := db.TxWithRetryAndResponse(c, time.Minute*5, req)
	if txErr != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), txErr)
		return
	}

	if resp.GetMetrics().NumUids["mutation_cost"] == 0 {
		return errors.New("nothing was deleted")
	}

	return
}
