package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"ppo/internal/delivery/http/dto"
	"ppo/internal/delivery/http/middleware"
	"ppo/internal/services"
)

// @Summary      Subscribe user to dataset
// @Description  Подписывает пользователя на обновления датасета
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        body  body      dto.SubscribeRequest  true  "UserID & DatasetID"
// @Success      204  {object}  nil
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      409  {object}  dto.ErrorResponse
// @Router       /subscriptions [post]
func SubscribeHandler(svc services.SubscriptionService) middleware.HandlerWithError {
	return func(w http.ResponseWriter, r *http.Request) error {
		var req dto.SubscribeRequest
		if err := dto.DecodeJSON(r.Body, &req); err != nil {
			return err // mapErrorToStatus ⇒ 400 или 500
		}

		if err := svc.Subscribe(r.Context(), req.UserID, req.DatasetID); err != nil {
			return err // ErrAlreadySubscribed ⇒ 409, ErrNotSubscribed не ожидается здесь, иначе 500
		}

		w.WriteHeader(http.StatusNoContent)
		return nil
	}
}

// @Summary      Unsubscribe user from dataset
// @Description  Отписывает пользователя от обновлений датасета
// @Tags         subscriptions
// @Accept       json
// @Produce      json
// @Param        body  body      dto.SubscribeRequest  true  "UserID & DatasetID"
// @Success      204  {object}  nil
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Router       /subscriptions [delete]
func UnsubscribeHandler(svc services.SubscriptionService) middleware.HandlerWithError {
	return func(w http.ResponseWriter, r *http.Request) error {
		var req dto.SubscribeRequest
		if err := dto.DecodeJSON(r.Body, &req); err != nil {
			return err // неверный JSON => 400
		}

		if err := svc.Unsubscribe(r.Context(), req.UserID, req.DatasetID); err != nil {
			return err // ErrNotSubscribed ⇒ 404, иначе 500
		}

		w.WriteHeader(http.StatusNoContent)
		return nil
	}
}

// @Summary      List subscribers for a dataset
// @Description  Возвращает список user_id, подписанных на датасет
// @Tags         subscriptions
// @Produce      json
// @Param        id   path      int  true  "Dataset ID"
// @Success      200  {object}  dto.ListSubscribersResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Router       /datasets/{id}/subscribers [get]
func ListSubscribersHandler(svc services.SubscriptionService) middleware.HandlerWithError {
	return func(w http.ResponseWriter, r *http.Request) error {
		idParam := chi.URLParam(r, "id")
		datasetID, err := strconv.ParseUint(idParam, 10, 64)
		if err != nil {
			return &dto.BadRequestError{Message: "invalid dataset ID"}
		}

		subs, err := svc.ListSubscribers(r.Context(), datasetID)
		if err != nil {
			return err // Возможны ошибки 404 (например, dataset не найден) или 500
		}

		resp := dto.ListSubscribersResponse{Subscribers: dto.ToSubscriberIDs(subs)}
		dto.WriteJSON(w, http.StatusOK, resp)
		return nil
	}
}

// @Summary      List subscriptions for a user
// @Description  Возвращает список dataset_id, на которые подписан пользователь
// @Tags         subscriptions
// @Produce      json
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  dto.ListSubscriptionsResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Router       /users/{id}/subscriptions [get]
func ListSubscriptionsHandler(svc services.SubscriptionService) middleware.HandlerWithError {
	return func(w http.ResponseWriter, r *http.Request) error {
		idParam := chi.URLParam(r, "id")
		userID, err := strconv.ParseUint(idParam, 10, 64)
		if err != nil {
			return &dto.BadRequestError{Message: "invalid user ID"}
		}

		datasets, err := svc.ListSubscriptions(r.Context(), userID)
		if err != nil {
			return err // 404 если юзер не найден, либо 500
		}

		resp := dto.ListSubscriptionsResponse{Subscriptions: dto.ToDatasetIDs(datasets)}
		dto.WriteJSON(w, http.StatusOK, resp)
		return nil
	}
}
