package db

import (
	"backend/external"
	"context"
	"github.com/dgraph-io/dgo/v210/protos/api"
)

// AlterSchemaAddSpendingGaps adds the spending gap predicate
func AlterSchemaAddSpendingGaps(c external.Database) error {
	return c.Alter(context.Background(), &api.Operation{
		Schema: `
			Heuristic.excludeSpendingGaps: bool .

			type Heuristic {
				Heuristic.type
				Heuristic.parameter
				Heuristic.transaction
				Heuristic.clusters
				Heuristic.ts
				Heuristic.parent
				Heuristic.clusterTypes
				Heuristic.excludeAddresses
				Heuristic.excludeSpendingGaps
			}
		`,
	})
}
