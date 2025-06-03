package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"ppo/internal/delivery/web/dto"
	"ppo/internal/delivery/web/templates"
	"ppo/internal/services"
)

// CategoryHandler отвечает за CRUD‐операции с категориями.
type CategoryHandler struct {
	service services.CategoryService
	logger  *zap.Logger
}

// NewCategoryHandler создает новый контроллер.
func NewCategoryHandler(svc services.CategoryService, log *zap.Logger) *CategoryHandler {
	return &CategoryHandler{service: svc, logger: log}
}

// RegisterRoutes регистрирует маршрут для CategoryHandler.
func (h *CategoryHandler) RegisterRoutes(r chi.Router) {
	r.Get("/categories", h.List)
	r.Get("/categories/new", h.NewForm)
	r.Post("/categories", h.Create)
	r.Get("/categories/{id}", h.Show)
	r.Get("/categories/{id}/edit", h.EditForm)
	r.Post("/categories/{id}", h.Update)
	r.Post("/categories/{id}/delete", h.Delete)
}

// List отображает список всех категорий.
// GET /categories
func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	cats, err := h.service.ListCategories(r.Context())
	if err != nil {
		h.logger.Error("ListCategories failed", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Собираем данные для шаблона
	data := struct {
		Title      string
		Categories []*dto.CategoryDTO
	}{
		Title:      "Список категорий",
		Categories: dto.ToCategoryDTOs(cats),
	}

	// Парсим только layout.tmpl + category_list.tmpl
	tpl := template.Must(template.ParseFS(
		templates.TemplatesFS,
		"layout.tmpl",
		"category_list.tmpl",
	))

	if err := tpl.ExecuteTemplate(w, "layout.tmpl", data); err != nil {
		h.logger.Error("template execution error", zap.Error(err))
	}
}

// Show отображает одну категорию.
// GET /categories/{id}
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

	data := struct {
		Title    string
		Category *dto.CategoryDTO
	}{
		Title:    fmt.Sprintf("Категория #%d", cat.ID),
		Category: dto.ToCategoryDTO(cat),
	}

	// Парсим только layout.tmpl + category_view.tmpl
	tpl := template.Must(template.ParseFS(
		templates.TemplatesFS,
		"layout.tmpl",
		"category_show.tmpl",
	))

	if err := tpl.ExecuteTemplate(w, "layout.tmpl", data); err != nil {
		h.logger.Error("template execution error", zap.Error(err))
	}
}

// NewForm показывает форму создания новой категории.
// GET /categories/new
func (h *CategoryHandler) NewForm(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Title      string
		IsNew      bool
		FormAction string
		Form       *dto.CreateCategoryForm
	}{
		Title:      "Новая категория",
		IsNew:      true,
		FormAction: "/categories",
		Form:       &dto.CreateCategoryForm{},
	}

	// Парсим только layout.tmpl + category_form.tmpl
	tpl := template.Must(template.ParseFS(
		templates.TemplatesFS,
		"layout.tmpl",
		"category_form.tmpl",
	))

	if err := tpl.ExecuteTemplate(w, "layout.tmpl", data); err != nil {
		h.logger.Error("template execution error", zap.Error(err))
	}
}

// Create обрабатывает сохранение новой категории.
// POST /categories
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

// EditForm показывает форму редактирования.
// GET /categories/{id}/edit
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

	data := struct {
		Title      string
		IsNew      bool
		FormAction string
		Form       *dto.UpdateCategoryForm
	}{
		Title:      fmt.Sprintf("Редактирование категории #%d", id),
		IsNew:      false,
		FormAction: fmt.Sprintf("/categories/%d", id),
		Form: &dto.UpdateCategoryForm{
			ID:          cat.ID,
			Name:        cat.Name,
			Description: cat.Description,
		},
	}

	// Парсим только layout.tmpl + category_form.tmpl
	tpl := template.Must(template.ParseFS(
		templates.TemplatesFS,
		"layout.tmpl",
		"category_form.tmpl",
	))

	if err := tpl.ExecuteTemplate(w, "layout.tmpl", data); err != nil {
		h.logger.Error("template execution error", zap.Error(err))
	}
}

// Update обрабатывает обновление.
// POST /categories/{id}
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

// Delete обрабатывает удаление.
// POST /categories/{id}/delete
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
