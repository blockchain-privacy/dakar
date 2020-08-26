package cliutil

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
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
	StartBlockID
	StopBlockID
	IsPrintStatus
	HttpServerPort
	StartHttpServer
	Logfile
	TxSearch
	TxInfo
	ClusterAddr
)

type Arguments struct {
	Continuous      bool
	IgnoreSafeguard bool
	ResetDB         bool
	RpcUser         string
	RpcPassword     string
	StartBlockID    uint64
	StopBlockID     uint64
	IsPrintStatus   bool
	RpcEndpoint     string
	Logfile         string
	TxSearch        string
	TxInfo          string
	ClusterAddr     string
	HttpServerPort  uint
	StartHttpServer bool
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

// creates a string in the format of "rpcHost:rpcPort"
func buildEndpoint(rpcHost string, rpcPort uint) (string, error) {
	// check if ip is valid
	if ip := net.ParseIP(rpcHost); ip == nil {
		return "", errors.New("IP is not valid")
	}

	// build endpoint string
	return rpcHost + ":" + strconv.Itoa(int(rpcPort)), nil
}

func GetLogfile(fileName string) (f *os.File, err error) {
	if len(fileName) > 0 {
		f, err = os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return
		}
		log.SetFlags(log.Lshortfile)
		log.SetOutput(io.MultiWriter(os.Stdout, f))
	}
	err = errors.New("name for log file is invalid")
	return
}

// parse arguments specified by the provided flags
func BuildArgs(flags ...Flag) (args Arguments, err error) {
	var rpcHostString string
	var rpcPortNumber uint

	isPortSet := false

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
			break
		case RpcUser:
			addRpcUser(&args.RpcUser)
			break
		case RpcPassword:
			addRpcPassword(&args.RpcPassword)
			break
		case RpcHost:
			addRpcHost(&rpcHostString)
			break
		case RpcPort:
			addRpcPort(&rpcPortNumber)
			isPortSet = true
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
		case StartHttpServer:
			addStartHttpServer(&args.StartHttpServer)
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
		default:
			err = errors.New("flag not recognized")
			return args, err
		}
	}

	flag.Parse()

	// if host and port are not empty build the endpoint
	if len(rpcHostString) > 0 && isPortSet {
		endpoint, err := buildEndpoint(rpcHostString, rpcPortNumber)
		if err != nil {
			return args, err
		}

		args.RpcEndpoint = endpoint
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

func addLogfile(v *string) {
	flag.StringVar(v, "logfile", "", "Specify log file (default: none)")
}

func addHttpServerPort(v *uint) {
	flag.UintVar(v, "serverport", 8081, "Http server port (default: 8081)")
}

func addStartHttpServer(v *bool) {
	flag.BoolVar(v, "startserver", false, "Start the http server (default: false)")
}
