module dashrpc

go 1.13

require (
	github.com/btcsuite/btcd v0.0.0-20190427004231-96897255fd17
	github.com/btcsuite/btclog v0.0.0-20170628155309-84c8d2346e9f
	github.com/btcsuite/btcutil v0.0.0-20190425235716-9e5f4b9a998d
	github.com/dgraph-io/badger v1.6.0
	github.com/pkg/errors v0.8.1
)

replace github.com/btcsuite/btcd => ./btcsuite/btcd

replace github.com/btcsuite/btcd/chaincfg => ./btcsuite/btcd/chaincfg

replace github.com/btcsuite/btcutil => ./btcsuite/btcutil
