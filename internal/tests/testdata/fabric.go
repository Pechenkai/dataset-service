package testdata

import (
	"bytes"
	"fmt"
	"io"
	"time"

	"ppo/internal/entities"
	"ppo/internal/services"
)

type Fabric struct {
	now    time.Time
	nextID uint64
}

func NewFabric() *Fabric {
	return &Fabric{
		now:    time.Now().UTC(),
		nextID: 1000,
	}
}

func NewFabricAt(t time.Time) *Fabric {
	return &Fabric{
		now:    t.UTC(),
		nextID: 1000,
	}
}

func (f *Fabric) advanceID() uint64 {
	f.nextID++
	return f.nextID
}

func (f *Fabric) Category() *entities.Category {
	return NewCategoryBuilder().
		WithID(f.advanceID()).
		WithName(fmt.Sprintf("category-%d", f.nextID)).
		WithDescription("generated category").
		Build()
}

func (f *Fabric) CreateCategoryCommand() services.CreateCategoryCmd {
	return NewCreateCategoryCmdBuilder().Build()
}

func (f *Fabric) InvalidCreateCategoryCommand() services.CreateCategoryCmd {
	return NewCreateCategoryCmdBuilder().WithoutName().Build()
}

func (f *Fabric) UpdateCategoryCommand(cat *entities.Category) services.UpdateCategoryCmd {
	return NewUpdateCategoryCmdBuilder().
		WithID(cat.ID).
		WithName(cat.Name + " updated").
		WithDescription(cat.Description + " updated").
		Build()
}

func (f *Fabric) InvalidUpdateCategoryCommand(cat *entities.Category) services.UpdateCategoryCmd {
	return NewUpdateCategoryCmdBuilder().
		WithID(cat.ID).
		WithoutName().
		Build()
}

func (f *Fabric) NotifySubscribersCommand(datasetID uint64, message string) services.NotifySubscribersCmd {
	return NewNotifySubscribersCmdBuilder().
		WithDataset(datasetID).
		WithMessage(message).
		Build()
}

func (f *Fabric) InvalidNotifySubscribersCommand(datasetID uint64) services.NotifySubscribersCmd {
	return NewNotifySubscribersCmdBuilder().
		WithDataset(datasetID).
		WithoutMessage().
		Build()
}

func (f *Fabric) CreateReviewCommand(user *entities.User, dataset *entities.Dataset, rating entities.Rating) services.CreateReviewCmd {
	return NewCreateReviewCmdBuilder().
		WithUser(user.ID).
		WithDataset(dataset.ID).
		WithRating(rating).
		WithText("great dataset").
		Build()
}

func (f *Fabric) InvalidCreateReviewCommand(user *entities.User, dataset *entities.Dataset) services.CreateReviewCmd {
	return NewCreateReviewCmdBuilder().
		WithUser(user.ID).
		WithDataset(dataset.ID).
		WithRating(entities.Rating(0)).
		Build()
}

func (f *Fabric) UpdateReviewCommand(review *entities.Review, rating entities.Rating) services.UpdateReviewCmd {
	return NewUpdateReviewCmdBuilder().
		WithID(review.ID).
		WithRating(rating).
		WithText(review.Text + " updated").
		Build()
}

func (f *Fabric) RegisterUserCommand() services.RegisterUserCmd {
	return NewRegisterUserCmdBuilder().Build()
}

func (f *Fabric) InvalidRegisterUserCommand() services.RegisterUserCmd {
	return NewRegisterUserCmdBuilder().WithoutEmail().Build()
}

func (f *Fabric) AuthenticateUserCommand(email, password string) services.AuthenticateUserCmd {
	return NewAuthenticateUserCmdBuilder().WithEmail(email).WithPassword(password).Build()
}

func (f *Fabric) UpdateUserCommand(user *entities.User) services.UpdateUserCmd {
	return NewUpdateUserCmdBuilder().
		WithID(user.ID).
		WithUsername(user.Username + "_updated").
		WithEmail(user.Email).
		WithCountry("US").
		Blocked(false).
		WithRole(user.Role).
		Build()
}

func (f *Fabric) RegularUser() *entities.User {
	return NewUserBuilder().
		WithID(f.advanceID()).
		WithRegistrationDate(f.now.Add(-24 * time.Hour)).
		Build()
}

func (f *Fabric) AdminUser() *entities.User {
	return NewUserBuilder().
		WithID(f.advanceID()).
		WithRole(entities.RoleAdmin).
		WithRegistrationDate(f.now.Add(-48 * time.Hour)).
		Build()
}

func (f *Fabric) Dataset(owner *entities.User, category *entities.Category) *entities.Dataset {
	return NewDatasetBuilder().
		WithID(f.advanceID()).
		WithOwner(owner.ID).
		WithCategory(category.ID).
		WithCreatedAt(f.now.Add(-2 * time.Hour)).
		Build()
}

func (f *Fabric) DatasetVersion(dataset *entities.Dataset, number string) *entities.DatasetVersion {
	if number == "" {
		number = "v1.0.0"
	}
	return NewDatasetVersionBuilder().
		WithID(f.advanceID()).
		WithDatasetID(dataset.ID).
		WithNumber(number).
		WithFilepath(fmt.Sprintf("datasets/%d/%s.bin", dataset.ID, number)).
		WithUploadDate(f.now.Add(-time.Hour)).
		Build()
}

func (f *Fabric) Metadata(version *entities.DatasetVersion) *entities.Metadata {
	return NewMetadataBuilder().
		WithID(f.advanceID()).
		WithVersionID(version.ID).
		Build()
}

func (f *Fabric) Notification(user *entities.User, dataset *entities.Dataset) *entities.Notification {
	return NewNotificationBuilder().
		WithID(f.advanceID()).
		WithUserID(user.ID).
		WithDatasetID(dataset.ID).
		WithMessage(fmt.Sprintf("dataset %d updated", dataset.ID)).
		WithCreatedAt(f.now.Add(-30 * time.Minute)).
		Build()
}

func (f *Fabric) Subscription(user *entities.User, dataset *entities.Dataset) *entities.Subscription {
	return NewSubscriptionBuilder().
		WithID(f.advanceID()).
		WithUserID(user.ID).
		WithDatasetID(dataset.ID).
		WithCreatedAt(f.now.Add(-12 * time.Hour)).
		Build()
}

func (f *Fabric) ReviewPositive(user *entities.User, dataset *entities.Dataset) *entities.Review {
	return NewReviewBuilder().
		WithID(f.advanceID()).
		WithUserID(user.ID).
		WithDatasetID(dataset.ID).
		WithRating(entities.Rating5).
		WithCreatedAt(f.now.Add(-3 * time.Hour)).
		WithText("excellent dataset").
		Build()
}

func (f *Fabric) ReviewNegative(user *entities.User, dataset *entities.Dataset) *entities.Review {
	return NewReviewBuilder().
		WithID(f.advanceID()).
		WithUserID(user.ID).
		WithDatasetID(dataset.ID).
		WithRating(entities.Rating1).
		WithCreatedAt(f.now.Add(-2 * time.Hour)).
		WithText("needs improvement").
		Build()
}

func (f *Fabric) AccessRequestPending(user *entities.User, dataset *entities.Dataset) *entities.AccessRequest {
	return NewAccessRequestBuilder().
		WithID(f.advanceID()).
		WithUserID(user.ID).
		WithDatasetID(dataset.ID).
		WithStatus(entities.AccessStatusPending).
		WithCreatedAt(f.now.Add(-time.Hour)).
		Build()
}

func (f *Fabric) AccessRequestApproved(user *entities.User, dataset *entities.Dataset) *entities.AccessRequest {
	return NewAccessRequestBuilder().
		WithID(f.advanceID()).
		WithUserID(user.ID).
		WithDatasetID(dataset.ID).
		WithStatus(entities.AccessStatusApproved).
		WithCreatedAt(f.now.Add(-45 * time.Minute)).
		Build()
}

func (f *Fabric) CreateDatasetCommand(category *entities.Category, actor *entities.User) services.CreateDatasetCmd {
	return NewCreateDatasetCmdBuilder().
		WithActor(actor.ID).
		WithCategory(category.ID).
		WithFileName("dataset.bin").
		WithMetadata("csv", "auto,generated", 2048).
		Build()
}

func (f *Fabric) InvalidCreateDatasetCommand(category *entities.Category, actor *entities.User) services.CreateDatasetCmd {
	return NewCreateDatasetCmdBuilder().
		WithActor(actor.ID).
		WithCategory(category.ID).
		WithoutName().
		Build()
}

func (f *Fabric) AddVersionCommand(dataset *entities.Dataset, actor *entities.User) services.AddVersionCmd {
	return NewAddVersionCmdBuilder().
		WithActor(actor.ID).
		WithDataset(dataset.ID).
		WithFileName("dataset-v2.bin").
		WithChangeLog("new samples added").
		WithMetadata("json", "delta", 4096).
		Build()
}

func (f *Fabric) InvalidAddVersionCommand(dataset *entities.Dataset, actor *entities.User) services.AddVersionCmd {
	return NewAddVersionCmdBuilder().
		WithActor(actor.ID).
		WithFileName(fmt.Sprintf("dataset-%d-v2.bin", dataset.ID)).
		WithoutDataset().
		Build()
}

func (f *Fabric) UpdateDatasetCommand(dataset *entities.Dataset, category *entities.Category) services.UpdateDatasetCmd {
	return NewUpdateDatasetCmdBuilder().
		WithID(dataset.ID).
		WithCategory(category.ID).
		WithName(dataset.Name + " updated").
		WithDescription(dataset.Description + " updated").
		Public(true).
		Build()
}

func (f *Fabric) InvalidUpdateDatasetCommand(dataset *entities.Dataset, category *entities.Category) services.UpdateDatasetCmd {
	return NewUpdateDatasetCmdBuilder().
		WithID(dataset.ID).
		WithCategory(category.ID).
		WithoutName().
		Build()
}

func (f *Fabric) DatasetFixtureWithMetadata() DatasetFixture {
	owner := f.RegularUser()
	category := f.Category()
	dataset := f.Dataset(owner, category)
	version := f.DatasetVersion(dataset, "v1.0.0")
	metadata := f.Metadata(version)

	return DatasetFixture{
		Owner:    owner,
		Category: category,
		Dataset:  dataset,
		Version:  version,
		Metadata: metadata,
	}
}

func (f *Fabric) DatasetFixture() DatasetFixture {
	fixture := f.DatasetFixtureWithMetadata()
	fixture.Metadata = nil
	return fixture
}

func (f *Fabric) DatasetFile(payload string) (io.Reader, int64) {
	if payload == "" {
		payload = "dataset payload"
	}
	data := []byte(payload)
	return bytes.NewReader(data), int64(len(data))
}

func (f *Fabric) DatasetVersionFile(payload string) (io.Reader, int64) {
	if payload == "" {
		payload = "version payload"
	}
	data := []byte(payload)
	return bytes.NewReader(data), int64(len(data))
}

type DatasetFixture struct {
	Owner    *entities.User
	Category *entities.Category
	Dataset  *entities.Dataset
	Version  *entities.DatasetVersion
	Metadata *entities.Metadata
}

func (d DatasetFixture) WithoutMetadata() DatasetFixture {
	d.Metadata = nil
	return d
}

func (d DatasetFixture) WithoutVersion() DatasetFixture {
	d.Version = nil
	d.Metadata = nil
	return d
}
