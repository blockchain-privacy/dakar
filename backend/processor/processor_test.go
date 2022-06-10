package processor

import (
	dbaddr "backend/db/address"
	dbop "backend/db/output"
	"backend/mocks"
	"errors"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

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
	addresses := make(map[string]dbaddr.Address)
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

func TestDecodeAddress(t *testing.T) {
	cfgDash := NewDashConfig()
	cases := []struct {
		works bool
		asm   string
	}{
		{true, "02e5a489b4fb934af831a9553c6521b3bf1f155bfa5a6c72d13461039ebc594cff"},
		{true, "036fc4628db906187d44ab325f8f8977c3686bc29ca76e1905b3afe82119bc5a7f"},
		{true, "0204d9dc3bd0f5e83901fcec9410c4a28907405b783200d47bbc3e5c7e9414acca"},
		{true, "029e2c658a15d18f7d4f44bbcdd80a18069b49ee2636b5937e28a2490536612b84"},
		{true, "0226baf8c8c3707d6062e754e10ee46fa13dcd71aa64cc1a1c6d433fd53de45c07"},
		{true, "02636234f12e46575b95e0c5a1251a211368dc0ec28abdf02986dc249be28f4eda"},
		{false, "02636234f12e46575b95e0c5a1251a211368dc0ec28abdf02986dc249be28f4eda1"},
		{false, ""},
		{false, "asdfafdssafd.123,"},
		{false, "  "},
	}

	for _, c := range cases {
		address, err := decodeAddress(c.asm, cfgDash.PubKeyHashAddrID)
		if c.works {
			require.Nil(t, err)
			require.NotEmpty(t, address)
		} else {
			require.NotNil(t, err)
		}
	}
}

func TestProcessAddresses(t *testing.T) {
	const (
		fistAddress   = "XsAptUZUmtL8onHcuJSvGM8MyvR7QCpw9u"
		secondAddress = "Xoi9jutn8qbtxvd2V3xqqSQnpdqPrCzP1K"
		txHash        = "123456"
	)

	oMap := make(map[string]outputMapping)
	oMap[fistAddress] = outputMapping{
		hash:    fistAddress,
		indexes: []uint32{0},
	}
	oMap[secondAddress] = outputMapping{
		hash:    secondAddress,
		indexes: []uint32{1},
	}

	txMap := transactionMapping{
		hash:    txHash,
		outputs: oMap,
	}

	zero := uint32(0)
	one := uint32(1)
	one64 := int64(1)
	fourNines := int64(9999)
	wrong := false
	output1 := dbop.Output{
		UID:         "0x59b84",
		OutputIndex: &zero,
		InputIndex:  nil,
		TxType:      "pubkeyhash",
		Amount:      &one64,
		IsCoinbase:  &wrong,
		DType:       nil,
	}

	output2 := dbop.Output{
		UID:         "0x59b85",
		OutputIndex: &one,
		InputIndex:  nil,
		TxType:      "pubkeyhash",
		Amount:      &fourNines,
		IsCoinbase:  &wrong,
		DType:       nil,
	}

	outputArr := []dbop.Output{
		output1, output2,
	}

	cache := newOutputCache()
	require.Nil(t, cache.setOutputs(txHash, outputArr))

	db := new(mocks.Database)
	mocks.MapUpsertAddresses(db)
	require.Nil(t, processAddresses(db, cache, []transactionMapping{txMap}))
}

func TestWaitForNextRPCBlock(t *testing.T) {
	var rpcClient mocks.RPCClient
	hash := mocks.RPCVal.HeightMap[1423340]
	var nilHash *chainhash.Hash
	expectedBlock := mocks.RPCVal.BlockStore[hash]
	interrupt := make(chan struct{})
	blkInfo := mocks.RPCVal.BlockchainInfo
	cfg := NewDashConfig()
	// for a quick test
	cfg.NewBlockIntervalTime = 1

	rpcClient.On("GetBlockVerbose", &hash).Return(&expectedBlock, nil)
	rpcClient.On("GetBlockVerbose", nilHash).Return(nil, errors.New("invalid argument"))
	rpcClient.On("GetBlockChainInfo").Return(&blkInfo, nil)

	// normal operation
	currentBlock, wasInterrupted, err := waitForNextRPCBlock(&rpcClient, interrupt, &hash, uint64(blkInfo.Blocks-1), cfg)
	require.Nil(t, err)
	require.False(t, wasInterrupted, "the interrupt flag should have been false")
	require.NotNil(t, currentBlock)

	// missing hash
	currentBlock, wasInterrupted, err = waitForNextRPCBlock(&rpcClient, interrupt, nil, uint64(blkInfo.Blocks-1), cfg)
	require.NotNil(t, err)
	require.False(t, wasInterrupted, "the interrupt flag should have been false")
	require.Nil(t, currentBlock)
	var test struct{}

	go func() {
		interrupt <- test
	}()

	// normal operation but interrupted and higher block
	// count as available, so it must wait or in this case get interrupted
	cfg.NewBlockIntervalTime = time.Second
	currentBlock, wasInterrupted, err = waitForNextRPCBlock(&rpcClient, interrupt, &hash, uint64(blkInfo.Blocks+1), cfg)
	require.Nil(t, err)
	require.True(t, wasInterrupted, "the interrupt flag should have been true")
	require.Nil(t, currentBlock)

	rpcClient.AssertExpectations(t)
}

func TestGetRPCNumberOfBlocks(t *testing.T) {
	var rpcClient mocks.RPCClient
	rpcClient.On("GetBlockChainInfo").Return(&mocks.RPCVal.BlockchainInfo, nil)

	numBlocks, err := getRPCNumberOfBlocks(&rpcClient)
	require.Nil(t, err)
	require.NotZerof(t, numBlocks, "number of blocks should not be zero")
}
