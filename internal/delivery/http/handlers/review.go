package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"ppo/internal/delivery/http/dto"
	"ppo/internal/delivery/http/middleware"
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
func CreateReviewHandler(svc services.ReviewService) middleware.HandlerWithError {
	return func(w http.ResponseWriter, r *http.Request) error {
		var req dto.CreateReviewRequest
		// 1) Прочитать тело JSON
		if err := dto.DecodeJSON(r.Body, &req); err != nil {
			return err // mapErrorToStatus → 400 или 500
		}

		// 2) Преобразовать в команду
		cmd := req.ToCommand()

		// 3) Вызвать сервис
		id, err := svc.CreateReview(r.Context(), cmd)
		if err != nil {
			return err // может быть ErrInvalidRating→400, или прочие→500
		}

		// 4) После успешного создания, получить сам объект (т. е. Review) по ID
		review, err := svc.GetReviewByID(r.Context(), id)
		if err != nil {
			return err // если вдруг не найден→404, иначе→500
		}

		// 5) Отправить JSON 201
		resp := dto.FromEntityReview(review)
		dto.WriteJSON(w, http.StatusCreated, resp)
		return nil
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
func UpdateReviewHandler(svc services.ReviewService) middleware.HandlerWithError {
	return func(w http.ResponseWriter, r *http.Request) error {
		// 1) Парсим review ID из URL
		idParam := chi.URLParam(r, "id")
		id, err := strconv.ParseUint(idParam, 10, 64)
		if err != nil {
			return &dto.BadRequestError{Message: "invalid review ID"}
		}

		// 2) Читаем тело JSON → dto.UpdateReviewRequest
		var req dto.UpdateReviewRequest
		if err := dto.DecodeJSON(r.Body, &req); err != nil {
			return err
		}

		// 3) Собираем команду
		cmd := req.ToCommand(id)

		// 4) Вызов бизнес-логики
		if err := svc.UpdateReview(r.Context(), cmd); err != nil {
			return err // ErrReviewNotFound→404, ErrInvalidRating→400, иначе→500
		}

		// 5) Если всё ок, просто 204 No Content
		w.WriteHeader(http.StatusNoContent)
		return nil
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
func DeleteReviewHandler(svc services.ReviewService) middleware.HandlerWithError {
	return func(w http.ResponseWriter, r *http.Request) error {
		// 1) Парсим review ID из URL
		idParam := chi.URLParam(r, "id")
		id, err := strconv.ParseUint(idParam, 10, 64)
		if err != nil {
			return &dto.BadRequestError{Message: "invalid review ID"}
		}

		// 2) Вызов бизнес-логики
		if err := svc.DeleteReview(r.Context(), id); err != nil {
			return err // ErrReviewNotFound→404, иначе→500
		}

		// 3) Если всё прошло успешно → 204 No Content
		w.WriteHeader(http.StatusNoContent)
		return nil
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
func GetReviewByIDHandler(svc services.ReviewService) middleware.HandlerWithError {
	return func(w http.ResponseWriter, r *http.Request) error {
		// 1) Парсим review ID
		idParam := chi.URLParam(r, "id")
		id, err := strconv.ParseUint(idParam, 10, 64)
		if err != nil {
			return &dto.BadRequestError{Message: "invalid review ID"}
		}

		// 2) Вызов сервиса
		review, err := svc.GetReviewByID(r.Context(), id)
		if err != nil {
			return err // ErrReviewNotFound→404, иначе→500
		}

		// 3) Возвращаем JSON{ ... }
		dto.WriteJSON(w, http.StatusOK, dto.FromEntityReview(review))
		return nil
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
func ListReviewsByDatasetHandler(svc services.ReviewService) middleware.HandlerWithError {
	return func(w http.ResponseWriter, r *http.Request) error {
		// 1) Парсим dataset ID
		idParam := chi.URLParam(r, "id")
		datasetID, err := strconv.ParseUint(idParam, 10, 64)
		if err != nil {
			return &dto.BadRequestError{Message: "invalid dataset ID"}
		}

		// 2) Сервис: взять все отзывы по datasetID
		reviews, err := svc.ListByDataset(r.Context(), datasetID)
		if err != nil {
			return err
		}

		// 3) Формируем JSON-массив
		dto.WriteJSON(w, http.StatusOK, dto.FromEntityReviewList(reviews))
		return nil
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
func ListReviewsByUserHandler(svc services.ReviewService) middleware.HandlerWithError {
	return func(w http.ResponseWriter, r *http.Request) error {
		// 1) Парсим user ID
		idParam := chi.URLParam(r, "id")
		userID, err := strconv.ParseUint(idParam, 10, 64)
		if err != nil {
			return &dto.BadRequestError{Message: "invalid user ID"}
		}

		// 2) Сервис: взять все отзывы по userID
		reviews, err := svc.ListByUser(r.Context(), userID)
		if err != nil {
			return err
		}

		// 3) JSON-массив
		dto.WriteJSON(w, http.StatusOK, dto.FromEntityReviewList(reviews))
		return nil
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
func GetRatingSummaryHandler(svc services.ReviewService) middleware.HandlerWithError {
	return func(w http.ResponseWriter, r *http.Request) error {
		// 1) Парсим dataset ID
		idParam := chi.URLParam(r, "id")
		datasetID, err := strconv.ParseUint(idParam, 10, 64)
		if err != nil {
			return &dto.BadRequestError{Message: "invalid dataset ID"}
		}

		// 2) Запрос в сервис
		summary, err := svc.GetRatingSummary(r.Context(), datasetID)
		if err != nil {
			return err
		}

		// 3) Формируем JSON ответа
		resp := dto.FromServiceSummary(summary)
		dto.WriteJSON(w, http.StatusOK, resp)
		return nil
	}
}
