// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package workspace

import (
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

	val1 := int64(1)
	valPoint01 := int64(1000000)
	valPoint1 := int64(10000000)

	opt := workspace.TxPropOptions{
		StartDate:   &startDate1,
		EndDate:     &endDate1,
		InputSum:    &workspace.AmountRange{Min: &val1},
		InputRange:  &workspace.AmountRange{Min: &valPoint01, Max: &valPoint1},
		OutputRange: &workspace.AmountRange{Min: &val1, Max: &valPoint1},
	}
	m := NewMutex()
	ctx, cancel := db.GetTaskContext()
	defer cancel()

	// add parent
	parentSelector, _, err := AddSelector(ctx, dbHandle, m, opt,
		workspace.TypeTxProp, "", workspaceUID, userUID)
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
			selectorType: workspace.TypeTxProp,
			wantErr:      false,
		},
		{
			options:      opt,
			selectorType: workspace.TypeTxProp,
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

	selectorUID1, err := workspace.InsertSelector(ctx, dbHandle, &workspace.Selector{
		Type:    workspace.TypeTxProp,
		Status:  workspace.StatusWaiting,
		Options: "dummy options",
	}, userUID1, workspaceUID1)
	require.NoError(t, err)

	// invalid because wrong status
	isValid, err := isValidParent(ctx, dbHandle, selectorUID1, workspaceUID1, userUID1)
	require.NoError(t, err)
	require.False(t, isValid)

	// invalid because wrong status + wrong user
	isValid, err = isValidParent(ctx, dbHandle, selectorUID1, workspaceUID2, userUID2)
	require.NoError(t, err)
	require.False(t, isValid)

	selectorUID2, err := workspace.InsertSelector(ctx, dbHandle, &workspace.Selector{
		Type:    workspace.TypeTxProp,
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
