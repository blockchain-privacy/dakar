// SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package workspace

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/dgraph-io/dgo/v250/protos/api"
	"gitlab.com/blockchain-privacy/dakar/db"
	"gitlab.com/blockchain-privacy/dakar/external"
	"gitlab.com/blockchain-privacy/gomisc/serror"
)

// AddWorkspace creates a new workspace
func AddWorkspace(ctx context.Context, c external.Database, name string, userUID string) (workspaceUID string, err error) {
	if name == "" || userUID == "" {
		err = serror.New(db.ErrEmptyRequestArgument)
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
		err = serror.New(err)
		return
	}

	resp, err := c.Mutate(ctx, &api.Request{
		Query: `query Q($userUID: string) {
					u as var(func: uid($userUID))@filter(type(User))
				  }`,
		Vars: map[string]string{"$userUID": userUID},
		Mutations: []*api.Mutation{{
			Cond:    "@if(eq(len(u), 1))",
			SetJson: pb,
		}},
		CommitNow: true,
	})
	if err != nil {
		err = serror.New(err)
		return
	}

	workspaceUID, ok := resp.GetUids()[newWorkspaceDummyUID]
	if !ok {
		err = serror.FromStr("new workspace was not created")
		return
	}

	return
}

// RenameWorkspace renames a workspace
func RenameWorkspace(ctx context.Context, c external.Database, name string,
	userUID string, workspaceUID string) (err error) {
	if name == "" || userUID == "" {
		return serror.New(db.ErrEmptyRequestArgument)
	}
	w := Workspace{
		UID:              workspaceUID,
		Name:             name,
		ModificationTime: time.Now().UTC().Format(time.RFC3339),
		ClusterHeight:    nil, // unset
	}
	w.SetDType()

	pb, err := json.Marshal(dummyUser{UID: userUID, Workspaces: []Workspace{w}})
	if err != nil {
		err = serror.New(err)
		return
	}

	_, err = c.Mutate(ctx, &api.Request{Mutations: []*api.Mutation{{SetJson: pb}}, CommitNow: true})

	return
}

// SetWorkspaceState sets the state of the specified workspace
func SetWorkspaceState(ctx context.Context, c external.Database, userUID string, workspaceUID string,
	state string, clusterHeight *int64) (err error) {
	if workspaceUID == "" || userUID == "" || state == "" {
		return serror.New(db.ErrEmptyRequestArgument)
	}
	w := Workspace{
		UID:              "uid(v)",
		State:            state,
		ModificationTime: time.Now().UTC().Format(time.RFC3339),
		ClusterHeight:    clusterHeight,
	}
	w.SetDType()

	pb, err := json.Marshal(dummyUser{UID: userUID, Workspaces: []Workspace{w}})
	if err != nil {
		err = serror.New(err)
		return
	}

	return db.MutationWithRetry(ctx, c, &api.Request{
		Query: "query Q($uid:string){var(func: uid($uid))@filter(has(Workspace.name)){v as uid}}",
		Vars:  map[string]string{"$uid": workspaceUID},
		Mutations: []*api.Mutation{{
			Cond:    "@if(gt(len(v), 0))",
			SetJson: pb,
		}},
		CommitNow: true,
	})
}

// GetFrontendWorkspaces returns all workspaces of the current user without its state
func GetFrontendWorkspaces(ctx context.Context, c external.Database, userUID string) ([]Workspace, error) {
	if userUID == "" {
		return nil, serror.New(db.ErrEmptyRequestArgument)
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

	resp, err := c.Query(ctx, query, map[string]string{"$user": userUID})
	if err != nil {
		return nil, serror.New(err)
	}

	// json struct
	var r struct {
		Workspaces []Workspace `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		return nil, serror.New(err)
	}

	return r.Workspaces, nil
}

// isStateEmpty returns true if the given string does not represent an empty state
func isStateEmpty(state string) bool {
	return state == "" || state == "[]" || state == "{}"
}

// GetFrontendWorkspace returns the specified workspace
func GetFrontendWorkspace(ctx context.Context, c external.Database, uid string, userUID string) (*DecodedWorkspace, error) {
	if userUID == "" || uid == "" {
		return nil, serror.New(db.ErrEmptyRequestArgument)
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

	resp, err := c.Query(ctx, query, map[string]string{"$user": userUID, "$workspace": uid})
	if err != nil {
		return nil, serror.New(err)
	}

	// json struct
	var r struct {
		Workspaces []Workspace `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		return nil, serror.New(err)
	}

	if len(r.Workspaces) != 1 {
		return nil, serror.FromStr("invalid number of workspaces returned: " + strconv.Itoa(len(r.Workspaces)))
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
		return nil, serror.New(err)
	}

	return &decodedWorkspace, nil
}

// GetWorkspaceState returns the state of the specified workspace
func GetWorkspaceState(ctx context.Context, c external.Database, uid string, userUID string) (string, error) {
	if userUID == "" || uid == "" {
		return "", serror.New(db.ErrEmptyRequestArgument)
	}

	// request uid and state, so if the state is empty a workspace item is still returned and passed the check below
	query := `query Q($user:string,$workspace:string){
			var(func: uid($user)){
				w as User.workspaces@filter(uid($workspace))
			}

			q(func: uid(w)){
				uid
				Workspace.state
			}
		}`

	resp, err := c.Query(ctx, query, map[string]string{"$user": userUID, "$workspace": uid})
	if err != nil {
		return "", serror.New(err)
	}

	// json struct
	var r struct {
		Workspaces []Workspace `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		return "", serror.New(err)
	}

	if len(r.Workspaces) != 1 {
		return "", serror.FromStr("invalid number of workspaces returned: " + strconv.Itoa(len(r.Workspaces)))
	}

	return r.Workspaces[0].State, nil
}

// DeleteWorkspace deletes a user's workspace including all their selectors.
// If the workspace UID is not set, all workspaces of the user are deleted.
func DeleteWorkspace(ctx context.Context, c external.Database, userUID string, workspaceUID string) error {
	var filterWorkspaces string

	if workspaceUID != "" {
		filterWorkspaces = "@filter(uid($workspace))"
	}

	req := &api.Request{
		Query: `query Q($user:string, $workspace:string){
				var(func: uid($user)){
					w as User.workspaces` + filterWorkspaces + `{
						s as Workspace.selectors{
							hc as Selector.results@filter(has(HeuristicCluster.results))
						}
					}
				}
			  }`,
		Vars: map[string]string{"$user": userUID, "$workspace": workspaceUID},
		Mutations: []*api.Mutation{{
			DelNquads: []byte(` uid(hc) * * .
								uid(s) * * .
								uid(w) * * .
								<` + userUID + "> <User.workspaces> uid(w) ."),
		}},
		CommitNow: true,
	}

	return db.MutationWithRetry(ctx, c, req)
}

func IsWorkspaceStateOutdated(ctx context.Context, c external.Database,
	height int64, nodeUIDs []string) (isOutdated bool, err error) {
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

	resp, err := db.QueryVarWithRetry(ctx, c, query, map[string]string{"$uids": db.CreateCommaArray(nodeUIDs)})
	if err != nil {
		err = serror.New(err)
		return
	}

	// json struct
	var r struct {
		Height []struct {
			MaxHeight *int64 `json:"max_height,omitempty"`
		} `json:"q,omitempty"`
	}

	if err = json.Unmarshal(resp.GetJson(), &r); err != nil {
		err = serror.New(err)
		return
	}

	if len(r.Height) != 1 || r.Height[0].MaxHeight == nil {
		err = serror.FromFormat("invalid max height returned: %v", r.Height)
		return
	}

	isOutdated = height < *r.Height[0].MaxHeight

	return
}
