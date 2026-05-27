// SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package status

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/blockchain-privacy/dakar/db"
)

func TestGetCrawlerStatus(t *testing.T) {
	dbHandle := db.GetBareDBConnection(t)
	ctx, cancel := db.GetTaskContext()
	defer cancel()

	// crawler status not yet set
	_, err := GetCrawlerStatus(ctx, dbHandle)
	require.Error(t, err)

	// set up schema
	require.NoError(t, db.SetupSchema(dbHandle))

	// set crawling
	require.NoError(t, SetCrawling(ctx, dbHandle, true))

	status, err := GetCrawlerStatus(ctx, dbHandle)
	require.NoError(t, err)
	require.True(t, *status.IsCrawling)

	// set not crawling
	require.NoError(t, SetCrawling(ctx, dbHandle, false))

	status, err = GetCrawlerStatus(ctx, dbHandle)
	require.NoError(t, err)
	require.False(t, *status.IsCrawling)
}

func TestGetClassifierStatus(t *testing.T) {
	dbHandle := db.GetBareDBConnection(t)
	ctx, cancel := db.GetTaskContext()
	defer cancel()

	// classifier status not yet set
	_, err := GetClassifierStatus(ctx, dbHandle)
	require.Error(t, err)

	// set up schema
	require.NoError(t, db.SetupSchema(dbHandle))

	// set classifying
	require.NoError(t, SetClassifying(ctx, dbHandle, true))

	status, err := GetClassifierStatus(ctx, dbHandle)
	require.NoError(t, err)
	require.True(t, *status.IsClassifying)

	// set not classifying
	require.NoError(t, SetClassifying(ctx, dbHandle, false))

	status, err = GetClassifierStatus(ctx, dbHandle)
	require.NoError(t, err)
	require.False(t, *status.IsClassifying)
}

func TestGetClusteringFMIStatus(t *testing.T) {
	dbHandle := db.GetBareDBConnection(t)
	ctx, cancel := db.GetTaskContext()
	defer cancel()

	// clustering status not yet set
	_, err := GetClusteringFMIStatus(ctx, dbHandle)
	require.Error(t, err)

	// set up schema
	require.NoError(t, db.SetupSchema(dbHandle))

	// set clustering
	require.NoError(t, SetClusteringFMI(ctx, dbHandle, true))

	status, err := GetClusteringFMIStatus(ctx, dbHandle)
	require.NoError(t, err)
	require.True(t, *status.IsClustering)

	// set not clustering
	require.NoError(t, SetClusteringFMI(ctx, dbHandle, false))

	status, err = GetClusteringFMIStatus(ctx, dbHandle)
	require.NoError(t, err)
	require.False(t, *status.IsClustering)
}

func TestGetHighestBlockID(t *testing.T) {
	dbHandle := db.GetDBConnection(t, db.UseBlockFile)

	ctx, cancel := db.GetTaskContext()
	defer cancel()

	blockHeight, err := GetHighestBlockID(ctx, dbHandle)
	require.NoError(t, err)
	require.EqualValues(t, db.BlockFileLastBlock, blockHeight)
}

func TestGetFrontendStatus(t *testing.T) {
	dbHandle := db.GetBareDBConnection(t)
	ctx, cancel := db.GetTaskContext()
	defer cancel()

	// should not fail even if no status is set at all
	status, err := GetFrontendStatus(ctx, dbHandle)
	require.NoError(t, err)
	require.Nil(t, status.IsCrawling)
	require.Nil(t, status.IsClassifying)

	// set up schema
	require.NoError(t, db.SetupSchema(dbHandle))

	// set crawling
	require.NoError(t, SetCrawling(ctx, dbHandle, true))
	require.NoError(t, SetLastBlockID(ctx, dbHandle, 50))

	status, err = GetFrontendStatus(t.Context(), dbHandle)
	require.NoError(t, err)
	require.NotNil(t, status.IsCrawling)
	require.True(t, *status.IsCrawling)
	require.EqualValues(t, 50, *status.LastBlockID)
}

func TestGetMeta(t *testing.T) {
	dbHandle := db.GetBareDBConnection(t)
	ctx, cancel := db.GetTaskContext()
	defer cancel()

	// nothing set yet -> should fail
	_, err := GetMeta(ctx, dbHandle)
	require.Error(t, err)

	// set up schema
	require.NoError(t, db.SetupSchema(dbHandle))

	// set schema version
	require.NoError(t, InitializeMeta(ctx, dbHandle, "Dash"))

	metaResult, err := GetMeta(ctx, dbHandle)
	require.NoError(t, err)
	require.NotNil(t, metaResult.SchemaVersion)
	require.Equal(t, db.SchemaVersion, *metaResult.SchemaVersion)
	require.Equal(t, "Dash", metaResult.BlockchainMode)
	require.NotEmpty(t, metaResult.CreationTime)
}
