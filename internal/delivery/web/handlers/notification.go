package handlers

import (
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

type NotificationHandler struct {
	service        services.NotificationService
	datasetService services.DatasetService
	logger         *zap.Logger
}

func NewNotificationHandler(svc services.NotificationService, dssvc services.DatasetService, log *zap.Logger) *NotificationHandler {
	return &NotificationHandler{
		service:        svc,
		datasetService: dssvc,
		logger:         log,
	}
}

func (h *NotificationHandler) RegisterRoutes(r chi.Router) {
	r.Get("/notifications", h.List)
	r.Get("/notifications/new", h.NewForm)
	r.Post("/notifications", h.Create)
	r.Post("/notifications/{id}/read", h.MarkRead)
}

func (h *NotificationHandler) List(w http.ResponseWriter, r *http.Request) {
	currentUID, currentRole := middleware.FromContext(r.Context())
	userID := currentUID

	notifs, err := h.service.GetNotificationsByUser(r.Context(), userID)
	if err != nil {
		h.logger.Error("GetNotificationsByUser failed", zap.Error(err), zap.Uint64("userID", userID))
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	notifications := dto.ToNotificationDTOs(notifs)

	for _, nd := range notifications {
		ds, err := h.datasetService.GetDataset(r.Context(), nd.DatasetID)
		if err != nil {
			nd.DatasetName = "—"
		} else {
			nd.DatasetName = ds.Name
		}
	}

	tpl := template.Must(template.ParseFS(
		templates.TemplatesFS,
		"layout.tmpl",
		"notification_list.tmpl",
	))

	data := struct {
		Title         string
		Role          string
		UserID        uint64
		Notifications []*dto.NotificationDTO
	}{
		Title:         "Мои уведомления",
		Role:          currentRole,
		UserID:        currentUID,
		Notifications: notifications,
	}

	if err := tpl.ExecuteTemplate(w, "layout.tmpl", data); err != nil {
		h.logger.Error("template execution error", zap.Error(err))
	}
}

func (h *NotificationHandler) NewForm(w http.ResponseWriter, r *http.Request) {
	formDTO := &dto.CreateNotificationForm{}

	currentUID, currentRole := middleware.FromContext(r.Context())

	tpl := template.Must(template.ParseFS(
		templates.TemplatesFS,
		"layout.tmpl",
		"notification_form.tmpl",
	))

	data := struct {
		Title      string
		Role       string
		UserID     uint64
		FormAction string
		Form       *dto.CreateNotificationForm
	}{
		Title:      "Новое уведомление",
		Role:       currentRole,
		UserID:     currentUID,
		FormAction: "/notifications",
		Form:       formDTO,
	}

	if err := tpl.ExecuteTemplate(w, "layout.tmpl", data); err != nil {
		h.logger.Error("template execution error", zap.Error(err))
	}
}

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

	http.Redirect(w, r, "/notifications", http.StatusSeeOther)
}

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
