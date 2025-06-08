package repositories

import "errors"

var (
	// Query errors
	ErrCategoryQueryBuild = errors.New("CategoryQueryBuild")
	ErrDatasetQueryBuild  = errors.New("failed to build dataset SQL")
	ErrUserQueryBuild     = errors.New("failed to build user SQL")
	ErrRequestQueryBuild  = errors.New("failed to build request SQL")

	// Dataset errors
	ErrDatasetNotFound = errors.New("dataset not found")
	ErrDatasetCreate   = errors.New("failed to create dataset")
	ErrDatasetUpdate   = errors.New("failed to update dataset")
	ErrDatasetDelete   = errors.New("failed to delete dataset")
	ErrDatasetScan     = errors.New("failed to scan dataset")
	ErrDatasetList     = errors.New("failed to list datasets")

	// Version errors
	ErrVersionQueryBuild = errors.New("failed to build version SQL")
	ErrVersionCreate     = errors.New("failed to create version")
	ErrVersionUpdate     = errors.New("failed to update version")
	ErrVersionDelete     = errors.New("failed to delete version")
	ErrVersionScan       = errors.New("failed to scan version row")
	ErrVersionList       = errors.New("failed to list versions")
	ErrVersionNotFound   = errors.New("version not found")

	// Metadata errors
	ErrMetadataQueryBuild = errors.New("failed to build metadata SQL")
	ErrMetadataCreate     = errors.New("failed to create metadata")
	ErrMetadataUpdate     = errors.New("failed to update metadata")
	ErrMetadataDelete     = errors.New("failed to delete metadata")
	ErrMetadataScan       = errors.New("failed to scan metadata row")
	ErrMetadataNotFound   = errors.New("metadata not found")
	ErrMetadataList       = errors.New("failed to list metadata")

	// Notification errors
	ErrNotificationQueryBuild  = errors.New("notification: failed to build SQL query")
	ErrNotificationCreate      = errors.New("notification: failed to create")
	ErrNotificationNotFound    = errors.New("notification not found")
	ErrNotificationDeleteFail  = errors.New("notification: failed to delete")
	ErrNotificationUpdateFail  = errors.New("notification: failed to update")
	ErrNotificationScanRow     = errors.New("notification: failed to scan row")
	ErrNotificationIterateRows = errors.New("notification: failed to iterate rows")

	// Review errors
	ErrReviewNotFound = errors.New("review not found")

	// Subscription errors
	ErrAlreadySubscribed    = errors.New("already subscribed")
	ErrSubscriptionNotFound = errors.New("subscription not found")

	// User errors
	ErrUserNotFound       = errors.New("user not found")
	ErrUserCreate         = errors.New("create user failed")
	ErrUserUpdate         = errors.New("update user failed")
	ErrUserDelete         = errors.New("delete user failed")
	ErrUserGet            = errors.New("get user failed")
	ErrUserGetAll         = errors.New("get all users failed")
	ErrUserScan           = errors.New("scan user row failed")
	ErrEmailAlreadyExists = errors.New("email already exists")

	// Category errors
	ErrCategoryNotFound      = errors.New("category not found")
	ErrCategoryNotEmpty      = errors.New("category not empty")
	ErrCategoryAlreadyExists = errors.New("category already exists")
	ErrCategoryCreate        = errors.New("category create")
	ErrCategoryUpdate        = errors.New("category update")
	ErrCategoryDelete        = errors.New("category delete")
	ErrCategoryFind          = errors.New("category find")
	ErrCategoryScan          = errors.New("category scan")
	ErrCategoryIterate       = errors.New("category iterate")

	// Request errors
	ErrRequestNotFound = errors.New("request not found")
	ErrRequestScan     = errors.New("scan request row failed")
)
