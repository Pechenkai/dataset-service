package v2

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	chi "github.com/go-chi/chi/v5"

	"ppo/internal/delivery/http/dto"
	"ppo/internal/entities"
	"ppo/internal/services"
)

func (h *Handler) ListAccessRequests(w http.ResponseWriter, r *http.Request) error {
	limit, offset, err := parsePagination(r, 50, 200)
	if err != nil {
		return err
	}

	user, err := currentUserOrError(r.Context())
	if err != nil {
		return err
	}

	filters, err := parseAccessRequestFilters(r, user)
	if err != nil {
		return err
	}

	datasetIDs, err := h.datasetIDsForAccessRequests(r.Context(), user, filters)
	if err != nil {
		return err
	}

	requests, err := h.collectAccessRequests(r.Context(), datasetIDs)
	if err != nil {
		return err
	}

	filtered := filterAccessRequests(requests, filters)
	items := paginateAccessRequests(filtered, limit, offset)

	writeJSON(w, http.StatusOK, AccessRequestsResponse{
		Items: items,
		Meta: PaginationMeta{
			Total:  len(filtered),
			Limit:  limit,
			Offset: offset,
		},
	})
	return nil
}

func (h *Handler) CreateAccessRequest(w http.ResponseWriter, r *http.Request) error {
	user, err := currentUserOrError(r.Context())
	if err != nil {
		return err
	}

	var req struct {
		DatasetID uint64 `json:"dataset_id"`
	}
	if err := decodeJSON(r, &req); err != nil {
		return err
	}
	if req.DatasetID == 0 {
		return &dto.BadRequestError{Message: "dataset_id is required"}
	}

	cmd := services.RequestAccessCmd{
		DatasetID: req.DatasetID,
		UserID:    user.ID,
	}
	if err := h.access.Request(r.Context(), cmd); err != nil {
		return err
	}

	// Получаем созданный запрос
	accessReq, err := h.access.Find(r.Context(), req.DatasetID, user.ID)
	if err != nil {
		return err
	}
	if accessReq == nil {
		return services.ErrRequestNotFound
	}

	writeJSON(w, http.StatusCreated, toAccessRequestResponse(accessReq))
	return nil
}

func (h *Handler) GetAccessRequest(w http.ResponseWriter, r *http.Request) error {
	id, err := parseIDParam(chi.URLParam(r, "requestId"))
	if err != nil {
		return err
	}

	user, err := currentUserOrError(r.Context())
	if err != nil {
		return err
	}

	req, err := h.access.FindByRequestID(r.Context(), id)
	if err != nil {
		return err
	}

	// Проверяем права доступа
	ds, err := h.datasets.GetDataset(r.Context(), req.DatasetID)
	if err != nil {
		return err
	}

	// Доступен владельцу датасета, автору запроса или администратору
	if ds.OwnerID != user.ID && req.UserID != user.ID && !isAdmin(user) {
		return services.ErrRequestForbidden
	}

	writeJSON(w, http.StatusOK, toAccessRequestResponse(req))
	return nil
}

func (h *Handler) UpdateAccessRequest(w http.ResponseWriter, r *http.Request) error {
	id, err := parseIDParam(chi.URLParam(r, "requestId"))
	if err != nil {
		return err
	}

	user, err := currentUserOrError(r.Context())
	if err != nil {
		return err
	}

	status, err := parseAccessStatusUpdate(r)
	if err != nil {
		return err
	}

	accessReq, err := h.access.FindByRequestID(r.Context(), id)
	if err != nil {
		return err
	}

	if err := h.ensureAccessUpdateAllowed(r.Context(), user, accessReq.DatasetID); err != nil {
		return err
	}

	if err := h.applyAccessStatus(r.Context(), id, user.ID, status); err != nil {
		return err
	}

	updated, err := h.access.FindByRequestID(r.Context(), id)
	if err != nil {
		return err
	}

	writeJSON(w, http.StatusOK, toAccessRequestResponse(updated))
	return nil
}

func toAccessRequestResponse(req *entities.AccessRequest) AccessRequestResponse {
	return AccessRequestResponse{
		ID:        req.ID,
		DatasetID: req.DatasetID,
		UserID:    req.UserID,
		Status:    string(req.Status),
		CreatedAt: req.CreatedAt,
	}
}

type accessRequestFilters struct {
	datasetID *uint64
	userID    *uint64
	status    string
	ownedOnly bool
	isAdmin   bool
}

func parseAccessRequestFilters(r *http.Request, user *entities.User) (accessRequestFilters, error) {
	q := r.URL.Query()
	datasetID, err := parseUintPtr(q, "dataset_id")
	if err != nil {
		return accessRequestFilters{}, err
	}
	userIDFilter, err := parseUintPtr(q, "user_id")
	if err != nil {
		return accessRequestFilters{}, err
	}
	status := strings.TrimSpace(q.Get("status"))
	if status != "" && !validAccessStatus(status) {
		return accessRequestFilters{}, &dto.BadRequestError{Message: "invalid status"}
	}

	ownedOnly, err := parseOwnedOnlyFlag(q.Get("owned_only"))
	if err != nil {
		return accessRequestFilters{}, err
	}
	isAdmin := isAdmin(user)
	if !isAdmin {
		ownedOnly = true
	}
	if userIDFilter != nil && !isAdmin && *userIDFilter != user.ID {
		return accessRequestFilters{}, services.ErrRequestForbidden
	}

	return accessRequestFilters{
		datasetID: datasetID,
		userID:    userIDFilter,
		status:    status,
		ownedOnly: ownedOnly,
		isAdmin:   isAdmin,
	}, nil
}

func validAccessStatus(status string) bool {
	switch status {
	case string(entities.AccessStatusPending), string(entities.AccessStatusApproved), string(entities.AccessStatusDenied):
		return true
	default:
		return false
	}
}

func parseOwnedOnlyFlag(raw string) (bool, error) {
	if raw == "" {
		return true, nil
	}
	val, err := strconv.ParseBool(raw)
	if err != nil {
		return false, &dto.BadRequestError{Message: "owned_only must be boolean"}
	}
	return val, nil
}

func (h *Handler) datasetIDsForAccessRequests(ctx context.Context, user *entities.User, filters accessRequestFilters) ([]uint64, error) {
	if filters.datasetID != nil {
		return h.datasetIDsFromSingle(ctx, user, *filters.datasetID)
	}
	if filters.ownedOnly {
		return h.datasetIDsByOwner(ctx, user.ID)
	}
	return h.datasetIDsByOwner(ctx, 0)
}

func (h *Handler) datasetIDsFromSingle(ctx context.Context, user *entities.User, datasetID uint64) ([]uint64, error) {
	ds, err := h.datasets.GetDataset(ctx, datasetID)
	if err != nil {
		return nil, err
	}
	if err := ensureDatasetOwnerOrAdmin(user, ds.OwnerID); err != nil {
		return nil, err
	}
	return []uint64{ds.ID}, nil
}

func (h *Handler) datasetIDsByOwner(ctx context.Context, ownerID uint64) ([]uint64, error) {
	var ownerPtr *uint64
	if ownerID != 0 {
		ownerPtr = &ownerID
	}
	datasets, err := h.datasets.ListDatasets(ctx, false, ownerPtr)
	if err != nil {
		return nil, err
	}
	result := make([]uint64, 0, len(datasets))
	for _, ds := range datasets {
		result = append(result, ds.ID)
	}
	return result, nil
}

func (h *Handler) collectAccessRequests(ctx context.Context, datasetIDs []uint64) ([]*entities.AccessRequest, error) {
	requests := make([]*entities.AccessRequest, 0)
	seen := make(map[uint64]struct{})
	for _, dsID := range datasetIDs {
		reqs, err := h.access.ListByDatasetID(ctx, dsID)
		if err != nil {
			return nil, err
		}
		for _, req := range reqs {
			if _, ok := seen[req.ID]; ok {
				continue
			}
			seen[req.ID] = struct{}{}
			requests = append(requests, req)
		}
	}
	return requests, nil
}

func filterAccessRequests(requests []*entities.AccessRequest, filters accessRequestFilters) []*entities.AccessRequest {
	filtered := make([]*entities.AccessRequest, 0, len(requests))
	for _, req := range requests {
		if filters.userID != nil && req.UserID != *filters.userID {
			continue
		}
		if filters.status != "" && string(req.Status) != filters.status {
			continue
		}
		filtered = append(filtered, req)
	}
	return filtered
}

func paginateAccessRequests(requests []*entities.AccessRequest, limit, offset int) []AccessRequestResponse {
	total := len(requests)
	start := clamp(offset, 0, total)
	end := clamp(offset+limit, start, total)

	items := make([]AccessRequestResponse, 0, end-start)
	for _, req := range requests[start:end] {
		items = append(items, toAccessRequestResponse(req))
	}
	return items
}

func parseAccessStatusUpdate(r *http.Request) (string, error) {
	var req struct {
		Status string `json:"status"`
	}
	if err := decodeJSON(r, &req); err != nil {
		return "", err
	}
	validStatuses := map[string]bool{
		string(entities.AccessStatusApproved): true,
		string(entities.AccessStatusDenied):   true,
	}
	if !validStatuses[req.Status] {
		return "", &dto.BadRequestError{Message: "status must be 'approved' or 'denied'"}
	}
	return req.Status, nil
}

func (h *Handler) ensureAccessUpdateAllowed(ctx context.Context, user *entities.User, datasetID uint64) error {
	ds, err := h.datasets.GetDataset(ctx, datasetID)
	if err != nil {
		return err
	}
	if ds.OwnerID != user.ID && !isAdmin(user) {
		return services.ErrRequestForbidden
	}
	return nil
}

func (h *Handler) applyAccessStatus(ctx context.Context, requestID, actorID uint64, status string) error {
	switch status {
	case string(entities.AccessStatusApproved):
		return h.access.Approve(ctx, requestID, actorID)
	case string(entities.AccessStatusDenied):
		return h.access.Deny(ctx, requestID, actorID)
	default:
		return &dto.BadRequestError{Message: "unknown status"}
	}
}
