package main

import (
	"context"
	"github.com/dgraph-io/dgo/v2"
	"log"
	"net/http"
	"strconv"
	"time"
)

// creates a http server on the given port
func createServer(port uint, dgraph *dgo.Dgraph) *http.Server {
	// setup REST API
	setupHandlers(dgraph)

	// create server
	srv := &http.Server{
		Addr: ":" + strconv.FormatUint(uint64(port), 10),
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalln("server error:", err)
		}
	}()

	log.Println("Starting server at endpoint http://localhost", srv.Addr)
	return srv
}

// sends a shutdown signal to the server with a timout of 5 seconds
func shutdownServer(server *http.Server) {
	log.Println("### Shutting down server###")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		// extra handling here
		cancel()
	}()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalln("Server Shutdown Failed:", err)
	}
}
