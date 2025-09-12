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

	MigrantUUID string `json:"migrant_uuid" gorm:"type:varchar(255);not null"`

	// Coordonnées géographiques
	Latitude  float64  `json:"latitude" validate:"required,min=-90,max=90"`
	Longitude float64  `json:"longitude" validate:"required,min=-180,max=180"`

	// Informations contextuelles
	TypeLocalisation string `json:"type_localisation" validate:"required,oneof=residence_actuelle lieu_travail point_passage frontiere centre_accueil urgence"`
	Description      string `json:"description"`
	Adresse          string `json:"adresse"`
	Ville            string `json:"ville"`
	Pays             string `json:"pays" validate:"required"` 

	// Informations de mouvement
	TypeMouvement string `json:"type_mouvement" validate:"oneof=arrivee depart transit residence_temporaire residence_permanente"`
	DureeSejour   *int   `json:"duree_sejour"` // en jours
	ProchaineDest string `json:"prochaine_destination"`

	// Relation avec Migrant
	Migrant Migrant `json:"migrant" gorm:"constraint:OnDelete:CASCADE"`
}

func (g *Geolocalisation) TableName() string {
	return "geolocalisations"
}
