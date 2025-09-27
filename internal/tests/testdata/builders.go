package testdata

import (
	"math/rand"
	"strings"
	"time"

	"ppo/internal/entities"
	"ppo/internal/services"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

type CategoryBuilder struct {
	id          uint64
	name        string
	description string
}

func NewCategoryBuilder() *CategoryBuilder {
	return &CategoryBuilder{
		name:        randomString("category"),
		description: "auto-generated category",
	}
}

func (b *CategoryBuilder) WithID(id uint64) *CategoryBuilder {
	b.id = id
	return b
}

func (b *CategoryBuilder) WithName(name string) *CategoryBuilder {
	b.name = name
	return b
}

func (b *CategoryBuilder) WithDescription(desc string) *CategoryBuilder {
	b.description = desc
	return b
}

func (b *CategoryBuilder) Build() *entities.Category {
	cat, err := entities.NewCategory(b.name, b.description)
	if err != nil {
		panic(err)
	}
	if b.id != 0 {
		cat.ID = b.id
	}
	return cat
}

type DatasetBuilder struct {
	id          uint64
	name        string
	description string
	ownerID     uint64
	categoryID  uint64
	isPublic    bool
	createdAt   time.Time
}

func NewDatasetBuilder() *DatasetBuilder {
	return &DatasetBuilder{
		name:        randomString("dataset"),
		description: "auto-generated dataset",
		ownerID:     1,
		categoryID:  1,
		isPublic:    true,
		createdAt:   time.Now().Add(-time.Minute),
	}
}

func (b *DatasetBuilder) WithID(id uint64) *DatasetBuilder {
	b.id = id
	return b
}

func (b *DatasetBuilder) WithName(name string) *DatasetBuilder {
	b.name = name
	return b
}

func (b *DatasetBuilder) WithDescription(description string) *DatasetBuilder {
	b.description = description
	return b
}

func (b *DatasetBuilder) WithOwner(id uint64) *DatasetBuilder {
	b.ownerID = id
	return b
}

func (b *DatasetBuilder) WithCategory(id uint64) *DatasetBuilder {
	b.categoryID = id
	return b
}

func (b *DatasetBuilder) WithCreatedAt(t time.Time) *DatasetBuilder {
	b.createdAt = t
	return b
}

func (b *DatasetBuilder) Public(isPublic bool) *DatasetBuilder {
	b.isPublic = isPublic
	return b
}

func (b *DatasetBuilder) Build() *entities.Dataset {
	ds, err := entities.NewDataset(b.name, b.description, b.ownerID, b.categoryID, b.isPublic, b.createdAt)
	if err != nil {
		panic(err)
	}
	if b.id != 0 {
		ds.ID = b.id
	}
	return ds
}

type DatasetVersionBuilder struct {
	id         uint64
	number     string
	filepath   string
	datasetID  uint64
	changeLog  string
	uploadDate time.Time
}

func NewDatasetVersionBuilder() *DatasetVersionBuilder {
	return &DatasetVersionBuilder{
		number:     "v1.0.0",
		filepath:   "/tmp/dataset-v1.bin",
		datasetID:  1,
		changeLog:  "initial upload",
		uploadDate: time.Now().Add(-time.Hour),
	}
}

func (b *DatasetVersionBuilder) WithID(id uint64) *DatasetVersionBuilder {
	b.id = id
	return b
}

func (b *DatasetVersionBuilder) WithNumber(number string) *DatasetVersionBuilder {
	b.number = number
	return b
}

func (b *DatasetVersionBuilder) WithFilepath(path string) *DatasetVersionBuilder {
	b.filepath = path
	return b
}

func (b *DatasetVersionBuilder) WithDatasetID(id uint64) *DatasetVersionBuilder {
	b.datasetID = id
	return b
}

func (b *DatasetVersionBuilder) WithChangeLog(log string) *DatasetVersionBuilder {
	b.changeLog = log
	return b
}

func (b *DatasetVersionBuilder) WithUploadDate(t time.Time) *DatasetVersionBuilder {
	b.uploadDate = t
	return b
}

func (b *DatasetVersionBuilder) Build() *entities.DatasetVersion {
	version, err := entities.NewDatasetVersion(b.number, b.filepath, b.changeLog, b.datasetID, b.uploadDate)
	if err != nil {
		panic(err)
	}
	if b.id != 0 {
		version.ID = b.id
	}
	return version
}

type MetadataBuilder struct {
	id        uint64
	format    string
	size      uint64
	tags      string
	versionID uint64
}

func NewMetadataBuilder() *MetadataBuilder {
	return &MetadataBuilder{
		format:    "csv",
		size:      1024,
		tags:      "auto,generated",
		versionID: 1,
	}
}

func (b *MetadataBuilder) WithID(id uint64) *MetadataBuilder {
	b.id = id
	return b
}

func (b *MetadataBuilder) WithFormat(format string) *MetadataBuilder {
	b.format = format
	return b
}

func (b *MetadataBuilder) WithSize(size uint64) *MetadataBuilder {
	b.size = size
	return b
}

func (b *MetadataBuilder) WithTags(tags string) *MetadataBuilder {
	b.tags = tags
	return b
}

func (b *MetadataBuilder) WithVersionID(id uint64) *MetadataBuilder {
	b.versionID = id
	return b
}

func (b *MetadataBuilder) Build() *entities.Metadata {
	metadata, err := entities.NewMetadata(b.format, b.tags, b.size, b.versionID)
	if err != nil {
		panic(err)
	}
	if b.id != 0 {
		metadata.ID = b.id
	}
	return metadata
}

type NotificationBuilder struct {
	id        uint64
	userID    uint64
	datasetID uint64
	message   string
	isRead    bool
	createdAt time.Time
}

func NewNotificationBuilder() *NotificationBuilder {
	return &NotificationBuilder{
		userID:    1,
		datasetID: 1,
		message:   "dataset updated",
		createdAt: time.Now().Add(-10 * time.Minute),
	}
}

func (b *NotificationBuilder) WithID(id uint64) *NotificationBuilder {
	b.id = id
	return b
}

func (b *NotificationBuilder) WithUserID(id uint64) *NotificationBuilder {
	b.userID = id
	return b
}

func (b *NotificationBuilder) WithDatasetID(id uint64) *NotificationBuilder {
	b.datasetID = id
	return b
}

func (b *NotificationBuilder) WithMessage(msg string) *NotificationBuilder {
	b.message = msg
	return b
}

func (b *NotificationBuilder) WithCreatedAt(t time.Time) *NotificationBuilder {
	b.createdAt = t
	return b
}

func (b *NotificationBuilder) Read(isRead bool) *NotificationBuilder {
	b.isRead = isRead
	return b
}

func (b *NotificationBuilder) Build() *entities.Notification {
	notif, err := entities.NewNotification(b.userID, b.datasetID, b.message, b.createdAt)
	if err != nil {
		panic(err)
	}
	if b.id != 0 {
		notif.ID = b.id
	}
	notif.IsRead = b.isRead
	return notif
}

type ReviewBuilder struct {
	id        uint64
	userID    uint64
	datasetID uint64
	rating    entities.Rating
	text      string
	createdAt time.Time
}

func NewReviewBuilder() *ReviewBuilder {
	return &ReviewBuilder{
		userID:    1,
		datasetID: 1,
		rating:    entities.Rating4,
		text:      "helpful dataset",
		createdAt: time.Now().Add(-2 * time.Hour),
	}
}

func (b *ReviewBuilder) WithID(id uint64) *ReviewBuilder {
	b.id = id
	return b
}

func (b *ReviewBuilder) WithUserID(id uint64) *ReviewBuilder {
	b.userID = id
	return b
}

func (b *ReviewBuilder) WithDatasetID(id uint64) *ReviewBuilder {
	b.datasetID = id
	return b
}

func (b *ReviewBuilder) WithRating(r entities.Rating) *ReviewBuilder {
	b.rating = r
	return b
}

func (b *ReviewBuilder) WithText(text string) *ReviewBuilder {
	b.text = text
	return b
}

func (b *ReviewBuilder) WithCreatedAt(t time.Time) *ReviewBuilder {
	b.createdAt = t
	return b
}

func (b *ReviewBuilder) Build() *entities.Review {
	review, err := entities.NewReview(b.userID, b.datasetID, b.rating, b.createdAt, b.text)
	if err != nil {
		panic(err)
	}
	if b.id != 0 {
		review.ID = b.id
	}
	return review
}

type SubscriptionBuilder struct {
	id        uint64
	userID    uint64
	datasetID uint64
	createdAt time.Time
}

func NewSubscriptionBuilder() *SubscriptionBuilder {
	return &SubscriptionBuilder{
		userID:    1,
		datasetID: 1,
		createdAt: time.Now().Add(-24 * time.Hour),
	}
}

func (b *SubscriptionBuilder) WithID(id uint64) *SubscriptionBuilder {
	b.id = id
	return b
}

func (b *SubscriptionBuilder) WithUserID(id uint64) *SubscriptionBuilder {
	b.userID = id
	return b
}

func (b *SubscriptionBuilder) WithDatasetID(id uint64) *SubscriptionBuilder {
	b.datasetID = id
	return b
}

func (b *SubscriptionBuilder) WithCreatedAt(t time.Time) *SubscriptionBuilder {
	b.createdAt = t
	return b
}

func (b *SubscriptionBuilder) Build() *entities.Subscription {
	sub, err := entities.NewSubscription(b.userID, b.datasetID, b.createdAt)
	if err != nil {
		panic(err)
	}
	if b.id != 0 {
		sub.ID = b.id
	}
	return sub
}

type AccessRequestBuilder struct {
	id        uint64
	datasetID uint64
	userID    uint64
	status    entities.AccessStatus
	createdAt time.Time
}

func NewAccessRequestBuilder() *AccessRequestBuilder {
	return &AccessRequestBuilder{
		datasetID: 1,
		userID:    1,
		status:    entities.AccessStatusPending,
		createdAt: time.Now().Add(-30 * time.Minute),
	}
}

func (b *AccessRequestBuilder) WithID(id uint64) *AccessRequestBuilder {
	b.id = id
	return b
}

func (b *AccessRequestBuilder) WithDatasetID(id uint64) *AccessRequestBuilder {
	b.datasetID = id
	return b
}

func (b *AccessRequestBuilder) WithUserID(id uint64) *AccessRequestBuilder {
	b.userID = id
	return b
}

func (b *AccessRequestBuilder) WithStatus(status entities.AccessStatus) *AccessRequestBuilder {
	b.status = status
	return b
}

func (b *AccessRequestBuilder) WithCreatedAt(t time.Time) *AccessRequestBuilder {
	b.createdAt = t
	return b
}

func (b *AccessRequestBuilder) Build() *entities.AccessRequest {
	request, err := entities.NewAccessRequest(b.datasetID, b.userID)
	if err != nil {
		panic(err)
	}
	if b.id != 0 {
		request.ID = b.id
	}
	if !b.createdAt.IsZero() {
		request.CreatedAt = b.createdAt
	}
	if b.status != "" {
		request.Status = b.status
	}
	return request
}

type UserBuilder struct {
	id       uint64
	username string
	email    string
	password string
	country  string
	role     string
	now      time.Time
	blocked  bool
}

func NewUserBuilder() *UserBuilder {
	return &UserBuilder{
		username: randomString("user"),
		email:    randomString("user") + "@example.com",
		password: "password123",
		country:  "RU",
		role:     entities.RoleUser,
		now:      time.Now().UTC(),
	}
}

func (b *UserBuilder) WithID(id uint64) *UserBuilder {
	b.id = id
	return b
}

func (b *UserBuilder) WithUsername(username string) *UserBuilder {
	b.username = username
	return b
}

func (b *UserBuilder) WithEmail(email string) *UserBuilder {
	b.email = email
	return b
}

func (b *UserBuilder) WithPassword(password string) *UserBuilder {
	b.password = password
	return b
}

func (b *UserBuilder) WithCountry(country string) *UserBuilder {
	b.country = country
	return b
}

func (b *UserBuilder) WithRole(role string) *UserBuilder {
	b.role = role
	return b
}

func (b *UserBuilder) WithRegistrationDate(t time.Time) *UserBuilder {
	b.now = t
	return b
}

func (b *UserBuilder) Blocked(blocked bool) *UserBuilder {
	b.blocked = blocked
	return b
}

func (b *UserBuilder) Build() *entities.User {
	user, err := entities.NewUser(b.username, b.email, b.password, b.country, b.role, b.now)
	if err != nil {
		panic(err)
	}
	if b.id != 0 {
		user.ID = b.id
	}
	user.IsBlocked = b.blocked
	return user
}

func randomString(prefix string) string {
	const letters = "abcdefghijklmnopqrstuvwxyz"
	var sb strings.Builder
	sb.WriteString(prefix)
	sb.WriteRune('-')
	for i := 0; i < 6; i++ {
		sb.WriteByte(letters[rand.Intn(len(letters))])
	}
	return sb.String()
}

type CreateDatasetCmdBuilder struct {
	actorID     uint64
	name        string
	description string
	categoryID  uint64
	fileName    string
	isPublic    bool
	metaFormat  string
	metaTags    string
	metaSize    uint64
}

func NewCreateDatasetCmdBuilder() *CreateDatasetCmdBuilder {
	return &CreateDatasetCmdBuilder{
		actorID:     1,
		name:        randomString("dataset"),
		description: "auto-generated dataset",
		categoryID:  1,
		fileName:    "dataset.bin",
		isPublic:    true,
	}
}

func (b *CreateDatasetCmdBuilder) WithActor(id uint64) *CreateDatasetCmdBuilder {
	b.actorID = id
	return b
}

func (b *CreateDatasetCmdBuilder) WithName(name string) *CreateDatasetCmdBuilder {
	b.name = name
	return b
}

func (b *CreateDatasetCmdBuilder) WithDescription(description string) *CreateDatasetCmdBuilder {
	b.description = description
	return b
}

func (b *CreateDatasetCmdBuilder) WithCategory(id uint64) *CreateDatasetCmdBuilder {
	b.categoryID = id
	return b
}

func (b *CreateDatasetCmdBuilder) WithFileName(fileName string) *CreateDatasetCmdBuilder {
	b.fileName = fileName
	return b
}

func (b *CreateDatasetCmdBuilder) Public(isPublic bool) *CreateDatasetCmdBuilder {
	b.isPublic = isPublic
	return b
}

func (b *CreateDatasetCmdBuilder) WithMetadata(format, tags string, size uint64) *CreateDatasetCmdBuilder {
	b.metaFormat = format
	b.metaTags = tags
	b.metaSize = size
	return b
}

func (b *CreateDatasetCmdBuilder) WithoutMetadata() *CreateDatasetCmdBuilder {
	b.metaFormat = ""
	b.metaTags = ""
	b.metaSize = 0
	return b
}

func (b *CreateDatasetCmdBuilder) WithoutName() *CreateDatasetCmdBuilder {
	b.name = "   "
	return b
}

func (b *CreateDatasetCmdBuilder) WithoutCategory() *CreateDatasetCmdBuilder {
	b.categoryID = 0
	return b
}

func (b *CreateDatasetCmdBuilder) Build() services.CreateDatasetCmd {
	return services.CreateDatasetCmd{
		ActorID:     b.actorID,
		Name:        b.name,
		Description: b.description,
		CategoryID:  b.categoryID,
		FileName:    b.fileName,
		IsPublic:    b.isPublic,
		MetaFormat:  b.metaFormat,
		MetaTags:    b.metaTags,
		MetaSize:    b.metaSize,
	}
}

type AddVersionCmdBuilder struct {
	actorID    uint64
	datasetID  uint64
	changeLog  string
	fileName   string
	metaFormat string
	metaTags   string
	metaSize   uint64
}

func NewAddVersionCmdBuilder() *AddVersionCmdBuilder {
	return &AddVersionCmdBuilder{
		actorID:   1,
		datasetID: 1,
		changeLog: "minor update",
		fileName:  "dataset.bin",
	}
}

func (b *AddVersionCmdBuilder) WithActor(id uint64) *AddVersionCmdBuilder {
	b.actorID = id
	return b
}

func (b *AddVersionCmdBuilder) WithDataset(id uint64) *AddVersionCmdBuilder {
	b.datasetID = id
	return b
}

func (b *AddVersionCmdBuilder) WithChangeLog(changeLog string) *AddVersionCmdBuilder {
	b.changeLog = changeLog
	return b
}

func (b *AddVersionCmdBuilder) WithFileName(name string) *AddVersionCmdBuilder {
	b.fileName = name
	return b
}

func (b *AddVersionCmdBuilder) WithMetadata(format, tags string, size uint64) *AddVersionCmdBuilder {
	b.metaFormat = format
	b.metaTags = tags
	b.metaSize = size
	return b
}

func (b *AddVersionCmdBuilder) WithoutMetadata() *AddVersionCmdBuilder {
	b.metaFormat = ""
	b.metaTags = ""
	b.metaSize = 0
	return b
}

func (b *AddVersionCmdBuilder) WithoutDataset() *AddVersionCmdBuilder {
	b.datasetID = 0
	return b
}

func (b *AddVersionCmdBuilder) Build() services.AddVersionCmd {
	return services.AddVersionCmd{
		ActorID:    b.actorID,
		DatasetID:  b.datasetID,
		ChangeLog:  b.changeLog,
		FileName:   b.fileName,
		MetaFormat: b.metaFormat,
		MetaTags:   b.metaTags,
		MetaSize:   b.metaSize,
	}
}

type UpdateDatasetCmdBuilder struct {
	id          uint64
	name        string
	description string
	categoryID  uint64
	isPublic    bool
}

func NewUpdateDatasetCmdBuilder() *UpdateDatasetCmdBuilder {
	return &UpdateDatasetCmdBuilder{
		id:          1,
		name:        randomString("dataset"),
		description: "updated dataset description",
		categoryID:  1,
		isPublic:    true,
	}
}

func (b *UpdateDatasetCmdBuilder) WithID(id uint64) *UpdateDatasetCmdBuilder {
	b.id = id
	return b
}

func (b *UpdateDatasetCmdBuilder) WithName(name string) *UpdateDatasetCmdBuilder {
	b.name = name
	return b
}

func (b *UpdateDatasetCmdBuilder) WithDescription(description string) *UpdateDatasetCmdBuilder {
	b.description = description
	return b
}

func (b *UpdateDatasetCmdBuilder) WithCategory(id uint64) *UpdateDatasetCmdBuilder {
	b.categoryID = id
	return b
}

func (b *UpdateDatasetCmdBuilder) Public(isPublic bool) *UpdateDatasetCmdBuilder {
	b.isPublic = isPublic
	return b
}

func (b *UpdateDatasetCmdBuilder) WithoutName() *UpdateDatasetCmdBuilder {
	b.name = ""
	return b
}

func (b *UpdateDatasetCmdBuilder) Build() services.UpdateDatasetCmd {
	return services.UpdateDatasetCmd{
		ID:          b.id,
		Name:        b.name,
		Description: b.description,
		CategoryID:  b.categoryID,
		IsPublic:    b.isPublic,
	}
}
