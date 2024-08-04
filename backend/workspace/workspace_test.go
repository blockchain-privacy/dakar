package workspace

import (
	"backend/analytics"
	"backend/analytics/graph"
	"backend/db"
	"backend/db/analytics/selectors"
	"backend/db/user"
	"backend/db/workspace"
	"backend/testhelper"
	"context"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

var dbHandle = &testhelper.TestDB{IsDirty: true}

func TestMain(m *testing.M) {
	InitLogger()
	db.InitLogger()
	graph.InitLogger()
	analytics.InitLogger()

	testhelper.RunDgraphTests(m, &dbHandle.DB)
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

func TestAddSelector(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	db.SetupDB(t, dbHandle, testhelper.UseClassifierFile)

	userUID, workspaceUID, err := createUserAndWorkspace()
	require.NoError(t, err)

	startDate1, err := time.Parse(time.RFC3339, "2021-10-20T00:00:00+01:00")
	require.NoError(t, err)
	endDate1, err := time.Parse(time.RFC3339, "2021-10-22T00:00:00+01:00")
	require.NoError(t, err)

	val1 := int64(1)
	valPoint01 := int64(1000000)
	valPoint1 := int64(10000000)

	opt := selectors.Options{
		StartDate:   &startDate1,
		EndDate:     &endDate1,
		InputSum:    &selectors.AmountRange{Min: &val1},
		InputRange:  &selectors.AmountRange{Min: &valPoint01, Max: &valPoint1},
		OutputRange: &selectors.AmountRange{Min: &val1, Max: &valPoint1},
	}

	m := NewMutex()
	ctx := context.Background()
	parentSelector, err := AddSelector(ctx, dbHandle, m, opt,
		selectors.TypeTransactionProperties, "", workspaceUID, userUID)
	require.NoError(t, err)

	tests := []struct {
		options      selectors.Options
		selectorType string
		parent       string
		wantErr      bool
	}{
		{
			wantErr: true,
		},
		{
			options:      opt,
			selectorType: selectors.TypeTransactionProperties,
			wantErr:      false,
		},
		{
			options:      opt,
			selectorType: selectors.TypeTransactionProperties,
			parent:       parentSelector,
			wantErr:      false,
		},
	}
	for _, tt := range tests {
		selector, err := AddSelector(ctx, dbHandle, m, tt.options, tt.selectorType, tt.parent, workspaceUID, userUID)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.NotNil(t, selector)
		}
	}
}
