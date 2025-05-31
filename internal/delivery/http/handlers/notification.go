package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"ppo/internal/delivery/http/dto"
	"ppo/internal/delivery/http/middleware"
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
func NotifySubscribersHandler(svc services.NotificationService) middleware.HandlerWithError {
	return func(w http.ResponseWriter, r *http.Request) error {
		// 1) Парсим dataset ID из URL
		idParam := chi.URLParam(r, "id")
		datasetID, err := strconv.ParseUint(idParam, 10, 64)
		if err != nil {
			// Неверный формат числа → 400 Bad Request
			return &dto.BadRequestError{Message: "invalid dataset ID"}
		}

		// 2) Считываем тело JSON (dto.NotifyRequest) и маппим в команду
		var req dto.NotifyRequest
		if err := dto.DecodeJSON(r.Body, &req); err != nil {
			return err // если это JSON-парсер, mapErrorToStatus отдаст 400 или 500
		}

		cmd := req.ToCommand(datasetID)

		// 3) Вызываем сервис
		sentCount, err := svc.NotifySubscribers(r.Context(), cmd)
		if err != nil {
			return err
		}

		// 4) Собираем ответ и возвращаем JSON{ "sent": sentCount }
		resp := dto.NotifyResponse{Sent: sentCount}
		dto.WriteJSON(w, http.StatusOK, resp)
		return nil
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
func GetUserNotificationsHandler(svc services.NotificationService) middleware.HandlerWithError {
	return func(w http.ResponseWriter, r *http.Request) error {
		// 1) Парсим user ID из URL
		idParam := chi.URLParam(r, "id")
		userID, err := strconv.ParseUint(idParam, 10, 64)
		if err != nil {
			return &dto.BadRequestError{Message: "invalid user ID"}
		}

		// 2) Вызываем сервис, чтобы получить список уведомлений
		notifs, err := svc.GetNotificationsByUser(r.Context(), userID)
		if err != nil {
			return err
		}

		// 3) Пишем JSON-ответ (массив уведомлений)
		resp := dto.FromEntityList(notifs) // предполагается, что это []dto.NotificationResponse
		dto.WriteJSON(w, http.StatusOK, resp)
		return nil
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
func MarkAsReadHandler(svc services.NotificationService) middleware.HandlerWithError {
	return func(w http.ResponseWriter, r *http.Request) error {
		// 1) Парсим notification ID
		idParam := chi.URLParam(r, "id")
		notifID, err := strconv.ParseUint(idParam, 10, 64)
		if err != nil {
			return &dto.BadRequestError{Message: "invalid notification ID"}
		}

		// 2) Вызываем сервис
		if err := svc.MarkAsRead(r.Context(), notifID); err != nil {
			return err
		}

		// 3) Возвращаем 204 No Content
		w.WriteHeader(http.StatusNoContent)
		return nil
	}
}
