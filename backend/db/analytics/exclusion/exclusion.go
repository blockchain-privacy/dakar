package exclusion

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

// AddAddressExclusions adds the given address exclusions to the database
func AddAddressExclusions(c external.Database, user User) error {
	if len(user.Exclusions) == 0 {
		return errors.New("nothing to add")
	}

	pb, err := json.Marshal(user)
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

// GetAddressExclusionUIDs returns all UIDs of the excluded addresses of a user
func GetAddressExclusionUIDs(c external.Database, userID string) (exclusions []string, err error) {
	const query = `query Q($user:string) {
				var(func:uid($user))@filter(type(User)){
					a as User.addressExclusions
				}

				q(func: uid(a)){
					uid
				}
			  }`

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*3, query, map[string]string{"$user": userID})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r struct {
		Exclusions []struct {
			UID string `json:"uid,omitempty"`
		} `json:"q,omitempty"`
	}
	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	for _, e := range r.Exclusions {
		exclusions = append(exclusions, e.UID)
	}

	return
}

// GetAddressExclusions returns all address hashes of the excluded addresses of a user.
// Response limited to 30 address hashes.
func GetAddressExclusions(c external.Database, userID string) (addresses []string, count int64, err error) {
	const query = `query Q($user:string) {
				var(func:uid($user))@filter(type(User)){
					a as User.addressExclusions
				}

				x(func:uid(a)){count(uid)}

				q(func: uid(a), first: 30){
					addresshash
				}
			  }`

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*3, query, map[string]string{"$user": userID})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r struct {
		Exclusions []struct {
			AddressHash string `json:"addresshash,omitempty"`
		} `json:"q,omitempty"`
		Count []struct {
			Count int64 `json:"count,omitempty"`
		} `json:"x,omitempty"`
	}
	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if len(r.Count) != 1 {
		err = errors.New("invalid response from database")
		return
	}

	for _, e := range r.Exclusions {
		addresses = append(addresses, e.AddressHash)
	}

	count = r.Count[0].Count

	return
}

// DeleteAddressExclusion deletes the given address exclusion
func DeleteAddressExclusion(c external.Database, userID string, addressHash string) (err error) {
	req := &api.Request{
		Query: `query Q($user:string,$hash:string) {
					a as var(func: eq(addresshash,$hash))@cascade{
						~User.addressExclusions@filter(uid($user))
					}
			  }`,
		Vars: map[string]string{"$user": userID, "$hash": addressHash},
		Mutations: []*api.Mutation{{
			DelNquads: []byte("<" + userID + "> <User.addressExclusions> uid(a) ."),
		}},
		CommitNow: true,
	}
	resp, txErr := db.TxWithRetryAndResponse(c, time.Minute*5, req)
	if txErr != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), txErr)
		return
	}

	// check if there was actually something mutated
	if resp.GetMetrics().NumUids["mutation_cost"] == 0 {
		return errors.New("nothing was deleted")
	}

	return
}

// DeleteAllAddressExclusions deletes all address exclusions of a given user
func DeleteAllAddressExclusions(c external.Database, userID string) (err error) {
	req := &api.Request{
		Mutations: []*api.Mutation{{
			DelNquads: []byte("<" + userID + "> <User.addressExclusions> * ."),
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

// GetAddressExclusionStatus returns true if the given address is part of the users address exclusion list
func GetAddressExclusionStatus(c external.Database, addressHash string, userID string) (isExcluded bool, err error) {
	const query = `query Q($user:string,$hash:string) {
					q(func: eq(addresshash,$hash))@filter(uid_in(~User.addressExclusions,$user)){
						uid
					}
				  }`

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*3, query,
		map[string]string{"$user": userID, "$hash": addressHash})
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	var r struct {
		Address []struct {
			UID string `json:"uid,omitempty"`
		} `json:"q,omitempty"`
	}
	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}
	if len(r.Address) == 0 {
		isExcluded = false
		return
	} else if len(r.Address) > 1 {
		err = errors.New("invalid response from database")
		return
	}

	isExcluded = r.Address[0].UID != ""

	return
}
