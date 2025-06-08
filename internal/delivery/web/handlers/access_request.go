package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"ppo/internal/delivery/web/dto"
	"ppo/internal/delivery/web/middleware"
	"ppo/internal/delivery/web/templates"
	"ppo/internal/services"
)

type AccessRequestHandler struct {
	accessSvc services.AccessService
	notifSvc  services.NotificationService
	dsSvc     services.DatasetService
	userSvc   services.UserService
	logger    *zap.Logger
}

func NewAccessRequestHandler(
	accessSvc services.AccessService,
	notifSvc services.NotificationService,
	dsSvc services.DatasetService,
	userSvc services.UserService,
	logger *zap.Logger,
) *AccessRequestHandler {
	return &AccessRequestHandler{
		accessSvc: accessSvc,
		notifSvc:  notifSvc,
		dsSvc:     dsSvc,
		userSvc:   userSvc,
		logger:    logger,
	}
}

func (h *AccessRequestHandler) RegisterRoutes(r chi.Router) {
	r.With(middleware.RequireRole("user", "admin")).Post(
		"/datasets/{id}/request-access", h.RequestAccess,
	)
	r.With(middleware.RequireRole("user", "admin")).Get(
		"/access-requests", h.ListRequests,
	)
	r.With(middleware.RequireRole("user", "admin")).Post(
		"/access-requests/{rid}/approve", h.Approve,
	)
	r.With(middleware.RequireRole("user", "admin")).Post(
		"/access-requests/{rid}/deny", h.Deny,
	)
}

func (h *AccessRequestHandler) RequestAccess(w http.ResponseWriter, r *http.Request) {
	dsID, _ := strconv.ParseUint(chi.URLParam(r, "id"), 10, 64)
	uid, _ := middleware.FromContext(r.Context())
	cmd := services.RequestAccessCmd{DatasetID: dsID, UserID: uid}
	if err := h.accessSvc.Request(r.Context(), cmd); err != nil {
		http.Error(w, "Не удалось создать запрос", http.StatusBadRequest)
		return
	}

	user, _ := h.userSvc.GetUserByID(r.Context(), uid)
	ds, _ := h.dsSvc.GetDataset(r.Context(), dsID)
	msg := fmt.Sprintf("Пользователь %s запросил доступ к вашему датасету '%s'", user.Username, ds.Name)
	_ = h.notifSvc.NotifyUser(r.Context(), ds.OwnerID, ds.ID, msg)
	http.Redirect(w, r, "/datasets/"+chi.URLParam(r, "id"), http.StatusSeeOther)
}

func (h *AccessRequestHandler) ListRequests(w http.ResponseWriter, r *http.Request) {
	uid, role := middleware.FromContext(r.Context())
	reqs, err := h.accessSvc.ListPending(r.Context(), uid)
	if err != nil {
		http.Error(w, "Ошибка получения запросов", http.StatusInternalServerError)
		return
	}

	dtos := make([]*dto.AccessRequestDTO, len(reqs))
	for i, ar := range reqs {
		user, _ := h.userSvc.GetUserByID(r.Context(), ar.UserID)
		ds, _ := h.dsSvc.GetDataset(r.Context(), ar.DatasetID)
		dtos[i] = &dto.AccessRequestDTO{
			ID:          ar.ID,
			DatasetID:   ar.DatasetID,
			DatasetName: ds.Name,
			UserID:      ar.UserID,
			Username:    user.Username,
			Status:      string(ar.Status),
			CreatedAt:   ar.CreatedAt,
		}
	}
	tpl := template.Must(template.ParseFS(templates.TemplatesFS, "layout.tmpl", "access_list.tmpl"))
	data := struct {
		Title    string
		Role     string
		UserID   uint64
		Requests []*dto.AccessRequestDTO
	}{
		Title:    "Запросы доступа",
		Role:     role,
		UserID:   uid,
		Requests: dtos,
	}
	tpl.ExecuteTemplate(w, "layout.tmpl", data)
}

func (h *AccessRequestHandler) Approve(w http.ResponseWriter, r *http.Request) {
	rid, _ := strconv.ParseUint(chi.URLParam(r, "rid"), 10, 64)
	ownerID, _ := middleware.FromContext(r.Context())
	if err := h.accessSvc.Approve(r.Context(), rid, ownerID); err != nil {
		http.Error(w, "Не удалось одобрить запрос", http.StatusInternalServerError)
		return
	}
	ar, _ := h.accessSvc.FindByRequestID(r.Context(), rid)
	msg := fmt.Sprintf("Ваш запрос #%d на доступ к датасету одобрен", rid)
	_ = h.notifSvc.NotifyUser(r.Context(), ar.UserID, ar.DatasetID, msg)
	http.Redirect(w, r, "/access-requests", http.StatusSeeOther)
}

func (h *AccessRequestHandler) Deny(w http.ResponseWriter, r *http.Request) {
	rid, _ := strconv.ParseUint(chi.URLParam(r, "rid"), 10, 64)
	ownerID, _ := middleware.FromContext(r.Context())
	if err := h.accessSvc.Deny(r.Context(), rid, ownerID); err != nil {
		http.Error(w, "Не удалось отклонить запрос", http.StatusInternalServerError)
		return
	}
	ar, _ := h.accessSvc.FindByRequestID(r.Context(), rid)
	msg := fmt.Sprintf("Ваш запрос #%d на доступ к датасету отклонён", rid)
	_ = h.notifSvc.NotifyUser(r.Context(), ar.UserID, ar.DatasetID, msg)
	http.Redirect(w, r, "/access-requests", http.StatusSeeOther)
}
