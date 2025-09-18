package testdata

import (
    "math/rand"
    "strings"
    "time"

    "ppo/internal/entities"
)

func init() {
    rand.Seed(time.Now().UnixNano())
}

type CategoryBuilder struct {
    name        string
    description string
}

func NewCategoryBuilder() *CategoryBuilder {
    return &CategoryBuilder{
        name:        randomString("category"),
        description: "auto-generated category",
    }
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
    return cat
}

type DatasetBuilder struct {
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

func (b *DatasetBuilder) WithName(name string) *DatasetBuilder {
    b.name = name
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
    return ds
}

type UserBuilder struct {
    username string
    email    string
    password string
    country  string
    role     string
    now      time.Time
}

func NewUserBuilder() *UserBuilder {
    return &UserBuilder{
        username: randomString("user"),
        email:    randomString("user") + "@example.com",
        password: "password123",
        country:  "RU",
        role:     "member",
        now:      time.Now().UTC(),
    }
}

func (b *UserBuilder) WithEmail(email string) *UserBuilder {
    b.email = email
    return b
}

func (b *UserBuilder) WithRole(role string) *UserBuilder {
    b.role = role
    return b
}

func (b *UserBuilder) Build() *entities.User {
    user, err := entities.NewUser(b.username, b.email, b.password, b.country, b.role, b.now)
    if err != nil {
        panic(err)
    }
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

