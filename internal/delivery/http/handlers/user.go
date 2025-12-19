package handlers

import (
	"net/http"
	"strconv"

	chi "github.com/go-chi/chi/v5"
	"ppo/internal/delivery/http/dto"
	"ppo/internal/delivery/http/middleware"
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
func RegisterUserHandler(svc services.UserService) middleware.HandlerWithError {
	return func(w http.ResponseWriter, r *http.Request) error {
		var req dto.RegisterUserRequest
		if err := dto.DecodeJSON(r.Body, &req); err != nil {
			return &dto.BadRequestError{Message: "invalid JSON payload"}
		}

		id, err := svc.Register(r.Context(), req.ToCommand())
		if err != nil {
			return err // ErrUserExists ⇒ 409, ErrInvalidPassword ⇒ 400, др. ⇒ 500
		}

		user, err := svc.GetUserByID(r.Context(), id)
		if err != nil {
			return err // обычно не случится, но маппинг на 500, если что
		}

		dto.WriteJSON(w, http.StatusCreated, dto.FromEntityUser(user))
		return nil
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
func AuthenticateUserHandler(svc services.UserService) middleware.HandlerWithError {
	return func(w http.ResponseWriter, r *http.Request) error {
		var req dto.AuthenticateUserRequest
		if err := dto.DecodeJSON(r.Body, &req); err != nil {
			return &dto.BadRequestError{Message: "invalid JSON payload"}
		}

		user, err := svc.Authenticate(r.Context(), req.ToCommand())
		if err != nil {
			return err // ErrInvalidCredentials ⇒ 401, ErrUserNotFound ⇒ 404? (в зависимости от mapErrorToStatus)
		}

		resp := dto.AuthenticateResponse{
			User:  dto.FromEntityUser(user),
			Token: "",
		}
		dto.WriteJSON(w, http.StatusOK, resp)
		return nil
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
func GetUserHandler(svc services.UserService) middleware.HandlerWithError {
	return func(w http.ResponseWriter, r *http.Request) error {
		idParam := chi.URLParam(r, "id")
		id, err := strconv.ParseUint(idParam, 10, 64)
		if err != nil {
			return &dto.BadRequestError{Message: "invalid user ID"}
		}

		user, err := svc.GetUserByID(r.Context(), id)
		if err != nil {
			return err // ErrUserNotFound ⇒ 404, иначе 500
		}

		dto.WriteJSON(w, http.StatusOK, dto.FromEntityUser(user))
		return nil
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
func UpdateUserHandler(svc services.UserService) middleware.HandlerWithError {
	return func(w http.ResponseWriter, r *http.Request) error {
		idParam := chi.URLParam(r, "id")
		id, err := strconv.ParseUint(idParam, 10, 64)
		if err != nil {
			return &dto.BadRequestError{Message: "invalid user ID"}
		}

		var req dto.UpdateUserRequest
		if err := dto.DecodeJSON(r.Body, &req); err != nil {
			return &dto.BadRequestError{Message: "invalid JSON payload"}
		}

		if err := svc.UpdateUser(r.Context(), req.ToCommand(id)); err != nil {
			return err // ErrUserNotFound ⇒ 404, ErrUserExists ⇒ 409, и т. д.
		}

		w.WriteHeader(http.StatusNoContent)
		return nil
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
func DeleteUserHandler(svc services.UserService) middleware.HandlerWithError {
	return func(w http.ResponseWriter, r *http.Request) error {
		idParam := chi.URLParam(r, "id")
		id, err := strconv.ParseUint(idParam, 10, 64)
		if err != nil {
			return &dto.BadRequestError{Message: "invalid user ID"}
		}

		if err := svc.DeleteUser(r.Context(), id); err != nil {
			return err // ErrUserNotFound ⇒ 404, иначе 500
		}

		w.WriteHeader(http.StatusNoContent)
		return nil
	}
}
