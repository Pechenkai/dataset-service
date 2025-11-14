package v2

import (
	"encoding/json"
	"net/http"
	"time"
)

type apiError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type PaginationMeta struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type CategoryResponse struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CategoriesResponse struct {
	Items []CategoryResponse `json:"items"`
	Meta  PaginationMeta     `json:"meta"`
}

type DatasetMetadata struct {
	Format string   `json:"format,omitempty"`
	Tags   []string `json:"tags,omitempty"`
	Size   uint64   `json:"size,omitempty"`
}

type DatasetRatingSummary struct {
	Average float64   `json:"average"`
	Count   int       `json:"count"`
	Synced  time.Time `json:"synced_at"`
}

type DatasetVersionResponse struct {
	ID         uint64           `json:"id"`
	DatasetID  uint64           `json:"dataset_id"`
	Number     string           `json:"number"`
	ChangeLog  string           `json:"change_log,omitempty"`
	FileURL    string           `json:"file_url,omitempty"`
	UploadAt   time.Time        `json:"upload_date"`
	UploadedBy uint64           `json:"uploaded_by,omitempty"`
	Metadata   *DatasetMetadata `json:"metadata,omitempty"`
}

type DatasetResponse struct {
	ID            uint64                  `json:"id"`
	Name          string                  `json:"name"`
	Description   string                  `json:"description"`
	CategoryID    uint64                  `json:"category_id"`
	OwnerID       uint64                  `json:"owner_id"`
	IsPublic      bool                    `json:"is_public"`
	CreatedAt     time.Time               `json:"created_at"`
	UpdatedAt     *time.Time              `json:"updated_at,omitempty"`
	LatestVersion *DatasetVersionResponse `json:"latest_version,omitempty"`
	Metadata      *DatasetMetadata        `json:"metadata,omitempty"`
	RatingSummary DatasetRatingSummary    `json:"rating_summary"`
}

type DatasetsResponse struct {
	Items []DatasetResponse `json:"items"`
	Meta  PaginationMeta    `json:"meta"`
}

type DatasetVersionsResponse struct {
	Items []DatasetVersionResponse `json:"items"`
	Meta  PaginationMeta           `json:"meta"`
}

type NotificationResponse struct {
	ID        uint64    `json:"id"`
	UserID    uint64    `json:"user_id"`
	DatasetID uint64    `json:"dataset_id"`
	Message   string    `json:"message"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}

type NotificationsResponse struct {
	Items []NotificationResponse `json:"items"`
	Meta  PaginationMeta         `json:"meta"`
}

type ReviewResponse struct {
	ID        uint64    `json:"id"`
	DatasetID uint64    `json:"dataset_id"`
	UserID    uint64    `json:"user_id"`
	Rating    int       `json:"rating"`
	Text      string    `json:"text"`
	CreatedAt time.Time `json:"created_at"`
}

type ReviewsResponse struct {
	Items []ReviewResponse `json:"items"`
	Meta  PaginationMeta   `json:"meta"`
}

type SubscriptionResponse struct {
	ID           uint64    `json:"id"`
	DatasetID    uint64    `json:"dataset_id"`
	UserID       uint64    `json:"user_id"`
	SubscribedAt time.Time `json:"subscribed_at"`
}

type SubscriptionsResponse struct {
	Items []SubscriptionResponse `json:"items"`
	Meta  PaginationMeta         `json:"meta"`
}

type SubscribersResponse struct {
	Items []SubscriberResponse `json:"items"`
	Meta  PaginationMeta       `json:"meta"`
}

type SubscriberResponse struct {
	UserID uint64 `json:"user_id"`
}

type UserResponse struct {
	ID               uint64    `json:"id"`
	Username         string    `json:"username"`
	Email            string    `json:"email"`
	Country          string    `json:"country"`
	Role             string    `json:"role"`
	IsBlocked        bool      `json:"is_blocked"`
	RegistrationDate time.Time `json:"registration_date"`
}

type UsersResponse struct {
	Items []UserResponse `json:"items"`
	Meta  PaginationMeta `json:"meta"`
}

type AuthenticateResponse struct {
	Token     string       `json:"token"`
	ExpiresAt time.Time    `json:"expires_at"`
	User      UserResponse `json:"user"`
}

func writeJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, apiError{
		Code:    statusToCode(status),
		Message: err.Error(),
	})
}

func statusToCode(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "bad_request"
	case http.StatusUnauthorized:
		return "unauthorized"
	case http.StatusNotFound:
		return "not_found"
	case http.StatusConflict:
		return "conflict"
	case http.StatusFailedDependency:
		return "failed_dependency"
	default:
		return "internal_error"
	}
}
