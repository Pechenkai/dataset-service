package handlers

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"ppo/internal/delivery/http/dto"
	"ppo/internal/services"
)

// @Summary      Register new user
// @Description  Регистрация нового пользователя
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        body  body      dto.RegisterUserRequest  true  "Username, Email, Password, Country, Role"
// @Success      201   {object}  dto.UserResponse
// @Failure      400   {object}  dto.ErrorResponse
// @Failure      409   {object}  dto.ErrorResponse  "email already exists"
// @Failure      500   {object}  dto.ErrorResponse
// @Router       /users/register [post]
func RegisterUserHandler(svc services.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.RegisterUserRequest
		if err := dto.DecodeJSON(r.Body, &req); err != nil {
			dto.WriteError(w, err)
			return
		}

		id, err := svc.Register(r.Context(), req.ToCommand())
		if err != nil {
			dto.WriteError(w, err)
			return
		}

		user, err := svc.GetUserByID(r.Context(), id)
		if err != nil {
			dto.WriteError(w, err)
			return
		}

		dto.WriteJSON(w, http.StatusCreated, dto.FromEntityUser(user))
	}
}

// @Summary      Authenticate user
// @Description  Аутентификация: email + password
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        body  body      dto.AuthenticateUserRequest  true  "Email and Password"
// @Success      200   {object}  dto.AuthenticateResponse
// @Failure      400   {object}  dto.ErrorResponse
// @Failure      401   {object}  dto.ErrorResponse  "invalid credentials"
// @Failure      500   {object}  dto.ErrorResponse
// @Router       /users/authenticate [post]
func AuthenticateUserHandler(svc services.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req dto.AuthenticateUserRequest
		if err := dto.DecodeJSON(r.Body, &req); err != nil {
			dto.WriteError(w, err)
			return
		}

		user, err := svc.Authenticate(r.Context(), req.ToCommand())
		if err != nil {
			dto.WriteError(w, err)
			return
		}

		resp := dto.AuthenticateResponse{User: dto.FromEntityUser(user), Token: ""} // токен можно добавить позже
		dto.WriteJSON(w, http.StatusOK, resp)
	}
}

// @Summary      Get user by ID
// @Description  Возвращает данные пользователя по ID
// @Tags         users
// @Produce      json
// @Param        id   path      int  true  "User ID"
// @Success      200  {object}  dto.UserResponse
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /users/{id} [get]
func GetUserHandler(svc services.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "id")
		id, err := strconv.ParseUint(idParam, 10, 64)
		if err != nil {
			dto.WriteError(w, &dto.BadRequestError{Message: "invalid user ID"})
			return
		}

		user, err := svc.GetUserByID(r.Context(), id)
		if err != nil {
			dto.WriteError(w, err)
			return
		}

		dto.WriteJSON(w, http.StatusOK, dto.FromEntityUser(user))
	}
}

// @Summary      Update user
// @Description  Обновляет данные пользователя
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        id    path      int                       true  "User ID"
// @Param        body  body      dto.UpdateUserRequest     true  "Новые данные пользователя"
// @Success      204   {object}  nil
// @Failure      400   {object}  dto.ErrorResponse
// @Failure      404   {object}  dto.ErrorResponse
// @Failure      500   {object}  dto.ErrorResponse
// @Router       /users/{id} [put]
func UpdateUserHandler(svc services.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "id")
		id, err := strconv.ParseUint(idParam, 10, 64)
		if err != nil {
			dto.WriteError(w, &dto.BadRequestError{Message: "invalid user ID"})
			return
		}

		var req dto.UpdateUserRequest
		if err := dto.DecodeJSON(r.Body, &req); err != nil {
			dto.WriteError(w, err)
			return
		}

		if err := svc.UpdateUser(r.Context(), req.ToCommand(id)); err != nil {
			dto.WriteError(w, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

// @Summary      Delete user
// @Description  Удаляет пользователя по ID
// @Tags         users
// @Produce      json
// @Param        id   path      int  true  "User ID"
// @Success      204  {object}  nil
// @Failure      400  {object}  dto.ErrorResponse
// @Failure      404  {object}  dto.ErrorResponse
// @Failure      500  {object}  dto.ErrorResponse
// @Router       /users/{id} [delete]
func DeleteUserHandler(svc services.UserService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "id")
		id, err := strconv.ParseUint(idParam, 10, 64)
		if err != nil {
			dto.WriteError(w, &dto.BadRequestError{Message: "invalid user ID"})
			return
		}

		if err := svc.DeleteUser(r.Context(), id); err != nil {
			dto.WriteError(w, err)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
