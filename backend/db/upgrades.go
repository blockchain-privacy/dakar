package db

import (
	"backend/external"
	"context"
	"github.com/dgraph-io/dgo/v230/protos/api"
)

// AlterSchemaAddWorkspaces adds the workspace type
func AlterSchemaAddWorkspaces(c external.Database) error {
	return c.Alter(context.Background(), &api.Operation{
		Schema: `
			User.heuristics: [uid] @reverse .
			User.addressExclusions: [uid] @count @reverse .
			User.workspaces: [uid] @reverse .

			type User {
				User.heuristics
				User.addressExclusions
				User.workspaces
			}
			
			Workspace.name: string . # the workspace name
			Workspace.ts: dateTime @index(day) . # modification date of the workspace
			Workspace.state: string . # JSON encoded state of the workspace
			Workspace.clusterHeight: int . # last clustered block at which this workspace was updated
			Workspace.heuristics: [uid] @reverse . # heuristics which are managed by this workspace

			type Workspace {
				Workspace.name
				Workspace.ts
				Workspace.state
				Workspace.clusterHeight
				Workspace.heuristics
			}`,
	})
}
