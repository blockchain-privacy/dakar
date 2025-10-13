// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package analytics

import (
	"backend/db"
	"backend/db/analytics/exclusion"
	"backend/db/user"
	"backend/external"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestImportAddressExclusions(t *testing.T) {
	dbHandle := db.GetDBConnection(t, db.UseBlockFile)

	ctx, cancel := db.GetTaskContext()
	defer cancel()

	// create dgraph user for tests
	userUID, err := user.CreateNewUser(ctx, dbHandle)
	require.NoError(t, err)

	addresses := []string{
		"XgiLmHQ4czfkGvoqLAQJ8SVMNeho1EiFRv",
		"Xe5GhnraNWanA3fY1XrjC1RnKQZfWmWygh",
		"Xrwhr9kHpnk5CmKLitCcm3aeMv5zNYFZcw",
	}

	type args struct {
		dgraph     external.Database
		exclusions []string
		userID     string
	}
	tests := []struct {
		args    args
		wantErr bool
	}{
		{
			args: args{
				dgraph:     nil,
				exclusions: nil,
				userID:     "",
			},
			wantErr: true,
		},
		{
			args: args{
				dgraph:     nil,
				exclusions: nil,
				userID:     "0xFFFFFF",
			},
			wantErr: true,
		},
		{
			args: args{
				dgraph:     dbHandle,
				exclusions: []string{"some_invalid_address"},
				userID:     "0xFFFFFF",
			},
			wantErr: true,
		},
		{
			args: args{
				dgraph:     dbHandle,
				exclusions: addresses,
				userID:     userUID,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		err := ImportAddressExclusions(t.Context(), tt.args.dgraph, tt.args.exclusions, tt.args.userID)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}

func Test_buildDatabaseAddressExclusions(t *testing.T) {
	type args struct {
		exclusions []string
		userID     string
	}
	tests := []struct {
		args args
		want exclusion.User
	}{
		{
			args: args{
				exclusions: nil,
				userID:     "some_uid",
			},
			want: exclusion.User{UID: "some_uid", Exclusions: []db.UIDNode{}},
		},
		{
			args: args{
				exclusions: []string{"some_other_uid1", "some_other_uid2"},
				userID:     "some_uid",
			},
			want: exclusion.User{UID: "some_uid", Exclusions: []db.UIDNode{
				{UID: "some_other_uid1"},
				{UID: "some_other_uid2"},
			}},
		},
	}
	for _, tt := range tests {
		got := buildDatabaseAddressExclusions(tt.args.exclusions, tt.args.userID)
		require.Equal(t, tt.want, got)
	}
}

func Test_validateExclusionAddresses(t *testing.T) {
	dbHandle := db.GetDBConnection(t, db.UseBlockFile)

	type args struct {
		dgraph     external.Database
		exclusions []string
	}
	tests := []struct {
		args    args
		wantErr bool
	}{
		{
			args: args{
				dgraph:     dbHandle,
				exclusions: nil,
			},
			wantErr: true,
		},
		{
			args: args{
				dgraph:     dbHandle,
				exclusions: []string{"some_invalid_address"},
			},
			wantErr: true,
		},
		{
			args: args{
				dgraph: dbHandle,
				exclusions: []string{
					"XgiLmHQ4czfkGvoqLAQJ8SVMNeho1EiFRv",
					"Xe5GhnraNWanA3fY1XrjC1RnKQZfWmWygh",
					"Xrwhr9kHpnk5CmKLitCcm3aeMv5zNYFZcw",
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		got, err := validateExclusionAddresses(t.Context(), tt.args.dgraph, tt.args.exclusions)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.NotEmpty(t, got)
		}
	}
}
