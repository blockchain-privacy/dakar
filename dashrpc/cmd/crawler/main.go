package main

import (
	"dashrpc"
	cli "dashrpc/cmd/cliutil"
	"dashrpc/db"
	"dashrpc/db/status"
	"dashrpc/rpcclient"
	"errors"
	"flag"
	"fmt"
	"github.com/dgraph-io/dgo/v2"
	"log"
)

const benchmarkStartBlockID = 901500
const benchmarkStopBlockID = 901250

func getCLIArgs() (cliArgs cli.Arguments, err error) {
	cliArgs, err = cli.BuildArgs(cli.ProcessContinue, cli.ResetDB, cli.RpcUser, cli.RpcPassword, cli.StartBlockID,
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

// checks if we can continue crawling with the provided arguments
func canContinue(dgraph *dgo.Dgraph, startBlock uint64, stopBlock uint64) error {
	if startBlock != 0 {
		return errors.New("cannot use -continue and start/stop options in the command line")
	}

	dbStatus, err := status.Get(dgraph)
	if err != nil {
		return err
	} else if dbStatus.IsCrawling == nil {
		return errors.New("was not able to status successfully")
	}

	if !*dbStatus.IsCrawling && stopBlock == 0 {
		return errors.New("when processing is finished, provide -stop option to continue provide")
	}

	return nil
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
	if f, err := cli.GetLogfile(cliArgs.Logfile, "crawler"); err != nil {
		defer func() {
			if err = f.Close(); err != nil {
				fmt.Println(err)
			}
		}()
	}

	if cliArgs.IsBenchmark {
		log.Print("Benchmarking is currently not supported")
		return
	}

	// create dgraph client
	dgraph, c, err := db.CreateDefaultClient()
	if err != nil {
		log.Print(err)
		return
	}
	defer func() {
		if err = c.Close(); err != nil {
			log.Println(err)
		}
	}()

	if cliArgs.IsPrintStatus {
		status.PrintStatus(dgraph)
		return
	}

	if cliArgs.ResetDB {
		err = db.DropAll(dgraph)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("dropped all data")
		err = db.SetupSchema(dgraph)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("setup new schema")
	}

	if cliArgs.ProcessContinue {
		if err = canContinue(dgraph, cliArgs.StartBlockID, cliArgs.StopBlockID); err != nil {
			log.Println(err)
			return
		}
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

	err = dashrpc.ProcessNewBlocks(dgraph, client, !cliArgs.ExcludeAddresses, cliArgs.StartBlockID, cliArgs.StopBlockID)
	if err != nil {
		log.Println(err)
		return
	}
}
