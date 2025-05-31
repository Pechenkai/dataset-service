package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"ppo/internal/delivery/http/dto"
	"ppo/internal/services"
)

// @Summary      List categories
// @Description  Returns all categories
// @Tags         categories
// @Produce      json
// @Success      200  {array}   dto.CategoryResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /categories [get]
func ListCategories(svc services.CategoryService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cats, err := svc.ListCategories(r.Context())
		if err != nil {
			dto.WriteError(w, err)
			return
		}
		resp := dto.FromEntities(cats)
		dto.WriteJSON(w, http.StatusOK, resp)
	}
}

// @Summary      Create category
// @Description  Create a new category
// @Tags         categories
// @Accept       json
// @Produce      json
// @Param        body  body      dto.CreateCategoryRequest  true  "Name and description"
// @Success      201   {object}  dto.CategoryResponse
// @Failure      400   {object}  dto.ErrorResponse
// @Failure      409   {object}  dto.ErrorResponse  "category exists"
// @Router       /categories [post]
func CreateCategory(svc services.CategoryService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.CreateCategoryRequest

		if err := dto.DecodeJSON(r.Body, &req); err != nil {
			dto.WriteError(w, err)
			return
		}

		cmd := req.ToCommand()

		id, err := svc.CreateCategory(r.Context(), cmd)
		if err != nil {
			dto.WriteError(w, err)
			return
		}

		cat, err := svc.GetCategoryByID(r.Context(), id)
		if err != nil {
			dto.WriteError(w, err)
			return
		}

		dto.WriteJSON(w, http.StatusCreated, dto.FromEntity(cat))
	}
}

// @Summary      Get category by ID
// @Description  Returns a category by its ID
// @Tags         categories
// @Produce      json
// @Param        id   path      int  true  "Category ID"
// @Success      200  {object}  dto.CategoryResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Router       /categories/{id} [get]
func GetCategory(svc services.CategoryService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "id")
		id, err := strconv.ParseUint(idParam, 10, 64)
		if err != nil {
			dto.WriteError(w, err)
			return
		}
		cat, err := svc.GetCategoryByID(r.Context(), id)
		if err != nil {
			dto.WriteError(w, err)
			return
		}
		dto.WriteJSON(w, http.StatusOK, dto.FromEntity(cat))
	}
}

// @Summary      Update category
// @Description  Update a category by its ID
// @Tags         categories
// @Accept       json
// @Produce      json
// @Param        id    path      int                        true  "Category ID"
// @Param        body  body      dto.UpdateCategoryRequest  true  "Updated name and description"
// @Success      204  {object}  nil
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Router       /categories/{id} [put]
func UpdateCategory(svc services.CategoryService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "id")
		id, err := strconv.ParseUint(idParam, 10, 64)
		if err != nil {
			dto.WriteError(w, err)
			return
		}

		var req dto.UpdateCategoryRequest
		if err := dto.DecodeJSON(r.Body, &req); err != nil {
			dto.WriteError(w, err)
			return
		}

		cmd := req.ToCommand(id)

		cmd.ID = id
		if err := svc.UpdateCategory(r.Context(), cmd); err != nil {
			dto.WriteError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}

// @Summary      Delete category
// @Description  Delete a category by its ID
// @Tags         categories
// @Produce      json
// @Param        id   path      int  true  "Category ID"
// @Success      204  {object}  nil
// @Failure      404  {object}  dto.ErrorResponse
// @Router       /categories/{id} [delete]
func DeleteCategory(svc services.CategoryService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "id")
		id, err := strconv.ParseUint(idParam, 10, 64)
		if err != nil {
			dto.WriteError(w, err)
			return
		}
		if err := svc.DeleteCategory(r.Context(), id); err != nil {
			dto.WriteError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
