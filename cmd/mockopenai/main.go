package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"ppo/internal/integrations/openai"
)

func main() {
	addr := getenv("MOCK_OPENAI_ADDR", ":8088")
	latency := flag.Duration("latency", 0, "artificial latency")
	flag.StringVar(&addr, "addr", addr, "listen address, e.g. :8088")
	flag.Parse()

	handler := openai.NewMockHandler(openai.MockConfig{Latency: *latency})
	server := &http.Server{
		Addr:         addr,
		Handler:      handler,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	log.Printf("mock OpenAI server listening on %s (latency=%s)", addr, latency.String())
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("mock OpenAI server error: %v", err)
	}
}

func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
