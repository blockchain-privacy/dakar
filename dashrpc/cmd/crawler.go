package main

import (
	"dashrpc"
	"dashrpc/rpcclient"
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"
)

// "random" temporary badger DB location for benchmarking. Not used for anything else.
// should be deleted automatically
const badgerBenchmarkPath = "/tmp/tmpBadger_321"

//
// The main crawler for the system. It needs to be run prior to using any of the other
// commands that rely on the Badger DB to be pre-created.
//
// DashRPC client traverses the Dash blockchain and creates a Badger database entry for each transaction
// starting from a given block, and, working backwards, until a given stop block.
//
// Note: in the future, the crawler could be integrated with the backend-web service as
// to run continuously in the background and share the DB with other API queries.
//
func main() {
	fmt.Printf("Go DashRPC client  %s\nBlock crawler\n\n", dashrpc.VersionString)

	badgerDir := flag.String("db", "/tmp/badger", "badger database location")
	processContinue := flag.Bool("continue", false, "continue the previously started DB build process")
	rpcUser := flag.String("rpcuser", "rpc1user", "Dash RPC user")
	rpcPassword := flag.String("rpcpassword", "1234pass", "Dash RPC password")
	startBlockID := flag.String("start", "", "Start Block Id")
	stopBlockID := flag.String("stop", "", "Stop Block Id")
	startBlockHash := flag.String("hash", "", "Start Block Hash")
	isPrintStatus := flag.Bool("status", false, "prints current processing status")
	isBenchmark := flag.Bool("benchmark", false, "run short performance test")

	flag.Parse()
	if !*isPrintStatus &&
		!*processContinue &&
		!*isBenchmark && (*startBlockID == "" || *startBlockHash == "" || *stopBlockID == "") {
		flag.PrintDefaults()
		return
	}
	if *isBenchmark {
		fmt.Println("Benchmark is ON.")
		fmt.Println("Command line options -start -stop -hash -continue -path are ignored")
		fmt.Printf("\nIt takes about %v minutes to complete the benchmark" +
			" on a high-end laptop.\n\n", 2)
		*startBlockHash = "000000000000002ded278008e12198d0687682a299795bdbbcac8084d59cd607"
		*startBlockID = "901500"
		*stopBlockID = "901250"
		*processContinue = false
		*isPrintStatus = false
		*badgerDir = badgerBenchmarkPath
	}
	db := dashrpc.SetupBadgerDB(*badgerDir)
	defer func() {
		e := db.Close()
		if e != nil { /* ignore */
		}
	}()


	dbBlockCount := dashrpc.DbGetBlockCount(db)
	dbTxCount := dashrpc.DbGetGlobalTxCount(db)
	fmt.Printf("DB block count: %v  TX count: %v\n", dbBlockCount, dbTxCount)
	if *isPrintStatus {
		dashrpc.PrintStatus(db)
		return
	}

	var dbStatus string
	dashrpc.DbGetStatus(db, &dbStatus)
	fmt.Printf("DB status: %s\n", dbStatus)
	if dbStatus == dashrpc.DbBlockStatusFinished && *processContinue && *stopBlockID == ""{
		fmt.Println("\nError: when processing is finished to continue provide -stop option")
		return
	}
	if *processContinue && (*startBlockHash != "" || *startBlockID != "") {
		fmt.Println("\nError: cannot use -continue and start/stop options in the command line")
	}

	// Setup the RPC connection
	var conn = rpcclient.ConnConfig{
		Host:       "0.0.0.0:9998",
		User:       *rpcUser,
		Pass:       *rpcPassword,
		DisableTLS: true,
	}
	client, err := rpcclient.New(&conn)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	count, err := client.GetBlockCount()
	if err != nil {
		fmt.Printf("\nError: problem with count() %s\n", err.Error())
		return
	}
	fmt.Printf("Current block count in the chain: %v\n", count)

	if *processContinue {
		err = dashrpc.DbGetString(db, dashrpc.DbBlockLastBlockId, startBlockID)
		if err != nil {
			fmt.Printf("\nError: problem reading LastBlockID from DB: %s\n", err.Error())
			return
		}
		err = dashrpc.DbGetString(db, dashrpc.DbBlockLastBlockHash, startBlockHash)
		if err != nil {
			fmt.Printf("\nError: problem reading LastBlockHash from DB: %s\n", err.Error())
			return
		}
		err = dashrpc.DbGetString(db, dashrpc.DbBlockStopBlockId, stopBlockID)
		if err != nil {
			fmt.Printf("\nError: problem reading StopBlockID from DB: %s\n", err.Error())
			return
		}
	}

	if *startBlockID == "" || *startBlockHash == "" {
		fmt.Println("\nMissing block ID or block Hash. They need to match!")
		return
	}

	startingBlockId, err := strconv.ParseUint(*startBlockID, 10, 64)
	if err != nil {
		fmt.Printf("\nError %v\n", err.Error())
		return
	}
	startingBlockHash := *startBlockHash

	stopingBlockId, err := strconv.ParseUint(*stopBlockID, 10, 64)
	if err != nil {
		fmt.Printf("\nError %v\n", err.Error())
		return
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

	err = dashrpc.ProcessNewBlocks(db, client, startingBlockHash, startingBlockId, stopingBlockId)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	err = db.Close()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}
	if *isBenchmark {
		time.Sleep(time.Second * 5) // need to give time to Badger to shutdown
		err = os.RemoveAll(badgerBenchmarkPath)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}
}


