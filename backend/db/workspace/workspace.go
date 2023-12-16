package workspace

import (
	"backend/cmd/cliutil"
	"backend/db"
	"backend/external"
	"encoding/json"
	"github.com/dgraph-io/dgo/v230/protos/api"
	"strconv"
	"time"
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
		err = cliutil.NewStackErrorStr("no new heuristic created")
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
