package main

import (
	"context"
	"dashrpc/rpcclient"
	"fmt"
	"github.com/dgraph-io/dgo/v2"
	"net/http"
	"strconv"
	"sync"
	"time"
)

// creates a http server on the given port
func createServer(wg *sync.WaitGroup, port uint, dgraph *dgo.Dgraph, client *rpcclient.Client) *http.Server {
	// setup REST API
	setupHandlers(dgraph, client)

	// create server
	srv := &http.Server{
		Addr: ":" + strconv.FormatUint(uint64(port), 10),
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverFatal("server error:", err)
		}
		wg.Done()
	}()

	serverInfo(fmt.Sprintf("Starting server at endpoint http://localhost%s", srv.Addr))
	return srv
}

// sends a shutdown signal to the server with a timout of 5 seconds
func shutdownServer(server *http.Server) {
	if server == nil {
		return
	}
	serverInfo("### Shutting down server###")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		// extra handling here
		cancel()
	}()

	if err := server.Shutdown(ctx); err != nil {
		serverInfo("Server Shutdown Failed:", err)
	}
}
