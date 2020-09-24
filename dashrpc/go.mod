module dashrpc

go 1.14

require (
	github.com/btcsuite/btcd v0.20.1-beta
	github.com/btcsuite/btclog v0.0.0-20170628155309-84c8d2346e9f
	github.com/btcsuite/btcutil v1.0.2
	github.com/dgraph-io/dgo/v2 v2.2.0
	github.com/kr/pretty v0.2.0 // indirect
	golang.org/x/lint v0.0.0-20200302205851-738671d3881b // indirect
	golang.org/x/tools v0.0.0-20200923182640-463111b69878 // indirect
	google.golang.org/grpc v1.30.0
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
)

replace github.com/btcsuite/btcd => ./btcsuite/btcd

replace github.com/btcsuite/btcd/chaincfg => ./btcsuite/btcd/chaincfg

replace github.com/btcsuite/btcutil => ./btcsuite/btcutil

replace github.com/btcsuite/btclog => ./btcsuite/btclog
