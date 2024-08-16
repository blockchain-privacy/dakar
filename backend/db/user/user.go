package user

import (
	"backend/db"
	"backend/external"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"github.com/dgraph-io/dgo/v230/protos/api"
	"github.com/qrest/gomisc/serror"
	"time"
)

// CreateNewUser creates a new user
func CreateNewUser(c external.Database) (string, error) {
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

	resp, err := db.TxWithRetryAndResponse(c, time.Minute*5, req)
	if err != nil {
		return "", err
	}

	// check if insert was successful
	if len(resp.Uids) != 1 {
		return "", serror.FromFormat("invalid number of uids returned: %s", resp.Uids)
	}

	var userUID string
	for _, v := range resp.GetUids() {
		userUID = v
	}

	return userUID, nil
}

// GetUserCount returns the number of user currently in the database
func GetUserCount(c external.Database) (userCount int, err error) {
	query := `{
				q(func: type(User)){
					count(uid)
				}
			  }`

	resp, err := db.ReadOnlyTxWithRetry(c, time.Minute*2, query)
	if err != nil {
		return
	}

	// json struct
	var r struct {
		Users []struct {
			Count int `json:"count"`
		} `json:"q"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = serror.New(err)
		return
	}

	userCount = r.Users[0].Count
	return
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
		Q []struct {
			UID string
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
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
		err = serror.FromStr("error user does not exist")
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

// generateRandomPassword returns a random string if fixed length of 22
func generateRandomPassword() (string, error) {
	// Generate a Salt
	pwByte := make([]byte, 16)
	if _, err := rand.Read(pwByte); err != nil {
		return "", serror.New(err)
	}

	return base64.RawStdEncoding.EncodeToString(pwByte), nil
}
