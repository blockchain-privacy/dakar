package main

import (
	"dashrpc"
	"dashrpc/rpcclient"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/dgraph-io/badger"
	"log"
	"net/http"
	"strconv"
)

//
// This is the ENTRY point for the web-based REST API backend.
//
// Work in progress.
//

const (
	// default server port
	SERVER_PORT = 8081
)



//
// handler builder
//
// API pattern: "/txt/<hash>"
// OUTPUT: dashrpc.Transaction
//
func handlerTxDetails(db *badger.DB, client *rpcclient.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r * http.Request) {
		txHashString := r.URL.Path[4:]
		txDetails := dashrpc.TxDetails{}
		err := dashrpc.DbGetTxDetails(db, txHashString, &txDetails)
		if err != nil {
			http.Error(w, err.Error() + " Key: "+txHashString, http.StatusNotFound)
			return
		}
		txHash, err := chainhash.NewHashFromStr(txHashString)
		if err != nil {
			http.Error(w, err.Error() + " Key: "+txHashString, http.StatusNotFound)
			return
		}

		tx, err := client.GetRawTransactionVerbose(txHash)
		if err != nil {
			http.Error(w, err.Error() + " Key: "+txHashString, http.StatusNotFound)
			return
		}
		block := dashrpc.Block{}
		err = dashrpc.DbGetBlock(db, tx.BlockHash, &block)
		if err != nil {
			http.Error(w, err.Error() + " Block hash: "+tx.BlockHash, http.StatusNotFound)
			return
		}
		transaction := dashrpc.Transaction{}
		transaction.Bhash = tx.BlockHash
		transaction.Bheight = block.Id
		transaction.Bts = block.Timestamp.Unix()
		transaction.Confirmations = tx.Confirmations
		transaction.Version = tx.Version
		transaction.Tx = txDetails
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8080")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, PUT, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "X-Requested-With, Content-Type, Authorization, Origin, Accept")
		// common['Access-Control-Request-Method'] = '*'")
		err = json.NewEncoder(w).Encode(transaction)
		if err != nil {
			http.Error(w, err.Error() + " TxDetails: "+txDetails.String(), http.StatusInternalServerError)
		}
	}
}


//
// Simple web-based utility to browse/lookup the TXs from the badger database
// It provides the API through HTTP
// Work in Progress. NOT WORKING YET.
//
func main() {
	fmt.Printf("Go DashRPC client  %s\nBlock explorer\n\n", dashrpc.VersionString)

	badgerDir := flag.String("db", "/tmp/badger", "badger database location")
	rpcUser := flag.String("rpcuser", "rpc1user", "Dash RPC user")
	rpcPassword := flag.String("rpcpassword", "1234pass", "Dash RPC password")
	isPrintStatus := flag.Bool("status", false, "prints current processing status")

	flag.Parse()
	db := dashrpc.SetupBadgerDB(*badgerDir)
	defer func() {
		e := db.Close()
		if e != nil { /* ignore */
		}
	}()

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

	// API end points
	http.HandleFunc("/tx/", handlerTxDetails(db, client))

	// start the server
	// TODO make the server port taken from the command-line parameters
	log.Fatal(http.ListenAndServe(":" + strconv.FormatUint(SERVER_PORT, 10), nil))
}
