package db

import (
	"backend/external"
	"context"
	"github.com/dgraph-io/dgo/v210/protos/api"
)

// AlterSchemaAddKratosID adds the kratosID predicate
func AlterSchemaAddKratosID(c external.Database) error {
	return c.Alter(context.Background(), &api.Operation{
		Schema: `
			User.kratosID: string @index(hash) .
	
			type User {
				User.email
				User.pwhash
				User.roles
				User.created
				User.modified
				User.heuristics
				User.addressExclusions
				User.kratosID
			}
		`,
	})
}
