package v2

import (
	"net/http"

	chi "github.com/go-chi/chi/v5"

	"ppo/internal/delivery/http/dto"
	"ppo/internal/entities"
	"ppo/internal/services"
)

func (h *Handler) ListNotifications(w http.ResponseWriter, r *http.Request) error {
	limit, offset, err := parsePagination(r, 50, 100)
	if err != nil {
		return err
	}
	q := r.URL.Query()
	userID, err := parseUintPtr(q, "user_id")
	if err != nil {
		return err
	}
	current, err := currentUserOrError(r.Context())
	if err != nil {
		return err
	}
	if userID == nil {
		id := current.ID
		userID = &id
	} else {
		if *userID != current.ID && !isAdmin(current) {
			return services.ErrRequestForbidden
		}
	}
	datasetID, err := parseUintPtr(q, "dataset_id")
	if err != nil {
		return err
	}
	isRead, err := parseBoolPtr(q, "is_read")
	if err != nil {
		return err
	}

	notifs, err := h.notifications.GetNotificationsByUser(r.Context(), *userID)
	if err != nil {
		return err
	}

	filtered := filterNotifications(notifs, datasetID, isRead)
	writeJSON(w, http.StatusOK, paginateNotifications(filtered, limit, offset))
	return nil
}

func (h *Handler) GetNotification(w http.ResponseWriter, r *http.Request) error {
	id, err := parseIDParam(chi.URLParam(r, "notificationId"))
	if err != nil {
		return err
	}
	notif, err := h.notifications.GetNotificationByID(r.Context(), id)
	if err != nil {
		return err
	}

	user, err := currentUserOrError(r.Context())
	if err != nil {
		return err
	}
	if notif.UserID != user.ID && !isAdmin(user) {
		return services.ErrRequestForbidden
	}

	writeJSON(w, http.StatusOK, toNotificationResponse(notif))
	return nil
}

func (h *Handler) UpdateNotification(w http.ResponseWriter, r *http.Request) error {
	id, err := parseIDParam(chi.URLParam(r, "notificationId"))
	if err != nil {
		return err
	}
	var req struct {
		IsRead *bool `json:"is_read"`
	}
	if err := decodeJSON(r, &req); err != nil {
		return err
	}
	if req.IsRead == nil {
		return &dto.BadRequestError{Message: "is_read is required"}
	}

	notif, err := h.notifications.GetNotificationByID(r.Context(), id)
	if err != nil {
		return err
	}
	user, err := currentUserOrError(r.Context())
	if err != nil {
		return err
	}
	if notif.UserID != user.ID && !isAdmin(user) {
		return services.ErrRequestForbidden
	}

	if err := h.notifications.SetReadStatus(r.Context(), id, *req.IsRead); err != nil {
		return err
	}
	updated, err := h.notifications.GetNotificationByID(r.Context(), id)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, toNotificationResponse(updated))
	return nil
}

func (h *Handler) ListUserNotifications(w http.ResponseWriter, r *http.Request) error {
	userID, err := parseIDParam(chi.URLParam(r, "userId"))
	if err != nil {
		return err
	}
	// Доступно самому пользователю и администраторам
	if _, err := ensureUserOrAdmin(r.Context(), userID); err != nil {
		return err
	}

	limit, offset, err := parsePagination(r, 50, 100)
	if err != nil {
		return err
	}

	notifs, err := h.notifications.GetNotificationsByUser(r.Context(), userID)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, paginateNotifications(toNotificationResponses(notifs), limit, offset))
	return nil
}

func filterNotifications(list []*entities.Notification, datasetID *uint64, isRead *bool) []NotificationResponse {
	res := make([]NotificationResponse, 0, len(list))
	for _, notif := range list {
		if datasetID != nil && notif.DatasetID != *datasetID {
			continue
		}
		if isRead != nil && notif.IsRead != *isRead {
			continue
		}
		res = append(res, toNotificationResponse(notif))
	}
	return res
}

func paginateNotifications(list []NotificationResponse, limit, offset int) NotificationsResponse {
	total := len(list)
	start := clamp(offset, 0, total)
	end := clamp(offset+limit, start, total)
	return NotificationsResponse{
		Items: list[start:end],
		Meta: PaginationMeta{
			Total:  total,
			Limit:  limit,
			Offset: offset,
		},
	}
}

func toNotificationResponses(list []*entities.Notification) []NotificationResponse {
	res := make([]NotificationResponse, 0, len(list))
	for _, notif := range list {
		res = append(res, toNotificationResponse(notif))
	}
	return res
}

func toNotificationResponse(notif *entities.Notification) NotificationResponse {
	return NotificationResponse{
		ID:        notif.ID,
		UserID:    notif.UserID,
		DatasetID: notif.DatasetID,
		Message:   notif.Message,
		IsRead:    notif.IsRead,
		CreatedAt: notif.CreatedAt,
	}
}
