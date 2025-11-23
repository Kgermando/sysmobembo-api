package models

import (
	"time"

	"gorm.io/gorm"
)

// Geolocalisation représente les coordonnées géographiques liées au migrant
type Geolocalisation struct {
	UUID      string         `gorm:"type:varchar(255);primary_key" json:"uuid"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`

	// Relation avec Identite
	IdentiteUUID string   `json:"identite_uuid" gorm:"not null" validate:"required"`
	Identite     Identite `json:"identite" gorm:"foreignKey:IdentiteUUID;constraint:OnDelete:CASCADE"`

	// Coordonnées géographiques
	Latitude  float64 `json:"latitude" validate:"required,min=-90,max=90"`
	Longitude float64 `json:"longitude" validate:"required,min=-180,max=180"`
}

func (g *Geolocalisation) TableName() string {
	return "geolocalisations"
}
