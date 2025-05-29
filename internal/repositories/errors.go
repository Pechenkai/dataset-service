package repositories

import "errors"

var (
	// Dataset errors
	ErrDatasetNotFound = errors.New("dataset not found")

	// Version errors
	ErrVersionNotFound = errors.New("version not found")

	// Metadata errors
	ErrMetadataNotFound = errors.New("metadata not found")

	// Notification errors
	ErrNotificationNotFound = errors.New("notification not found")

	// Review errors
	ErrReviewNotFound = errors.New("review not found")

	// Subscription errors
	ErrAlreadySubscribed    = errors.New("already subscribed")
	ErrSubscriptionNotFound = errors.New("subscription not found")

	// Use errors
	ErrUserNotFound = errors.New("user not found")

	// Category errors
	ErrCategoryNotFound      = errors.New("category not found")
	ErrCategoryNotEmpty      = errors.New("category not empty")
	ErrCategoryAlreadyExists = errors.New("category already exists")
)
