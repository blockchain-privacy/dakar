package main

import (
	"context"
	heuristic "dashrpc/analytics/heuristics/transaction"
	"dashrpc/btcjson"
	"dashrpc/cmd/cliutil"
	dbaddr "dashrpc/db/address"
	dban "dashrpc/db/analytics"
	dbtxh "dashrpc/db/analytics/heuristics/transaction"
	dbblk "dashrpc/db/block"
	dbstat "dashrpc/db/status"
	dbtx "dashrpc/db/transaction"
	"dashrpc/rpcclient"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/dgraph-io/dgo/v2"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"sync"
)

const (
	routePrefix              string = "/api/v1/"
	routeTransaction         string = "tx/"
	routeBlock               string = "blk/"
	routeAddress             string = "address/"
	routeMeta                string = "meta/"
	routePaths               string = "paths/"
	routeHeuristics          string = "heuristics/"
	routeHeuristicsSummary   string = "heuristicsSummary/"
	routeHeuristicsExecution string = "executeHeuristics/"
	routeHeuristicDetails    string = "heuristicDetails/"
)

const (
	maxOrigins = 1000
)

var (
	errorPath               = "error getting paths"
	errorHeuristics         = "error getting heuristics"
	errorHeuristicExecution = "error executing heuristics"
	errorHeuristicDetails   = "error getting heuristic details"
)

type reply struct {
	Message string `json:"msg,omitempty"`
}

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

func getRouteMeta() string {
	return getRoute(routeMeta)
}

func getRouteOrigins() string {
	return getRoute(routePaths)
}

func getRouteHeuristics() string {
	return getRoute(routeHeuristics)
}
func getRouteHeuristicsSummary() string {
	return getRoute(routeHeuristicsSummary)
}

func getRouteHeuristicsExecution() string {
	return getRoute(routeHeuristicsExecution)
}

func getRouteHeuristicDetails() string {
	return getRoute(routeHeuristicDetails)
}

func setDefaultHeader(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:8080")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, PUT, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "X-Requested-With, Content-Type, Authorization, Origin, Accept")
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

		rpcInfo, err := client.GetBlockChainInfo()
		if err != nil {
			http.Error(w, "error getting status information", http.StatusInternalServerError)
			serverInfo(cliutil.ShowCallInfo(), err)
			return
		}

		type metaStatus struct {
			Status  dbstat.FrontendStatus           `json:"status"`
			RPCInfo btcjson.GetBlockChainInfoResult `json:"rpcinfo"`
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

var lock sync.Mutex

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
		if originCount > maxOrigins {
			http.Error(w, "getting paths is only supported up to "+strconv.Itoa(maxOrigins)+" origins", http.StatusNotFound)
			return
		}

		lock.Lock()
		paths, transactions, err := dban.GetPaths(dgraph, txHashString)
		if err != nil {
			http.Error(w, errorPath, http.StatusNotFound)
			serverInfo(cliutil.ShowCallInfo(), err)
			return
		}
		lock.Unlock()

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

// API pattern: "/api/v1/heuristicsSummary/<hash>"
func handlerHeuristicsSummary(dgraph *dgo.Dgraph) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		txHashString := r.URL.Path[len(getRouteHeuristicsSummary()):]

		if !isValid(txHashString) {
			http.Error(w, errorPath, http.StatusNotFound)
			return
		}

		cHeuristic, err := dbtxh.GetFrontendHeuristic(dgraph, txHashString)
		if err != nil {
			log.Println(err)
			return
		}

		if len(cHeuristic.Heuristics) == 0 {
			http.Error(w, errorPath, http.StatusNotFound)
			return
		}

		shortestPaths := make(map[string]int)

		lock.Lock()
		for _, h := range cHeuristic.Heuristics {

			for _, r := range h.Results {
				if _, ok := shortestPaths[r.Uid]; !ok {
					pathLen, err := dbtxh.GetShortestPathLength(dgraph, cHeuristic.Uid, r.Uid)
					if err != nil {
						log.Println(err)
						return
					}
					shortestPaths[r.Uid] = pathLen
				}
			}
		}
		lock.Unlock()

		// headers for streaming data to client
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s.csv", txHashString))
		w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
		w.Header().Set("Content-Length", r.Header.Get("Content-Length"))

		csvWriter := csv.NewWriter(w)
		csvWriter.Comma = ';'

		header := []string{"heuristic uid", "parent heuristic uid", "child heuristic uid",
			"heuristic type", "heuristic parameter", "heuristic timestamp",
			"origin uid", "origin transaction hash", "origin timestamp",
			"origin address hash", "origin shortest path"}
		if err = csvWriter.Write(header); err != nil {
			http.Error(w, "Error writing to csv stream", http.StatusInternalServerError)
			serverInfo(cliutil.ShowCallInfo(), err)
		}

		for _, h := range cHeuristic.Heuristics {
			for _, r := range h.Results {
				var row []string
				// per heuristic information
				row = append(row, h.Uid)
				var parentHeuristic string
				if len(h.ParentHeuristic) > 0 {
					// only one parent heuristic is possible
					parentHeuristic = h.ParentHeuristic[0].Uid
				}
				row = append(row, parentHeuristic)

				var childHeuristics string
				for i, c := range h.ChildHeuristics {
					childHeuristics += c.Uid
					if i+1 < len(h.ChildHeuristics) {
						childHeuristics += ","
					}
				}

				row = append(row, childHeuristics)
				row = append(row, h.Type)
				row = append(row, h.Parameter)
				row = append(row, h.Timestamp)

				// per origin information
				row = append(row, r.Uid)
				row = append(row, r.TxHash)
				row = append(row, r.Timestamp)
				row = append(row, r.AddressHash)
				row = append(row, strconv.Itoa(shortestPaths[r.Uid]))

				if err = csvWriter.Write(row); err != nil {
					http.Error(w, "Error writing to csv stream", http.StatusInternalServerError)
					serverInfo(cliutil.ShowCallInfo(), err)
				}
			}
			csvWriter.Flush()
		}
	}
}

// API pattern: "/api/v1/heuristics/<hash>"
func handlerHeuristics(dgraph *dgo.Dgraph) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		txHashString := r.URL.Path[len(getRouteHeuristics()):]

		if !isValid(txHashString) {
			http.Error(w, errorHeuristics, http.StatusNotFound)
			return
		}

		heuristics, err := dbtxh.GetBasicFrontendHeuristic(dgraph, txHashString)
		if err != nil {
			http.Error(w, errorHeuristics, http.StatusNotFound)
			serverInfo(cliutil.ShowCallInfo(), err)
			return
		}

		// encoding
		err = json.NewEncoder(w).Encode(heuristics)
		if err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			serverInfo(cliutil.ShowCallInfo(), err)
		}
	}
}

// API pattern: "/api/v1/heuristicDetails/<hash>"
func handlerHeuristicsDetails(dgraph *dgo.Dgraph) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		txHashString := r.URL.Path[len(getRouteHeuristicDetails()):]

		if !isValid(txHashString) {
			http.Error(w, errorHeuristicExecution, http.StatusNotFound)
			return
		}

		type request struct {
			HeuristicUid string `json:"uid,omitempty"`
		}

		var heuristicRequest request

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&heuristicRequest)
		if err != nil {
			http.Error(w, errorHeuristicExecution, http.StatusNotFound)
			serverInfo(cliutil.ShowCallInfo(), err)
			return
		}

		if len(heuristicRequest.HeuristicUid) == 0 {
			http.Error(w, errorHeuristicDetails, http.StatusNotFound)
			return
		}

		frontendHeuristic, err := dbtxh.GetFrontendHeuristicByUid(dgraph, heuristicRequest.HeuristicUid, txHashString)
		if err != nil {
			http.Error(w, errorHeuristicDetails, http.StatusNotFound)
			return
		}

		// encoding
		err = json.NewEncoder(w).Encode(frontendHeuristic)
		if err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			serverInfo(cliutil.ShowCallInfo(), err)
		}
	}
}

// API pattern: "/api/v1/executeHeuristics/<hash>"
func handlerHeuristicsExecution(dgraph *dgo.Dgraph, worker *heuristic.Worker) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		setDefaultHeader(w)

		txHashString := r.URL.Path[len(getRouteHeuristicsExecution()):]

		if !isValid(txHashString) {
			http.Error(w, errorHeuristicExecution, http.StatusNotFound)
			return
		}

		type request struct {
			Changed []dbtxh.FrontendHeuristic `json:"changed,omitempty"`
			Deleted []string                  `json:"deleted,omitempty"`
		}

		var heuristicRequest request

		decoder := json.NewDecoder(r.Body)
		err := decoder.Decode(&heuristicRequest)
		if err != nil {
			http.Error(w, errorHeuristicExecution, http.StatusNotFound)
			serverInfo(cliutil.ShowCallInfo(), err)
			return
		}

		if len(heuristicRequest.Changed) == 0 && len(heuristicRequest.Deleted) == 0 {
			http.Error(w, errorHeuristicExecution, http.StatusNotFound)
			return
		}

		log.Println("Received", len(heuristicRequest.Changed), "changed heuristics")
		log.Println("Received", len(heuristicRequest.Deleted), "deleted heuristics")

		work, err := heuristic.CreateWork(dgraph, txHashString, heuristicRequest.Changed,
			heuristicRequest.Deleted)
		if err != nil {
			http.Error(w, errorHeuristicExecution, http.StatusNotFound)
			serverInfo(cliutil.ShowCallInfo(), err)
			return
		}

		log.Println("Added work:", worker.AddWork(txHashString, work))

		msg := reply{Message: fmt.Sprintf("Received %d changed and %d deleted heuristics",
			len(heuristicRequest.Changed), len(heuristicRequest.Deleted))}

		// encoding
		err = json.NewEncoder(w).Encode(msg)
		if err != nil {
			http.Error(w, "encoding error", http.StatusInternalServerError)
			serverInfo(cliutil.ShowCallInfo(), err)
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
func setupHandlers(ctx context.Context, dgraph *dgo.Dgraph, client *rpcclient.Client) {
	worker := heuristic.NewWorker()
	worker.StartWorking(ctx, dgraph)

	// API end points
	http.HandleFunc(getRouteTransaction(), handlerTxDetails(dgraph))
	http.HandleFunc(getRouteAddress(), handlerAddressDetails(dgraph))
	http.HandleFunc(getRouteBlock(), handlerBlockDetails(dgraph))
	http.HandleFunc(getRouteMeta(), handlerMeta(dgraph, client))
	http.HandleFunc(getRouteOrigins(), handlerPaths(dgraph))
	http.HandleFunc(getRouteHeuristicsSummary(), handlerHeuristicsSummary(dgraph))
	http.HandleFunc(getRouteHeuristics(), handlerHeuristics(dgraph))
	http.HandleFunc(getRouteHeuristicsExecution(), handlerHeuristicsExecution(dgraph, &worker))
	http.HandleFunc(getRouteHeuristicDetails(), handlerHeuristicsDetails(dgraph))
}
