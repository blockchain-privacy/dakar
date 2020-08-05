package main

import (
	dbstat "dashrpc/db/status"
	"dashrpc/rpcclient"
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/dgo/v2"
	"log"
	"net/http"
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

//// API pattern: "/api/v1/blk/<hash>"
//// OUTPUT: dashrpc.BlkDetails
//func handlerBlockDetails(db *badger.DB, client *rpcclient.Client) func(http.ResponseWriter, *http.Request) {
//	return func(w http.ResponseWriter, r *http.Request) {
//		log.Println("Accessed", r.URL.Path)
//		setDefaultHeader(w)
//
//		blkHashString := r.URL.Path[len(getRouteBlock()):]
//		block := dashrpc.Block{}
//		err := dashrpc.DbGetBlock(db, blkHashString, &block)
//		if err != nil {
//			http.Error(w, err.Error()+" Block hash: "+blkHashString, http.StatusNotFound)
//			return
//		}
//
//		// assignment to output struct
//		blkDetails := dashrpc.BlkDetails{
//			Hash:          block.Hash.String(),
//			Id:            block.Id,
//			NextBlockHash: block.NextBlockHash.String(),
//			PrevBlockHash: block.PrevBlockHash.String(),
//			TxHashes:      block.TxHashes,
//			Timestamp:     block.Timestamp,
//		}
//
//		// encoding
//		err = json.NewEncoder(w).Encode(blkDetails)
//		if err != nil {
//			http.Error(w, err.Error()+" Block: "+blkDetails.String(), http.StatusInternalServerError)
//		}
//	}
//}
//
//// API pattern: "/api/v1/address/<hash>"
//// OUTPUT: dashrpc.AddressData
//func handlerAddressDetails(db *badger.DB, client *rpcclient.Client) func(http.ResponseWriter, *http.Request) {
//	return func(w http.ResponseWriter, r *http.Request) {
//		log.Println("Accessed", r.URL.Path)
//		setDefaultHeader(w)
//
//		addressHashString := r.URL.Path[len(getRouteAddress()):]
//		addressData := dashrpc.AddressData{}
//
//		// assignment to output struct
//		err := dashrpc.DbGetDataForAddress(db, addressHashString, &addressData)
//		if err != nil {
//			http.Error(w, err.Error()+" Key: "+addressHashString, http.StatusNotFound)
//			return
//		}
//
//		// encoding
//		err = json.NewEncoder(w).Encode(addressData)
//		if err != nil {
//			http.Error(w, err.Error()+" AddressData: "+addressData.Address, http.StatusInternalServerError)
//		}
//	}
//}
//
//// API pattern: "/api/v1/tx/<hash>"
//// OUTPUT: dashrpc.Transaction
//func handlerTxDetails(db *badger.DB, client *rpcclient.Client) func(http.ResponseWriter, *http.Request) {
//	return func(w http.ResponseWriter, r *http.Request) {
//		log.Println("Accessed", r.URL.Path)
//		setDefaultHeader(w)
//
//		txHashString := r.URL.Path[len(getRouteTransaction()):]
//		log.Println(txHashString)
//		txDetails := dashrpc.TxDetails{}
//		err := dashrpc.DbGetTxDetails(db, txHashString, &txDetails)
//		if err != nil {
//			http.Error(w, err.Error()+" Key: "+txHashString, http.StatusNotFound)
//			return
//		}
//		txHash, err := chainhash.NewHashFromStr(txHashString)
//		if err != nil {
//			http.Error(w, err.Error()+" Key: "+txHashString, http.StatusNotFound)
//			return
//		}
//
//		tx, err := client.GetRawTransactionVerbose(txHash)
//		if err != nil {
//			http.Error(w, err.Error()+" Key: "+txHashString, http.StatusNotFound)
//			return
//		}
//		block := dashrpc.Block{}
//		err = dashrpc.DbGetBlock(db, tx.BlockHash, &block)
//		if err != nil {
//			http.Error(w, err.Error()+" Block hash: "+tx.BlockHash, http.StatusNotFound)
//			return
//		}
//
//		// assignment to output struct
//		transaction := dashrpc.Transaction{
//			Bhash:         tx.BlockHash,
//			Bheight:       block.Id,
//			Bts:           block.Timestamp.Unix(),
//			Confirmations: tx.Confirmations,
//			Version:       tx.Version,
//			Tx:            txDetails,
//		}
//
//		// encoding
//		err = json.NewEncoder(w).Encode(transaction)
//		if err != nil {
//			http.Error(w, err.Error()+" TxDetails: "+txDetails.String(), http.StatusInternalServerError)
//		}
//	}
//}

// API pattern: "/api/v1/meta/"
// OUTPUT: dashrpc.Transaction
func handlerMeta(dgraph *dgo.Dgraph) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("Accessed", r.URL.Path)
		setDefaultHeader(w)

		verboseStatus, err := dbstat.GetVerbose(dgraph)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// encoding
		err = json.NewEncoder(w).Encode(verboseStatus)
		if err != nil {
			http.Error(w, err.Error()+" Meta information: "+verboseStatus.String(), http.StatusInternalServerError)
		}
	}
}

// creates endpoint handlers
func setupHandlers(db *dgo.Dgraph, client *rpcclient.Client) {
	// API end points
	//http.HandleFunc(getRouteTransaction(), handlerTxDetails(db, client))
	//http.HandleFunc(getRouteAddress(), handlerAddressDetails(db, client))
	//http.HandleFunc(getRouteBlock(), handlerBlockDetails(db, client))
	http.HandleFunc(getRouteMeta(), handlerMeta(db))
	http.HandleFunc(getRouteRoot(), handlerRoot)
}
