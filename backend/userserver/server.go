// SPDX-FileCopyrightText: 2025 Michael Ziegler <michael.h.ziegler@ntnu.no>
// SPDX-FileCopyrightText: 2025 Mariusz Nowostawski <mariusz.nowostawski@ntnu.no>
// SPDX-License-Identifier: AGPL-3.0-or-later

package userserver

import (
	"backend/external"
	"backend/server"
	"errors"
	mw "gitlab.com/blockchain-privacy/gomisc/middleware"
	"gitlab.com/blockchain-privacy/gomisc/serror"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"
)

func info(msg string, v ...any) {
	slog.Info(msg, append([]any{"module", "server"}, v...)...)
}

func warn(err error, v ...any) {
	serror.Log(slog.Default(), err, v...)
}

type Server struct {
	// dgraph database
	db external.Database
	// HTTP mux
	handler *http.ServeMux
}

func NewServer(db external.Database) *Server {
	return &Server{
		db:      db,
		handler: http.NewServeMux(),
	}
}

// StartServer creates a user server on the given port
func (s *Server) StartServer(wg *sync.WaitGroup, port uint) *http.Server {
	handler := http.NewServeMux()

	const routeUsers = "users"
	const routeHealth = "health"
	handler.Handle(server.BuildPattern(http.MethodPost, routeUsers, ""),
		mw.Adapt(s.handlerCreateUser(), mw.MaxBody5MiB()))
	handler.Handle(server.BuildPattern(http.MethodDelete, routeUsers, "uid"),
		mw.Adapt(s.handlerDeleteUser(), mw.MaxBody5MiB()))
	handler.Handle(server.BuildPattern(http.MethodGet, routeHealth, ""),
		mw.Adapt(s.handlerHealth(), mw.MaxBody5MiB()))

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

	info("Started user server at endpoint http://localhost" + srv.Addr)

	return srv
}
