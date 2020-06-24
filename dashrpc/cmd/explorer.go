package main

import (
	"dashrpc"
	cli "dashrpc/cmd/cliutil"
	"dashrpc/rpcclient"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/dgraph-io/badger/v2"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
)

const (
	routePrefix      string = "/api/v1/"
	routeTransaction string = "tx/"
	routeBlock       string = "blk/"
	routeAddress     string = "address/"
	routeMeta        string = "meta/"
	routeRoot        string = ""
)

func getRoute(r string) string {
	return routePrefix + r
}

func getRouteTransaction() string {
	return getRoute(routeTransaction)
}

func getRouteBlock() string {
	return getRoute(routeBlock)
}

func getRouteAddress() string {
	return getRoute(routeAddress)
}

func getRouteRoot() string {
	return getRoute(routeRoot)
}

func getRouteMeta() string {
	return getRoute(routeMeta)
}

// Block represents a simple block
type meta struct {
	LastBlockId      uint64 `json:"lastblockid"`
	StopBlockId      uint64 `json:"stopblockid"`
	LastBlockHash    string `json:"lastblockhash"`
	Status           string `json:"status"`
	RangeUp          uint64 `json:"rangeup"`
	RangeDown        uint64 `json:"rangedown"`
	GlobalBlockCount uint64 `json:"globalblockcount"`
	GlobalTxCount    uint64 `json:"globaltxcount"`
}

func (m meta) String() string {
	return fmt.Sprintf("LastBlockId: %s\nStopBlockId:\n%LastBlockHash:\n%v\n",
		m.LastBlockId, m.StopBlockId, m.LastBlockHash)
}

func setDefaultHeader(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8080")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "X-Requested-With, Content-Type, Authorization, Origin, Accept")
}

// API pattern: "/api/v1/"
// OUTPUT: List of patterns
func handlerRoot(w http.ResponseWriter, r *http.Request) {
	log.Println("Accessed", r.URL.Path)
	setDefaultHeader(w)
	// not handling possible errors
	_, e := fmt.Fprintln(w, "Possible routes:")
	if e != nil {
		log.Println(e)
	}

	_, e = fmt.Fprintln(w, "/\t\t-> This page")
	if e != nil {
		log.Println(e)
	}

	_, e = fmt.Fprintln(w, "/tx/<hash>\t-> Transaction details")
	if e != nil {
		log.Println(e)
	}

	_, e = fmt.Fprintln(w, "/address/<hash>\t-> Address details")
	if e != nil {
		log.Println(e)
	}

	_, e = fmt.Fprintln(w, "/blk/<hash>\t-> Block details")
	if e != nil {
		log.Println(e)
	}
	_, e = fmt.Fprintln(w, "/meta/\t\t-> Database meta information")
	if e != nil {
		log.Println(e)
	}
}

// API pattern: "/api/v1/blk/<hash>"
// OUTPUT: dashrpc.BlkDetails
func handlerBlockDetails(db *badger.DB, client *rpcclient.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Accessed", r.URL.Path)
		setDefaultHeader(w)

		blkHashString := r.URL.Path[len(getRouteBlock()):]
		block := dashrpc.Block{}
		err := dashrpc.DbGetBlock(db, blkHashString, &block)
		if err != nil {
			http.Error(w, err.Error()+" Block hash: "+blkHashString, http.StatusNotFound)
			return
		}

		blkDetails := dashrpc.BlkDetails{}

		blkDetails.Hash = block.Hash.String()
		blkDetails.Id = block.Id
		blkDetails.NextBlockHash = block.NextBlockHash.String()
		blkDetails.PrevBlockHash = block.PrevBlockHash.String()
		blkDetails.TxHashes = block.TxHashes
		blkDetails.Timestamp = block.Timestamp

		err = json.NewEncoder(w).Encode(blkDetails)
		if err != nil {
			http.Error(w, err.Error()+" Block: "+blkDetails.String(), http.StatusInternalServerError)
		}
	}
}

// API pattern: "/api/v1/address/<hash>"
// OUTPUT: dashrpc.AddressData
func handlerAddressDetails(db *badger.DB, client *rpcclient.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Accessed", r.URL.Path)
		setDefaultHeader(w)

		addressHashString := r.URL.Path[len(getRouteAddress()):]
		addressData := dashrpc.AddressData{}
		err := dashrpc.DbGetDataForAddress(db, addressHashString, &addressData)
		if err != nil {
			http.Error(w, err.Error()+" Key: "+addressHashString, http.StatusNotFound)
			return
		}

		err = json.NewEncoder(w).Encode(addressData)
		if err != nil {
			http.Error(w, err.Error()+" AddressData: "+addressData.Address, http.StatusInternalServerError)
		}
	}
}

// API pattern: "/api/v1/tx/<hash>"
// OUTPUT: dashrpc.Transaction
func handlerTxDetails(db *badger.DB, client *rpcclient.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Accessed", r.URL.Path)
		setDefaultHeader(w)

		txHashString := r.URL.Path[len(getRouteTransaction()):]
		log.Println(txHashString)
		txDetails := dashrpc.TxDetails{}
		err := dashrpc.DbGetTxDetails(db, txHashString, &txDetails)
		if err != nil {
			http.Error(w, err.Error()+" Key: "+txHashString, http.StatusNotFound)
			return
		}
		txHash, err := chainhash.NewHashFromStr(txHashString)
		if err != nil {
			http.Error(w, err.Error()+" Key: "+txHashString, http.StatusNotFound)
			return
		}

		tx, err := client.GetRawTransactionVerbose(txHash)
		if err != nil {
			http.Error(w, err.Error()+" Key: "+txHashString, http.StatusNotFound)
			return
		}
		block := dashrpc.Block{}
		err = dashrpc.DbGetBlock(db, tx.BlockHash, &block)
		if err != nil {
			http.Error(w, err.Error()+" Block hash: "+tx.BlockHash, http.StatusNotFound)
			return
		}
		transaction := dashrpc.Transaction{}
		transaction.Bhash = tx.BlockHash
		transaction.Bheight = block.Id
		transaction.Bts = block.Timestamp.Unix()
		transaction.Confirmations = tx.Confirmations
		transaction.Version = tx.Version
		transaction.Tx = txDetails
		// common['Access-Control-Request-Method'] = '*'")
		err = json.NewEncoder(w).Encode(transaction)
		if err != nil {
			http.Error(w, err.Error()+" TxDetails: "+txDetails.String(), http.StatusInternalServerError)
		}
	}
}

// API pattern: "/api/v1/meta/"
// OUTPUT: dashrpc.Transaction
func handlerMeta(db *badger.DB, client *rpcclient.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Accessed", r.URL.Path)
		setDefaultHeader(w)

		metaInformation := meta{
			LastBlockId:      dashrpc.DbGetLastBlockId(db),
			StopBlockId:      dashrpc.DbGetStopBlockId(db),
			LastBlockHash:    dashrpc.DbGetLastBlockHash(db),
			Status:           dashrpc.DbGetStatus(db),
			RangeUp:          dashrpc.DbGetRangeUp(db),
			RangeDown:        dashrpc.DbGetRangeDown(db),
			GlobalBlockCount: dashrpc.DbGetBlockCount(db),
			GlobalTxCount:    dashrpc.DbGetGlobalTxCount(db),
		}

		err := json.NewEncoder(w).Encode(metaInformation)
		if err != nil {
			http.Error(w, err.Error()+" Meta information: "+metaInformation.String(), http.StatusInternalServerError)
		}
	}
}

// setup cli
func getExplorerCLIArgs() (cliArgs cli.Arguments, err error) {
	cliArgs, err = cli.BuildArgs(cli.BadgerDirectory, cli.RpcUser, cli.RpcPassword, cli.RpcHost,
		cli.RpcPort, cli.Logfile, cli.IsPrintStatus, cli.ExplorerServerPort)

	if err != nil {
		flag.PrintDefaults()
		return cliArgs, err
	}

	return cliArgs, err
}

// Simple web-based utility to browse/lookup the TXs from the badger database
// It provides the API through HTTP
// Work in Progress. NOT WORKING YET.
func main() {
	fmt.Printf("Go DashRPC client  %s\nBlock explorer\n\n", dashrpc.VersionString)

	cliArgs, err := getExplorerCLIArgs()
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
		log.SetPrefix("explorer ")
		log.SetOutput(io.MultiWriter(os.Stdout, f))
	}

	db := dashrpc.SetupBadgerDB(cliArgs.BadgerDir)
	defer func() {
		e := db.Close()
		if e != nil { /* ignore */
		}
	}()

	// Setup the RPC connection
	var conn = rpcclient.ConnConfig{
		Host:       cliArgs.RpcEndpoint,
		User:       cliArgs.RpcUser,
		Pass:       cliArgs.RpcPassword,
		DisableTLS: true,
	}
	client, err := rpcclient.New(&conn)
	if err != nil {
		log.Printf("Error: %v\n", err)
		return
	}

	dbBlockCount := dashrpc.DbGetBlockCount(db)
	dbTxCount := dashrpc.DbGetGlobalTxCount(db)
	log.Printf("DB block count: %v  TX count: %v\n", dbBlockCount, dbTxCount)
	if cliArgs.IsPrintStatus {
		dashrpc.PrintStatus(db)
		return
	}

	dbStatus := dashrpc.DbGetStatus(db)
	log.Printf("DB status: %s\n", dbStatus)

	// API end points

	http.HandleFunc(getRouteTransaction(), handlerTxDetails(db, client))
	http.HandleFunc(getRouteAddress(), handlerAddressDetails(db, client))
	http.HandleFunc(getRouteBlock(), handlerBlockDetails(db, client))
	http.HandleFunc(getRouteMeta(), handlerMeta(db, client))
	http.HandleFunc(getRouteRoot(), handlerRoot)

	// start the server
	log.Println("Starting server on port", cliArgs.ExplorerServerPort)
	log.Fatal(http.ListenAndServe(":"+strconv.FormatUint(uint64(cliArgs.ExplorerServerPort), 10), nil))
}
