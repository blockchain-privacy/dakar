package processor

import (
	"backend/db"
	"backend/external"
	"backend/testhelper"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/integration/rpctest"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"log"
	"sync"
	"testing"
	"time"
)

var (
	dbHandle external.Database
	client   *rpcclient.Client
	// batch client not needed for now
	// batchClient *rpcclient.Client
)

const blockFileName = "../db/testdata/blocks_60000_60020.json"

func TestMain(m *testing.M) {
	if !testhelper.IsCIActive() {
		m.Run()
		return
	}

	// create dgraph client
	graphDB, c, err := external.CreateClient(string(testhelper.ContainerNameProcessor) + ":9080")
	if err != nil {
		log.Panic(err)
		return
	}
	defer func(c *grpc.ClientConn) {
		err := c.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(c)

	if !external.WaitForDatabase(graphDB) {
		log.Panic("Could not connect to database", err)
		return
	}

	// create test harness. Automatic build of btcd is not working somehow, so it is built at the CI stage
	harness, err := rpctest.New(&chaincfg.SimNetParams, nil, []string{"--rejectnonstd", "--txindex"}, "btcd")
	if err != nil {
		log.Panic("unable to create primary harness: ", err)
		return
	}

	defer func(harness *rpctest.Harness) {
		_ = harness.TearDown()
	}(harness)

	// Initialize the primary mining node with a chain of length 125,
	// providing 25 mature coinbases to allow spending from for testing
	// purposes.
	if err := harness.SetUp(true, 25); err != nil {
		log.Panic("unable to setup test chain: ", err)
		return
	}

	dbHandle = graphDB

	client = harness.Client
	// batch client not needed for now
	// batchClient = harness.BatchClient

	m.Run()
}

func TestIncrementProcessingState(t *testing.T) {
	const (
		firstHash   = "000000003b36901a4771aebad94ab2707e55b19ba62898bedcea9a69265f8e7"
		secondHash  = "00000000251b4f191d09553f115383e12108fdf98d3d77530a9e96bc9dd6dd6a"
		toLongHash  = "000000000251b4f191d09553f115383e12108fdf98d3d77530a9e96bc9dd6dd6a"
		invalidHash = "."
	)

	var p crawlerState

	err := p.increment(firstHash)
	require.Nil(t, err)
	require.NotEmpty(t, p.String())

	if p.id != 1 || p.hash != firstHash {
		t.Fatal("incrementation not successful")
	}

	err = p.increment(secondHash)
	require.Nil(t, err)
	if p.id != 2 || p.hash != secondHash {
		t.Fatal("incrementation not successful")
	}

	p2 := p

	err = p2.increment(toLongHash)
	require.NotNil(t, err)

	err = p.increment(invalidHash)
	require.NotNil(t, err)
}

func TestAddOutputToMapping(t *testing.T) {
	outputMappings := make(map[string]outputMapping)
	const (
		firstAddress  = "XtUa1xzS8rr4UMv1bopfTKipspwrUvaBMp"
		secondAddress = "Xhbf3Cj7YVih5Ze8WCgPacRR7oVsaJqRJ3"
	)
	outputMappings = addOutputToMapping(outputMappings, firstAddress, 0)
	if val, ok := outputMappings[firstAddress]; ok {
		if len(val.indexes) != 1 {
			t.Fatal("wrong length of ids")
		} else if val.hash != firstAddress {
			t.Fatal("wrong hash")
		} else if val.indexes[0] != 0 {
			t.Fatal("wrong id")
		}
	} else {
		t.Fatal("Error getting address mapping")
	}

	if len(outputMappings) != 1 {
		t.Fatal("wrong length of output mapping")
	}

	outputMappings = addOutputToMapping(outputMappings, secondAddress, 10)
	if val, ok := outputMappings[secondAddress]; ok {
		if len(val.indexes) != 1 {
			t.Fatal("wrong length of ids")
		} else if val.hash != secondAddress {
			t.Fatal("wrong hash")
		} else if val.indexes[0] != 10 {
			t.Fatal("wrong id")
		}
	} else {
		t.Fatal("Error getting address mapping")
	}

	if len(outputMappings) != 2 {
		t.Fatal("wrong length of output mapping")
	}

	outputMappings = addOutputToMapping(outputMappings, firstAddress, 5)
	if val, ok := outputMappings[firstAddress]; ok {
		if len(val.indexes) != 2 {
			t.Fatal("wrong length of ids")
		} else if val.hash != firstAddress {
			t.Fatal("wrong hash")
		} else if val.indexes[0] != 0 || val.indexes[1] != 5 {
			t.Fatal("wrong id")
		}
	} else {
		t.Fatal("Error getting address mapping")
	}

	if len(outputMappings) != 2 {
		t.Fatal("wrong length of output mapping")
	}
}

func TestAddOutputsToAddresses(t *testing.T) {
	addresses := make(map[string]db.Address)
	cases := []struct {
		address        string
		uids           []string
		requiredLength int
	}{
		{
			address:        "a",
			uids:           []string{"o1", "o2", "o3"},
			requiredLength: 1,
		},
		{
			address:        "b",
			uids:           []string{"o1", "o2", "o3"},
			requiredLength: 2,
		},
		{
			address:        "c",
			uids:           []string{"o1", "o2", "o3"},
			requiredLength: 3,
		},
		{
			address:        "a",
			uids:           []string{"o1", "o2", "o3"},
			requiredLength: 3,
		},
	}

	for _, c := range cases {
		addOutputsToAddresses(addresses, c.address, c.uids)
		require.Len(t, addresses, c.requiredLength)
	}
}

func TestCreateOutputUid(t *testing.T) {
	outputUID := createOutputUID("asdf", 50)
	require.NotEmpty(t, outputUID)
	if len(outputUID) < 2 {
		t.Fatal("output uid is too short:", outputUID)
	}
	require.EqualValues(t, "_:", outputUID[:2])
}

func TestProcessAddresses(t *testing.T) {
	testhelper.SkipIfNotCI(t)
	db.SetupDB(t, dbHandle, blockFileName)

	// calling with empty mapping is allowed
	require.NoError(t, processAddresses(dbHandle, nil, nil))

	// cache is necessary if mapping is not empty
	require.Error(t, processAddresses(dbHandle, nil, []transactionMapping{{}}))

	const (
		fistAddress   = "XonqFxADHJxSwZCuka5h46HXAdFfBMQc21"
		secondAddress = "XvdH1vasQtDv7LvQuD2u124ibKFwNsPFv9"
		txHash        = "fd89e6e3bb0968da20d0253dbddb9e8634bc97e1f173b7c497e0c61e7231398b"
	)

	mapping, err := db.GetTransactionsOutputs(dbHandle, []string{txHash})
	require.NoError(t, err)
	require.Len(t, mapping, 1)
	require.Len(t, mapping[0].Outputs, 2)

	txMap := transactionMapping{
		hash: txHash,
		outputs: map[string]outputMapping{
			fistAddress: {
				hash:    fistAddress,
				indexes: []uint32{0},
			},
			secondAddress: {
				hash:    secondAddress,
				indexes: []uint32{1},
			},
		},
	}

	var outputs [2]db.Output
	outputs[0] = mapping[0].Outputs[0]
	outputs[1] = mapping[0].Outputs[1]

	cache := newOutputCache()
	require.NoError(t, cache.setOutputs(txHash, outputs[:]))

	require.NoError(t, processAddresses(dbHandle, cache, []transactionMapping{txMap}))
}

func TestWaitForNextRPCBlock(t *testing.T) {
	testhelper.SkipIfNotCI(t)
	interrupt := make(chan struct{})
	cfg := NewDashConfig()
	// for a fast test
	cfg.NewBlockIntervalTime = 1

	blkCount, err := client.GetBlockCount()
	require.NoError(t, err)
	// add two blocks, so the first block has a reference to the next block
	hashes, err := client.Generate(2)
	require.NoError(t, err)

	// normal operation
	currentBlock, wasInterrupted, err := waitForNextRPCBlock(client, interrupt, hashes[0], uint64(blkCount), cfg)
	require.NoError(t, err)
	require.False(t, wasInterrupted, "the interrupt flag should have been false")
	require.NotNil(t, currentBlock)

	// missing hash
	currentBlock, wasInterrupted, err = waitForNextRPCBlock(client, interrupt, nil, uint64(blkCount), cfg)
	require.Error(t, err)
	require.False(t, wasInterrupted, "the interrupt flag should have been false")
	require.Nil(t, currentBlock)
	var test struct{}

	go func() {
		interrupt <- test
	}()

	// normal operation but interrupted and higher block
	// count as available, so it must wait or in this case get interrupted
	cfg.NewBlockIntervalTime = time.Minute
	currentBlock, wasInterrupted, err = waitForNextRPCBlock(client, interrupt, hashes[0], uint64(blkCount+2), cfg)
	require.NoError(t, err)
	require.True(t, wasInterrupted, "the interrupt flag should have been true")
	require.Nil(t, currentBlock)
}

func TestGetRPCNumberOfBlocks(t *testing.T) {
	testhelper.SkipIfNotCI(t)
	numBlocks, err := getRPCNumberOfBlocks(client)
	require.NoError(t, err)
	require.NotZerof(t, numBlocks, "number of blocks should not be zero")
}

func Test_crawlerState_String(t *testing.T) {
	state := crawlerState{}
	require.NotEmpty(t, state.String())
	state.id = 1
	state.hash = "asdf"
	require.NotEmpty(t, state.String())
}

func Test_crawlerState_increment(t *testing.T) {
	state := crawlerState{}
	require.NoError(t, state.increment(""))
	require.EqualValues(t, 0, state.id)

	require.Error(t, state.increment("asdf"))
	require.EqualValues(t, 0, state.id)

	require.NoError(t, state.increment("000007248b1005ffdcf3f41f3a5630b5cb0078ca5733d931223839821f7f5faa"))
	require.EqualValues(t, 1, state.id)
}

func getPointer[number any](n number) *number {
	return &n
}

func Test_buildAddresses(t *testing.T) {
	oCache := newOutputCache()
	err := oCache.setOutputs("asdf", []db.Output{
		{OutputIndex: getPointer[uint32](1)},
		{OutputIndex: getPointer[uint32](2)},
		{OutputIndex: getPointer[uint32](3)},
	})
	require.NoError(t, err)

	type args struct {
		cache   *outputCache
		txHash  string
		outputs map[string]outputMapping
		addrMap map[string]db.Address
	}
	tests := []struct {
		args    args
		wantErr bool
	}{
		{
			args: args{
				cache:   nil,
				txHash:  "",
				outputs: nil,
				addrMap: nil,
			},
			wantErr: true,
		},
		{
			args: args{
				cache:   newOutputCache(),
				txHash:  "",
				outputs: nil,
				addrMap: nil,
			},
			wantErr: true,
		},
		{
			args: args{
				cache:   newOutputCache(),
				txHash:  "asdf",
				outputs: nil,
				addrMap: nil,
			},
			wantErr: false,
		},
		{
			args: args{
				cache:  newOutputCache(),
				txHash: "asdf",
				outputs: map[string]outputMapping{"": {
					hash:    "",
					indexes: []uint32{1, 2, 3},
				}},
				addrMap: map[string]db.Address{},
			},
			wantErr: true,
		},
		{
			args: args{
				cache:  oCache,
				txHash: "asdf",
				outputs: map[string]outputMapping{"": {
					hash:    "",
					indexes: []uint32{1, 2, 3},
				}},
				addrMap: map[string]db.Address{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		err := buildAddresses(new(sync.Mutex), tt.args.cache, tt.args.txHash, tt.args.outputs, tt.args.addrMap)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}

func Test_buildTransactionMapping(t *testing.T) {
	testhelper.SkipIfNotCI(t)

	blockHashes, err := client.Generate(1)
	require.NoError(t, err)
	require.Len(t, blockHashes, 1)

	block, err := client.GetBlock(blockHashes[0])
	require.NoError(t, err)
	require.NotEmpty(t, block.Transactions)

	txHash := block.Transactions[0].TxHash()
	rawTxResult, err := client.GetRawTransactionVerbose(&txHash)
	require.NoError(t, err)
	require.NotNil(t, rawTxResult)

	type args struct {
		rawTransaction  btcjson.TxRawResult
		txHashMap       map[string]btcjson.TxRawResult
		externalOutputs map[string]map[uint32]db.Output
		config          Config
		cache           *outputCache
	}
	tests := []struct {
		args          args
		wantTxDetails db.Transaction
		wantTMap      transactionMapping
		wantErr       bool
	}{
		{
			args: args{
				rawTransaction:  *rawTxResult,
				txHashMap:       map[string]btcjson.TxRawResult{},
				externalOutputs: map[string]map[uint32]db.Output{},
				config:          NewBitcoinConfig(),
				cache:           newOutputCache(),
			},
			wantTxDetails: db.Transaction{},
			wantTMap:      transactionMapping{},
			wantErr:       false,
		},
	}
	for _, tt := range tests {
		_, _, err := buildTransactionMapping(tt.args.rawTransaction, tt.args.txHashMap, tt.args.externalOutputs, tt.args.config, tt.args.cache)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}
