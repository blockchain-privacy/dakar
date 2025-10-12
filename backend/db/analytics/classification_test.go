package analytics

import (
	"backend/db"
	"testing"

	"github.com/stretchr/testify/require"
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
