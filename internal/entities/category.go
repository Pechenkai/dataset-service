package entities

import (
	"strings"
)

type Category struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func NewCategory(name, description string) (*Category, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrEmptyCategoryName
	}
	if len(name) > 50 {
		return nil, ErrCategoryNameTooLong
	}

	return &Category{
		Name:        name,
		Description: strings.TrimSpace(description),
	}, nil
}
