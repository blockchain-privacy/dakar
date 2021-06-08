package processor

import (
	dbaddr "backend/db/address"
	dbop "backend/db/output"
	"errors"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"testing"
)

type mockRPC struct {
	mock.Mock
	blockHeight          int64
	blockStore           map[chainhash.Hash]btcjson.GetBlockVerboseResult
	txStore              map[chainhash.Hash]btcjson.TxRawResult
	heightMap            map[int64]chainhash.Hash
	blockchainInfo       btcjson.GetBlockChainInfoResult
	futureBlockchainInfo rpcclient.FutureGetBlockChainInfoResult
}

func newMockRPC() mockRPC {
	blkInfo := btcjson.GetBlockChainInfoResult{
		Chain:                "main",
		Blocks:               1483900,
		Headers:              1483900,
		BestBlockHash:        "000000000000000e92425469f883c8b61fa1008c32bb77d18abedaf92032a991",
		Difficulty:           144233434.3550941,
		MedianTime:           1623066730,
		VerificationProgress: 0.9999998218452665,
		Pruned:               false,
		PruneHeight:          0,
		ChainWork:            "000000000000000000000000000000000000000000005a7453e98c68a20099fe",
		SoftForks:            nil,
		UnifiedSoftForks:     nil,
	}

	store := make(map[chainhash.Hash]btcjson.GetBlockVerboseResult)

	blockHash1, err := chainhash.NewHashFromStr("0000000000000005fc2136de2af0c2de7d905150ba609b0a2250a896741000de")
	if err != nil {
		return mockRPC{}
	}

	blockHash2, err := chainhash.NewHashFromStr("00000000000000085f0674b61c363b5600efc7e640e41f964ca0917c111ed320")
	if err != nil {
		return mockRPC{}
	}

	blockHash3, err := chainhash.NewHashFromStr("000000000000000f61e863e914dde7a2146c555ac1608e8f1db940033319ca11")
	if err != nil {
		return mockRPC{}
	}

	txHash1, err := chainhash.NewHashFromStr("9b75129c5d8538287e3c34035da0420905b6ff1651f06a6f292392bc825a4af5")
	if err != nil {
		return mockRPC{}
	}

	txHash2, err := chainhash.NewHashFromStr("8c5aa04bf4ac5047b85e8252b3adf92602a14ed1cf21b798b789233047f826d0")
	if err != nil {
		return mockRPC{}
	}

	txHash3, err := chainhash.NewHashFromStr("83958bf3982d4ad722a1e39c06d6049a0e4e48d2146ac2791dfbc193302b20cf")
	if err != nil {
		return mockRPC{}
	}

	block1 := btcjson.GetBlockVerboseResult{
		Hash:          "0000000000000005fc2136de2af0c2de7d905150ba609b0a2250a896741000de",
		Confirmations: 60558,
		StrippedSize:  0,
		Size:          38515,
		Weight:        0,
		Height:        1423340,
		Version:       536870912,
		VersionHex:    "20000000",
		MerkleRoot:    "a56160ccabcd1ad5e3e79577a9bb384a9145affc87d15dca0add23de7daedbd9",
		Tx: []string{
			"9b75129c5d8538287e3c34035da0420905b6ff1651f06a6f292392bc825a4af5",
			"8c5aa04bf4ac5047b85e8252b3adf92602a14ed1cf21b798b789233047f826d0",
			"83958bf3982d4ad722a1e39c06d6049a0e4e48d2146ac2791dfbc193302b20cf",
			"757bf538649bca154a9471f9e86267f893df7dbf4b7b6ddb917fdc9db2832a9c",
			"a6411195c5cd52efcaa0892b6798c5ab83269257711f1239af0bce006257d7f7",
			"491c34c1373edd74364e43759744623577fd8237a73f2ee63219260ff7e03e05",
			"a7a9496878ea0f3e620424b0a0090073d058b04aa86b1fa36dcfa61e112822b2",
			"30ee604c4a51290877d02f4739d29cf925d26188612141aab59706c2c5a755c0",
			"2a2a41145b9e8d73ab369e0867e057a96d3bd2d345ff6af4437a52b18301120f",
			"be03b98e2bd2d549ab31565324b919602f2b43cb579c5aea4f41370173c7f9dd",
			"0768b069bdea56d17918df47688f7a2ffabaf6391b83b45d11d24994ad0606c8",
			"94d9c1337fb984d61c1168c9b56a3496cdcfaea6a1489a02863227537712f7b4",
		},
		RawTx:        nil,
		Time:         1613520445,
		Nonce:        1524677851,
		Bits:         "19176191",
		Difficulty:   183691028.7073135,
		PreviousHash: "000000000000000af1342de65ddb266425e0a3df020a46dc94c8f2289bf92aba",
		NextHash:     "00000000000000085f0674b61c363b5600efc7e640e41f964ca0917c111ed320",
	}

	block2 := btcjson.GetBlockVerboseResult{
		Hash:          "00000000000000085f0674b61c363b5600efc7e640e41f964ca0917c111ed320",
		Confirmations: 60561,
		StrippedSize:  0,
		Size:          38918,
		Weight:        0,
		Height:        1423341,
		Version:       536870912,
		VersionHex:    "20000000",
		MerkleRoot:    "a3dd5f571ff68eef584db8539acc0a0daae114c1490b7e5eade55d3a584d4cba",
		Tx: []string{
			"6dcdbf5119665dcda1a15828fc9f2ca37629e5c4770a08a074663cb4e6481066",
			"7cdafb2500c7fc29a5680de1cb15b8c9fa410b65e54337458858b6bd7f2fb708",
			"b8820802c68667bbbd7e01737d37450fb1dd3a42bb39b4c59df1deea1f6a93dc",
			"e8756c3387b5eabe3ded2d98674204c002c6b0dd46b17e7f64c0d522bd61eff6",
			"2129d248001f490883a3873c14d1100e14613f279cd2d8ef4b16dd5d2060ff2f",
			"d4f8082dfb73a3534d5cb276618fd36a328afeeb1e27ddb38ac387638e30c874",
		},
		RawTx:        nil,
		Time:         1613520531,
		Nonce:        2919555214,
		Bits:         "19173043",
		Difficulty:   185216707.5260828,
		PreviousHash: "0000000000000005fc2136de2af0c2de7d905150ba609b0a2250a896741000de",
		NextHash:     "000000000000000f61e863e914dde7a2146c555ac1608e8f1db940033319ca11",
	}

	block3 := btcjson.GetBlockVerboseResult{
		Hash:          "000000000000000f61e863e914dde7a2146c555ac1608e8f1db940033319ca11",
		Confirmations: 60560,
		StrippedSize:  0,
		Size:          16403,
		Weight:        0,
		Height:        1423342,
		Version:       536870912,
		VersionHex:    "20000000",
		MerkleRoot:    "30a578142522c3c4b44fbc00751af71056d6207ace6f821df9265855f7288fd4",
		Tx: []string{
			"05ac43990effe11c9b46e3a5788e33e55c8d747b685fdd494883a446ef063db8",
			"5e4504a870095e08fcaed04fbc75be6ef3cd60d117d096e0c83ca13e16d08ba9",
			"8cf1eb5aa14855d58bb98fa4023a816733af76df0faf0824eaf29423320ad531",
			"6187a7d1c4c8520519076f3f5e52abf71af4f21d3672270c1f426005664a7171",
		},
		RawTx:        nil,
		Time:         1613520586,
		Nonce:        2083211797,
		Bits:         "19172440",
		Difficulty:   185592243.8384606,
		PreviousHash: "00000000000000085f0674b61c363b5600efc7e640e41f964ca0917c111ed320",
		NextHash:     "00000000000000015a377874b283924e5745fd11decab0c0be8704874a5aa948",
	}
	store[*blockHash1] = block1
	store[*blockHash2] = block2
	store[*blockHash3] = block3

	hMap := make(map[int64]chainhash.Hash)
	hMap[1423340] = *blockHash1
	hMap[1423341] = *blockHash2
	hMap[1423342] = *blockHash3

	txMap := make(map[chainhash.Hash]btcjson.TxRawResult)

	txMap[*txHash1] = btcjson.TxRawResult{
		Hex:      "03000500010000000000000000000000000000000000000000000000000000000000000000ffffffff2003ecb7151b5c4c55584f525c000000002988056645c219dbf6a043bb9e008b0500000000021a586008000000001976a9148bf9a55d3864e49d0bb9f8cc75b6ab3949526ab888ac9ad3d208000000001976a91458e09f6b914214ef3bf2907dabd496822616654388ac00000000460200ecb715009bd04fb02da560fdd1d3640c6db267ec8ab05416cf87befa7dfb5c24b6b4b59b19f8270cdb792df824dfa25504fdff26b879c793fc800ae21a8ce061a8f9ebdc",
		Txid:     "9b75129c5d8538287e3c34035da0420905b6ff1651f06a6f292392bc825a4af5",
		Hash:     "",
		Size:     222,
		Vsize:    0,
		Weight:   0,
		Version:  3,
		LockTime: 0,
		Vin: []btcjson.Vin{
			{
				Coinbase:  "03ecb7151b5c4c55584f525c000000002988056645c219dbf6a043bb9e008b05",
				Txid:      "",
				Vout:      0,
				ScriptSig: nil,
				Sequence:  0,
				Witness:   nil,
			},
		},
		Vout: []btcjson.Vout{
			{
				Value: 1.40531738,
				N:     0,
				ScriptPubKey: btcjson.ScriptPubKeyResult{
					Asm:       "OP_DUP OP_HASH160 8bf9a55d3864e49d0bb9f8cc75b6ab3949526ab8 OP_EQUALVERIFY OP_CHECKSIG",
					Hex:       "76a9148bf9a55d3864e49d0bb9f8cc75b6ab3949526ab888ac",
					ReqSigs:   1,
					Type:      "pubkeyhash",
					Addresses: []string{"XoSxpd5VYNQvKbXbEaDKt6P1aZANzAkXrJ"},
				},
			},
			{
				Value: 1.48034458,
				N:     1,
				ScriptPubKey: btcjson.ScriptPubKeyResult{
					Asm:       "OP_DUP OP_HASH160 58e09f6b914214ef3bf2907dabd4968226166543 OP_EQUALVERIFY OP_CHECKSIG",
					Hex:       "76a91458e09f6b914214ef3bf2907dabd496822616654388ac",
					ReqSigs:   1,
					Type:      "pubkeyhash",
					Addresses: []string{"XinnPCJsEqyvHiusgikA1bs7aVDBq42f9K"},
				},
			},
		},
		BlockHash:     "0000000000000005fc2136de2af0c2de7d905150ba609b0a2250a896741000de",
		Confirmations: 60571,
		Time:          1613520445,
		Blocktime:     1613520445,
	}

	txMap[*txHash2] = btcjson.TxRawResult{
		Hex:      "0200000006a4e1c37d0dc4e9acfcb5c5c31a678ed90769fbc6e9a382545c9893de472a8614040000006b483045022100a6dedfc533202a186902ba80354aed94448f2c348ffb219e75ffef756dd93cb602205c207114a7e9433c9f7973e2cb150d45daf02ae062141a0627e385670d6c9c23812102146d0fd0e96964116bb3ef3459cfe20ad4e8b671d1b244d9d875f20a62429a20ffffffff83c5e495506cf6450d74a17de4964dfa0879cb7076ce670cfa24882fac1d8a240a0000006a473044022053b2cba6a230d5dba17592eb2a04193cd4061b1d651e18622ca5454c80f2617d02205e76654b63ef5134118e7a15541d129b2b292b53cb6bd80a446a21cf4612a816812103ce29df4edfc3f2db18dc8691d14e9857af231db2582e46d5d27e7cbfaf9033e1ffffffff43886059990c949eaba1fa08d195e748b63c45e546141a19b8d3ca70d5e4974d020000006b48304502210084d31643d8ce802cb53f013c78f79806dc74488a60a2cddba5dcde9e0578c56502201fce4445b867802b67a8fb5ff90ce9598f226ec05419005396662a012e7c3a56812103309dea62304772c548acb2846962faf5c1c4f0ded1b9b1b86b7defd7e085a943ffffffff033f4a626eb58e78e3e9863a1253bd7a94174da84a472d7e836eb1d2d19f954e030000006a47304402203d220c5da637ce9f2491aac4faec5fb1a31062f008cd5fa7e344063f7dbdf59602207391173aa548bf351da9984ce556aa9f292c304d14116ab99a4b76ec7ec156868121036b16106b827ea3c55d2b4e3b60cbceb161e3dab7ce9f9877c999e40e90642202ffffffff0e91c6fb0be2c7a778990286b826e0c0cad0022cca7d869b6b1956dd02c86876000000006b483045022100c227a45ce2a70f13e3996e20001061bc5ab84e7fde637abb80b3f369caa6020102200ab5360ebec5f042e58c92d1bb9db67b70a3ba52575ea671b566d4b83d7dc460812102af83cfa5120ebd72ab19219b59ee2a57d6e1fdeb64d47b09ceeb5f28c1d48859ffffffff98af88a645380f4cdfa2edd210137c8b2c3070298d6888f901057750bc7c09cb0a0000006b483045022100e12efb4a9c7a8a8962fabb380944e8e087bd7c466705f34473489a2d4042560a02204cc5d63bdbf6e5cc89e89bf2bf4a51caa89cfe3e1e19cabd3faa0f1619e099e88121026ad2558403314dcc28e94f01213ed7caa8c3fe8ebafb797cfbc0cc385f63c908ffffffff064a420f00000000001976a9140a00e72b922416c12779191500b100bf959f531988ac4a420f00000000001976a9148361fc4a3b474263ae2b7a3fb229000162c5709688ac4a420f00000000001976a914b86e3ebfdc69689971a441a58d867199410529b988ac4a420f00000000001976a914c6ef7ef72e3975177c4544fa941c2b198c030ea088ac4a420f00000000001976a914df6e49e6fa7a5a7b6e304f14e5ad8dc9b9192cec88ac4a420f00000000001976a914f7ae957e3152f3b9d23de5e551de6973c839d7a988ac00000000",
		Txid:     "8c5aa04bf4ac5047b85e8252b3adf92602a14ed1cf21b798b789233047f826d0",
		Hash:     "",
		Size:     1100,
		Vsize:    0,
		Weight:   0,
		Version:  2,
		LockTime: 0,
		Vin: []btcjson.Vin{
			{
				Coinbase: "",
				Txid:     "14862a47de93985c5482a3e9c6fb6907d98e671ac3c5b5fcace9c40d7dc3e1a4",
				Vout:     4,
				ScriptSig: &btcjson.ScriptSig{
					Asm: "3045022100a6dedfc533202a186902ba80354aed94448f2c348ffb219e75ffef756dd93cb602205c207114a7e9433c9f7973e2cb150d45daf02ae062141a0627e385670d6c9c23[ALL|ANYONECANPAY] 02146d0fd0e96964116bb3ef3459cfe20ad4e8b671d1b244d9d875f20a62429a20",
					Hex: "483045022100a6dedfc533202a186902ba80354aed94448f2c348ffb219e75ffef756dd93cb602205c207114a7e9433c9f7973e2cb150d45daf02ae062141a0627e385670d6c9c23812102146d0fd0e96964116bb3ef3459cfe20ad4e8b671d1b244d9d875f20a62429a20",
				},
				Sequence: 4294967295,
				Witness:  nil,
			},
			{
				Coinbase: "",
				Txid:     "248a1dac2f8824fa0c67ce7670cb7908fa4d96e47da1740d45f66c5095e4c583",
				Vout:     10,
				ScriptSig: &btcjson.ScriptSig{
					Asm: "3044022053b2cba6a230d5dba17592eb2a04193cd4061b1d651e18622ca5454c80f2617d02205e76654b63ef5134118e7a15541d129b2b292b53cb6bd80a446a21cf4612a816[ALL|ANYONECANPAY] 03ce29df4edfc3f2db18dc8691d14e9857af231db2582e46d5d27e7cbfaf9033e1[ALL|ANYONECANPAY] 02146d0fd0e96964116bb3ef3459cfe20ad4e8b671d1b244d9d875f20a62429a20",
					Hex: "473044022053b2cba6a230d5dba17592eb2a04193cd4061b1d651e18622ca5454c80f2617d02205e76654b63ef5134118e7a15541d129b2b292b53cb6bd80a446a21cf4612a816812103ce29df4edfc3f2db18dc8691d14e9857af231db2582e46d5d27e7cbfaf9033e1",
				},
				Sequence: 4294967295,
				Witness:  nil,
			},
			{
				Coinbase: "",
				Txid:     "4d97e4d570cad3b8191a1446e5453cb648e795d108faa1ab9e940c9959608843",
				Vout:     2,
				ScriptSig: &btcjson.ScriptSig{
					Asm: "304502210084d31643d8ce802cb53f013c78f79806dc74488a60a2cddba5dcde9e0578c56502201fce4445b867802b67a8fb5ff90ce9598f226ec05419005396662a012e7c3a56[ALL|ANYONECANPAY] 03309dea62304772c548acb2846962faf5c1c4f0ded1b9b1b86b7defd7e085a943[ALL|ANYONECANPAY] 03ce29df4edfc3f2db18dc8691d14e9857af231db2582e46d5d27e7cbfaf9033e1[ALL|ANYONECANPAY] 02146d0fd0e96964116bb3ef3459cfe20ad4e8b671d1b244d9d875f20a62429a20",
					Hex: "48304502210084d31643d8ce802cb53f013c78f79806dc74488a60a2cddba5dcde9e0578c56502201fce4445b867802b67a8fb5ff90ce9598f226ec05419005396662a012e7c3a56812103309dea62304772c548acb2846962faf5c1c4f0ded1b9b1b86b7defd7e085a943",
				},
				Sequence: 4294967295,
				Witness:  nil,
			},
			{
				Coinbase: "",
				Txid:     "4e959fd1d2b16e837e2d474aa84d17947abd53123a86e9e3788eb56e624a3f03",
				Vout:     3,
				ScriptSig: &btcjson.ScriptSig{
					Asm: "304402203d220c5da637ce9f2491aac4faec5fb1a31062f008cd5fa7e344063f7dbdf59602207391173aa548bf351da9984ce556aa9f292c304d14116ab99a4b76ec7ec15686[ALL|ANYONECANPAY] 036b16106b827ea3c55d2b4e3b60cbceb161e3dab7ce9f9877c999e40e90642202[ALL|ANYONECANPAY] 03309dea62304772c548acb2846962faf5c1c4f0ded1b9b1b86b7defd7e085a943[ALL|ANYONECANPAY] 03ce29df4edfc3f2db18dc8691d14e9857af231db2582e46d5d27e7cbfaf9033e1[ALL|ANYONECANPAY] 02146d0fd0e96964116bb3ef3459cfe20ad4e8b671d1b244d9d875f20a62429a20",
					Hex: "47304402203d220c5da637ce9f2491aac4faec5fb1a31062f008cd5fa7e344063f7dbdf59602207391173aa548bf351da9984ce556aa9f292c304d14116ab99a4b76ec7ec156868121036b16106b827ea3c55d2b4e3b60cbceb161e3dab7ce9f9877c999e40e90642202",
				},
				Sequence: 4294967295,
				Witness:  nil,
			},
			{
				Coinbase: "",
				Txid:     "7668c802dd56196b9b867dca2c02d0cac0e026b886029978a7c7e20bfbc6910e",
				Vout:     0,
				ScriptSig: &btcjson.ScriptSig{
					Asm: "3045022100c227a45ce2a70f13e3996e20001061bc5ab84e7fde637abb80b3f369caa6020102200ab5360ebec5f042e58c92d1bb9db67b70a3ba52575ea671b566d4b83d7dc460[ALL|ANYONECANPAY] 02af83cfa5120ebd72ab19219b59ee2a57d6e1fdeb64d47b09ceeb5f28c1d48859[ALL|ANYONECANPAY] 036b16106b827ea3c55d2b4e3b60cbceb161e3dab7ce9f9877c999e40e90642202[ALL|ANYONECANPAY] 03309dea62304772c548acb2846962faf5c1c4f0ded1b9b1b86b7defd7e085a943[ALL|ANYONECANPAY] 03ce29df4edfc3f2db18dc8691d14e9857af231db2582e46d5d27e7cbfaf9033e1[ALL|ANYONECANPAY] 02146d0fd0e96964116bb3ef3459cfe20ad4e8b671d1b244d9d875f20a62429a20",
					Hex: "483045022100c227a45ce2a70f13e3996e20001061bc5ab84e7fde637abb80b3f369caa6020102200ab5360ebec5f042e58c92d1bb9db67b70a3ba52575ea671b566d4b83d7dc460812102af83cfa5120ebd72ab19219b59ee2a57d6e1fdeb64d47b09ceeb5f28c1d48859",
				},
				Sequence: 4294967295,
				Witness:  nil,
			},
			{
				Coinbase: "",
				Txid:     "cb097cbc50770501f988688d2970302c8b7c1310d2eda2df4c0f3845a688af98",
				Vout:     10,
				ScriptSig: &btcjson.ScriptSig{
					Asm: "3045022100e12efb4a9c7a8a8962fabb380944e8e087bd7c466705f34473489a2d4042560a02204cc5d63bdbf6e5cc89e89bf2bf4a51caa89cfe3e1e19cabd3faa0f1619e099e8[ALL|ANYONECANPAY] 026ad2558403314dcc28e94f01213ed7caa8c3fe8ebafb797cfbc0cc385f63c908[ALL|ANYONECANPAY] 02af83cfa5120ebd72ab19219b59ee2a57d6e1fdeb64d47b09ceeb5f28c1d48859[ALL|ANYONECANPAY] 036b16106b827ea3c55d2b4e3b60cbceb161e3dab7ce9f9877c999e40e90642202[ALL|ANYONECANPAY] 03309dea62304772c548acb2846962faf5c1c4f0ded1b9b1b86b7defd7e085a943[ALL|ANYONECANPAY] 03ce29df4edfc3f2db18dc8691d14e9857af231db2582e46d5d27e7cbfaf9033e1[ALL|ANYONECANPAY] 02146d0fd0e96964116bb3ef3459cfe20ad4e8b671d1b244d9d875f20a62429a20",
					Hex: "483045022100e12efb4a9c7a8a8962fabb380944e8e087bd7c466705f34473489a2d4042560a02204cc5d63bdbf6e5cc89e89bf2bf4a51caa89cfe3e1e19cabd3faa0f1619e099e88121026ad2558403314dcc28e94f01213ed7caa8c3fe8ebafb797cfbc0cc385f63c908",
				},
				Sequence: 4294967295,
				Witness:  nil,
			},
		},
		Vout: []btcjson.Vout{
			{
				Value: 0.0100001,
				N:     1,
				ScriptPubKey: btcjson.ScriptPubKeyResult{
					Asm:       "OP_DUP OP_HASH160 0a00e72b922416c12779191500b100bf959f5319 OP_EQUALVERIFY OP_CHECKSIG",
					Hex:       "76a9140a00e72b922416c12779191500b100bf959f531988ac",
					ReqSigs:   1,
					Type:      "pubkeyhash",
					Addresses: []string{"XbbjfuxcuFQZyXG3dS1WAcuD48WuQmz3wr"},
				},
			},
			{
				Value: 0.0100001,
				N:     0,
				ScriptPubKey: btcjson.ScriptPubKeyResult{
					Asm:       "OP_DUP OP_HASH160 8361fc4a3b474263ae2b7a3fb229000162c57096 OP_EQUALVERIFY OP_CHECKSIG",
					Hex:       "76a9148361fc4a3b474263ae2b7a3fb229000162c5709688ac",
					ReqSigs:   1,
					Type:      "pubkeyhash",
					Addresses: []string{"XnfXjY2GFd8WRKkh37e4B7UCNsfRYS8N39"},
				},
			},
			{
				Value: 0.0100001,
				N:     2,
				ScriptPubKey: btcjson.ScriptPubKeyResult{
					Asm:       "OP_DUP OP_HASH160 b86e3ebfdc69689971a441a58d867199410529b9 OP_EQUALVERIFY OP_CHECKSIG",
					Hex:       "76a914b86e3ebfdc69689971a441a58d867199410529b988ac",
					ReqSigs:   1,
					Type:      "pubkeyhash",
					Addresses: []string{"XsW2EWqBuuX6p3L9LPG5zBQHZmqSAtQfmX"},
				},
			},
			{
				Value: 0.0100001,
				N:     3,
				ScriptPubKey: btcjson.ScriptPubKeyResult{
					Asm:       "OP_DUP OP_HASH160 c6ef7ef72e3975177c4544fa941c2b198c030ea0 OP_EQUALVERIFY OP_CHECKSIG",
					Hex:       "76a914c6ef7ef72e3975177c4544fa941c2b198c030ea088ac",
					ReqSigs:   1,
					Type:      "pubkeyhash",
					Addresses: []string{"XtpiXhx7WebRYzcP9fSRVjBpfVDFTFd9D9"},
				},
			},
			{
				Value: 0.0100001,
				N:     4,
				ScriptPubKey: btcjson.ScriptPubKeyResult{
					Asm:       "OP_DUP OP_HASH160 df6e49e6fa7a5a7b6e304f14e5ad8dc9b9192cec OP_EQUALVERIFY OP_CHECKSIG",
					Hex:       "76a914df6e49e6fa7a5a7b6e304f14e5ad8dc9b9192cec88ac",
					ReqSigs:   1,
					Type:      "pubkeyhash",
					Addresses: []string{"Xw4Ee2KNuEDYYLzKnwqwAUrCW7hXGU2s7i"},
				},
			},
			{
				Value: 0.0100001,
				N:     5,
				ScriptPubKey: btcjson.ScriptPubKeyResult{
					Asm:       "OP_DUP OP_HASH160 f7ae957e3152f3b9d23de5e551de6973c839d7a9 OP_EQUALVERIFY OP_CHECKSIG",
					Hex:       "76a914f7ae957e3152f3b9d23de5e551de6973c839d7a988ac",
					ReqSigs:   1,
					Type:      "pubkeyhash",
					Addresses: []string{"XyGTswZdfsZH6WzgY1xN9nR2jPr1mwND5o"},
				},
			},
		},
		BlockHash:     "0000000000000005fc2136de2af0c2de7d905150ba609b0a2250a896741000de",
		Confirmations: 60571,
		Time:          1613520445,
		Blocktime:     1613520445,
	}

	txMap[*txHash3] = btcjson.TxRawResult{
		Hex:      "03000500010000000000000000000000000000000000000000000000000000000000000000ffffffff2003ecb7151b5c4c55584f525c000000002988056645c219dbf6a043bb9e008b0500000000021a586008000000001976a9148bf9a55d3864e49d0bb9f8cc75b6ab3949526ab888ac9ad3d208000000001976a91458e09f6b914214ef3bf2907dabd496822616654388ac00000000460200ecb715009bd04fb02da560fdd1d3640c6db267ec8ab05416cf87befa7dfb5c24b6b4b59b19f8270cdb792df824dfa25504fdff26b879c793fc800ae21a8ce061a8f9ebdc",
		Txid:     "9b75129c5d8538287e3c34035da0420905b6ff1651f06a6f292392bc825a4af5",
		Hash:     "",
		Size:     222,
		Vsize:    0,
		Weight:   0,
		Version:  3,
		LockTime: 0,
		Vin: []btcjson.Vin{
			{
				Coinbase:  "03ecb7151b5c4c55584f525c000000002988056645c219dbf6a043bb9e008b05",
				Txid:      "",
				Vout:      0,
				ScriptSig: nil,
				Sequence:  0,
				Witness:   nil,
			},
		},
		Vout: []btcjson.Vout{
			{
				Value: 1.40531738,
				N:     0,
				ScriptPubKey: btcjson.ScriptPubKeyResult{
					Asm:       "OP_DUP OP_HASH160 8bf9a55d3864e49d0bb9f8cc75b6ab3949526ab8 OP_EQUALVERIFY OP_CHECKSIG",
					Hex:       "76a9148bf9a55d3864e49d0bb9f8cc75b6ab3949526ab888ac",
					ReqSigs:   1,
					Type:      "pubkeyhash",
					Addresses: []string{"XoSxpd5VYNQvKbXbEaDKt6P1aZANzAkXrJ"},
				},
			},
			{
				Value: 1.48034458,
				N:     1,
				ScriptPubKey: btcjson.ScriptPubKeyResult{
					Asm:       "OP_DUP OP_HASH160 58e09f6b914214ef3bf2907dabd4968226166543 OP_EQUALVERIFY OP_CHECKSIG",
					Hex:       "76a91458e09f6b914214ef3bf2907dabd496822616654388ac",
					ReqSigs:   1,
					Type:      "pubkeyhash",
					Addresses: []string{"XinnPCJsEqyvHiusgikA1bs7aVDBq42f9K"},
				},
			},
		},
		BlockHash:     "0000000000000005fc2136de2af0c2de7d905150ba609b0a2250a896741000de",
		Confirmations: 60571,
		Time:          1613520445,
		Blocktime:     1613520445,
	}

	return mockRPC{
		blockHeight:    50000,
		blockchainInfo: blkInfo,
		blockStore:     store,
		txStore:        txMap,
		heightMap:      hMap,
	}
}

func (m *mockRPC) GetBlockCount() (int64, error) {
	_ = m.Called()
	return m.blockHeight, nil
}

func (m *mockRPC) GetBlockVerbose(blockHash *chainhash.Hash) (*btcjson.GetBlockVerboseResult, error) {
	block, ok := m.blockStore[*blockHash]
	if ok {
		return &block, nil
	}

	return nil, errors.New("block not found")
}

func (m *mockRPC) GetBlockChainInfo() (*btcjson.GetBlockChainInfoResult, error) {
	return &m.blockchainInfo, nil
}

func (m *mockRPC) GetBlockHash(blockHeight int64) (*chainhash.Hash, error) {
	hash, ok := m.heightMap[blockHeight]
	if ok {
		return &hash, nil
	}

	return nil, errors.New("hash not found")
}

func (m *mockRPC) GetRawTransactionVerbose(txHash *chainhash.Hash) (*btcjson.TxRawResult, error) {
	tx, ok := m.txStore[*txHash]
	if ok {
		return &tx, nil
	}

	return nil, errors.New("transaction not found")
}

func (m *mockRPC) GetBlockChainInfoAsync() rpcclient.FutureGetBlockChainInfoResult {
	return m.futureBlockchainInfo
}

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
	outputUid := createOutputUid("asdf", 50)
	require.NotEmpty(t, outputUid)
	if len(outputUid) < 2 {
		t.Fatal("output uid is too short:", outputUid)
	}
	require.EqualValues(t, "_:", outputUid[:2])
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
		Uid:         "0x59b84",
		OutputIndex: &zero,
		InputIndex:  nil,
		TxType:      "pubkeyhash",
		Amount:      &one64,
		IsCoinbase:  &wrong,
		DType:       nil,
	}

	output2 := dbop.Output{
		Uid:         "0x59b85",
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
		Uid:  "",
		Hash: fistAddress,
		Outputs: []dbop.Output{{
			Uid:         "0x59b81",
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
		} else if val.Outputs[0].Uid != "0x59b81" ||
			val.Outputs[1].Uid != "0x59b84" {
			t.Error("Uids not set")
		}
	}
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
