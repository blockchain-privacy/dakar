package main

import (
	"dashrpc/btcjson"
	"dashrpc/cmd/cliutil"
	dbaddr "dashrpc/db/address"
	dbblk "dashrpc/db/block"
	dbstat "dashrpc/db/status"
	dbtx "dashrpc/db/transaction"
	"dashrpc/rpcclient"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgraph-io/dgo/v2"
	"log"
	"net/http"
)

func serverInfo(v ...interface{}) {
	log.SetPrefix("\033[0;34mserver\u001B[0m\t")
	log.Println(v)
	log.SetPrefix("")
}

func serverFatal(v ...interface{}) {
	log.SetPrefix("\033[0;34mserver\u001B[0m\t")
	log.Fatalln(v)
	log.SetPrefix("")
}

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
	serverInfo("Accessed", r.URL.Path)
	setDefaultHeader(w)
	// not handling possible errors
	_, e := fmt.Fprintln(w, "Possible routes:")
	if e != nil {
		serverInfo(e)
	}

	_, e = fmt.Fprintln(w, "/\t\t-> This page")
	if e != nil {
		serverInfo(e)
	}

	_, e = fmt.Fprintln(w, "/tx/<hash>\t-> Transaction details")
	if e != nil {
		serverInfo(e)
	}

	_, e = fmt.Fprintln(w, "/address/<hash>\t-> Address details")
	if e != nil {
		serverInfo(e)
	}

	_, e = fmt.Fprintln(w, "/blk/<hash>\t-> Block details")
	if e != nil {
		serverInfo(e)
	}
	_, e = fmt.Fprintln(w, "/meta/\t\t-> Database meta information")
	if e != nil {
		serverInfo(e)
	}
}

// API pattern: "/api/v1/blk/<hash>"
// OUTPUT: dashrpc.BlkDetails
func handlerBlockDetails(dgraph *dgo.Dgraph) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		serverInfo("Accessed", r.URL.Path)
		setDefaultHeader(w)

		blkHashString := r.URL.Path[len(getRouteBlock()):]

		block, err := dbblk.GetFrontendBlock(dgraph, blkHashString)
		if err != nil {
			http.Error(w, "Block hash: "+blkHashString, http.StatusNotFound)

			// only print error if it is not expected
			if !errors.Is(err, dbblk.ErrorBlockNotFound) {
				serverInfo(cliutil.ShowCallInfo(), err)
			}

			return
		}

		// encoding
		err = json.NewEncoder(w).Encode(block)
		if err != nil {
			http.Error(w, "Block: "+block.String(), http.StatusInternalServerError)
			serverInfo(cliutil.ShowCallInfo(), err)
		}
	}
}

// API pattern: "/api/v1/address/<hash>"
// OUTPUT: dashrpc.AddressData
func handlerAddressDetails(dgraph *dgo.Dgraph) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		serverInfo("Accessed", r.URL.Path)
		setDefaultHeader(w)

		addressHashString := r.URL.Path[len(getRouteAddress()):]

		address, err := dbaddr.GetFrontendAddress(dgraph, addressHashString)
		if err != nil {
			http.Error(w, "Address: "+addressHashString, http.StatusNotFound)

			// only print error if it is not expected
			if !errors.Is(err, dbaddr.ErrorAddressNotFound) {
				serverInfo(cliutil.ShowCallInfo(), err)
			}

			return
		}

		// encoding
		err = json.NewEncoder(w).Encode(address)
		if err != nil {
			http.Error(w, "AddressData: "+address.String(), http.StatusInternalServerError)
			serverInfo(cliutil.ShowCallInfo(), err)
		}
	}
}

// API pattern: "/api/v1/tx/<hash>"
// OUTPUT: dashrpc.Transaction
func handlerTxDetails(dgraph *dgo.Dgraph) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		serverInfo("Accessed", r.URL.Path)
		setDefaultHeader(w)

		txHashString := r.URL.Path[len(getRouteTransaction()):]

		vTx, err := dbtx.GetFrontendTransaction(dgraph, txHashString)
		if err != nil {
			http.Error(w, "Transaction: "+txHashString, http.StatusInternalServerError)

			// only print error if it is not expected
			if !errors.Is(err, dbtx.ErrorTransactionNotFound) {
				serverInfo(cliutil.ShowCallInfo(), err)
			}

			return
		}

		// encoding
		err = json.NewEncoder(w).Encode(vTx)
		if err != nil {
			http.Error(w, "TxDetails: "+vTx.String(), http.StatusInternalServerError)
			serverInfo(cliutil.ShowCallInfo(), err)
		}
	}
}

// API pattern: "/api/v1/meta/"
// OUTPUT: dashrpc.Transaction
func handlerMeta(dgraph *dgo.Dgraph, client *rpcclient.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		serverInfo("Accessed", r.URL.Path)
		setDefaultHeader(w)

		verboseStatus, err := dbstat.GetFrontendStatus(dgraph)
		if err != nil {
			http.Error(w, "error getting status information", http.StatusInternalServerError)

			// only print error if it is not expected
			if errors.Is(err, dbaddr.ErrorAddressNotFound) {
				serverInfo(cliutil.ShowCallInfo(), err)
			}

			return
		}

		rpcInfo, err := client.GetInfo()
		if err != nil {
			http.Error(w, "error getting status information", http.StatusInternalServerError)
			serverInfo(cliutil.ShowCallInfo(), err)
			return
		}

		type metaStatus struct {
			Status  dbstat.FrontendStatus    `json:"status"`
			RPCInfo btcjson.InfoWalletResult `json:"rpcinfo"`
		}

		stat := metaStatus{
			Status:  verboseStatus,
			RPCInfo: *rpcInfo,
		}

		// encoding
		err = json.NewEncoder(w).Encode(stat)
		if err != nil {
			http.Error(w, "Meta information: "+verboseStatus.String(), http.StatusInternalServerError)
			serverInfo(cliutil.ShowCallInfo(), err)
		}
	}
}

// creates endpoint handlers
func setupHandlers(dgraph *dgo.Dgraph, client *rpcclient.Client) {
	// API end points
	http.HandleFunc(getRouteTransaction(), handlerTxDetails(dgraph))
	http.HandleFunc(getRouteAddress(), handlerAddressDetails(dgraph))
	http.HandleFunc(getRouteBlock(), handlerBlockDetails(dgraph))
	http.HandleFunc(getRouteMeta(), handlerMeta(dgraph, client))
	http.HandleFunc(getRouteRoot(), handlerRoot)
}
