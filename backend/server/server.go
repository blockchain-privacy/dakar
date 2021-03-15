package server

import (
	"backend/external"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/dgraph-io/dgo/v2"
)

// loggerPrefix is the prefix which is printed for each log message
const loggerPrefix = "\033[0;34mserver\u001B[0m\t"

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

type Server struct {
	server  *http.Server
	context context.Context
	cancel  context.CancelFunc
}

// CreateServer creates a http server on the given port
func CreateServer(wg *sync.WaitGroup, port uint, dgraph *dgo.Dgraph, client external.RPCClient) Server {
	ctx, cancelFunc := context.WithCancel(context.Background())
	// setup REST API
	setupHandlers(ctx, dgraph, client)

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

	info(fmt.Sprintf("Starting server at endpoint http://localhost%s", srv.Addr))

	return Server{server: srv, context: ctx, cancel: cancelFunc}
}

// ShutdownServer sends a shutdown signal to the server with a timout of 5 seconds
func (s *Server) ShutdownServer() {
	if s.server == nil {
		return
	}
	info("### Shutting down server###")

	s.cancel()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		// extra handling here
		cancel()
	}()

	if err := s.server.Shutdown(ctx); err != nil {
		info("Server Shutdown Failed:", err)
	}
}
