package entities

import "errors"

// ==== Category errors ====
var (
	ErrEmptyCategoryName        = errors.New("название категории не может быть пустым")
	ErrCategoryNameTooLong      = errors.New("название категории слишком длинное")
	ErrCategoryNameInvalidChars = errors.New("название категории содержит недопустимые символы")
)

// ==== Dataset errors ====
var (
	ErrEmptyDatasetName   = errors.New("название датасета не может быть пустым")
	ErrDatasetNameTooLong = errors.New("название датасета слишком длинное")
	ErrInvalidOwnerID     = errors.New("владелец не указан")
	ErrInvalidCategoryID  = errors.New("категория не указана")
	ErrInvalidCreatedAt   = errors.New("дата создания не может быть в будущем")
)

// ==== DatasetVersion errors ====
var (
	ErrEmptyVersionNumber   = errors.New("номер версии не может быть пустым")
	ErrVersionNumberTooLong = errors.New("номер версии слишком длинный")
	ErrEmptyFilepath        = errors.New("путь к файлу обязателен")
	ErrInvalidDatasetID     = errors.New("dataset ID не может быть 0")
	ErrUploadDateInFuture   = errors.New("дата загрузки не может быть в будущем")
)

// ==== Metadata errors ====
var (
	ErrEmptyFormat      = errors.New("формат файла не указан")
	ErrZeroSize         = errors.New("размер файла не может быть 0")
	ErrMissingVersionID = errors.New("не указана версия датасета")
)

// ==== Notification errors ====
var (
	ErrMissingUserID        = errors.New("не указан получатель уведомления")
	ErrMissingDatasetID     = errors.New("не указан датасет")
	ErrEmptyMessage         = errors.New("сообщение не может быть пустым")
	ErrNotificationInFuture = errors.New("время уведомления не может быть в будущем")
)

// ==== Review errors ====
var (
	ErrInvalidRating = errors.New("рейтинг должен быть от 1 до 5")
)

// ==== Subscription errors ====
var (
	ErrInvalidSubscriptionDate = errors.New("дата подписки не может быть в будущем")
)

// ==== User errors ====
var (
	ErrEmptyUsername           = errors.New("имя пользователя не может быть пустым")
	ErrEmptyEmail              = errors.New("email не может быть пустым")
	ErrEmptyPassword           = errors.New("пароль не может быть пустым")
	ErrInvalidRole             = errors.New("роль пользователя недопустима")
	ErrInvalidRegistrationDate = errors.New("дата регистрации не может быть в будущем")
)
