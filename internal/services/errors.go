package services

import "errors"

var (
	// Review errors
	ErrNilReview      = errors.New("review is nil")
	ErrInvalidRating  = errors.New("rating must be between 1 and 5")
	ErrReviewNotFound = errors.New("review not found")

	// User errors
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrUserNotFound       = errors.New("user not found")
	ErrNilUser            = errors.New("user is nil")
	ErrInvalidCredentials = errors.New("invalid credentials")

	// Dataset errors
	ErrNilDataset      = errors.New("dataset is nil")
	ErrVersionFail     = errors.New("failed to create version")
	ErrInvalidMetadata = errors.New("invalid metadata")
	ErrDatasetNotFound = errors.New("dataset not found")

	// Category errors
	ErrNilCategory      = errors.New("category is nil")
	ErrCategoryExists   = errors.New("category already exists")
	ErrCategoryNotEmpty = errors.New("category not empty")
	ErrCategoryNotFound = errors.New("category not found")

	// Notification errors
	ErrNotificationNotFound = errors.New("notification not found")
	ErrNoSubscribers        = errors.New("no subscribers found")

	ErrAlreadySubscribed = errors.New("user already subscribed to dataset")
	ErrNotSubscribed     = errors.New("subscription not found")
)
