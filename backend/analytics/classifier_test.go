package analytics

import (
	"backend/db"
	"backend/db/status"
	"backend/external"
	"backend/testhelper"
	"context"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

var dbHandle external.Database

const blockFileName = "../db/testdata/blocks_60000_60020.json"
const classificationFile = "../db/testdata/blocks_1649985_1650050.json"

func getNumPointer[number int64 | uint64 | uint32](n number) *number {
	return &n
}

func TestMain(m *testing.M) {
	db.RunDgraphTests(m, &dbHandle, db.ContainerNameAnalytics)
}

func TestIsMixing(t *testing.T) {
	type transactionTest struct {
		tx         db.Transaction
		shouldFail bool
	}

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
	type transactionTest struct {
		tx         db.Transaction
		shouldFail bool
	}

	shouldWork1 := db.Transaction{
		Fee:  getNumPointer[int64](minCollateral),
		Hash: "9b6306c63f6f57d23a41a904f2a5d8e41d41623a37bbc03da57813a325c342b2",
		Outputs: []db.Output{
			{Amount: getNumPointer[int64](minCollateral)},
		},
		Inputs: []db.Output{
			{Amount: getNumPointer[int64](minCollateral)},
		},
	}

	noFee := db.Transaction{
		Fee:  new(int64),
		Hash: "9b6306c63f6f57d23a41a904f2a5d8e41d41623a37bbc03da57813a325c342b2",
		Outputs: []db.Output{
			{Amount: getNumPointer[int64](minCollateral)},
		},
		Inputs: []db.Output{
			{Amount: getNumPointer[int64](minCollateral)},
		},
	}

	multipleInputs := db.Transaction{
		Fee:  getNumPointer[int64](minCollateral),
		Hash: "9b6306c63f6f57d23a41a904f2a5d8e41d41623a37bbc03da57813a325c342b2",
		Outputs: []db.Output{
			{Amount: getNumPointer[int64](minCollateral)},
			{Amount: getNumPointer[int64](minCollateral)},
		},
		Inputs: []db.Output{
			{Amount: getNumPointer[int64](minCollateral)},
			{Amount: getNumPointer[int64](minCollateral)},
		},
	}

	bigInput := db.Transaction{
		Fee:  getNumPointer[int64](minCollateral),
		Hash: "9b6306c63f6f57d23a41a904f2a5d8e41d41623a37bbc03da57813a325c342b2",
		Outputs: []db.Output{
			{Amount: getNumPointer[int64](500000000000)},
		},
		Inputs: []db.Output{
			{Amount: getNumPointer[int64](500000000000)},
		},
	}

	smallInput := db.Transaction{
		Fee:  getNumPointer[int64](minCollateral),
		Hash: "9b6306c63f6f57d23a41a904f2a5d8e41d41623a37bbc03da57813a325c342b2",
		Outputs: []db.Output{
			{Amount: getNumPointer[int64](1)},
		},
		Inputs: []db.Output{
			{Amount: getNumPointer[int64](1)},
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

// unregisterCollectors unregisters all collectors of the classifier.
// This is needed because collectors can not be registered twice with the same default config.
func unregisterCollectors(c *Classifier) {
	if c == nil {
		return
	}

	prometheus.Unregister(c.blocks)
	prometheus.Unregister(c.transactions)
	prometheus.Unregister(c.blockHeight)
}

func TestNewClassifier(t *testing.T) {
	classifier := NewClassifier(context.Background(), nil, Config{
		ClassifierStartAfterBlock: 0,
		BlockchainName:            "Dash",
		IsHeuristicWorkerEnabled:  false,
		IsClassifyingEnabled:      false,
		IsHMIClusteringEnabled:    false,
		IsFMIClusteringEnabled:    false,
	})
	unregisterCollectors(classifier)
	require.NotNil(t, classifier)
}

func TestClassifier_Name(t *testing.T) {
	classifier := NewClassifier(context.Background(), nil, Config{})
	unregisterCollectors(classifier)
	require.NotEmpty(t, classifier.Name())
}

func TestClassifier_Logger(t *testing.T) {
	classifier := NewClassifier(context.Background(), nil, Config{})
	unregisterCollectors(classifier)
	require.NotNil(t, classifier.Logger())
}

func TestClassifier_Context(t *testing.T) {
	classifier := NewClassifier(context.Background(), nil, Config{})
	unregisterCollectors(classifier)
	require.NotNil(t, classifier.Context())
}

func TestClassifier_IncrementState(t *testing.T) {
	classifier := NewClassifier(context.Background(), nil, Config{})
	unregisterCollectors(classifier)

	for i := 0; i < 100; i++ {
		require.NoError(t, classifier.IncrementState())
	}

	require.EqualValues(t, 100, classifier.state.ID)
}

func TestClassifier_Empty(t *testing.T) {
	classifier := NewClassifier(context.Background(), nil, Config{})
	unregisterCollectors(classifier)

	require.False(t, classifier.Empty())
	require.NoError(t, classifier.IncrementState())
	require.True(t, classifier.Empty())
}

func TestClassifier_CalculateInitialState(t *testing.T) {
	testhelper.SkipIfNotCI(t)
	db.SetupDBWithoutData(t, dbHandle)

	classifier := NewClassifier(context.Background(), nil, Config{})
	unregisterCollectors(classifier)

	require.Error(t, classifier.CalculateInitialState())

	classifier.config.IsClassifyingEnabled = true

	// panics because database is nil
	require.Panics(t, func() {
		_ = classifier.CalculateInitialState()
	})

	classifier.db = dbHandle

	// status not set yet
	require.Error(t, classifier.CalculateInitialState())

	yes := true
	require.NoError(t, status.SetCrawlerStatus(dbHandle, status.CrawlerStatus{
		IsCrawling:  &yes,
		LastBlockID: getNumPointer[uint64](5),
	}))

	require.NoError(t, classifier.CalculateInitialState())
	require.EqualValues(t, 5, classifier.state.Top)
	require.EqualValues(t, 1, classifier.state.ID)
}

func Test_getUids(t *testing.T) {
	type args struct {
		txs []db.Transaction
	}
	tests := []struct {
		args args
		want []string
	}{
		{
			args: args{txs: nil},
			want: nil,
		},
		{
			args: args{txs: []db.Transaction{{UID: "some_uid1"}, {UID: "some_uid2"}}},
			want: []string{"some_uid1", "some_uid2"},
		},
		{
			args: args{txs: []db.Transaction{{UID: "some_uid"}}},
			want: []string{"some_uid"},
		},
	}
	for _, tt := range tests {
		require.Len(t, getUids(tt.args.txs), len(tt.want))
	}
}

func Test_getConnectedCollaterals(t *testing.T) {
	testhelper.SkipIfNotCI(t)
	db.SetupDB(t, dbHandle, classificationFile)

	transactions, err := db.GetTransactionByBlock(dbHandle, 1650016)
	require.NoError(t, err)

	var wantedTx = map[string]bool{
		"8508e32dbd5e6ae6fbf3a1ca47e967ca1754fdc64d4ae00de27d32a891b9365b": true,
		"6e0de143dbdd5544be262067dbc3d6f9767b3a4a725f12034c222236e10c0f1e": true,
		"921d30bb00c2f27c655f915a2486d674d4873170786f8e5774de6029f34726c0": true,
	}

	i := 0
	var txs = make([]db.Transaction, len(wantedTx))
	for _, tx := range transactions {
		if wantedTx[tx.Hash] {
			tx.PrivacyType = nil
			txs[i] = tx
			i++
		}
	}

	// transaction must be in returned set
	require.Len(t, txs, len(wantedTx))

	type args struct {
		dgraph                          external.Database
		potentialCollateralTransactions []db.Transaction
		blockHeight                     uint64
	}
	tests := []struct {
		args            args
		wantOriginCCLen int
		wantOriginCPLen int
		wantErr         bool
	}{
		{
			args: args{
				dgraph:                          nil,
				potentialCollateralTransactions: nil,
				blockHeight:                     0,
			},
			wantOriginCCLen: 0,
			wantOriginCPLen: 0,
			wantErr:         false,
		},
		{
			args: args{
				dgraph:                          dbHandle,
				potentialCollateralTransactions: txs,
				blockHeight:                     1650044,
			},
			wantOriginCCLen: 1,
			wantOriginCPLen: 2,
			wantErr:         false,
		},
	}
	for _, tt := range tests {
		gotOriginCC, gotOriginCP, err := getConnectedCollaterals(tt.args.dgraph,
			tt.args.potentialCollateralTransactions, tt.args.blockHeight)

		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.Len(t, gotOriginCC, tt.wantOriginCCLen)
			require.Len(t, gotOriginCP, tt.wantOriginCPLen)
		}
	}
}

func TestClassifier_NextBlock(t *testing.T) {
	testhelper.SkipIfNotCI(t)
	db.SetupDB(t, dbHandle, blockFileName)

	no := false
	require.NoError(t, status.SetCrawlerStatus(dbHandle, status.CrawlerStatus{
		IsCrawling:  &no,
		LastBlockID: getNumPointer[uint64](60020),
	}))

	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*20)
	defer cancelFunc()

	classifier := NewClassifier(ctx, dbHandle, NewDashConfig())
	unregisterCollectors(classifier)

	const firstBlock = 60000
	const lastBlock = 60020
	// set to first available block
	classifier.state.ID = firstBlock
	classifier.state.Top = firstBlock

	got, err := classifier.NextBlock()
	require.NoError(t, err)
	require.True(t, got)
	require.EqualValues(t, lastBlock, classifier.state.Top)
}

func TestClassifier_CurrentBlock(t *testing.T) {
	classifier := NewClassifier(context.Background(), nil, NewDashConfig())
	unregisterCollectors(classifier)

	require.EqualValues(t, 0, classifier.CurrentBlock())
}

func TestClassifier_Iterate(t *testing.T) {
	testhelper.SkipIfNotCI(t)
	db.SetupDB(t, dbHandle, classificationFile)

	const firstBlock = 1649985

	no := false
	require.NoError(t, status.SetCrawlerStatus(dbHandle, status.CrawlerStatus{
		IsCrawling: &no,
		// first block of the file
		LastBlockID: getNumPointer[uint64](firstBlock),
	}))

	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*20)
	defer cancelFunc()

	classifier := NewClassifier(ctx, dbHandle, NewDashConfig())
	unregisterCollectors(classifier)

	// state is set to block 0, which does not exist in database
	_, err := classifier.Iterate()
	require.Error(t, err)

	classifier.state.ID = firstBlock
	classifier.state.Top = firstBlock

	_, err = classifier.Iterate()
	require.NoError(t, err)
}

func TestClassifier_PostExecution(t *testing.T) {
	testhelper.SkipIfNotCI(t)
	db.SetupDBWithoutData(t, dbHandle)

	classifier := NewClassifier(context.Background(), dbHandle, NewDashConfig())
	unregisterCollectors(classifier)

	require.NoError(t, classifier.PostExecution())
}

func Test_setInitialClassifierID(t *testing.T) {
	testhelper.SkipIfNotCI(t)
	db.SetupDBWithoutData(t, dbHandle)
	yes := true
	require.NoError(t, status.SetClassifierStatus(dbHandle, status.ClassifierStatus{
		IsClassifying:         &yes,
		LastClassifiedBlockID: getNumPointer[uint64](700),
	}))

	tests := []struct {
		startBlockClassifier uint64
		wantErr              bool
	}{
		{
			startBlockClassifier: 0,
			wantErr:              false,
		},
		{
			startBlockClassifier: 10000,
			wantErr:              false,
		},
		{
			startBlockClassifier: 500,
			wantErr:              false,
		},
		{
			startBlockClassifier: 1,
			wantErr:              false,
		},
	}
	for _, tt := range tests {
		err := setInitialClassifierID(dbHandle, tt.startBlockClassifier)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}
