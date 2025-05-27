package dto

import "ppo/internal/services"

// CreateCategoryRequest — payload для POST /api/v1/categories
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

// UpdateCategoryRequest — payload для PUT /api/v1/categories/{id}
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

// CategoryResponse — модель в ответе GET /api/v1/categories/{id}
type CategoryResponse struct {
	ID          uint64 `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// CategoriesResponse — список для GET /api/v1/categories
type CategoriesResponse struct {
	Categories []CategoryResponse `json:"categories"`
}
