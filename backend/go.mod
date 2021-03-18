module backend

go 1.16

require (
	github.com/btcsuite/btcd v0.20.1-beta
	github.com/btcsuite/btcutil v1.0.2
	github.com/dgraph-io/dgo/v2 v2.2.0
	github.com/dgraph-io/ristretto v0.0.3
	github.com/o1egl/paseto v1.0.0
	github.com/test/test v0.0.0-00010101000000-000000000000
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	golang.org/x/net v0.0.0-20200822124328-c89045814202 // indirect
	google.golang.org/grpc v1.30.0
)

replace github.com/btcsuite/btcd => github.com/decfi/btcd v0.21.0-beta-dakar

replace github.com/test/test => github.com/btcsuite/btcd v0.21.0-beta
