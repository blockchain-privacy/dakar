// SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package analytics

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/blockchain-privacy/dakar/db"
)

func TestRemoveTransactionTypeOfAllTransactions(t *testing.T) {
	dbHandle := db.GetDBConnection(t, db.UseClassifierFile)

	ctx, cancel := db.GetTaskContext()
	defer cancel()

	// check mixing count before deleting
	mixingCount, _, _, _, _, err := GetDashTransactionTypeCount(ctx, dbHandle)
	require.NoError(t, err)
	require.NotEmpty(t, mixingCount)

	require.NoError(t, RemoveTransactionTypeOfAllTransactions(ctx, dbHandle))

	// check mixing count after deletion
	mixingCount, _, _, _, _, err = GetDashTransactionTypeCount(ctx, dbHandle)
	require.NoError(t, err)
	require.Empty(t, mixingCount)
}
