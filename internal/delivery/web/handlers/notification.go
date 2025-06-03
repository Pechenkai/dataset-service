package handlers

import (
	"html/template"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"ppo/internal/delivery/web/dto"
	"ppo/internal/delivery/web/templates"
	"ppo/internal/services"
)

// NotificationHandler отвечает за работу с уведомлениями.
type NotificationHandler struct {
	service services.NotificationService
	logger  *zap.Logger
}

// NewNotificationHandler создаёт новый NotificationHandler без предзагрузки всех шаблонов.
func NewNotificationHandler(svc services.NotificationService, log *zap.Logger) *NotificationHandler {
	return &NotificationHandler{
		service: svc,
		logger:  log,
	}
}

// RegisterRoutes регистрирует HTTP‐маршруты для уведомлений.
func (h *NotificationHandler) RegisterRoutes(r chi.Router) {
	r.Get("/notifications", h.List)
	r.Get("/notifications/new", h.NewForm)
	r.Post("/notifications", h.Create)
	r.Post("/notifications/{id}/read", h.MarkRead)
}

// List показывает все уведомления для текущего пользователя.
// GET /notifications
func (h *NotificationHandler) List(w http.ResponseWriter, r *http.Request) {
	// TODO: вместо 1 забрать реальный userID из контекста/сессии
	userID := uint64(1)

	notifs, err := h.service.GetNotificationsByUser(r.Context(), userID)
	if err != nil {
		h.logger.Error("GetNotificationsByUser failed", zap.Error(err), zap.Uint64("userID", userID))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Конвертируем в DTO
	notifications := dto.ToNotificationDTOs(notifs)

	// Парсим только layout.tmpl + notification_list.tmpl
	tpl := template.Must(template.ParseFS(
		templates.TemplatesFS,
		"layout.tmpl",
		"notification_list.tmpl",
	))

	data := struct {
		Title         string
		Notifications []*dto.NotificationDTO
	}{
		Title:         "Мои уведомления",
		Notifications: notifications,
	}

	if err := tpl.ExecuteTemplate(w, "layout.tmpl", data); err != nil {
		h.logger.Error("template execution error", zap.Error(err))
	}
}

// NewForm отображает форму для рассылки уведомления подписчикам.
// GET /notifications/new
func (h *NotificationHandler) NewForm(w http.ResponseWriter, r *http.Request) {
	// DTO для формы создания уведомления
	formDTO := &dto.CreateNotificationForm{}

	// Парсим только layout.tmpl + notification_form.tmpl
	tpl := template.Must(template.ParseFS(
		templates.TemplatesFS,
		"layout.tmpl",
		"notification_form.tmpl",
	))

	data := struct {
		Title      string
		FormAction string
		Form       *dto.CreateNotificationForm
	}{
		Title:      "Новое уведомление",
		FormAction: "/notifications",
		Form:       formDTO,
	}

	if err := tpl.ExecuteTemplate(w, "layout.tmpl", data); err != nil {
		h.logger.Error("template execution error", zap.Error(err))
	}
}

// Create обрабатывает POST /notifications и рассылает уведомление подписчикам.
func (h *NotificationHandler) Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		h.logger.Warn("ParseForm error", zap.Error(err))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	datasetID64, err := strconv.ParseUint(r.FormValue("dataset_id"), 10, 64)
	if err != nil {
		h.logger.Warn("invalid dataset_id", zap.Error(err))
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}
	message := r.FormValue("message")

	cmd := services.NotifySubscribersCmd{
		DatasetID: datasetID64,
		Message:   message,
	}

	_, err = h.service.NotifySubscribers(r.Context(), cmd)
	if err != nil {
		h.logger.Error("NotifySubscribers failed", zap.Error(err), zap.Uint64("datasetID", datasetID64))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// После успеха редиректим обратно к списку уведомлений
	http.Redirect(w, r, "/notifications", http.StatusSeeOther)
}

// MarkRead обрабатывает POST /notifications/{id}/read — помечает уведомление как прочитанное.
// POST /notifications/{id}/read
func (h *NotificationHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if err := h.service.MarkAsRead(r.Context(), id); err != nil {
		h.logger.Error("MarkAsRead failed", zap.Uint64("id", id), zap.Error(err))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/notifications", http.StatusSeeOther)
}
