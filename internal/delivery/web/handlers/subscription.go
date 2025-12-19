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
	"ppo/internal/services"
)

type SubscriptionHandler struct {
	service services.SubscriptionService
	logger  *zap.Logger
}

func NewSubscriptionHandler(svc services.SubscriptionService, log *zap.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{
		service: svc,
		logger:  log,
	}
}

func (h *SubscriptionHandler) RegisterRoutes(r chi.Router) {
	r.With(middleware.RequireRole("user", "admin")).
		Get("/subscriptions", h.ListSubscriptions)

	r.With(middleware.RequireRole("user", "admin")).
		Get("/subscriptions/new", h.NewForm)
	r.With(middleware.RequireRole("user", "admin")).
		Post("/subscriptions", h.Create)

	r.With(middleware.RequireRole("user", "admin")).
		Post("/subscriptions/{dataset_id}/unsubscribe", h.Unsubscribe)

	r.With(middleware.RequireRole("admin")).
		Get("/subscribers/{dataset_id}", h.ListSubscribers)
}

func (h *SubscriptionHandler) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
	currentUID, currentRole := middleware.FromContext(r.Context())

	subs, err := h.service.ListSubscriptions(r.Context(), currentUID)
	if err != nil {
		h.logger.Error("ListSubscriptions failed", zap.Error(err), zap.Uint64("userID", currentUID))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	dtoList := dto.ToSubscriptionDTOs(subs)

	tpl := template.Must(template.ParseFS(
		templates.TemplatesFS,
		"layout.tmpl",
		"subscription_list.tmpl",
	))

	data := struct {
		Title         string
		Subscriptions []*dto.SubscriptionDTO
		User          uint64
		Role          string
	}{
		Title:         "Мои подписки",
		Subscriptions: dtoList,
		User:          currentUID,
		Role:          currentRole,
	}

	if err := tpl.ExecuteTemplate(w, "layout.tmpl", data); err != nil {
		h.logger.Error("template execution error", zap.Error(err))
	}
}

func (h *SubscriptionHandler) NewForm(w http.ResponseWriter, r *http.Request) {
	currentUID, currentRole := middleware.FromContext(r.Context())

	formDTO := &dto.CreateSubscriptionForm{}

	tpl := template.Must(template.ParseFS(
		templates.TemplatesFS,
		"layout.tmpl",
		"subscription_form.tmpl",
	))

	data := struct {
		Title      string
		FormAction string
		Form       *dto.CreateSubscriptionForm
		User       uint64
		Role       string
	}{
		Title:      "Новая подписка",
		FormAction: "/subscriptions",
		Form:       formDTO,
		User:       currentUID,
		Role:       currentRole,
	}

	if err := tpl.ExecuteTemplate(w, "layout.tmpl", data); err != nil {
		h.logger.Error("template execution error", zap.Error(err))
	}
}

func (h *SubscriptionHandler) Create(w http.ResponseWriter, r *http.Request) {
	currentUID, _ := middleware.FromContext(r.Context())

	if err := r.ParseForm(); err != nil {
		h.logger.Warn("ParseForm error", zap.Error(err))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	dsID64, err := strconv.ParseUint(r.FormValue("dataset_id"), 10, 64)
	if err != nil {
		h.logger.Warn("invalid dataset_id", zap.Error(err))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	if err := h.service.Subscribe(r.Context(), currentUID, dsID64); err != nil {
		h.logger.Error("Subscribe failed", zap.Error(err), zap.Uint64("userID", currentUID), zap.Uint64("datasetID", dsID64))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/subscriptions", http.StatusSeeOther)
}

func (h *SubscriptionHandler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	currentUID, _ := middleware.FromContext(r.Context())

	dsIDStr := chi.URLParam(r, "dataset_id")
	dsID, err := strconv.ParseUint(dsIDStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if err := h.service.Unsubscribe(r.Context(), currentUID, dsID); err != nil {
		h.logger.Error("Unsubscribe failed", zap.Error(err), zap.Uint64("userID", currentUID), zap.Uint64("datasetID", dsID))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/subscriptions", http.StatusSeeOther)
}

func (h *SubscriptionHandler) ListSubscribers(w http.ResponseWriter, r *http.Request) {
	dsIDStr := chi.URLParam(r, "dataset_id")
	dsID, err := strconv.ParseUint(dsIDStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	currentUID, currentRole := middleware.FromContext(r.Context())

	list, err := h.service.ListSubscribers(r.Context(), dsID)
	if err != nil {
		h.logger.Error("ListSubscribers failed", zap.Error(err), zap.Uint64("datasetID", dsID))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	subscribers := dto.ToSubscriberDTOs(list)

	tpl := template.Must(template.ParseFS(
		templates.TemplatesFS,
		"layout.tmpl",
		"subscriber_list.tmpl",
	))

	data := struct {
		Title       string
		DatasetID   uint64
		Subscribers []*dto.SubscriberDTO
		User        uint64
		Role        string
	}{
		Title:       fmt.Sprintf("Подписчики датасета #%d", dsID),
		DatasetID:   dsID,
		Subscribers: subscribers,
		User:        currentUID,
		Role:        currentRole,
	}

	if err := tpl.ExecuteTemplate(w, "layout.tmpl", data); err != nil {
		h.logger.Error("template execution error", zap.Error(err))
	}
}
