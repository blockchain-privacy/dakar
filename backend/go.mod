module backend

go 1.18

require (
	github.com/btcsuite/btcd v0.23.2
	github.com/btcsuite/btcd/btcutil v1.1.2
	github.com/btcsuite/btcd/chaincfg/chainhash v1.0.1
	github.com/dgraph-io/dgo/v210 v210.0.0-20220113041351-ba0e5dfc4c3e
	github.com/dgraph-io/ristretto v0.1.0
	github.com/ory/kratos-client-go v0.10.1
	github.com/prometheus/client_golang v1.13.0
	github.com/stretchr/testify v1.8.0
	gonum.org/v1/gonum v0.12.0
	google.golang.org/grpc v1.50.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/btcsuite/btcd/btcec/v2 v2.2.1 // indirect
	github.com/btcsuite/btclog v0.0.0-20170628155309-84c8d2346e9f // indirect
	github.com/btcsuite/go-socks v0.0.0-20170105172521-4720035b7bfd // indirect
	github.com/btcsuite/websocket v0.0.0-20150119174127-31079b680792 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/decred/dcrd/crypto/blake256 v1.0.0 // indirect
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.1.0 // indirect
	github.com/dustin/go-humanize v1.0.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/glog v1.0.0 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/kr/pretty v0.2.1 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.2 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.2.0 // indirect
	github.com/prometheus/common v0.37.0 // indirect
	github.com/prometheus/procfs v0.8.0 // indirect
	golang.org/x/crypto v0.0.0-20221005025214-4161e89ecf1b // indirect
	golang.org/x/net v0.0.0-20221004154528-8021a29435af // indirect
	golang.org/x/oauth2 v0.0.0-20221006150949-b44042a4b9c1 // indirect
	golang.org/x/sys v0.0.0-20221006211917-84dc82d7e875 // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/genproto v0.0.0-20220930163606-c98284e70a91 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
)

replace github.com/btcsuite/btcd => github.com/qrest/btcd v0.23.2-dakar
