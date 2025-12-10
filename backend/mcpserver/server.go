// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package mcpserver

import (
	"backend/analytics/graph"
	"backend/constants"
	"backend/db"
	"backend/external"
	"backend/server"
	"backend/workspace"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"

	mw "gitlab.com/blockchain-privacy/gomisc/middleware"
	"gitlab.com/blockchain-privacy/gomisc/serror"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func info(msg string, v ...any) {
	slog.Info(msg, append([]any{"module", "mcpserver"}, v...)...)
}

func warn(err error, v ...any) {
	serror.Log(slog.Default(), err, v...)
}

type Server struct {
	// dgraph database
	db external.Database
	// worker which sequentially processes work packages (currently only used for heuristics)
	worker *workspace.Worker
	// in-memory transaction and address graph of all classified transactions
	graphWrapper *graph.Wrapper
	// HTTP mux
	handler *http.ServeMux
	// duration after which every handler timeout
	handlerTimeout time.Duration
	// what blockchain data this MCP server handles (e.g. dash or btc)
	blockchainMode string
}

func NewServer(db external.Database, worker *workspace.Worker, graphWrapper *graph.Wrapper, blockchainMode string) *Server {
	return &Server{
		db:             db,
		worker:         worker,
		graphWrapper:   graphWrapper,
		handler:        http.NewServeMux(),
		blockchainMode: blockchainMode,
		handlerTimeout: time.Minute * 3,
	}
}

// blockchainTitle converts a given blockchain mode key to a human-readable string
func blockchainTitle(key string) string {
	switch key {
	case constants.BlockchainModeBTC:
		return "Bitcoin"
	case constants.BlockchainModeDash:
		return "Dash"
	default:
		return "invalid"
	}
}

func blockchainDisclaimer(key string) string {
	// leading space is intended
	return fmt.Sprintf(" This tool responds with data from the %s blockchain.", blockchainTitle(key))
}

// StartServer creates a user server on the given port
func (s *Server) StartServer(wg *sync.WaitGroup, port uint) *http.Server {
	mcpServer := mcp.NewServer(&mcp.Implementation{
		Name:    "dakar-mcp",
		Version: "1.0.0",
		Title:   "Dakar - CoinJoin Forensic Analysis",
	}, &mcp.ServerOptions{
		Instructions: fmt.Sprintf("This MCP server provides tools to analyse CoinJoin transactions. "+
			"It only works with the %s blockchain. Don't explain the data to the user, only respond with it", blockchainTitle(s.blockchainMode)),
	})

	mcp.AddTool[TransactionParams, *db.FrontendTransaction](mcpServer, &mcp.Tool{
		Name: "get_transaction", Description: "get full transaction details." + blockchainDisclaimer(s.blockchainMode)}, s.getTransaction())

	h := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return mcpServer }, nil)

	handler := http.NewServeMux()

	handler.Handle("/", s.adapt(h, server.Authorization()))

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

	info("Started mcp server at endpoint http://localhost" + srv.Addr)

	return srv
}

// adapt calls mw.Adapt() and inserts an http.TimeoutHandler into the adapter chain
func (s *Server) adapt(h http.Handler, adapters ...mw.Adapter) http.Handler {
	return mw.Adapt(h, append([]mw.Adapter{s.timeout()}, adapters...)...)
}

func (s *Server) timeout() mw.Adapter {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.TimeoutHandler(h, s.handlerTimeout, "request timed out").ServeHTTP(w, r)
		})
	}
}
