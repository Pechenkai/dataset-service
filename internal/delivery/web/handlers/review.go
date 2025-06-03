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
	"ppo/internal/entities"
	"ppo/internal/services"
)

// ReviewHandler отвечает за CRUD‐операции с отзывами.
type ReviewHandler struct {
	service services.ReviewService
	logger  *zap.Logger
}

// NewReviewHandler создаёт новый ReviewHandler без предзагрузки шаблонов.
func NewReviewHandler(svc services.ReviewService, log *zap.Logger) *ReviewHandler {
	return &ReviewHandler{
		service: svc,
		logger:  log,
	}
}

// RegisterRoutes регистрирует маршруты для отзывов.
func (h *ReviewHandler) RegisterRoutes(r chi.Router) {
	r.Get("/reviews", h.List)
	r.Get("/reviews/new", h.NewForm)
	r.Post("/reviews", h.Create)
	r.Get("/reviews/{id}/edit", h.EditForm)
	r.Post("/reviews/{id}", h.Update)
	r.Post("/reviews/{id}/delete", h.Delete)
	r.Get("/reviews/summary/{dataset_id}", h.Summary)
}

// List показывает все отзывы (с опциональным фильтром по dataset_id или user_id).
// GET /reviews
func (h *ReviewHandler) List(w http.ResponseWriter, r *http.Request) {
	// Если задан query-параметр dataset_id, показываем только для него
	if dsIDStr := r.URL.Query().Get("dataset_id"); dsIDStr != "" {
		dsID, err := strconv.ParseUint(dsIDStr, 10, 64)
		if err == nil {
			revs, err := h.service.ListByDataset(r.Context(), dsID)
			if err != nil {
				h.logger.Error("ListByDataset failed", zap.Error(err), zap.Uint64("datasetID", dsID))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			// Конвертируем в DTO
			reviewDTOs := dto.ToReviewDTOs(revs)

			// Парсим layout.tmpl + review_list.tmpl
			tpl := template.Must(template.ParseFS(
				templates.TemplatesFS,
				"layout.tmpl",
				"review_list.tmpl",
			))

			data := struct {
				Title           string
				Reviews         []*dto.ReviewDTO
				FilterDatasetID uint64
			}{
				Title:           fmt.Sprintf("Отзывы для датасета #%d", dsID),
				Reviews:         reviewDTOs,
				FilterDatasetID: dsID,
			}

			if err := tpl.ExecuteTemplate(w, "layout.tmpl", data); err != nil {
				h.logger.Error("template execution error", zap.Error(err))
			}
			return
		}
	}

	// Если задан query-параметр user_id, показываем только для пользователя
	if uIDStr := r.URL.Query().Get("user_id"); uIDStr != "" {
		uID, err := strconv.ParseUint(uIDStr, 10, 64)
		if err == nil {
			revs, err := h.service.ListByUser(r.Context(), uID)
			if err != nil {
				h.logger.Error("ListByUser failed", zap.Error(err), zap.Uint64("userID", uID))
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}

			reviewDTOs := dto.ToReviewDTOs(revs)

			tpl := template.Must(template.ParseFS(
				templates.TemplatesFS,
				"layout.tmpl",
				"review_list.tmpl",
			))

			data := struct {
				Title        string
				Reviews      []*dto.ReviewDTO
				FilterUserID uint64
			}{
				Title:        fmt.Sprintf("Отзывы пользователя #%d", uID),
				Reviews:      reviewDTOs,
				FilterUserID: uID,
			}

			if err := tpl.ExecuteTemplate(w, "layout.tmpl", data); err != nil {
				h.logger.Error("template execution error", zap.Error(err))
			}
			return
		}
	}

	// Иначе – выводим все отзывы (или можно запретить без фильтра)
	// Здесь, чтобы вернуть хоть что-то, используем ListByUser(0), но можно завести специальный метод.
	revs, err := h.service.ListByUser(r.Context(), 0)
	if err != nil {
		h.logger.Error("ListByUser(all) failed", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	reviewDTOs := dto.ToReviewDTOs(revs)

	tpl := template.Must(template.ParseFS(
		templates.TemplatesFS,
		"layout.tmpl",
		"review_list.tmpl",
	))

	data := struct {
		Title   string
		Reviews []*dto.ReviewDTO
	}{
		Title:   "Все отзывы",
		Reviews: reviewDTOs,
	}

	if err := tpl.ExecuteTemplate(w, "layout.tmpl", data); err != nil {
		h.logger.Error("template execution error", zap.Error(err))
	}
}

// NewForm отображает форму создания нового отзыва.
// GET /reviews/new
func (h *ReviewHandler) NewForm(w http.ResponseWriter, r *http.Request) {
	formDTO := &dto.CreateReviewForm{}

	tpl := template.Must(template.ParseFS(
		templates.TemplatesFS,
		"layout.tmpl",
		"review_form.tmpl",
	))

	data := struct {
		Title      string
		FormAction string
		Form       *dto.CreateReviewForm
	}{
		Title:      "Новый отзыв",
		FormAction: "/reviews",
		Form:       formDTO,
	}

	if err := tpl.ExecuteTemplate(w, "layout.tmpl", data); err != nil {
		h.logger.Error("template execution error", zap.Error(err))
	}
}

// Create обрабатывает POST /reviews.
func (h *ReviewHandler) Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.logger.Warn("ParseForm error", zap.Error(err))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// Читаем поля формы
	dsID64, err := strconv.ParseUint(r.FormValue("dataset_id"), 10, 64)
	if err != nil {
		h.logger.Warn("invalid dataset_id", zap.Error(err))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	ratingInt, err := strconv.Atoi(r.FormValue("rating"))
	if err != nil || ratingInt < int(entities.Rating1) || ratingInt > int(entities.Rating5) {
		h.logger.Warn("invalid rating", zap.Error(err))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	text := r.FormValue("text")

	// TODO: взять реальный userID из контекста
	userID := uint64(1)

	cmd := services.CreateReviewCmd{
		UserID:    userID,
		DatasetID: dsID64,
		Rating:    entities.Rating(ratingInt),
		Text:      text,
	}
	if _, err = h.service.CreateReview(r.Context(), cmd); err != nil {
		h.logger.Error("CreateReview failed", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	// После успешного создания перенаправляем на список с фильтром по dataset_id
	http.Redirect(w, r, fmt.Sprintf("/reviews?dataset_id=%d", dsID64), http.StatusSeeOther)
}

// EditForm отображает форму редактирования отзыва.
// GET /reviews/{id}/edit
func (h *ReviewHandler) EditForm(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	rev, err := h.service.GetReviewByID(r.Context(), id)
	if err != nil {
		h.logger.Warn("GetReviewByID failed", zap.Uint64("id", id), zap.Error(err))
		http.NotFound(w, r)
		return
	}

	form := &dto.UpdateReviewForm{
		ID:     rev.ID,
		Rating: int(rev.Rating),
		Text:   rev.Text,
	}

	tpl := template.Must(template.ParseFS(
		templates.TemplatesFS,
		"layout.tmpl",
		"review_form.tmpl",
	))

	data := struct {
		Title      string
		IsNew      bool
		FormAction string
		Form       *dto.UpdateReviewForm
	}{
		Title:      fmt.Sprintf("Редактирование отзыва #%d", id),
		IsNew:      false,
		FormAction: fmt.Sprintf("/reviews/%d", id),
		Form:       form,
	}

	if err := tpl.ExecuteTemplate(w, "layout.tmpl", data); err != nil {
		h.logger.Error("template execution error", zap.Error(err))
	}
}

// Update обрабатывает POST /reviews/{id}.
func (h *ReviewHandler) Update(w http.ResponseWriter, r *http.Request) {
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
	ratingInt, err := strconv.Atoi(r.FormValue("rating"))
	if err != nil || ratingInt < int(entities.Rating1) || ratingInt > int(entities.Rating5) {
		h.logger.Warn("invalid rating", zap.Error(err))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	text := r.FormValue("text")

	cmd := services.UpdateReviewCmd{
		ReviewID: id,
		Rating:   entities.Rating(ratingInt),
		Text:     text,
	}
	if err := h.service.UpdateReview(r.Context(), cmd); err != nil {
		h.logger.Error("UpdateReview failed", zap.Error(err), zap.Uint64("id", id))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// После обновления получаем ревью, чтобы узнать datasetID (для редиректа с фильтром)
	rev, _ := h.service.GetReviewByID(r.Context(), id)
	http.Redirect(w, r, fmt.Sprintf("/reviews?dataset_id=%d", rev.DatasetID), http.StatusSeeOther)
}

// Delete обрабатывает POST /reviews/{id}/delete.
func (h *ReviewHandler) Delete(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	rev, err := h.service.GetReviewByID(r.Context(), id)
	if err != nil {
		h.logger.Warn("GetReviewByID failed", zap.Uint64("id", id), zap.Error(err))
		http.NotFound(w, r)
		return
	}
	dsID := rev.DatasetID

	if err := h.service.DeleteReview(r.Context(), id); err != nil {
		h.logger.Error("DeleteReview failed", zap.Error(err), zap.Uint64("id", id))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/reviews?dataset_id=%d", dsID), http.StatusSeeOther)
}

// Summary отображает средний рейтинг и общее число отзывов для dataset_id.
// GET /reviews/summary/{dataset_id}
func (h *ReviewHandler) Summary(w http.ResponseWriter, r *http.Request) {
	dsIDStr := chi.URLParam(r, "dataset_id")
	dsID, err := strconv.ParseUint(dsIDStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	summary, err := h.service.GetRatingSummary(r.Context(), dsID)
	if err != nil {
		h.logger.Error("GetRatingSummary failed", zap.Error(err), zap.Uint64("datasetID", dsID))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	dtoSummary := dto.RatingSummaryDTO{
		Average: summary.Average,
		Count:   summary.Count,
	}

	tpl := template.Must(template.ParseFS(
		templates.TemplatesFS,
		"layout.tmpl",
		"review_summary.tmpl",
	))

	data := struct {
		Title     string
		DatasetID uint64
		Summary   dto.RatingSummaryDTO
	}{
		Title:     fmt.Sprintf("Сводка по рейтингу для датасета #%d", dsID),
		DatasetID: dsID,
		Summary:   dtoSummary,
	}

	if err := tpl.ExecuteTemplate(w, "layout.tmpl", data); err != nil {
		h.logger.Error("template execution error", zap.Error(err))
	}
}
