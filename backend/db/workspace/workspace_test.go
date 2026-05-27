// SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package workspace

import (
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/blockchain-privacy/dakar/db"
	"gitlab.com/blockchain-privacy/dakar/db/user"
)

func TestAddWorkspace(t *testing.T) {
	dbHandle := db.GetDBConnection(t, "")

	ctx, cancel := db.GetTaskContext()
	defer cancel()

	// create dgraph user for tests
	userUID, err := user.CreateNewUser(ctx, dbHandle)
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
		{
			name:    "TTTT TTTT TTTT TTTT TTTT TTTT TTTT TTTT TTTT TTTT T",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		newWorkspaceUID, err := AddWorkspace(t.Context(), dbHandle, tt.name, tt.userUID)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.NotEmpty(t, newWorkspaceUID)
		}
	}
}

func TestGetFrontendWorkspaces(t *testing.T) {
	dbHandle := db.GetDBConnection(t, "")

	ctx := t.Context()
	// create dgraph user and workspace for tests
	userUID, err := user.CreateNewUser(ctx, dbHandle)
	require.NoError(t, err)
	_, err = AddWorkspace(ctx, dbHandle, "test", userUID)
	require.NoError(t, err)

	userUID2, err := user.CreateNewUser(ctx, dbHandle)
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

func TestFindDescendantSelectorUIDs(t *testing.T) {
	filledMap := map[string]Node{
		"0x1": {
			UID:      "0x1",
			Type:     NodeTypeSelector,
			Children: nil,
		},
		"0x2": {
			UID:      "0x2",
			Type:     NodeTypeSelector,
			Children: nil,
		},
		"0x3": {
			UID:      "0x3",
			Type:     NodeTypeSelector,
			Children: nil,
		},
		"0x4": {
			UID:      "0x4",
			Type:     NodeTypeSelector,
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
		require.Equal(t, tt.want, FindDescendantSelectorUIDs(tt.nodes, tt.nodeUID))
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
		require.Equal(t, tt.want, DeleteNodes(tt.nodes, tt.uids))
	}
}

func TestDeleteWorkspace(t *testing.T) {
	dbHandle := db.GetDBConnection(t, "")

	ctx, cancel := db.GetTaskContext()
	defer cancel()
	// create dgraph user and workspace for tests
	userUID, err := user.CreateNewUser(ctx, dbHandle)
	require.NoError(t, err)
	workspaceUID, err := AddWorkspace(ctx, dbHandle, "test", userUID)
	require.NoError(t, err)

	tests := []struct {
		workspaceUID string
		wantErr      bool
	}{
		{
			workspaceUID: workspaceUID,
			wantErr:      false,
		},
		// delete all workspaces
		{
			workspaceUID: "",
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		if err := DeleteWorkspace(ctx, dbHandle, userUID, tt.workspaceUID); tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}
