// Package main runs the HTTP API
package main

import (
	"log"
	"net/http"

	"github.com/zaidmasri/business-planning-tool/internal/api"
	"github.com/zaidmasri/business-planning-tool/internal/store"
)

func main() {
	memStore := store.NewMemoryStore()

	handler := api.NewHandler(memStore)

	mux := http.NewServeMux()

	handler.RegisterRoutes(mux)

	port := ":8080"
	log.Printf("Server starting on port %s...", port)
	if err := http.ListenAndServe(port, mux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
