module backend

go 1.16

require (
	github.com/aead/chacha20poly1305 v0.0.0-20201124145622-1a5aba2a8b29 // indirect
	github.com/btcsuite/btcd v0.22.0-beta
	github.com/btcsuite/btcutil v1.0.3-0.20201208143702-a53e38424cce
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/dgraph-io/dgo/v210 v210.0.0-20210825123656-d3f867fe9cc3
	github.com/dgraph-io/ristretto v0.1.0
	github.com/go-echarts/go-echarts/v2 v2.2.4
	github.com/golang/glog v1.0.0 // indirect
	github.com/o1egl/paseto v1.0.0
	github.com/prometheus/client_golang v1.11.0
	github.com/prometheus/common v0.30.0 // indirect
	github.com/prometheus/procfs v0.7.3 // indirect
	github.com/stretchr/objx v0.3.0 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/wcharczuk/go-chart/v2 v2.1.0
	golang.org/x/crypto v0.0.0-20210920023735-84f357641f63
	golang.org/x/image v0.0.0-20210628002857-a66eb6448b8d // indirect
	golang.org/x/net v0.0.0-20210917221730-978cfadd31cf // indirect
	golang.org/x/sys v0.0.0-20210921065528-437939a70204 // indirect
	golang.org/x/text v0.3.7 // indirect
	gonum.org/v1/gonum v0.9.3
	google.golang.org/genproto v0.0.0-20210920155426-26f343e4c215 // indirect
	google.golang.org/grpc v1.40.0
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b
)

replace github.com/btcsuite/btcd => github.com/decfi/btcd v0.22.0-beta
