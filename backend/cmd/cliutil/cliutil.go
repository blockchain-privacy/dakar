package cliutil

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"runtime"
	"strconv"
	"strings"
)

type Flag int

// flag enum
const (
	Continuous Flag = iota
	IgnoreSafeguard
	ResetDB
	RpcUser
	RpcPassword
	RpcHost
	RpcPort
	DBHost
	DBPort
	StartBlockID
	StopBlockID
	IsPrintStatus
	HttpServerPort
	DisableHttpServer
	DisableCrawler
	DisableAnalyzer
	Logfile
	TxSearch
	TxInfo
	ClusterAddr
	BTC
)

type Arguments struct {
	Continuous        bool
	IgnoreSafeguard   bool
	ResetDB           bool
	RpcUser           string
	RpcPassword       string
	StartBlockID      uint64
	StopBlockID       uint64
	IsPrintStatus     bool
	RpcEndpoint       string
	DBEndpoint        string
	Logfile           string
	TxSearch          string
	TxInfo            string
	ClusterAddr       string
	HttpServerPort    uint
	DisableHttpServer bool
	DisableCrawler    bool
	DisableAnalyzer   bool
	BTC               bool
}

func ShowCallInfo() string {
	pc, file, line, ok := runtime.Caller(1)
	if !ok {
		log.Fatal("not ok")
	}

	_, fileName := path.Split(file)
	parts := strings.Split(runtime.FuncForPC(pc).Name(), ".")
	pl := len(parts)
	funcName := parts[pl-1]

	if parts[pl-2][0] == '(' {
		funcName = parts[pl-2] + "." + funcName
	}

	return fmt.Sprintf("%s:%d %s", fileName, line, funcName)
}

// creates a string in the format of "host:port"
func buildEndpoint(host string, port uint) (string, error) {
	// the host can be in a form of IP address, or in a form of Label (e.g. in Docker), or proper hostname
	// it is complicated to actually validate it properly
	//
	// check if ip is valid
	// if ip := net.ParseIP(rpcHost); ip == nil {
	//	return "", errors.New("IP is not valid")
	// }

	host = strings.TrimSpace(host)
	response := host + ":" + strconv.Itoa(int(port))
	if len(host) > 0 && port > 0 {
		return response, nil
	}

	return response, errors.New("host or port is not valid")
}

func GetLogfile(fileName string) (f *os.File, err error) {
	if len(fileName) > 0 {
		f, err = os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return
		}
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		log.SetOutput(io.MultiWriter(os.Stdout, f))
	}
	err = errors.New("name for log file is invalid")
	return
}

// parse arguments specified by the provided flags
func BuildArgs(flags ...Flag) (args Arguments, err error) {
	var rpcHostString string
	var rpcPortNumber uint
	var dbHostString string
	var dbPortNumber uint

	var isRPCused bool
	var isDBused bool

	for _, f := range flags {
		switch f {
		case Continuous:
			addContinuous(&args.Continuous)
			break
		case IgnoreSafeguard:
			addIgnoreSafeguard(&args.IgnoreSafeguard)
			break
		case ResetDB:
			addResetDB(&args.ResetDB)
			isRPCused = true
			break
		case RpcUser:
			addRpcUser(&args.RpcUser)
			isRPCused = true
			break
		case RpcPassword:
			addRpcPassword(&args.RpcPassword)
			isRPCused = true
			break
		case RpcHost:
			addRpcHost(&rpcHostString)
			break
		case RpcPort:
			addRpcPort(&rpcPortNumber)
			break
		case DBHost:
			addDBHost(&dbHostString)
			isDBused = true
			break
		case DBPort:
			addDBPort(&dbPortNumber)
			isDBused = true
			break
		case StartBlockID:
			addStartBlockID(&args.StartBlockID)
			break
		case StopBlockID:
			addStopBlockID(&args.StopBlockID)
			break
		case IsPrintStatus:
			addIsPrintStatus(&args.IsPrintStatus)
			break
		case Logfile:
			addLogfile(&args.Logfile)
			break
		case HttpServerPort:
			addHttpServerPort(&args.HttpServerPort)
			break
		case DisableHttpServer:
			addDisableHttpServer(&args.DisableHttpServer)
			break
		case DisableCrawler:
			addDisableCrawler(&args.DisableCrawler)
			break
		case DisableAnalyzer:
			addDisableAnalyzer(&args.DisableAnalyzer)
			break
		case TxSearch:
			addTxSearch(&args.TxSearch)
			break
		case TxInfo:
			addTxInfo(&args.TxInfo)
			break
		case ClusterAddr:
			addClusterAddr(&args.ClusterAddr)
			break
		case BTC:
			addBTC(&args.BTC)
			break
		default:
			err = errors.New("flag not recognized")
			return args, err
		}
	}

	flag.Parse()

	if isRPCused {
		args.RpcEndpoint, err = buildEndpoint(rpcHostString, rpcPortNumber)
		if err != nil {
			return args, err
		}
	}

	if isDBused {
		args.DBEndpoint, err = buildEndpoint(dbHostString, dbPortNumber)
	}

	return args, err
}

func addClusterAddr(v *string) {
	flag.StringVar(v, "clusteraddr", "", "Create cluster for the given address (default: none)")
}

func addTxInfo(v *string) {
	flag.StringVar(v, "txinfo", "", "Get information about the given transaction hash (default: none)")
}

func addTxSearch(v *string) {
	flag.StringVar(v, "txsearch", "", "Last PrivateSend transaction hash (default: none)")
}

func addContinuous(v *bool) {
	flag.BoolVar(v, "continuous", false, "Continuously syncs the whole chain (default: false)")
}

func addIgnoreSafeguard(v *bool) {
	flag.BoolVar(v, "ignoresafeguard", false, "Ignore the crawling safe guard (default: false)")
}

func addResetDB(v *bool) {
	flag.BoolVar(v, "reset", false, "Remove all data from the database (default: false)")
}

func addRpcUser(v *string) {
	flag.StringVar(v, "rpcuser", "rpc1user", "Dash RPC user (default: rpc1user)")
}

func addRpcPassword(v *string) {
	flag.StringVar(v, "rpcpassword", "1234pass", "Dash RPC password (default: 1234pass)")
}

func addStartBlockID(v *uint64) {
	flag.Uint64Var(v, "start", 0, "Start Block Id")
}

func addStopBlockID(v *uint64) {
	flag.Uint64Var(v, "stop", 0, "Stop Block Id")
}

func addIsPrintStatus(v *bool) {
	flag.BoolVar(v, "status", false, "Prints current processing status (default: false)")
}

func addRpcHost(v *string) {
	flag.StringVar(v, "rpchost", "0.0.0.0", "Dash RPC host IP (default: 0.0.0.0)")
}

func addRpcPort(v *uint) {
	flag.UintVar(v, "rpcport", 9998, "Dash RPC port (default: 9998)")
}

func addDBHost(v *string) {
	flag.StringVar(v, "dbhost", "0.0.0.0", "Dgraph host IP (default: 0.0.0.0)")
}

func addDBPort(v *uint) {
	flag.UintVar(v, "dbport", 9080, "Dgraph port (default: 9080)")
}

func addLogfile(v *string) {
	flag.StringVar(v, "logfile", "", "Specify log file (default: none)")
}

func addHttpServerPort(v *uint) {
	flag.UintVar(v, "serverport", 8081, "Http server port (default: 8081)")
}

func addDisableHttpServer(v *bool) {
	flag.BoolVar(v, "disableserver", false, "Disable the http server (default: false)")
}

func addDisableCrawler(v *bool) {
	flag.BoolVar(v, "disablecrawler", false, "Disable the crawler (default: false)")
}

func addDisableAnalyzer(v *bool) {
	flag.BoolVar(v, "disableanalyzer", false, "Disable the analyzer (default: false)")
}

func addBTC(v *bool) {
	flag.BoolVar(v, "btc", false, "Switch to Bitcoin mode (default: false)")
}
