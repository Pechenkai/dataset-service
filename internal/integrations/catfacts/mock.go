package catfacts

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

type MockConfig struct {
	Facts   []string
	Latency time.Duration
}

func NewMockHandler(cfg MockConfig) http.Handler {
	facts := cfg.Facts
	if len(facts) == 0 {
		facts = []string{
			"Dataset owners love cats that sleep on warm GPUs.",
			"Every cleaned dataset earns a cat purr of approval.",
			"Metadata is like cat whiskers: it keeps you out of trouble.",
		}
	}

	var counter uint64
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/fact" {
			http.NotFound(w, r)
			return
		}

		if cfg.Latency > 0 {
			time.Sleep(cfg.Latency)
		}

		idx := int(atomic.AddUint64(&counter, 1)-1) % len(facts)
		fact := strings.TrimSpace(facts[idx])
		if fact == "" {
			fact = "Cats like deterministic mock responses."
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"fact":   fact,
			"length": len(fact),
		})
	})
}

func DefaultMockHandler() http.Handler {
	return NewMockHandler(MockConfig{})
}
