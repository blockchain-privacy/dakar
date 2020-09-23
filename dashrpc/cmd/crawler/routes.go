package main

import (
	"dashrpc/btcjson"
	"dashrpc/cmd/cliutil"
	dbaddr "dashrpc/db/address"
	dban "dashrpc/db/analytics"
	dbblk "dashrpc/db/block"
	dbstat "dashrpc/db/status"
	dbtx "dashrpc/db/transaction"
	"dashrpc/rpcclient"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgraph-io/dgo/v2"
	"net/http"
	"regexp"
	"strconv"
)

const (
	routePrefix      string = "/api/v1/"
	routeTransaction string = "tx/"
	routeBlock       string = "blk/"
	routeAddress     string = "address/"
	routeMeta        string = "meta/"
	routePaths       string = "paths/"
	routeRoot        string = ""
)

var (
	errorPath = "error getting paths"
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

func getRouteOrigins() string {
	return getRoute(routePaths)
}

func setDefaultHeader(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8080")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "X-Requested-With, Content-Type, Authorization, Origin, Accept")
}

// API pattern: "/api/v1/"
func handlerRoot(w http.ResponseWriter, r *http.Request) {
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

	_, e = fmt.Fprintln(w, "/paths/<hash>\t\t-> Get CSV file of paths")
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
func handlerBlockDetails(dgraph *dgo.Dgraph) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		blkHashString := r.URL.Path[len(getRouteBlock()):]

		if !isValid(blkHashString) {
			http.Error(w, "Block hash: "+blkHashString, http.StatusNotFound)
			return
		}

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
func handlerAddressDetails(dgraph *dgo.Dgraph) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		addressHashString := r.URL.Path[len(getRouteAddress()):]

		if !isValid(addressHashString) {
			http.Error(w, "Address: "+addressHashString, http.StatusNotFound)
			return
		}

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
func handlerTxDetails(dgraph *dgo.Dgraph) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		txHashString := r.URL.Path[len(getRouteTransaction()):]

		if !isValid(txHashString) {
			http.Error(w, "Transaction: "+txHashString, http.StatusNotFound)
			return
		}

		vTx, err := dbtx.GetFrontendTransaction(dgraph, txHashString)
		if err != nil {
			http.Error(w, "Transaction: "+txHashString, http.StatusNotFound)

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
func handlerMeta(dgraph *dgo.Dgraph, client *rpcclient.Client) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		verboseStatus, err := dbstat.GetFrontendStatus(dgraph)
		if err != nil {
			http.Error(w, "error getting status information", http.StatusInternalServerError)

			serverInfo(cliutil.ShowCallInfo(), err)

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

// API pattern: "/api/v1/paths/"
func handlerPaths(dgraph *dgo.Dgraph) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		txHashString := r.URL.Path[len(getRouteOrigins()):]

		if !isValid(txHashString) {
			http.Error(w, errorPath, http.StatusNotFound)
			return
		}

		originCount, err := dban.GetOriginCount(dgraph, txHashString)
		if err != nil {
			http.Error(w, errorPath, http.StatusNotFound)
			serverInfo(cliutil.ShowCallInfo(), err)
			return
		}

		// returned data is getting to big
		if originCount > 700 {
			http.Error(w, "getting paths is only supported up to 700 origins", http.StatusNotFound)
			return
		}

		paths, transactions, err := dban.GetPaths(dgraph, txHashString)
		if err != nil {
			http.Error(w, errorPath, http.StatusNotFound)
			serverInfo(cliutil.ShowCallInfo(), err)
			return
		}

		if len(paths) == 0 {
			http.Error(w, errorPath, http.StatusNotFound)
			return
		}

		// headers for streaming data to client
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.csv", txHashString))
		w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
		w.Header().Set("Content-Length", r.Header.Get("Content-Length"))

		csvWriter := csv.NewWriter(w)
		csvWriter.Comma = ';'

		header := []string{"path id", "path step", "tx hash", "type", "block hash", "block height", "timestamp"}
		if err = csvWriter.Write(header); err != nil {
			http.Error(w, "Error writing to csv stream", http.StatusInternalServerError)
			serverInfo(cliutil.ShowCallInfo(), err)
		}

		for i, p := range paths {
			for j, e := range p {
				tx := transactions[e.Hash]
				var row []string
				row = append(row, strconv.Itoa(i+1))
				row = append(row, strconv.Itoa(j+1))
				row = append(row, e.Hash)

				if e.IsOrigin {
					row = append(row, dbtx.PrivacyOrigin)
				} else {
					row = append(row, dbtx.PrivacyMixing)
				}

				row = append(row, tx.BlockHash)
				row = append(row, strconv.FormatUint(tx.BlockId, 10))
				row = append(row, tx.BlockTimestamp)
				if err = csvWriter.Write(row); err != nil {
					http.Error(w, "Error writing to csv stream", http.StatusInternalServerError)
					serverInfo(cliutil.ShowCallInfo(), err)
				}
			}
			csvWriter.Flush()
		}
	}
}

var isValidInput = regexp.MustCompile(`^[a-zA-Z0-9]*$`).MatchString

func isValid(input string) bool {
	if len(input) == 0 {
		return false
	}

	return isValidInput(input)
}

// creates endpoint handlers
func setupHandlers(dgraph *dgo.Dgraph, client *rpcclient.Client) {
	// API end points
	http.HandleFunc(getRouteTransaction(), handlerTxDetails(dgraph))
	http.HandleFunc(getRouteAddress(), handlerAddressDetails(dgraph))
	http.HandleFunc(getRouteBlock(), handlerBlockDetails(dgraph))
	http.HandleFunc(getRouteMeta(), handlerMeta(dgraph, client))
	http.HandleFunc(getRouteOrigins(), handlerPaths(dgraph))
	http.HandleFunc(getRouteRoot(), handlerRoot)
}
