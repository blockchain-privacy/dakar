package userserver

import (
	"backend/external"
	"backend/server"
	"errors"
	mw "github.com/qrest/gomisc/middleware"
	"github.com/qrest/gomisc/serror"
	"log/slog"
	"net/http"
	"strconv"
	"sync"
	"time"
)

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
	// User
	handler.Handle(server.BuildPattern(http.MethodPost, routeUsers, ""),
		mw.Adapt(s.handlerCreateUser(), mw.MaxBody5MiB()))
	handler.Handle(server.BuildPattern(http.MethodDelete, routeUsers, "uid"),
		mw.Adapt(s.handlerDeleteUser(), mw.MaxBody5MiB()))

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

	info("Started user server at endpoint http://localhost" + srv.Addr)

	return srv
}
