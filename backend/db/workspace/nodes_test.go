package workspace

import (
	"backend/db"
	"backend/testhelper"
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_getWorkspaceConnectionsRaw(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	db.SetupDB(t, dbHandle, testhelper.UseBlockFile)

	ctx := context.Background()

	txUID, err := db.GetTransactionUID(ctx, dbHandle, "91609034d29949f9e19dc62637f0665bdc1b161e11b7f360ee692d15b46c8cdb")
	require.NoError(t, err)
	txUID2, err := db.GetTransactionUID(ctx, dbHandle, "ae52511e1f61977ee2993e47f387d6fe409140dee5783f6df07703360c81a542")
	require.NoError(t, err)
	addressUIDs, err := db.GetAddressUIDs(ctx, dbHandle, []string{"XqwLfPeZSQb9z1QFbCaDkvNQvfutRR4JL3"})
	require.NoError(t, err)
	require.NotEmpty(t, addressUIDs)

	tests := []struct {
		uids         []string
		workspaceUID string
		userUID      string
		wantErr      bool
	}{
		{
			workspaceUID: "0x123",
			userUID:      "0x123",
			uids:         []string{"0xef9504a", "0x4b3acba", "0x4b3ac10", "0xef94d9e", "0x4b3ab85"},
			wantErr:      false,
		},
		{
			workspaceUID: "0x123",
			userUID:      "0x123",
			uids:         []string{"0xf6010fd", "0xf5e0fae"},
			wantErr:      false,
		},
		{
			workspaceUID: "0x123",
			userUID:      "0x123",
			uids:         nil,
			wantErr:      true,
		},
		{
			workspaceUID: "0x123",
			userUID:      "0x123",
			uids:         []string{txUID, txUID2, addressUIDs[0].UID},
			wantErr:      false,
		},
	}
	for _, tt := range tests {
		result, err := getWorkspaceConnectionsRaw(dbHandle, tt.uids, tt.userUID, tt.workspaceUID)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.NotEmpty(t, result)
		}
	}
}
