package dto

import (
	"ppo/internal/entities"
	"ppo/internal/services"
)

type CreateCategoryRequest struct {
	Name        string `json:"name" validate:"required,max=50"`
	Description string `json:"description"`
}

func (r *CreateCategoryRequest) ToCommand() services.CreateCategoryCmd {
	return services.CreateCategoryCmd{
		Name:        r.Name,
		Description: r.Description,
	}
}

type UpdateCategoryRequest struct {
	Name        string `json:"name" validate:"required,max=50"`
	Description string `json:"description"`
}

func (r *UpdateCategoryRequest) ToCommand(id uint64) services.UpdateCategoryCmd {
	return services.UpdateCategoryCmd{
		ID:          id,
		Name:        r.Name,
		Description: r.Description,
	}
}

type CategoryResponse struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

func FromEntity(e *entities.Category) CategoryResponse {
	return CategoryResponse{
		ID:          e.ID,
		Name:        e.Name,
		Description: e.Description,
	}
}

type CategoriesResponse struct {
	Categories []CategoryResponse `json:"categories"`
}

func FromEntities(list []*entities.Category) CategoriesResponse {
	resp := make([]CategoryResponse, len(list))
	for i, e := range list {
		resp[i] = FromEntity(e)
	}
	return CategoriesResponse{Categories: resp}
}
