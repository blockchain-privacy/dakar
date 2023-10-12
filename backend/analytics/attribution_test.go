package analytics

import (
	"backend/db"
	"backend/db/user"
	"backend/external"
	"backend/testhelper"
	"github.com/stretchr/testify/require"
	"testing"
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
	testhelper.SkipIfNoDB(t)
	db.SetupDB(t, dbHandle, testhelper.UseBlockFile)

	// create dgraph user for tests
	userUID, err := user.CreateNewUser(dbHandle)
	require.NoError(t, err)

	type args struct {
		dgraph       external.Database
		attributions []Attribution
		userID       string
		isPublic     bool
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
				isPublic:     false,
			},
			wantErr: true,
		},
		{
			args: args{
				dgraph:       dbHandle,
				attributions: nil,
				userID:       userUID,
				isPublic:     false,
			},
			wantErr: true,
		},
		{
			args: args{
				dgraph:       dbHandle,
				attributions: attributions,
				userID:       userUID,
				isPublic:     false,
			},
			wantErr: false,
		},
		{
			args: args{
				dgraph:       dbHandle,
				attributions: attributions,
				userID:       userUID,
				isPublic:     true,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		err := ImportAttribution(tt.args.dgraph, tt.args.attributions, tt.args.userID, tt.args.isPublic)
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
		isPublic     bool
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
				isPublic:     false,
			},
			wantEmpty: true,
		},
		{
			args: args{
				attributions: attributions,
				userID:       "some_user_uid",
				hashToUID:    hashToUID,
				isPublic:     false,
			},
			wantEmpty: false,
		},
	}
	for _, tt := range tests {
		got := buildDatabaseAttributions(tt.args.attributions, tt.args.userID, tt.args.hashToUID, tt.args.isPublic)
		if tt.wantEmpty {
			require.Empty(t, got)
		} else {
			require.NotEmpty(t, got)
		}
	}
}

func Test_validateAddresses(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	db.SetupDB(t, dbHandle, testhelper.UseBlockFile)

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
		got, err := validateAddresses(tt.args.dgraph, tt.args.attributions)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.NotEmpty(t, got)
		}
	}
}
