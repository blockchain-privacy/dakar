package workspace

import (
	"backend/db"
	"backend/db/user"
	"backend/testhelper"
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

var dbHandle = &testhelper.TestDB{IsDirty: true}

func TestMain(m *testing.M) {
	db.InitLogger()
	testhelper.RunDgraphTests(m, &dbHandle.DB)
}

func TestAddWorkspace(t *testing.T) {
	testhelper.SkipIfNoDB(t)

	// create dgraph user for tests
	userUID, err := user.CreateNewUser(dbHandle)
	require.NoError(t, err)

	tests := []struct {
		name    string
		userUID string
		wantErr bool
	}{
		{
			name:    "",
			userUID: "",
			wantErr: true,
		},
		{
			name:    "test",
			wantErr: true,
		},
		{
			name:    "test",
			userUID: userUID,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		newWorkspaceUID, err := AddWorkspace(context.Background(), dbHandle, tt.name, tt.userUID)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.NotEmpty(t, newWorkspaceUID)
		}
	}
}

func TestGetFrontendWorkspaces(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	ctx := context.Background()
	// create dgraph user and workspace for tests
	userUID, err := user.CreateNewUser(dbHandle)
	require.NoError(t, err)
	_, err = AddWorkspace(ctx, dbHandle, "test", userUID)
	require.NoError(t, err)

	userUID2, err := user.CreateNewUser(dbHandle)
	require.NoError(t, err)
	_, err = AddWorkspace(ctx, dbHandle, "test", userUID2)
	require.NoError(t, err)
	_, err = AddWorkspace(ctx, dbHandle, "test", userUID2)
	require.NoError(t, err)
	_, err = AddWorkspace(ctx, dbHandle, "test", userUID2)
	require.NoError(t, err)

	tests := []struct {
		userUID       string
		numWorkspaces int
		wantErr       bool
	}{
		{
			userUID: "",
			wantErr: true,
		},
		{
			userUID:       "0x123", // user id does not exist, but should not error
			numWorkspaces: 0,
			wantErr:       false,
		},
		{
			userUID:       userUID2,
			numWorkspaces: 3,
			wantErr:       false,
		},
	}
	for _, tt := range tests {
		workspaces, err := GetFrontendWorkspaces(ctx, dbHandle, tt.userUID)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.Len(t, workspaces, tt.numWorkspaces)
		}
	}
}

func TestFindDescandantHeuristicUIDs(t *testing.T) {
	filledMap := map[string]Node{
		"0x1": {
			UID:      "0x1",
			Type:     "heuristic",
			Children: nil,
		},
		"0x2": {
			UID:      "0x2",
			Type:     "heuristic",
			Children: nil,
		},
		"0x3": {
			UID:      "0x3",
			Type:     "heuristic",
			Children: nil,
		},
		"0x4": {
			UID:      "0x4",
			Type:     "heuristic",
			Children: []string{"0x2", "0x3"},
		},
	}

	tests := []struct {
		nodes   map[string]Node
		nodeUID string
		want    []string
	}{
		{},
		// node not in map
		{
			nodes:   map[string]Node{},
			nodeUID: "0x123",
			want:    nil,
		},
		// node in map
		{
			nodes:   filledMap,
			nodeUID: "0x2",
			want:    []string{"0x2"},
		},
		// node with children in map
		{
			nodes:   filledMap,
			nodeUID: "0x4",
			want:    []string{"0x4", "0x2", "0x3"},
		},
	}
	for _, tt := range tests {
		require.EqualValues(t, tt.want, FindDescendantHeuristicUIDs(tt.nodes, tt.nodeUID))
	}
}

func TestDeleteNodes(t *testing.T) {
	tests := []struct {
		nodes []Node
		uids  []string
		want  []Node
	}{
		{
			nodes: []Node{{UID: "0x1"}, {UID: "0x2"}, {UID: "0x3"}, {UID: "0x4"}, {UID: "0x5"}},
			uids:  []string{"0x2", "0x5"},
			want:  []Node{{UID: "0x1"}, {UID: "0x3"}, {UID: "0x4"}},
		},
		{
			nodes: []Node{{UID: "0x1"}, {UID: "0x2"}, {UID: "0x3"}, {UID: "0x4"}, {UID: "0x5"}},
			uids:  []string{"0x1"},
			want:  []Node{{UID: "0x2"}, {UID: "0x3"}, {UID: "0x4"}, {UID: "0x5"}},
		},
	}
	for _, tt := range tests {
		require.EqualValues(t, tt.want, DeleteNodes(tt.nodes, tt.uids))
	}
}
