package server

import (
	"backend/analytics/graph"
	"backend/external"
	"backend/worker"
	"backend/workspace"
	"errors"
	ory "github.com/ory/kratos-client-go"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	mw "github.com/qrest/gomisc/middleware"
	"github.com/qrest/gomisc/serror"
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
	serror.Log(thisLogger, err, v...)
}

type Server struct {
	// dgraph database
	db external.Database
	// Dash or Bitcoin RPC client
	client external.RPCClient
	// worker which sequentially processes work packages (currently only used for heuristics)
	worker *worker.Worker
	// in-memory transaction and address graph of all privacy transactions
	graphWrapper *graph.Wrapper
	// cache factory
	cacheFactory func(duration time.Duration) mw.Adapter
	// mutex map which synchronizes access to workspaces
	workspaceMutex *workspace.Mutex
	// ory kratos authentifaction handle
	auth *ory.APIClient
	// ory kratos admin authentifaction handle
	adminAuth *ory.APIClient
	// HTTP mux
	handler *http.ServeMux
}

func NewServer(db external.Database, adminAuth *ory.APIClient, auth *ory.APIClient, client external.RPCClient,
	worker *worker.Worker, graphWrapper *graph.Wrapper) (*Server, error) {
	if adminAuth == nil || auth == nil {
		return nil, serror.FromStr("authentication handles are not set")
	}

	if worker == nil {
		return nil, serror.FromStr("worker pointer is nil")
	}

	factory, err := mw.NewCacheFactory(thisLogger)
	if err != nil {
		return nil, err
	}

	return &Server{
		db:             db,
		client:         client,
		worker:         worker,
		graphWrapper:   graphWrapper,
		cacheFactory:   factory,
		auth:           auth,
		adminAuth:      adminAuth,
		workspaceMutex: workspace.NewMutex(),
		handler:        http.NewServeMux(),
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
			warn(serror.New(err))
		}
		wg.Done()
	}()

	info("Started API server at endpoint http://localhost" + srv.Addr)

	return srv
}

// StartMetrics creates a metrics server on the given port
func StartMetrics(wg *sync.WaitGroup, port uint) *http.Server {
	handler := http.NewServeMux()
	handler.Handle(http.MethodGet+" "+routeMetrics, mw.Adapt(promhttp.Handler(), mw.MaxBody5MiB()))

	// create server
	srv := &http.Server{
		Addr:              ":" + strconv.FormatUint(uint64(port), 10),
		Handler:           handler,
		ReadTimeout:       time.Minute,
		ReadHeaderTimeout: time.Second * 5,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			warn(serror.New(err))
		}
		wg.Done()
	}()

	info("Started metrics server at endpoint http://localhost" + srv.Addr)

	return srv
}
