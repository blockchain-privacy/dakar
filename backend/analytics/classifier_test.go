package analytics

import (
	"backend/blockiterator"
	"backend/constants"
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

var dbHandle = &testhelper.TestDB{IsDirty: true}

func getPointer[number any](n number) *number {
	return &n
}

func TestMain(m *testing.M) {
	testhelper.RunDgraphTests(m, &dbHandle.DB, testhelper.ContainerNameDB)
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
		Fee:  getPointer[int64](minCollateral),
		Hash: "9b6306c63f6f57d23a41a904f2a5d8e41d41623a37bbc03da57813a325c342b2",
		Outputs: []db.Output{
			{Amount: getPointer[int64](minCollateral)},
		},
		Inputs: []db.Output{
			{Amount: getPointer[int64](minCollateral)},
		},
	}

	noFee := db.Transaction{
		Fee:  new(int64),
		Hash: "9b6306c63f6f57d23a41a904f2a5d8e41d41623a37bbc03da57813a325c342b2",
		Outputs: []db.Output{
			{Amount: getPointer[int64](minCollateral)},
		},
		Inputs: []db.Output{
			{Amount: getPointer[int64](minCollateral)},
		},
	}

	multipleInputs := db.Transaction{
		Fee:  getPointer[int64](minCollateral),
		Hash: "9b6306c63f6f57d23a41a904f2a5d8e41d41623a37bbc03da57813a325c342b2",
		Outputs: []db.Output{
			{Amount: getPointer[int64](minCollateral)},
			{Amount: getPointer[int64](minCollateral)},
		},
		Inputs: []db.Output{
			{Amount: getPointer[int64](minCollateral)},
			{Amount: getPointer[int64](minCollateral)},
		},
	}

	bigInput := db.Transaction{
		Fee:  getPointer[int64](minCollateral),
		Hash: "9b6306c63f6f57d23a41a904f2a5d8e41d41623a37bbc03da57813a325c342b2",
		Outputs: []db.Output{
			{Amount: getPointer[int64](500000000000)},
		},
		Inputs: []db.Output{
			{Amount: getPointer[int64](500000000000)},
		},
	}

	smallInput := db.Transaction{
		Fee:  getPointer[int64](minCollateral),
		Hash: "9b6306c63f6f57d23a41a904f2a5d8e41d41623a37bbc03da57813a325c342b2",
		Outputs: []db.Output{
			{Amount: getPointer[int64](1)},
		},
		Inputs: []db.Output{
			{Amount: getPointer[int64](1)},
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
		LastBlockID: getPointer[uint64](5),
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
	db.SetupDB(t, dbHandle, testhelper.UseClassifierFile)

	txHashes := []string{
		"55a6030d087b42682e3c3fdd0605e15ccf0923192fccaa83a6cf42a036d472e4",
		"f44eb76b592c5b16a79fd81277c55306f4db6cb783b01f3fde675867bc8af2b7",
		"15e89abaa5d3062e8b694b06db4d9bd7ebf121527d70559ba3a56d62234c4296"}

	txs := make([]db.Transaction, len(txHashes))
	for i, hash := range txHashes {
		transaction, err := db.GetTransaction(dbHandle, hash)
		require.NoError(t, err)
		transaction.PrivacyType = nil
		txs[i] = transaction
	}

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
				blockHeight:                     1557780,
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
	db.SetupDB(t, dbHandle, testhelper.UseBlockFile)

	no := false
	require.NoError(t, status.SetCrawlerStatus(dbHandle, status.CrawlerStatus{
		IsCrawling:  &no,
		LastBlockID: getPointer[uint64](60020),
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
	db.SetupDB(t, dbHandle, testhelper.UseClassifierFile)

	const firstBlock = 1557778

	no := false
	require.NoError(t, status.SetCrawlerStatus(dbHandle, status.CrawlerStatus{
		IsCrawling: &no,
		// first block of the file
		LastBlockID: getPointer[uint64](firstBlock),
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
		LastClassifiedBlockID: getPointer[uint64](700),
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

func Test_isCollateralCreation(t *testing.T) {
	testhelper.SkipIfNotCI(t)
	db.SetupDB(t, dbHandle, testhelper.UseClassifierFile)

	ccTx, err := db.GetTransaction(dbHandle, "f44eb76b592c5b16a79fd81277c55306f4db6cb783b01f3fde675867bc8af2b7")
	require.NoError(t, err)
	cpTx, err := db.GetTransaction(dbHandle, "8f85c5c61fac409ce4b07c25d51d93dc8bcd1054d5dad3da2c1d7754bdc98d5e")
	require.NoError(t, err)
	unclassifiedTx, err := db.GetTransaction(dbHandle, "c071b12871b6f2b2eaded80e156273a021a95fde407a729fa968afd38e996242")
	require.NoError(t, err)
	mixingTx, err := db.GetTransaction(dbHandle, "6bae9c7d40899c501fdd00c3ff5b6e5dc78687d1ca192fe9afe685ccdcc15389")
	require.NoError(t, err)

	tests := []struct {
		t       db.Transaction
		want    bool
		wantErr bool
	}{
		{t: ccTx, want: true, wantErr: false},
		{t: cpTx, want: false, wantErr: false},
		{t: mixingTx, want: false, wantErr: false},
		{t: unclassifiedTx, want: false, wantErr: false},
	}
	for _, tt := range tests {
		got, err := isCollateralCreation(dbHandle, tt.t)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.EqualValues(t, tt.want, got)
		}
	}
}

func Test_newCollateralCreationTransaction(t *testing.T) {
	cc := constants.PrivacyCollateralCreation
	tests := []struct {
		uid  string
		want db.Transaction
	}{
		{
			uid:  "some_uid",
			want: db.Transaction{UID: "some_uid", PrivacyType: &cc},
		},
		{
			uid:  "some_uid2",
			want: db.Transaction{UID: "some_uid2", PrivacyType: &cc},
		},
	}
	for _, tt := range tests {
		tx := newCollateralCreationTransaction(tt.uid)
		require.Equal(t, tt.want.UID, tx.UID)
		require.Equal(t, *tt.want.PrivacyType, *tx.PrivacyType)
	}
}

func Test_isCollateralPayment(t *testing.T) {
	testhelper.SkipIfNotCI(t)
	db.SetupDB(t, dbHandle, testhelper.UseClassifierFile)

	cp, err := db.GetTransaction(dbHandle, "8f85c5c61fac409ce4b07c25d51d93dc8bcd1054d5dad3da2c1d7754bdc98d5e")
	require.NoError(t, err)
	ccTx, err := db.GetTransaction(dbHandle, "f44eb76b592c5b16a79fd81277c55306f4db6cb783b01f3fde675867bc8af2b7")
	require.NoError(t, err)
	unclassifiedTx, err := db.GetTransaction(dbHandle, "c071b12871b6f2b2eaded80e156273a021a95fde407a729fa968afd38e996242")
	require.NoError(t, err)
	mixingTx, err := db.GetTransaction(dbHandle, "6bae9c7d40899c501fdd00c3ff5b6e5dc78687d1ca192fe9afe685ccdcc15389")
	require.NoError(t, err)

	tests := []struct {
		t    db.Transaction
		want bool
	}{
		{t: cp, want: true},
		{t: ccTx, want: false},
		{t: mixingTx, want: false},
		{t: unclassifiedTx, want: false},
	}
	for _, tt := range tests {
		require.EqualValues(t, tt.want, isCollateralPayment(tt.t))
	}
}

func Test_newCollateralPaymentTransaction(t *testing.T) {
	cp := constants.PrivacyCollateralPayment
	tests := []struct {
		uid  string
		want db.Transaction
	}{
		{
			uid:  "some_uid",
			want: db.Transaction{UID: "some_uid", PrivacyType: &cp},
		},
		{
			uid:  "some_uid2",
			want: db.Transaction{UID: "some_uid2", PrivacyType: &cp},
		},
	}
	for _, tt := range tests {
		tx := newCollateralPaymentTransaction(tt.uid)
		require.Equal(t, tt.want.UID, tx.UID)
		require.Equal(t, *tt.want.PrivacyType, *tx.PrivacyType)
	}
}

func Test_isMixing(t *testing.T) {
	testhelper.SkipIfNotCI(t)
	db.SetupDB(t, dbHandle, testhelper.UseClassifierFile)

	cp, err := db.GetTransaction(dbHandle, "8f85c5c61fac409ce4b07c25d51d93dc8bcd1054d5dad3da2c1d7754bdc98d5e")
	require.NoError(t, err)
	ccTx, err := db.GetTransaction(dbHandle, "f44eb76b592c5b16a79fd81277c55306f4db6cb783b01f3fde675867bc8af2b7")
	require.NoError(t, err)
	unclassifiedTx, err := db.GetTransaction(dbHandle, "c071b12871b6f2b2eaded80e156273a021a95fde407a729fa968afd38e996242")
	require.NoError(t, err)
	mixingTx, err := db.GetTransaction(dbHandle, "6bae9c7d40899c501fdd00c3ff5b6e5dc78687d1ca192fe9afe685ccdcc15389")
	require.NoError(t, err)
	mixingTx2, err := db.GetTransaction(dbHandle, "8a1b7adf54e37a2165f3bfba9df4abd4552a50af703dbd4ba5ba59b0562ded2f")
	require.NoError(t, err)

	tests := []struct {
		t    db.Transaction
		want int
	}{
		{t: cp, want: -1},
		{t: ccTx, want: -1},
		{t: mixingTx, want: 3},
		{t: mixingTx2, want: 4},
		{t: unclassifiedTx, want: -1},
	}
	for _, tt := range tests {
		require.EqualValues(t, tt.want, isMixing(tt.t))
	}
}

func Test_newMixingTransaction(t *testing.T) {
	tests := []struct {
		uid                   string
		denominationTypeIndex int
		want                  db.Transaction
	}{
		{
			uid:                   "some_uid",
			denominationTypeIndex: 3,
			want:                  db.Transaction{UID: "some_uid", PrivacyType: getPointer[constants.PrivacyType](constants.MixingTypes[3])},
		},
		{
			uid:                   "some_uid2",
			denominationTypeIndex: 0,
			want:                  db.Transaction{UID: "some_uid2", PrivacyType: getPointer[constants.PrivacyType](constants.MixingTypes[0])},
		},
	}
	for _, tt := range tests {
		tx := newMixingTransaction(tt.uid, tt.denominationTypeIndex)
		require.Equal(t, tt.want.UID, tx.UID)
		require.EqualValues(t, *tt.want.PrivacyType, *tx.PrivacyType)
	}
}

func Test_newOriginTransaction(t *testing.T) {
	cp := constants.PrivacyOrigin
	tests := []struct {
		uid  string
		want db.Transaction
	}{
		{
			uid:  "some_uid",
			want: db.Transaction{UID: "some_uid", PrivacyType: &cp},
		},
		{
			uid:  "some_uid2",
			want: db.Transaction{UID: "some_uid2", PrivacyType: &cp},
		},
	}
	for _, tt := range tests {
		tx := newOriginTransaction(tt.uid)
		require.Equal(t, tt.want.UID, tx.UID)
		require.Equal(t, *tt.want.PrivacyType, *tx.PrivacyType)
	}
}

func Test_hasValidPrivacyType(t *testing.T) {
	tests := []struct {
		tx   db.Transaction
		want bool
	}{
		{tx: db.Transaction{PrivacyType: nil}, want: false},
		{tx: db.Transaction{PrivacyType: getPointer[constants.PrivacyType](0)}, want: true},
		{tx: db.Transaction{PrivacyType: getPointer[constants.PrivacyType](constants.PrivacyCollateralPaymentLast + 1)},
			want: false},
		{tx: db.Transaction{PrivacyType: getPointer[constants.PrivacyType](5)}, want: true},
	}
	for _, tt := range tests {
		require.Equal(t, tt.want, hasValidPrivacyType(tt.tx))
	}
}

func Test_classifyTransactions(t *testing.T) {
	testhelper.SkipIfNotCI(t)
	db.SetupDB(t, dbHandle, testhelper.UseClassifierFile)

	txHashes := []string{
		"6bae9c7d40899c501fdd00c3ff5b6e5dc78687d1ca192fe9afe685ccdcc15389",
		"55a6030d087b42682e3c3fdd0605e15ccf0923192fccaa83a6cf42a036d472e4",
		"f44eb76b592c5b16a79fd81277c55306f4db6cb783b01f3fde675867bc8af2b7",
		"15e89abaa5d3062e8b694b06db4d9bd7ebf121527d70559ba3a56d62234c4296"}

	txs := make([]db.Transaction, len(txHashes))
	for i, hash := range txHashes {
		transaction, err := db.GetTransaction(dbHandle, hash)
		require.NoError(t, err)
		transaction.PrivacyType = nil
		txs[i] = transaction
	}

	tests := []struct {
		transactions    []db.Transaction
		wantMixingLen   int
		wantOriginCCLen int
		wantOriginCPLen int
		wantErr         bool
	}{
		{
			transactions:    nil,
			wantMixingLen:   0,
			wantOriginCCLen: 0,
			wantOriginCPLen: 0,
			wantErr:         false,
		},
		{
			transactions:    txs,
			wantMixingLen:   1,
			wantOriginCCLen: 1,
			wantOriginCPLen: 2,
			wantErr:         false,
		},
	}
	for _, tt := range tests {
		mixing, cc, cp, err := classifyTransactions(dbHandle, tt.transactions)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.Len(t, mixing, tt.wantMixingLen)
			require.Len(t, cc, tt.wantOriginCCLen)
			require.Len(t, cp, tt.wantOriginCPLen)
		}
	}
}

func TestBlockIterator(t *testing.T) {
	testhelper.SkipIfNotCI(t)
	db.SetupDB(t, dbHandle, testhelper.UseClassifierFile)

	const firstBlock = 1557778

	no := false
	require.NoError(t, status.SetCrawlerStatus(dbHandle, status.CrawlerStatus{
		IsCrawling: &no,
		// let's classify 2 blocks
		LastBlockID: getPointer[uint64](firstBlock + 1),
	}))
	require.NoError(t, status.SetClassifierStatus(dbHandle, status.ClassifierStatus{
		IsClassifying: &no,
		// let's classify 3 blocks
		LastClassifiedBlockID: getPointer[uint64](firstBlock),
	}))

	ctx, cancelFunc := context.WithTimeout(context.Background(), time.Second*20)
	defer cancelFunc()

	require.NoError(t, blockiterator.StartIteration(NewClassifier(ctx, dbHandle, NewDashConfig())))
}
