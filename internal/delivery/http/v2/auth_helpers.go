package v2

import (
	"context"
	"strings"

	"ppo/internal/delivery/http/middleware"
	"ppo/internal/entities"
	"ppo/internal/services"
)

func currentUser(ctx context.Context) *entities.User {
	return middleware.CurrentUser(ctx)
}

func currentUserOrError(ctx context.Context) (*entities.User, error) {
	user := currentUser(ctx)
	if user == nil {
		return nil, services.ErrInvalidCredentials
	}
	return user, nil
}

func requireAdmin(ctx context.Context) (*entities.User, error) {
	user, err := currentUserOrError(ctx)
	if err != nil {
		return nil, err
	}
	if !isAdmin(user) {
		return nil, services.ErrRequestForbidden
	}
	return user, nil
}

func ensureUserOrAdmin(ctx context.Context, targetID uint64) (*entities.User, error) {
	user, err := currentUserOrError(ctx)
	if err != nil {
		return nil, err
	}
	if user.ID != targetID && !isAdmin(user) {
		return nil, services.ErrRequestForbidden
	}
	return user, nil
}

func isAdmin(user *entities.User) bool {
	return user != nil && strings.EqualFold(user.Role, entities.RoleAdmin)
}

func ensureDatasetReadable(user *entities.User, ownerID uint64, isPublic bool) error {
	if isPublic {
		return nil
	}
	if user == nil {
		return services.ErrInvalidCredentials
	}
	if user.ID == ownerID || isAdmin(user) {
		return nil
	}
	return services.ErrRequestForbidden
}

func ensureDatasetOwnerOrAdmin(user *entities.User, ownerID uint64) error {
	if user == nil {
		return services.ErrInvalidCredentials
	}
	if user.ID == ownerID || isAdmin(user) {
		return nil
	}
	return services.ErrRequestForbidden
}
