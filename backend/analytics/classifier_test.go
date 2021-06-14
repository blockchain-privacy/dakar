package analytics

import (
	dbop "backend/db/output"
	dbtx "backend/db/transaction"
	"github.com/stretchr/testify/require"
	"testing"
)

var denominations = [5]int64{1000010000, 100001000, 10000100, 1000010, 100001}

type transactionTest struct {
	tx         dbtx.Transaction
	shouldFail bool
}

func TestIsMixing(t *testing.T) {
	shouldWork1 := dbtx.Transaction{
		Fee:  new(int64),
		Hash: "9b6306c63f6f57d23a41a904f2a5d8e41d41623a37bbc03da57813a325c342b2",
		Outputs: []dbop.Output{
			{Amount: &denominations[0]},
			{Amount: &denominations[0]},
			{Amount: &denominations[0]},
			{Amount: &denominations[0]},
			{Amount: &denominations[0]},
			{Amount: &denominations[0]},
			{Amount: &denominations[0]},
			{Amount: &denominations[0]},
		},
		Inputs: []dbop.Output{
			{Amount: &denominations[0]},
			{Amount: &denominations[0]},
			{Amount: &denominations[0]},
			{Amount: &denominations[0]},
			{Amount: &denominations[0]},
			{Amount: &denominations[0]},
			{Amount: &denominations[0]},
			{Amount: &denominations[0]},
		},
	}

	shouldWork2 := dbtx.Transaction{
		Fee:  new(int64),
		Hash: "9b6306c63f6f57d23a41a904f2a5d8e41d41623a37bbc03da57813a325c342b2",
		Outputs: []dbop.Output{
			{Amount: &denominations[1]},
			{Amount: &denominations[1]},
			{Amount: &denominations[1]},
		},
		Inputs: []dbop.Output{
			{Amount: &denominations[1]},
			{Amount: &denominations[1]},
			{Amount: &denominations[1]},
		},
	}

	shouldWork3 := dbtx.Transaction{
		Fee:  new(int64),
		Hash: "9b6306c63f6f57d23a41a904f2a5d8e41d41623a37bbc03da57813a325c342b2",
		Outputs: []dbop.Output{
			{Amount: &denominations[2]},
			{Amount: &denominations[2]},
			{Amount: &denominations[2]},
		},
		Inputs: []dbop.Output{
			{Amount: &denominations[2]},
			{Amount: &denominations[2]},
			{Amount: &denominations[2]},
		},
	}

	shouldWork4 := dbtx.Transaction{
		Fee:  new(int64),
		Hash: "9b6306c63f6f57d23a41a904f2a5d8e41d41623a37bbc03da57813a325c342b2",
		Outputs: []dbop.Output{
			{Amount: &denominations[3]},
			{Amount: &denominations[3]},
			{Amount: &denominations[3]},
		},
		Inputs: []dbop.Output{
			{Amount: &denominations[3]},
			{Amount: &denominations[3]},
			{Amount: &denominations[3]},
		},
	}

	shouldWork5 := dbtx.Transaction{
		Fee:  new(int64),
		Hash: "9b6306c63f6f57d23a41a904f2a5d8e41d41623a37bbc03da57813a325c342b2",
		Outputs: []dbop.Output{
			{Amount: &denominations[4]},
			{Amount: &denominations[4]},
			{Amount: &denominations[4]},
		},
		Inputs: []dbop.Output{
			{Amount: &denominations[4]},
			{Amount: &denominations[4]},
			{Amount: &denominations[4]},
		},
	}

	fee := int64(5)
	hasFee := dbtx.Transaction{
		Fee:  &fee,
		Hash: "9b6306c63f6f57d23a41a904f2a5d8e41d41623a37bbc03da57813a325c342b2",
		Outputs: []dbop.Output{
			{Amount: &denominations[0]},
			{Amount: &denominations[0]},
			{Amount: &denominations[0]},
		},
		Inputs: []dbop.Output{
			{Amount: &denominations[0]},
			{Amount: &denominations[0]},
			{Amount: &denominations[0]},
		},
	}

	notEqualAmountsOfInputsAndOutputs := dbtx.Transaction{
		Fee:  new(int64),
		Hash: "9b6306c63f6f57d23a41a904f2a5d8e41d41623a37bbc03da57813a325c342b2",
		Outputs: []dbop.Output{
			{Amount: &denominations[0]},
			{Amount: &denominations[0]},
			{Amount: &denominations[0]},
		},
		Inputs: []dbop.Output{
			{Amount: &denominations[0]},
			{Amount: &denominations[0]},
		},
	}

	mixedDenominations := dbtx.Transaction{
		Fee:  new(int64),
		Hash: "9b6306c63f6f57d23a41a904f2a5d8e41d41623a37bbc03da57813a325c342b2",
		Outputs: []dbop.Output{
			{Amount: &denominations[0]},
			{Amount: &denominations[1]},
			{Amount: &denominations[0]},
		},
		Inputs: []dbop.Output{
			{Amount: &denominations[0]},
			{Amount: &denominations[0]},
			{Amount: &denominations[1]},
		},
	}
	one := int64(1)
	notOnlyDenominations := dbtx.Transaction{
		Fee:  new(int64),
		Hash: "9b6306c63f6f57d23a41a904f2a5d8e41d41623a37bbc03da57813a325c342b2",
		Outputs: []dbop.Output{
			{Amount: &denominations[0]},
			{Amount: &denominations[0]},
			{Amount: &one},
		},
		Inputs: []dbop.Output{
			{Amount: &denominations[0]},
			{Amount: &denominations[0]},
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
	minCollateral := int64(dbop.MinCollateral)
	shouldWork1 := dbtx.Transaction{
		Fee:  &minCollateral,
		Hash: "9b6306c63f6f57d23a41a904f2a5d8e41d41623a37bbc03da57813a325c342b2",
		Outputs: []dbop.Output{
			{Amount: &minCollateral},
		},
		Inputs: []dbop.Output{
			{Amount: &minCollateral},
		},
	}

	noFee := dbtx.Transaction{
		Fee:  new(int64),
		Hash: "9b6306c63f6f57d23a41a904f2a5d8e41d41623a37bbc03da57813a325c342b2",
		Outputs: []dbop.Output{
			{Amount: &minCollateral},
		},
		Inputs: []dbop.Output{
			{Amount: &minCollateral},
		},
	}

	multipleInputs := dbtx.Transaction{
		Fee:  &minCollateral,
		Hash: "9b6306c63f6f57d23a41a904f2a5d8e41d41623a37bbc03da57813a325c342b2",
		Outputs: []dbop.Output{
			{Amount: &minCollateral},
			{Amount: &minCollateral},
		},
		Inputs: []dbop.Output{
			{Amount: &minCollateral},
			{Amount: &minCollateral},
		},
	}

	bigAmount := int64(500000000000)
	bigInput := dbtx.Transaction{
		Fee:  &minCollateral,
		Hash: "9b6306c63f6f57d23a41a904f2a5d8e41d41623a37bbc03da57813a325c342b2",
		Outputs: []dbop.Output{
			{Amount: &bigAmount},
		},
		Inputs: []dbop.Output{
			{Amount: &bigAmount},
		},
	}

	one := int64(1)
	smallInput := dbtx.Transaction{
		Fee:  &minCollateral,
		Hash: "9b6306c63f6f57d23a41a904f2a5d8e41d41623a37bbc03da57813a325c342b2",
		Outputs: []dbop.Output{
			{Amount: &one},
		},
		Inputs: []dbop.Output{
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
