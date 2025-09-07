package models

import (
	"time"

	"gorm.io/gorm"
)

// MotifDeplacement représente les raisons du déplacement du migrant
type MotifDeplacement struct {
	UUID      string         `gorm:"type:varchar(255);primary_key" json:"uuid"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`

	MigrantUUID string `json:"migrant_uuid" gorm:"type:varchar(255);not null"`

	// Types de motifs
	TypeMotif       string `json:"type_motif" validate:"required,oneof=economique politique persecution naturelle familial education sanitaire"`
	MotifPrincipal  string `json:"motif_principal" validate:"required"`
	MotifSecondaire string `json:"motif_secondaire"`
	Description     string `json:"description" gorm:"type:text"`

	// Contexte du déplacement
	CaractereVolontaire bool      `json:"caractere_volontaire" gorm:"default:true"`
	Urgence             string    `json:"urgence" validate:"oneof=faible moyenne elevee critique"`
	DateDeclenchement   time.Time `json:"date_declenchement" validate:"required"`
	DureeEstimee        int       `json:"duree_estimee"` // en jours

	// Facteurs externes
	ConflitArme          bool `json:"conflit_arme" gorm:"default:false"`
	CatastropheNaturelle bool `json:"catastrophe_naturelle" gorm:"default:false"`
	Persecution          bool `json:"persecution" gorm:"default:false"`
	ViolenceGeneralisee  bool `json:"violence_generalisee" gorm:"default:false"`

	// Relation avec Migrant
	Migrant Migrant `json:"migrant" gorm:"constraint:OnDelete:CASCADE"`
}

func (md *MotifDeplacement) TableName() string {
	return "motif_deplacements"
}
