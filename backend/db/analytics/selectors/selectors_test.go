package selectors

import (
	"backend/db"
	"backend/db/user"
	"backend/db/workspace"
	"backend/testhelper"
	"context"
	"encoding/json"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

var dbHandle = &testhelper.TestDB{IsDirty: true}

func TestMain(m *testing.M) {
	db.InitLogger()
	testhelper.RunDgraphTests(m, &dbHandle.DB)
}

func TestDoSelection(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	db.SetupDB(t, dbHandle, testhelper.UseClassifierFile)

	startDate1, err := time.Parse(time.RFC3339, "2021-10-20T00:00:00+01:00")
	require.NoError(t, err)
	endDate1, err := time.Parse(time.RFC3339, "2021-10-22T00:00:00+01:00")
	require.NoError(t, err)

	val1 := int64(1)
	valPoint01 := int64(1000000)
	valPoint1 := int64(10000000)
	yes := true

	tests := []struct {
		o       Options
		wantErr bool
	}{
		{
			o:       Options{},
			wantErr: true,
		},
		{
			o: Options{
				StartDate: &startDate1,
			},
			wantErr: true,
		},
		{
			o: Options{
				StartDate: &startDate1,
				EndDate:   &endDate1,
			},
			wantErr: true,
		},
		{
			o: Options{
				StartDate: &startDate1,
				EndDate:   &endDate1,
				InputSum:  &AmountRange{Min: &valPoint1, Max: &valPoint01},
			},
			wantErr: true,
		},
		{
			o: Options{
				StartDate: &startDate1,
				EndDate:   &endDate1,
				OutputSum: &AmountRange{Min: &valPoint1, Max: &valPoint01},
			},
			wantErr: true,
		},
		{
			o: Options{
				StartDate: &startDate1,
				EndDate:   &endDate1,
				OutputSum: &AmountRange{Min: &valPoint1},
			},
			wantErr: false,
		},
		{
			o: Options{
				StartDate: &startDate1,
				EndDate:   &endDate1,
				InputSum:  &AmountRange{Min: &valPoint1},
				OutputSum: &AmountRange{Min: &valPoint01},
			},
			wantErr: false,
		},
		{
			o: Options{
				StartDate: &startDate1,
				EndDate:   &endDate1,
				InputSum:  &AmountRange{Min: &val1, Max: &valPoint1},
			},
			wantErr: false,
		},
		{
			o: Options{
				StartDate:  &startDate1,
				EndDate:    &endDate1,
				InputRange: &AmountRange{Min: &valPoint01, Max: &val1},
			},
			wantErr: true,
		},

		{
			o: Options{
				StartDate:   &startDate1,
				EndDate:     &endDate1,
				OutputRange: &AmountRange{Min: &valPoint01, Max: &val1},
			},
			wantErr: true,
		},
		{
			o: Options{
				StartDate:  &startDate1,
				EndDate:    &endDate1,
				InputRange: &AmountRange{Min: &valPoint01, Max: &valPoint1},
			},
			wantErr: false,
		},
		{
			o: Options{
				StartDate:   &startDate1,
				EndDate:     &endDate1,
				InputRange:  &AmountRange{Min: &valPoint01, Max: &valPoint1},
				OutputRange: &AmountRange{Min: &val1, Max: &valPoint1},
			},
			wantErr: false,
		},
		{
			o: Options{
				StartDate:   &startDate1,
				EndDate:     &endDate1,
				InputSum:    &AmountRange{Min: &val1},
				InputRange:  &AmountRange{Min: &valPoint01, Max: &valPoint1},
				OutputRange: &AmountRange{Min: &val1, Max: &valPoint1},
			},
			wantErr: false,
		},
		{
			o: Options{
				StartDate:    &startDate1,
				EndDate:      &endDate1,
				PrivacyTypes: []int{0, 2},
				InputSum:     &AmountRange{Min: &val1},
				InputRange:   &AmountRange{Min: &valPoint01, Max: &valPoint1},
				OutputRange:  &AmountRange{Min: &val1, Max: &valPoint1},
			},
			wantErr: false,
		},
		{
			o: Options{
				StartDate:                  &startDate1,
				EndDate:                    &endDate1,
				PrivacyTypes:               []int{0, 2},
				ExcludePrivacyTransactions: &yes,
			},
			wantErr: true,
		},
		{
			o: Options{
				StartDate:                  &startDate1,
				EndDate:                    &endDate1,
				ExcludePrivacyTransactions: &yes,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		selection, err := DoSelection(context.Background(), dbHandle, tt.o)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.NotEmpty(t, selection)
		}
	}
}

func createUserAndWorkspace() (string, string, error) {
	userUID, err := user.CreateNewUser(dbHandle)
	if err != nil {
		return "", "", err
	}

	workspaceUID, err := workspace.AddWorkspace(context.Background(), dbHandle, "test", userUID)
	if err != nil {
		return "", "", err
	}

	return userUID, workspaceUID, nil
}

func doSelection() ([]string, []byte, error) {
	startDate1, err := time.Parse(time.RFC3339, "2021-10-20T00:00:00+01:00")
	if err != nil {
		return nil, nil, err
	}
	endDate1, err := time.Parse(time.RFC3339, "2021-10-22T00:00:00+01:00")
	if err != nil {
		return nil, nil, err
	}

	val1 := int64(1)
	valPoint01 := int64(1000000)
	valPoint1 := int64(10000000)

	opt := Options{
		StartDate:   &startDate1,
		EndDate:     &endDate1,
		InputSum:    &AmountRange{Min: &val1},
		InputRange:  &AmountRange{Min: &valPoint01, Max: &valPoint1},
		OutputRange: &AmountRange{Min: &val1, Max: &valPoint1},
	}

	optJSON, err := json.Marshal(opt)
	if err != nil {
		return nil, nil, err
	}

	selection, err := DoSelection(context.Background(), dbHandle, opt)
	if err != nil {
		return nil, nil, err
	}

	return selection, optJSON, nil
}

func TestInsertSelector(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	db.SetupDB(t, dbHandle, testhelper.UseClassifierFile)

	ctx := context.Background()

	userUID, workspaceUID, err := createUserAndWorkspace()
	require.NoError(t, err)

	resultUIDs, optJSON, err := doSelection()
	require.NoError(t, err)

	results := make([]db.UIDNode, len(resultUIDs))
	for i, result := range resultUIDs {
		results[i] = db.UIDNode{UID: result}
	}

	now := time.Now().UTC().Format(time.RFC3339)

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
			selector: &Selector{
				Created: "",
			},
			wantErr: true,
		},
		{
			selector: &Selector{
				Created:  now,
				Modified: now,
				Type:     "invalidType",
				Status:   "invalidStatus",
				Options:  string(optJSON),
			},
			wantErr: true,
		},
		{
			selector: &Selector{
				Created:  now,
				Modified: now,
				Type:     "transactionProperties",
				Status:   "success",
				Options:  string(optJSON),
				Results:  results,
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

func TestGetFrontendSelectorByUID(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	db.SetupDB(t, dbHandle, testhelper.UseClassifierFile)

	ctx := context.Background()

	userUID, workspaceUID, err := createUserAndWorkspace()
	require.NoError(t, err)

	resultUIDs, optJSON, err := doSelection()
	require.NoError(t, err)

	results := make([]db.UIDNode, len(resultUIDs))
	for i, result := range resultUIDs {
		results[i] = db.UIDNode{UID: result}
	}

	now := time.Now().UTC().Format(time.RFC3339)
	selectorUID, err := InsertSelector(ctx, dbHandle, &Selector{
		Created:  now,
		Modified: now,
		Type:     typeTransactionProperties,
		Status:   statusSuccess,
		Options:  string(optJSON),
		Results:  results,
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
		selector, err := GetFrontendSelectorByUID(ctx, dbHandle, tt.selectorUID, userUID, workspaceUID)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.NotNil(t, selector)
		}
	}
}

func TestUpdateSelector(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	db.SetupDB(t, dbHandle, testhelper.UseClassifierFile)

	ctx := context.Background()

	userUID, workspaceUID, err := createUserAndWorkspace()
	require.NoError(t, err)

	resultUIDs, optJSON, err := doSelection()
	require.NoError(t, err)

	results := make([]db.UIDNode, len(resultUIDs))
	for i, result := range resultUIDs {
		results[i] = db.UIDNode{UID: result}
	}

	now := time.Now().UTC().Format(time.RFC3339)
	selectorUID, err := InsertSelector(ctx, dbHandle, &Selector{
		Created:  now,
		Modified: now,
		Type:     typeTransactionProperties,
		Status:   statusSuccess,
		Options:  string(optJSON),
		Results:  results,
	}, userUID, workspaceUID)
	require.NoError(t, err)

	tests := []struct {
		status  string
		wantErr bool
	}{
		{
			status:  statusError,
			wantErr: false,
		},
		{
			status:  statusWaiting,
			wantErr: false,
		},
		{
			status:  "invalidStatus",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		err := UpdateSelector(ctx, dbHandle, &Selector{UID: selectorUID, Status: tt.status}, userUID, workspaceUID)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)

			selector, err := GetFrontendSelectorByUID(ctx, dbHandle, selectorUID, userUID, workspaceUID)
			require.NoError(t, err)
			require.EqualValues(t, tt.status, selector.Status)
		}
	}
}

func TestDeleteUserSelectors(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	db.SetupDB(t, dbHandle, testhelper.UseClassifierFile)

	ctx := context.Background()

	userUID, workspaceUID, err := createUserAndWorkspace()
	require.NoError(t, err)

	resultUIDs, optJSON, err := doSelection()
	require.NoError(t, err)

	results := make([]db.UIDNode, len(resultUIDs))
	for i, result := range resultUIDs {
		results[i] = db.UIDNode{UID: result}
	}

	now := time.Now().UTC().Format(time.RFC3339)
	selectorUID, err := InsertSelector(ctx, dbHandle, &Selector{
		Created:  now,
		Modified: now,
		Type:     typeTransactionProperties,
		Status:   statusSuccess,
		Options:  string(optJSON),
		Results:  results,
	}, userUID, workspaceUID)
	require.NoError(t, err)

	err = DeleteUserSelectors(ctx, dbHandle, []string{selectorUID}, userUID, workspaceUID)
	require.NoError(t, err)

	_, err = GetFrontendSelectorByUID(ctx, dbHandle, selectorUID, userUID, workspaceUID)
	require.Error(t, err)
}
