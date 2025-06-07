package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"ppo/internal/delivery/web/middleware"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"ppo/internal/delivery/web/dto"
	"ppo/internal/delivery/web/templates"
	"ppo/internal/services"
)

type CategoryHandler struct {
	service    services.CategoryService
	datasetSvc services.DatasetService
	logger     *zap.Logger
}

func NewCategoryHandler(svc services.CategoryService, dssvc services.DatasetService, log *zap.Logger) *CategoryHandler {
	return &CategoryHandler{service: svc, logger: log, datasetSvc: dssvc}
}

func (h *CategoryHandler) RegisterRoutes(r chi.Router) {
	r.Get("/categories", h.List)
	r.Get("/categories/{id}", h.Show)

	r.With(middleware.RequireRole("admin")).Get("/categories/new", h.NewForm)
	r.With(middleware.RequireRole("admin")).Post("/categories", h.Create)

	r.With(middleware.RequireRole("admin")).Get("/categories/{id}/edit", h.EditForm)
	r.With(middleware.RequireRole("admin")).Post("/categories/{id}", h.Update)

	r.With(middleware.RequireRole("admin")).Post("/categories/{id}/delete", h.Delete)
}

func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	cats, err := h.service.ListCategories(r.Context())
	if err != nil {
		h.logger.Error("ListCategories failed", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	currentUID, currentRole := middleware.FromContext(r.Context())

	data := struct {
		Title      string
		Role       string
		UserID     uint64
		Categories []*dto.CategoryDTO
	}{
		Title:      "Список категорий",
		Role:       currentRole,
		UserID:     currentUID,
		Categories: dto.ToCategoryDTOs(cats),
	}

	tpl := template.Must(template.ParseFS(
		templates.TemplatesFS,
		"layout.tmpl",
		"category_list.tmpl",
	))

	if err := tpl.ExecuteTemplate(w, "layout.tmpl", data); err != nil {
		h.logger.Error("template execution error", zap.Error(err))
	}
}

func (h *CategoryHandler) Show(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	cat, err := h.service.GetCategoryByID(r.Context(), id)
	if err != nil {
		h.logger.Warn("GetCategoryByID failed", zap.Uint64("id", id), zap.Error(err))
		http.NotFound(w, r)
		return
	}

	dsets, err := h.datasetSvc.ListByCategory(r.Context(), id)
	if err != nil {
		h.logger.Warn("ListByCategory failed", zap.Uint64("catID", id), zap.Error(err))
	}

	currentUID, currentRole := middleware.FromContext(r.Context())

	data := struct {
		Title    string
		Category *dto.CategoryDTO
		Datasets []*dto.DatasetDTO
		Role     string
		UserID   uint64
	}{
		Title:    fmt.Sprintf("Категория #%d", cat.ID),
		Category: dto.ToCategoryDTO(cat),
		Datasets: dto.ToDatasetDTOs(dsets),
		Role:     currentRole,
		UserID:   currentUID,
	}

	tpl := template.Must(template.ParseFS(
		templates.TemplatesFS,
		"layout.tmpl",
		"category_show.tmpl",
	))

	if err := tpl.ExecuteTemplate(w, "layout.tmpl", data); err != nil {
		h.logger.Error("template execution error", zap.Error(err))
	}
}

func (h *CategoryHandler) NewForm(w http.ResponseWriter, r *http.Request) {
	currentUID, currentRole := middleware.FromContext(r.Context())

	data := struct {
		Title      string
		Role       string
		UserID     uint64
		IsNew      bool
		FormAction string
		Form       *dto.CreateCategoryForm
	}{
		Title:      "Новая категория",
		Role:       currentRole,
		UserID:     currentUID,
		IsNew:      true,
		FormAction: "/categories",
		Form:       &dto.CreateCategoryForm{},
	}

	tpl := template.Must(template.ParseFS(
		templates.TemplatesFS,
		"layout.tmpl",
		"category_form.tmpl",
	))

	if err := tpl.ExecuteTemplate(w, "layout.tmpl", data); err != nil {
		h.logger.Error("template execution error", zap.Error(err))
	}
}

func (h *CategoryHandler) Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.logger.Warn("ParseForm error", zap.Error(err))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	name := r.PostForm.Get("name")
	desc := r.PostForm.Get("description")

	id, err := h.service.CreateCategory(r.Context(), services.CreateCategoryCmd{
		Name:        name,
		Description: desc,
	})
	if err != nil {
		h.logger.Error("CreateCategory failed", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/categories/%d", id), http.StatusSeeOther)
}

func (h *CategoryHandler) EditForm(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	cat, err := h.service.GetCategoryByID(r.Context(), id)
	if err != nil {
		h.logger.Warn("GetCategoryByID failed", zap.Uint64("id", id), zap.Error(err))
		http.NotFound(w, r)
		return
	}

	currentUID, currentRole := middleware.FromContext(r.Context())

	data := struct {
		Title      string
		IsNew      bool
		Role       string
		UserID     uint64
		FormAction string
		Form       *dto.UpdateCategoryForm
	}{
		Title:      fmt.Sprintf("Редактирование категории #%d", id),
		IsNew:      false,
		Role:       currentRole,
		UserID:     currentUID,
		FormAction: fmt.Sprintf("/categories/%d", id),
		Form: &dto.UpdateCategoryForm{
			ID:          cat.ID,
			Name:        cat.Name,
			Description: cat.Description,
		},
	}

	tpl := template.Must(template.ParseFS(
		templates.TemplatesFS,
		"layout.tmpl",
		"category_form.tmpl",
	))

	if err := tpl.ExecuteTemplate(w, "layout.tmpl", data); err != nil {
		h.logger.Error("template execution error", zap.Error(err))
	}
}

func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if err := r.ParseForm(); err != nil {
		h.logger.Warn("ParseForm error", zap.Error(err))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	name := r.PostForm.Get("name")
	desc := r.PostForm.Get("description")

	err = h.service.UpdateCategory(r.Context(), services.UpdateCategoryCmd{
		ID:          id,
		Name:        name,
		Description: desc,
	})
	if err != nil {
		h.logger.Error("UpdateCategory failed", zap.Error(err), zap.Uint64("id", id))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/categories/%d", id), http.StatusSeeOther)
}

func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if err := h.service.DeleteCategory(r.Context(), id); err != nil {
		h.logger.Error("DeleteCategory failed", zap.Error(err), zap.Uint64("id", id))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/categories", http.StatusSeeOther)
}
