module dashrpc

go 1.12

require (
	github.com/AndreasBriese/bbloom v0.0.0-20190306092124-e2d15f34fcf9 // indirect
	github.com/btcsuite/btcd v0.0.0-20190427004231-96897255fd17
	github.com/btcsuite/btclog v0.0.0-20170628155309-84c8d2346e9f
	github.com/btcsuite/btcutil v0.0.0-20190425235716-9e5f4b9a998d
	github.com/dgraph-io/badger v1.5.4
	github.com/dgryski/go-farm v0.0.0-20190423205320-6a90982ecee2 // indirect
	github.com/pkg/errors v0.8.1
	github.com/stretchr/testify v1.3.0 // indirect
)

replace github.com/btcsuite/btcd => ./btcsuite/btcd

replace github.com/btcsuite/btcd/chaincfg => ./btcsuite/btcd/chaincfg

replace github.com/btcsuite/btcutil => ./btcsuite/btcutil
