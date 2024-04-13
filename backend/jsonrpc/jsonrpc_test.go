package jsonrpc

import (
	cli "backend/cmd/cliutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/rpcclient"
	"testing"
)

func getOldRPC() (*rpcclient.Client, error) {
	rpcEndpoint, err := cli.BuildEndpoint("0.0.0.0", 9998)
	if err != nil {
		return nil, err
	}

	connection := rpcclient.ConnConfig{
		Host:                rpcEndpoint,
		User:                "rpc1user",
		Pass:                "1234pass",
		DisableConnectOnNew: true,
		DisableTLS:          true,
		HTTPPostMode:        true,
	}

	client, err := rpcclient.New(&connection, nil)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func getOldBatchRPC() (*rpcclient.Client, error) {
	rpcEndpoint, err := cli.BuildEndpoint("0.0.0.0", 9998)
	if err != nil {
		return nil, err
	}

	connection := rpcclient.ConnConfig{
		Host:                rpcEndpoint,
		User:                "rpc1user",
		Pass:                "1234pass",
		DisableConnectOnNew: true,
		DisableTLS:          true,
		HTTPPostMode:        true,
	}

	client, err := rpcclient.NewBatch(&connection)
	if err != nil {
		return nil, err
	}
	return client, nil
}

func BenchmarkGetBlockCount(b *testing.B) {
	rpc := NewDashClient("0.0.0.0:9998", "rpc1user", "1234pass", nil)

	for range b.N {
		_, err := rpc.GetBlockCount()
		if err != nil {
			b.Error(err)
			return
		}
	}
}

func BenchmarkGetBlockCount_old(b *testing.B) {
	rpc, err := getOldRPC()
	if err != nil {
		b.Error(err)
		return
	}
	for range b.N {
		_, err := rpc.GetBlockCount()
		if err != nil {
			b.Error(err)
			return
		}
	}
}

func BenchmarkGetBlockHash(b *testing.B) {
	rpc := NewDashClient("0.0.0.0:9998", "rpc1user", "1234pass", nil)

	for i := range b.N {
		_, err := rpc.GetBlockHash(100000 + int64(i))
		if err != nil {
			b.Error(err)
			return
		}
	}
}

func BenchmarkGetBlockHash_old(b *testing.B) {
	rpc, err := getOldRPC()
	if err != nil {
		b.Error(err)
		return
	}
	for i := range b.N {
		_, err := rpc.GetBlockHash(100000 + int64(i))
		if err != nil {
			b.Error(err)
			return
		}
	}
}

var testTransactions4 = []string{
	"429c8f23491a597f8961912a4813ad057ee9f28dea4a67b9ecd7697ce950ae40",
	"1a92c7f34a041082fe485f9e17bc8ee84b88d31ce063387e1d192da89e49f596",
	"1ed18374598734fe718226bb08b8f8f6c40d9e2e24da7f5323d506e6c8970c0c",
	"650dbdc2562578236ffbd479a7aa05bb7fa233645d3b284230afdb109fde59da",
}

var testTransactions40 = []string{
	"429c8f23491a597f8961912a4813ad057ee9f28dea4a67b9ecd7697ce950ae40",
	"1a92c7f34a041082fe485f9e17bc8ee84b88d31ce063387e1d192da89e49f596",
	"1ed18374598734fe718226bb08b8f8f6c40d9e2e24da7f5323d506e6c8970c0c",
	"650dbdc2562578236ffbd479a7aa05bb7fa233645d3b284230afdb109fde59da",
	"429c8f23491a597f8961912a4813ad057ee9f28dea4a67b9ecd7697ce950ae40",
	"1a92c7f34a041082fe485f9e17bc8ee84b88d31ce063387e1d192da89e49f596",
	"1ed18374598734fe718226bb08b8f8f6c40d9e2e24da7f5323d506e6c8970c0c",
	"650dbdc2562578236ffbd479a7aa05bb7fa233645d3b284230afdb109fde59da",
	"429c8f23491a597f8961912a4813ad057ee9f28dea4a67b9ecd7697ce950ae40",
	"1a92c7f34a041082fe485f9e17bc8ee84b88d31ce063387e1d192da89e49f596",
	"1ed18374598734fe718226bb08b8f8f6c40d9e2e24da7f5323d506e6c8970c0c",
	"650dbdc2562578236ffbd479a7aa05bb7fa233645d3b284230afdb109fde59da",
	"429c8f23491a597f8961912a4813ad057ee9f28dea4a67b9ecd7697ce950ae40",
	"1a92c7f34a041082fe485f9e17bc8ee84b88d31ce063387e1d192da89e49f596",
	"1ed18374598734fe718226bb08b8f8f6c40d9e2e24da7f5323d506e6c8970c0c",
	"650dbdc2562578236ffbd479a7aa05bb7fa233645d3b284230afdb109fde59da",
	"429c8f23491a597f8961912a4813ad057ee9f28dea4a67b9ecd7697ce950ae40",
	"1a92c7f34a041082fe485f9e17bc8ee84b88d31ce063387e1d192da89e49f596",
	"1ed18374598734fe718226bb08b8f8f6c40d9e2e24da7f5323d506e6c8970c0c",
	"650dbdc2562578236ffbd479a7aa05bb7fa233645d3b284230afdb109fde59da",
	"429c8f23491a597f8961912a4813ad057ee9f28dea4a67b9ecd7697ce950ae40",
	"1a92c7f34a041082fe485f9e17bc8ee84b88d31ce063387e1d192da89e49f596",
	"1ed18374598734fe718226bb08b8f8f6c40d9e2e24da7f5323d506e6c8970c0c",
	"650dbdc2562578236ffbd479a7aa05bb7fa233645d3b284230afdb109fde59da",
	"429c8f23491a597f8961912a4813ad057ee9f28dea4a67b9ecd7697ce950ae40",
	"1a92c7f34a041082fe485f9e17bc8ee84b88d31ce063387e1d192da89e49f596",
	"1ed18374598734fe718226bb08b8f8f6c40d9e2e24da7f5323d506e6c8970c0c",
	"650dbdc2562578236ffbd479a7aa05bb7fa233645d3b284230afdb109fde59da",
	"429c8f23491a597f8961912a4813ad057ee9f28dea4a67b9ecd7697ce950ae40",
	"1a92c7f34a041082fe485f9e17bc8ee84b88d31ce063387e1d192da89e49f596",
	"1ed18374598734fe718226bb08b8f8f6c40d9e2e24da7f5323d506e6c8970c0c",
	"650dbdc2562578236ffbd479a7aa05bb7fa233645d3b284230afdb109fde59da",
	"429c8f23491a597f8961912a4813ad057ee9f28dea4a67b9ecd7697ce950ae40",
	"1a92c7f34a041082fe485f9e17bc8ee84b88d31ce063387e1d192da89e49f596",
	"1ed18374598734fe718226bb08b8f8f6c40d9e2e24da7f5323d506e6c8970c0c",
	"650dbdc2562578236ffbd479a7aa05bb7fa233645d3b284230afdb109fde59da",
	"429c8f23491a597f8961912a4813ad057ee9f28dea4a67b9ecd7697ce950ae40",
	"1a92c7f34a041082fe485f9e17bc8ee84b88d31ce063387e1d192da89e49f596",
	"1ed18374598734fe718226bb08b8f8f6c40d9e2e24da7f5323d506e6c8970c0c",
	"650dbdc2562578236ffbd479a7aa05bb7fa233645d3b284230afdb109fde59da",
	"429c8f23491a597f8961912a4813ad057ee9f28dea4a67b9ecd7697ce950ae40",
	"1a92c7f34a041082fe485f9e17bc8ee84b88d31ce063387e1d192da89e49f596",
	"1ed18374598734fe718226bb08b8f8f6c40d9e2e24da7f5323d506e6c8970c0c",
	"650dbdc2562578236ffbd479a7aa05bb7fa233645d3b284230afdb109fde59da",
	"429c8f23491a597f8961912a4813ad057ee9f28dea4a67b9ecd7697ce950ae40",
	"1a92c7f34a041082fe485f9e17bc8ee84b88d31ce063387e1d192da89e49f596",
	"1ed18374598734fe718226bb08b8f8f6c40d9e2e24da7f5323d506e6c8970c0c",
	"650dbdc2562578236ffbd479a7aa05bb7fa233645d3b284230afdb109fde59da",
	"429c8f23491a597f8961912a4813ad057ee9f28dea4a67b9ecd7697ce950ae40",
	"1a92c7f34a041082fe485f9e17bc8ee84b88d31ce063387e1d192da89e49f596",
	"1ed18374598734fe718226bb08b8f8f6c40d9e2e24da7f5323d506e6c8970c0c",
	"650dbdc2562578236ffbd479a7aa05bb7fa233645d3b284230afdb109fde59da",
	"429c8f23491a597f8961912a4813ad057ee9f28dea4a67b9ecd7697ce950ae40",
	"1a92c7f34a041082fe485f9e17bc8ee84b88d31ce063387e1d192da89e49f596",
	"1ed18374598734fe718226bb08b8f8f6c40d9e2e24da7f5323d506e6c8970c0c",
	"650dbdc2562578236ffbd479a7aa05bb7fa233645d3b284230afdb109fde59da",
	"429c8f23491a597f8961912a4813ad057ee9f28dea4a67b9ecd7697ce950ae40",
	"1a92c7f34a041082fe485f9e17bc8ee84b88d31ce063387e1d192da89e49f596",
	"1ed18374598734fe718226bb08b8f8f6c40d9e2e24da7f5323d506e6c8970c0c",
	"650dbdc2562578236ffbd479a7aa05bb7fa233645d3b284230afdb109fde59da",
	"429c8f23491a597f8961912a4813ad057ee9f28dea4a67b9ecd7697ce950ae40",
	"1a92c7f34a041082fe485f9e17bc8ee84b88d31ce063387e1d192da89e49f596",
	"1ed18374598734fe718226bb08b8f8f6c40d9e2e24da7f5323d506e6c8970c0c",
	"650dbdc2562578236ffbd479a7aa05bb7fa233645d3b284230afdb109fde59da",
	"429c8f23491a597f8961912a4813ad057ee9f28dea4a67b9ecd7697ce950ae40",
	"1a92c7f34a041082fe485f9e17bc8ee84b88d31ce063387e1d192da89e49f596",
	"1ed18374598734fe718226bb08b8f8f6c40d9e2e24da7f5323d506e6c8970c0c",
	"650dbdc2562578236ffbd479a7aa05bb7fa233645d3b284230afdb109fde59da",
	"429c8f23491a597f8961912a4813ad057ee9f28dea4a67b9ecd7697ce950ae40",
	"1a92c7f34a041082fe485f9e17bc8ee84b88d31ce063387e1d192da89e49f596",
	"1ed18374598734fe718226bb08b8f8f6c40d9e2e24da7f5323d506e6c8970c0c",
	"650dbdc2562578236ffbd479a7aa05bb7fa233645d3b284230afdb109fde59da",
	"429c8f23491a597f8961912a4813ad057ee9f28dea4a67b9ecd7697ce950ae40",
	"1a92c7f34a041082fe485f9e17bc8ee84b88d31ce063387e1d192da89e49f596",
	"1ed18374598734fe718226bb08b8f8f6c40d9e2e24da7f5323d506e6c8970c0c",
	"650dbdc2562578236ffbd479a7aa05bb7fa233645d3b284230afdb109fde59da",
	"429c8f23491a597f8961912a4813ad057ee9f28dea4a67b9ecd7697ce950ae40",
	"1a92c7f34a041082fe485f9e17bc8ee84b88d31ce063387e1d192da89e49f596",
	"1ed18374598734fe718226bb08b8f8f6c40d9e2e24da7f5323d506e6c8970c0c",
	"650dbdc2562578236ffbd479a7aa05bb7fa233645d3b284230afdb109fde59da",
	"429c8f23491a597f8961912a4813ad057ee9f28dea4a67b9ecd7697ce950ae40",
	"1a92c7f34a041082fe485f9e17bc8ee84b88d31ce063387e1d192da89e49f596",
	"1ed18374598734fe718226bb08b8f8f6c40d9e2e24da7f5323d506e6c8970c0c",
	"650dbdc2562578236ffbd479a7aa05bb7fa233645d3b284230afdb109fde59da",
	"429c8f23491a597f8961912a4813ad057ee9f28dea4a67b9ecd7697ce950ae40",
	"1a92c7f34a041082fe485f9e17bc8ee84b88d31ce063387e1d192da89e49f596",
	"1ed18374598734fe718226bb08b8f8f6c40d9e2e24da7f5323d506e6c8970c0c",
	"650dbdc2562578236ffbd479a7aa05bb7fa233645d3b284230afdb109fde59da",
	"429c8f23491a597f8961912a4813ad057ee9f28dea4a67b9ecd7697ce950ae40",
	"1a92c7f34a041082fe485f9e17bc8ee84b88d31ce063387e1d192da89e49f596",
	"1ed18374598734fe718226bb08b8f8f6c40d9e2e24da7f5323d506e6c8970c0c",
	"650dbdc2562578236ffbd479a7aa05bb7fa233645d3b284230afdb109fde59da",
	"429c8f23491a597f8961912a4813ad057ee9f28dea4a67b9ecd7697ce950ae40",
	"1a92c7f34a041082fe485f9e17bc8ee84b88d31ce063387e1d192da89e49f596",
	"1ed18374598734fe718226bb08b8f8f6c40d9e2e24da7f5323d506e6c8970c0c",
	"650dbdc2562578236ffbd479a7aa05bb7fa233645d3b284230afdb109fde59da",
	"429c8f23491a597f8961912a4813ad057ee9f28dea4a67b9ecd7697ce950ae40",
	"1a92c7f34a041082fe485f9e17bc8ee84b88d31ce063387e1d192da89e49f596",
	"1ed18374598734fe718226bb08b8f8f6c40d9e2e24da7f5323d506e6c8970c0c",
	"650dbdc2562578236ffbd479a7aa05bb7fa233645d3b284230afdb109fde59da",
	"429c8f23491a597f8961912a4813ad057ee9f28dea4a67b9ecd7697ce950ae40",
	"1a92c7f34a041082fe485f9e17bc8ee84b88d31ce063387e1d192da89e49f596",
	"1ed18374598734fe718226bb08b8f8f6c40d9e2e24da7f5323d506e6c8970c0c",
	"650dbdc2562578236ffbd479a7aa05bb7fa233645d3b284230afdb109fde59da",
	"429c8f23491a597f8961912a4813ad057ee9f28dea4a67b9ecd7697ce950ae40",
	"1a92c7f34a041082fe485f9e17bc8ee84b88d31ce063387e1d192da89e49f596",
	"1ed18374598734fe718226bb08b8f8f6c40d9e2e24da7f5323d506e6c8970c0c",
	"650dbdc2562578236ffbd479a7aa05bb7fa233645d3b284230afdb109fde59da",
	"429c8f23491a597f8961912a4813ad057ee9f28dea4a67b9ecd7697ce950ae40",
	"1a92c7f34a041082fe485f9e17bc8ee84b88d31ce063387e1d192da89e49f596",
	"1ed18374598734fe718226bb08b8f8f6c40d9e2e24da7f5323d506e6c8970c0c",
	"650dbdc2562578236ffbd479a7aa05bb7fa233645d3b284230afdb109fde59da",
	"429c8f23491a597f8961912a4813ad057ee9f28dea4a67b9ecd7697ce950ae40",
	"1a92c7f34a041082fe485f9e17bc8ee84b88d31ce063387e1d192da89e49f596",
	"1ed18374598734fe718226bb08b8f8f6c40d9e2e24da7f5323d506e6c8970c0c",
	"650dbdc2562578236ffbd479a7aa05bb7fa233645d3b284230afdb109fde59da",
	"429c8f23491a597f8961912a4813ad057ee9f28dea4a67b9ecd7697ce950ae40",
	"1a92c7f34a041082fe485f9e17bc8ee84b88d31ce063387e1d192da89e49f596",
	"1ed18374598734fe718226bb08b8f8f6c40d9e2e24da7f5323d506e6c8970c0c",
	"650dbdc2562578236ffbd479a7aa05bb7fa233645d3b284230afdb109fde59da",
	"429c8f23491a597f8961912a4813ad057ee9f28dea4a67b9ecd7697ce950ae40",
	"1a92c7f34a041082fe485f9e17bc8ee84b88d31ce063387e1d192da89e49f596",
	"1ed18374598734fe718226bb08b8f8f6c40d9e2e24da7f5323d506e6c8970c0c",
	"650dbdc2562578236ffbd479a7aa05bb7fa233645d3b284230afdb109fde59da",
	"429c8f23491a597f8961912a4813ad057ee9f28dea4a67b9ecd7697ce950ae40",
	"1a92c7f34a041082fe485f9e17bc8ee84b88d31ce063387e1d192da89e49f596",
	"1ed18374598734fe718226bb08b8f8f6c40d9e2e24da7f5323d506e6c8970c0c",
	"650dbdc2562578236ffbd479a7aa05bb7fa233645d3b284230afdb109fde59da",
	"429c8f23491a597f8961912a4813ad057ee9f28dea4a67b9ecd7697ce950ae40",
	"1a92c7f34a041082fe485f9e17bc8ee84b88d31ce063387e1d192da89e49f596",
	"1ed18374598734fe718226bb08b8f8f6c40d9e2e24da7f5323d506e6c8970c0c",
	"650dbdc2562578236ffbd479a7aa05bb7fa233645d3b284230afdb109fde59da",
	"429c8f23491a597f8961912a4813ad057ee9f28dea4a67b9ecd7697ce950ae40",
	"1a92c7f34a041082fe485f9e17bc8ee84b88d31ce063387e1d192da89e49f596",
	"1ed18374598734fe718226bb08b8f8f6c40d9e2e24da7f5323d506e6c8970c0c",
	"650dbdc2562578236ffbd479a7aa05bb7fa233645d3b284230afdb109fde59da",
	"429c8f23491a597f8961912a4813ad057ee9f28dea4a67b9ecd7697ce950ae40",
	"1a92c7f34a041082fe485f9e17bc8ee84b88d31ce063387e1d192da89e49f596",
	"1ed18374598734fe718226bb08b8f8f6c40d9e2e24da7f5323d506e6c8970c0c",
	"650dbdc2562578236ffbd479a7aa05bb7fa233645d3b284230afdb109fde59da",
	"429c8f23491a597f8961912a4813ad057ee9f28dea4a67b9ecd7697ce950ae40",
	"1a92c7f34a041082fe485f9e17bc8ee84b88d31ce063387e1d192da89e49f596",
	"1ed18374598734fe718226bb08b8f8f6c40d9e2e24da7f5323d506e6c8970c0c",
	"650dbdc2562578236ffbd479a7aa05bb7fa233645d3b284230afdb109fde59da",
	"429c8f23491a597f8961912a4813ad057ee9f28dea4a67b9ecd7697ce950ae40",
	"1a92c7f34a041082fe485f9e17bc8ee84b88d31ce063387e1d192da89e49f596",
	"1ed18374598734fe718226bb08b8f8f6c40d9e2e24da7f5323d506e6c8970c0c",
	"650dbdc2562578236ffbd479a7aa05bb7fa233645d3b284230afdb109fde59da",
	"429c8f23491a597f8961912a4813ad057ee9f28dea4a67b9ecd7697ce950ae40",
	"1a92c7f34a041082fe485f9e17bc8ee84b88d31ce063387e1d192da89e49f596",
	"1ed18374598734fe718226bb08b8f8f6c40d9e2e24da7f5323d506e6c8970c0c",
	"650dbdc2562578236ffbd479a7aa05bb7fa233645d3b284230afdb109fde59da",
	"429c8f23491a597f8961912a4813ad057ee9f28dea4a67b9ecd7697ce950ae40",
	"1a92c7f34a041082fe485f9e17bc8ee84b88d31ce063387e1d192da89e49f596",
	"1ed18374598734fe718226bb08b8f8f6c40d9e2e24da7f5323d506e6c8970c0c",
	"650dbdc2562578236ffbd479a7aa05bb7fa233645d3b284230afdb109fde59da",
	"429c8f23491a597f8961912a4813ad057ee9f28dea4a67b9ecd7697ce950ae40",
	"1a92c7f34a041082fe485f9e17bc8ee84b88d31ce063387e1d192da89e49f596",
	"1ed18374598734fe718226bb08b8f8f6c40d9e2e24da7f5323d506e6c8970c0c",
	"650dbdc2562578236ffbd479a7aa05bb7fa233645d3b284230afdb109fde59da",
}

func BenchmarkGetRawTransactionVerbose(b *testing.B) {
	rpc := NewDashClient("0.0.0.0:9998", "rpc1user", "1234pass", nil)

	for i := range b.N {
		_, err := rpc.GetRawTransactionVerbose(testTransactions4[i%len(testTransactions4)])
		if err != nil {
			b.Error(err)
			return
		}
	}
}
func BenchmarkGetRawTransactionVerbose_old(b *testing.B) {
	rpc, err := getOldRPC()
	if err != nil {
		b.Error(err)
		return
	}

	hashes := make([]*chainhash.Hash, len(testTransactions4))
	for i, tx := range testTransactions4 {
		h, err := chainhash.NewHashFromStr(tx)
		if err != nil {
			b.Error(err)
			return
		}

		hashes[i] = h
	}

	for i := range b.N {
		_, err := rpc.GetRawTransactionVerbose(hashes[i%len(hashes)])
		if err != nil {
			b.Error(err)
			return
		}
	}
}

func BenchmarkGetRawTransactionVerboseSimulatedBatch(b *testing.B) {
	rpc := NewDashClient("0.0.0.0:9998", "rpc1user", "1234pass", nil)

	for range b.N {
		for y := range len(testTransactions40) {
			_, err := rpc.GetRawTransactionVerbose(testTransactions40[y])
			if err != nil {
				b.Error(err)
				return
			}
		}
	}
}

func BenchmarkGetRawTransactionVerboseBatch_old(b *testing.B) {
	rpc, err := getOldBatchRPC()
	if err != nil {
		b.Error(err)
		return
	}

	hashes := make([]*chainhash.Hash, len(testTransactions40))
	for i, tx := range testTransactions40 {
		h, err := chainhash.NewHashFromStr(tx)
		if err != nil {
			b.Error(err)
			return
		}

		hashes[i] = h
	}

	for range b.N {
		var futures []rpcclient.FutureGetRawTransactionVerboseResult
		for y := range len(hashes) {
			futures = append(futures, rpc.GetRawTransactionVerboseAsync(hashes[y]))
		}

		if err := rpc.Send(); err != nil {
			b.Error(err)
			return
		}

		for _, f := range futures {
			_, err := f.Receive()
			if err != nil {
				b.Error(err)
				return
			}
		}
	}
}

func BenchmarkGetRawTransactionVerboseBatch(b *testing.B) {
	rpc := NewDashClient("0.0.0.0:9998", "rpc1user", "1234pass", nil)

	for range b.N {
		_, err := rpc.GetRawTransactionVerboseBatch(testTransactions40)
		if err != nil {
			b.Error(err)
			return
		}
	}
}
