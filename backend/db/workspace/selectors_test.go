// SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package workspace

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"gitlab.com/blockchain-privacy/dakar/constants"
	"gitlab.com/blockchain-privacy/dakar/db"
	"gitlab.com/blockchain-privacy/dakar/db/user"
	"gitlab.com/blockchain-privacy/dakar/external"
)

func TestDoSelection(t *testing.T) {
	dbHandle := db.GetDBConnection(t, db.UseClassifierFile)

	startDate1, err := time.Parse(time.RFC3339, "2021-10-20T00:00:00+01:00")
	require.NoError(t, err)
	endDate1, err := time.Parse(time.RFC3339, "2021-10-22T00:00:00+01:00")
	require.NoError(t, err)

	val1 := int64(1)
	valPoint01 := int64(1000000)
	valPoint1 := int64(10000000)
	yes := true

	tests := []struct {
		o       TxPropOptions
		wantErr bool
	}{
		{
			o:       TxPropOptions{},
			wantErr: true,
		},
		{
			o: TxPropOptions{
				StartDate: &startDate1,
			},
			wantErr: true,
		},
		{
			o: TxPropOptions{
				StartDate: &startDate1,
				EndDate:   &endDate1,
			},
			wantErr: true,
		},
		{
			o: TxPropOptions{
				StartDate: &startDate1,
				EndDate:   &endDate1,
				InputSum:  &AmountRange{Min: &valPoint1, Max: &valPoint01},
			},
			wantErr: true,
		},
		{
			o: TxPropOptions{
				StartDate: &startDate1,
				EndDate:   &endDate1,
				OutputSum: &AmountRange{Min: &valPoint1, Max: &valPoint01},
			},
			wantErr: true,
		},
		{
			o: TxPropOptions{
				StartDate: &startDate1,
				EndDate:   &endDate1,
				OutputSum: &AmountRange{Min: &valPoint1},
			},
			wantErr: false,
		},
		{
			o: TxPropOptions{
				StartDate: &startDate1,
				EndDate:   &endDate1,
				InputSum:  &AmountRange{Min: &valPoint1},
				OutputSum: &AmountRange{Min: &valPoint01},
			},
			wantErr: false,
		},
		{
			o: TxPropOptions{
				StartDate: &startDate1,
				EndDate:   &endDate1,
				InputSum:  &AmountRange{Min: &val1, Max: &valPoint1},
			},
			wantErr: false,
		},
		{
			o: TxPropOptions{
				StartDate:  &startDate1,
				EndDate:    &endDate1,
				InputRange: &AmountRange{Min: &valPoint01, Max: &val1},
			},
			wantErr: true,
		},

		{
			o: TxPropOptions{
				StartDate:   &startDate1,
				EndDate:     &endDate1,
				OutputRange: &AmountRange{Min: &valPoint01, Max: &val1},
			},
			wantErr: true,
		},
		{
			o: TxPropOptions{
				StartDate:  &startDate1,
				EndDate:    &endDate1,
				InputRange: &AmountRange{Min: &valPoint01, Max: &valPoint1},
			},
			wantErr: false,
		},
		{
			o: TxPropOptions{
				StartDate:   &startDate1,
				EndDate:     &endDate1,
				InputRange:  &AmountRange{Min: &valPoint01, Max: &valPoint1},
				OutputRange: &AmountRange{Min: &val1, Max: &valPoint1},
			},
			wantErr: false,
		},
		{
			o: TxPropOptions{
				StartDate:   &startDate1,
				EndDate:     &endDate1,
				InputSum:    &AmountRange{Min: &val1},
				InputRange:  &AmountRange{Min: &valPoint01, Max: &valPoint1},
				OutputRange: &AmountRange{Min: &val1, Max: &valPoint1},
			},
			wantErr: false,
		},
		{
			o: TxPropOptions{
				StartDate:        &startDate1,
				EndDate:          &endDate1,
				TransactionTypes: []string{constants.TypeDashOrigin, constants.TypeDashMixing},
				InputSum:         &AmountRange{Min: &val1},
				InputRange:       &AmountRange{Min: &valPoint01, Max: &valPoint1},
				OutputRange:      &AmountRange{Min: &val1, Max: &valPoint1},
			},
			wantErr: false,
		},
		{
			o: TxPropOptions{
				StartDate:                  &startDate1,
				EndDate:                    &endDate1,
				TransactionTypes:           []string{constants.TypeDashOrigin, constants.TypeDashOrigin},
				ExcludePrivacyTransactions: &yes,
			},
			wantErr: true,
		},
		{
			o: TxPropOptions{
				StartDate:                  &startDate1,
				EndDate:                    &endDate1,
				ExcludePrivacyTransactions: &yes,
			},
			wantErr: false,
		},
		{
			// start date after end date
			o: TxPropOptions{
				StartDate: &endDate1,
				EndDate:   &startDate1,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		flag := tt.o.IsValid(t.Context(), dbHandle, "")
		if tt.wantErr {
			require.False(t, flag)
			continue
		}
		require.True(t, flag)

		selection, totalResultCount, err := DoSelection(t.Context(), dbHandle, tt.o, "")
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.NotEmpty(t, selection)
			require.NotZero(t, totalResultCount)
		}
	}
}

func createUserAndWorkspace(dbHandle external.Database) (string, string, error) {
	userUID, err := user.CreateNewUser(context.Background(), dbHandle)
	if err != nil {
		return "", "", err
	}

	workspaceUID, err := AddWorkspace(context.Background(), dbHandle, "test", userUID, "")
	if err != nil {
		return "", "", err
	}

	return userUID, workspaceUID, nil
}

func doSelection(dbHandle external.Database) ([]string, []byte, error) {
	startDate1, err := time.Parse(time.RFC3339, "2021-10-20T00:00:00+01:00")
	if err != nil {
		return nil, nil, err
	}
	endDate1, err := time.Parse(time.RFC3339, "2021-10-22T00:00:00+01:00")
	if err != nil {
		return nil, nil, err
	}

	opt := TxPropOptions{
		StartDate:   &startDate1,
		EndDate:     &endDate1,
		InputSum:    &AmountRange{Min: new(int64(1))},
		InputRange:  &AmountRange{Min: new(int64(1000000)), Max: new(int64(10000000))},
		OutputRange: &AmountRange{Min: new(int64(1)), Max: new(int64(10000000))},
	}

	optJSON, err := json.Marshal(opt)
	if err != nil {
		return nil, nil, err
	}

	selection, _, err := DoSelection(context.Background(), dbHandle, opt, "")
	if err != nil {
		return nil, nil, err
	}

	return selection, optJSON, nil
}

func TestInsertSelector(t *testing.T) {
	dbHandle := db.GetDBConnection(t, db.UseClassifierFile)

	userUID, workspaceUID, err := createUserAndWorkspace(dbHandle)
	require.NoError(t, err)

	resultUIDs, optJSON, err := doSelection(dbHandle)
	require.NoError(t, err)

	results := make([]any, len(resultUIDs))
	for i, result := range resultUIDs {
		results[i] = db.UIDNode{UID: result}
	}

	ctx := t.Context()

	// for parent test
	parentSelectorUID, err := InsertSelector(ctx, dbHandle, &Selector{
		Type:    constants.TypeTxProp,
		Status:  StatusSuccess,
		Options: string(optJSON),
		Results: results,
	}, userUID, workspaceUID)
	require.NoError(t, err)

	tests := []struct {
		selector     *Selector
		userUID      string
		workspaceUID string
		wantErr      bool
	}{
		{
			selector: nil,
			wantErr:  true,
		},
		{
			selector: &Selector{},
			wantErr:  true,
		},
		{
			selector: &Selector{
				Type:    "invalidType",
				Status:  "invalidStatus",
				Options: string(optJSON),
			},
			wantErr: true,
		},
		{
			selector: &Selector{
				Type:    constants.TypeTxProp,
				Status:  StatusSuccess,
				Options: string(optJSON),
				Results: results,
			},
			userUID:      userUID,
			workspaceUID: workspaceUID,
			wantErr:      false,
		},
		// invalid parent UID
		{
			selector: &Selector{
				Type:    constants.TypeTxProp,
				Status:  StatusError,
				Options: string(optJSON),
				Parent:  &db.UIDNode{UID: "0x123"},
			},
			userUID:      userUID,
			workspaceUID: workspaceUID,
			wantErr:      true,
		},
		// empty parent UID
		{
			selector: &Selector{
				Type:    constants.TypeTxProp,
				Status:  StatusError,
				Options: string(optJSON),
				Parent:  &db.UIDNode{UID: ""},
			},
			userUID:      userUID,
			workspaceUID: workspaceUID,
			wantErr:      true,
		},
		{
			selector: &Selector{
				Type:    constants.TypeTxProp,
				Status:  StatusSuccess,
				Options: string(optJSON),
				Parent:  &db.UIDNode{UID: parentSelectorUID},
			},
			userUID:      userUID,
			workspaceUID: workspaceUID,
			wantErr:      false,
		},
	}
	for _, tt := range tests {
		selector, err := InsertSelector(ctx, dbHandle, tt.selector, tt.userUID, tt.workspaceUID)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.NotNil(t, selector)
		}
	}
}

func TestGetSelectorResultsByUID(t *testing.T) {
	dbHandle := db.GetDBConnection(t, db.UseClassifierFile)

	userUID, workspaceUID, err := createUserAndWorkspace(dbHandle)
	require.NoError(t, err)

	resultUIDs, optJSON, err := doSelection(dbHandle)
	require.NoError(t, err)

	results := make([]any, len(resultUIDs))
	for i, result := range resultUIDs {
		results[i] = db.UIDNode{UID: result}
	}
	ctx := t.Context()
	selectorUID, err := InsertSelector(ctx, dbHandle, &Selector{
		Type:    constants.TypeTxProp,
		Status:  StatusSuccess,
		Options: string(optJSON),
		Results: results,
	}, userUID, workspaceUID)
	require.NoError(t, err)

	tests := []struct {
		selectorUID string
		wantErr     bool
	}{
		{
			selectorUID: selectorUID,
			wantErr:     false,
		},
		{
			// invalid selector uid
			selectorUID: "0x123",
			wantErr:     true,
		},
	}
	for _, tt := range tests {
		transactions, err := GetSelectorResultsByUID(ctx, dbHandle, tt.selectorUID, userUID, workspaceUID)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.NotNil(t, transactions)
			require.NotEmpty(t, transactions)
		}
	}
}

func TestUpdateSelector(t *testing.T) {
	dbHandle := db.GetDBConnection(t, db.UseClassifierFile)

	ctx := t.Context()

	userUID, workspaceUID, err := createUserAndWorkspace(dbHandle)
	require.NoError(t, err)

	resultUIDs, optJSON, err := doSelection(dbHandle)
	require.NoError(t, err)

	results := make([]any, len(resultUIDs))
	for i, result := range resultUIDs {
		results[i] = db.UIDNode{UID: result}
	}

	selectorUID, err := InsertSelector(ctx, dbHandle, &Selector{
		Type:    constants.TypeTxProp,
		Status:  StatusSuccess,
		Options: string(optJSON),
		Results: results,
	}, userUID, workspaceUID)
	require.NoError(t, err)

	tests := []struct {
		status       string
		selectorType string
		wantErr      bool
	}{
		{
			status:       StatusError,
			selectorType: constants.TypeTxProp,
			wantErr:      false,
		},
		{
			status:  StatusWaiting,
			wantErr: false,
		},
		{
			status:  "invalidStatus",
			wantErr: true,
		},
		{
			selectorType: "invalidType",
			wantErr:      true,
		},
	}
	for _, tt := range tests {
		err := UpdateSelector(ctx, dbHandle, &Selector{UID: selectorUID, Status: tt.status, Type: tt.selectorType},
			userUID, workspaceUID)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}

func TestDeleteUserSelectors(t *testing.T) {
	dbHandle := db.GetDBConnection(t, db.UseClassifierFile)

	ctx := t.Context()

	userUID, workspaceUID, err := createUserAndWorkspace(dbHandle)
	require.NoError(t, err)

	resultUIDs, optJSON, err := doSelection(dbHandle)
	require.NoError(t, err)

	results := make([]any, len(resultUIDs))
	for i, result := range resultUIDs {
		results[i] = db.UIDNode{UID: result}
	}

	selectorUID, err := InsertSelector(ctx, dbHandle, &Selector{
		Type:    constants.TypeTxProp,
		Status:  StatusSuccess,
		Options: string(optJSON),
		Results: results,
	}, userUID, workspaceUID)
	require.NoError(t, err)

	err = DeleteUserSelectors(ctx, dbHandle, []string{selectorUID}, userUID, workspaceUID)
	require.NoError(t, err)

	// should throw error because selector does not exist anymore
	_, err = GetSelectorResultsByUID(ctx, dbHandle, selectorUID, userUID, workspaceUID)
	require.Error(t, err)
}

func TestGetWaitingSelectors(t *testing.T) {
	dbHandle := db.GetDBConnection(t, db.UseClassifierFile)

	ctx := t.Context()

	userUID, workspaceUID, err := createUserAndWorkspace(dbHandle)
	require.NoError(t, err)

	resultUIDs, optJSON, err := doSelection(dbHandle)
	require.NoError(t, err)

	results := make([]any, len(resultUIDs))
	for i, result := range resultUIDs {
		results[i] = db.UIDNode{UID: result}
	}

	_, err = InsertSelector(ctx, dbHandle, &Selector{
		Type:    constants.TypeTxProp,
		Status:  StatusWaiting,
		Options: string(optJSON),
		Results: results,
	}, userUID, workspaceUID)
	require.NoError(t, err)

	tests := []struct {
		wantReturnCount int
		wantErr         bool
	}{
		{
			wantReturnCount: 1,
			wantErr:         false,
		},
	}
	for _, tt := range tests {
		selectors, err := GetWaitingSelectors(ctx, dbHandle, 5)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.Len(t, selectors, tt.wantReturnCount)
		}
	}
}

func TestGetSelectorStatus(t *testing.T) {
	dbHandle := db.GetDBConnection(t, db.UseClassifierFile)

	ctx := t.Context()

	userUID, workspaceUID, err := createUserAndWorkspace(dbHandle)
	require.NoError(t, err)

	resultUIDs, optJSON, err := doSelection(dbHandle)
	require.NoError(t, err)

	results := make([]any, len(resultUIDs))
	for i, result := range resultUIDs {
		results[i] = db.UIDNode{UID: result}
	}

	// selector uid does not exist
	_, err = GetSelectorStatus(ctx, dbHandle, "0x123", workspaceUID, userUID)
	require.Error(t, err)

	selectorUID, err := InsertSelector(ctx, dbHandle, &Selector{
		Type:    constants.TypeTxProp,
		Status:  StatusWaiting,
		Options: string(optJSON),
		Results: results,
	}, userUID, workspaceUID)
	require.NoError(t, err)

	status, err := GetSelectorStatus(ctx, dbHandle, selectorUID, workspaceUID, userUID)
	require.NoError(t, err)
	require.Equal(t, StatusWaiting, status)
}

func TestDoGraphSelection(t *testing.T) {
	dbHandle := db.GetDBConnection(t, db.UseClassifierFile)

	ctx, cancel := db.GetTaskContext()
	defer cancel()

	uid, err := db.GetTransactionUID(ctx, dbHandle, "4652f5875e3805d9cc7b15d6f5796bfacd8fc58b3a289e1ea187055f3d9d3916")
	require.NoError(t, err)

	tests := []struct {
		o         TxGraphOptions
		parentUID string
		wantErr   bool
	}{
		{
			o:         TxGraphOptions{},
			parentUID: "",
			wantErr:   true,
		},
		{
			o: TxGraphOptions{
				Depth: new(10),
			},
			parentUID: "",
			wantErr:   true,
		},
		{
			o: TxGraphOptions{
				Depth: new(2),
			},
			parentUID: "",
			wantErr:   true,
		},
		{
			o: TxGraphOptions{
				Depth:     new(2),
				IsForward: true,
				MaxItems:  new(5),
			},
			parentUID: uid,
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		flag := tt.o.IsValid(t.Context(), dbHandle, tt.parentUID)
		if tt.wantErr {
			require.False(t, flag)
			continue
		}
		require.True(t, flag)

		selection, totalResultCount, err := DoGraphSelection(t.Context(), dbHandle, tt.o, tt.parentUID)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.NotEmpty(t, selection)
			require.NotZero(t, totalResultCount)
		}
	}
}
