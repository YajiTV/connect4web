package main

import (
	"log"
	"net/http"
	"power4/internal/app"
)

// main bootstraps the application and starts the HTTP server
func main() {
	// Boot returns the mux, a cleanup function, and an error if init fails
	mux, err := app.Boot("data")
	if err != nil {
		log.Fatal(err)
	}

	addr := ":8090"
	log.Println("Server started on", addr)

	// Starts the HTTP server
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}
