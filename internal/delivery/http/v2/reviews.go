package v2

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"ppo/internal/delivery/http/dto"
	"ppo/internal/entities"
	"ppo/internal/services"
)

func (h *Handler) ListReviews(w http.ResponseWriter, r *http.Request) error {
	limit, offset, err := parsePagination(r, 50, 200)
	if err != nil {
		return err
	}
	q := r.URL.Query()
	datasetID, err := parseUintPtr(q, "dataset_id")
	if err != nil {
		return err
	}
	userID, err := parseUintPtr(q, "user_id")
	if err != nil {
		return err
	}
	if datasetID == nil && userID == nil {
		current, err := currentUserOrError(r.Context())
		if err != nil {
			return err
		}
		id := current.ID
		userID = &id
	}
	minRating, err := parseRatingBound(q, "min_rating")
	if err != nil {
		return err
	}
	maxRating, err := parseRatingBound(q, "max_rating")
	if err != nil {
		return err
	}
	if minRating != nil && maxRating != nil && *minRating > *maxRating {
		return &dto.BadRequestError{Message: "min_rating cannot exceed max_rating"}
	}

	var list []*entities.Review
	if datasetID != nil {
		list, err = h.reviews.ListByDataset(r.Context(), *datasetID)
	} else {
		list, err = h.reviews.ListByUser(r.Context(), *userID)
	}
	if err != nil {
		return err
	}
	filtered := filterReviews(list, datasetID, userID, minRating, maxRating)
	writeJSON(w, http.StatusOK, paginateReviews(filtered, limit, offset))
	return nil
}

func (h *Handler) CreateReview(w http.ResponseWriter, r *http.Request) error {
	var req struct {
		UserID    uint64 `json:"user_id"`
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
	if req.UserID == 0 {
		user, err := currentUserOrError(r.Context())
		if err != nil {
			return err
		}
		req.UserID = user.ID
	}
	if req.Rating < int(entities.Rating1) || req.Rating > int(entities.Rating5) {
		return &dto.BadRequestError{Message: "rating must be between 1 and 5"}
	}

	cmd := services.CreateReviewCmd{
		UserID:    req.UserID,
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
	var req struct {
		Rating *int    `json:"rating"`
		Text   *string `json:"text"`
	}
	if err := decodeJSON(r, &req); err != nil {
		return err
	}
	if req.Rating == nil && req.Text == nil {
		return &dto.BadRequestError{Message: "nothing to update"}
	}

	current, err := h.reviews.GetReviewByID(r.Context(), id)
	if err != nil {
		return err
	}

	rating := current.Rating
	if req.Rating != nil {
		if *req.Rating < int(entities.Rating1) || *req.Rating > int(entities.Rating5) {
			return &dto.BadRequestError{Message: "rating must be between 1 and 5"}
		}
		rating = entities.Rating(*req.Rating)
	}
	text := current.Text
	if req.Text != nil {
		text = *req.Text
	}

	cmd := services.UpdateReviewCmd{
		ReviewID: id,
		Rating:   rating,
		Text:     text,
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
