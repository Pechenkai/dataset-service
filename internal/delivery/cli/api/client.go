package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path"
	"strings"
	"time"
)

type Client struct {
	baseURL string
	client  *http.Client
	store   TokenStore
}

func NewClient(baseURL string, store TokenStore) (*Client, error) {
	if !strings.HasPrefix(baseURL, "http") {
		return nil, fmt.Errorf("invalid api base url: %s", baseURL)
	}
	return &Client{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		client:  &http.Client{Timeout: 30 * time.Second},
		store:   store,
	}, nil
}

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func (c *Client) get(ctx context.Context, p string, auth bool, v interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.makeURL(p), nil)
	if err != nil {
		return err
	}
	return c.doJSON(req, auth, v)
}

func (c *Client) postJSON(ctx context.Context, p string, payload interface{}, auth bool, v interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.makeURL(p), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.doJSON(req, auth, v)
}

func (c *Client) patchJSON(ctx context.Context, p string, payload interface{}, auth bool, v interface{}) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.makeURL(p), bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	return c.doJSON(req, auth, v)
}

func (c *Client) delete(ctx context.Context, p string, auth bool) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, c.makeURL(p), nil)
	if err != nil {
		return err
	}
	return c.doJSON(req, auth, nil)
}

func (c *Client) doJSON(req *http.Request, auth bool, v interface{}) error {
	if auth {
		token, err := c.store.Get()
		if err != nil {
			return err
		}
		if token == "" {
			return errors.New("please login first via 'user login'")
		}
		req.Header.Set("Authorization", "Bearer "+strings.TrimSpace(token))
	}
	req.Header.Set("Accept", "application/json")
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		var apiErr apiError
		if err := json.NewDecoder(resp.Body).Decode(&apiErr); err == nil && apiErr.Message != "" {
			return fmt.Errorf("%s", apiErr.Message)
		}
		return fmt.Errorf("api error: %s", resp.Status)
	}
	if v == nil {
		return nil
	}
	return json.NewDecoder(resp.Body).Decode(v)
}

func (c *Client) makeURL(p string) string {
	return c.baseURL + path.Clean("/"+p)
}

func (c *Client) writeMultipart(ctx context.Context, p string, fields map[string]string, fileField string, fileName string, file io.Reader, auth bool, v interface{}) error {
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	for key, value := range fields {
		if err := writer.WriteField(key, value); err != nil {
			return err
		}
	}
	if file != nil {
		fw, err := writer.CreateFormFile(fileField, fileName)
		if err != nil {
			return err
		}
		if _, err := io.Copy(fw, file); err != nil {
			return err
		}
	}
	if err := writer.Close(); err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.makeURL(p), &buf)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	return c.doJSON(req, auth, v)
}

func boolToString(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

func (c *Client) Logout() error {
	return c.store.Clear()
}

// additional helper types -------------------------------------------------

type categoryPayload struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type categoriesResponse struct {
	Items []categoryPayload `json:"items"`
}

type datasetPayload struct {
	ID          uint64    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CategoryID  uint64    `json:"category_id"`
	OwnerID     uint64    `json:"owner_id"`
	IsPublic    bool      `json:"is_public"`
	CreatedAt   time.Time `json:"created_at"`
}

type datasetsResponse struct {
	Items []datasetPayload `json:"items"`
}

type notificationPayload struct {
	ID        uint64    `json:"id"`
	DatasetID uint64    `json:"dataset_id"`
	UserID    uint64    `json:"user_id"`
	Message   string    `json:"message"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}

type notificationsResponse struct {
	Items []notificationPayload `json:"items"`
}

type reviewPayload struct {
	ID        uint64    `json:"id"`
	DatasetID uint64    `json:"dataset_id"`
	UserID    uint64    `json:"user_id"`
	Rating    int       `json:"rating"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

type reviewsResponse struct {
	Items []reviewPayload `json:"items"`
}

type authenticateResponse struct {
	Token     string      `json:"token"`
	ExpiresAt time.Time   `json:"expires_at"`
	User      userPayload `json:"user"`
}

type twoFAChallengeResponse struct {
	ChallengeID  string    `json:"challenge_id"`
	ExpiresAt    time.Time `json:"expires_at"`
	AttemptsLeft int       `json:"attempts_left"`
	Delivery     string    `json:"delivery"`
}

type userPayload struct {
	ID               uint64    `json:"id"`
	Username         string    `json:"username"`
	Email            string    `json:"email"`
	Country          string    `json:"country"`
	Role             string    `json:"role"`
	RegistrationDate time.Time `json:"registration_date"`
	IsBlocked        bool      `json:"is_blocked"`
}

type subscriptionPayload struct {
	UserID    uint64    `json:"user_id"`
	DatasetID uint64    `json:"dataset_id"`
	CreatedAt time.Time `json:"subscribed_at"`
}

type subscriptionsResponse struct {
	Items []subscriptionPayload `json:"items"`
}

type ratingSummaryPayload struct {
	Average float64 `json:"average"`
	Count   int     `json:"count"`
}

type datasetWithRating struct {
	ID            uint64               `json:"id"`
	Name          string               `json:"name"`
	CategoryID    uint64               `json:"category_id"`
	OwnerID       uint64               `json:"owner_id"`
	IsPublic      bool                 `json:"is_public"`
	CreatedAt     time.Time            `json:"created_at"`
	RatingSummary ratingSummaryPayload `json:"rating_summary"`
}
