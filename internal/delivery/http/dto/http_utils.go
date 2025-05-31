package dto

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"ppo/internal/services"
)

func WriteJSON(w http.ResponseWriter, status int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func DecodeJSON(r io.Reader, v interface{}) error {
	dec := json.NewDecoder(r)
	dec.DisallowUnknownFields()
	if err := dec.Decode(v); err != nil {
		return &BadRequestError{Message: fmt.Sprintf("invalid JSON: %v", err)}
	}
	return nil
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type BadRequestError struct{ Message string }

func (e *BadRequestError) Error() string { return e.Message }

func WriteError(w http.ResponseWriter, err error) {
	switch {
	case isBadRequest(err):
		WriteJSON(w, http.StatusBadRequest, ErrorResponse{Error: err.Error()})
	case isNotFound(err):
		WriteJSON(w, http.StatusNotFound, ErrorResponse{Error: err.Error()})
	case isConflict(err):
		WriteJSON(w, http.StatusConflict, ErrorResponse{Error: err.Error()})
	default:
		WriteJSON(w, http.StatusInternalServerError, ErrorResponse{Error: err.Error()})
	}
}

func ToCategoryList(list []CategoryResponse) map[string]interface{} {
	return map[string]interface{}{"categories": list}
}

func isBadRequest(err error) bool {
	_, ok := err.(*BadRequestError)
	return ok
}

func isNotFound(err error) bool {
	return errorsIs(err, services.ErrReviewNotFound) ||
		errorsIs(err, services.ErrDatasetNotFound) ||
		errorsIs(err, services.ErrUserNotFound)
}

func isConflict(err error) bool {
	return errorsIs(err, services.ErrInvalidRating)
}

func WriteStatusError(w http.ResponseWriter, status int, err error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(ErrorResponse{
		Error: err.Error(),
	})
}
