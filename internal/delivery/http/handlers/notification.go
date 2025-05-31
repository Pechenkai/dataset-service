package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"ppo/internal/delivery/http/dto"
	"ppo/internal/services"
)

// @Summary      Notify subscribers
// @Description  Send a message to all subscribers of a dataset
// @Tags         notifications
// @Accept       json
// @Produce      json
// @Param        id     path      int             true  "Dataset ID"
// @Param        body   body      dto.NotifyRequest   true  "Notification message"
// @Success      200    {object}  dto.NotifyResponse
// @Failure      400    {object}  dto.ErrorResponse
// @Failure      404    {object}  dto.ErrorResponse
// @Failure      500    {object}  dto.ErrorResponse
// @Router       /datasets/{id}/notifications [post]
func NotifySubscribersHandler(svc services.NotificationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "id")
		datasetID, err := strconv.ParseUint(idParam, 10, 64)
		if err != nil {
			dto.WriteError(w, &dto.BadRequestError{Message: "invalid dataset ID"})
			return
		}

		var req dto.NotifyRequest
		if err := dto.DecodeJSON(r.Body, &req); err != nil {
			dto.WriteError(w, err)
			return
		}

		cmd := req.ToCommand(datasetID)
		sentCount, err := svc.NotifySubscribers(r.Context(), cmd)
		if err != nil {
			dto.WriteError(w, err)
			return
		}

		resp := dto.NotifyResponse{Sent: sentCount}
		dto.WriteJSON(w, http.StatusOK, resp)
	}
}

// @Summary      List user notifications
// @Description  Returns all notifications for a given user
// @Tags         notifications
// @Produce      json
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  dto.NotificationsResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /users/{id}/notifications [get]
func GetUserNotificationsHandler(svc services.NotificationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "id")
		userID, err := strconv.ParseUint(idParam, 10, 64)
		if err != nil {
			dto.WriteError(w, &dto.BadRequestError{Message: "invalid user ID"})
			return
		}

		notifs, err := svc.GetNotificationsByUser(r.Context(), userID)
		if err != nil {
			dto.WriteError(w, err)
			return
		}

		resp := dto.FromEntityList(notifs)
		dto.WriteJSON(w, http.StatusOK, resp)
	}
}

// @Summary      Mark notification as read
// @Description  Mark a notification as read by its ID
// @Tags         notifications
// @Produce      json
// @Param        id   path      int  true  "Notification ID"
// @Success      204  {object}  nil
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /notifications/{id}/read [put]
func MarkAsReadHandler(svc services.NotificationService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "id")
		notifID, err := strconv.ParseUint(idParam, 10, 64)
		if err != nil {
			dto.WriteError(w, &dto.BadRequestError{Message: "invalid notification ID"})
			return
		}

		if err := svc.MarkAsRead(r.Context(), notifID); err != nil {
			dto.WriteError(w, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
