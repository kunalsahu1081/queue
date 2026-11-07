package main

import (
	"log"
	"net/http"

	transport "letsgo/internal/transport/http"
)

func main() {

	router := transport.NewRouter()

	// Start HTTP server
	server := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	log.Println("Server listening on :8080")

	if err := server.ListenAndServe(); err != nil &&
		err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
