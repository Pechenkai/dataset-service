package v2

import (
	"errors"
	"net/http"
	"strings"

	"ppo/internal/delivery/http/dto"
	"ppo/internal/services"
)

var errTokenServiceDisabled = errors.New("token service is not configured")

func (h *Handler) IssueToken(w http.ResponseWriter, r *http.Request) error {
	if h.tokens == nil {
		return errTokenServiceDisabled
	}
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := decodeJSON(r, &req); err != nil {
		return err
	}
	if strings.TrimSpace(req.Email) == "" || strings.TrimSpace(req.Password) == "" {
		return &dto.BadRequestError{Message: "email and password are required"}
	}

	user, err := h.users.Authenticate(r.Context(), services.AuthenticateUserCmd{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return err
	}

	record, err := h.tokens.IssueToken(r.Context(), user.ID, h.tokenTTL)
	if err != nil {
		return err
	}

	writeJSON(w, http.StatusOK, AuthenticateResponse{
		Token:     record.ID,
		ExpiresAt: record.ExpiresAt,
		User:      toUserResponse(user),
	})
	return nil
}

func (h *Handler) RevokeToken(w http.ResponseWriter, r *http.Request) error {
	if h.tokens == nil {
		return errTokenServiceDisabled
	}
	var req struct {
		TokenID *string `json:"token_id"`
	}
	_ = decodeJSON(r, &req)

	tokenID := ""
	if req.TokenID != nil {
		tokenID = strings.TrimSpace(*req.TokenID)
	}
	if tokenID == "" {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			return &dto.BadRequestError{Message: "token_id or Authorization header required"}
		}
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return &dto.BadRequestError{Message: "invalid Authorization header"}
		}
		tokenID = strings.TrimSpace(parts[1])
	}

	if err := h.tokens.RevokeToken(r.Context(), tokenID); err != nil {
		if errors.Is(err, services.ErrTokenNotFound) {
			return services.ErrTokenNotFound
		}
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}
