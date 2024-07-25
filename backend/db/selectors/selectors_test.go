package selectors

import (
	"backend/db"
	"backend/testhelper"
	"context"
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
