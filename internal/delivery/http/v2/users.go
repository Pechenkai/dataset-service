package v2

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"ppo/internal/delivery/http/dto"
	"ppo/internal/entities"
	"ppo/internal/services"
)

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) error {
	limit, offset, err := parsePagination(r, 50, 200)
	if err != nil {
		return err
	}
	q := r.URL.Query()
	role := strings.TrimSpace(q.Get("role"))
	email := strings.TrimSpace(q.Get("email"))
	country := strings.TrimSpace(q.Get("country"))
	isBlocked, err := parseBoolPtr(q, "is_blocked")
	if err != nil {
		return err
	}

	users, err := h.users.ListAllUsers(r.Context())
	if err != nil {
		return err
	}

	filtered := make([]UserResponse, 0, len(users))
	for _, user := range users {
		if role != "" && !strings.EqualFold(user.Role, role) {
			continue
		}
		if email != "" && !strings.Contains(strings.ToLower(user.Email), strings.ToLower(email)) {
			continue
		}
		if country != "" && !strings.EqualFold(user.Country, country) {
			continue
		}
		if isBlocked != nil && user.IsBlocked != *isBlocked {
			continue
		}
		filtered = append(filtered, toUserResponse(user))
	}

	total := len(filtered)
	start := clamp(offset, 0, total)
	end := clamp(offset+limit, start, total)
	writeJSON(w, http.StatusOK, UsersResponse{
		Items: filtered[start:end],
		Meta: PaginationMeta{
			Total:  total,
			Limit:  limit,
			Offset: offset,
		},
	})
	return nil
}

func (h *Handler) RegisterUser(w http.ResponseWriter, r *http.Request) error {
	var req services.RegisterUserCmd
	if err := decodeJSON(r, &req); err != nil {
		return err
	}
	if strings.TrimSpace(req.Username) == "" ||
		strings.TrimSpace(req.Email) == "" ||
		strings.TrimSpace(req.Password) == "" ||
		strings.TrimSpace(req.Country) == "" ||
		strings.TrimSpace(req.Role) == "" {
		return &dto.BadRequestError{Message: "all fields are required"}
	}

	id, err := h.users.Register(r.Context(), req)
	if err != nil {
		return err
	}
	user, err := h.users.GetUserByID(r.Context(), id)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusCreated, toUserResponse(user))
	return nil
}

func (h *Handler) GetUser(w http.ResponseWriter, r *http.Request) error {
	id, err := parseIDParam(chi.URLParam(r, "userId"))
	if err != nil {
		return err
	}
	user, err := h.users.GetUserByID(r.Context(), id)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, toUserResponse(user))
	return nil
}

func (h *Handler) UpdateUser(w http.ResponseWriter, r *http.Request) error {
	id, err := parseIDParam(chi.URLParam(r, "userId"))
	if err != nil {
		return err
	}

	var req struct {
		Username  *string `json:"username"`
		Email     *string `json:"email"`
		Password  *string `json:"password"`
		Country   *string `json:"country"`
		Role      *string `json:"role"`
		IsBlocked *bool   `json:"is_blocked"`
	}
	if err := decodeJSON(r, &req); err != nil {
		return err
	}
	if req.Username == nil && req.Email == nil && req.Password == nil && req.Country == nil && req.Role == nil && req.IsBlocked == nil {
		return &dto.BadRequestError{Message: "nothing to update"}
	}

	current, err := h.users.GetUserByID(r.Context(), id)
	if err != nil {
		return err
	}

	cmd := services.UpdateUserCmd{
		ID:        id,
		Username:  current.Username,
		Email:     current.Email,
		Country:   current.Country,
		Role:      current.Role,
		IsBlocked: current.IsBlocked,
	}
	if req.Username != nil {
		cmd.Username = *req.Username
	}
	if req.Email != nil {
		cmd.Email = *req.Email
	}
	if req.Password != nil {
		cmd.Password = *req.Password
	}
	if req.Country != nil {
		cmd.Country = *req.Country
	}
	if req.Role != nil {
		cmd.Role = *req.Role
	}
	if req.IsBlocked != nil {
		cmd.IsBlocked = *req.IsBlocked
	}

	if err := h.users.UpdateUser(r.Context(), cmd); err != nil {
		return err
	}
	updated, err := h.users.GetUserByID(r.Context(), id)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, toUserResponse(updated))
	return nil
}

func (h *Handler) DeleteUser(w http.ResponseWriter, r *http.Request) error {
	id, err := parseIDParam(chi.URLParam(r, "userId"))
	if err != nil {
		return err
	}
	if err := h.users.DeleteUser(r.Context(), id); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func (h *Handler) ListUserSubscriptions(w http.ResponseWriter, r *http.Request) error {
	userID, err := parseIDParam(chi.URLParam(r, "userId"))
	if err != nil {
		return err
	}
	limit, offset, err := parsePagination(r, 50, 200)
	if err != nil {
		return err
	}
	items, err := h.collectSubscriptions(r.Context(), &userID, nil)
	if err != nil {
		return err
	}
	writeJSON(w, http.StatusOK, paginateSubscriptions(items, limit, offset))
	return nil
}

func toUserResponse(user *entities.User) UserResponse {
	return UserResponse{
		ID:               user.ID,
		Username:         user.Username,
		Email:            user.Email,
		Country:          user.Country,
		Role:             user.Role,
		IsBlocked:        user.IsBlocked,
		RegistrationDate: user.RegistrationDate,
	}
}
