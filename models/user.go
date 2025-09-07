package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	UUID      string `gorm:"type:varchar(255);primary_key" json:"uuid"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	Fullname        string     `gorm:"not null" json:"fullname"`
	Email           string     `gorm:"unique; not null" json:"email"`
	Telephone       string     `gorm:"unique; not null" json:"telephone"`
	Password        string     `json:"password" validate:"required"`
	PasswordConfirm string     `json:"password_confirm" gorm:"-"`
	Role            string     `json:"role"` // Agent, Manager, Supervisor, Administrator
	Permission      string     `json:"permission"`
	Status          bool       `gorm:"default:false" json:"status"`
	Signature       string     `json:"signature"` 
}

type UserResponse struct {
	UUID           string     `json:"uuid"`
	Fullname       string     `json:"fullname"`
	Email          string     `json:"email"`
	Telephone      string     `json:"telephone"`
	Role           string     `json:"role"`
	Permission     string     `json:"permission"`
	Status         bool       `json:"status"` 
	Signature      string     `json:"signature"`
	CreatedAt      time.Time
	UpdatedAt      time.Time 
}

type Login struct { 
	Identifier string `json:"identifier" validate:"required"`
	Password   string `json:"password" validate:"required"`
}

func (u *User) SetPassword(p string) {
	hp, _ := bcrypt.GenerateFromPassword([]byte(p), 14)
	u.Password = string(hp)
}

func (u *User) ComparePassword(p string) error {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(p))
	return err
}
