// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package analytics

import (
	"backend/db"
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/blockchain-privacy/gomisc/serror"
)

func TestGetAllFMIClusters(t *testing.T) {
	dbHandle := db.GetDBConnection(t, db.UseBlockFile)

	ctx, cancel := db.GetTaskContext()
	defer cancel()

	clusters, err := GetAllFMIClusters(ctx, dbHandle)
	require.NoError(t, err)
	require.NotEmpty(t, clusters)
}

func TestGetTransactionCountPerCluster(t *testing.T) {
	dbHandle := db.GetDBConnection(t, db.UseBlockFile)

	ctx, cancel := db.GetTaskContext()
	defer cancel()

	clusters, err := GetAllFMIClusters(ctx, dbHandle)
	require.NoError(t, err)
	require.NotEmpty(t, clusters)

	for i, c := range clusters {
		inputCount, outputCount, err := GetTransactionCountPerCluster(ctx, dbHandle, c)
		require.NoError(t, err)
		require.NotZero(t, inputCount+outputCount)

		if i > 100 {
			break
		}
	}
}

func Test_isDeadlineExceeded(t *testing.T) {
	err := errors.New("test")
	require.False(t, isDeadlineExceeded(err))
	err = serror.FromStr("test")
	require.False(t, isDeadlineExceeded(err))
	ctxErr := context.DeadlineExceeded
	require.True(t, isDeadlineExceeded(ctxErr))
}
