package v2

import (
	"context"
	"net/http"

	"github.com/go-chi/chi/v5"

	"ppo/internal/delivery/http/dto"
	"ppo/internal/entities"
	"ppo/internal/services"
)

func (h *Handler) ListSubscriptions(w http.ResponseWriter, r *http.Request) error {
	limit, offset, err := parsePagination(r, 50, 200)
	if err != nil {
		return err
	}
	q := r.URL.Query()
	userID, err := parseUintPtr(q, "user_id")
	if err != nil {
		return err
	}
	datasetID, err := parseUintPtr(q, "dataset_id")
	if err != nil {
		return err
	}

	items, err := h.collectSubscriptions(r.Context(), userID, datasetID)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, paginateSubscriptions(items, limit, offset))
	return nil
}

func (h *Handler) CreateSubscription(w http.ResponseWriter, r *http.Request) error {
	var req struct {
		UserID    uint64 `json:"user_id"`
		DatasetID uint64 `json:"dataset_id"`
	}
	if err := decodeJSON(r, &req); err != nil {
		return err
	}
	if req.UserID == 0 || req.DatasetID == 0 {
		return &dto.BadRequestError{Message: "user_id and dataset_id are required"}
	}
	if err := h.subscriptions.Subscribe(r.Context(), req.UserID, req.DatasetID); err != nil {
		return err
	}
	sub, err := h.getSubscriptionRecord(r.Context(), req.UserID, req.DatasetID)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusCreated, toSubscriptionResponse(sub))
	return nil
}

func (h *Handler) GetSubscription(w http.ResponseWriter, r *http.Request) error {
	idValue, err := parseIDParam(chi.URLParam(r, "subscriptionId"))
	if err != nil {
		return err
	}
	userID, datasetID := services.DecodeSubscriptionID(idValue)
	if userID == 0 || datasetID == 0 {
		return &dto.BadRequestError{Message: "invalid subscription identifier"}
	}
	sub, err := h.getSubscriptionRecord(r.Context(), userID, datasetID)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, toSubscriptionResponse(sub))
	return nil
}

func (h *Handler) DeleteSubscription(w http.ResponseWriter, r *http.Request) error {
	idValue, err := parseIDParam(chi.URLParam(r, "subscriptionId"))
	if err != nil {
		return err
	}
	userID, datasetID := services.DecodeSubscriptionID(idValue)
	if userID == 0 || datasetID == 0 {
		return &dto.BadRequestError{Message: "invalid subscription identifier"}
	}
	if err := h.subscriptions.Unsubscribe(r.Context(), userID, datasetID); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (h *Handler) collectSubscriptions(ctx context.Context, userID, datasetID *uint64) ([]SubscriptionResponse, error) {
	switch {
	case userID != nil:
		datasets, err := h.subscriptions.ListSubscriptions(ctx, *userID)
		if err != nil {
			return nil, err
		}
		items := make([]SubscriptionResponse, 0, len(datasets))
		for _, sub := range datasets {
			if datasetID != nil && sub.DatasetID != *datasetID {
				continue
			}
			items = append(items, toSubscriptionResponse(sub))
		}
		return items, nil
	case datasetID != nil:
		users, err := h.subscriptions.ListSubscribers(ctx, *datasetID)
		if err != nil {
			return nil, err
		}
		items := make([]SubscriptionResponse, 0, len(users))
		for _, sub := range users {
			items = append(items, toSubscriptionResponse(sub))
		}
		return items, nil
	default:
		subs, err := h.subscriptions.ListAllSubscriptions(ctx)
		if err != nil {
			return nil, err
		}
		items := make([]SubscriptionResponse, 0, len(subs))
		for _, sub := range subs {
			items = append(items, toSubscriptionResponse(sub))
		}
		return items, nil
	}
}

func (h *Handler) getSubscriptionRecord(ctx context.Context, userID, datasetID uint64) (*entities.Subscription, error) {
	subs, err := h.subscriptions.ListSubscriptions(ctx, userID)
	if err != nil {
		return nil, err
	}
	for _, sub := range subs {
		if sub.DatasetID == datasetID {
			return sub, nil
		}
	}
	return nil, services.ErrNotSubscribed
}

func toSubscriptionResponse(sub *entities.Subscription) SubscriptionResponse {
	return SubscriptionResponse{
		ID:           services.EncodeSubscriptionID(sub.UserID, sub.DatasetID),
		UserID:       sub.UserID,
		DatasetID:    sub.DatasetID,
		SubscribedAt: sub.CreatedAt,
	}
}

func paginateSubscriptions(items []SubscriptionResponse, limit, offset int) SubscriptionsResponse {
	total := len(items)
	start := clamp(offset, 0, total)
	end := clamp(offset+limit, start, total)
	return SubscriptionsResponse{
		Items: items[start:end],
		Meta: PaginationMeta{
			Total:  total,
			Limit:  limit,
			Offset: offset,
		},
	}
}
