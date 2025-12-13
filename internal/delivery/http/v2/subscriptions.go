package v2

import (
	"context"
	"net/http"
	"strings"

	chi "github.com/go-chi/chi/v5"

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
	currentUser, err := currentUserOrError(r.Context())
	if err != nil {
		return err
	}

	var req struct {
		UserID    *uint64 `json:"user_id,omitempty"`
		DatasetID uint64  `json:"dataset_id"`
	}
	if err := decodeJSON(r, &req); err != nil {
		return err
	}
	if req.DatasetID == 0 {
		return &dto.BadRequestError{Message: "dataset_id is required"}
	}

	targetUserID := currentUser.ID
	if req.UserID != nil {
		if *req.UserID == 0 {
			return &dto.BadRequestError{Message: "user_id must be positive"}
		}
		if !strings.EqualFold(currentUser.Role, entities.RoleAdmin) && *req.UserID != currentUser.ID {
			return services.ErrRequestForbidden
		}
		targetUserID = *req.UserID
	}

	if err := h.subscriptions.Subscribe(r.Context(), targetUserID, req.DatasetID); err != nil {
		return err
	}
	sub, err := h.getSubscriptionRecord(r.Context(), targetUserID, req.DatasetID)
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
	if userID != nil {
		return h.subscriptionsForUser(ctx, *userID, datasetID)
	}
	if datasetID != nil {
		return h.subscribersForDataset(ctx, *datasetID)
	}
	return h.allSubscriptions(ctx)
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

func (h *Handler) subscriptionsForUser(ctx context.Context, userID uint64, datasetID *uint64) ([]SubscriptionResponse, error) {
	datasets, err := h.subscriptions.ListSubscriptions(ctx, userID)
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
}

func (h *Handler) subscribersForDataset(ctx context.Context, datasetID uint64) ([]SubscriptionResponse, error) {
	users, err := h.subscriptions.ListSubscribers(ctx, datasetID)
	if err != nil {
		return nil, err
	}
	items := make([]SubscriptionResponse, 0, len(users))
	for _, sub := range users {
		items = append(items, toSubscriptionResponse(sub))
	}
	return items, nil
}

func (h *Handler) allSubscriptions(ctx context.Context) ([]SubscriptionResponse, error) {
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
