package workspace

import (
	"backend/db"
	"backend/external"
	"encoding/json"
	"errors"
	"github.com/dgraph-io/dgo/v230/protos/api"
	"github.com/qrest/gomisc/serror"
	"strconv"
	"time"
)

var (
	// ErrNoMutationHappened is returned if no mutation occurred
	ErrNoMutationHappened = errors.New("no mutation happened")
)

// AddWorkspace creates a new workspace
func AddWorkspace(c external.Database, name string, userUID string) (workspaceUID string, err error) {
	if name == "" || userUID == "" {
		err = serror.NewStackError(db.ErrEmptyRequestArgument)
		return
	}
	const newWorkspaceDummyUID = "new_w"
	w := Workspace{
		UID:              "_:" + newWorkspaceDummyUID,
		Name:             name,
		ModificationTime: time.Now().UTC().Format(time.RFC3339),
		ClusterHeight:    nil, // unset
	}
	w.SetDType()

	type dummyUser struct {
		UID        string      `json:"uid,omitempty"`
		Workspaces []Workspace `json:"User.workspaces,omitempty"`
	}

	pb, err := json.Marshal(dummyUser{UID: userUID, Workspaces: []Workspace{w}})
	if err != nil {
		err = serror.NewStackError(err)
		return
	}

	resp, err := db.TxWithRetryAndResponse(c, time.Minute*10, &api.Request{
		Mutations: []*api.Mutation{{SetJson: pb}},
		CommitNow: true,
	})
	if err != nil {
		return
	}

	workspaceUID, ok := resp.GetUids()[newWorkspaceDummyUID]
	if !ok {
		err = serror.NewStackErrorStr("new workspace was not created")
		return
	}

	return
}

// RenameWorkspace renames a workspace
func RenameWorkspace(c external.Database, name string, userUID string, workspaceUID string) (err error) {
	if name == "" || userUID == "" {
		return serror.NewStackError(db.ErrEmptyRequestArgument)
	}
	w := Workspace{
		UID:              workspaceUID,
		Name:             name,
		ModificationTime: time.Now().UTC().Format(time.RFC3339),
		ClusterHeight:    nil, // unset
	}
	w.SetDType()

	type dummyUser struct {
		UID        string      `json:"uid,omitempty"`
		Workspaces []Workspace `json:"User.workspaces,omitempty"`
	}

	pb, err := json.Marshal(dummyUser{UID: userUID, Workspaces: []Workspace{w}})
	if err != nil {
		err = serror.NewStackError(err)
		return
	}

	_, err = db.TxWithRetryAndResponse(c, time.Minute*10, &api.Request{
		Mutations: []*api.Mutation{{SetJson: pb}},
		CommitNow: true,
	})

	return
}

// SetWorkspaceState sets the state of the specified workspace
func SetWorkspaceState(c external.Database, userUID string, workspaceUID string,
	state string, clusterHeight *int64) (err error) {
	if workspaceUID == "" || userUID == "" || state == "" {
		return serror.NewStackError(db.ErrEmptyRequestArgument)
	}
	w := Workspace{
		UID:              "uid(v)",
		State:            state,
		ModificationTime: time.Now().UTC().Format(time.RFC3339),
		ClusterHeight:    clusterHeight,
	}
	w.SetDType()

	type dummyUser struct {
		UID        string      `json:"uid,omitempty"`
		Workspaces []Workspace `json:"User.workspaces,omitempty"`
	}

	pb, err := json.Marshal(dummyUser{UID: userUID, Workspaces: []Workspace{w}})
	if err != nil {
		err = serror.NewStackError(err)
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
		return nil, serror.NewStackError(db.ErrEmptyRequestArgument)
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
		return nil, serror.NewStackError(err)
	}

	// json struct
	var r struct {
		Workspaces []Workspace `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		return nil, serror.NewStackError(err)
	}

	return r.Workspaces, nil
}

// isStateEmpty returns true if the given string does not represent an empty state
func isStateEmpty(state string) bool {
	return state == "" || state == "[]" || state == "{}"
}

// GetFrontendWorkspace returns the specified workspace
func GetFrontendWorkspace(c external.Database, uid string, userUID string) (*DecodedWorkspace, error) {
	if userUID == "" {
		return nil, serror.NewStackError(db.ErrEmptyRequestArgument)
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
				Workspace.clusterHeight
			}
		}`

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*2, query, map[string]string{"$user": userUID, "$workspace": uid})
	if err != nil {
		return nil, serror.NewStackError(err)
	}

	// json struct
	var r struct {
		Workspaces []Workspace `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		return nil, serror.NewStackError(err)
	}

	if len(r.Workspaces) != 1 {
		return nil, serror.NewStackErrorStr("invalid number of workspaces returned: " + strconv.Itoa(len(r.Workspaces)))
	}

	decodedWorkspace := DecodedWorkspace{
		UID:              r.Workspaces[0].UID,
		Name:             r.Workspaces[0].Name,
		ModificationTime: r.Workspaces[0].ModificationTime,
		ClusterHeight:    r.Workspaces[0].ClusterHeight,
	}

	if isStateEmpty(r.Workspaces[0].State) {
		return &decodedWorkspace, nil
	}

	if err := json.Unmarshal([]byte(r.Workspaces[0].State), &decodedWorkspace.Nodes); err != nil {
		return nil, serror.NewStackError(err)
	}

	return &decodedWorkspace, nil
}

// DeleteAllWorkspaces deletes a user's workspaces and their heuristics
func DeleteAllWorkspaces(c external.Database, userUID string) error {
	return DeleteWorkspace(c, userUID, "")
}

// DeleteWorkspace deletes a user's workspace
func DeleteWorkspace(c external.Database, userUID string, workspaceUID string) error {
	var filterWorkspaces string

	if workspaceUID != "" {
		filterWorkspaces = "@filter(uid($workspace))"
	}

	req := &api.Request{
		Query: `query Q($user:string, $workspace:string){
				var(func: uid($user)){
					w as User.workspaces` + filterWorkspaces + `{
						h as Workspace.heuristics{
							hc as Heuristic.clusters{
								hr as HeuristicCluster.results
							}
						}
					}
				}
			  }`,
		Vars: map[string]string{"$user": userUID, "$workspace": workspaceUID},
		Mutations: []*api.Mutation{{
			DelNquads: []byte(` uid(hr) * * .
								uid(hc) * * .
								uid(h) * * .
								uid(w) * * .
								<` + userUID + "> <User.workspaces> uid(w) ."),
		}},
		CommitNow: true,
	}

	_, err := db.TxWithRetryAndResponse(c, time.Minute*5, req)
	return err
}

func IsWorkspaceStateOutdated(c external.Database, height int64, nodeUIDs []string) (isOutdated bool, err error) {
	const query = `query Q($uids:string){
					var(func: uid($uids)){
						 ~Cluster.addresses@filter(eq(Cluster.type, "fmi")){
							Cluster.transaction{
								~transactions{
									h as id
								}
							}
						}
					}

					q(){
						max_height:max(val(h))
					}
				}`

	resp, err := db.ReadOnlyTxVarWithRetry(c, time.Minute*2, query,
		map[string]string{"$uids": db.CreateCommaArray(nodeUIDs)})
	if err != nil {
		err = serror.NewStackError(err)
		return
	}

	// json struct
	var r struct {
		Height []struct {
			MaxHeight *int64 `json:"max_height,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.Json, &r); err != nil {
		err = serror.NewStackError(err)
		return
	}

	if len(r.Height) != 1 || r.Height[0].MaxHeight == nil {
		err = serror.NewStackErrorf("invalid max height returned: %v", r.Height)
		return
	}

	isOutdated = height < *r.Height[0].MaxHeight

	return
}
