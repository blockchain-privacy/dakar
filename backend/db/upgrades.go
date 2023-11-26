package db

import (
	"backend/cmd/cliutil"
	"backend/external"
	"context"
	"github.com/dgraph-io/dgo/v230/protos/api"
)

// AlterSchemaDropRoles removes the Role type
func AlterSchemaDropRoles(c external.Database) error {
	err := c.Alter(context.Background(), &api.Operation{
		DropAttr: "Role.name",
	})
	if err != nil {
		return cliutil.NewStackError(err)
	}

	return c.Alter(context.Background(), &api.Operation{
		DropOp:    api.Operation_TYPE,
		DropValue: "Role",
	})
}

// AlterSchemaDropUserPredicates removes all not needed User predicates
func AlterSchemaDropUserPredicates(c external.Database) error {
	err := c.Alter(context.Background(), &api.Operation{
		DropAttr: "User.email",
	})
	if err != nil {
		return cliutil.NewStackError(err)
	}

	err = c.Alter(context.Background(), &api.Operation{
		DropAttr: "User.pwhash",
	})
	if err != nil {
		return cliutil.NewStackError(err)
	}

	err = c.Alter(context.Background(), &api.Operation{
		DropAttr: "User.created",
	})
	if err != nil {
		return cliutil.NewStackError(err)
	}

	err = c.Alter(context.Background(), &api.Operation{
		DropAttr: "User.modified",
	})
	if err != nil {
		return cliutil.NewStackError(err)
	}

	err = c.Alter(context.Background(), &api.Operation{
		DropAttr: "User.roles",
	})
	if err != nil {
		return cliutil.NewStackError(err)
	}

	return c.Alter(context.Background(), &api.Operation{
		Schema: `
			User.heuristics: [uid] @reverse .
			User.addressExclusions: [uid] @count @reverse .
	
			type User {
				User.heuristics
				User.addressExclusions
			}`,
	})
}
