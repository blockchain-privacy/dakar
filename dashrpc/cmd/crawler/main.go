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
	cliArgs, err = cli.BuildArgs(cli.Continuous, cli.ResetDB, cli.RpcUser, cli.RpcPassword, cli.StartBlockID,
		cli.StopBlockID, cli.IsPrintStatus, cli.IsBenchmark, cli.RpcHost, cli.RpcPort, cli.Logfile)

	if err != nil {
		flag.PrintDefaults()
		return cliArgs, err
	}

	if !cliArgs.IsPrintStatus && !cliArgs.Continuous && !cliArgs.IsBenchmark &&
		(cliArgs.StartBlockID == 0 || cliArgs.StopBlockID == 0) {
		flag.PrintDefaults()
		err = errors.New("missing block information")
		return
	}

	if cliArgs.Continuous && (cliArgs.StartBlockID > 0 || cliArgs.StopBlockID > 0) {
		flag.PrintDefaults()
		err = errors.New("continuous syncing can not be used together with start and stop block id")
		return
	}

	// startBlockID must be smaller or equal than stopBlockID, as we go forward
	if cliArgs.StartBlockID > cliArgs.StopBlockID {
		flag.PrintDefaults()
		err = errors.New("start must be smaller or equal than stop")
		return
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
	//}

	return
}

// checks if a crawling process is already running
func isCrawling(dgraph *dgo.Dgraph) (bool, error) {
	dbStatus, err := status.Get(dgraph)
	if err != nil {
		// no status information found -> database is completely new
		// and thus no crawling is happening right now
		if err.Error() == status.ErrorStatusNotFound {
			return false, nil
		}

		return true, err
	} else if dbStatus.IsCrawling == nil {
		return true, errors.New("was not able to get crawling status successfully")
	}

	return *dbStatus.IsCrawling, nil
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

	if ok, err := isCrawling(dgraph); err != nil {
		log.Println(err)
		return
	} else if ok {
		log.Println("Crawling process is already running. Aborting ...")
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
	log.Printf("Current block count in the chain of the RPC client: %v\n", count)

	if err := status.SetCrawling(dgraph, true); err != nil {
		log.Println("could not set crawling status:", err)
		return
	}

	err = dashrpc.ProcessNewBlocks(dgraph, client, cliArgs.Continuous, cliArgs.StartBlockID, cliArgs.StopBlockID)
	if err != nil {
		log.Println(err)
	}

	if err = status.SetCrawling(dgraph, false); err != nil {
		log.Println(err)
		return
	}
}
