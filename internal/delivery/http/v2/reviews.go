package v2

import (
	"context"
	"net/http"

	chi "github.com/go-chi/chi/v5"

	"ppo/internal/delivery/http/dto"
	"ppo/internal/entities"
	"ppo/internal/services"
)

func (h *Handler) ListReviews(w http.ResponseWriter, r *http.Request) error {
	limit, offset, err := parsePagination(r, 50, 200)
	if err != nil {
		return err
	}
	filters, err := parseReviewFilters(r)
	if err != nil {
		return err
	}

	list, err := h.loadReviews(r.Context(), filters)
	if err != nil {
		return err
	}

	filtered := filterReviews(list, filters.datasetID, filters.userID, filters.minRating, filters.maxRating)
	writeJSON(w, http.StatusOK, paginateReviews(filtered, limit, offset))
	return nil
}

func (h *Handler) CreateReview(w http.ResponseWriter, r *http.Request) error {
	// Согласно спецификации, отзыв создается от имени текущего аутентифицированного пользователя
	user, err := currentUserOrError(r.Context())
	if err != nil {
		return err
	}

	var req struct {
		DatasetID uint64 `json:"dataset_id"`
		Rating    int    `json:"rating"`
		Text      string `json:"text"`
	}
	if err := decodeJSON(r, &req); err != nil {
		return err
	}
	if req.DatasetID == 0 {
		return &dto.BadRequestError{Message: "dataset_id is required"}
	}
	if req.Rating < int(entities.Rating1) || req.Rating > int(entities.Rating5) {
		return &dto.BadRequestError{Message: "rating must be between 1 and 5"}
	}

	cmd := services.CreateReviewCmd{
		UserID:    user.ID,
		DatasetID: req.DatasetID,
		Rating:    entities.Rating(req.Rating),
		Text:      req.Text,
	}
	id, err := h.reviews.CreateReview(r.Context(), cmd)
	if err != nil {
		return err
	}
	review, err := h.reviews.GetReviewByID(r.Context(), id)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusCreated, toReviewResponse(review))
	return nil
}

func (h *Handler) GetReview(w http.ResponseWriter, r *http.Request) error {
	id, err := parseIDParam(chi.URLParam(r, "reviewId"))
	if err != nil {
		return err
	}
	review, err := h.reviews.GetReviewByID(r.Context(), id)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, toReviewResponse(review))
	return nil
}

func (h *Handler) UpdateReview(w http.ResponseWriter, r *http.Request) error {
	id, err := parseIDParam(chi.URLParam(r, "reviewId"))
	if err != nil {
		return err
	}
	req, err := decodeUpdateReviewRequest(r)
	if err != nil {
		return err
	}

	current, err := h.reviews.GetReviewByID(r.Context(), id)
	if err != nil {
		return err
	}

	if err := ensureReviewAccess(r.Context(), current.UserID); err != nil {
		return err
	}

	cmd, err := buildUpdateReviewCmd(id, current, req)
	if err != nil {
		return err
	}
	if err := h.reviews.UpdateReview(r.Context(), cmd); err != nil {
		return err
	}
	updated, err := h.reviews.GetReviewByID(r.Context(), id)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, toReviewResponse(updated))
	return nil
}

func (h *Handler) DeleteReview(w http.ResponseWriter, r *http.Request) error {
	id, err := parseIDParam(chi.URLParam(r, "reviewId"))
	if err != nil {
		return err
	}

	// Проверяем права доступа: только автор отзыва или администратор
	current, err := h.reviews.GetReviewByID(r.Context(), id)
	if err != nil {
		return err
	}
	user, err := currentUserOrError(r.Context())
	if err != nil {
		return err
	}
	if current.UserID != user.ID && !isAdmin(user) {
		return services.ErrRequestForbidden
	}

	if err := h.reviews.DeleteReview(r.Context(), id); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (h *Handler) ListUserReviews(w http.ResponseWriter, r *http.Request) error {
	userID, err := parseIDParam(chi.URLParam(r, "userId"))
	if err != nil {
		return err
	}
	// Доступно самому пользователю и администраторам
	if _, err := ensureUserOrAdmin(r.Context(), userID); err != nil {
		return err
	}

	list, err := h.reviews.ListByUser(r.Context(), userID)
	if err != nil {
		return err
	}
	limit, offset, err := parsePagination(r, 50, 200)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, paginateReviews(toReviewResponses(list), limit, offset))
	return nil
}

type reviewFilters struct {
	datasetID *uint64
	userID    *uint64
	minRating *float64
	maxRating *float64
}

func parseReviewFilters(r *http.Request) (reviewFilters, error) {
	q := r.URL.Query()
	datasetID, err := parseUintPtr(q, "dataset_id")
	if err != nil {
		return reviewFilters{}, err
	}
	userID, err := parseUintPtr(q, "user_id")
	if err != nil {
		return reviewFilters{}, err
	}
	minRating, err := parseRatingBound(q, "min_rating")
	if err != nil {
		return reviewFilters{}, err
	}
	maxRating, err := parseRatingBound(q, "max_rating")
	if err != nil {
		return reviewFilters{}, err
	}
	if minRating != nil && maxRating != nil && *minRating > *maxRating {
		return reviewFilters{}, &dto.BadRequestError{Message: "min_rating cannot exceed max_rating"}
	}

	return reviewFilters{
		datasetID: datasetID,
		userID:    userID,
		minRating: minRating,
		maxRating: maxRating,
	}, nil
}

func (h *Handler) loadReviews(ctx context.Context, filters reviewFilters) ([]*entities.Review, error) {
	switch {
	case filters.datasetID != nil:
		return h.reviews.ListByDataset(ctx, *filters.datasetID)
	case filters.userID != nil:
		return h.reviews.ListByUser(ctx, *filters.userID)
	default:
		return h.collectPublicReviews(ctx)
	}
}

func (h *Handler) collectPublicReviews(ctx context.Context) ([]*entities.Review, error) {
	publicDatasets, err := h.datasets.ListDatasets(ctx, true, nil)
	if err != nil {
		return nil, err
	}
	allReviews := make([]*entities.Review, 0)
	seenReviews := make(map[uint64]bool)
	for _, ds := range publicDatasets {
		reviews, listErr := h.reviews.ListByDataset(ctx, ds.ID)
		if listErr != nil {
			continue
		}
		for _, rev := range reviews {
			if !seenReviews[rev.ID] {
				allReviews = append(allReviews, rev)
				seenReviews[rev.ID] = true
			}
		}
	}
	return allReviews, nil
}

func filterReviews(list []*entities.Review, datasetID, userID *uint64, minRating, maxRating *float64) []ReviewResponse {
	res := make([]ReviewResponse, 0, len(list))
	for _, rev := range list {
		if datasetID != nil && rev.DatasetID != *datasetID {
			continue
		}
		if userID != nil && rev.UserID != *userID {
			continue
		}
		if minRating != nil && float64(rev.Rating) < *minRating {
			continue
		}
		if maxRating != nil && float64(rev.Rating) > *maxRating {
			continue
		}
		res = append(res, toReviewResponse(rev))
	}
	return res
}

func paginateReviews(list []ReviewResponse, limit, offset int) ReviewsResponse {
	total := len(list)
	start := clamp(offset, 0, total)
	end := clamp(offset+limit, start, total)
	return ReviewsResponse{
		Items: list[start:end],
		Meta: PaginationMeta{
			Total:  total,
			Limit:  limit,
			Offset: offset,
		},
	}
}

func toReviewResponses(list []*entities.Review) []ReviewResponse {
	res := make([]ReviewResponse, 0, len(list))
	for _, rev := range list {
		res = append(res, toReviewResponse(rev))
	}
	return res
}

type updateReviewRequest struct {
	Rating *int    `json:"rating"`
	Text   *string `json:"text"`
}

func decodeUpdateReviewRequest(r *http.Request) (updateReviewRequest, error) {
	var req updateReviewRequest
	if err := decodeJSON(r, &req); err != nil {
		return updateReviewRequest{}, err
	}
	if req.Rating == nil && req.Text == nil {
		return updateReviewRequest{}, &dto.BadRequestError{Message: "nothing to update"}
	}
	if req.Rating != nil && (*req.Rating < int(entities.Rating1) || *req.Rating > int(entities.Rating5)) {
		return updateReviewRequest{}, &dto.BadRequestError{Message: "rating must be between 1 and 5"}
	}
	return req, nil
}

func ensureReviewAccess(ctx context.Context, authorID uint64) error {
	user, err := currentUserOrError(ctx)
	if err != nil {
		return err
	}
	if user.ID != authorID && !isAdmin(user) {
		return services.ErrRequestForbidden
	}
	return nil
}

func buildUpdateReviewCmd(id uint64, current *entities.Review, req updateReviewRequest) (services.UpdateReviewCmd, error) {
	rating := current.Rating
	if req.Rating != nil {
		rating = entities.Rating(*req.Rating)
	}
	text := current.Text
	if req.Text != nil {
		text = *req.Text
	}
	return services.UpdateReviewCmd{
		ReviewID: id,
		Rating:   rating,
		Text:     text,
	}, nil
}

func toReviewResponse(rev *entities.Review) ReviewResponse {
	return ReviewResponse{
		ID:        rev.ID,
		DatasetID: rev.DatasetID,
		UserID:    rev.UserID,
		Rating:    int(rev.Rating),
		Text:      rev.Text,
		CreatedAt: rev.CreatedAt,
	}
}
