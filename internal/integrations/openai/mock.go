package openai

import (
	"encoding/json"
	"net/http"
	"strings"
	"sync/atomic"
	"time"
)

type MockConfig struct {
	Summaries []string
	Latency   time.Duration
}

func NewMockHandler(cfg MockConfig) http.Handler {
	summaries := cfg.Summaries
	if len(summaries) == 0 {
		summaries = []string{
			"Краткая сводка датасета: данные подготовлены и готовы к анализу.",
			"Датасет включает описания и метаданные для быстрой оценки качества.",
		}
	}
	var counter uint64

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet && strings.HasSuffix(r.URL.Path, "/chat/completions") {
			w.WriteHeader(http.StatusOK)
			return
		}
		if r.Method != http.MethodPost || !strings.HasSuffix(r.URL.Path, "/chat/completions") {
			http.NotFound(w, r)
			return
		}
		if cfg.Latency > 0 {
			time.Sleep(cfg.Latency)
		}
		idx := int(atomic.AddUint64(&counter, 1)-1) % len(summaries)
		summary := strings.TrimSpace(summaries[idx])
		if summary == "" {
			summary = "Сводка отсутствует."
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"model": "gpt-mock",
			"choices": []map[string]any{
				{
					"message": map[string]any{
						"content": summary,
					},
				},
			},
		})
	})
}

func DefaultMockHandler() http.Handler {
	return NewMockHandler(MockConfig{})
}
