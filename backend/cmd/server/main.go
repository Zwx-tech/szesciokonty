package main

import (
	"log"
	"net/http"
	"os"

	"github.com/Zwx-tech/szesciokonty/backend/internal/ws"
)

func main() {
	addr := env("ADDR", ":3000")
	hub := ws.NewHub()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		_, _ = w.Write([]byte("ok"))
	})
	mux.Handle("GET /ws", hub.Handler())

	log.Printf("listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
