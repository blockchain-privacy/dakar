package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/swaggo/http-swagger/v2"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	// import for openapi docs
	_ "backend/openapi"
)

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/swagger/", httpSwagger.Handler(httpSwagger.URL("http://localhost:1323/swagger/doc.json")))

	srv := http.Server{
		Addr:         ":1323",
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// handle shutdown signals
	chSignal := make(chan os.Signal, 1)
	signal.Notify(chSignal, os.Interrupt, syscall.SIGTERM)

	go func() {
		fmt.Println("Server endpoint: http://localhost:1323/swagger/index.html")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Println(err)
		}
	}()

	// wait for interrupt signal
	if <-chSignal; true {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer func() {
			// extra handling here
			cancel()
		}()
		if err := srv.Shutdown(ctx); err != nil {
			log.Printf("Server was shutdown and returned error: %v", err)
		}
	}
}
