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

// SubscriptionHandler отвечает за работу с подписками.
type SubscriptionHandler struct {
	service services.SubscriptionService
	logger  *zap.Logger
}

// NewSubscriptionHandler конструирует контроллер для подписок.
func NewSubscriptionHandler(svc services.SubscriptionService, log *zap.Logger) *SubscriptionHandler {
	return &SubscriptionHandler{
		service: svc,
		logger:  log,
	}
}

// RegisterRoutes регистрирует все URL для подписок/подписчиков.
func (h *SubscriptionHandler) RegisterRoutes(r chi.Router) {
	r.Get("/subscriptions", h.ListSubscriptions)
	r.Get("/subscriptions/new", h.NewForm)
	r.Post("/subscriptions", h.Create)
	r.Post("/subscriptions/{dataset_id}/unsubscribe", h.Unsubscribe)
	r.Get("/subscribers/{dataset_id}", h.ListSubscribers)
}

// ListSubscriptions показывает все подписки (datasetID) для текущего пользователя.
// GET /subscriptions
func (h *SubscriptionHandler) ListSubscriptions(w http.ResponseWriter, r *http.Request) {
	// TODO: получить реальный userID из сессии/контекста
	userID := uint64(1)

	list, err := h.service.ListSubscriptions(r.Context(), userID)
	if err != nil {
		h.logger.Error("ListSubscriptions failed", zap.Error(err), zap.Uint64("userID", userID))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	subscriptions := dto.ToSubscriptionDTOs(list)

	// Парсим только layout.tmpl + subscription_list.tmpl
	tpl := template.Must(template.ParseFS(
		templates.TemplatesFS,
		"layout.tmpl",
		"subscription_list.tmpl",
	))

	data := struct {
		Title         string
		Subscriptions []*dto.SubscriptionDTO
	}{
		Title:         "Мои подписки",
		Subscriptions: subscriptions,
	}

	if err := tpl.ExecuteTemplate(w, "layout.tmpl", data); err != nil {
		h.logger.Error("template execution error", zap.Error(err))
	}
}

// NewForm рендерит форму для создания подписки.
// GET /subscriptions/new
func (h *SubscriptionHandler) NewForm(w http.ResponseWriter, r *http.Request) {
	formDTO := &dto.CreateSubscriptionForm{}

	// Парсим только layout.tmpl + subscription_form.tmpl
	tpl := template.Must(template.ParseFS(
		templates.TemplatesFS,
		"layout.tmpl",
		"subscription_form.tmpl",
	))

	data := struct {
		Title      string
		FormAction string
		Form       *dto.CreateSubscriptionForm
	}{
		Title:      "Новая подписка",
		FormAction: "/subscriptions",
		Form:       formDTO,
	}

	if err := tpl.ExecuteTemplate(w, "layout.tmpl", data); err != nil {
		h.logger.Error("template execution error", zap.Error(err))
	}
}

// Create обрабатывает POST /subscriptions и создаёт подписку.
func (h *SubscriptionHandler) Create(w http.ResponseWriter, r *http.Request) {
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
	// TODO: получить настоящий userID из контекста
	userID := uint64(1)

	if err := h.service.Subscribe(r.Context(), userID, dsID64); err != nil {
		h.logger.Error("Subscribe failed", zap.Error(err), zap.Uint64("userID", userID), zap.Uint64("datasetID", dsID64))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// После подписки возвращаемся к списку подписок пользователя
	http.Redirect(w, r, "/subscriptions", http.StatusSeeOther)
}

// Unsubscribe обрабатывает POST /subscriptions/{dataset_id}/unsubscribe.
func (h *SubscriptionHandler) Unsubscribe(w http.ResponseWriter, r *http.Request) {
	dsIDStr := chi.URLParam(r, "dataset_id")
	dsID, err := strconv.ParseUint(dsIDStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	// TODO: получить настоящий userID из контекста
	userID := uint64(1)

	if err := h.service.Unsubscribe(r.Context(), userID, dsID); err != nil {
		h.logger.Error("Unsubscribe failed", zap.Error(err), zap.Uint64("userID", userID), zap.Uint64("datasetID", dsID))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/subscriptions", http.StatusSeeOther)
}

// ListSubscribers показывает всех пользователей, подписанных на datasetID.
// GET /subscribers/{dataset_id}
func (h *SubscriptionHandler) ListSubscribers(w http.ResponseWriter, r *http.Request) {
	dsIDStr := chi.URLParam(r, "dataset_id")
	dsID, err := strconv.ParseUint(dsIDStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	list, err := h.service.ListSubscribers(r.Context(), dsID)
	if err != nil {
		h.logger.Error("ListSubscribers failed", zap.Error(err), zap.Uint64("datasetID", dsID))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	subscribers := dto.ToSubscriberDTOs(list)

	// Парсим только layout.tmpl + subscriber_list.tmpl
	tpl := template.Must(template.ParseFS(
		templates.TemplatesFS,
		"layout.tmpl",
		"subscriber_list.tmpl",
	))

	data := struct {
		Title       string
		DatasetID   uint64
		Subscribers []*dto.SubscriberDTO
	}{
		Title:       fmt.Sprintf("Подписчики датасета #%d", dsID),
		DatasetID:   dsID,
		Subscribers: subscribers,
	}

	if err := tpl.ExecuteTemplate(w, "layout.tmpl", data); err != nil {
		h.logger.Error("template execution error", zap.Error(err))
	}
}
