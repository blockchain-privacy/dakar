package user

import (
	"backend/cmd/cliutil"
	"backend/db"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgraph-io/dgo/v2"
	"github.com/dgraph-io/dgo/v2/protos/api"
	"time"
)

var (
	errorInvalidUser   = errors.New("error invalid user")
	ErrorEmailExists   = errors.New("error email already exists")
	ErrorUsersNotFound = errors.New("error no users found")
)

// CreateUser creates a new user
func CreateUser(c *dgo.Dgraph, user User) error {
	if len(user.Email) == 0 || len(user.Roles) == 0 {
		return errorInvalidUser
	}

	user.Uid = ""
	timeNow := time.Now()
	user.Created = timeNow
	user.Modified = timeNow
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
	ctx, cancel := db.GetBackendContext()
	defer cancel()
	resp, dbErr := db.TxWithRetryAndResponse(c, ctx, req)
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
					user_email
					user_modified
					user_created
					user_roles {
						role_name
					}
				}
			  }`

	ctx, cancel := db.GetFrontendContext()
	defer cancel()
	resp, err := db.ReadOnlyTxWithRetry(c, ctx, query)
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

// GetUserCount gets the number of users in the database
func GetUserCount(c *dgo.Dgraph) (uint64, error) {
	return db.GetCount(c, DTypeUser)
}

// GetRoleCount gets the number of roles in the database
func GetRoleCount(c *dgo.Dgraph) (uint64, error) {
	return db.GetCount(c, DTypeRole)
}
