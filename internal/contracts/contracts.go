package contracts

import (
	"encoding/json"
	"time"

	"ppo/internal/entities"
)

const (
	MsgTypeHealthRequest              = "health.request"
	MsgTypeHealthResponse             = "health.response"
	MsgTypeCategoryListRequest        = "categories.list.request"
	MsgTypeCategoryListResponse       = "categories.list.response"
	MsgTypeDatasetListRequest         = "datasets.list.request"
	MsgTypeDatasetListResponse        = "datasets.list.response"
	MsgTypeDatasetGetRequest          = "datasets.get.request"
	MsgTypeDatasetGetResponse         = "datasets.get.response"
	MsgTypeVersionsListRequest        = "versions.list.request"
	MsgTypeVersionsListResponse       = "versions.list.response"
	MsgTypeVersionGetRequest          = "versions.get.request"
	MsgTypeVersionGetResponse         = "versions.get.response"
	MsgTypeNotificationsRequest       = "notifications.list.request"
	MsgTypeNotificationsResponse      = "notifications.list.response"
	MsgTypeAccessPendingRequest       = "access.pending.request"
	MsgTypeAccessPendingResponse      = "access.pending.response"
	MsgTypeAccessFindRequest          = "access.find.request"
	MsgTypeAccessFindResponse         = "access.find.response"
	MsgTypeAccessUpdateRequest        = "access.update.request"
	MsgTypeAccessUpdateResponse       = "access.update.response"
	MsgTypeNotificationCreateRequest  = "notification.create.request"
	MsgTypeNotificationCreateResponse = "notification.create.response"
	MsgTypeNotificationMarkRequest    = "notification.mark.request"
	MsgTypeNotificationMarkResponse   = "notification.mark.response"
	MsgTypeDatasetCreateRequest       = "dataset.create.request"
	MsgTypeDatasetCreateResponse      = "dataset.create.response"
	MsgTypeDatasetUpdateRequest       = "dataset.update.request"
	MsgTypeDatasetUpdateResponse      = "dataset.update.response"
	MsgTypeDatasetDeleteRequest       = "dataset.delete.request"
	MsgTypeDatasetDeleteResponse      = "dataset.delete.response"
	MsgTypeVersionAddRequest          = "version.add.request"
	MsgTypeVersionAddResponse         = "version.add.response"
	MsgTypeVersionDeleteRequest       = "version.delete.request"
	MsgTypeVersionDeleteResponse      = "version.delete.response"
	MsgTypeReviewCreateRequest        = "review.create.request"
	MsgTypeReviewCreateResponse       = "review.create.response"
	MsgTypeReviewUpdateRequest        = "review.update.request"
	MsgTypeReviewUpdateResponse       = "review.update.response"
	MsgTypeReviewDeleteRequest        = "review.delete.request"
	MsgTypeReviewDeleteResponse       = "review.delete.response"
	MsgTypeSubscriptionCreateRequest  = "subscription.create.request"
	MsgTypeSubscriptionCreateResponse = "subscription.create.response"
	MsgTypeSubscriptionDeleteRequest  = "subscription.delete.request"
	MsgTypeSubscriptionDeleteResponse = "subscription.delete.response"
	MsgTypeCategoryCreateRequest      = "category.create.request"
	MsgTypeCategoryCreateResponse     = "category.create.response"
	MsgTypeCategoryUpdateRequest      = "category.update.request"
	MsgTypeCategoryUpdateResponse     = "category.update.response"
	MsgTypeCategoryDeleteRequest      = "category.delete.request"
	MsgTypeCategoryDeleteResponse     = "category.delete.response"
	MsgTypeUserCreateRequest          = "user.create.request"
	MsgTypeUserCreateResponse         = "user.create.response"
	MsgTypeUserUpdateRequest          = "user.update.request"
	MsgTypeUserUpdateResponse         = "user.update.response"
	MsgTypeUserDeleteRequest          = "user.delete.request"
	MsgTypeUserDeleteResponse         = "user.delete.response"
	MsgTypeUserGetRequest             = "user.get.request"
	MsgTypeUserGetResponse            = "user.get.response"
	MsgTypeUnknown                    = "unknown"
	DefaultServiceNameGateway         = "gateway"
	DefaultServiceNameCore            = "core"
	DefaultServiceNameData            = "data"
	MessageErrorCodeUnknownType       = "unknown_type"
	MessageErrorCodeProcessError      = "processing_error"
)

type Envelope struct {
	ID            string          `json:"id"`
	Type          string          `json:"type"`
	Source        string          `json:"source,omitempty"`
	CorrelationID string          `json:"correlation_id,omitempty"`
	TraceID       string          `json:"trace_id,omitempty"`
	ActorID       uint64          `json:"actor_id,omitempty"`
	ActorRole     string          `json:"actor_role,omitempty"`
	OccurredAt    time.Time       `json:"occurred_at"`
	Payload       json.RawMessage `json:"payload,omitempty"`
	Error         *ErrorPayload   `json:"error,omitempty"`
	ReplyTo       string          `json:"reply_to,omitempty"`
}

type ErrorPayload struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

type ListCategoriesRequest struct {
	OnlyPublic bool `json:"only_public"`
}

type ListCategoriesResponse struct {
	Categories []*entities.Category `json:"categories"`
}

type HealthResponse struct {
	Service   string    `json:"service"`
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
}

type DatasetListRequest struct {
	OnlyPublic bool    `json:"only_public"`
	OwnerID    *uint64 `json:"owner_id,omitempty"`
}

type DatasetListResponse struct {
	Datasets []*entities.Dataset `json:"datasets"`
}

type DatasetGetRequest struct {
	ID uint64 `json:"id"`
}

type DatasetGetResponse struct {
	Dataset *entities.Dataset `json:"dataset"`
}

type VersionsListRequest struct {
	DatasetID uint64 `json:"dataset_id"`
}

type VersionsListResponse struct {
	Versions []*entities.DatasetVersion `json:"versions"`
}

type VersionGetRequest struct {
	VersionID uint64 `json:"version_id"`
}

type VersionGetResponse struct {
	Version *entities.DatasetVersion `json:"version"`
}

type NotificationsRequest struct {
	UserID uint64 `json:"user_id"`
}

type NotificationsResponse struct {
	Notifications []*entities.Notification `json:"notifications"`
}

type AccessPendingRequest struct {
	OwnerID uint64 `json:"owner_id"`
}

type AccessPendingResponse struct {
	Requests []*entities.AccessRequest `json:"requests"`
}

type AccessFindRequest struct {
	DatasetID uint64 `json:"dataset_id"`
	UserID    uint64 `json:"user_id"`
}

type AccessFindResponse struct {
	Request *entities.AccessRequest `json:"request,omitempty"`
}

type AccessUpdateRequest struct {
	RequestID uint64                `json:"request_id"`
	Status    entities.AccessStatus `json:"status"`
}

type AccessUpdateResponse struct {
	Updated bool `json:"updated"`
}

type NotificationCreateRequest struct {
	UserID    uint64 `json:"user_id"`
	DatasetID uint64 `json:"dataset_id"`
	Message   string `json:"message"`
}

type NotificationCreateResponse struct {
	ID uint64 `json:"id"`
}

type NotificationMarkRequest struct {
	NotificationID uint64 `json:"notification_id"`
	IsRead         bool   `json:"is_read"`
}

type NotificationMarkResponse struct {
	Updated bool `json:"updated"`
}

type DatasetCreateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	CategoryID  uint64 `json:"category_id"`
	OwnerID     uint64 `json:"owner_id"`
	IsPublic    bool   `json:"is_public"`
}

type DatasetCreateResponse struct {
	ID uint64 `json:"id"`
}

type DatasetUpdateRequest struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CategoryID  uint64 `json:"category_id"`
	IsPublic    bool   `json:"is_public"`
}

type DatasetUpdateResponse struct {
	Updated bool `json:"updated"`
}

type DatasetDeleteRequest struct {
	ID uint64 `json:"id"`
}

type DatasetDeleteResponse struct {
	Deleted bool `json:"deleted"`
}

type VersionAddRequest struct {
	DatasetID uint64 `json:"dataset_id"`
	Number    string `json:"number"`
	Filepath  string `json:"filepath"`
	ChangeLog string `json:"change_log"`
}

type VersionAddResponse struct {
	ID uint64 `json:"id"`
}

type VersionDeleteRequest struct {
	VersionID uint64 `json:"version_id"`
}

type VersionDeleteResponse struct {
	Deleted bool `json:"deleted"`
}

type ReviewCreateRequest struct {
	UserID    uint64          `json:"user_id"`
	DatasetID uint64          `json:"dataset_id"`
	Rating    entities.Rating `json:"rating"`
	Text      string          `json:"text"`
}

type ReviewCreateResponse struct {
	ID uint64 `json:"id"`
}

type ReviewUpdateRequest struct {
	ID     uint64          `json:"id"`
	Rating entities.Rating `json:"rating"`
	Text   string          `json:"text"`
}

type ReviewUpdateResponse struct {
	Updated bool `json:"updated"`
}

type ReviewDeleteRequest struct {
	ID uint64 `json:"id"`
}

type ReviewDeleteResponse struct {
	Deleted bool `json:"deleted"`
}

type SubscriptionCreateRequest struct {
	UserID    uint64 `json:"user_id"`
	DatasetID uint64 `json:"dataset_id"`
}

type SubscriptionCreateResponse struct {
	Created bool `json:"created"`
}

type SubscriptionDeleteRequest struct {
	UserID    uint64 `json:"user_id"`
	DatasetID uint64 `json:"dataset_id"`
}

type SubscriptionDeleteResponse struct {
	Deleted bool `json:"deleted"`
}

type CategoryCreateRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CategoryCreateResponse struct {
	ID uint64 `json:"id"`
}

type CategoryUpdateRequest struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CategoryUpdateResponse struct {
	Updated bool `json:"updated"`
}

type CategoryDeleteRequest struct {
	ID uint64 `json:"id"`
}

type CategoryDeleteResponse struct {
	Deleted bool `json:"deleted"`
}

type UserCreateRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
	Country  string `json:"country"`
	Role     string `json:"role"`
}

type UserCreateResponse struct {
	ID uint64 `json:"id"`
}

type UserUpdateRequest struct {
	ID       uint64 `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Country  string `json:"country"`
	Role     string `json:"role"`
	Block    *bool  `json:"block,omitempty"`
}

type UserUpdateResponse struct {
	Updated bool `json:"updated"`
}

type UserDeleteRequest struct {
	ID uint64 `json:"id"`
}

type UserDeleteResponse struct {
	Deleted bool `json:"deleted"`
}

type UserGetRequest struct {
	ID uint64 `json:"id"`
}

type UserGetResponse struct {
	User *entities.User `json:"user,omitempty"`
}
