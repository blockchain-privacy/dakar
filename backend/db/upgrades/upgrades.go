// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package upgrades

import (
	"backend/db"
	"backend/db/status"
	"backend/external"
	"context"
	"log/slog"
	"reflect"
	"runtime"

	"gitlab.com/blockchain-privacy/gomisc/serror"
)

// availableUpgrades contains all available schema upgrades.
// The key is the schema version to which the database should
// be set after its updates haven been applied.
var availableUpgrades = map[int]UpgradePackage{
	15: {upgrades: []schemaUpgrade{AlterSchemaAddErrorCode}},
}

func info(msg string, v ...any) {
	slog.Info(msg, append([]any{"module", "database upgrade"}, v...)...)
}

func getFunctionName(i interface{}) string {
	f := runtime.FuncForPC(reflect.ValueOf(i).Pointer())
	if f == nil {
		return "<unknown function name>"
	}
	return f.Name()
}

// GetSchemaVersion returns the schema version of the database
func GetSchemaVersion(ctx context.Context, db external.Database) (int, error) {
	meta, err := status.GetMeta(ctx, db)
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
func upgradeDatabaseToNextVersion(ctx context.Context, c external.Database,
	upgrades map[int]UpgradePackage, currentSchemaVersion int) error {
	upgradePackage, ok := upgrades[currentSchemaVersion+1]
	if !ok {
		return serror.FromStrWithContext("can not find upgrade package",
			"current database version", currentSchemaVersion)
	}

	if len(upgradePackage.upgrades) == 0 {
		return serror.FromStr("upgrade package contains no upgrades")
	}

	for _, upgrade := range upgradePackage.upgrades {
		info("applying upgrade", "function name", getFunctionName(upgrade))
		if err := upgrade(c); err != nil {
			return err
		}
	}

	if err := status.SetSchemaVersion(ctx, c, currentSchemaVersion+1); err != nil {
		return err
	}

	return nil
}

// UpgradeDatabase upgrades the database schema to the newest version
func UpgradeDatabase(c external.Database) error {
	// no timeout on context
	return applyUpgrades(context.Background(), c, availableUpgrades)
}

// applyUpgrades upgrades the database schema to the newest version, by applying  the given UpgradePackages
func applyUpgrades(ctx context.Context, c external.Database, upgrades map[int]UpgradePackage) error {
	currentSchemaVersion, err := GetSchemaVersion(ctx, c)
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
		if err := upgradeDatabaseToNextVersion(ctx, c, upgrades, currentSchemaVersion); err != nil {
			return err
		}

		newVersion, err := GetSchemaVersion(ctx, c)
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

// AlterSchemaAddErrorCode adds the Selector.errorCode predicate
func AlterSchemaAddErrorCode(c external.Database) error {
	return c.SetSchema(context.Background(), `
			Selector.errorCode: string . # reason for the error status, if any

			type Selector {
				Selector.created
				Selector.modified
				Selector.type
				Selector.status
				Selector.errorCode
				Selector.parent
				Selector.options
				Selector.results
				Selector.totalResultCount
			}`)
}
