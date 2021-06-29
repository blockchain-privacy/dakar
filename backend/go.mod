module backend

go 1.16

require (
	github.com/aead/chacha20poly1305 v0.0.0-20201124145622-1a5aba2a8b29 // indirect
	github.com/btcsuite/btcd v0.22.0-beta
	github.com/btcsuite/btcutil v1.0.2
	github.com/dgraph-io/dgo/v210 v210.0.0-20210421093152-78a2fece3ebd
	github.com/dgraph-io/ristretto v0.1.0
	github.com/go-echarts/go-echarts/v2 v2.2.4
	github.com/golang/glog v0.0.0-20210429001901-424d2337a529 // indirect
	github.com/o1egl/paseto v1.0.0
	github.com/prometheus/client_golang v1.11.0
	github.com/prometheus/common v0.29.0 // indirect
	github.com/stretchr/objx v0.3.0 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/wcharczuk/go-chart/v2 v2.1.0
	golang.org/x/crypto v0.0.0-20210616213533-5ff15b29337e
	golang.org/x/image v0.0.0-20210628002857-a66eb6448b8d // indirect
	golang.org/x/net v0.0.0-20210614182718-04defd469f4e // indirect
	golang.org/x/sys v0.0.0-20210616094352-59db8d763f22 // indirect
	gonum.org/v1/gonum v0.9.2
	google.golang.org/genproto v0.0.0-20210624195500-8bfb893ecb84 // indirect
	google.golang.org/grpc v1.38.0
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
)

replace github.com/btcsuite/btcd => github.com/decfi/btcd v0.21.0-beta-dakar
