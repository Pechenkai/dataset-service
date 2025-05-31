package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"ppo/internal/delivery/http/dto"
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
func SubscribeHandler(svc services.SubscriptionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.SubscribeRequest
		if err := dto.DecodeJSON(r.Body, &req); err != nil {
			dto.WriteError(w, err)
			return
		}
		if err := svc.Subscribe(r.Context(), req.UserID, req.DatasetID); err != nil {
			dto.WriteError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
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
func UnsubscribeHandler(svc services.SubscriptionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.SubscribeRequest
		if err := dto.DecodeJSON(r.Body, &req); err != nil {
			dto.WriteError(w, err)
			return
		}
		if err := svc.Unsubscribe(r.Context(), req.UserID, req.DatasetID); err != nil {
			dto.WriteError(w, err)
			return
		}
		w.WriteHeader(http.StatusNoContent)
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
func ListSubscribersHandler(svc services.SubscriptionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "id")
		datasetID, err := strconv.ParseUint(idParam, 10, 64)
		if err != nil {
			dto.WriteError(w, &dto.BadRequestError{Message: "invalid dataset ID"})
			return
		}

		subs, err := svc.ListSubscribers(r.Context(), datasetID)
		if err != nil {
			dto.WriteError(w, err)
			return
		}

		resp := dto.ListSubscribersResponse{Subscribers: subs}
		dto.WriteJSON(w, http.StatusOK, resp)
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
func ListSubscriptionsHandler(svc services.SubscriptionService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "id")
		userID, err := strconv.ParseUint(idParam, 10, 64)
		if err != nil {
			dto.WriteError(w, &dto.BadRequestError{Message: "invalid user ID"})
			return
		}

		datasets, err := svc.ListSubscriptions(r.Context(), userID)
		if err != nil {
			dto.WriteError(w, err)
			return
		}

		resp := dto.ListSubscriptionsResponse{Subscriptions: datasets}
		dto.WriteJSON(w, http.StatusOK, resp)
	}
}
