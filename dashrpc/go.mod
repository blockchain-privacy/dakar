module dashrpc

go 1.14

require (
	github.com/btcsuite/btcd v0.20.1-beta
	github.com/btcsuite/btclog v0.0.0-20170628155309-84c8d2346e9f
	github.com/btcsuite/btcutil v1.0.2
	github.com/dgraph-io/badger/v2 v2.0.3
	github.com/dgraph-io/ristretto v0.0.2 // indirect
	github.com/kr/pretty v0.2.0 // indirect
	github.com/pkg/errors v0.8.1
)

replace github.com/btcsuite/btcd => ./btcsuite/btcd

replace github.com/btcsuite/btcd/chaincfg => ./btcsuite/btcd/chaincfg

replace github.com/btcsuite/btcutil => ./btcsuite/btcutil

replace github.com/btcsuite/btclog => ./btcsuite/btclog
