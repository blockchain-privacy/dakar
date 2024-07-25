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
	db.SetupDB(t, dbHandle, testhelper.UseBlockFile)

	startDate1, err := time.Parse(time.RFC3339, "2014-04-28T00:00:00+01:00")
	require.NoError(t, err)
	endDate1, err := time.Parse(time.RFC3339, "2014-04-30T00:00:00+01:00")
	require.NoError(t, err)

	val1 := int64(1)
	val150 := int64(1500000000)
	val200 := int64(2000000000)

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
				InputSum:  &AmountRange{Min: &val200, Max: &val150},
			},
			wantErr: true,
		},
		{
			o: Options{
				StartDate: &startDate1,
				EndDate:   &endDate1,
				OutputSum: &AmountRange{Min: &val200, Max: &val150},
			},
			wantErr: true,
		},
		{
			o: Options{
				StartDate: &startDate1,
				EndDate:   &endDate1,
				OutputSum: &AmountRange{Min: &val200},
			},
			wantErr: false,
		},
		{
			o: Options{
				StartDate: &startDate1,
				EndDate:   &endDate1,
				InputSum:  &AmountRange{Min: &val200},
				OutputSum: &AmountRange{Min: &val150},
			},
			wantErr: false,
		},
		{
			o: Options{
				StartDate: &startDate1,
				EndDate:   &endDate1,
				InputSum:  &AmountRange{Min: &val1, Max: &val200},
			},
			wantErr: false,
		},
		{
			o: Options{
				StartDate:  &startDate1,
				EndDate:    &endDate1,
				InputRange: &AmountRange{Min: &val150, Max: &val1},
			},
			wantErr: true,
		},

		{
			o: Options{
				StartDate:   &startDate1,
				EndDate:     &endDate1,
				OutputRange: &AmountRange{Min: &val150, Max: &val1},
			},
			wantErr: true,
		},
		{
			o: Options{
				StartDate:  &startDate1,
				EndDate:    &endDate1,
				InputRange: &AmountRange{Min: &val150, Max: &val200},
			},
			wantErr: false,
		},
		{
			o: Options{
				StartDate:   &startDate1,
				EndDate:     &endDate1,
				InputRange:  &AmountRange{Min: &val150, Max: &val200},
				OutputRange: &AmountRange{Min: &val1, Max: &val200},
			},
			wantErr: false,
		},
		{
			o: Options{
				StartDate:   &startDate1,
				EndDate:     &endDate1,
				InputSum:    &AmountRange{Min: &val1},
				InputRange:  &AmountRange{Min: &val150, Max: &val200},
				OutputRange: &AmountRange{Min: &val1, Max: &val200},
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
