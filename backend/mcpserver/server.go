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
	// what blockchain data this MCP server handles (e.g. dash or btc)
	blockchainMode string
	// if auth middleware shall be disabled
	disableAuth bool
}

func NewServer(db external.Database, worker *workspace.Worker, graphWrapper *graph.Wrapper, blockchainMode string) *Server {
	return &Server{
		db:             db,
		worker:         worker,
		graphWrapper:   graphWrapper,
		handler:        http.NewServeMux(),
		blockchainMode: blockchainMode,
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

	const (
		toolGetTransaction   = "get_transaction"
		toolListHeuristics   = "list_heuristics"
		toolExecuteHeuristic = "execute_heuristic"
	)

	mcp.AddTool[TransactionParams, *db.FrontendTransaction](mcpServer, &mcp.Tool{
		Name: toolGetTransaction, Description: "get full transaction details." + blockchainDisclaimer(s.blockchainMode)},
		s.getTransaction())
	mcp.AddTool[any, *ListHeuristicsResult](mcpServer, &mcp.Tool{
		Name: toolListHeuristics, Description: "get a list of available CoinJoin heuristics." +
			blockchainDisclaimer(s.blockchainMode)}, s.listHeuristics())
	mcp.AddTool[ExecuteHeuristicParams, *ExecuteHeuristicResult](mcpServer, &mcp.Tool{
		Name: toolExecuteHeuristic, Description: fmt.Sprintf("runs a heuristic. get possible heuristic types "+
			"and parameter restrictions from the %s tool. %s", toolListHeuristics,
			blockchainDisclaimer(s.blockchainMode))}, s.executeHeuristic())

	h := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return mcpServer }, &mcp.StreamableHTTPOptions{
		SessionTimeout: time.Minute * 2,
	})

	handler := http.NewServeMux()

	if s.disableAuth {
		handler.Handle("/", h)
	} else {
		handler.Handle("/", s.adapt(h, server.Authorization()))
	}

	srv := &http.Server{
		Addr:              ":" + strconv.FormatUint(uint64(port), 10),
		Handler:           handler,
		ReadTimeout:       time.Minute,
		ReadHeaderTimeout: time.Second * 5,
	}

	srv.RegisterOnShutdown(func() {
		for session := range mcpServer.Sessions() {
			_ = session.Close()
		}
	})

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
	return mw.Adapt(h, adapters...)
}
