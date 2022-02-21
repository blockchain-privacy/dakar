package server

import (
	heuristic "backend/analytics/heuristics"
	"backend/external"

	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"sync"
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
}

func fatal(v ...interface{}) {
	thisLogger.Fatalln(v...)
}

type Server struct {
	db              external.Database
	client          external.RPCClient
	worker          *heuristic.Worker
	basicAuthUser   string
	basicAuthHash   string
	tokenPublicKey  []byte
	tokenPrivateKey []byte
	handler         *http.ServeMux
}

func NewServer(db external.Database, client external.RPCClient, worker *heuristic.Worker,
	basicAuthUser string, basicAuthHash string, tokenPublicKey string, tokenPrivateKey string) (Server, error) {

	if tokenPublicKey == "" || tokenPrivateKey == "" {
		return Server{}, errors.New("keys are not set")
	}

	if basicAuthUser == "" {
		return Server{}, errors.New("basic authentication user is not set")
	}

	if basicAuthHash == "" {
		return Server{}, errors.New("basic authentication hash is not set")
	}

	if worker == nil {
		return Server{}, errors.New("worker pointer is nil")
	}

	privateKey, err := hex.DecodeString(tokenPrivateKey)
	if err != nil {
		return Server{}, err
	}

	publicKey, err := hex.DecodeString(tokenPublicKey)
	if err != nil {
		return Server{}, err
	}

	return Server{
		db:              db,
		client:          client,
		worker:          worker,
		basicAuthUser:   basicAuthUser,
		basicAuthHash:   basicAuthHash,
		tokenPublicKey:  publicKey,
		tokenPrivateKey: privateKey,
		handler:         http.NewServeMux(),
	}, nil
}

// StartServer creates a http server on the given port
func (s *Server) StartServer(wg *sync.WaitGroup, port uint) *http.Server {
	// setup REST API
	s.setupHandlers()

	// create server
	srv := &http.Server{
		Addr:    ":" + strconv.FormatUint(uint64(port), 10),
		Handler: s.handler,
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
