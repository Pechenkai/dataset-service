package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"ppo/internal/delivery/web/middleware"
	"strconv"

	chi "github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"ppo/internal/delivery/web/dto"
	"ppo/internal/delivery/web/templates"
	"ppo/internal/entities"
	"ppo/internal/services"
)

type ReviewHandler struct {
	service services.ReviewService
	logger  *zap.Logger
}

func NewReviewHandler(svc services.ReviewService, log *zap.Logger) *ReviewHandler {
	return &ReviewHandler{
		service: svc,
		logger:  log,
	}
}

func (h *ReviewHandler) RegisterRoutes(r chi.Router) {
	r.Get("/reviews", h.List)
	r.With(middleware.RequireRole("user", "admin")).
		Get("/reviews/new", h.NewForm)
	r.With(middleware.RequireRole("user", "admin")).
		Post("/reviews", h.Create)
	r.With(middleware.RequireRole("user", "admin")).
		Get("/reviews/{id}/edit", h.EditForm)
	r.With(middleware.RequireRole("user", "admin")).
		Post("/reviews/{id}", h.Update)
	r.With(middleware.RequireRole("user", "admin")).
		Post("/reviews/{id}/delete", h.Delete)
	r.Get("/reviews/summary/{dataset_id}", h.Summary)
}

func (h *ReviewHandler) List(w http.ResponseWriter, r *http.Request) {
	currentUID, currentRole := middleware.FromContext(r.Context())

	if dsIDStr := r.URL.Query().Get("dataset_id"); dsIDStr != "" {
		dsID, err := strconv.ParseUint(dsIDStr, 10, 64)
		if err == nil {
			revs, err := h.service.ListByDataset(r.Context(), dsID)
			if err != nil {
				h.logger.Error("ListByDataset failed", zap.Error(err), zap.Uint64("datasetID", dsID))
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
				Title           string
				Reviews         []*dto.ReviewDTO
				Role            string
				User            uint64
				FilterDatasetID uint64
			}{
				Title:           fmt.Sprintf("Отзывы для датасета #%d", dsID),
				Reviews:         reviewDTOs,
				Role:            currentRole,
				User:            currentUID,
				FilterDatasetID: dsID,
			}

			if err := tpl.ExecuteTemplate(w, "layout.tmpl", data); err != nil {
				h.logger.Error("template execution error", zap.Error(err))
			}
			return
		}
	}

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
				User         uint64
				Role         string
				FilterUserID uint64
			}{
				Title:        fmt.Sprintf("Отзывы пользователя #%d", uID),
				Reviews:      reviewDTOs,
				User:         currentUID,
				Role:         currentRole,
				FilterUserID: uID,
			}

			if err := tpl.ExecuteTemplate(w, "layout.tmpl", data); err != nil {
				h.logger.Error("template execution error", zap.Error(err))
			}
			return
		}
	}

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
		Role    string
		User    uint64
	}{
		Title:   "Все отзывы",
		User:    currentUID,
		Role:    currentRole,
		Reviews: reviewDTOs,
	}

	if err := tpl.ExecuteTemplate(w, "layout.tmpl", data); err != nil {
		h.logger.Error("template execution error", zap.Error(err))
	}
}

func (h *ReviewHandler) NewForm(w http.ResponseWriter, r *http.Request) {
	formDTO := &dto.CreateReviewForm{}

	currentUID, currentRole := middleware.FromContext(r.Context())

	tpl := template.Must(template.ParseFS(
		templates.TemplatesFS,
		"layout.tmpl",
		"review_form.tmpl",
	))

	data := struct {
		Title      string
		FormAction string
		Form       *dto.CreateReviewForm
		User       uint64
		Role       string
	}{
		Title:      "Новый отзыв",
		FormAction: "/reviews",
		Form:       formDTO,
		User:       currentUID,
		Role:       currentRole,
	}

	if err := tpl.ExecuteTemplate(w, "layout.tmpl", data); err != nil {
		h.logger.Error("template execution error", zap.Error(err))
	}
}

func (h *ReviewHandler) Create(w http.ResponseWriter, r *http.Request) {
	currentUID, _ := middleware.FromContext(r.Context())
	if currentUID == 0 {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if err := r.ParseForm(); err != nil {
		h.logger.Warn("ParseForm error", zap.Error(err))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	dsID, err := strconv.ParseUint(r.FormValue("dataset_id"), 10, 64)
	if err != nil {
		h.logger.Warn("invalid dataset_id", zap.Error(err))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	ratingInt, err := strconv.Atoi(r.FormValue("rating"))
	if err != nil || ratingInt < 1 || ratingInt > 5 {
		h.logger.Warn("invalid rating", zap.Error(err))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	text := r.FormValue("text")

	cmd := services.CreateReviewCmd{
		UserID:    currentUID,
		DatasetID: dsID,
		Rating:    entities.Rating(ratingInt),
		Text:      text,
	}
	_, err = h.service.CreateReview(r.Context(), cmd)
	if err != nil {
		h.logger.Error("CreateReview failed", zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/datasets/%d", dsID), http.StatusSeeOther)
}

func (h *ReviewHandler) EditForm(w http.ResponseWriter, r *http.Request) {
	currentUID, currentRole := middleware.FromContext(r.Context())

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

	if currentRole != "admin" && currentUID != rev.UserID {
		http.Error(w, "Forbidden", http.StatusForbidden)
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
		User       uint64
		Role       string
	}{
		Title:      fmt.Sprintf("Редактирование отзыва #%d", id),
		IsNew:      false,
		FormAction: fmt.Sprintf("/reviews/%d", id),
		Form:       form,
		User:       currentUID,
		Role:       currentRole,
	}

	if err := tpl.ExecuteTemplate(w, "layout.tmpl", data); err != nil {
		h.logger.Error("template execution error", zap.Error(err))
	}
}

func (h *ReviewHandler) Update(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	currentUID, currentRole := middleware.FromContext(r.Context())

	rev, err := h.service.GetReviewByID(r.Context(), id)
	if err != nil {
		h.logger.Warn("GetReviewByID failed", zap.Uint64("id", id), zap.Error(err))
		http.NotFound(w, r)
		return
	}

	if currentRole != "admin" && currentUID != rev.UserID {
		http.Error(w, "Forbidden", http.StatusForbidden)
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

	updated, _ := h.service.GetReviewByID(r.Context(), id)
	http.Redirect(w, r, fmt.Sprintf("/reviews?dataset_id=%d", updated.DatasetID), http.StatusSeeOther)
}

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
