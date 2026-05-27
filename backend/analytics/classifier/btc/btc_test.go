// SPDX-FileCopyrightText: 2026 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2026 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package btc

import (
	"encoding/json"
	"strconv"
	"testing"

	"github.com/dgraph-io/dgo/v250/protos/api"
	"github.com/stretchr/testify/require"
	"gitlab.com/blockchain-privacy/dakar/constants"
	"gitlab.com/blockchain-privacy/dakar/db"
	"gitlab.com/blockchain-privacy/dakar/db/analytics"
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
			outputs: []db.Output{{Amount: &denom0}, {Amount: &denom1}, {Amount: new(denominationsTypesWasabi2[2])},
				{Amount: new(denominationsTypesWasabi2[3])}, {Amount: new(denominationsTypesWasabi2[4])}},
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

func Test_isWasabi2MixingProperties(t *testing.T) {
	type transactionTest struct {
		tx         db.Transaction
		shouldFail bool
	}

	onlyOneDenomination := db.Transaction{
		Fee:  new(int64),
		Hash: "9b6306c63f6f57d23a41a904f2a5d8e41d41623a37bbc03da57813a325c342b2",
		Inputs: []db.Output{
			{Amount: new(int64(5001))},
			{Amount: new(int64(5002))},
			{Amount: new(int64(5003))},
			{Amount: new(int64(5004))},
			{Amount: new(int64(5005))},
			{Amount: new(int64(5006))},
			{Amount: new(int64(5007))},
			{Amount: new(int64(5008))},
			{Amount: new(int64(5009))},
			{Amount: new(int64(50010))},
			{Amount: new(int64(50011))},
			{Amount: new(int64(50012))},
			{Amount: new(int64(50013))},
			{Amount: new(int64(50014))},
			{Amount: new(int64(50015))},
		},
		Outputs: []db.Output{
			{Amount: new(int64(100000000000))},
			{Amount: new(int64(100000000000))},
			{Amount: new(int64(100000000000))},
			{Amount: new(int64(100000000000))},
			{Amount: new(int64(100000000000))},
			{Amount: new(int64(100000000000))},
			{Amount: new(int64(100000000000))},
			{Amount: new(int64(100000000000))},
			{Amount: new(int64(100000000000))},
			{Amount: new(int64(100000000000))},
		},
	}

	lowNumberOfDenominations := db.Transaction{
		Fee:  new(int64),
		Hash: "9b6306c63f6f57d23a41a904f2a5d8e41d41623a37bbc03da57813a325c342b2",
		Inputs: []db.Output{
			{Amount: new(int64(5001))},
			{Amount: new(int64(5002))},
			{Amount: new(int64(5003))},
			{Amount: new(int64(5004))},
			{Amount: new(int64(5005))},
			{Amount: new(int64(5006))},
			{Amount: new(int64(5007))},
			{Amount: new(int64(5008))},
			{Amount: new(int64(5009))},
			{Amount: new(int64(50010))},
			{Amount: new(int64(50011))},
			{Amount: new(int64(50012))},
			{Amount: new(int64(50013))},
			{Amount: new(int64(50014))},
			{Amount: new(int64(50015))},
		},
		Outputs: []db.Output{
			{Amount: new(int64(100000000000))},
			{Amount: new(int64(100000000000))},
			{Amount: new(int64(100000000000))},
			{Amount: new(int64(1))},
			{Amount: new(int64(2))},
			{Amount: new(int64(3))},
			{Amount: new(int64(4))},
			{Amount: new(int64(5))},
			{Amount: new(int64(6))},
			{Amount: new(int64(7))},
		},
	}

	shouldWork := db.Transaction{
		Fee:  new(int64),
		Hash: "9b6306c63f6f57d23a41a904f2a5d8e41d41623a37bbc03da57813a325c342b2",
		Inputs: []db.Output{
			{Amount: new(int64(5001))},
			{Amount: new(int64(5002))},
			{Amount: new(int64(5003))},
			{Amount: new(int64(5004))},
			{Amount: new(int64(5005))},
			{Amount: new(int64(5006))},
			{Amount: new(int64(5007))},
			{Amount: new(int64(5008))},
			{Amount: new(int64(5009))},
			{Amount: new(int64(50010))},
			{Amount: new(int64(50011))},
			{Amount: new(int64(50012))},
			{Amount: new(int64(50013))},
			{Amount: new(int64(50014))},
			{Amount: new(int64(50015))},
		},
		Outputs: []db.Output{
			{Amount: new(int64(258280326))},
			{Amount: new(int64(4782969))},
			{Amount: new(int64(8388608))},
			{Amount: new(int64(19683))},
			{Amount: new(int64(20000))},
			{Amount: new(int64(1))},
			{Amount: new(int64(2))},
			{Amount: new(int64(3))},
			{Amount: new(int64(4))},
			{Amount: new(int64(5))},
		},
	}

	lowNumberOfOutputs := db.Transaction{
		Fee:  new(int64),
		Hash: "9b6306c63f6f57d23a41a904f2a5d8e41d41623a37bbc03da57813a325c342b2",
		Inputs: []db.Output{
			{Amount: new(int64(5001))},
			{Amount: new(int64(5002))},
		},
		Outputs: []db.Output{
			{Amount: new(int64(258280326))},
			{Amount: new(int64(4782969))},
		},
	}

	noUncommonDenomination := db.Transaction{
		Fee:  new(int64),
		Hash: "f7a484919cee80b71d61a37cb901a6aa6204e7f26535cb7e479079722d571b10",
		Inputs: []db.Output{
			{Amount: new(int64(874642))},
			{Amount: new(int64(874642))},
			{Amount: new(int64(874642))},
			{Amount: new(int64(874642))},
			{Amount: new(int64(874642))},
			{Amount: new(int64(874642))},
			{Amount: new(int64(874642))},
			{Amount: new(int64(874642))},
			{Amount: new(int64(874642))},
			{Amount: new(int64(874642))},
			{Amount: new(int64(874642))},
			{Amount: new(int64(874642))},
			{Amount: new(int64(874642))},
			{Amount: new(int64(874642))},
			{Amount: new(int64(874642))},
			{Amount: new(int64(874642))},
			{Amount: new(int64(874642))},
			{Amount: new(int64(874642))},
			{Amount: new(int64(874642))},
			{Amount: new(int64(874642))},
			{Amount: new(int64(874642))},
			{Amount: new(int64(874642))},
			{Amount: new(int64(874642))},
			{Amount: new(int64(874642))},
			{Amount: new(int64(874642))},
			{Amount: new(int64(874642))},
			{Amount: new(int64(874642))},
			{Amount: new(int64(874642))},
			{Amount: new(int64(874642))},
			{Amount: new(int64(874642))},
		},
		Outputs: []db.Output{
			{Amount: new(int64(133200000))},
			{Amount: new(int64(5000000))},
			{Amount: new(int64(150000000))},
			{Amount: new(int64(3799000000))},
			{Amount: new(int64(100000000))},
			{Amount: new(int64(1700000))},
			{Amount: new(int64(1000000))},
			{Amount: new(int64(2000000))},
			{Amount: new(int64(482500000))},
			{Amount: new(int64(100000000))},
			{Amount: new(int64(234403702357))},
		},
	}

	var cases = []transactionTest{
		{onlyOneDenomination, true},
		{lowNumberOfDenominations, true},
		{lowNumberOfOutputs, true},
		{shouldWork, false},
		{noUncommonDenomination, true},
	}

	for _, c := range cases {
		require.Equal(t, !c.shouldFail, isWasabi2MixingProperties(c.tx))
	}
}

func Test_isWasabi2Mixing(t *testing.T) {
	dbHandle := db.GetDBConnection(t, "")

	t1 := db.Transaction{
		Fee:  new(int64),
		Hash: "A",
		Inputs: []db.Output{
			{Amount: new(int64(5001))},
			{Amount: new(int64(5002))},
			{Amount: new(int64(5003))},
			{Amount: new(int64(5004))},
			{Amount: new(int64(5005))},
			{Amount: new(int64(5006))},
			{Amount: new(int64(5007))},
			{Amount: new(int64(5008))},
			{Amount: new(int64(5009))},
			{Amount: new(int64(50010))},
			{Amount: new(int64(50011))},
			{Amount: new(int64(50012))},
			{Amount: new(int64(50013))},
			{Amount: new(int64(50014))},
			{Amount: new(int64(50015))},
		},
		Outputs: []db.Output{
			{Amount: new(int64(258280326))},
			{Amount: new(int64(4782969))},
			{Amount: new(int64(8388608))},
			{Amount: new(int64(19683))},
			{Amount: new(int64(20000))},
			{Amount: new(int64(1))},
			{Amount: new(int64(2))},
			{Amount: new(int64(3))},
			{Amount: new(int64(4))},
			{Amount: new(int64(5))},
		},
	}

	t2 := t1

	t2.Hash = "B"

	pb, err := json.Marshal([]db.Transaction{t1, t2})
	require.NoError(t, err)

	require.NoError(t, db.MutationWithRetry(t.Context(), dbHandle, &api.Request{Mutations: []*api.Mutation{{SetJson: pb}}, CommitNow: true}))

	ta, err := db.GetTransaction(t.Context(), dbHandle, "A")
	require.NoError(t, err)

	tb, err := db.GetTransaction(t.Context(), dbHandle, "B")
	require.NoError(t, err)

	addresses := make([]db.Address, len(ta.Outputs)+len(tb.Outputs))
	var y int
	for i, output := range ta.Outputs {
		addresses[y] = db.Address{
			UID:     "uid(a" + strconv.Itoa(y) + ")",
			Hash:    "a" + strconv.Itoa(i),
			Outputs: []db.Output{output},
		}
		y++
	}

	for i, output := range tb.Outputs {
		addresses[y] = db.Address{
			UID:     "uid(a" + strconv.Itoa(y) + ")",
			Hash:    "b" + strconv.Itoa(i),
			Outputs: []db.Output{output},
		}
		y++
	}

	// add the output of the last element to the second last, so one address points to two outputs. remove the last element
	addresses[len(addresses)-2].Outputs = append(addresses[len(addresses)-2].Outputs, addresses[len(addresses)-1].Outputs...)
	addresses = addresses[:len(addresses)-1]

	require.NoError(t, db.UpsertAddresses(t.Context(), dbHandle, addresses))

	type transactionTest struct {
		tx         db.Transaction
		shouldFail bool
	}

	var cases = []transactionTest{
		{ta, false},
		{tb, true},
	}

	for _, c := range cases {
		require.Equal(t, !c.shouldFail, isWasabi2Mixing(t.Context(), dbHandle, c.tx), c.tx)
	}
}

func Test_isWhirlpoolMixing(t *testing.T) {
	type transactionTest struct {
		tx         db.Transaction
		shouldFail bool
	}

	shouldWork := db.Transaction{
		Fee:  new(int64),
		Hash: "1",
		Inputs: []db.Output{
			{Amount: new(int64(50000000))},
			{Amount: new(int64(50000302))},
			{Amount: new(int64(50000302))},
			{Amount: new(int64(50000302))},
			{Amount: new(int64(50000000))},
		},
		Outputs: []db.Output{
			{Amount: new(int64(50000000))},
			{Amount: new(int64(50000000))},
			{Amount: new(int64(50000000))},
			{Amount: new(int64(50000000))},
			{Amount: new(int64(50000000))},
		},
	}

	shouldWork2 := db.Transaction{
		Fee:  new(int64),
		Hash: "2",
		Inputs: []db.Output{
			{Amount: new(int64(50006655))},
			{Amount: new(int64(50002218))},
			{Amount: new(int64(50005646))},
			{Amount: new(int64(50002218))},
			{Amount: new(int64(50002420))},
		},
		Outputs: []db.Output{
			{Amount: new(int64(50000000))},
			{Amount: new(int64(50000000))},
			{Amount: new(int64(50000000))},
			{Amount: new(int64(50000000))},
			{Amount: new(int64(50000000))},
		},
	}

	lowerDenomination := db.Transaction{
		Fee:  new(int64),
		Hash: "3",
		Inputs: []db.Output{
			{Amount: new(int64(100050))},
			{Amount: new(int64(100000))},
			{Amount: new(int64(100030))},
			{Amount: new(int64(100000))},
			{Amount: new(int64(100000))},
		},
		Outputs: []db.Output{
			{Amount: new(int64(100000))},
			{Amount: new(int64(100000))},
			{Amount: new(int64(100000))},
			{Amount: new(int64(100000))},
			{Amount: new(int64(100000))},
		},
	}

	fail := db.Transaction{
		Fee:  new(int64),
		Hash: "4",
		Inputs: []db.Output{
			{Amount: new(int64(100050))},
			{Amount: new(int64(100000))},
			{Amount: new(int64(100030))},
			{Amount: new(int64(100000))},
			{Amount: new(int64(250000))},
		},
		Outputs: []db.Output{
			{Amount: new(int64(100000))},
			{Amount: new(int64(100000))},
			{Amount: new(int64(100000))},
			{Amount: new(int64(100000))},
			{Amount: new(int64(100000))},
		},
	}

	fail2 := db.Transaction{
		Fee:  new(int64),
		Hash: "5",
		Inputs: []db.Output{
			{Amount: new(int64(100050))},
			{Amount: new(int64(100000))},
			{Amount: new(int64(100030))},
			{Amount: new(int64(100000))},
		},
		Outputs: []db.Output{
			{Amount: new(int64(100000))},
			{Amount: new(int64(100000))},
			{Amount: new(int64(100000))},
			{Amount: new(int64(100000))},
			{Amount: new(int64(100000))},
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
		require.Equal(t, !c.shouldFail, isWhirlpoolMixing(c.tx))
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
			{Amount: new(int64(2130879553))},
		},
		Outputs: []db.Output{
			{Amount: nil},
			{Amount: new(int64(2500000))},
			{Amount: new(int64(28361985))},
			{Amount: new(int64(50000302))},
			{Amount: new(int64(50000302))},
			{Amount: new(int64(50000302))},
			{Amount: new(int64(50000302))},
			{Amount: new(int64(50000302))},
			{Amount: new(int64(50000302))},
		},
	}

	shouldWork2 := db.Transaction{
		Fee:  new(int64),
		Hash: "9045ef9690fa9e41c7a541984dcbb2a61947467ca6a21c79c8ca4899bb060230",
		Inputs: []db.Output{
			{Amount: new(int64(2130879553))},
		},
		Outputs: []db.Output{
			{Amount: new(int64(0))},
			{Amount: new(int64(2500000))},
			{Amount: new(int64(28361985))},
			{Amount: new(int64(50000302))},
			{Amount: new(int64(50000302))},
			{Amount: new(int64(50000302))},
			{Amount: new(int64(50000302))},
			{Amount: new(int64(50000302))},
			{Amount: new(int64(50000302))},
		},
	}

	shouldWork3 := db.Transaction{
		Fee:  new(int64),
		Hash: "2a3f2423a7964206be75075460a8623f0a9b1fe7fb57c2424602f39ed50a25d9",
		Inputs: []db.Output{
			{Amount: new(int64(1502688613))},
		},
		Outputs: []db.Output{
			{Amount: new(int64(0))},
			{Amount: new(int64(250000))},
			{Amount: new(int64(5010000))},
			{Amount: new(int64(5010000))},
			{Amount: new(int64(5010000))},
			{Amount: new(int64(5010000))},
			{Amount: new(int64(5010000))},
			{Amount: new(int64(1477358853))},
		},
	}

	noNilAmount := db.Transaction{
		Fee:  new(int64),
		Hash: "9045ef9690fa9e41c7a541984dcbb2a61947467ca6a21c79c8ca4899bb060230",
		Inputs: []db.Output{
			{Amount: new(int64(2130879553))},
		},
		Outputs: []db.Output{
			{Amount: new(int64(1))},
			{Amount: new(int64(2500000))},
			{Amount: new(int64(28361985))},
			{Amount: new(int64(50000302))},
			{Amount: new(int64(50000302))},
			{Amount: new(int64(50000302))},
			{Amount: new(int64(50000302))},
			{Amount: new(int64(50000302))},
			{Amount: new(int64(50000302))},
		},
	}

	twoNilAmount := db.Transaction{
		Fee:  new(int64),
		Hash: "9045ef9690fa9e41c7a541984dcbb2a61947467ca6a21c79c8ca4899bb060230",
		Inputs: []db.Output{
			{Amount: new(int64(2130879553))},
		},
		Outputs: []db.Output{
			{Amount: nil},
			{Amount: nil},
			{Amount: new(int64(28361985))},
			{Amount: new(int64(50000302))},
			{Amount: new(int64(50000302))},
			{Amount: new(int64(50000302))},
			{Amount: new(int64(50000302))},
			{Amount: new(int64(50000302))},
			{Amount: new(int64(50000302))},
		},
	}

	mostFrequentAmountNotCloseToDenomination := db.Transaction{
		Fee:  new(int64),
		Hash: "9045ef9690fa9e41c7a541984dcbb2a61947467ca6a21c79c8ca4899bb060230",
		Inputs: []db.Output{
			{Amount: new(int64(2130879553))},
		},
		Outputs: []db.Output{
			{Amount: nil},
			{Amount: new(int64(2500000))},
			{Amount: new(int64(28361985))},
			{Amount: new(int64(5))},
			{Amount: new(int64(5))},
			{Amount: new(int64(5))},
		},
	}

	twoDenominationGroups := db.Transaction{
		Fee:  new(int64),
		Hash: "9045ef9690fa9e41c7a541984dcbb2a61947467ca6a21c79c8ca4899bb060230",
		Inputs: []db.Output{
			{Amount: new(int64(2130879553))},
		},
		Outputs: []db.Output{
			{Amount: nil},
			{Amount: new(int64(2500000))},
			{Amount: new(int64(28361985))},
			{Amount: new(int64(50000302))},
			{Amount: new(int64(50000302))},
			{Amount: new(int64(50000302))},
			{Amount: new(int64(5000302))},
			{Amount: new(int64(5000302))},
			{Amount: new(int64(5000302))},
		},
	}

	noExactDenomination := db.Transaction{
		Fee:  new(int64),
		Hash: "9045ef9690fa9e41c7a541984dcbb2a61947467ca6a21c79c8ca4899bb060230",
		Inputs: []db.Output{
			{Amount: new(int64(2130879553))},
		},
		Outputs: []db.Output{
			{Amount: nil},
			{Amount: new(int64(2500000))},
			{Amount: new(int64(28361985))},
			{Amount: new(int64(50000000))},
			{Amount: new(int64(50000000))},
			{Amount: new(int64(50000000))},
		},
	}

	mostFrequentAmountNotCloseToDenomination2 := db.Transaction{
		Fee:  new(int64),
		Hash: "9045ef9690fa9e41c7a541984dcbb2a61947467ca6a21c79c8ca4899bb060230",
		Inputs: []db.Output{
			{Amount: new(int64(2130879553))},
		},
		Outputs: []db.Output{
			{Amount: nil},
			{Amount: new(int64(2500000))},
			{Amount: new(int64(28361985))},
			{Amount: new(int64(70000000))},
			{Amount: new(int64(70000000))},
			{Amount: new(int64(70000000))},
		},
	}

	lowDenominationCount := db.Transaction{
		Fee:  new(int64),
		Hash: "9045ef9690fa9e41c7a541984dcbb2a61947467ca6a21c79c8ca4899bb060230",
		Inputs: []db.Output{
			{Amount: new(int64(2130879553))},
		},
		Outputs: []db.Output{
			{Amount: nil},
			{Amount: new(int64(2500000))},
			{Amount: new(int64(28361985))},
			{Amount: new(int64(5000302))},
		},
	}

	shouldFail := db.Transaction{
		Fee:  new(int64),
		Hash: "1fc25f4a3b89dda0e021869045aa461092a946a8ca0534780d7c83a7ea7820a8",
		Inputs: []db.Output{
			{Amount: new(int64(8945818))},
		},
		Outputs: []db.Output{
			{Amount: new(int64(8918271))},
			{Amount: new(int64(0))},
			{Amount: new(int64(546))},
		},
	}

	var cases = []transactionTest{
		{shouldFail, true},
		{shouldWork, false},
		{shouldWork2, false},
		{shouldWork3, false},
		{twoDenominationGroups, true},
		{lowDenominationCount, false},
		{noNilAmount, true},
		{twoNilAmount, true},
		{mostFrequentAmountNotCloseToDenomination, true},
		{mostFrequentAmountNotCloseToDenomination2, true},
		{noExactDenomination, true},
	}

	for _, c := range cases {
		require.Equal(t, !c.shouldFail, isWhirlpoolOrigin(c.tx))
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
					{Amount: new(int64(2130879553))},
				},
				Outputs: []db.Output{
					{Amount: nil},
					{Amount: new(int64(2500000))},
					{Amount: new(int64(28361985))},
					{Amount: new(int64(50000302))},
					{Amount: new(int64(50000302))},
					{Amount: new(int64(50000302))},
					{Amount: new(int64(50000302))},
					{Amount: new(int64(50000302))},
					{Amount: new(int64(50000302))},
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
						{Amount: new(int64(2130879553))},
					},
					Outputs: []db.Output{
						{Amount: nil},
						{Amount: new(int64(2500000))},
						{Amount: new(int64(28361985))},
						{Amount: new(int64(50000302))},
						{Amount: new(int64(50000302))},
						{Amount: new(int64(50000302))},
						{Amount: new(int64(50000302))},
						{Amount: new(int64(50000302))},
						{Amount: new(int64(50000302))},
					},
				},
				{
					UID: "0x2",
					Inputs: []db.Output{
						{Amount: new(int64(2130879553))},
					},
					Outputs: []db.Output{
						{Amount: new(int64(0))},
						{Amount: new(int64(2500000))},
						{Amount: new(int64(28361985))},
						{Amount: new(int64(50000302))},
						{Amount: new(int64(50000302))},
						{Amount: new(int64(50000302))},
						{Amount: new(int64(50000302))},
						{Amount: new(int64(50000302))},
						{Amount: new(int64(50000302))},
					},
				},
				{
					UID: "0x3",
					Inputs: []db.Output{
						{Amount: new(int64(2130879553))},
					},
					Outputs: []db.Output{
						{Amount: new(int64(1))},
						{Amount: new(int64(2500000))},
						{Amount: new(int64(28361985))},
						{Amount: new(int64(50000302))},
						{Amount: new(int64(50000302))},
						{Amount: new(int64(50000302))},
						{Amount: new(int64(50000302))},
						{Amount: new(int64(50000302))},
						{Amount: new(int64(50000302))},
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
	dbHandle := db.GetDBConnection(t, db.UseBTCPrivacyFile)

	ctx, cancel := db.GetTaskContext()
	defer cancel()

	require.NoError(t, analytics.RemoveTransactionTypeOfAllTransactions(ctx, dbHandle))

	ok, err := Iterate(ctx, dbHandle, 574040, 574040)
	require.NoError(t, err)
	require.True(t, ok)
}

func TestCountAmountWhirlpoolFuzzyDenominations(t *testing.T) {
	type testCase struct {
		amounts []int64
		result  [NumWhirlpoolDenominations]int
	}

	// with fee
	var cases = []testCase{
		{
			amounts: []int64{1, 2, 0, 4, 0},
			result:  [NumWhirlpoolDenominations]int{0, 0, 0, 0},
		},
		{
			amounts: []int64{100105, 100105, 100105},
			result:  [NumWhirlpoolDenominations]int{3, 0, 0, 0},
		},
		{
			amounts: []int64{1000105, 1000105, 1000105, 6, 9, -1},
			result:  [NumWhirlpoolDenominations]int{0, 3, 0, 0},
		},
		{
			amounts: []int64{100105, 1000105, 5000105, 50000105},
			result:  [NumWhirlpoolDenominations]int{1, 1, 1, 1},
		},
	}

	for _, c := range cases {
		require.Equal(t, c.result, CountAmountWhirlpoolFuzzyDenominations(c.amounts, 100))
	}
	// without fee
	cases = []testCase{
		{
			amounts: []int64{100000, 100000, 100000},
			result:  [NumWhirlpoolDenominations]int{3, 0, 0, 0},
		},
		{
			amounts: []int64{1000000, 1000000, 1000000, 6, 9, -1},
			result:  [NumWhirlpoolDenominations]int{0, 3, 0, 0},
		},
		{
			amounts: []int64{100000, 1000000, 5000000, 50000000},
			result:  [NumWhirlpoolDenominations]int{1, 1, 1, 1},
		},
	}

	for _, c := range cases {
		require.Equal(t, c.result, CountAmountWhirlpoolFuzzyDenominations(c.amounts, 0))
	}
}

func TestCountAmountWhirlpoolDenominations(t *testing.T) {
	type testCase struct {
		amounts []int64
		result  [NumWhirlpoolDenominations]int
	}

	// with fee
	var cases = []testCase{
		{
			amounts: []int64{1, 2, 0, 4, 0},
			result:  [NumWhirlpoolDenominations]int{0, 0, 0, 0},
		},
		{
			amounts: []int64{100105, 100105, 100105},
			result:  [NumWhirlpoolDenominations]int{0, 0, 0, 0},
		},
		{
			amounts: []int64{1000105, 1000105, 1000105, 6, 9, -1},
			result:  [NumWhirlpoolDenominations]int{0, 0, 0, 0},
		},
		{
			amounts: []int64{100105, 1000105, 5000105, 50000105},
			result:  [NumWhirlpoolDenominations]int{0, 0, 0, 0},
		},
	}

	for _, c := range cases {
		require.Equal(t, c.result, CountAmountWhirlpoolDenominations(c.amounts))
	}
	// without fee
	cases = []testCase{
		{
			amounts: []int64{100000, 100000, 100000},
			result:  [NumWhirlpoolDenominations]int{3, 0, 0, 0},
		},
		{
			amounts: []int64{1000000, 1000000, 1000000, 6, 9, -1},
			result:  [NumWhirlpoolDenominations]int{0, 3, 0, 0},
		},
		{
			amounts: []int64{100000, 1000000, 5000000, 50000000},
			result:  [NumWhirlpoolDenominations]int{1, 1, 1, 1},
		},
	}

	for _, c := range cases {
		require.Equal(t, c.result, CountAmountWhirlpoolDenominations(c.amounts))
	}
}
