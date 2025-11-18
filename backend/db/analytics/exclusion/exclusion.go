// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package exclusion

import (
	"backend/db"
	"backend/external"
	"context"
	"encoding/json"
	"github.com/dgraph-io/dgo/v250/protos/api"
	"gitlab.com/blockchain-privacy/gomisc/serror"
)

// AddAddressExclusions adds the given address exclusions to the database
func AddAddressExclusions(ctx context.Context, c external.Database, user User) error {
	if len(user.Exclusions) == 0 {
		return serror.FromStr("nothing to add")
	}

	pb, err := json.Marshal(user)
	if err != nil {
		return serror.New(err)
	}

	_, err = c.Mutate(ctx, &api.Request{
		Mutations: []*api.Mutation{{SetJson: pb}},
		CommitNow: true,
	})
	if err != nil {
		return serror.New(err)
	}

	return nil
}

// GetAddressExclusionUIDs returns all UIDs of the excluded addresses of a user
func GetAddressExclusionUIDs(ctx context.Context, c external.Database, userID string) (exclusions []string, err error) {
	const query = `query Q($user:string) {
				var(func:uid($user))@filter(type(User)){
					a as User.addressExclusions
				}

				q(func: uid(a)){
					uid
				}
			  }`

	resp, err := db.QueryVarWithRetry(ctx, c, query, map[string]string{"$user": userID})
	if err != nil {
		return
	}

	var r struct {
		Exclusions []db.UIDNode `json:"q,omitempty"`
	}
	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		err = serror.New(err)
		return
	}

	for _, e := range r.Exclusions {
		exclusions = append(exclusions, e.UID)
	}

	return
}

// GetAddressExclusions returns all address hashes of the excluded addresses of a user.
// Response limited to 30 address hashes.
func GetAddressExclusions(ctx context.Context, c external.Database,
	userID string) (addresses []string, count int64, err error) {
	const query = `query Q($user:string) {
				var(func:uid($user))@filter(type(User)){
					a as User.addressExclusions
				}

				x(func:uid(a)){count(uid)}

				q(func: uid(a), first: 30){
					addresshash
				}
			  }`

	resp, err := c.Query(ctx, query, map[string]string{"$user": userID})
	if err != nil {
		return nil, 0, serror.New(err)
	}

	var r struct {
		Exclusions []struct {
			AddressHash string `json:"addresshash,omitempty"`
		} `json:"q,omitempty"`
		Count []struct {
			Count int64 `json:"count,omitempty"`
		} `json:"x,omitempty"`
	}
	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		err = serror.New(err)
		return
	}

	if len(r.Count) != 1 {
		err = serror.FromStr("invalid response from database")
		return
	}

	for _, e := range r.Exclusions {
		addresses = append(addresses, e.AddressHash)
	}

	count = r.Count[0].Count

	return
}

// DeleteAddressExclusion deletes the given address exclusion
func DeleteAddressExclusion(ctx context.Context, c external.Database, userID string, addressHash string) error {
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
	resp, err := db.MutationWithRetryAndResponse(ctx, c, req)
	if err != nil {
		return err
	}

	// check if there was actually something mutated
	if !db.HasMutationCost(resp) {
		return serror.FromStr("nothing was deleted")
	}

	return nil
}

// DeleteAllAddressExclusions deletes all address exclusions of a given user
func DeleteAllAddressExclusions(ctx context.Context, c external.Database, userID string) error {
	req := &api.Request{
		Mutations: []*api.Mutation{{
			DelNquads: []byte("<" + userID + "> <User.addressExclusions> * ."),
		}},
		CommitNow: true,
	}

	return db.MutationWithRetry(ctx, c, req)
}

// GetAddressExclusionStatus returns true if the given address is part of the users address exclusion list
func GetAddressExclusionStatus(ctx context.Context, c external.Database,
	addressHash string, userID string) (isExcluded bool, err error) {
	const query = `query Q($user:string,$hash:string) {
					q(func: eq(addresshash,$hash))@filter(uid_in(~User.addressExclusions,$user)){
						uid
					}
				  }`

	resp, err := c.Query(ctx, query, map[string]string{"$user": userID, "$hash": addressHash})
	if err != nil {
		err = serror.New(err)
		return
	}

	var r struct {
		Address []db.UIDNode `json:"q,omitempty"`
	}
	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		err = serror.New(err)
		return
	}
	if len(r.Address) == 0 {
		isExcluded = false
		return
	} else if len(r.Address) > 1 {
		err = serror.FromStr("invalid response from database")
		return
	}

	isExcluded = r.Address[0].UID != ""

	return
}
