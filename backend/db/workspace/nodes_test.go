// SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package workspace

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/blockchain-privacy/dakar/db"
)

func Test_getWorkspaceConnectionsRaw(t *testing.T) {
	dbHandle := db.GetDBConnection(t, db.UseBlockFile)

	ctx := t.Context()

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
		result, err := getWorkspaceConnectionsRaw(ctx, dbHandle, tt.uids, tt.userUID, tt.workspaceUID)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.NotEmpty(t, result)
		}
	}
}

func TestCheckDuplicateAddress(t *testing.T) {
	dbHandle := db.GetDBConnection(t, db.UseClassifierFile)

	ctx := t.Context()

	// both addresses belong to the same cluster
	addresses, err := db.GetAddressUIDs(ctx, dbHandle, []string{
		"Xw63PFg7vy3rZumKaWn5DRmopKGDcoJ4au",
		"XgVrFBit8H8jnKa3HVeQqiyCJWAp2bN8LU"})
	require.NoError(t, err)
	require.Len(t, addresses, 2)

	duplicateAddress, err := CheckDuplicateAddress(ctx, dbHandle,
		[]string{addresses[0].UID}, addresses[1].UID)
	require.NoError(t, err)
	require.Equal(t, addresses[0].UID, duplicateAddress)

	otherAddress, err := db.GetAddressUIDs(ctx, dbHandle, []string{"Xkti8tVmisBMzmRMAGKVDizk2zVjwWnhAt"})
	require.NoError(t, err)
	require.Len(t, addresses, 2)

	duplicateAddress, err = CheckDuplicateAddress(ctx, dbHandle, []string{addresses[0].UID, addresses[1].UID}, otherAddress[0].UID)
	require.NoError(t, err)
	require.Empty(t, duplicateAddress)
}
