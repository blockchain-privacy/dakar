package workspace

import (
	"backend/cmd/cliutil"
	"backend/db"
	"backend/external"
	"encoding/json"
	"errors"
	"github.com/dgraph-io/dgo/v230/protos/api"
	"strconv"
	"time"
)

var (
	// ErrNoMutationHappened is returned if no mutation occurred
	ErrNoMutationHappened = errors.New("no mutation happened")
)

// AddWorkspace creates a new workspace with the given name
func AddWorkspace(c external.Database, name string, userUID string) (err error) {
	if name == "" || userUID == "" {
		return cliutil.NewStackError(db.ErrEmptyRequestArgument)
	}
	const newWorkspaceDummyUID = "new_w"
	w := Workspace{
		UID:              "_:" + newWorkspaceDummyUID,
		Name:             name,
		ModificationTime: time.Now().UTC().Format(time.RFC3339),
	}
	w.SetDType()

	type dummyUser struct {
		UID        string      `json:"uid,omitempty"`
		Workspaces []Workspace `json:"User.workspaces,omitempty"`
	}

	pb, err := json.Marshal(dummyUser{UID: userUID, Workspaces: []Workspace{w}})
	if err != nil {
		err = cliutil.NewStackError(err)
		return
	}

	resp, err := db.TxWithRetryAndResponse(c, time.Minute*10, &api.Request{
		Mutations: []*api.Mutation{{
			SetJson: pb,
		}},
		CommitNow: true,
	})
	if err != nil {
		return
	}

	_, ok := resp.GetUids()[newWorkspaceDummyUID]
	if !ok {
		err = cliutil.NewStackErrorStr("new workspace was not created")
		return
	}

	return
}

// SetWorkspaceState sets the state of the specified workspace
func SetWorkspaceState(c external.Database, userUID string, workspaceUID string,
	state string, timeStamp time.Time) (err error) {
	if workspaceUID == "" || userUID == "" || state == "" {
		return cliutil.NewStackError(db.ErrEmptyRequestArgument)
	}
	w := Workspace{
		UID:              "uid(v)",
		State:            state,
		ModificationTime: timeStamp.UTC().Format(time.RFC3339),
	}
	w.SetDType()

	type dummyUser struct {
		UID        string      `json:"uid,omitempty"`
		Workspaces []Workspace `json:"User.workspaces,omitempty"`
	}

	pb, err := json.Marshal(dummyUser{UID: userUID, Workspaces: []Workspace{w}})
	if err != nil {
		err = cliutil.NewStackError(err)
		return
	}

	_, err = db.TxWithRetryAndResponse(c, time.Minute*10, &api.Request{
		Query: "query Q($uid:string){var(func: uid($uid))@filter(type(Workspace)){v as uid}}",
		Vars:  map[string]string{"$uid": workspaceUID},
		Mutations: []*api.Mutation{{
			Cond:    "@if(gt(len(v), 0))",
			SetJson: pb,
		}},
		CommitNow: true,
	})
	if err != nil {
		return
	}

	return
}

// GetFrontendWorkspaces returns all workspaces of the current user without its state
func GetFrontendWorkspaces(c external.Database, userUID string) ([]Workspace, error) {
	if userUID == "" {
		return nil, cliutil.NewStackError(db.ErrEmptyRequestArgument)
	}

	query := `query Q($user:string){
			var(func: uid($user)){
				w as User.workspaces
			}

			q(func: uid(w)){
				uid
				Workspace.name
				Workspace.ts
			}
		}`

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*2, query, map[string]string{"$user": userUID})
	if err != nil {
		return nil, cliutil.NewStackError(err)
	}

	// json struct
	var r struct {
		Workspaces []Workspace `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		return nil, cliutil.NewStackError(err)
	}

	return r.Workspaces, nil
}

// GetFrontendWorkspace returns the specified workspace
func GetFrontendWorkspace(c external.Database, uid string, userUID string) (*Workspace, error) {
	if userUID == "" {
		return nil, cliutil.NewStackError(db.ErrEmptyRequestArgument)
	}

	query := `query Q($user:string,$workspace:string){
			var(func: uid($user)){
				w as User.workspaces@filter(uid($workspace))
			}

			q(func: uid(w)){
				uid
				Workspace.name
				Workspace.ts
				Workspace.state
			}
		}`

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*2, query, map[string]string{"$user": userUID, "$workspace": uid})
	if err != nil {
		return nil, cliutil.NewStackError(err)
	}

	// json struct
	var r struct {
		Workspaces []Workspace `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		return nil, cliutil.NewStackError(err)
	}

	if len(r.Workspaces) != 1 {
		return nil, cliutil.NewStackErrorStr("invalid number of workspaces returned: " + strconv.Itoa(len(r.Workspaces)))
	}

	return &r.Workspaces[0], nil
}

// DeleteAllWorkspaces deletes all workspaces of a user
func DeleteAllWorkspaces(c external.Database, userUID string) error {
	req := &api.Request{
		Query: "query Q($user:string){var(func: uid($user)){w as User.workspaces}}}",
		Vars:  map[string]string{"$user": userUID},
		Mutations: []*api.Mutation{{
			DelNquads: []byte(` uid(w) * * .
								<` + userUID + "> <User.workspaces> uid(w) ."),
		}},
		CommitNow: true,
	}

	_, err := db.TxWithRetryAndResponse(c, time.Minute*10, req)
	return err
}

// DeleteWorkspace deletes a user's workspace
func DeleteWorkspace(c external.Database, workspaceUID string, userUID string) error {
	query := `query Q($user:string, $workspace:string){
				var(func: uid($user)){
					w as User.workspaces@filter(uid($workspace))
				}
			  }`

	req := &api.Request{
		Query: query,
		Vars:  map[string]string{"$user": userUID, "$workspace": workspaceUID},
		Mutations: []*api.Mutation{{
			DelNquads: []byte(` uid(w) * * .
								<` + userUID + "> <User.workspaces> uid(w) ."),
		}},
		CommitNow: true,
	}

	resp, err := db.TxWithRetryAndResponse(c, time.Minute*5, req)
	if err != nil {
		return err
	}

	if v, ok := resp.Metrics.NumUids["mutation_cost"]; !ok || v == 0 {
		return cliutil.NewStackError(ErrNoMutationHappened)
	}

	return nil
}
