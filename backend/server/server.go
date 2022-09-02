package server

import (
	heuristic "backend/analytics/heuristics"
	"backend/external"
	"github.com/dgraph-io/ristretto"
	ory "github.com/ory/kratos-client-go"
	"net/http/cookiejar"

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
	cache           *ristretto.Cache
	auth            *ory.APIClient
	adminAuth       *ory.APIClient
	basicAuthUser   string
	basicAuthHash   string
	tokenPublicKey  []byte
	tokenPrivateKey []byte
	handler         *http.ServeMux
}

func newOryClient(endpoint string) (*ory.APIClient, error) {

	cj, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	conf := ory.NewConfiguration()
	conf.Servers = ory.ServerConfigurations{{URL: endpoint}}

	conf.HTTPClient = &http.Client{Jar: cj}

	return ory.NewAPIClient(conf), nil
}

func NewServer(db external.Database, client external.RPCClient, worker *heuristic.Worker, basicAuthUser string,
	basicAuthHash string, tokenPublicKey string, tokenPrivateKey string) (*Server, error) {

	if tokenPublicKey == "" || tokenPrivateKey == "" {
		return nil, errors.New("keys are not set")
	}

	if basicAuthUser == "" {
		return nil, errors.New("basic authentication user is not set")
	}

	if basicAuthHash == "" {
		return nil, errors.New("basic authentication hash is not set")
	}

	if worker == nil {
		return nil, errors.New("worker pointer is nil")
	}

	privateKey, err := hex.DecodeString(tokenPrivateKey)
	if err != nil {
		return nil, err
	}

	publicKey, err := hex.DecodeString(tokenPublicKey)
	if err != nil {
		return nil, err
	}

	auth, err := newOryClient("http://localhost:4433")
	if err != nil {
		return nil, err
	} else if auth == nil {
		return nil, errors.New("authentication client is null")
	}

	adminAuth, err := newOryClient("http://localhost:4434")
	if err != nil {
		return nil, err
	} else if adminAuth == nil {
		return nil, errors.New("authentication client is null")
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
		db:              db,
		client:          client,
		worker:          worker,
		cache:           cache,
		auth:            auth,
		adminAuth:       adminAuth,
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
