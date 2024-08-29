package analytics

import (
	"backend/db"
	"backend/testhelper"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRemoveTransactionTypeOfAllTransactions(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	db.SetupDB(t, dbHandle, testhelper.UseClassifierFile)

	ctx, cancel := db.GetBackendContext()
	defer cancel()

	// check mixing count before deleting
	mixingCount, _, _, _, _, err := GetTransactionTypeCount(dbHandle)
	require.NoError(t, err)
	require.NotEmpty(t, mixingCount)

	require.NoError(t, RemoveTransactionTypeOfAllTransactions(ctx, dbHandle))

	// check mixing count after deletion
	mixingCount, _, _, _, _, err = GetTransactionTypeCount(dbHandle)
	require.NoError(t, err)
	require.Empty(t, mixingCount)
}
