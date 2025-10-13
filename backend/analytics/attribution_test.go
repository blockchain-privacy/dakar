// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package analytics

import (
	"backend/db"
	"backend/db/user"
	"backend/external"
	"testing"

	"github.com/stretchr/testify/require"
)

var attributions = []Attribution{
	{
		AddressHash: "XgiLmHQ4czfkGvoqLAQJ8SVMNeho1EiFRv",
		Tag:         "tag1",
		Description: "description1",
		Source:      "source1",
		Category:    "category1",
	},
	{
		AddressHash: "Xe5GhnraNWanA3fY1XrjC1RnKQZfWmWygh",
		Tag:         "tag2",
		Description: "description2",
		Source:      "source2",
		Category:    "category2",
	},
	{
		AddressHash: "Xrwhr9kHpnk5CmKLitCcm3aeMv5zNYFZcw",
		Tag:         "tag3",
		Description: "description3",
		Source:      "source3",
		Category:    "category3",
	},
}

func TestImportAttribution(t *testing.T) {
	dbHandle := db.GetDBConnection(t, db.UseBlockFile)

	ctx, cancel := db.GetTaskContext()
	defer cancel()

	// create dgraph user for tests
	userUID, err := user.CreateNewUser(ctx, dbHandle)
	require.NoError(t, err)

	type args struct {
		dgraph       external.Database
		attributions []Attribution
		userID       string
	}
	tests := []struct {
		args    args
		wantErr bool
	}{
		{
			args: args{
				dgraph:       dbHandle,
				attributions: attributions,
				userID:       "",
			},
			wantErr: true,
		},
		{
			args: args{
				dgraph:       dbHandle,
				attributions: nil,
				userID:       userUID,
			},
			wantErr: true,
		},
		{
			args: args{
				dgraph:       dbHandle,
				attributions: attributions,
				userID:       userUID,
			},
			wantErr: false,
		},
		{
			args: args{
				dgraph:       dbHandle,
				attributions: attributions,
				userID:       userUID,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		err := ImportAttribution(t.Context(), tt.args.dgraph, tt.args.attributions, tt.args.userID)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}

func Test_buildDatabaseAttributions(t *testing.T) {
	var hashToUID = map[string]string{
		"XgiLmHQ4czfkGvoqLAQJ8SVMNeho1EiFRv": "uid1",
		"Xe5GhnraNWanA3fY1XrjC1RnKQZfWmWygh": "uid2",
		"Xrwhr9kHpnk5CmKLitCcm3aeMv5zNYFZcw": "uid3",
	}

	type args struct {
		attributions []Attribution
		userID       string
		hashToUID    map[string]string
	}
	tests := []struct {
		args      args
		wantEmpty bool
	}{
		{
			args: args{
				attributions: nil,
				userID:       "",
				hashToUID:    nil,
			},
			wantEmpty: true,
		},
		{
			args: args{
				attributions: attributions,
				userID:       "some_user_uid",
				hashToUID:    hashToUID,
			},
			wantEmpty: false,
		},
	}
	for _, tt := range tests {
		got := buildDatabaseAttributions(tt.args.attributions, tt.args.userID, tt.args.hashToUID)
		if tt.wantEmpty {
			require.Empty(t, got)
		} else {
			require.NotEmpty(t, got)
		}
	}
}

func Test_validateAddresses(t *testing.T) {
	dbHandle := db.GetDBConnection(t, db.UseBlockFile)

	type args struct {
		dgraph       external.Database
		attributions []Attribution
	}
	tests := []struct {
		args    args
		wantErr bool
	}{
		{
			args: args{
				dgraph:       nil,
				attributions: nil,
			},
			wantErr: true,
		},
		{
			args: args{
				dgraph:       dbHandle,
				attributions: attributions,
			},
			wantErr: false,
		},
		{
			args: args{
				dgraph: dbHandle,
				attributions: []Attribution{{
					AddressHash: "invalid_address_hash",
				}},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		got, err := validateAddresses(t.Context(), tt.args.dgraph, tt.args.attributions, false)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.NotEmpty(t, got)
		}
	}
}
