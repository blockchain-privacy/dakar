package cliutil

import (
	"errors"
	"flag"
	"net"
	"strconv"
)

type Flag int

// flag enum
const (
	BadgerDirectory Flag = iota
	ProcessContinue
	RpcUser
	RpcPassword
	RpcHost
	RpcPort
	StartBlockID
	StopBlockID
	StartBlockHash
	IsPrintStatus
	IsBenchmark
	ExcludeAddresses
	ExplorerServerPort
	Logfile
)

type Arguments struct {
	BadgerDir          string
	ProcessContinue    bool
	RpcUser            string
	RpcPassword        string
	StartBlockID       uint64
	StopBlockID        uint64
	StartBlockHash     string
	IsPrintStatus      bool
	IsBenchmark        bool
	ExcludeAddresses   bool
	RpcEndpoint        string
	Logfile            string
	ExplorerServerPort uint
}

func buildEndpoint(rpcHost string, rpcPort uint) (string, error) {
	// check if ip is valid
	if ip := net.ParseIP(rpcHost); ip == nil {
		return "", errors.New("IP is not valid")
	}

	// build endpoint string
	return rpcHost + ":" + strconv.Itoa(int(rpcPort)), nil
}

// parse arguments specified by the provided flags
func BuildArgs(flags ...Flag) (args Arguments, err error) {
	var rpcHostString string
	var rpcPortNumber uint

	badgerDirRequested := false
	isPortSet := false

	for _, f := range flags {
		switch f {
		case BadgerDirectory:
			addBadgerDir(&args.BadgerDir)
			badgerDirRequested = true
			break
		case ProcessContinue:
			addProcessContinue(&args.ProcessContinue)
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
		case StartBlockHash:
			addStartBlockHash(&args.StartBlockHash)
			break
		case IsPrintStatus:
			addIsPrintStatus(&args.IsPrintStatus)
			break
		case IsBenchmark:
			addIsBenchmark(&args.IsBenchmark)
			break
		case ExcludeAddresses:
			addExcludeAddresses(&args.ExcludeAddresses)
			break
		case Logfile:
			addLogfile(&args.Logfile)
			break
		case ExplorerServerPort:
			addExplorerServerPort(&args.ExplorerServerPort)
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

	// check if database directory is empty
	if badgerDirRequested && len(args.BadgerDir) == 0 {
		err = errors.New("badger dir is empty")
		return args, err
	}

	return args, err
}

func addBadgerDir(v *string) {
	flag.StringVar(v, "db", "/tmp/badger", "Badger database location (default: /tmp/badger)")
}

func addProcessContinue(v *bool) {
	flag.BoolVar(v, "continue", false, "Continue the previously started DB build process")
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

func addStartBlockHash(v *string) {
	flag.StringVar(v, "hash", "", "Start Block Hash")
}

func addIsPrintStatus(v *bool) {
	flag.BoolVar(v, "status", false, "Prints current processing status (default: false)")
}

func addIsBenchmark(v *bool) {
	flag.BoolVar(v, "benchmark", false, "Run short performance test (default: false)")
}

func addExcludeAddresses(v *bool) {
	flag.BoolVar(v, "excludeaddresses", false, "Exclude addresses from saving into the database (default: false)")
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

func addExplorerServerPort(v *uint) {
	flag.UintVar(v, "serverport", 8081, "Explorer server port (default: 8081)")
}
