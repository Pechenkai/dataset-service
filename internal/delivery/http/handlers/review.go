package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"ppo/internal/delivery/http/dto"
	"ppo/internal/services"
)

// @Summary      Create review
// @Description  Создает новый отзыв (rating + text) от пользователя для датасета.
// @Tags         reviews
// @Accept       json
// @Produce      json
// @Param        body  body      dto.CreateReviewRequest  true  "UserID, DatasetID, Rating, Text"
// @Success      201   {object}  dto.ReviewResponse
// @Failure      400   {object}  dto.ErrorResponse
// @Failure      404   {object}  dto.ErrorResponse
// @Failure      500   {object}  dto.ErrorResponse
// @Router       /reviews [post]
func CreateReviewHandler(svc services.ReviewService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.CreateReviewRequest
		if err := dto.DecodeJSON(r.Body, &req); err != nil {
			dto.WriteError(w, err)
			return
		}

		cmd := req.ToCommand()

		id, err := svc.CreateReview(r.Context(), cmd)
		if err != nil {
			dto.WriteError(w, err)
			return
		}

		review, err := svc.GetReviewByID(r.Context(), id)
		if err != nil {
			dto.WriteError(w, err)
			return
		}

		resp := dto.FromEntityReview(review)
		dto.WriteJSON(w, http.StatusCreated, resp)
	}
}

// @Summary      Update review
// @Description  Обновляет рейтинг и текст существующего отзыва.
// @Tags         reviews
// @Accept       json
// @Produce      json
// @Param        id    path      int                      true  "Review ID"
// @Param        body  body      dto.UpdateReviewRequest  true  "New rating & text"
// @Success      204  {object}  nil
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /reviews/{id} [put]
func UpdateReviewHandler(svc services.ReviewService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "id")
		id, err := strconv.ParseUint(idParam, 10, 64)
		if err != nil {
			dto.WriteError(w, &dto.BadRequestError{Message: "invalid review ID"})
			return
		}

		var req dto.UpdateReviewRequest
		if err := dto.DecodeJSON(r.Body, &req); err != nil {
			dto.WriteError(w, err)
			return
		}

		cmd := req.ToCommand(id)

		if err := svc.UpdateReview(r.Context(), cmd); err != nil {
			dto.WriteError(w, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// @Summary      Delete review
// @Description  Удаляет отзыв по ID.
// @Tags         reviews
// @Produce      json
// @Param        id   path      int  true  "Review ID"
// @Success      204  {object}  nil
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /reviews/{id} [delete]
func DeleteReviewHandler(svc services.ReviewService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "id")
		id, err := strconv.ParseUint(idParam, 10, 64)
		if err != nil {
			dto.WriteError(w, &dto.BadRequestError{Message: "invalid review ID"})
			return
		}

		if err := svc.DeleteReview(r.Context(), id); err != nil {
			dto.WriteError(w, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// @Summary      Get review by ID
// @Description  Возвращает отзыв по его ID
// @Tags         reviews
// @Produce      json
// @Param        id   path      int  true  "Review ID"
// @Success      200  {object}  dto.ReviewResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /reviews/{id} [get]
func GetReviewByIDHandler(svc services.ReviewService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "id")
		id, err := strconv.ParseUint(idParam, 10, 64)
		if err != nil {
			dto.WriteError(w, &dto.BadRequestError{Message: "invalid review ID"})
			return
		}

		review, err := svc.GetReviewByID(r.Context(), id)
		if err != nil {
			dto.WriteError(w, err)
			return
		}

		dto.WriteJSON(w, http.StatusOK, dto.FromEntityReview(review))
	}
}

// @Summary      List reviews by dataset
// @Description  Возвращает все отзывы для указанного датасета
// @Tags         reviews
// @Produce      json
// @Param        id   path      int  true  "Dataset ID"
// @Success      200  {object}  dto.ReviewsResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /datasets/{id}/reviews [get]
func ListReviewsByDatasetHandler(svc services.ReviewService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "id")
		datasetID, err := strconv.ParseUint(idParam, 10, 64)
		if err != nil {
			dto.WriteError(w, &dto.BadRequestError{Message: "invalid dataset ID"})
			return
		}

		reviews, err := svc.ListByDataset(r.Context(), datasetID)
		if err != nil {
			dto.WriteError(w, err)
			return
		}

		dto.WriteJSON(w, http.StatusOK, dto.FromEntityReviewList(reviews))
	}
}

// @Summary      List reviews by user
// @Description  Возвращает все отзывы, созданные указанным пользователем
// @Tags         reviews
// @Produce      json
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  dto.ReviewsResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /users/{id}/reviews [get]
func ListReviewsByUserHandler(svc services.ReviewService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "id")
		userID, err := strconv.ParseUint(idParam, 10, 64)
		if err != nil {
			dto.WriteError(w, &dto.BadRequestError{Message: "invalid user ID"})
			return
		}

		reviews, err := svc.ListByUser(r.Context(), userID)
		if err != nil {
			dto.WriteError(w, err)
			return
		}

		dto.WriteJSON(w, http.StatusOK, dto.FromEntityReviewList(reviews))
	}
}

// @Summary      Get rating summary
// @Description  Возвращает средний рейтинг и количество отзывов для датасета
// @Tags         reviews
// @Produce      json
// @Param        id   path      int  true  "Dataset ID"
// @Success      200  {object}  dto.RatingSummaryResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /datasets/{id}/reviews/summary [get]
func GetRatingSummaryHandler(svc services.ReviewService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "id")
		datasetID, err := strconv.ParseUint(idParam, 10, 64)
		if err != nil {
			dto.WriteError(w, &dto.BadRequestError{Message: "invalid dataset ID"})
			return
		}

		summary, err := svc.GetRatingSummary(r.Context(), datasetID)
		if err != nil {
			dto.WriteError(w, err)
			return
		}

		resp := dto.FromServiceSummary(summary)
		dto.WriteJSON(w, http.StatusOK, resp)
	}
}
