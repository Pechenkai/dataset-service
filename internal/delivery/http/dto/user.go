package dto

import (
	"regexp"
	"time"

	"ppo/internal/entities"
	"ppo/internal/services"
)

type RegisterUserRequest struct {
	Username string `json:"username" validate:"required,min=3,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Country  string `json:"country" validate:"required"`
	Role     string `json:"role" validate:"required,oneof=guest user admin"`
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

type UpdateUserRequest struct {
	Username  string `json:"username" validate:"omitempty,min=3,max=50"`
	Email     string `json:"email" validate:"omitempty,email"`
	Password  string `json:"password" validate:"omitempty,min=6"`
	Country   string `json:"country" validate:"omitempty"`
	IsBlocked *bool  `json:"is_blocked"` // optional
	Role      string `json:"role" validate:"omitempty,oneof=guest user admin"`
}

func (r *UpdateUserRequest) ToCommand(id uint64) services.UpdateUserCmd {
	cmd := services.UpdateUserCmd{ID: id}
	if r.Username != "" {
		cmd.Username = r.Username
	}
	if r.Email != "" {
		cmd.Email = r.Email
	}
	if r.Password != "" {
		cmd.Password = r.Password
	}
	if r.Country != "" {
		cmd.Country = r.Country
	}
	if r.IsBlocked != nil {
		cmd.IsBlocked = *r.IsBlocked
	}
	if r.Role != "" {
		cmd.Role = r.Role
	}
	return cmd
}

type UserResponse struct {
	ID               uint64    `json:"id"`
	Username         string    `json:"username"`
	Email            string    `json:"email"`
	Country          string    `json:"country"`
	Role             string    `json:"role"`
	RegistrationDate time.Time `json:"registration_date"`
	IsBlocked        bool      `json:"is_blocked"`
}

func FromEntityUser(u *entities.User) UserResponse {
	return UserResponse{
		ID:               u.ID,
		Username:         u.Username,
		Email:            u.Email,
		Country:          u.Country,
		Role:             u.Role,
		RegistrationDate: u.RegistrationDate,
		IsBlocked:        u.IsBlocked,
	}
}

type AuthenticateResponse struct {
	User  UserResponse `json:"user"`
	Token string       `json:"token"` // если планируется JWT, иначе пустое
}

type UsersResponse struct {
	Users []UserResponse `json:"users"`
}

var emailRegexp = regexp.MustCompile(`^[^@\s]+@[^@\s]+\.[^@\s]+$`)

func isValidEmail(email string) bool {
	return emailRegexp.MatchString(email)
}
