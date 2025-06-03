package dto

import "ppo/internal/entities"

type CategoryDTO struct {
	ID          uint64
	Name        string
	Description string
}

type CreateCategoryForm struct {
	Name        string
	Description string
}

type UpdateCategoryForm struct {
	ID          uint64
	Name        string
	Description string
}

func ToCategoryDTO(c *entities.Category) *CategoryDTO {
	return &CategoryDTO{
		ID:          c.ID,
		Name:        c.Name,
		Description: c.Description,
	}
}

func ToCategoryDTOs(list []*entities.Category) []*CategoryDTO {
	result := make([]*CategoryDTO, 0, len(list))
	for _, c := range list {
		result = append(result, ToCategoryDTO(c))
	}
	return result
}
