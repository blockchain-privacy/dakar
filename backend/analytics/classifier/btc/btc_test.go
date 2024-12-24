package btc

import (
	"backend/constants"
	"backend/db"
	"backend/db/analytics"
	"backend/testhelper"
	"github.com/stretchr/testify/require"
	"testing"
)

var dbHandle = &testhelper.TestDB{}

func TestMain(m *testing.M) {
	testhelper.RunDgraphTests(m, dbHandle)
}

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

	noUncommonDenomation := db.Transaction{
		Fee:  new(int64),
		Hash: "f7a484919cee80b71d61a37cb901a6aa6204e7f26535cb7e479079722d571b10",
		Inputs: []db.Output{
			{Amount: testhelper.GetPointer[int64](874642)},
			{Amount: testhelper.GetPointer[int64](874642)},
			{Amount: testhelper.GetPointer[int64](874642)},
			{Amount: testhelper.GetPointer[int64](874642)},
			{Amount: testhelper.GetPointer[int64](874642)},
			{Amount: testhelper.GetPointer[int64](874642)},
			{Amount: testhelper.GetPointer[int64](874642)},
			{Amount: testhelper.GetPointer[int64](874642)},
			{Amount: testhelper.GetPointer[int64](874642)},
			{Amount: testhelper.GetPointer[int64](874642)},
			{Amount: testhelper.GetPointer[int64](874642)},
			{Amount: testhelper.GetPointer[int64](874642)},
			{Amount: testhelper.GetPointer[int64](874642)},
			{Amount: testhelper.GetPointer[int64](874642)},
			{Amount: testhelper.GetPointer[int64](874642)},
			{Amount: testhelper.GetPointer[int64](874642)},
			{Amount: testhelper.GetPointer[int64](874642)},
			{Amount: testhelper.GetPointer[int64](874642)},
			{Amount: testhelper.GetPointer[int64](874642)},
			{Amount: testhelper.GetPointer[int64](874642)},
			{Amount: testhelper.GetPointer[int64](874642)},
			{Amount: testhelper.GetPointer[int64](874642)},
			{Amount: testhelper.GetPointer[int64](874642)},
			{Amount: testhelper.GetPointer[int64](874642)},
			{Amount: testhelper.GetPointer[int64](874642)},
			{Amount: testhelper.GetPointer[int64](874642)},
			{Amount: testhelper.GetPointer[int64](874642)},
			{Amount: testhelper.GetPointer[int64](874642)},
			{Amount: testhelper.GetPointer[int64](874642)},
			{Amount: testhelper.GetPointer[int64](874642)},
		},
		Outputs: []db.Output{
			{Amount: testhelper.GetPointer[int64](133200000), KeyAsm: "1"},
			{Amount: testhelper.GetPointer[int64](5000000), KeyAsm: "2"},
			{Amount: testhelper.GetPointer[int64](150000000), KeyAsm: "3"},
			{Amount: testhelper.GetPointer[int64](3799000000), KeyAsm: "4"},
			{Amount: testhelper.GetPointer[int64](100000000), KeyAsm: "5"},
			{Amount: testhelper.GetPointer[int64](1700000), KeyAsm: "6"},
			{Amount: testhelper.GetPointer[int64](1000000), KeyAsm: "7"},
			{Amount: testhelper.GetPointer[int64](2000000), KeyAsm: "8"},
			{Amount: testhelper.GetPointer[int64](482500000), KeyAsm: "9"},
			{Amount: testhelper.GetPointer[int64](100000000), KeyAsm: "10"},
			{Amount: testhelper.GetPointer[int64](234403702357), KeyAsm: "11"},
		},
	}

	var cases = []transactionTest{
		{onlyOneDenomation, true},
		{lowNumberOfDenominations, true},
		{lowNumberOfOutputs, true},
		{shouldWork, false},
		{noUncommonDenomation, true},
	}

	for _, c := range cases {
		require.EqualValues(t, !c.shouldFail, isWasabi2Mixing(c.tx))
	}
}

func Test_isWhirlpoolMixing(t *testing.T) {
	type transactionTest struct {
		tx         db.Transaction
		shouldFail bool
	}

	shouldWork := db.Transaction{
		Fee:  new(int64),
		Hash: "a18b0025ca64ce09c5e11cd5a923e4ca8e2b04565d4d233ffcb6bc949b924d1e",
		Inputs: []db.Output{
			{Amount: testhelper.GetPointer[int64](50000000)},
			{Amount: testhelper.GetPointer[int64](50000302)},
			{Amount: testhelper.GetPointer[int64](50000302)},
			{Amount: testhelper.GetPointer[int64](50000302)},
			{Amount: testhelper.GetPointer[int64](50000000)},
		},
		Outputs: []db.Output{
			{Amount: testhelper.GetPointer[int64](50000000), KeyAsm: "1"},
			{Amount: testhelper.GetPointer[int64](50000000), KeyAsm: "2"},
			{Amount: testhelper.GetPointer[int64](50000000), KeyAsm: "3"},
			{Amount: testhelper.GetPointer[int64](50000000), KeyAsm: "4"},
			{Amount: testhelper.GetPointer[int64](50000000), KeyAsm: "5"},
		},
	}

	shouldWork2 := db.Transaction{
		Fee:  new(int64),
		Hash: "a18b0025ca64ce09c5e11cd5a923e4ca8e2b04565d4d233ffcb6bc949b924d1e",
		Inputs: []db.Output{
			{Amount: testhelper.GetPointer[int64](50006655)},
			{Amount: testhelper.GetPointer[int64](50002218)},
			{Amount: testhelper.GetPointer[int64](50005646)},
			{Amount: testhelper.GetPointer[int64](50002218)},
			{Amount: testhelper.GetPointer[int64](50002420)},
		},
		Outputs: []db.Output{
			{Amount: testhelper.GetPointer[int64](50000000), KeyAsm: "1"},
			{Amount: testhelper.GetPointer[int64](50000000), KeyAsm: "2"},
			{Amount: testhelper.GetPointer[int64](50000000), KeyAsm: "3"},
			{Amount: testhelper.GetPointer[int64](50000000), KeyAsm: "4"},
			{Amount: testhelper.GetPointer[int64](50000000), KeyAsm: "5"},
		},
	}

	lowerDenomination := db.Transaction{
		Fee:  new(int64),
		Hash: "a18b0025ca64ce09c5e11cd5a923e4ca8e2b04565d4d233ffcb6bc949b924d1e",
		Inputs: []db.Output{
			{Amount: testhelper.GetPointer[int64](100050)},
			{Amount: testhelper.GetPointer[int64](100000)},
			{Amount: testhelper.GetPointer[int64](100030)},
			{Amount: testhelper.GetPointer[int64](100000)},
			{Amount: testhelper.GetPointer[int64](100000)},
		},
		Outputs: []db.Output{
			{Amount: testhelper.GetPointer[int64](100000), KeyAsm: "1"},
			{Amount: testhelper.GetPointer[int64](100000), KeyAsm: "2"},
			{Amount: testhelper.GetPointer[int64](100000), KeyAsm: "3"},
			{Amount: testhelper.GetPointer[int64](100000), KeyAsm: "4"},
			{Amount: testhelper.GetPointer[int64](100000), KeyAsm: "5"},
		},
	}

	fail := db.Transaction{
		Fee:  new(int64),
		Hash: "a18b0025ca64ce09c5e11cd5a923e4ca8e2b04565d4d233ffcb6bc949b924d1e",
		Inputs: []db.Output{
			{Amount: testhelper.GetPointer[int64](100050)},
			{Amount: testhelper.GetPointer[int64](100000)},
			{Amount: testhelper.GetPointer[int64](100030)},
			{Amount: testhelper.GetPointer[int64](100000)},
			{Amount: testhelper.GetPointer[int64](250000)},
		},
		Outputs: []db.Output{
			{Amount: testhelper.GetPointer[int64](100000), KeyAsm: "1"},
			{Amount: testhelper.GetPointer[int64](100000), KeyAsm: "2"},
			{Amount: testhelper.GetPointer[int64](100000), KeyAsm: "3"},
			{Amount: testhelper.GetPointer[int64](100000), KeyAsm: "4"},
			{Amount: testhelper.GetPointer[int64](100000), KeyAsm: "5"},
		},
	}

	fail2 := db.Transaction{
		Fee:  new(int64),
		Hash: "a18b0025ca64ce09c5e11cd5a923e4ca8e2b04565d4d233ffcb6bc949b924d1e",
		Inputs: []db.Output{
			{Amount: testhelper.GetPointer[int64](100050)},
			{Amount: testhelper.GetPointer[int64](100000)},
			{Amount: testhelper.GetPointer[int64](100030)},
			{Amount: testhelper.GetPointer[int64](100000)},
		},
		Outputs: []db.Output{
			{Amount: testhelper.GetPointer[int64](100000), KeyAsm: "1"},
			{Amount: testhelper.GetPointer[int64](100000), KeyAsm: "2"},
			{Amount: testhelper.GetPointer[int64](100000), KeyAsm: "3"},
			{Amount: testhelper.GetPointer[int64](100000), KeyAsm: "4"},
			{Amount: testhelper.GetPointer[int64](100000), KeyAsm: "5"},
		},
	}

	var cases = []transactionTest{
		{shouldWork, false},
		{shouldWork2, false},
		{lowerDenomination, false},
		{fail, true},
		{fail2, true},
	}

	for _, c := range cases {
		require.EqualValues(t, !c.shouldFail, isWhirlpoolMixing(c.tx))
	}
}

func Test_isWhirlpoolOrigin(t *testing.T) {
	type transactionTest struct {
		tx         db.Transaction
		shouldFail bool
	}

	shouldWork := db.Transaction{
		Fee:  new(int64),
		Hash: "9045ef9690fa9e41c7a541984dcbb2a61947467ca6a21c79c8ca4899bb060230",
		Inputs: []db.Output{
			{Amount: testhelper.GetPointer[int64](2130879553)},
		},
		Outputs: []db.Output{
			{Amount: nil},
			{Amount: testhelper.GetPointer[int64](2500000)},
			{Amount: testhelper.GetPointer[int64](28361985)},
			{Amount: testhelper.GetPointer[int64](50000302)},
			{Amount: testhelper.GetPointer[int64](50000302)},
			{Amount: testhelper.GetPointer[int64](50000302)},
			{Amount: testhelper.GetPointer[int64](50000302)},
			{Amount: testhelper.GetPointer[int64](50000302)},
			{Amount: testhelper.GetPointer[int64](50000302)},
		},
	}

	shouldWork2 := db.Transaction{
		Fee:  new(int64),
		Hash: "9045ef9690fa9e41c7a541984dcbb2a61947467ca6a21c79c8ca4899bb060230",
		Inputs: []db.Output{
			{Amount: testhelper.GetPointer[int64](2130879553)},
		},
		Outputs: []db.Output{
			{Amount: testhelper.GetPointer[int64](0)},
			{Amount: testhelper.GetPointer[int64](2500000)},
			{Amount: testhelper.GetPointer[int64](28361985)},
			{Amount: testhelper.GetPointer[int64](50000302)},
			{Amount: testhelper.GetPointer[int64](50000302)},
			{Amount: testhelper.GetPointer[int64](50000302)},
			{Amount: testhelper.GetPointer[int64](50000302)},
			{Amount: testhelper.GetPointer[int64](50000302)},
			{Amount: testhelper.GetPointer[int64](50000302)},
		},
	}

	shouldWork3 := db.Transaction{
		Fee:  new(int64),
		Hash: "2a3f2423a7964206be75075460a8623f0a9b1fe7fb57c2424602f39ed50a25d9",
		Inputs: []db.Output{
			{Amount: testhelper.GetPointer[int64](1502688613)},
		},
		Outputs: []db.Output{
			{Amount: testhelper.GetPointer[int64](0)},
			{Amount: testhelper.GetPointer[int64](250000)},
			{Amount: testhelper.GetPointer[int64](5010000)},
			{Amount: testhelper.GetPointer[int64](5010000)},
			{Amount: testhelper.GetPointer[int64](5010000)},
			{Amount: testhelper.GetPointer[int64](5010000)},
			{Amount: testhelper.GetPointer[int64](5010000)},
			{Amount: testhelper.GetPointer[int64](1477358853)},
		},
	}

	noNilAmount := db.Transaction{
		Fee:  new(int64),
		Hash: "9045ef9690fa9e41c7a541984dcbb2a61947467ca6a21c79c8ca4899bb060230",
		Inputs: []db.Output{
			{Amount: testhelper.GetPointer[int64](2130879553)},
		},
		Outputs: []db.Output{
			{Amount: testhelper.GetPointer[int64](1)},
			{Amount: testhelper.GetPointer[int64](2500000)},
			{Amount: testhelper.GetPointer[int64](28361985)},
			{Amount: testhelper.GetPointer[int64](50000302)},
			{Amount: testhelper.GetPointer[int64](50000302)},
			{Amount: testhelper.GetPointer[int64](50000302)},
			{Amount: testhelper.GetPointer[int64](50000302)},
			{Amount: testhelper.GetPointer[int64](50000302)},
			{Amount: testhelper.GetPointer[int64](50000302)},
		},
	}

	twoNilAmount := db.Transaction{
		Fee:  new(int64),
		Hash: "9045ef9690fa9e41c7a541984dcbb2a61947467ca6a21c79c8ca4899bb060230",
		Inputs: []db.Output{
			{Amount: testhelper.GetPointer[int64](2130879553)},
		},
		Outputs: []db.Output{
			{Amount: nil},
			{Amount: nil},
			{Amount: testhelper.GetPointer[int64](28361985)},
			{Amount: testhelper.GetPointer[int64](50000302)},
			{Amount: testhelper.GetPointer[int64](50000302)},
			{Amount: testhelper.GetPointer[int64](50000302)},
			{Amount: testhelper.GetPointer[int64](50000302)},
			{Amount: testhelper.GetPointer[int64](50000302)},
			{Amount: testhelper.GetPointer[int64](50000302)},
		},
	}

	mostFrequentAmountNotCloseToDenomation := db.Transaction{
		Fee:  new(int64),
		Hash: "9045ef9690fa9e41c7a541984dcbb2a61947467ca6a21c79c8ca4899bb060230",
		Inputs: []db.Output{
			{Amount: testhelper.GetPointer[int64](2130879553)},
		},
		Outputs: []db.Output{
			{Amount: nil},
			{Amount: testhelper.GetPointer[int64](2500000)},
			{Amount: testhelper.GetPointer[int64](28361985)},
			{Amount: testhelper.GetPointer[int64](5)},
			{Amount: testhelper.GetPointer[int64](5)},
			{Amount: testhelper.GetPointer[int64](5)},
		},
	}

	twoDenominationGroups := db.Transaction{
		Fee:  new(int64),
		Hash: "9045ef9690fa9e41c7a541984dcbb2a61947467ca6a21c79c8ca4899bb060230",
		Inputs: []db.Output{
			{Amount: testhelper.GetPointer[int64](2130879553)},
		},
		Outputs: []db.Output{
			{Amount: nil},
			{Amount: testhelper.GetPointer[int64](2500000)},
			{Amount: testhelper.GetPointer[int64](28361985)},
			{Amount: testhelper.GetPointer[int64](50000302)},
			{Amount: testhelper.GetPointer[int64](50000302)},
			{Amount: testhelper.GetPointer[int64](50000302)},
			{Amount: testhelper.GetPointer[int64](5000302)},
			{Amount: testhelper.GetPointer[int64](5000302)},
			{Amount: testhelper.GetPointer[int64](5000302)},
		},
	}

	noExactDenomation := db.Transaction{
		Fee:  new(int64),
		Hash: "9045ef9690fa9e41c7a541984dcbb2a61947467ca6a21c79c8ca4899bb060230",
		Inputs: []db.Output{
			{Amount: testhelper.GetPointer[int64](2130879553)},
		},
		Outputs: []db.Output{
			{Amount: nil},
			{Amount: testhelper.GetPointer[int64](2500000)},
			{Amount: testhelper.GetPointer[int64](28361985)},
			{Amount: testhelper.GetPointer[int64](50000000)},
			{Amount: testhelper.GetPointer[int64](50000000)},
			{Amount: testhelper.GetPointer[int64](50000000)},
		},
	}

	mostFrequentAmountNotCloseToDenomation2 := db.Transaction{
		Fee:  new(int64),
		Hash: "9045ef9690fa9e41c7a541984dcbb2a61947467ca6a21c79c8ca4899bb060230",
		Inputs: []db.Output{
			{Amount: testhelper.GetPointer[int64](2130879553)},
		},
		Outputs: []db.Output{
			{Amount: nil},
			{Amount: testhelper.GetPointer[int64](2500000)},
			{Amount: testhelper.GetPointer[int64](28361985)},
			{Amount: testhelper.GetPointer[int64](70000000)},
			{Amount: testhelper.GetPointer[int64](70000000)},
			{Amount: testhelper.GetPointer[int64](70000000)},
		},
	}

	lowDenomationCount := db.Transaction{
		Fee:  new(int64),
		Hash: "9045ef9690fa9e41c7a541984dcbb2a61947467ca6a21c79c8ca4899bb060230",
		Inputs: []db.Output{
			{Amount: testhelper.GetPointer[int64](2130879553)},
		},
		Outputs: []db.Output{
			{Amount: nil},
			{Amount: testhelper.GetPointer[int64](2500000)},
			{Amount: testhelper.GetPointer[int64](28361985)},
			{Amount: testhelper.GetPointer[int64](5000302)},
		},
	}

	shouldFail := db.Transaction{
		Fee:  new(int64),
		Hash: "1fc25f4a3b89dda0e021869045aa461092a946a8ca0534780d7c83a7ea7820a8",
		Inputs: []db.Output{
			{Amount: testhelper.GetPointer[int64](8945818)},
		},
		Outputs: []db.Output{
			{Amount: testhelper.GetPointer[int64](8918271)},
			{Amount: testhelper.GetPointer[int64](0)},
			{Amount: testhelper.GetPointer[int64](546)},
		},
	}

	var cases = []transactionTest{
		{shouldFail, true},
		{shouldWork, false},
		{shouldWork2, false},
		{shouldWork3, false},
		{twoDenominationGroups, true},
		{lowDenomationCount, false},
		{noNilAmount, true},
		{twoNilAmount, true},
		{mostFrequentAmountNotCloseToDenomation, true},
		{mostFrequentAmountNotCloseToDenomation2, true},
		{noExactDenomation, true},
	}

	for _, c := range cases {
		require.EqualValues(t, !c.shouldFail, isWhirlpoolOrigin(c.tx))
	}
}

func Test_classifyWhirlpoolOriginTransactions(t *testing.T) {
	tests := []struct {
		origins           []db.Transaction
		originToMixingMap map[string][]string
		want              []db.Transaction
	}{
		{},
		{
			origins: []db.Transaction{{
				UID: "0x1",
				Inputs: []db.Output{
					{Amount: testhelper.GetPointer[int64](2130879553)},
				},
				Outputs: []db.Output{
					{Amount: nil},
					{Amount: testhelper.GetPointer[int64](2500000)},
					{Amount: testhelper.GetPointer[int64](28361985)},
					{Amount: testhelper.GetPointer[int64](50000302)},
					{Amount: testhelper.GetPointer[int64](50000302)},
					{Amount: testhelper.GetPointer[int64](50000302)},
					{Amount: testhelper.GetPointer[int64](50000302)},
					{Amount: testhelper.GetPointer[int64](50000302)},
					{Amount: testhelper.GetPointer[int64](50000302)},
				},
			}},
			want: []db.Transaction{{UID: "0x1", Type: constants.TypeWhirlpoolOrigin}},
		},
		{
			originToMixingMap: map[string][]string{"0x1": {"0x10", "0x11"}, "0x2": {"0x12", "0x13"}},
			want: []db.Transaction{
				{UID: "0x1", Type: constants.TypeWhirlpoolOrigin},
				{UID: "0x2", Type: constants.TypeWhirlpoolOrigin},
				{UID: "0x10", Type: constants.TypeWhirlpoolMixing},
				{UID: "0x11", Type: constants.TypeWhirlpoolMixing},
				{UID: "0x12", Type: constants.TypeWhirlpoolMixing},
				{UID: "0x13", Type: constants.TypeWhirlpoolMixing},
			},
			origins: []db.Transaction{
				{
					UID: "0x1",
					Inputs: []db.Output{
						{Amount: testhelper.GetPointer[int64](2130879553)},
					},
					Outputs: []db.Output{
						{Amount: nil},
						{Amount: testhelper.GetPointer[int64](2500000)},
						{Amount: testhelper.GetPointer[int64](28361985)},
						{Amount: testhelper.GetPointer[int64](50000302)},
						{Amount: testhelper.GetPointer[int64](50000302)},
						{Amount: testhelper.GetPointer[int64](50000302)},
						{Amount: testhelper.GetPointer[int64](50000302)},
						{Amount: testhelper.GetPointer[int64](50000302)},
						{Amount: testhelper.GetPointer[int64](50000302)},
					},
				},
				{
					UID: "0x2",
					Inputs: []db.Output{
						{Amount: testhelper.GetPointer[int64](2130879553)},
					},
					Outputs: []db.Output{
						{Amount: testhelper.GetPointer[int64](0)},
						{Amount: testhelper.GetPointer[int64](2500000)},
						{Amount: testhelper.GetPointer[int64](28361985)},
						{Amount: testhelper.GetPointer[int64](50000302)},
						{Amount: testhelper.GetPointer[int64](50000302)},
						{Amount: testhelper.GetPointer[int64](50000302)},
						{Amount: testhelper.GetPointer[int64](50000302)},
						{Amount: testhelper.GetPointer[int64](50000302)},
						{Amount: testhelper.GetPointer[int64](50000302)},
					},
				},
				{
					UID: "0x3",
					Inputs: []db.Output{
						{Amount: testhelper.GetPointer[int64](2130879553)},
					},
					Outputs: []db.Output{
						{Amount: testhelper.GetPointer[int64](1)},
						{Amount: testhelper.GetPointer[int64](2500000)},
						{Amount: testhelper.GetPointer[int64](28361985)},
						{Amount: testhelper.GetPointer[int64](50000302)},
						{Amount: testhelper.GetPointer[int64](50000302)},
						{Amount: testhelper.GetPointer[int64](50000302)},
						{Amount: testhelper.GetPointer[int64](50000302)},
						{Amount: testhelper.GetPointer[int64](50000302)},
						{Amount: testhelper.GetPointer[int64](50000302)},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		txs := classifyWhirlpoolOriginTransactions(tt.origins, tt.originToMixingMap)
		require.Len(t, txs, len(tt.want))
		for _, tx := range txs {
			require.Contains(t, tt.want, tx)
		}
	}
}

func TestIterate(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	db.SetupDB(t, dbHandle, testhelper.UseBTCPrivacyFile)

	ctx, cancel := db.GetTaskContext()
	defer cancel()

	require.NoError(t, analytics.RemoveTransactionTypeOfAllTransactions(ctx, dbHandle))

	ok, err := Iterate(ctx, dbHandle, 574040, 574040)
	require.NoError(t, err)
	require.True(t, ok)
}
