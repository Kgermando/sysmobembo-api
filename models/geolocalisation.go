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
	Altitude  *float64 `json:"altitude"`  // en mètres
	Precision *float64 `json:"precision"` // en mètres

	// Informations contextuelles
	TypeLocalisation string `json:"type_localisation" validate:"required,oneof=residence_actuelle lieu_travail point_passage frontiere centre_accueil urgence"`
	Description      string `json:"description"`
	Adresse          string `json:"adresse"`
	Ville            string `json:"ville"`
	Pays             string `json:"pays" validate:"required"`
	CodePostal       string `json:"code_postal"`

	// Métadonnées de capture
	DateEnregistrement time.Time `json:"date_enregistrement" validate:"required"`
	MethodeCapture     string    `json:"methode_capture" validate:"oneof=gps manuel automatique"`
	DisposifSource     string    `json:"dispositif_source"`
	FiabiliteSource    string    `json:"fiabilite_source" validate:"oneof=elevee moyenne faible"`

	// Statut et validité
	Actif          bool       `json:"actif" gorm:"default:true"`
	DateValidation *time.Time `json:"date_validation"`
	ValidePar      string     `json:"valide_par"`
	Commentaire    string     `json:"commentaire" gorm:"type:text"`

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
