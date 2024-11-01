package btc

import (
	"backend/db"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCountAmountWasabi2Denominations(t *testing.T) {
	type testCase struct {
		amounts []int64
		result  [NumWasabi2Denominations]int
	}

	var cases = []testCase{
		{
			amounts: []int64{1, 2, 0, 4, 0},
			result:  [NumWasabi2Denominations]int{},
		},
		{
			amounts: []int64{5000, 5000, 5000},
			result:  [NumWasabi2Denominations]int{3},
		},
		{
			amounts: []int64{6561, 6561, 6561, 6, 9, -1},
			result:  [NumWasabi2Denominations]int{0, 3},
		},
		{
			amounts: []int64{5000, 6561, 8192, 10000, 13122},
			result:  [NumWasabi2Denominations]int{1, 1, 1, 1, 1},
		},
	}

	for _, c := range cases {
		require.Equal(t, c.result, CountAmountWasabi2Denominations(c.amounts))
	}
}

func TestCountWasabi2Denominations(t *testing.T) {
	type testCase struct {
		outputs []db.Output
		result  [NumWasabi2Denominations]int
	}

	notDenom0 := int64(5)
	notDenom1 := int64(-1)
	notDenom2 := int64(0)

	// copy denominations
	denom0 := denominationsTypesWasabi2[0]
	denom1 := denominationsTypesWasabi2[1]
	denom2 := denominationsTypesWasabi2[2]
	denom3 := denominationsTypesWasabi2[3]
	denom4 := denominationsTypesWasabi2[4]

	var cases = []testCase{
		{
			outputs: []db.Output{{Amount: &notDenom0}, {Amount: &notDenom1}, {Amount: &notDenom2}},
			result:  [NumWasabi2Denominations]int{},
		},
		{
			outputs: []db.Output{{Amount: &denom0}, {Amount: &denom0}, {Amount: &denom0}},
			result:  [NumWasabi2Denominations]int{3, 0, 0, 0, 0},
		},
		{
			outputs: []db.Output{{Amount: &denom1}, {Amount: &denom1}, {Amount: &denom1},
				{Amount: &notDenom0}, {Amount: &notDenom1}, {Amount: &notDenom2}},
			result: [NumWasabi2Denominations]int{0, 3},
		},
		{
			outputs: []db.Output{{Amount: &denom0}, {Amount: &denom1}, {Amount: &denom2},
				{Amount: &denom3}, {Amount: &denom4}},
			result: [NumWasabi2Denominations]int{1, 1, 1, 1, 1},
		},
		{
			// one empty Output should result in an empty result
			outputs: []db.Output{{Amount: &denom0}, {}},
			result:  [NumWasabi2Denominations]int{},
		},
	}

	for _, c := range cases {
		require.Equal(t, c.result, countWasabi2Denominations(c.outputs))
	}
}
