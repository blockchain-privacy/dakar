package server

import (
	heuristic "backend/analytics/heuristics/transaction"
	"backend/external"

	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
)

// loggerPrefix is the prefix which is printed for each log message
const loggerPrefix = "\033[0;34mserver\u001B[0m\t"

// maxBodySize is the maximum number of bytes a body can contain without an error being thrown while being read
const maxBodySize = 1048576 // 1048576 = 1024 * 1024 -> 1 MiB

var thisLogger = log.New(log.Writer(), loggerPrefix, log.Flags())

// InitLogger creates new loggers with the given parameters.
func InitLogger(out io.Writer, flag int) {
	thisLogger = log.New(out, loggerPrefix, flag)
}

func info(v ...interface{}) {
	thisLogger.Println(v...)
}

func fatal(v ...interface{}) {
	thisLogger.Fatalln(v...)
}

// StartServer creates a http server on the given port
func StartServer(wg *sync.WaitGroup, port uint, basicAuthUser string, basicAuthHash string, tokenPublicKey string,
	tokenPrivateKey string, dgraph external.Database, client external.RPCClient,
	worker *heuristic.Worker) *http.Server {
	// setup REST API
	setupHandlers(dgraph, client, worker, basicAuthUser, basicAuthHash, tokenPublicKey, tokenPrivateKey)

	// create server
	srv := &http.Server{
		Addr: ":" + strconv.FormatUint(uint64(port), 10),
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fatal("server error:", err)
		}
		wg.Done()
	}()

	info(fmt.Sprintf("Started server at endpoint http://localhost%s", srv.Addr))

	return srv
}
