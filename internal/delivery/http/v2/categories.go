package v2

import (
	"net/http"
	"strconv"
	"strings"

	chi "github.com/go-chi/chi/v5"

	"ppo/internal/delivery/http/dto"
	"ppo/internal/entities"
	"ppo/internal/services"
)

func (h *Handler) ListCategories(w http.ResponseWriter, r *http.Request) error {
	limit, offset, err := parsePagination(r, 50, 200)
	if err != nil {
		return err
	}

	search := strings.ToLower(strings.TrimSpace(r.URL.Query().Get("search")))

	categories, err := h.categories.ListCategories(r.Context())
	if err != nil {
		return err
	}

	filtered := make([]CategoryResponse, 0, len(categories))
	for _, cat := range categories {
		if search != "" && !strings.Contains(strings.ToLower(cat.Name), search) && !strings.Contains(strings.ToLower(cat.Description), search) {
			continue
		}
		filtered = append(filtered, CategoryResponse{
			ID:          cat.ID,
			Name:        cat.Name,
			Description: cat.Description,
		})
	}

	total := len(filtered)
	start := clamp(offset, 0, total)
	end := clamp(offset+limit, start, total)

	resp := CategoriesResponse{
		Items: filtered[start:end],
		Meta: PaginationMeta{
			Total:  total,
			Limit:  limit,
			Offset: offset,
		},
	}

	writeJSON(w, http.StatusOK, resp)
	return nil
}

func (h *Handler) CreateCategory(w http.ResponseWriter, r *http.Request) error {
	if _, err := requireAdmin(r.Context()); err != nil {
		return err
	}

	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := decodeJSON(r, &req); err != nil {
		return err
	}
	if strings.TrimSpace(req.Name) == "" {
		return &dto.BadRequestError{Message: "name is required"}
	}

	cmd := services.CreateCategoryCmd{
		Name:        req.Name,
		Description: req.Description,
	}
	id, err := h.categories.CreateCategory(r.Context(), cmd)
	if err != nil {
		return err
	}

	cat, err := h.categories.GetCategoryByID(r.Context(), id)
	if err != nil {
		return err
	}

	writeJSON(w, http.StatusCreated, CategoryResponse{
		ID:          cat.ID,
		Name:        cat.Name,
		Description: cat.Description,
	})
	return nil
}

func (h *Handler) GetCategory(w http.ResponseWriter, r *http.Request) error {
	id, err := parseIDParam(chi.URLParam(r, "categoryId"))
	if err != nil {
		return err
	}
	cat, err := h.categories.GetCategoryByID(r.Context(), id)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, CategoryResponse{
		ID:          cat.ID,
		Name:        cat.Name,
		Description: cat.Description,
	})
	return nil
}

func (h *Handler) UpdateCategory(w http.ResponseWriter, r *http.Request) error {
	id, err := parseIDParam(chi.URLParam(r, "categoryId"))
	if err != nil {
		return err
	}

	if _, err := requireAdmin(r.Context()); err != nil {
		return err
	}

	req, err := decodeUpdateCategoryRequest(r)
	if err != nil {
		return err
	}

	current, err := h.categories.GetCategoryByID(r.Context(), id)
	if err != nil {
		return err
	}

	cmd := buildUpdateCategoryCmd(id, current, req)
	if err := h.categories.UpdateCategory(r.Context(), cmd); err != nil {
		return err
	}

	updated, err := h.categories.GetCategoryByID(r.Context(), id)
	if err != nil {
		return err
	}

	writeJSON(w, http.StatusOK, CategoryResponse{
		ID:          updated.ID,
		Name:        updated.Name,
		Description: updated.Description,
	})
	return nil
}

func (h *Handler) DeleteCategory(w http.ResponseWriter, r *http.Request) error {
	id, err := parseIDParam(chi.URLParam(r, "categoryId"))
	if err != nil {
		return err
	}
	if _, err := requireAdmin(r.Context()); err != nil {
		return err
	}

	if err := h.categories.DeleteCategory(r.Context(), id); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func parseIDParam(value string) (uint64, error) {
	id, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return 0, &dto.BadRequestError{Message: "invalid identifier"}
	}
	return id, nil
}

type updateCategoryRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

func decodeUpdateCategoryRequest(r *http.Request) (updateCategoryRequest, error) {
	var req updateCategoryRequest
	if err := decodeJSON(r, &req); err != nil {
		return updateCategoryRequest{}, err
	}
	if req.Name == nil && req.Description == nil {
		return updateCategoryRequest{}, &dto.BadRequestError{Message: "nothing to update"}
	}
	return req, nil
}

func buildUpdateCategoryCmd(id uint64, current *entities.Category, req updateCategoryRequest) services.UpdateCategoryCmd {
	cmd := services.UpdateCategoryCmd{
		ID:          id,
		Name:        current.Name,
		Description: current.Description,
	}
	if req.Name != nil {
		cmd.Name = *req.Name
	}
	if req.Description != nil {
		cmd.Description = *req.Description
	}
	return cmd
}
