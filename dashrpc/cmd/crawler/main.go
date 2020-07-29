package main

import (
	"dashrpc"
	cli "dashrpc/cmd/cliutil"
	"dashrpc/db"
	"dashrpc/rpcclient"
	"errors"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
)

const benchmarkStartBlockID = 901500
const benchmarkStopBlockID = 901250

func getCLIArgs() (cliArgs cli.Arguments, err error) {
	cliArgs, err = cli.BuildArgs(cli.ProcessContinue, cli.RpcUser, cli.RpcPassword, cli.StartBlockID,
		cli.StopBlockID, cli.IsPrintStatus, cli.IsBenchmark, cli.ExcludeAddresses, cli.RpcHost, cli.RpcPort, cli.Logfile)

	if err != nil {
		flag.PrintDefaults()
		return cliArgs, err
	}

	if !cliArgs.IsPrintStatus && !cliArgs.ProcessContinue && !cliArgs.IsBenchmark &&
		(cliArgs.StartBlockID == 0 || cliArgs.StopBlockID == 0) {
		flag.PrintDefaults()
		err = errors.New("missing block information")
		return cliArgs, err
	}

	// startBlockID must be smaller than stopBlockID, as we go forward
	if cliArgs.StartBlockID > cliArgs.StopBlockID {
		flag.PrintDefaults()
		err = errors.New("start must be smaller than stop")
		return cliArgs, err
	}

	// todo: do we still need benchmarks?
	//if cliArgs.IsBenchmark {
	//	cliArgs.StartBlockID = benchmarkStartBlockID
	//	cliArgs.StopBlockID = benchmarkStopBlockID
	//	cliArgs.ProcessContinue = false
	//	cliArgs.IsPrintStatus = false
	//
	//	// temp dir will be deleted later on
	//	dirName, err := ioutil.TempDir("", "dashrpc")
	//
	//	if err != nil {
	//		flag.PrintDefaults()
	//		return cliArgs, err
	//	}
	//	cliArgs.BadgerDir = dirName
	//}

	return cliArgs, err
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
	cliArgs, err := getCLIArgs()
	if err != nil {
		fmt.Println(err)
		return
	}

	// setup Logging
	if len(cliArgs.Logfile) > 0 {
		f, err := os.OpenFile(cliArgs.Logfile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
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

	if cliArgs.IsBenchmark {
		log.Print("Benchmarking is currently not supported")
		return
	}

	// create dgraph client
	dbClient, c, err := db.CreateDefaultClient()
	if err != nil {
		log.Print(err)
		return
	}
	defer func() {
		if err = c.Close(); err != nil {
			log.Println(err)
		}
	}()

	// drop all data todo: remove
	err = db.DropAll(dbClient)
	if err != nil {
		log.Print(err)
		return
	}

	// create new db schema todo: remove and only do if requested via CLI argument
	err = db.SetupSchema(dbClient)
	if err != nil {
		log.Print(err)
		return
	}

	if cliArgs.IsPrintStatus {
		//dashrpc.PrintStatus(db)
		return
	}

	//if dashrpc.DbGetStatus(db) == dashrpc.DbBlockStatusFinished && cliArgs.ProcessContinue && cliArgs.StopBlockID == 0 {
	//	log.Println("\nError: when processing is finished, provide -stop option to continue provide")
	//	return
	//}

	if cliArgs.ProcessContinue && cliArgs.StartBlockID != 0 {
		log.Println("\nError: cannot use -continue and start/stop options in the command line")
		return
	}

	// Setup the RPC connection
	client, err := rpcclient.New(&rpcclient.ConnConfig{
		Host:       cliArgs.RpcEndpoint,
		User:       cliArgs.RpcUser,
		Pass:       cliArgs.RpcPassword,
		DisableTLS: true,
	})
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

	//if cliArgs.ProcessContinue {
	//	err = dashrpc.DbGetUint64(db, dashrpc.DbBlockLastBlockId, &cliArgs.StartBlockID)
	//	if err != nil {
	//		log.Printf("\nError: problem reading LastBlockID from DB: %s\n", err.Error())
	//		return
	//	}
	//	err = dashrpc.DbGetString(db, dashrpc.DbBlockLastBlockHash, &cliArgs.StartBlockHash)
	//	if err != nil {
	//		log.Printf("\nError: problem reading LastBlockHash from DB: %s\n", err.Error())
	//		return
	//	}
	//	err = dashrpc.DbGetUint64(db, dashrpc.DbBlockStopBlockId, &cliArgs.StopBlockID)
	//	if err != nil {
	//		log.Printf("\nError: problem reading StopBlockID from DB: %s\n", err.Error())
	//		return
	//	}
	//}

	err = dashrpc.ProcessNewBlocks(dbClient, client, !cliArgs.ExcludeAddresses, cliArgs.StartBlockID, cliArgs.StopBlockID)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}
}
