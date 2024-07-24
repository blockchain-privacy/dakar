package user

import (
	"backend/cmd/cliutil"
	"backend/db"
	"backend/external"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"github.com/dgraph-io/dgo/v230/protos/api"
	ory "github.com/ory/kratos-client-go"
	"time"
)

// CreateNewUser creates a new user
func CreateNewUser(c external.Database) (string, error) {
	var usr User

	usr.UID = ""
	usr.SetDType()

	pb, err := json.Marshal(usr)
	if err != nil {
		return "", cliutil.NewStackError(err)
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
		return "", cliutil.NewStackErrorf("invalid number of uids returned: %s", resp.Uids)
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
		err = cliutil.NewStackError(err)
		return
	}

	userCount = r.Users[0].Count
	return
}

// existsUser checks if a User with the given uid exists
func existsUser(c external.Database, uid string) (found bool, err error) {
	query := "query Q($uid:string){q(func: uid($uid))@filter(eq(dgraph.type," + DType + ")){uid}}"

	ctx, cancel := db.GetFrontendContext()
	defer cancel()

	// no retry
	resp, txErr := c.Query(ctx, query, map[string]string{"$uid": uid})
	if txErr != nil {
		err = cliutil.NewStackError(txErr)
		return
	}

	var r struct {
		Q []struct {
			UID string
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = cliutil.NewStackError(err)
		return
	}

	if len(r.Q) != 1 || len(r.Q[0].UID) == 0 {
		return
	}

	found = true

	return
}

// DeleteUser deletes the User with the given uid
func DeleteUser(c external.Database, uid string) (err error) {
	if found, existsErr := existsUser(c, uid); existsErr != nil {
		err = existsErr
		return
	} else if !found {
		err = cliutil.NewStackErrorStr("error user does not exist")
		return
	}

	req := &api.Request{
		Query: "query Q($uid:string){h as var(func: uid($uid))@filter(eq(dgraph.type," + DType + "))}",
		Vars:  map[string]string{"$uid": uid},
		Mutations: []*api.Mutation{{
			DelNquads: []byte("uid(h) * * ."),
		}},
		CommitNow: true,
	}

	return db.TxWithRetry(c, time.Minute*5, req)
}

// CreateDgraphAndKratosUser creates a dgraph user and ory kratos identity.
// The UID of the dgraph user is written the metadata_admin of the new kratos identity.
// The credentials are set if not nil.
// state has to be either "active" or "inactive" (as per ory kratos documentation)
func CreateDgraphAndKratosUser(ctx context.Context, c external.Database, adminAuth *ory.APIClient,
	email string, credentials *ory.IdentityWithCredentials, roles []string, state string) error {
	// check state
	if !IsStateValid(state) {
		return cliutil.NewStackErrorStr("invalid identity state: " + state)
	}

	// create dgraph user
	newUserUID, userCreationError := CreateNewUser(c)
	if userCreationError != nil {
		return userCreationError
	}

	// create kratos identity
	_, _, err := adminAuth.IdentityAPI.CreateIdentity(ctx).CreateIdentityBody(ory.CreateIdentityBody{
		SchemaId:       "default_v0",
		Traits:         map[string]interface{}{"email": email},
		MetadataPublic: map[string]any{"roles": roles, "dgraph_uid": newUserUID},
		Credentials:    credentials,
		State:          &state,
	}).Execute()
	if err != nil {
		return cliutil.NewStackError(err)
	}

	return nil
}

// generateRandomPassword returns a random string if fixed length of 22
func generateRandomPassword() (string, error) {
	// Generate a Salt
	pwByte := make([]byte, 16)
	if _, err := rand.Read(pwByte); err != nil {
		return "", cliutil.NewStackError(err)
	}

	return base64.RawStdEncoding.EncodeToString(pwByte), nil
}

// CreateAdminUser creates a new admin account with a random password
func CreateAdminUser(c external.Database, adminAuth *ory.APIClient, email string) (string, error) {
	pw, err := generateRandomPassword()
	if err != nil {
		return "", err
	}

	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Minute*2)
	defer cancelFunc()

	// get ory handle and also create password
	err = CreateDgraphAndKratosUser(ctx, c, adminAuth, email, &ory.IdentityWithCredentials{
		Password: &ory.IdentityWithCredentialsPassword{
			Config: &ory.IdentityWithCredentialsPasswordConfig{Password: &pw}},
	}, []string{"admin"}, "active")
	if err != nil {
		return "", err
	}

	return pw, nil
}

// CreatePrivilegedUser creates a new privileged user account with the given password
func CreatePrivilegedUser(c external.Database, adminAuth *ory.APIClient, email string, pw string) error {
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Minute*2)
	defer cancelFunc()

	// get ory handle and also create password
	err := CreateDgraphAndKratosUser(ctx, c, adminAuth, email, &ory.IdentityWithCredentials{
		Password: &ory.IdentityWithCredentialsPassword{
			Config: &ory.IdentityWithCredentialsPasswordConfig{Password: &pw}},
	}, []string{"privileged"}, "active")
	if err != nil {
		return err
	}

	return nil
}
