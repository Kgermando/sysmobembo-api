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
	Nom           string    `json:"nom" gorm:"not null;default:''" validate:"required"`
	Postnom       string    `json:"postnom" gorm:"not null;default:''" validate:"required"`
	Prenom        string    `json:"prenom" gorm:"not null;default:''" validate:"required"`
	DateNaissance time.Time `json:"date_naissance" validate:"required"`
	LieuNaissance string    `json:"lieu_naissance" gorm:"default:''" validate:"required"`
	Sexe          string    `json:"sexe" gorm:"type:varchar(1);default:''" validate:"required,oneof=M F"`
	Nationalite   string    `json:"nationalite" gorm:"default:''" validate:"required"`

	Adresse    string `json:"adresse" gorm:"default:''"`
	Profession string `json:"profession" gorm:"default:''"`

	PaysEmetteur     string `json:"pays_emetteur" gorm:"not null;default:''" validate:"required"`
	AutoriteEmetteur string `json:"autorite_emetteur" gorm:"not null;default:''" validate:"required"`
	DateEmission     time.Time `json:"date_emission" validate:"required"`
	DateExpiration   time.Time `json:"date_expiration" validate:"required"`

	NumeroPasseport string `json:"numero_passeport" gorm:"unique;not null;default:''" validate:"required"`

	// Relations
	Migrants         []Migrant         `json:"migrants" gorm:"foreignKey:IdentiteUUID;constraint:OnDelete:CASCADE"`
	Geolocalisations []Geolocalisation `json:"geolocalisations" gorm:"foreignKey:IdentiteUUID;constraint:OnDelete:CASCADE"`
}

func (i *Identite) TableName() string {
	return "identites"
}
