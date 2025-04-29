package entities

import (
	"strings"
	"time"
)

type User struct {
	ID               uint64    `json:"id"`
	Username         string    `json:"username"`
	Email            string    `json:"email"`
	Password         string    `json:"-"`
	RegistrationDate time.Time `json:"registration_date"`
	Country          string    `json:"country"`
	IsBlocked        bool      `json:"is_blocked"`
	Role             string    `json:"role"`
}

const (
	RoleGuest = "guest"
	RoleUser  = "user"
	RoleAdmin = "admin"
)

var validRoles = map[string]struct{}{
	RoleGuest: {},
	RoleUser:  {},
	RoleAdmin: {},
}

func NewUser(username, email, password, country, role string, registrationDate time.Time) (*User, error) {
	username = strings.TrimSpace(username)
	email = strings.TrimSpace(email)
	password = strings.TrimSpace(password)
	country = strings.TrimSpace(country)
	role = strings.ToLower(strings.TrimSpace(role))

	if username == "" {
		return nil, ErrEmptyUsername
	}
	if email == "" {
		return nil, ErrEmptyEmail
	}
	if password == "" {
		return nil, ErrEmptyPassword
	}
	if _, ok := validRoles[role]; !ok {
		return nil, ErrInvalidRole
	}
	if registrationDate.After(time.Now()) {
		return nil, ErrInvalidRegistrationDate
	}

	return &User{
		Username:         username,
		Email:            email,
		Password:         password,
		Country:          country,
		IsBlocked:        false,
		Role:             role,
		RegistrationDate: registrationDate,
	}, nil
}
