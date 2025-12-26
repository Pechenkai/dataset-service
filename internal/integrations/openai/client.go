package openai

import (
	"bytes"
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

type Client interface {
	CreateSummary(ctx context.Context, prompt string) (Summary, error)
}

type Summary struct {
	Text      string
	Model     string
	Mode      string
	Source    string
	CreatedAt time.Time
}

type HTTPClient struct {
	cfg    config.OpenAIConfig
	client *http.Client
	logger *zap.Logger
	mode   string
	url    string
	model  string
}

func NewHTTPClient(cfg config.OpenAIConfig, logger *zap.Logger) *HTTPClient {
	mode := normalizeMode(cfg.Mode)
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	model := strings.TrimSpace(cfg.Model)
	if model == "" {
		model = "gpt-3.5-turbo"
	}
	baseURL := strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	if logger == nil {
		logger = zap.NewNop()
	}
	return &HTTPClient{
		cfg: cfg,
		client: &http.Client{
			Timeout: timeout,
		},
		logger: logger,
		mode:   mode,
		url:    baseURL,
		model:  model,
	}
}

func (c *HTTPClient) CreateSummary(ctx context.Context, prompt string) (Summary, error) {
	if strings.TrimSpace(prompt) == "" {
		return Summary{}, errors.New("empty prompt")
	}

	body := map[string]any{
		"model": c.model,
		"messages": []map[string]string{
			{"role": "system", "content": "You are a concise dataset summarizer. Provide 1-2 sentences."},
			{"role": "user", "content": prompt},
		},
		"max_tokens": clampMaxTokens(c.cfg.MaxTokens, 50, 512),
	}
	data, err := json.Marshal(body)
	if err != nil {
		return Summary{}, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.url+"/chat/completions", bytes.NewReader(data))
	if err != nil {
		return Summary{}, err
	}
	req.Header.Set("Content-Type", "application/json")
	if token := strings.TrimSpace(c.cfg.APIKey); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return Summary{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return Summary{}, fmt.Errorf("openai unexpected status: %d", resp.StatusCode)
	}

	var decoded completionResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return Summary{}, fmt.Errorf("decode openai response: %w", err)
	}
	content := decoded.FirstContent()
	if content == "" {
		return Summary{}, errors.New("openai response missing content")
	}

	return Summary{
		Text:      content,
		Model:     decoded.Model,
		Mode:      c.mode,
		Source:    c.url,
		CreatedAt: time.Now().UTC(),
	}, nil
}

type completionResponse struct {
	Model   string `json:"model"`
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (r completionResponse) FirstContent() string {
	if len(r.Choices) == 0 {
		return ""
	}
	return strings.TrimSpace(r.Choices[0].Message.Content)
}

func normalizeMode(raw string) string {
	mode := strings.ToLower(strings.TrimSpace(raw))
	if mode == "real" {
		return "real"
	}
	return "mock"
}

func clampMaxTokens(value, min, max int) int {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}
