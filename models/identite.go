package models

import (
	"time"

	"gorm.io/gorm"
)

// Identite représente les informations d'identité d'un passeport ordinaire
type Identite struct {
	UUID      string         `gorm:"type:varchar(255);primary_key" json:"uuid"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`

	// Informations personnelles (comme dans un passeport)
	Nom           string    `json:"nom" gorm:"not null" validate:"required"`
	Prenom        string    `json:"prenom" gorm:"not null" validate:"required"`
	DateNaissance time.Time `json:"date_naissance" validate:"required"`
	LieuNaissance string    `json:"lieu_naissance" validate:"required"`
	Sexe          string    `json:"sexe" gorm:"type:varchar(1)" validate:"required,oneof=M F"`
	Nationalite   string    `json:"nationalite" validate:"required"`

	Adresse    string `json:"adresse"`
	Profession string `json:"profession"`

	PaysEmetteur     string `json:"pays_emetteur" gorm:"not null" validate:"required"`
	AutoriteEmetteur string `json:"autorite_emetteur" gorm:"not null" validate:"required"`

	NumeroPasseport string `json:"numero_passeport" gorm:"unique;not null" validate:"required"`
}

func (i *Identite) TableName() string {
	return "identites"
}
