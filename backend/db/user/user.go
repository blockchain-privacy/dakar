package user

import (
	"backend/cmd/cliutil"
	"backend/db"
	"backend/user"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"time"
)

var (
	errorInvalidUser         = errors.New("error invalid user")
	ErrorEmailExists         = errors.New("error email already exists")
	ErrorUsersNotFound       = errors.New("error no users found")
	errorToManyUsersReturned = errors.New("error to many users returned")
)

// CreateUser creates a new user
func CreateUser(c *dgo.Dgraph, user User) error {
	if len(user.Email) == 0 ||
		len(user.Roles) == 0 ||
		len(user.PasswordHash) == 0 ||
		len(user.Roles) > 20 {
		return errorInvalidUser
	}

	user.Uid = ""
	timeNow := time.Now()
	user.Created = &timeNow
	user.Modified = &timeNow
	user.SetDType()

	queryVars := map[string]string{"$email": user.Email}
	queryStart := "query Q($email: string,"
	var queryRoles string
	queryEnd := "user as var(func: eq(user_email, $email))}"
	for i := range user.Roles {
		roleUidPlaceholder := fmt.Sprintf("r%d", i)

		user.Roles[i].Uid = "uid(" + roleUidPlaceholder + ")"
		user.Roles[i].SetDType()

		roleVarId := fmt.Sprintf("$role%d", i)
		queryVars[roleVarId] = user.Roles[i].Name

		queryStart += roleVarId + ":string"
		queryRoles += roleUidPlaceholder + " as var(func: eq(role_name," + roleVarId + "))\n"

		if i+1 < len(user.Roles) {
			queryStart += ","
		}
	}

	queryStart += "){"

	pb, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	req := &api.Request{
		Query: queryStart + queryRoles + queryEnd,
		Vars:  queryVars,
		Mutations: []*api.Mutation{{
			Cond:    "@if(eq(len(user), 0))",
			SetJson: pb,
		}},
		CommitNow: true,
	}

	resp, dbErr := db.TxWithRetryAndResponse(c, time.Minute*5, req)
	if dbErr != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), dbErr)
	}

	// check if insert was successful
	if resp.Uids == nil {
		return ErrorEmailExists
	}

	return err
}

// GetUsers gets all users currently in the database
func GetUsers(c *dgo.Dgraph) (users []User, err error) {
	query := `query {
				q(func: type(User)){
					uid
					user_email
					user_modified
					user_created
					user_roles {
						role_name
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

	users = r.Users

	return
}

// GetUserByEmail gets a User by E-mail from the db
func GetUserByEmail(c *dgo.Dgraph, email string) (user User, err error) {
	query := `query Q($email:string){
				q(func: eq(user_email,$email))@filter(eq(dgraph.type,` + DTypeUser + `)){
					uid
					user_email
					user_pwhash
					user_modified
					user_created
					user_roles{
						role_name
					}
				}
			  }`

	ctx, cancel := db.GetFrontendContext()
	defer cancel()

	// no retry
	resp, txErr := c.NewReadOnlyTxn().QueryWithVars(ctx, query, map[string]string{"$email": email})
	if txErr != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), txErr)
		return
	}

	var r struct {
		Users []User `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if len(r.Users) > 1 {
		err = errorToManyUsersReturned
		return
	} else if len(r.Users) == 0 {
		err = ErrorUsersNotFound
		return
	}

	user = r.Users[0]

	return
}

// GetUser gets a User by uid from the db
func GetUser(c *dgo.Dgraph, uid string) (user User, err error) {
	query := `query Q($uid:string){
				q(func: uid($uid))@filter(eq(dgraph.type,` + DTypeUser + `)){
					uid
					user_email
					user_pwhash
					user_modified
					user_created
					user_roles{
						role_name
					}
				}
			  }`

	ctx, cancel := db.GetFrontendContext()
	defer cancel()

	// no retry
	resp, txErr := c.NewReadOnlyTxn().QueryWithVars(ctx, query, map[string]string{"$uid": uid})
	if txErr != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), txErr)
		return
	}

	var r struct {
		Users []User `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if len(r.Users) > 1 {
		err = errorToManyUsersReturned
		return
	} else if len(r.Users) == 0 {
		err = ErrorUsersNotFound
		return
	}

	user = r.Users[0]

	return
}

// existsUser checks if a User with the given uid exists
func existsUser(c *dgo.Dgraph, uid string) (found bool, err error) {
	query := "query Q($uid:string){q(func: uid($uid))@filter(eq(dgraph.type," + DTypeUser + ")){uid}}"

	ctx, cancel := db.GetFrontendContext()
	defer cancel()

	// no retry
	resp, txErr := c.NewReadOnlyTxn().QueryWithVars(ctx, query, map[string]string{"$uid": uid})
	if txErr != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), txErr)
		return
	}

	var r struct {
		Q []struct {
			Uid string
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
		return
	}

	if len(r.Q) != 1 || len(r.Q[0].Uid) == 0 {
		return
	}

	found = true

	return
}

// DeleteUser deletes the User with the given uid
func DeleteUser(c *dgo.Dgraph, uid string) (err error) {
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

// GetUserCount gets the number of users in the database
func GetUserCount(c *dgo.Dgraph) (uint64, error) {
	return db.GetCount(c, DTypeUser)
}

// GetRoleCount gets the number of roles in the database
func GetRoleCount(c *dgo.Dgraph) (uint64, error) {
	return db.GetCount(c, DTypeRole)
}

// CreateAdminUser creates an new admin account with a random password
func CreateAdminUser(c *dgo.Dgraph, email string) (string, error) {
	pw, pwHash, err := user.GetRandomPasswordAndHash()
	if err != nil {
		return "", err
	}

	if err = CreateUser(c, User{
		Email:        email,
		PasswordHash: pwHash,
		Roles: []Role{{
			Name: "admin",
		}},
	}); err != nil {
		return "", err
	}

	return pw, nil
}

// ModifyUser modifies the given user in the database. The uid must be filled.
// Email and/or Roles can be set.
func ModifyUser(c *dgo.Dgraph, user User) (err error) {

	modifiedTime := time.Now()
	user.Modified = &modifiedTime

	queryStart := "query Q($uid: string"
	var queryRoles string
	queryEnd := "user as var(func: uid($uid))@filter(eq(dgraph.type," + DTypeUser + "))}"
	queryVars := map[string]string{"$uid": user.Uid}

	if len(user.Roles) > 0 {
		queryStart += ","
	}

	for i := range user.Roles {
		roleUidPlaceholder := fmt.Sprintf("r%d", i)

		user.Roles[i].Uid = "uid(" + roleUidPlaceholder + ")"
		user.Roles[i].SetDType()

		roleVarId := fmt.Sprintf("$role%d", i)
		queryVars[roleVarId] = user.Roles[i].Name

		queryStart += roleVarId + ":string"
		queryRoles += roleUidPlaceholder + " as var(func: eq(role_name," + roleVarId + "))\n"

		if i+1 < len(user.Roles) {
			queryStart += ","
		}
	}

	queryStart += "){"

	pb, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), err)
	}

	ctx, cancel := db.GetFrontendContext()
	defer cancel()

	resp, txErr := c.NewTxn().Do(ctx, &api.Request{
		Query: queryStart + queryRoles + queryEnd,
		Vars:  queryVars,
		Mutations: []*api.Mutation{{
			Cond:    "@if(eq(len(user), 1))",
			SetJson: pb,
		}},
		CommitNow: true,
	})
	_ = resp
	// no retry
	if txErr != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), txErr)
		return
	}

	// check if mutation was successful
	if numMutations, ok := resp.Metrics.NumUids["mutation_cost"]; !ok || numMutations == 0 {
		err = errors.New("error user was not modified")
		return
	}

	return
}

// RemoveRolesFromUser removes all roles from a given user
func RemoveRolesFromUser(c *dgo.Dgraph, uid string) (err error) {
	query := `query Q($uid: string) {
			 h as var(func: uid($uid))@filter(eq(dgraph.type,"` + DTypeUser + `"))
	}`

	req := &api.Request{
		Query: query,
		Vars:  map[string]string{"$uid": uid},
		Mutations: []*api.Mutation{{
			DelNquads: []byte("uid(h) <user_roles> * ."),
		}},
		CommitNow: true,
	}

	if txErr := db.TxWithRetry(c, time.Minute*5, req); txErr != nil {
		err = fmt.Errorf("%s: %w", cliutil.ShowCallInfo(), txErr)
		return
	}
	return
}
