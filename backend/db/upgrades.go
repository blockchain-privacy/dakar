package db

import (
	"backend/external"
	"context"
	"github.com/dgraph-io/dgo/v210/protos/api"
)

// AlterSchemaAddMeta adds the new meta type
func AlterSchemaAddMeta(c external.Database) error {
	return c.Alter(context.Background(), &api.Operation{
		Schema: `
			Meta.creationTime: dateTime . # The time when this database was initialized
			Meta.blockchainMode: string . # The blockchain mode of this database. Possible values: Bitcoin, Dash, Doge
			Meta.schemaVersion: int . # The schema version of this database. 

			type Meta {
				Meta.creationTime
				Meta.blockchainMode
				Meta.schemaVersion
			}
		`,
	})
}
