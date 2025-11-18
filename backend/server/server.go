// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package server

import (
	"backend/analytics/graph"
	"backend/external"
	"backend/workspace"
	"errors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	mw "gitlab.com/blockchain-privacy/gomisc/middleware"
	"gitlab.com/blockchain-privacy/gomisc/serror"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// maxBodySize is the maximum number of bytes a body can contain
// without an error being thrown while it being read
const maxBodySize = 5242880 // 5242880 = 1024 * 1024 * 5 -> 5 MiB

func info(msg string, v ...any) {
	slog.Info(msg, append([]any{"module", "server"}, v...)...)
}

func warn(err error, v ...any) {
	serror.Log(slog.Default(), err, v...)
}

type Server struct {
	// dgraph database
	db external.Database
	// Dash or Bitcoin RPC client. Can be nil.
	client external.RPCClient
	// worker which sequentially processes work packages (currently only used for heuristics)
	worker *workspace.Worker
	// in-memory transaction and address graph of all classified transactions
	graphWrapper *graph.Wrapper
	// cache factory
	cacheFactory func(duration time.Duration) mw.Adapter
	// mutex map which synchronizes access to workspaces
	workspaceMutex *workspace.Mutex
	// HTTP mux
	handler *http.ServeMux
	// duration after which every handler timesout
	handlerTimeout time.Duration
}

func NewServer(m *workspace.Mutex, db external.Database, client external.RPCClient,
	worker *workspace.Worker, graphWrapper *graph.Wrapper) (*Server, error) {
	if worker == nil {
		return nil, serror.FromStr("worker pointer is nil")
	}

	factory, err := mw.NewCacheFactory(1024, func(err error) { warn(err) })
	if err != nil {
		return nil, err
	}

	return &Server{
		db: db,
		// rpc client can be nil
		client:         client,
		worker:         worker,
		graphWrapper:   graphWrapper,
		cacheFactory:   factory,
		workspaceMutex: m,
		handler:        http.NewServeMux(),
		handlerTimeout: time.Minute * 3,
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
