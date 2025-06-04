package dto

import (
	"ppo/internal/entities"
	"time"
)

type UserDTO struct {
	ID               uint64
	Username         string
	Email            string
	Country          string
	Role             string
	IsBlocked        bool
	RegistrationDate time.Time
}

func ToUserDTO(u *entities.User) *UserDTO {
	return &UserDTO{
		ID:               u.ID,
		Username:         u.Username,
		Email:            u.Email,
		Country:          u.Country,
		Role:             u.Role,
		IsBlocked:        u.IsBlocked,
		RegistrationDate: u.RegistrationDate,
	}
}

func ToUserDTOs(list []*entities.User) []*UserDTO {
	res := make([]*UserDTO, 0, len(list))
	for _, u := range list {
		res = append(res, ToUserDTO(u))
	}
	return res
}

type CreateUserForm struct {
	Username string
	Email    string
	Password string
	Country  string
	Role     string
}

type UpdateUserForm struct {
	ID        uint64
	Username  string
	Email     string
	Password  string
	Country   string
	IsBlocked bool
	Role      string
}

type AuthenticateForm struct {
	Email    string
	Password string
}
