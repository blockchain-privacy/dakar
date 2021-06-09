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

	var p crawlerProcessingState

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
		newAddressMap := addOutputsToAddresses(addresses, c.address, c.uids)
		require.Len(t, newAddressMap, c.requiredLength)
		addresses = newAddressMap
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

func TestBuildAddressMapping(t *testing.T) {
	const (
		fistAddress   = "XsAptUZUmtL8onHcuJSvGM8MyvR7QCpw9u"
		secondAddress = "Xoi9jutn8qbtxvd2V3xqqSQnpdqPrCzP1K"
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

	zero := uint32(0)
	one := uint32(1)
	one64 := int64(1)
	zero64 := int64(0)
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
	addresses := make(map[string]dbaddr.Address)
	addresses[fistAddress] = dbaddr.Address{
		UID:  "",
		Hash: fistAddress,
		Outputs: []dbop.Output{{
			UID:         "0x59b81",
			OutputIndex: nil,
			InputIndex:  nil,
			TxType:      "",
			Amount:      &zero64,
			IsCoinbase:  nil,
			DType:       nil,
		}},
		DType: nil,
	}

	buildAddressMapping(oMap, outputArr, &addresses)

	if val, ok := addresses[fistAddress]; ok {
		if len(val.Outputs) != 2 {
			t.Error("Wrong number of outputs")
		} else if val.Outputs[0].UID != "0x59b81" ||
			val.Outputs[1].UID != "0x59b84" {
			t.Error("Uids not set")
		}
	}
}

func TestWaitForNextRPCBlock(t *testing.T) {
	var rpcClient mocks.RPCClient
	hash := mocks.RPCVal.HeightMap[1423340]
	nilHash := &chainhash.Hash{}
	nilHash = nil
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
	// count as available so it must wait or in this case get interrupted
	cfg.NewBlockIntervalTime = time.Second
	currentBlock, wasInterrupted, err = waitForNextRPCBlock(&rpcClient, interrupt, &hash, uint64(blkInfo.Blocks+1), cfg)
	require.Nil(t, err)
	require.True(t, wasInterrupted, "the interrupt flag should have been true")
	require.Nil(t, currentBlock)

	rpcClient.AssertExpectations(t)
}

//const block49998 = "000000000018692f3cd1e6255d9aa3edc427101e02da940f6e6673823118f016"
//const block49999 = "000000000014f796bbd2312686a63cbe17401a1026ab2a8149b74553e8dcb96d"
//const block50000 = "00000000000fa6230896498b3cc6f1015456b4512452ead9979f6b43ca0a74dc"
//
//func TestProcessBlock50000(t *testing.T) {
//	db := backend.setupDB(t)
//	defer backend.tearDownDB(t, db)
//	client := backend.setupRpcClient(t)
//
//	block := backend.Block{}
//
//	startHash, err := chainhash.NewHashFromStr(block50000)
//	if err != nil {
//		fmt.Printf("we have problem with HashFromStr() %s\n", err.Error())
//	}
//	startBlock, err := client.GetBlock(startHash)
//	if err != nil {
//		fmt.Printf("we have problem with getBlock() %s\n", err.Error())
//		t.Error(t)
//		return
//	}
//	err = ProcessBlock(db, startBlock, *startHash, chainhash.Hash{}, 50000, &block)
//	if err != nil {
//		t.Error(err)
//	}
//
//	block2 := backend.Block{}
//	err = backend.DbGetBlock(db, block50000, &block2)
//	if err != nil {
//		t.Error(err)
//	}
//
//	if block2.Hash != block.Hash ||
//		block2.NextBlockHash != block.NextBlockHash ||
//		block2.PrevBlockHash != block.PrevBlockHash ||
//		len(block2.TxHashes) != len(block.TxHashes) ||
//		block2.TxHashes[0] != block.TxHashes[0] {
//		t.Error("Blocks do not match")
//	}
//
//	if block2.PrevBlockHash.String() != block49999 ||
//		block.PrevBlockHash.String() != block49999 {
//		msg := fmt.Sprintf("PrevBlockHash does not match!\n%s\n%s\n%s\n",
//			block49999, block.PrevBlockHash, block2.PrevBlockHash)
//		t.Error(msg)
//	}
//
//	if block2.TxHashes[0] != "c13fc482603f574b7322da10398c20d64a431e14f8e886b054128591abaa66a4" {
//		t.Error("Output Transaction hash is WRONG")
//	}
//
//}
//
//func TestProcessBlock49999(t *testing.T) {
//	db := backend.setupDB(t)
//	defer backend.tearDownDB(t, db)
//
//	client := backend.setupRpcClient(t)
//
//	block := backend.Block{}
//
//	startHash, err := chainhash.NewHashFromStr(block49999)
//	if err != nil {
//		fmt.Printf("we have problem with HashFromStr() %s\n", err.Error())
//	}
//	block50000hash, err := chainhash.NewHashFromStr(block50000)
//	if err != nil {
//		fmt.Printf("we have problem with HashFromStr() %s\n", err.Error())
//	}
//
//	startBlock, err := client.GetBlock(startHash)
//	if err != nil {
//		fmt.Printf("we have problem with getBlock() %s\n", err.Error())
//		t.Error(t)
//		return
//	}
//	err = ProcessBlock(db, startBlock, *startHash, *block50000hash, 50000, &block)
//	if err != nil {
//		t.Error(err)
//	}
//
//	block2 := backend.Block{}
//	err = backend.DbGetBlock(db, block49999, &block2)
//	if err != nil {
//		t.Error(err)
//	}
//
//	if block2.Hash != block.Hash ||
//		block2.NextBlockHash != block.NextBlockHash ||
//		block2.PrevBlockHash != block.PrevBlockHash ||
//		len(block2.TxHashes) != len(block.TxHashes) ||
//		len(block2.TxHashes) != 9 ||
//		block2.TxHashes[0] != block.TxHashes[0] {
//		t.Error("Blocks do not match")
//	}
//
//	if block2.PrevBlockHash.String() != block49998 ||
//		block.NextBlockHash.String() != block50000 {
//		msg := fmt.Sprintf("PrevBlockHash does not match!\n%s\n%s\n%s\n",
//			block49999, block.PrevBlockHash, block2.PrevBlockHash)
//		t.Error(msg)
//	}
//
//	if block2.TxHashes[0] != "106f0dea7bdff518a5db6551dd43210d6639fffad84e56083e73231921c779f9" {
//		fmt.Printf("TxHashes %v\n", block2.TxHashes)
//		t.Error("Output Transaction hash is WRONG")
//	}
//
//}
//
//func TestProcessTxFromBlock50000(t *testing.T) {
//	db := backend.setupDB(t)
//	defer backend.tearDownDB(t, db)
//
//	client := backend.setupRpcClient(t)
//
//	block := backend.Block{}
//	txHash := "c13fc482603f574b7322da10398c20d64a431e14f8e886b054128591abaa66a4"
//
//	startHash, err := chainhash.NewHashFromStr(block50000)
//	if err != nil {
//		fmt.Printf("we have problem with HashFromStr() %s\n", err.Error())
//	}
//	startBlock, err := client.GetBlock(startHash)
//	if err != nil {
//		fmt.Printf("we have problem with getBlock() %s\n", err.Error())
//		t.Error(t)
//		return
//	}
//	err = ProcessBlock(db, startBlock, *startHash, chainhash.Hash{}, 50000, &block)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	// test without addresses
//	err = BuildTransactionMapping(db, client, false, txHash)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	txDetails := backend.TxDetails{}
//	err = backend.DbGetTxDetails(db, block.TxHashes[0], &txDetails)
//	if err != nil {
//		t.Error(err)
//	}
//	if len(txDetails.Outputs) != 1 ||
//		txDetails.Outputs[0].Amount != 16.00 ||
//		txDetails.Outputs[0].IsCoinbase != false ||
//		len(txDetails.Inputs) != 1 ||
//		txDetails.Inputs[0].IsCoinbase != true {
//		msg := fmt.Sprintf("Error: TX data does not match for TX c13fc48....\nData:\n%v\n", txDetails)
//		t.Error(msg)
//	}
//
//	if err != nil {
//		t.Error(err)
//	}
//
//}
//
//func TestProcessTxFromBlock49999WithoutAddresses(t *testing.T) {
//	db := backend.setupDB(t)
//	defer backend.tearDownDB(t, db)
//
//	client := backend.setupRpcClient(t)
//
//	txHash := "af530c23992d7439107b31d8840facb60d0606d370c9cdd35195eea87113ff1e"
//
//	// test with addresses
//	err := BuildTransactionMapping(db, client, false, txHash)
//	if err != nil {
//		t.Fatal(err)
//		return
//	}
//
//	txDetails := backend.TxDetails{}
//	err = backend.DbGetTxDetails(db, txHash, &txDetails)
//	if err != nil {
//		t.Error(err)
//	}
//	if len(txDetails.Outputs) != 2 ||
//		len(txDetails.Inputs) != 2 ||
//		txDetails.Outputs[0].Amount != 9.91547479 ||
//		txDetails.Outputs[1].Amount != 6.02335110 ||
//		len(txDetails.Outputs[0].Addresses) != 1 ||
//		len(txDetails.Outputs[1].Addresses) != 1 ||
//		txDetails.Outputs[0].Addresses[0] != "XrHBvi9hxQcrUfXsB9hK6V7hb2625s2kAV" ||
//		txDetails.Outputs[1].Addresses[0] != "Xstz9D2DNrrCWhAnsmiu1R144DesKNw22t" ||
//		txDetails.Outputs[0].IsCoinbase != false ||
//		txDetails.Outputs[1].IsCoinbase != false ||
//		txDetails.Outputs[0].Index != 0 ||
//		txDetails.Outputs[1].Index != 1 ||
//		len(txDetails.Inputs[0].Addresses) != 1 ||
//		len(txDetails.Inputs[1].Addresses) != 1 ||
//		txDetails.Inputs[0].Amount != 7.73616759 ||
//		txDetails.Inputs[1].Amount != 8.20365830 ||
//		txDetails.Inputs[0].Addresses[0] != "XooKzX2FFWZekaVg7X8T67oLWE2v1tpX5z" ||
//		txDetails.Inputs[1].Addresses[0] != "XnLNnQVYQ9P2zc6uQrY5vypLmXoTiqxrw7" ||
//		txDetails.Inputs[0].IsCoinbase != false ||
//		txDetails.Inputs[1].IsCoinbase != false {
//		msg := fmt.Sprintf("Error: TX data does not match for TX af530c23992d74....\nData:\n%v\n", txDetails)
//		t.Error(msg)
//	}
//
//	if err != nil {
//		t.Error(err)
//	}
//
//	// check if address is inserted
//	addressHash1 := "XooKzX2FFWZekaVg7X8T67oLWE2v1tpX5z"
//	addressHash2 := "XnLNnQVYQ9P2zc6uQrY5vypLmXoTiqxrw7"
//
//	addressData1 := backend.AddressData{}
//	addressData2 := backend.AddressData{}
//
//	err = backend.DbGetDataForAddress(db, addressHash1, &addressData1)
//	if err == nil {
//		msg := fmt.Sprintf("Error: address data should not be available, but is included in the database\nData:\n%v\n", addressData1)
//		t.Error(msg)
//	}
//
//	err = backend.DbGetDataForAddress(db, addressHash2, &addressData2)
//	if err == nil {
//		msg := fmt.Sprintf("Error: address data should not be available, but is included in the database\nData:\n%v\n", addressData2)
//		t.Error(msg)
//	}
//}
//
//func TestProcessTxFromBlock49999WithAddresses(t *testing.T) {
//	db := backend.setupDB(t)
//	defer backend.tearDownDB(t, db)
//
//	client := backend.setupRpcClient(t)
//
//	txHash := "af530c23992d7439107b31d8840facb60d0606d370c9cdd35195eea87113ff1e"
//
//	// test with addresses
//	err := BuildTransactionMapping(db, client, true, txHash)
//	if err != nil {
//		t.Fatal(err)
//		return
//	}
//
//	// check if TX details are okay
//	txDetails := backend.TxDetails{}
//	err = backend.DbGetTxDetails(db, txHash, &txDetails)
//	if err != nil {
//		t.Error(err)
//	}
//	if len(txDetails.Outputs) != 2 ||
//		len(txDetails.Inputs) != 2 ||
//		txDetails.Outputs[0].Amount != 9.91547479 ||
//		txDetails.Outputs[1].Amount != 6.02335110 ||
//		len(txDetails.Outputs[0].Addresses) != 1 ||
//		len(txDetails.Outputs[1].Addresses) != 1 ||
//		txDetails.Outputs[0].Addresses[0] != "XrHBvi9hxQcrUfXsB9hK6V7hb2625s2kAV" ||
//		txDetails.Outputs[1].Addresses[0] != "Xstz9D2DNrrCWhAnsmiu1R144DesKNw22t" ||
//		txDetails.Outputs[0].IsCoinbase != false ||
//		txDetails.Outputs[1].IsCoinbase != false ||
//		txDetails.Outputs[0].Index != 0 ||
//		txDetails.Outputs[1].Index != 1 ||
//		len(txDetails.Inputs[0].Addresses) != 1 ||
//		len(txDetails.Inputs[1].Addresses) != 1 ||
//		txDetails.Inputs[0].Amount != 7.73616759 ||
//		txDetails.Inputs[1].Amount != 8.20365830 ||
//		txDetails.Inputs[0].Addresses[0] != "XooKzX2FFWZekaVg7X8T67oLWE2v1tpX5z" ||
//		txDetails.Inputs[1].Addresses[0] != "XnLNnQVYQ9P2zc6uQrY5vypLmXoTiqxrw7" ||
//		txDetails.Inputs[0].IsCoinbase != false ||
//		txDetails.Inputs[1].IsCoinbase != false {
//		msg := fmt.Sprintf("Error: TX data does not match for TX af530c23992d74....\nData:\n%v\n", txDetails)
//		t.Error(msg)
//	}
//
//	if err != nil {
//		t.Error(err)
//	}
//
//	// check if address is inserted
//	addressHash1 := "XooKzX2FFWZekaVg7X8T67oLWE2v1tpX5z"
//	addressHash2 := "XnLNnQVYQ9P2zc6uQrY5vypLmXoTiqxrw7"
//
//	addressData1 := backend.AddressData{}
//	addressData2 := backend.AddressData{}
//
//	err = backend.DbGetDataForAddress(db, addressHash1, &addressData1)
//	if err != nil {
//		t.Error(err)
//	}
//
//	err = backend.DbGetDataForAddress(db, addressHash2, &addressData2)
//	if err != nil {
//		t.Error(err)
//	}
//
//	if addressData1.Address != addressHash1 ||
//		len(addressData1.Txs) != 1 ||
//		addressData1.Txs[0].TxHash != "6cf491409356e4ef6fbe758fac68a69370d3ddfcdc442051f9334120692085cc" ||
//		addressData1.Txs[0].Amount != 7.73616759 ||
//		addressData1.Txs[0].Index != 0 ||
//		addressData1.Txs[0].IsCoinbase != false {
//		msg := fmt.Sprintf("Error: address data does not match  XooKzX2FFWZeka....\nData:\n%v\n", addressData1)
//		t.Error(msg)
//	}
//
//	if addressData2.Address != addressHash2 ||
//		len(addressData1.Txs) != 1 ||
//		addressData2.Txs[0].TxHash != "f45e46db10ab19a662ba98fe864ffa5149ede795e6bb5802ec92f3518bd5b833" ||
//		addressData2.Txs[0].Amount != 8.2036583 ||
//		addressData2.Txs[0].Index != 0 ||
//		addressData2.Txs[0].IsCoinbase != false {
//		msg := fmt.Sprintf("Error: address data does not match  XnLNnQVYQ9P2zc....\nData:\n%v\n", addressData2)
//		t.Error(msg)
//	}
//}
