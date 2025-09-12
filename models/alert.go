package models

import (
	"time"

	"gorm.io/gorm"
)

// Alert représente les alertes liées au migrant
type Alert struct {
	UUID      string         `gorm:"type:varchar(255);primary_key" json:"uuid"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`

	MigrantUUID string `json:"migrant_uuid" gorm:"type:varchar(255);not null"`

	// Informations de l'alerte
	TypeAlerte    string `json:"type_alerte" validate:"required,oneof=securite sante juridique administrative humanitaire"`
	NiveauGravite string `json:"niveau_gravite" validate:"required,oneof=info warning danger critical"`
	Titre         string `json:"titre" validate:"required"`
	Description   string `json:"description" gorm:"type:text" validate:"required"`

	// Statut et traitement
	Statut              string     `json:"statut" gorm:"default:active" validate:"oneof=active resolved dismissed expired"`
	DateExpiration      *time.Time `json:"date_expiration"`
	ActionRequise       string     `json:"action_requise" gorm:"type:text"`
	PersonneResponsable string     `json:"personne_responsable"`

	// Métadonnées de traitement
	DateResolution        *time.Time `json:"date_resolution"`
	CommentaireResolution string     `json:"commentaire_resolution" gorm:"type:text"`

	// Relation avec Migrant
	Migrant Migrant `json:"migrant" gorm:"constraint:OnDelete:CASCADE"`
}

func (a *Alert) TableName() string {
	return "alertes"
}
