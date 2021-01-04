module backend

go 1.14

require (
	github.com/btcsuite/btcd v0.20.1-beta
	github.com/btcsuite/btcutil v1.0.2
	github.com/dgraph-io/dgo/v2 v2.2.0
	github.com/kr/pretty v0.2.0 // indirect
	golang.org/x/net v0.0.0-20200822124328-c89045814202 // indirect
	golang.org/x/xerrors v0.0.0-20200804184101-5ec99f83aff1 // indirect
	google.golang.org/grpc v1.30.0
	gopkg.in/check.v1 v1.0.0-20190902080502-41f04d3bba15 // indirect
)

replace github.com/btcsuite/btcd => ./btcsuite/btcd

replace github.com/btcsuite/btcutil => ./btcsuite/btcutil

replace github.com/btcsuite/btclog => ./btcsuite/btclog
