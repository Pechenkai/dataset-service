package v2

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"ppo/internal/delivery/http/dto"
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

	var req struct {
		Name        *string `json:"name"`
		Description *string `json:"description"`
	}
	if err := decodeJSON(r, &req); err != nil {
		return err
	}
	if req.Name == nil && req.Description == nil {
		return &dto.BadRequestError{Message: "nothing to update"}
	}

	current, err := h.categories.GetCategoryByID(r.Context(), id)
	if err != nil {
		return err
	}

	name := current.Name
	desc := current.Description
	if req.Name != nil {
		name = *req.Name
	}
	if req.Description != nil {
		desc = *req.Description
	}

	err = h.categories.UpdateCategory(r.Context(), services.UpdateCategoryCmd{
		ID:          id,
		Name:        name,
		Description: desc,
	})
	if err != nil {
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
