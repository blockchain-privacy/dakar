package main

import (
	"dashrpc"
	"dashrpc/rpcclient"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"time"
)

const benchmarkStartBlockID = 901500
const benchmarkStopBlockID = 901250
const benchmarkStartBlockHash = "000000000000002ded278008e12198d0687682a299795bdbbcac8084d59cd607"

type CLIArguments struct {
	badgerDir       string
	processContinue bool
	rpcUser         string
	rpcPassword     string
	startBlockID    uint64
	stopBlockID     uint64
	startBlockHash  string
	isPrintStatus   bool
	isBenchmark     bool
	saveAddresses   bool
	rpcEndpoint     string
	logfile         string
	err             error
}

func buildEndpoint(rpcHost string, rpcPort uint) (string, error) {
	// check if ip is valid
	if ip := net.ParseIP(rpcHost); ip == nil {
		return "", errors.New("IP is not valid")
	}

	// build endpoint string
	return rpcHost + ":" + strconv.Itoa(int(rpcPort)), nil
}

// saves cli arguments in cli structure
func getCLIArgs() (cliArgs CLIArguments) {
	badgerDir := flag.String("db", "/tmp/badger", "Badger database location")
	processContinue := flag.Bool("continue", false, "Continue the previously started DB build process")
	rpcUser := flag.String("rpcuser", "rpc1user", "Dash RPC user")
	rpcPassword := flag.String("rpcpassword", "1234pass", "Dash RPC password")
	startBlockID := flag.Uint64("start", 0, "Start Block Id")
	stopBlockID := flag.Uint64("stop", 0, "Stop Block Id")
	startBlockHash := flag.String("hash", "", "Start Block Hash")
	isPrintStatus := flag.Bool("status", false, "Prints current processing status (default: false)")
	isBenchmark := flag.Bool("benchmark", false, "Run short performance test (default: false)")
	saveAddresses := flag.Bool("addresses", false, "Save addresses into database (default: false)")
	rpcHost := flag.String("rpchost", "0.0.0.0", "Dash RPC host IP (default: 0.0.0.0)")
	rpcPort := flag.Uint("rpcport", 9998, "Dash RPC port (default: 9998)")
	logfile := flag.String("logfile", "", "Specify log file (default: crawler.log)")
	flag.Parse()

	cliArgs.badgerDir = *badgerDir
	cliArgs.processContinue = *processContinue
	cliArgs.rpcUser = *rpcUser
	cliArgs.rpcPassword = *rpcPassword
	cliArgs.startBlockID = *startBlockID
	cliArgs.stopBlockID = *stopBlockID
	cliArgs.startBlockHash = *startBlockHash
	cliArgs.isPrintStatus = *isPrintStatus
	cliArgs.isBenchmark = *isBenchmark
	cliArgs.saveAddresses = *saveAddresses
	cliArgs.logfile = *logfile

	ep, err := buildEndpoint(*rpcHost, *rpcPort)

	if err != nil {
		flag.PrintDefaults()
		cliArgs.err = err
		return cliArgs
	}

	cliArgs.rpcEndpoint = ep

	if !*isPrintStatus && !*processContinue && !*isBenchmark && (*startBlockID == 0 || *startBlockHash == "" || *stopBlockID == 0) {
		flag.PrintDefaults()
		cliArgs.err = errors.New("missing block information")
		return cliArgs
	}

	// startBlockID must be bigger than stopBlockID, as we go backwards
	if *startBlockID < *stopBlockID {
		flag.PrintDefaults()
		cliArgs.err = errors.New("start must be bigger than stop")
		return cliArgs
	}

	if *isBenchmark {
		cliArgs.startBlockHash = benchmarkStartBlockHash
		cliArgs.startBlockID = benchmarkStartBlockID
		cliArgs.stopBlockID = benchmarkStopBlockID
		cliArgs.processContinue = false
		cliArgs.isPrintStatus = false

		// temp dir will be deleted later on
		dirName, err := ioutil.TempDir("", "dashrpc")

		if err != nil {
			flag.PrintDefaults()
			cliArgs.err = err
			return cliArgs
		}
		cliArgs.badgerDir = dirName
	}

	return cliArgs
}

// The main crawler for the system. It needs to be run prior to using any of the other
// commands that rely on the Badger DB to be pre-created.
//
// DashRPC client traverses the Dash blockchain and creates a Badger database entry for each transaction
// starting from a given block, and, working backwards, until a given stop block.
//
// Note: in the future, the crawler could be integrated with the backend-web service as
// to run continuously in the background and share the DB with other API queries.
func main() {
	fmt.Printf("Go DashRPC client  %s\nBlock crawler\n\n", dashrpc.VersionString)
	cliArgs := getCLIArgs()
	if cliArgs.err != nil {
		fmt.Println(cliArgs.err)
		return
	}

	// setup Logging
	if len(cliArgs.logfile) > 0 {
		f, err := os.OpenFile(cliArgs.logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			fmt.Println("Error opening log file", err)
			return
		}
		defer func() {
			err = f.Close()
			if err != nil {
				fmt.Println(err)
			}
		}()
		log.SetPrefix("crawler ")
		log.SetOutput(io.MultiWriter(os.Stdout, f))
	}

	if cliArgs.isBenchmark {
		benchmarkStr := "Benchmark is ON."
		if cliArgs.saveAddresses {
			benchmarkStr = "Benchmark with addresses is ON"
		}
		log.Println(benchmarkStr)
		log.Println("Command line options -start -stop -hash -continue -path are ignored")
		log.Printf("It takes about %v minutes to complete the benchmark"+
			" on a high-end laptop.", 2)

		// remove temp dir at the end
		defer func() {
			err := os.RemoveAll(cliArgs.badgerDir)
			if err != nil {
				log.Printf("Error: %v\n", err)
			}
		}()
	}

	db := dashrpc.SetupBadgerDB(cliArgs.badgerDir)
	defer func() {
		e := db.Close()
		if e != nil { /* ignore */
		}
	}()

	dbBlockCount := dashrpc.DbGetBlockCount(db)
	dbTxCount := dashrpc.DbGetGlobalTxCount(db)
	log.Printf("DB block count: %v  TX count: %v\n", dbBlockCount, dbTxCount)
	if cliArgs.isPrintStatus {
		dashrpc.PrintStatus(db)
		return
	}

	var dbStatus string
	dashrpc.DbGetStatus(db, &dbStatus)
	log.Printf("DB status: %s\n", dbStatus)

	if dbStatus == dashrpc.DbBlockStatusFinished && cliArgs.processContinue && cliArgs.stopBlockID == 0 {
		log.Println("\nError: when processing is finished to continue provide -stop option")
		return
	}

	if cliArgs.processContinue && (cliArgs.startBlockHash != "" || cliArgs.startBlockID != 0) {
		log.Println("\nError: cannot use -continue and start/stop options in the command line")
		return
	}

	// Setup the RPC connection
	var conn = rpcclient.ConnConfig{
		Host:       cliArgs.rpcEndpoint,
		User:       cliArgs.rpcUser,
		Pass:       cliArgs.rpcPassword,
		DisableTLS: true,
	}

	client, err := rpcclient.New(&conn)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	count, err := client.GetBlockCount()
	if err != nil {
		log.Printf("\nError: problem with count() %s\n", err.Error())
		return
	}
	log.Printf("Current block count in the chain: %v\n", count)

	if cliArgs.processContinue {
		err = dashrpc.DbGetUint64(db, dashrpc.DbBlockLastBlockId, &cliArgs.startBlockID)
		if err != nil {
			log.Printf("\nError: problem reading LastBlockID from DB: %s\n", err.Error())
			return
		}
		err = dashrpc.DbGetString(db, dashrpc.DbBlockLastBlockHash, &cliArgs.startBlockHash)
		if err != nil {
			log.Printf("\nError: problem reading LastBlockHash from DB: %s\n", err.Error())
			return
		}
		err = dashrpc.DbGetUint64(db, dashrpc.DbBlockStopBlockId, &cliArgs.stopBlockID)
		if err != nil {
			log.Printf("\nError: problem reading StopBlockID from DB: %s\n", err.Error())
			return
		}
	}

	//startingBlockId := uint64(1060000)
	//startingBlockHash := "00000000000000132447e6bac9fe0d7d756851450eab29358787dc05d809bf07"

	// 2019-05-05 19:22
	// Block: 1065229
	// 0000000000000015b42d1e661ccffac1128a0fde14ae6ec5ed78f7b16a04820c
	//
	// startingBlockId := 1065229
	// startingBlockHash := "0000000000000015b42d1e661ccffac1128a0fde14ae6ec5ed78f7b16a04820c"

	//
	// Appeared in Dash 126744 (2014-08-28 19:47:52)
	// startingBlockHash := "00000000000d0b8cd2507d6ea244bc7109ff9c979a8653617caaff6df848452d"

	// startingBlockId := 50000
	// startingBlockHash := "00000000000fa6230896498b3cc6f1015456b4512452ead9979f6b43ca0a74dc"

	// 50 block
	// startingBlockHash := "00000f106b17cfec9d127b0cab42fd5b8c4102b39800be0e711b4cb38c017e7a"

	// 100 block
	// startingBlockHash := "00000fcef4b9e3b5aa2371dc7f310a8cc2e27171121d656e77f59464e7c0d400"

	err = dashrpc.ProcessNewBlocks(db, client, cliArgs.saveAddresses, cliArgs.startBlockHash, cliArgs.startBlockID, cliArgs.stopBlockID)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	err = db.Close()
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	if cliArgs.isBenchmark {
		time.Sleep(time.Second * 5) // need to give time to Badger to shutdown
	}
}
