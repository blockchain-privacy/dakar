module backend

go 1.16

require (
	github.com/btcsuite/btcd v0.20.1-beta
	github.com/btcsuite/btcutil v1.0.2
	github.com/dgraph-io/dgo/v210 v210.0.0-20210421093152-78a2fece3ebd
	github.com/dgraph-io/ristretto v0.0.3
	github.com/go-echarts/go-echarts/v2 v2.2.4
	github.com/o1egl/paseto v1.0.0
	github.com/wcharczuk/go-chart/v2 v2.1.0
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	gonum.org/v1/gonum v0.9.1
	google.golang.org/grpc v1.38.0
)

replace github.com/btcsuite/btcd => github.com/decfi/btcd v0.21.0-beta-dakar
