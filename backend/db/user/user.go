// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package user

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/dgraph-io/dgo/v250/protos/api"
	"gitlab.com/blockchain-privacy/dakar/db"
	"gitlab.com/blockchain-privacy/dakar/external"
	"gitlab.com/blockchain-privacy/gomisc/serror"
)

var ErrUserDoesNotExist = errors.New("user does not exist")

// CreateNewUser creates a new user
func CreateNewUser(ctx context.Context, c external.Database) (string, error) {
	var usr User

	usr.UID = ""
	usr.SetDType()

	pb, err := json.Marshal(usr)
	if err != nil {
		return "", serror.New(err)
	}

	req := &api.Request{
		Mutations: []*api.Mutation{{
			SetJson: pb,
		}},
		CommitNow: true,
	}

	resp, err := db.MutationWithRetryAndResponse(ctx, c, req)
	if err != nil {
		return "", err
	}

	// check if insert was successful
	if len(resp.GetUids()) != 1 {
		return "", serror.FromFormat("invalid number of uids returned: %s", resp.GetUids())
	}

	var userUID string
	for _, v := range resp.GetUids() {
		userUID = v
	}

	return userUID, nil
}

// existsUser checks if a User with the given uid exists
func existsUser(ctx context.Context, c external.Database, uid string) (found bool, err error) {
	query := "query Q($uid:string){q(func: uid($uid))@filter(eq(dgraph.type," + DType + ")){uid}}"

	resp, txErr := db.QueryVarWithRetry(ctx, c, query, map[string]string{"$uid": uid})
	if txErr != nil {
		err = serror.New(txErr)
		return
	}

	var r struct {
		Q []db.UIDNode `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		err = serror.New(err)
		return
	}

	if len(r.Q) != 1 || len(r.Q[0].UID) == 0 {
		return
	}

	found = true

	return
}

// DeleteUser deletes the User with the given uid
func DeleteUser(ctx context.Context, c external.Database, uid string) (err error) {
	if found, existsErr := existsUser(ctx, c, uid); existsErr != nil {
		err = existsErr
		return
	} else if !found {
		err = serror.NewWithContext(ErrUserDoesNotExist, "user", uid)
		return
	}

	req := &api.Request{
		Query:     "query Q($uid:string){h as var(func: uid($uid))@filter(eq(dgraph.type," + DType + "))}",
		Vars:      map[string]string{"$uid": uid},
		Mutations: []*api.Mutation{{DelNquads: []byte("uid(h) * * .")}},
		CommitNow: true,
	}

	resp, err := db.MutationWithRetryAndResponse(ctx, c, req)
	if err != nil {
		return err
	}

	if !db.HasMutationCost(resp) {
		return serror.New(db.ErrNoMutationHappened)
	}

	return nil
}
