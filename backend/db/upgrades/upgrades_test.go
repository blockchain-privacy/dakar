// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package upgrades

import (
	"backend/db"
	"backend/db/status"
	"backend/external"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
)

func getUpgradesWithError() map[int]UpgradePackage {
	fun := func(external.Database) error { return nil }
	errorFun := func(external.Database) error { return errors.New("error") }
	return map[int]UpgradePackage{
		2: {upgrades: []schemaUpgrade{fun, fun, fun}},
		3: {upgrades: []schemaUpgrade{fun, errorFun, fun}},
	}
}

func getUpgrades() map[int]UpgradePackage {
	fun := func(external.Database) error { return nil }
	upgrades := map[int]UpgradePackage{}
	for i := range db.SchemaVersion {
		upgrades[i+1] = UpgradePackage{upgrades: []schemaUpgrade{fun, fun, fun}}
	}

	return upgrades
}

func Test_upgradeDatabaseToNextVersion(t *testing.T) {
	dbHandle := db.GetDBConnection(t, "")

	ctx, cancel := db.GetTaskContext()
	defer cancel()

	upgrades := getUpgradesWithError()
	tests := []struct {
		upgrades             map[int]UpgradePackage
		currentSchemaVersion int
		wantErr              bool
	}{
		{
			upgrades: nil,
			wantErr:  true,
		},
		{
			upgrades:             map[int]UpgradePackage{2: {upgrades: nil}},
			currentSchemaVersion: 1,
			wantErr:              true,
		},
		// currentSchemaVersion is too low, so should fail
		{
			upgrades:             upgrades,
			currentSchemaVersion: 0,
			wantErr:              true,
		},
		{
			upgrades:             upgrades,
			currentSchemaVersion: 1,
			wantErr:              false,
		},
		// fails because one of the upgrades fails
		{
			upgrades:             upgrades,
			currentSchemaVersion: 2,
			wantErr:              true,
		},
	}
	for _, tt := range tests {
		err := upgradeDatabaseToNextVersion(ctx, dbHandle, tt.upgrades, tt.currentSchemaVersion)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}

func Test_applyUpgrades(t *testing.T) {
	dbHandle := db.GetDBConnection(t, "")
	ctx, cancel := db.GetTaskContext()
	defer cancel()
	require.NoError(t, status.SetMeta(ctx, dbHandle, status.Meta{SchemaVersion: new(1)}))

	tests := []struct {
		upgrades map[int]UpgradePackage
		wantErr  bool
	}{
		{
			upgrades: nil,
			wantErr:  true,
		},
		{
			upgrades: map[int]UpgradePackage{2: {upgrades: nil}},
			wantErr:  true,
		},
		// this test case has to be executed before a successful upgrade,
		// so the schema version is still low enough
		{
			upgrades: getUpgradesWithError(),
			wantErr:  true,
		},
		{
			upgrades: getUpgrades(),
			wantErr:  false,
		},
		// schema should already be updated, so nothing should happen
		{
			upgrades: getUpgrades(),
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		err := applyUpgrades(ctx, dbHandle, tt.upgrades)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}
