package processor

import (
	"backend/db"
	"backend/db/status"
	"backend/external"
	"backend/jsonrpc"
	"backend/testhelper"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"log"
	"sync"
	"testing"
	"time"
)

var (
	dbHandle          = &testhelper.TestDB{IsDirty: true}
	client            *jsonrpc.BlockchainClient
	generateToAddress string
)

func setupRPCTest(client *jsonrpc.BlockchainClient, numBlocks int) error {
	// wallet might already exist -> ignore error
	_, _ = client.CreateWallet("testwallet")
	// wallet might already be loaded -> ignore error
	_, _ = client.LoadWallet("testwallet")

	var err error
	generateToAddress, err = client.GetNewAddress()
	if err != nil {
		return err
	}

	_, err = client.GenerateToAddress(numBlocks, generateToAddress)
	return err
}

func TestMain(m *testing.M) {
	InitLogger()

	if testhelper.DoDBTests() {
		dbName, ok := testhelper.GetDBName()
		if !ok {
			log.Fatal("environment variable " + testhelper.EnvDBHostname + " is not set")
		}
		// create dgraph client
		graphDB, c, err := external.CreateClient(dbName + ":9080")
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

		dbHandle.DB = graphDB
	}

	if testhelper.DoRPCTests() {
		rpcHostname, ok := testhelper.GetRPCName()
		if !ok {
			log.Panic("environment variable " + testhelper.EnvRPCHostname + " is not set")
			return
		}

		client = jsonrpc.NewBlockchainClient(rpcHostname+":8131", "rpc1user", "1234pass", nil)

		if err := setupRPCTest(client, 5); err != nil {
			log.Panic("Could not setup RPC test", err)
			return
		}
	}

	m.Run()
}

func TestIncrementProcessingState(t *testing.T) {
	const (
		firstHash   = "000000003b36901a4771aebad94ab2707e55b19ba62898bedcea9a69265f8e7"
		secondHash  = "00000000251b4f191d09553f115383e12108fdf98d3d77530a9e96bc9dd6dd6a"
		invalidHash = "."
	)

	var p crawlerState

	err := p.increment(firstHash)
	require.NoError(t, err)
	require.NotEmpty(t, p.String())

	if p.id != 1 || p.hash != firstHash {
		t.Fatal("incrementation not successful")
	}

	err = p.increment(secondHash)
	require.NoError(t, err)
	if p.id != 2 || p.hash != secondHash {
		t.Fatal("incrementation not successful")
	}

	//  no error on invliad values
	err = p.increment(invalidHash)
	require.NoError(t, err)
}

func TestAddOutputToMapping(t *testing.T) {
	outputMappings := make(map[string]outputMapping)
	const (
		firstAddress  = "XtUa1xzS8rr4UMv1bopfTKipspwrUvaBMp"
		secondAddress = "Xhbf3Cj7YVih5Ze8WCgPacRR7oVsaJqRJ3"
	)
	outputMappings = addOutputToMapping(outputMappings, firstAddress, 0)
	if val, ok := outputMappings[firstAddress]; ok {
		switch {
		case len(val.indexes) != 1:
			t.Fatal("wrong length of ids")
		case val.hash != firstAddress:
			t.Fatal("wrong hash")
		case val.indexes[0] != 0:
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
		switch {
		case len(val.indexes) != 1:
			t.Fatal("wrong length of ids")
		case val.hash != secondAddress:
			t.Fatal("wrong hash")
		case val.indexes[0] != 10:
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
		switch {
		case len(val.indexes) != 2:
			t.Fatal("wrong length of ids")
		case val.hash != firstAddress:
			t.Fatal("wrong hash")
		case val.indexes[0] != 0 || val.indexes[1] != 5:
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
	testhelper.SkipIfNoDB(t)
	db.SetupDB(t, dbHandle, testhelper.UseBlockFile)

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
	testhelper.SkipIfNoDB(t)
	testhelper.SkipIfNoRPC(t)
	interrupt := make(chan struct{})
	cfg := NewDashConfig()
	// for a fast test
	cfg.NewBlockIntervalTime = 1

	blkCount, err := client.GetBlockCount()
	require.NoError(t, err)
	// add two blocks, so the first block has a reference to the next block
	hashes, err := client.GenerateToAddress(2, generateToAddress)
	require.NoError(t, err)

	// normal operation
	currentBlock, wasInterrupted, err := waitForNextRPCBlock(client, interrupt, hashes[0], uint64(blkCount), cfg)
	require.NoError(t, err)
	require.False(t, wasInterrupted, "the interrupt flag should have been false")
	require.NotNil(t, currentBlock)

	// missing hash
	currentBlock, wasInterrupted, err = waitForNextRPCBlock(client, interrupt, "", uint64(blkCount), cfg)
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
	testhelper.SkipIfNoDB(t)
	testhelper.SkipIfNoRPC(t)
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

	// no error check for invalid values
	require.NoError(t, state.increment("asdf"))
	require.EqualValues(t, 1, state.id)

	require.NoError(t, state.increment("000007248b1005ffdcf3f41f3a5630b5cb0078ca5733d931223839821f7f5faa"))
	require.EqualValues(t, 2, state.id)
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
	testhelper.SkipIfNoDB(t)
	testhelper.SkipIfNoRPC(t)

	blockHashes, err := client.GenerateToAddress(1, generateToAddress)
	require.NoError(t, err)
	require.Len(t, blockHashes, 1)

	block, err := client.GetBlockVerbose(blockHashes[0])
	require.NoError(t, err)
	require.NotEmpty(t, block.Tx)

	basicBlock, err := client.GetBlockVerbose(blockHashes[0])
	require.NoError(t, err)
	require.NotEmpty(t, basicBlock.Tx)

	rawTxResult, err := client.GetRawTransactionVerbose(basicBlock.Tx[0])
	require.NoError(t, err)
	require.NotNil(t, rawTxResult)

	txHashMap, err := createTransactionHashmap(client, block.Tx)
	require.NoError(t, err)

	txWithoutAddresses := rawTxResult
	for i := range txWithoutAddresses.Vout {
		txWithoutAddresses.Vout[i].ScriptPubKey.Addresses = nil
		txWithoutAddresses.Vout[i].ScriptPubKey.Type = "pubkeyhash"
	}

	type args struct {
		rawTransaction  jsonrpc.TxRawResult
		txHashMap       map[string]jsonrpc.TxRawResult
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
				txHashMap:       txHashMap,
				externalOutputs: map[string]map[uint32]db.Output{},
				config:          NewBitcoinConfig(),
				cache:           newOutputCache(),
			},
			wantTxDetails: db.Transaction{},
			wantTMap:      transactionMapping{},
			wantErr:       false,
		},
		{
			args: args{
				rawTransaction:  *txWithoutAddresses,
				txHashMap:       txHashMap,
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

func Test_filterExternalOutputs(t *testing.T) {
	txMap := map[string]jsonrpc.TxRawResult{"": {
		Vin: []jsonrpc.Vin{
			{Txid: "txhash1", Vout: 0},
			{Txid: "txhash1", Vout: 1},
			{Txid: "txhash1", Vout: 2},
			{Txid: "txhash2", Vout: 3},
			{Txid: "txhash2", Vout: 4},
			{Txid: "txhash2", Vout: 5},
		},
	},
	}

	cache := newOutputCache()
	require.NoError(t, cache.setOutputs("txhash2", []db.Output{
		{OutputIndex: getPointer[uint32](4)},
		{OutputIndex: getPointer[uint32](5)},
	}))

	type args struct {
		txHashMap map[string]jsonrpc.TxRawResult
		cache     *outputCache
	}
	tests := []struct {
		args args
		want map[string][]uint32
	}{
		{
			args: args{
				txHashMap: nil,
				cache:     nil,
			},
			want: map[string][]uint32{},
		},
		{
			args: args{
				txHashMap: txMap,
				cache:     newOutputCache(),
			},
			want: map[string][]uint32{"txhash1": {0, 1, 2}, "txhash2": {3, 4, 5}},
		},
		{
			args: args{
				txHashMap: txMap,
				cache:     cache,
			},
			want: map[string][]uint32{"txhash1": {0, 1, 2}, "txhash2": {3}},
		},
	}
	for _, tt := range tests {
		require.Equal(t, tt.want, filterExternalOutputs(tt.args.txHashMap, tt.args.cache))
	}
}

func Test_processTxVin(t *testing.T) {
	cache := newOutputCache()
	require.NoError(t, cache.setOutputs("txhash1", []db.Output{
		{
			OutputIndex: getPointer[uint32](0),
			Amount:      getPointer[int64](3),
		},
	}))
	type args struct {
		details         *db.Transaction
		externalOutputs map[string]map[uint32]db.Output
		vin             jsonrpc.Vin
		index           uint32
		txHashMap       map[string]jsonrpc.TxRawResult
		cache           *outputCache
	}
	tests := []struct {
		args    args
		wantErr bool
	}{
		{
			args: args{
				details:         &db.Transaction{Inputs: []db.Output{}},
				externalOutputs: nil,
				vin: jsonrpc.Vin{
					Txid: "txhash1",
					ScriptSig: &jsonrpc.ScriptSig{
						Asm: "some_asm",
					},
					Vout: 0,
				},
				index: 0,
				txHashMap: map[string]jsonrpc.TxRawResult{"txhash1": {
					Hash: "txhash1",
					Vout: []jsonrpc.Vout{{Value: 0}},
				}},
				cache: newOutputCache(),
			},
			wantErr: false,
		},
		{
			args: args{
				details:         &db.Transaction{Inputs: []db.Output{}},
				externalOutputs: nil,
				vin: jsonrpc.Vin{
					Txid: "txhash1",
					ScriptSig: &jsonrpc.ScriptSig{
						Asm: "some_asm",
					},
					Vout: 0,
				},
				index:     0,
				txHashMap: map[string]jsonrpc.TxRawResult{},
				cache:     newOutputCache(),
			},
			wantErr: true,
		},
		{
			args: args{
				details:         &db.Transaction{Inputs: []db.Output{}},
				externalOutputs: nil,
				vin: jsonrpc.Vin{
					Txid: "txhash1",
					ScriptSig: &jsonrpc.ScriptSig{
						Asm: "some_asm",
					},
					Vout: 0,
				},
				index:     0,
				txHashMap: map[string]jsonrpc.TxRawResult{},
				cache:     cache,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		err := processTxVin(tt.args.details, tt.args.externalOutputs, tt.args.vin,
			tt.args.index, tt.args.txHashMap, tt.args.cache)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.Len(t, tt.args.details.Inputs, 1)
		}
	}
}

func Test_processBlock(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	db.SetupDB(t, dbHandle, testhelper.UseBlockFile)

	transactions, err := db.GetTransactionByBlock(dbHandle, testhelper.BlockFileFirstBlock)
	require.NoError(t, err)

	type args struct {
		transactions  []db.Transaction
		currentHash   string
		blockID       uint64
		timestamp     string
		prevBlockHash string
	}
	tests := []struct {
		args    args
		wantErr bool
	}{
		{
			args: args{
				transactions:  transactions,
				currentHash:   "some_hash",
				blockID:       5,
				timestamp:     time.Now().Format(time.RFC3339),
				prevBlockHash: "some_other_hash",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		err := processBlock(dbHandle, tt.args.transactions, tt.args.currentHash,
			tt.args.blockID, tt.args.timestamp, tt.args.prevBlockHash)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}

func Test_getStartingID(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	db.SetupDBWithoutData(t, dbHandle)

	require.NoError(t, status.SetCrawling(dbHandle, true))

	gotStartID, err := getStartingID(dbHandle)
	require.NoError(t, err)
	require.EqualValues(t, 1, gotStartID)

	db.SetupDB(t, dbHandle, testhelper.UseBlockFile)

	require.NoError(t, status.SetCrawlerStatus(dbHandle, status.CrawlerStatus{
		IsCrawling: getPointer[bool](true),
		// make blocks not match
		LastBlockID: getPointer[uint64](5),
	}))
	_, err = getStartingID(dbHandle)
	require.Error(t, err)

	require.NoError(t, status.SetCrawlerStatus(dbHandle, status.CrawlerStatus{
		IsCrawling:  getPointer[bool](true),
		LastBlockID: getPointer[uint64](testhelper.BlockFileLastBlock),
	}))
	gotStartID, err = getStartingID(dbHandle)
	require.NoError(t, err)
	require.EqualValues(t, testhelper.BlockFileLastBlock, gotStartID)
}

func Test_processingInterrupted(t *testing.T) {
	require.NotPanics(t, func() {
		processingInterrupted()
	})
}

func Test_getInitialState(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	testhelper.SkipIfNoRPC(t)
	db.SetupDBWithoutData(t, dbHandle)

	_, err := getInitialState(dbHandle, client)
	require.Error(t, err)

	require.NoError(t, status.SetCrawling(dbHandle, true))

	_, err = getInitialState(dbHandle, client)
	require.NoError(t, err)
}

func Test_createTransactionHashmap(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	testhelper.SkipIfNoRPC(t)
	blockHashes, err := client.GenerateToAddress(1, generateToAddress)
	require.NoError(t, err)
	require.NotEmpty(t, blockHashes)

	verboseBlock, err := client.GetBlockVerbose(blockHashes[0])
	require.NoError(t, err)

	hashmap, err := createTransactionHashmap(client, verboseBlock.Tx)
	require.NoError(t, err)
	require.NotEmpty(t, hashmap)
}

func Test_getExternalOutputs(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	db.SetupDB(t, dbHandle, testhelper.UseBlockFile)

	tests := []struct {
		outputs  map[string][]uint32
		wantSize int
		wantErr  bool
	}{
		{
			outputs:  nil,
			wantSize: 0,
			wantErr:  false,
		},
		{
			outputs:  map[string][]uint32{"91609034d29949f9e19dc62637f0665bdc1b161e11b7f360ee692d15b46c8cdb": {0, 1}},
			wantSize: 1,
			wantErr:  false,
		},
		{
			// wrong indexes -> zero size
			outputs:  map[string][]uint32{"91609034d29949f9e19dc62637f0665bdc1b161e11b7f360ee692d15b46c8cdb": {10, 11}},
			wantSize: 0,
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		outputs, err := getExternalOutputs(dbHandle, tt.outputs)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.Len(t, outputs, tt.wantSize)
		}
	}
}

func Test_processRound(t *testing.T) {
	testhelper.SkipIfNoDB(t)
	testhelper.SkipIfNoRPC(t)
	db.SetupDBWithoutData(t, dbHandle)

	blockHashes, err := client.GenerateToAddress(1, generateToAddress)
	require.NoError(t, err)
	require.NotEmpty(t, blockHashes)

	verboseBlock, err := client.GetBlockVerbose(blockHashes[0])
	require.NoError(t, err)

	type args struct {
		state  crawlerState
		block  *jsonrpc.GetBlockVerboseResult
		config Config
		cache  *outputCache
	}
	tests := []struct {
		args    args
		wantErr bool
	}{
		{
			args: args{
				state:  crawlerState{top: uint64(3), id: uint64(1)},
				block:  verboseBlock,
				config: NewBitcoinConfig(),
				cache:  newOutputCache(),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		_, _, err := processRound(dbHandle, client, tt.args.state, tt.args.block, tt.args.config, tt.args.cache)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
		}
	}
}

func Test_getOutputAddress(t *testing.T) {
	tests := []struct {
		pubKey           *jsonrpc.ScriptPubKeyResult
		pubKeyHashAddrID byte
		want             string
		wantErr          bool
	}{
		{
			pubKey:  nil,
			wantErr: true,
		},
		{
			pubKey:  &jsonrpc.ScriptPubKeyResult{},
			want:    "",
			wantErr: false,
		},
		{
			pubKey: &jsonrpc.ScriptPubKeyResult{
				Address: "a",
			},
			want:    "a",
			wantErr: false,
		},
		{
			pubKey: &jsonrpc.ScriptPubKeyResult{
				Addresses: []string{"b"},
				Address:   "a",
			},
			want:    "a",
			wantErr: false,
		},
		{
			pubKey: &jsonrpc.ScriptPubKeyResult{
				Addresses: []string{"b"},
			},
			want:    "b",
			wantErr: false,
		},
		{
			pubKey: &jsonrpc.ScriptPubKeyResult{
				Type: "nulldata",
			},
			want:    "",
			wantErr: false,
		},
		{
			pubKey: &jsonrpc.ScriptPubKeyResult{
				Type: "nonstandard",
			},
			want:    "",
			wantErr: false,
		},
		{
			pubKey: &jsonrpc.ScriptPubKeyResult{
				Hex: "76a914bc89d6071dabc5b4494d303af761a052a5c70d5788ac",
			},
			pubKeyHashAddrID: 0x4c,
			want:             "XssjzLKgsfATYGqTQmiJURQzeKdpL5K1k3",
			wantErr:          false,
		},
		{
			pubKey: &jsonrpc.ScriptPubKeyResult{
				Hex: "76a914bc89d6071dabc5b4494d303af761a052a5c70d5788ac",
			},
			pubKeyHashAddrID: 0x00,
			want:             "1JBuA5fnuwwsPLEsYtQ5ctjCoz48JpqMmB",
			wantErr:          false,
		},
		{
			pubKey: &jsonrpc.ScriptPubKeyResult{
				// invalid hex
				Hex: "76a914bc89d6071dabc5b4494d303af761a052a5c70d5788a",
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		address, err := getOutputAddress(tt.pubKey, tt.pubKeyHashAddrID)
		if tt.wantErr {
			require.Error(t, err)
		} else {
			require.NoError(t, err)
			require.Equal(t, tt.want, address)
		}
	}
}
