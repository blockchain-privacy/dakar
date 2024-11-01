package btc

import (
	"backend/db"
	"backend/testhelper"
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

func Test_isWasabi2Mixing(t *testing.T) {
	type transactionTest struct {
		tx         db.Transaction
		shouldFail bool
	}

	onlyOneDenomation := db.Transaction{
		Fee:  new(int64),
		Hash: "9b6306c63f6f57d23a41a904f2a5d8e41d41623a37bbc03da57813a325c342b2",
		Inputs: []db.Output{
			{Amount: testhelper.GetPointer[int64](5001)},
			{Amount: testhelper.GetPointer[int64](5002)},
			{Amount: testhelper.GetPointer[int64](5003)},
			{Amount: testhelper.GetPointer[int64](5004)},
			{Amount: testhelper.GetPointer[int64](5005)},
			{Amount: testhelper.GetPointer[int64](5006)},
			{Amount: testhelper.GetPointer[int64](5007)},
			{Amount: testhelper.GetPointer[int64](5008)},
			{Amount: testhelper.GetPointer[int64](5009)},
			{Amount: testhelper.GetPointer[int64](50010)},
			{Amount: testhelper.GetPointer[int64](50011)},
			{Amount: testhelper.GetPointer[int64](50012)},
			{Amount: testhelper.GetPointer[int64](50013)},
			{Amount: testhelper.GetPointer[int64](50014)},
			{Amount: testhelper.GetPointer[int64](50015)},
		},
		Outputs: []db.Output{
			{Amount: testhelper.GetPointer[int64](100000000000), KeyAsm: "1"},
			{Amount: testhelper.GetPointer[int64](100000000000), KeyAsm: "2"},
			{Amount: testhelper.GetPointer[int64](100000000000), KeyAsm: "3"},
			{Amount: testhelper.GetPointer[int64](100000000000), KeyAsm: "4"},
			{Amount: testhelper.GetPointer[int64](100000000000), KeyAsm: "5"},
			{Amount: testhelper.GetPointer[int64](100000000000), KeyAsm: "6"},
			{Amount: testhelper.GetPointer[int64](100000000000), KeyAsm: "7"},
			{Amount: testhelper.GetPointer[int64](100000000000), KeyAsm: "8"},
			{Amount: testhelper.GetPointer[int64](100000000000), KeyAsm: "9"},
			{Amount: testhelper.GetPointer[int64](100000000000), KeyAsm: "10"},
		},
	}

	lowNumberOfDenominations := db.Transaction{
		Fee:  new(int64),
		Hash: "9b6306c63f6f57d23a41a904f2a5d8e41d41623a37bbc03da57813a325c342b2",
		Inputs: []db.Output{
			{Amount: testhelper.GetPointer[int64](5001)},
			{Amount: testhelper.GetPointer[int64](5002)},
			{Amount: testhelper.GetPointer[int64](5003)},
			{Amount: testhelper.GetPointer[int64](5004)},
			{Amount: testhelper.GetPointer[int64](5005)},
			{Amount: testhelper.GetPointer[int64](5006)},
			{Amount: testhelper.GetPointer[int64](5007)},
			{Amount: testhelper.GetPointer[int64](5008)},
			{Amount: testhelper.GetPointer[int64](5009)},
			{Amount: testhelper.GetPointer[int64](50010)},
			{Amount: testhelper.GetPointer[int64](50011)},
			{Amount: testhelper.GetPointer[int64](50012)},
			{Amount: testhelper.GetPointer[int64](50013)},
			{Amount: testhelper.GetPointer[int64](50014)},
			{Amount: testhelper.GetPointer[int64](50015)},
		},
		Outputs: []db.Output{
			{Amount: testhelper.GetPointer[int64](100000000000), KeyAsm: "1"},
			{Amount: testhelper.GetPointer[int64](100000000000), KeyAsm: "2"},
			{Amount: testhelper.GetPointer[int64](100000000000), KeyAsm: "3"},
			{Amount: testhelper.GetPointer[int64](1), KeyAsm: "4"},
			{Amount: testhelper.GetPointer[int64](2), KeyAsm: "5"},
			{Amount: testhelper.GetPointer[int64](3), KeyAsm: "6"},
			{Amount: testhelper.GetPointer[int64](4), KeyAsm: "7"},
			{Amount: testhelper.GetPointer[int64](5), KeyAsm: "8"},
			{Amount: testhelper.GetPointer[int64](6), KeyAsm: "9"},
			{Amount: testhelper.GetPointer[int64](7), KeyAsm: "10"},
		},
	}

	shouldWork := db.Transaction{
		Fee:  new(int64),
		Hash: "9b6306c63f6f57d23a41a904f2a5d8e41d41623a37bbc03da57813a325c342b2",
		Inputs: []db.Output{
			{Amount: testhelper.GetPointer[int64](5001)},
			{Amount: testhelper.GetPointer[int64](5002)},
			{Amount: testhelper.GetPointer[int64](5003)},
			{Amount: testhelper.GetPointer[int64](5004)},
			{Amount: testhelper.GetPointer[int64](5005)},
			{Amount: testhelper.GetPointer[int64](5006)},
			{Amount: testhelper.GetPointer[int64](5007)},
			{Amount: testhelper.GetPointer[int64](5008)},
			{Amount: testhelper.GetPointer[int64](5009)},
			{Amount: testhelper.GetPointer[int64](50010)},
			{Amount: testhelper.GetPointer[int64](50011)},
			{Amount: testhelper.GetPointer[int64](50012)},
			{Amount: testhelper.GetPointer[int64](50013)},
			{Amount: testhelper.GetPointer[int64](50014)},
			{Amount: testhelper.GetPointer[int64](50015)},
		},
		Outputs: []db.Output{
			{Amount: testhelper.GetPointer[int64](258280326), KeyAsm: "1"},
			{Amount: testhelper.GetPointer[int64](4782969), KeyAsm: "2"},
			{Amount: testhelper.GetPointer[int64](8388608), KeyAsm: "3"},
			{Amount: testhelper.GetPointer[int64](19683), KeyAsm: "4"},
			{Amount: testhelper.GetPointer[int64](20000), KeyAsm: "5"},
			{Amount: testhelper.GetPointer[int64](1), KeyAsm: "6"},
			{Amount: testhelper.GetPointer[int64](2), KeyAsm: "7"},
			{Amount: testhelper.GetPointer[int64](3), KeyAsm: "8"},
			{Amount: testhelper.GetPointer[int64](4), KeyAsm: "9"},
			{Amount: testhelper.GetPointer[int64](5), KeyAsm: "10"},
		},
	}

	lowNumberOfOutputs := db.Transaction{
		Fee:  new(int64),
		Hash: "9b6306c63f6f57d23a41a904f2a5d8e41d41623a37bbc03da57813a325c342b2",
		Inputs: []db.Output{
			{Amount: testhelper.GetPointer[int64](5001)},
			{Amount: testhelper.GetPointer[int64](5002)},
		},
		Outputs: []db.Output{
			{Amount: testhelper.GetPointer[int64](258280326), KeyAsm: "1"},
			{Amount: testhelper.GetPointer[int64](4782969), KeyAsm: "2"},
		},
	}

	var cases = []transactionTest{
		{onlyOneDenomation, true},
		{lowNumberOfDenominations, true},
		{lowNumberOfOutputs, true},
		{shouldWork, false},
	}

	for _, c := range cases {
		require.EqualValues(t, !c.shouldFail, isWasabi2Mixing(c.tx))
	}
}
