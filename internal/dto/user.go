package dto

import "ppo/internal/services"

type RegisterUserRequest struct {
	Username string `json:"username" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
	Country  string `json:"country"`
	Role     string `json:"role" validate:"oneof=guest user admin"`
}

func (r *RegisterUserRequest) ToCommand() services.RegisterUserCmd {
	return services.RegisterUserCmd{
		Username: r.Username,
		Email:    r.Email,
		Password: r.Password,
		Country:  r.Country,
		Role:     r.Role,
	}
}

// AuthenticateUserRequest — POST /api/v1/users/login
type AuthenticateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func (r *AuthenticateUserRequest) ToCommand() services.AuthenticateUserCmd {
	return services.AuthenticateUserCmd{
		Email:    r.Email,
		Password: r.Password,
	}
}

// UpdateUserRequest — PUT /api/v1/users/{id}
type UpdateUserRequest struct {
	Username  string `json:"username"`
	Email     string `json:"email" validate:"omitempty,email"`
	Password  string `json:"password"`
	Country   string `json:"country"`
	IsBlocked bool   `json:"is_blocked"`
	Role      string `json:"role" validate:"omitempty,oneof=guest user admin"`
}

func (r *UpdateUserRequest) ToCommand(id uint64) services.UpdateUserCmd {
	return services.UpdateUserCmd{
		ID:        id,
		Username:  r.Username,
		Email:     r.Email,
		Password:  r.Password,
		Country:   r.Country,
		IsBlocked: r.IsBlocked,
		Role:      r.Role,
	}
}

// UserResponse
type UserResponse struct {
	ID        uint64 `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Country   string `json:"country"`
	Role      string `json:"role"`
	CreatedAt string `json:"registration_date"`
}
