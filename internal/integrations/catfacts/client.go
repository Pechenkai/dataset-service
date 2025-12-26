package catfacts

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"

	"ppo/internal/config"
)

type Fact struct {
	Text   string
	Length int
	Source string
	Mode   string
}

type Client interface {
	RandomFact(ctx context.Context) (Fact, error)
}

type HTTPClient struct {
	baseURL string
	mode    string
	client  *http.Client
	logger  *zap.Logger
}

func NewHTTPClient(cfg config.CatFactsConfig, logger *zap.Logger) *HTTPClient {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 4 * time.Second
	}
	baseURL := strings.TrimRight(selectBaseURL(cfg), "/")
	mode := normalizeMode(cfg.Mode)
	return &HTTPClient{
		baseURL: baseURL,
		mode:    mode,
		client: &http.Client{
			Timeout: timeout,
		},
		logger: logger,
	}
}

func (c *HTTPClient) RandomFact(ctx context.Context) (Fact, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/fact", nil)
	if err != nil {
		return Fact{}, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return Fact{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return Fact{}, fmt.Errorf("catfacts unexpected status: %d", resp.StatusCode)
	}

	var payload struct {
		Fact   string `json:"fact"`
		Length int    `json:"length"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return Fact{}, fmt.Errorf("decode catfacts response: %w", err)
	}
	if strings.TrimSpace(payload.Fact) == "" {
		return Fact{}, errors.New("catfacts response missing fact")
	}
	if payload.Length == 0 {
		payload.Length = len(payload.Fact)
	}

	return Fact{
		Text:   payload.Fact,
		Length: payload.Length,
		Source: c.baseURL,
		Mode:   c.mode,
	}, nil
}

func selectBaseURL(cfg config.CatFactsConfig) string {
	mode := normalizeMode(cfg.Mode)
	if mode == "mock" && strings.TrimSpace(cfg.MockBaseURL) != "" {
		return cfg.MockBaseURL
	}
	if strings.TrimSpace(cfg.RealBaseURL) != "" {
		return cfg.RealBaseURL
	}
	return "https://catfact.ninja"
}

func normalizeMode(raw string) string {
	mode := strings.ToLower(strings.TrimSpace(raw))
	if mode == "" {
		return "real"
	}
	if mode != "mock" {
		return "real"
	}
	return mode
}
