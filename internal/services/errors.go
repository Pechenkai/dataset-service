package services

import "errors"

var (
	// Review errors
	ErrNilReview      = errors.New("review is nil")
	ErrInvalidRating  = errors.New("rating must be between 1 and 5")
	ErrReviewNotFound = errors.New("review not found")

	// User errors
	ErrUserExists             = errors.New("user already exists")
	ErrInvalidPassword        = errors.New("invalid password")
	ErrUserNotFound           = errors.New("user not found")
	ErrNilUser                = errors.New("user is nil")
	ErrInvalidCredentials     = errors.New("invalid credentials")
	ErrUserBlocked            = errors.New("user is blocked")
	ErrInvalidTwoFACode       = errors.New("invalid 2fa code")
	ErrTwoFAExpired           = errors.New("2fa challenge expired")
	ErrTwoFAChallengeNotFound = errors.New("2fa challenge not found")
	ErrTwoFADebugDisabled     = errors.New("two-factor debug access is disabled")

	// Dataset errors
	ErrNilDataset      = errors.New("dataset is nil")
	ErrVersionFail     = errors.New("failed to create version")
	ErrInvalidMetadata = errors.New("invalid metadata")
	ErrDatasetNotFound = errors.New("dataset not found")
	ErrVersionNotFound = errors.New("version not found")

	// Category errors
	ErrNilCategory      = errors.New("category is nil")
	ErrCategoryExists   = errors.New("category already exists")
	ErrCategoryNotEmpty = errors.New("category not empty")
	ErrCategoryNotFound = errors.New("category not found")
	ErrCategoryCreate   = errors.New("failed to create category")

	// Notification errors
	ErrNotificationNotFound = errors.New("notification not found")
	ErrNoSubscribers        = errors.New("no subscribers found")

	// Subscription errors
	ErrAlreadySubscribed = errors.New("user already subscribed to dataset")
	ErrNotSubscribed     = errors.New("subscription not found")

	// Auth errors
	ErrTokenNotFound = errors.New("token not found")
	ErrTokenInvalid  = errors.New("invalid token")

	// Access request errors
	ErrRequestNotFound      = errors.New("not found")
	ErrBadRequest           = errors.New("bad request")
	ErrRequestAlreadyExists = errors.New("request already exists")
	ErrRequestForbidden     = errors.New("request is forbidden")
)
