package models

import "time"

type PasswordReset struct {
	UUID           string    `json:"-"`
	Email          string    `json:"email" validate:"required,email"`
	Token          string    `json:"-"`
	ExpirationTime time.Time `json:"-"`
	CreatedAt      time.Time `json:"-"`
}

type Reset struct {
	Password        string `json:"password" validate:"required"`
	PasswordConfirm string `json:"password_confirm" validate:"required"`
}
