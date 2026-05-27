// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package workspace

import (
	"backend/constants"
	"backend/db"
	"backend/db/user"
	"backend/db/workspace"
	"backend/external"
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func createUserAndWorkspace(dbHandle external.Database) (string, string, error) {
	userUID, err := user.CreateNewUser(context.Background(), dbHandle)
	if err != nil {
		return "", "", err
	}

	workspaceUID, err := workspace.AddWorkspace(context.Background(), dbHandle, "test", userUID)
	if err != nil {
		return "", "", err
	}

	return userUID, workspaceUID, nil
}

func TestAddSelector(t *testing.T) {
	dbHandle := db.GetDBConnection(t, db.UseClassifierFile)

	userUID, workspaceUID, err := createUserAndWorkspace(dbHandle)
	require.NoError(t, err)

	startDate1, err := time.Parse(time.RFC3339, "2021-10-20T00:00:00+01:00")
	require.NoError(t, err)
	endDate1, err := time.Parse(time.RFC3339, "2021-10-22T00:00:00+01:00")
	require.NoError(t, err)

	opt := workspace.TxPropOptions{
		StartDate:   &startDate1,
		EndDate:     &endDate1,
		InputSum:    &workspace.AmountRange{Min: new(int64(1))},
		InputRange:  &workspace.AmountRange{Min: new(int64(1000000)), Max: new(int64(10000000))},
		OutputRange: &workspace.AmountRange{Min: new(int64(1)), Max: new(int64(10000000))},
	}
	m := NewMutex()
	ctx, cancel := db.GetTaskContext()
	defer cancel()

	// add parent
	parentSelector, _, err := AddSelector(ctx, dbHandle, m, opt,
		constants.TypeTxProp, "", workspaceUID, userUID)
	require.NoError(t, err)

	// set parent to be successful
	require.NoError(t, workspace.UpdateSelector(ctx, dbHandle, &workspace.Selector{
		UID:    parentSelector,
		Status: workspace.StatusSuccess,
	}, userUID, workspaceUID))

	tests := []struct {
		options      workspace.TxPropOptions
		selectorType string
		parent       string
		wantErr      bool
	}{
		{
			wantErr: true,
		},
		{
			options:      opt,
			selectorType: constants.TypeTxProp,
			wantErr:      false,
		},
		{
			options:      opt,
			selectorType: constants.TypeTxProp,
			parent:       parentSelector,
			wantErr:      false,
		},
	}
	for _, tt := range tests {
		selector, _, err := AddSelector(ctx, dbHandle, m, tt.options, tt.selectorType, tt.parent, workspaceUID, userUID)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.NotNil(t, selector)
		}
	}
}

func Test_isValidParent(t *testing.T) {
	dbHandle := db.GetDBConnection(t, db.UseClassifierFile)

	userUID1, workspaceUID1, err := createUserAndWorkspace(dbHandle)
	require.NoError(t, err)

	userUID2, workspaceUID2, err := createUserAndWorkspace(dbHandle)
	require.NoError(t, err)

	ctx, cancel := db.GetTaskContext()
	defer cancel()

	txUID, err := db.GetTransactionUID(t.Context(), dbHandle, "c8eda15137f6ffcfbe851847e621a6f524da04a2fcec2a3cb15b8a92f6bfe0b6")
	require.NoError(t, err)

	selectorUID1, err := workspace.InsertSelector(ctx, dbHandle, &workspace.Selector{
		Type:    constants.TypeTxProp,
		Status:  workspace.StatusWaiting,
		Options: "dummy options",
	}, userUID1, workspaceUID1)
	require.NoError(t, err)

	isValid, err := isValidParent(ctx, dbHandle, txUID, workspaceUID1, userUID1)
	require.NoError(t, err)
	require.True(t, isValid)

	// invalid because parent uid is empty
	isValid, err = isValidParent(ctx, dbHandle, "", workspaceUID1, userUID1)
	require.Error(t, err)
	require.False(t, isValid)

	// invalid because wrong status
	isValid, err = isValidParent(ctx, dbHandle, selectorUID1, workspaceUID1, userUID1)
	require.NoError(t, err)
	require.False(t, isValid)

	// invalid because wrong status + wrong user
	isValid, err = isValidParent(ctx, dbHandle, selectorUID1, workspaceUID2, userUID2)
	require.NoError(t, err)
	require.False(t, isValid)

	selectorUID2, err := workspace.InsertSelector(ctx, dbHandle, &workspace.Selector{
		Type:    constants.TypeTxProp,
		Status:  workspace.StatusSuccess,
		Options: "dummy options",
	}, userUID1, workspaceUID1)
	require.NoError(t, err)

	isValid, err = isValidParent(ctx, dbHandle, selectorUID2, workspaceUID1, userUID1)
	require.NoError(t, err)
	require.True(t, isValid)

	// invalid because wrong user
	isValid, err = isValidParent(ctx, dbHandle, selectorUID2, workspaceUID2, userUID2)
	require.NoError(t, err)
	require.False(t, isValid)
}
