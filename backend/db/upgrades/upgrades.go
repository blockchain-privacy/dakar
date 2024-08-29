package upgrades

import (
	"backend/db"
	"backend/db/analytics/clustering"
	"backend/db/analytics/heuristics"
	"backend/db/status"
	"backend/external"
	"context"
	"github.com/dgraph-io/dgo/v230/protos/api"
	"github.com/qrest/gomisc/serror"
	"log/slog"
	"reflect"
	"runtime"
)

// availableUpgrades contains all available schema upgrades.
// The key is the schema version to which the database should
// be set after its updates haven been applied.
var availableUpgrades = map[uint64]UpgradePackage{
	4: {upgrades: []schemaUpgrade{AlterSchemaAddWorkspaces}},
	5: {
		upgrades: []schemaUpgrade{
			clustering.DeleteAllFMIClusters,
			func() schemaUpgrade {
				return func(c external.Database) error {
					zero := uint64(0)
					return status.SetClusteringFMIStatus(c, status.ClusteringFlatMultiInputStatus{
						LastClusteredBlockID: &zero,
					})
				}
			}(),
			DropPredicateHex,
			AlterSchemaRemoveHex},
	},
	6:  {upgrades: []schemaUpgrade{DropPredicateWorkspaceHeuristics, AlterSchemaAddSelectors}},
	7:  {upgrades: []schemaUpgrade{DropPredicateUserHeuristics, AlterSchemaRemoveUserHeuristics, heuristics.DeleteAllHeuristics}},
	8:  {upgrades: []schemaUpgrade{DropTypeHeuristic, DropTypeHeuristicResult}},
	9:  {upgrades: []schemaUpgrade{AlterSchemaAddSelectorTotalResultCount}},
	10: {upgrades: []schemaUpgrade{DropPrivacyType, AlterSchemaAddTransactionType}},
}

var thisLogger *slog.Logger

// InitLogger creates new loggers with the given parameters.
func InitLogger() {
	thisLogger = slog.With(slog.String("module", "database upgrade"))
}

func info(msg string, v ...any) {
	thisLogger.Info(msg, v...)
}

func getFunctionName(i interface{}) string {
	f := runtime.FuncForPC(reflect.ValueOf(i).Pointer())
	if f == nil {
		return "<unknown function name>"
	}
	return f.Name()
}

// GetSchemaVersion returns the schema version of the database
func GetSchemaVersion(db external.Database) (uint64, error) {
	meta, err := status.GetMeta(db)
	if err != nil {
		return 0, err
	}

	if meta.SchemaVersion == nil {
		return 0, serror.FromStr("received nil schema version")
	}

	return *meta.SchemaVersion, nil
}

type schemaUpgrade func(database external.Database) error

type UpgradePackage struct {
	upgrades []schemaUpgrade
}

// upgradeDatabaseToNextVersion upgrades the database to the next schema version
func upgradeDatabaseToNextVersion(c external.Database, upgrades map[uint64]UpgradePackage, currentSchemaVersion uint64) error {
	upgradePackage, ok := upgrades[currentSchemaVersion+1]
	if !ok {
		return serror.FromStrWithContext("can not find upgrade package",
			"current database version", currentSchemaVersion)
	}

	if len(upgradePackage.upgrades) == 0 {
		return serror.FromStr("upgrade package contains no upgrades")
	}

	for _, upgrade := range upgradePackage.upgrades {
		info("applyling upgrade", "function name", getFunctionName(upgrade))
		if err := upgrade(c); err != nil {
			return err
		}
	}

	if err := status.SetSchemaVersion(c, currentSchemaVersion+1); err != nil {
		return err
	}

	return nil
}

// UpgradeDatabase upgrades the database schema to the newest version
func UpgradeDatabase(c external.Database) error {
	return applyUpgrades(c, availableUpgrades)
}

// applyUpgrades upgrades the database schema to the newest version, by applying  the given UpgradePackages
func applyUpgrades(c external.Database, upgrades map[uint64]UpgradePackage) error {
	currentSchemaVersion, err := GetSchemaVersion(c)
	if err != nil {
		return err
	}

	if db.SchemaVersion < currentSchemaVersion {
		return serror.FromStrWithContext("invalid schema version",
			"executable schema version", db.SchemaVersion, "database schema version", currentSchemaVersion)
	}

	if db.SchemaVersion == currentSchemaVersion {
		info("database schema is already up to date. No upgrades a necessary.")
		return nil
	}

	info("starting upgrade process", "current version", currentSchemaVersion, "target version", db.SchemaVersion)

	for currentSchemaVersion < db.SchemaVersion {
		info("upgrading database schema", "current version", currentSchemaVersion)
		if err := upgradeDatabaseToNextVersion(c, upgrades, currentSchemaVersion); err != nil {
			return err
		}

		newVersion, err := GetSchemaVersion(c)
		if err != nil {
			return err
		}

		if newVersion != currentSchemaVersion+1 {
			return serror.FromStr("database schema upgrade did not increase version")
		}

		currentSchemaVersion = newVersion
		info("database schema upgrade complete", "current version", currentSchemaVersion)
	}

	return nil
}

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

// DropPredicateHex drops the sighex and keyhex predicate
func DropPredicateHex(c external.Database) error {
	err := c.Alter(context.Background(), &api.Operation{DropAttr: "sighex"})
	if err != nil {
		return err
	}

	return c.Alter(context.Background(), &api.Operation{DropAttr: "keyhex"})
}

// AlterSchemaRemoveHex removes hex signature and script from Output type
func AlterSchemaRemoveHex(c external.Database) error {
	return c.Alter(context.Background(), &api.Operation{
		Schema: `
			outputindex: int .
			inputindex: int .
			txtype: string .
			amount: int .
			iscoinbase: bool .
			keyasm: string .
			sigasm: string .

			type Output {
				outputindex
				inputindex
				txtype
				amount
				iscoinbase
				keyasm
				sigasm
				<~tx_inputs>
				<~tx_outputs>
				<~addr_outputs>
			}`,
	})
}

func DropPredicateWorkspaceHeuristics(c external.Database) error {
	return c.Alter(context.Background(), &api.Operation{DropAttr: "Workspace.heuristics"})
}

func AlterSchemaAddSelectors(c external.Database) error {
	return c.Alter(context.Background(), &api.Operation{
		Schema: `
			Workspace.selectors: [uid] @reverse . # selectors which are managed by this workspace
			type Workspace {
				Workspace.name
				Workspace.ts
				Workspace.state
				Workspace.clusterHeight
				Workspace.selectors
			}
			
			Selector.created: dateTime @index(day) .  # creation date of the selector
			Selector.modified: dateTime @index(day) .  # modification date of the selector
			Selector.type: string @index(hash) . # type of the selector
			Selector.status: string @index(hash) . # status of the selector (waiting, error, success)
			Selector.parent: uid @reverse . # parent node from which a selector can use data
			Selector.options: string . # JSON encoded options of the selector
			Selector.results: [uid] @reverse . # results of the selector

			type Selector {
				Selector.created
				Selector.modified
				Selector.type
				Selector.status
				Selector.parent
				Selector.options
				Selector.results
			}`,
	})
}

func DropPredicateUserHeuristics(c external.Database) error {
	return c.Alter(context.Background(), &api.Operation{DropAttr: "User.heuristics"})
}

func AlterSchemaRemoveUserHeuristics(c external.Database) error {
	return c.Alter(context.Background(), &api.Operation{
		Schema: `
			User.addressExclusions: [uid] @count @reverse .
			User.workspaces: [uid] @reverse .

			type User {
				User.addressExclusions
				User.workspaces
			}`,
	})
}

func DropTypeHeuristic(c external.Database) error {
	if err := c.Alter(context.Background(), &api.Operation{DropAttr: "Heuristic.type"}); err != nil {
		return err
	}
	if err := c.Alter(context.Background(), &api.Operation{DropAttr: "Heuristic.parameter"}); err != nil {
		return err
	}
	if err := c.Alter(context.Background(), &api.Operation{DropAttr: "Heuristic.transaction"}); err != nil {
		return err
	}
	if err := c.Alter(context.Background(), &api.Operation{DropAttr: "Heuristic.clusters"}); err != nil {
		return err
	}
	if err := c.Alter(context.Background(), &api.Operation{DropAttr: "Heuristic.parent"}); err != nil {
		return err
	}
	if err := c.Alter(context.Background(), &api.Operation{DropAttr: "Heuristic.ts"}); err != nil {
		return err
	}
	if err := c.Alter(context.Background(), &api.Operation{DropAttr: "Heuristic.clusterTypes"}); err != nil {
		return err
	}
	if err := c.Alter(context.Background(), &api.Operation{DropAttr: "Heuristic.excludeAddresses"}); err != nil {
		return err
	}
	if err := c.Alter(context.Background(), &api.Operation{DropAttr: "Heuristic.excludeSpendingGaps"}); err != nil {
		return err
	}

	return c.Alter(context.Background(), &api.Operation{DropOp: api.Operation_TYPE, DropValue: "Heuristic"})
}

func DropTypeHeuristicResult(c external.Database) error {
	if err := c.Alter(context.Background(), &api.Operation{DropAttr: "HeuristicResult.origin"}); err != nil {
		return err
	}
	if err := c.Alter(context.Background(), &api.Operation{DropAttr: "HeuristicResult.destinations"}); err != nil {
		return err
	}

	return c.Alter(context.Background(), &api.Operation{DropOp: api.Operation_TYPE, DropValue: "HeuristicResult"})
}

func AlterSchemaAddSelectorTotalResultCount(c external.Database) error {
	return c.Alter(context.Background(), &api.Operation{
		Schema: `
			Selector.totalResultCount: int . # number of results found by the selector (can be higher than number of stored results)

			type Selector {
				Selector.created
				Selector.modified
				Selector.type
				Selector.status
				Selector.parent
				Selector.options
				Selector.results
				Selector.totalResultCount
			}`,
	})
}

func DropPrivacyType(c external.Database) error {
	return c.Alter(context.Background(), &api.Operation{DropAttr: "privacytype"})
}

func AlterSchemaAddTransactionType(c external.Database) error {
	return c.Alter(context.Background(), &api.Operation{
		Schema: `
			Transaction.type: string @index(hash) .

			type Transaction {
				txhash
				Transaction.type
				fee
				tx_outputs
				tx_inputs
			}`,
	})
}
