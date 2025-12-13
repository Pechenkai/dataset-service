package v2

import (
	"net/http"
	"strings"

	chi "github.com/go-chi/chi/v5"

	"ppo/internal/delivery/http/dto"
	"ppo/internal/entities"
	"ppo/internal/services"
)

func (h *Handler) ListUsers(w http.ResponseWriter, r *http.Request) error {
	// Доступно только администраторам
	if _, err := requireAdmin(r.Context()); err != nil {
		return err
	}

	limit, offset, err := parsePagination(r, 50, 200)
	if err != nil {
		return err
	}
	filters, err := parseUserFilters(r)
	if err != nil {
		return err
	}

	users, err := h.users.ListAllUsers(r.Context())
	if err != nil {
		return err
	}

	filtered := filterUsers(users, filters)

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
	var req struct {
		Username string  `json:"username"`
		Email    string  `json:"email"`
		Password string  `json:"password"`
		Country  string  `json:"country"`
		Role     *string `json:"role,omitempty"`
	}
	if err := decodeJSON(r, &req); err != nil {
		return err
	}
	if strings.TrimSpace(req.Username) == "" ||
		strings.TrimSpace(req.Email) == "" ||
		strings.TrimSpace(req.Password) == "" ||
		strings.TrimSpace(req.Country) == "" {
		return &dto.BadRequestError{Message: "all fields are required"}
	}

	cmd := services.RegisterUserCmd{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
		Country:  req.Country,
		Role:     entities.RoleUser,
	}

	id, err := h.users.Register(r.Context(), cmd)
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
	// Доступно самому пользователю и администраторам
	if _, err := ensureUserOrAdmin(r.Context(), id); err != nil {
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

	// Доступно только администраторам
	if _, err := requireAdmin(r.Context()); err != nil {
		return err
	}

	req, err := decodeUpdateUserRequest(r)
	if err != nil {
		return err
	}

	current, err := h.users.GetUserByID(r.Context(), id)
	if err != nil {
		return err
	}

	cmd := buildUpdateUserCmd(id, current, req)
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
	// Доступно только администраторам
	if _, err := requireAdmin(r.Context()); err != nil {
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
	// Доступно самому пользователю и администраторам
	if _, err := ensureUserOrAdmin(r.Context(), userID); err != nil {
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

type userFilters struct {
	role      string
	email     string
	country   string
	isBlocked *bool
}

func parseUserFilters(r *http.Request) (userFilters, error) {
	q := r.URL.Query()
	isBlocked, err := parseBoolPtr(q, "is_blocked")
	if err != nil {
		return userFilters{}, err
	}
	return userFilters{
		role:      strings.TrimSpace(q.Get("role")),
		email:     strings.TrimSpace(q.Get("email")),
		country:   strings.TrimSpace(q.Get("country")),
		isBlocked: isBlocked,
	}, nil
}

func filterUsers(users []*entities.User, filters userFilters) []UserResponse {
	filtered := make([]UserResponse, 0, len(users))
	for _, user := range users {
		if filters.role != "" && !strings.EqualFold(user.Role, filters.role) {
			continue
		}
		if filters.email != "" && !strings.Contains(strings.ToLower(user.Email), strings.ToLower(filters.email)) {
			continue
		}
		if filters.country != "" && !strings.EqualFold(user.Country, filters.country) {
			continue
		}
		if filters.isBlocked != nil && user.IsBlocked != *filters.isBlocked {
			continue
		}
		filtered = append(filtered, toUserResponse(user))
	}
	return filtered
}

type updateUserRequest struct {
	Username  *string `json:"username"`
	Email     *string `json:"email"`
	Password  *string `json:"password"`
	Country   *string `json:"country"`
	Role      *string `json:"role"`
	IsBlocked *bool   `json:"is_blocked"`
}

func decodeUpdateUserRequest(r *http.Request) (updateUserRequest, error) {
	var req updateUserRequest
	if err := decodeJSON(r, &req); err != nil {
		return updateUserRequest{}, err
	}
	if req.Username == nil && req.Email == nil && req.Password == nil && req.Country == nil && req.Role == nil && req.IsBlocked == nil {
		return updateUserRequest{}, &dto.BadRequestError{Message: "nothing to update"}
	}
	return req, nil
}

func buildUpdateUserCmd(id uint64, current *entities.User, req updateUserRequest) services.UpdateUserCmd {
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
	return cmd
}
