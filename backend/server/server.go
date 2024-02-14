package server

import (
	heuristic "backend/analytics/heuristics"
	"backend/cmd/cliutil"
	"backend/external"
	"errors"
	"github.com/dgraph-io/ristretto"
	ory "github.com/ory/kratos-client-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// maxBodySize is the maximum number of bytes a body can contain
// without an error being thrown while it being read
const maxBodySize = 5242880 // 5242880 = 1024 * 1024 * 5 -> 5 MiB

var thisLogger *slog.Logger

// InitLogger creates new loggers with the given parameters.
func InitLogger() {
	thisLogger = slog.With(slog.String("module", "server"))
}

func info(msg string, v ...any) {
	thisLogger.Info(msg, v...)
}

func warn(err error, v ...any) {
	cliutil.LogError(thisLogger, err, v...)
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
		return nil, cliutil.NewStackError(err)
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
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			warn(cliutil.NewStackError(err))
		}
		wg.Done()
	}()

	info("Started API server at endpoint http://localhost" + srv.Addr)

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
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			warn(cliutil.NewStackError(err))
		}
		wg.Done()
	}()

	info("Started metrics server at endpoint http://localhost" + srv.Addr)

	return srv
}
