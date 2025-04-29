package services

import "errors"

var (
	// Review errors
	ErrNilReview     = errors.New("review is nil")
	ErrInvalidRating = errors.New("rating must be between 1 and 5")

	// User errors
	ErrUserExists      = errors.New("user already exists")
	ErrInvalidPassword = errors.New("invalid password")
	ErrUserNotFound    = errors.New("user not found")
	ErrNilUser         = errors.New("user is nil")

	// Dataset errors
	ErrNilDataset  = errors.New("dataset is nil")
	ErrVersionFail = errors.New("failed to create version")

	// Category errors
	ErrNilCategory = errors.New("category is nil")

	// Notification errors
	ErrNotificationFound = errors.New("notification not found")
)
