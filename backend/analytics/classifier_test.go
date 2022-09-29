package analytics

import (
	"backend/db"
	"backend/mocks"
	"context"
	"github.com/stretchr/testify/require"
	"testing"
)

type transactionTest struct {
	tx         db.Transaction
	shouldFail bool
}

func TestIsMixing(t *testing.T) {
	shouldWork1 := db.Transaction{
		Fee:  new(int64),
		Hash: "9b6306c63f6f57d23a41a904f2a5d8e41d41623a37bbc03da57813a325c342b2",
		Outputs: []db.Output{
			{Amount: &denominationsTypes[0]},
			{Amount: &denominationsTypes[0]},
			{Amount: &denominationsTypes[0]},
			{Amount: &denominationsTypes[0]},
			{Amount: &denominationsTypes[0]},
			{Amount: &denominationsTypes[0]},
			{Amount: &denominationsTypes[0]},
			{Amount: &denominationsTypes[0]},
		},
		Inputs: []db.Output{
			{Amount: &denominationsTypes[0]},
			{Amount: &denominationsTypes[0]},
			{Amount: &denominationsTypes[0]},
			{Amount: &denominationsTypes[0]},
			{Amount: &denominationsTypes[0]},
			{Amount: &denominationsTypes[0]},
			{Amount: &denominationsTypes[0]},
			{Amount: &denominationsTypes[0]},
		},
	}

	shouldWork2 := db.Transaction{
		Fee:  new(int64),
		Hash: "9b6306c63f6f57d23a41a904f2a5d8e41d41623a37bbc03da57813a325c342b2",
		Outputs: []db.Output{
			{Amount: &denominationsTypes[1]},
			{Amount: &denominationsTypes[1]},
			{Amount: &denominationsTypes[1]},
		},
		Inputs: []db.Output{
			{Amount: &denominationsTypes[1]},
			{Amount: &denominationsTypes[1]},
			{Amount: &denominationsTypes[1]},
		},
	}

	shouldWork3 := db.Transaction{
		Fee:  new(int64),
		Hash: "9b6306c63f6f57d23a41a904f2a5d8e41d41623a37bbc03da57813a325c342b2",
		Outputs: []db.Output{
			{Amount: &denominationsTypes[2]},
			{Amount: &denominationsTypes[2]},
			{Amount: &denominationsTypes[2]},
		},
		Inputs: []db.Output{
			{Amount: &denominationsTypes[2]},
			{Amount: &denominationsTypes[2]},
			{Amount: &denominationsTypes[2]},
		},
	}

	shouldWork4 := db.Transaction{
		Fee:  new(int64),
		Hash: "9b6306c63f6f57d23a41a904f2a5d8e41d41623a37bbc03da57813a325c342b2",
		Outputs: []db.Output{
			{Amount: &denominationsTypes[3]},
			{Amount: &denominationsTypes[3]},
			{Amount: &denominationsTypes[3]},
		},
		Inputs: []db.Output{
			{Amount: &denominationsTypes[3]},
			{Amount: &denominationsTypes[3]},
			{Amount: &denominationsTypes[3]},
		},
	}

	shouldWork5 := db.Transaction{
		Fee:  new(int64),
		Hash: "9b6306c63f6f57d23a41a904f2a5d8e41d41623a37bbc03da57813a325c342b2",
		Outputs: []db.Output{
			{Amount: &denominationsTypes[4]},
			{Amount: &denominationsTypes[4]},
			{Amount: &denominationsTypes[4]},
		},
		Inputs: []db.Output{
			{Amount: &denominationsTypes[4]},
			{Amount: &denominationsTypes[4]},
			{Amount: &denominationsTypes[4]},
		},
	}

	fee := int64(5)
	hasFee := db.Transaction{
		Fee:  &fee,
		Hash: "9b6306c63f6f57d23a41a904f2a5d8e41d41623a37bbc03da57813a325c342b2",
		Outputs: []db.Output{
			{Amount: &denominationsTypes[0]},
			{Amount: &denominationsTypes[0]},
			{Amount: &denominationsTypes[0]},
		},
		Inputs: []db.Output{
			{Amount: &denominationsTypes[0]},
			{Amount: &denominationsTypes[0]},
			{Amount: &denominationsTypes[0]},
		},
	}

	notEqualAmountsOfInputsAndOutputs := db.Transaction{
		Fee:  new(int64),
		Hash: "9b6306c63f6f57d23a41a904f2a5d8e41d41623a37bbc03da57813a325c342b2",
		Outputs: []db.Output{
			{Amount: &denominationsTypes[0]},
			{Amount: &denominationsTypes[0]},
			{Amount: &denominationsTypes[0]},
		},
		Inputs: []db.Output{
			{Amount: &denominationsTypes[0]},
			{Amount: &denominationsTypes[0]},
		},
	}

	mixedDenominations := db.Transaction{
		Fee:  new(int64),
		Hash: "9b6306c63f6f57d23a41a904f2a5d8e41d41623a37bbc03da57813a325c342b2",
		Outputs: []db.Output{
			{Amount: &denominationsTypes[0]},
			{Amount: &denominationsTypes[1]},
			{Amount: &denominationsTypes[0]},
		},
		Inputs: []db.Output{
			{Amount: &denominationsTypes[0]},
			{Amount: &denominationsTypes[0]},
			{Amount: &denominationsTypes[1]},
		},
	}
	one := int64(1)
	notOnlyDenominations := db.Transaction{
		Fee:  new(int64),
		Hash: "9b6306c63f6f57d23a41a904f2a5d8e41d41623a37bbc03da57813a325c342b2",
		Outputs: []db.Output{
			{Amount: &denominationsTypes[0]},
			{Amount: &denominationsTypes[0]},
			{Amount: &one},
		},
		Inputs: []db.Output{
			{Amount: &denominationsTypes[0]},
			{Amount: &denominationsTypes[0]},
			{Amount: &one},
		},
	}

	var cases = []transactionTest{
		{shouldWork1, false},
		{shouldWork2, false},
		{shouldWork3, false},
		{shouldWork4, false},
		{shouldWork5, false},
		{hasFee, true},
		{notEqualAmountsOfInputsAndOutputs, true},
		{mixedDenominations, true},
		{notOnlyDenominations, true},
	}

	for _, c := range cases {
		require.EqualValues(t, !c.shouldFail, isMixing(c.tx) >= 0)
	}
}

func TestIsCollateralPayment(t *testing.T) {
	minCollateral := int64(minCollateral)
	shouldWork1 := db.Transaction{
		Fee:  &minCollateral,
		Hash: "9b6306c63f6f57d23a41a904f2a5d8e41d41623a37bbc03da57813a325c342b2",
		Outputs: []db.Output{
			{Amount: &minCollateral},
		},
		Inputs: []db.Output{
			{Amount: &minCollateral},
		},
	}

	noFee := db.Transaction{
		Fee:  new(int64),
		Hash: "9b6306c63f6f57d23a41a904f2a5d8e41d41623a37bbc03da57813a325c342b2",
		Outputs: []db.Output{
			{Amount: &minCollateral},
		},
		Inputs: []db.Output{
			{Amount: &minCollateral},
		},
	}

	multipleInputs := db.Transaction{
		Fee:  &minCollateral,
		Hash: "9b6306c63f6f57d23a41a904f2a5d8e41d41623a37bbc03da57813a325c342b2",
		Outputs: []db.Output{
			{Amount: &minCollateral},
			{Amount: &minCollateral},
		},
		Inputs: []db.Output{
			{Amount: &minCollateral},
			{Amount: &minCollateral},
		},
	}

	bigAmount := int64(500000000000)
	bigInput := db.Transaction{
		Fee:  &minCollateral,
		Hash: "9b6306c63f6f57d23a41a904f2a5d8e41d41623a37bbc03da57813a325c342b2",
		Outputs: []db.Output{
			{Amount: &bigAmount},
		},
		Inputs: []db.Output{
			{Amount: &bigAmount},
		},
	}

	one := int64(1)
	smallInput := db.Transaction{
		Fee:  &minCollateral,
		Hash: "9b6306c63f6f57d23a41a904f2a5d8e41d41623a37bbc03da57813a325c342b2",
		Outputs: []db.Output{
			{Amount: &one},
		},
		Inputs: []db.Output{
			{Amount: &one},
		},
	}

	var cases = []transactionTest{
		{shouldWork1, false},
		{noFee, true},
		{multipleInputs, true},
		{bigInput, true},
		{smallInput, true},
	}

	for _, c := range cases {
		require.EqualValues(t, !c.shouldFail, isCollateralPayment(c.tx))
	}
}

func TestNewClassifier(t *testing.T) {
	mockDB := new(mocks.Database)
	classifier := NewClassifier(context.Background(), mockDB, NewDashConfig())
	require.NotEmpty(t, classifier.Name())
	require.NotNil(t, classifier.Logger())
	require.NoError(t, classifier.IncrementState())

	mocks.MapSetClassifying(mockDB)
	mocks.MapGetClassifierStatus(mockDB)
	mocks.MapGetCrawlerStatus(mockDB)

	require.NoError(t, classifier.CalculateInitialState())
}

func TestCountAmountDenominations(t *testing.T) {
	type testCase struct {
		amounts []int64
		result  [NumDenominations]int
	}

	var cases = []testCase{
		{
			amounts: []int64{1, 2, 0, 4, 0},
			result:  [NumDenominations]int{0, 0, 0, 0, 0},
		},
		{
			amounts: []int64{1000010000, 1000010000, 1000010000},
			result:  [NumDenominations]int{3, 0, 0, 0, 0},
		},
		{
			amounts: []int64{100001000, 100001000, 100001000, 6, 9, -1},
			result:  [NumDenominations]int{0, 3, 0, 0, 0},
		},
		{
			amounts: []int64{1000010000, 100001000, 10000100, 1000010, 100001},
			result:  [NumDenominations]int{1, 1, 1, 1, 1},
		},
	}

	for _, c := range cases {
		require.Equal(t, c.result, CountAmountDenominations(c.amounts))
	}
}

func TestCountOutputDenominations(t *testing.T) {
	type testCase struct {
		outputs []db.Output
		result  [NumDenominations]int
	}

	notDenom0 := int64(5)
	notDenom1 := int64(-1)
	notDenom2 := int64(0)

	// copy denominations
	denom0 := denominationsTypes[0]
	denom1 := denominationsTypes[1]
	denom2 := denominationsTypes[2]
	denom3 := denominationsTypes[3]
	denom4 := denominationsTypes[4]

	var cases = []testCase{
		{
			outputs: []db.Output{{Amount: &notDenom0}, {Amount: &notDenom1}, {Amount: &notDenom2}},
			result:  [NumDenominations]int{0, 0, 0, 0, 0},
		},
		{
			outputs: []db.Output{{Amount: &denom0}, {Amount: &denom0}, {Amount: &denom0}},
			result:  [NumDenominations]int{3, 0, 0, 0, 0},
		},
		{
			outputs: []db.Output{{Amount: &denom1}, {Amount: &denom1}, {Amount: &denom1},
				{Amount: &notDenom0}, {Amount: &notDenom1}, {Amount: &notDenom2}},
			result: [NumDenominations]int{0, 3, 0, 0, 0},
		},
		{
			outputs: []db.Output{{Amount: &denom0}, {Amount: &denom1}, {Amount: &denom2},
				{Amount: &denom3}, {Amount: &denom4}},
			result: [NumDenominations]int{1, 1, 1, 1, 1},
		},
		{
			// one empty Output should result in an empty result
			outputs: []db.Output{{Amount: &denom0}, {}},
			result:  [NumDenominations]int{},
		},
	}

	for _, c := range cases {
		require.Equal(t, c.result, countOutputDenominations(c.outputs))
	}
}
