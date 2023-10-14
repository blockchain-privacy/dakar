package main

import (
	"fmt"
	"github.com/swaggo/http-swagger/v2"
	"net/http"
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

	if err := srv.ListenAndServe(); err != nil {
		fmt.Println(err)
		return
	}
}
