package main

import (
	"dashrpc"
	"dashrpc/rpcclient"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/dgraph-io/badger/v2"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
)

func setDefaultHeader(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8080")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "X-Requested-With, Content-Type, Authorization, Origin, Accept")
}

// API pattern: "/"
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
}

// API pattern: "/address/<hash>"
// OUTPUT: dashrpc.AddressData
func handlerAddressDetails(db *badger.DB, client *rpcclient.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Accessed", r.URL.Path)
		setDefaultHeader(w)

		addressHashString := r.URL.Path[9:]
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

// API pattern: "/txt/<hash>"
// OUTPUT: dashrpc.Transaction
func handlerTxDetails(db *badger.DB, client *rpcclient.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Accessed", r.URL.Path)
		setDefaultHeader(w)

		txHashString := r.URL.Path[4:]
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

// same as buildEndpoint in crawler.go
func buildExplorerEndpoint(rpcHost string, rpcPort uint) (string, error) {
	// check if ip is valid
	if ip := net.ParseIP(rpcHost); ip == nil {
		return "", errors.New("IP is not valid")
	}

	// build endpoint string
	return rpcHost + ":" + strconv.Itoa(int(rpcPort)), nil
}

type ExplorerCLIArguments struct {
	badgerDir     string
	rpcUser       string
	rpcPassword   string
	isPrintStatus bool
	rpcEndpoint   string
	logfile       string
	serverPort    uint16 // type also defines max port number
	err           error
}

func getExplorerCLIArgs() (cliArgs ExplorerCLIArguments) {
	badgerDir := flag.String("db", "/tmp/badger", "Badger database location")
	rpcUser := flag.String("rpcuser", "rpc1user", "Dash RPC user")
	rpcPassword := flag.String("rpcpassword", "1234pass", "Dash RPC password")
	rpcHost := flag.String("rpchost", "0.0.0.0", "Dash RPC host IP (default: 0.0.0.0)")
	rpcPort := flag.Uint("rpcport", 9998, "Dash RPC port (default: 9998)")
	serverPort := flag.Uint("serverport", 8081, "Explorer server port (default: 8081)")
	isPrintStatus := flag.Bool("status", false, "Prints current processing status")
	logfile := flag.String("logfile", "", "Specify log file (default: none)")
	flag.Parse()

	cliArgs.badgerDir = *badgerDir
	cliArgs.rpcUser = *rpcUser
	cliArgs.rpcPassword = *rpcPassword
	cliArgs.isPrintStatus = *isPrintStatus
	cliArgs.logfile = *logfile
	cliArgs.serverPort = uint16(*serverPort)

	ep, err := buildExplorerEndpoint(*rpcHost, *rpcPort)

	if err != nil {
		flag.PrintDefaults()
		cliArgs.err = err
		return cliArgs
	}

	cliArgs.rpcEndpoint = ep

	if len(cliArgs.badgerDir) == 0 {
		flag.PrintDefaults()
		cliArgs.err = errors.New("error: badger dir not set")
		return cliArgs
	}

	return cliArgs
}

//
// Simple web-based utility to browse/lookup the TXs from the badger database
// It provides the API through HTTP
// Work in Progress. NOT WORKING YET.
//
func main() {
	fmt.Printf("Go DashRPC client  %s\nBlock explorer\n\n", dashrpc.VersionString)

	cliArgs := getExplorerCLIArgs()
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
		log.SetPrefix("explorer ")
		log.SetOutput(io.MultiWriter(os.Stdout, f))
	}

	db := dashrpc.SetupBadgerDB(cliArgs.badgerDir)
	defer func() {
		e := db.Close()
		if e != nil { /* ignore */
		}
	}()

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

	// API end points
	http.HandleFunc("/tx/", handlerTxDetails(db, client))
	http.HandleFunc("/address/", handlerAddressDetails(db, client))
	http.HandleFunc("/", handlerRoot)

	// start the server
	log.Println("Starting server on port", cliArgs.serverPort)
	log.Fatal(http.ListenAndServe(":"+strconv.FormatUint(uint64(cliArgs.serverPort), 10), nil))
}
