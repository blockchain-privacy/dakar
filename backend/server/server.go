package server

import (
	heuristic "backend/analytics/heuristics"
	"backend/cmd/cliutil"
	"backend/external"
	"errors"
	"fmt"
	"github.com/dgraph-io/ristretto"
	ory "github.com/ory/kratos-client-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

// loggerPrefix is the prefix which is printed for each log message
const loggerPrefix = "\033[0;34mserver\u001B[0m\t"

// maxBodySize is the maximum number of bytes a body can contain
// without an error being thrown while it being read
const maxBodySize = 1048576 // 1048576 = 1024 * 1024 -> 1 MiB

var thisLogger = log.New(log.Writer(), loggerPrefix, log.Flags())

// InitLogger creates new loggers with the given parameters.
func InitLogger(out io.Writer, flag int) {
	thisLogger = log.New(out, loggerPrefix, flag)
}

func info(v ...interface{}) {
	thisLogger.Println(v...)
	cliutil.PrintStack(thisLogger, v...)
}

func fatal(v ...interface{}) {
	info(v...)
	os.Exit(1)
}

type Server struct {
	db        external.Database
	client    external.RPCClient
	worker    *heuristic.Worker
	cache     *ristretto.Cache
	auth      *ory.APIClient
	adminAuth *ory.APIClient
	handler   *http.ServeMux
}

func NewServer(db external.Database, adminAuth *ory.APIClient, auth *ory.APIClient, client external.RPCClient,
	worker *heuristic.Worker) (*Server, error) {
	if adminAuth == nil || auth == nil {
		return nil, cliutil.NewStackErrorStr("authentication handles are not set")
	}

	if worker == nil {
		return nil, cliutil.NewStackErrorStr("worker pointer is nil")
	}

	// init cache
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,     // number of keys to track frequency of (10 M).
		MaxCost:     1 << 30, // maximum cost of cache (1 GB).
		BufferItems: 64,      // number of keys per Get buffer.
	})
	if err != nil {
		return nil, err
	}

	return &Server{
		db:        db,
		client:    client,
		worker:    worker,
		cache:     cache,
		auth:      auth,
		adminAuth: adminAuth,
		handler:   http.NewServeMux(),
	}, nil
}

// StartServer creates an api server on the given port
func (s *Server) StartServer(wg *sync.WaitGroup, port uint) *http.Server {
	// setup REST API
	s.setupHandlers()

	// create server
	srv := &http.Server{
		Addr:              ":" + strconv.FormatUint(uint64(port), 10),
		Handler:           s.handler,
		ReadTimeout:       time.Minute,
		ReadHeaderTimeout: time.Second * 5,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			fatal("server error:", err)
		}
		wg.Done()
	}()

	info(fmt.Sprintf("Started API server at endpoint http://localhost%s", srv.Addr))

	return srv
}

// StartMetrics creates a metrics server on the given port
func StartMetrics(wg *sync.WaitGroup, port uint) *http.Server {
	handler := http.NewServeMux()
	handler.Handle(getRouteMetrics(), adapt(promhttp.Handler(), getRouteMetrics(),
		limitMethod("GET"), maxBody()))

	// create server
	srv := &http.Server{
		Addr:              ":" + strconv.FormatUint(uint64(port), 10),
		Handler:           handler,
		ReadTimeout:       time.Minute,
		ReadHeaderTimeout: time.Second * 5,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && errors.Is(err, http.ErrServerClosed) {
			fatal("server error:", err)
		}
		wg.Done()
	}()

	info(fmt.Sprintf("Started metrics server at endpoint http://localhost%s", srv.Addr))

	return srv
}
