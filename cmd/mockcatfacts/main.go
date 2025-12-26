package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"ppo/internal/integrations/catfacts"
)

func main() {
	addr := getenv("MOCK_CATFACTS_ADDR", ":9099")
	latency := flag.Duration("latency", 0, "artificial latency for testing (e.g. 200ms)")
	flag.StringVar(&addr, "addr", addr, "listen address")
	flag.Parse()

	handler := catfacts.NewMockHandler(catfacts.MockConfig{Latency: *latency})

	server := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}

	log.Printf("mock catfacts server listening on %s (latency=%s)", addr, latency.String())
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("mock catfacts server error: %v", err)
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
