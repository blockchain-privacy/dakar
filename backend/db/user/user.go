package user

import (
	"backend/cmd/cliutil"
	"backend/db"
	"backend/external"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	ory "github.com/ory/kratos-client-go"
	"time"

	"github.com/dgraph-io/dgo/v210/protos/api"
)

var (
	// ErrorUsersNotFound is returned if no users are found
	ErrorUsersNotFound = errors.New("error no users found")
)

// CreateNewUser creates a new user
func CreateNewUser(c external.Database) (string, error) {
	var usr User

	usr.UID = ""
	usr.SetDType()

	pb, err := json.Marshal(usr)
	if err != nil {
		return "", fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	req := &api.Request{
		Mutations: []*api.Mutation{{
			SetJson: pb,
		}},
		CommitNow: true,
	}

	resp, dbErr := db.TxWithRetryAndResponse(c, time.Minute*5, req)
	if dbErr != nil {
		return "", fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), dbErr)
	}

	// check if insert was successful
	if len(resp.Uids) != 1 {
		return "", fmt.Errorf("invalid number of uids returned: %s", resp.Uids)
	}

	var userUID string
	for _, v := range resp.GetUids() {
		userUID = v
	}

	return userUID, err
}

// GetUsers gets all users currently in the database
func GetUsers(c external.Database) (users []FrontendUserBackendState, err error) {
	query := `{
				q(func: type(User)){
					uid
					User.email
					User.modified
					User.created
					User.kratosID
					User.roles {
						Role.name
					}
				}
			  }`

	resp, err := db.ReadOnlyTxWithRetry(c, time.Minute*2, query)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	// json struct

	var r struct {
		Users []User `json:"q"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		return
	}

	if len(r.Users) == 0 {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), ErrorUsersNotFound)
		return
	}

	for _, u := range r.Users {
		users = append(users, u.ToFrontendUserBackendState())
	}

	return
}

// GetUsersWithCredentials gets all users currently in the database
func GetUsersWithCredentials(c external.Database) (users []FrontendUserClientStateWithCredentials, err error) {
	query := `{
				q(func: type(User)){
					uid
					User.email
					User.modified
					User.created
					User.kratosID
					User.roles {
						Role.name
					}
					User.pwhash
				}
			  }`

	resp, err := db.ReadOnlyTxWithRetry(c, time.Minute*2, query)
	if err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	// json struct

	var r struct {
		Users []User `json:"q"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		return
	}

	if len(r.Users) == 0 {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), ErrorUsersNotFound)
		return
	}

	for _, u := range r.Users {
		users = append(users, u.ToFrontendUserStateWithCredentials())
	}

	return
}

// existsUser checks if a User with the given uid exists
func existsUser(c external.Database, uid string) (found bool, err error) {
	query := "query Q($uid:string){q(func: uid($uid))@filter(eq(dgraph.type," + DTypeUser + ")){uid}}"

	ctx, cancel := db.GetFrontendContext()
	defer cancel()

	// no retry
	resp, txErr := c.Query(ctx, query, map[string]string{"$uid": uid})
	if txErr != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), txErr)
		return
	}

	var r struct {
		Q []struct {
			UID string
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
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
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), existsErr)
		return
	} else if !found {
		err = errors.New("error user does not exist")
		return
	}

	req := &api.Request{
		Query: "query Q($uid:string){h as var(func: uid($uid))@filter(eq(dgraph.type," + DTypeUser + "))}",
		Vars:  map[string]string{"$uid": uid},
		Mutations: []*api.Mutation{{
			DelNquads: []byte("uid(h) * * ."),
		}},
		CommitNow: true,
	}

	if txErr := db.TxWithRetry(c, time.Minute*5, req); txErr != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), txErr)
		return
	}

	return
}

// CreateDgraphAndKratosUser creates a dgraph user and ory kratos identity.
// The UID of the dgraph user is written the metadata_admin of the new kratos identity.
// The credentials are set if not nil.
func CreateDgraphAndKratosUser(ctx context.Context, c external.Database, adminAuth *ory.APIClient,
	email string, credentials *ory.AdminIdentityImportCredentials, roles []string) error {
	// create dgraph user
	newUserUID, userCreationError := CreateNewUser(c)
	if userCreationError != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), userCreationError)
	}

	// create kratos identity
	_, _, err := adminAuth.V0alpha2Api.AdminCreateIdentity(ctx).
		AdminCreateIdentityBody(ory.AdminCreateIdentityBody{
			SchemaId:       "default_v0",
			Traits:         map[string]interface{}{"email": email},
			MetadataPublic: map[string]any{"roles": roles, "dgraph_uid": newUserUID},
			Credentials:    credentials,
		}).Execute()
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	return nil
}

// CreateKratosUser creates a dgraph user.
// The UID of the dgraph user is written the metadata_admin of the new kratos identity.
// The credentials are set if not nil.
func CreateKratosUser(ctx context.Context, dgraphUID string, adminAuth *ory.APIClient,
	email string, credentials *ory.AdminIdentityImportCredentials, roles []string) error {
	// create kratos identity
	_, _, err := adminAuth.V0alpha2Api.AdminCreateIdentity(ctx).
		AdminCreateIdentityBody(ory.AdminCreateIdentityBody{
			SchemaId:       "default_v0",
			Traits:         map[string]interface{}{"email": email},
			MetadataPublic: map[string]any{"roles": roles, "dgraph_uid": dgraphUID},
			Credentials:    credentials,
		}).Execute()
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	return nil
}

// generateRandomPassword returns a random string if fixed length of 22
func generateRandomPassword() (string, error) {
	// Generate a Salt
	pwByte := make([]byte, 16)
	if _, err := rand.Read(pwByte); err != nil {
		return "", err
	}

	return base64.RawStdEncoding.EncodeToString(pwByte), nil
}

// CreateAdminUser creates a new admin account with a random password
func CreateAdminUser(c external.Database, adminAuth *ory.APIClient, email string) (string, error) {
	pw, err := generateRandomPassword()
	if err != nil {
		return "", fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Minute*2)
	defer cancelFunc()

	// get ory handle and also create password
	err = CreateDgraphAndKratosUser(ctx, c, adminAuth, email, &ory.AdminIdentityImportCredentials{
		Password: &ory.AdminCreateIdentityImportCredentialsPassword{
			Config: &ory.AdminCreateIdentityImportCredentialsPasswordConfig{Password: &pw}},
	}, []string{"admin"})
	if err != nil {
		return "", fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	return pw, nil
}

// CreatePrivilegedUser creates a new privileged user account with the given password
func CreatePrivilegedUser(c external.Database, adminAuth *ory.APIClient, email string, pw string) error {
	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Minute*2)
	defer cancelFunc()

	// get ory handle and also create password
	err := CreateDgraphAndKratosUser(ctx, c, adminAuth, email, &ory.AdminIdentityImportCredentials{
		Password: &ory.AdminCreateIdentityImportCredentialsPassword{
			Config: &ory.AdminCreateIdentityImportCredentialsPasswordConfig{Password: &pw}},
	}, []string{"privileged"})
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	return nil
}
