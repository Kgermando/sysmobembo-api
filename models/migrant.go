package models

import (
	"time"

	"gorm.io/gorm"
)

// Migrant représente l'identité principale d'un migrant
type Migrant struct {
	UUID      string         `gorm:"type:varchar(255);primary_key" json:"uuid"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`

	NumeroIdentifiant string `json:"numero_identifiant" gorm:"unique;not null" validate:"required"`

	// Relation avec Identite
	IdentiteUUID string   `json:"identite_uuid" gorm:"not null" validate:"required"`
	Identite     Identite `json:"identite" gorm:"foreignKey:IdentiteUUID;constraint:OnDelete:CASCADE"`

	// Informations de contact
	Telephone       string `json:"telephone"`
	Email           string `json:"email" gorm:"unique"`
	AdresseActuelle string `json:"adresse_actuelle"`
	VilleActuelle   string `json:"ville_actuelle"`
	PaysActuel      string `json:"pays_actuel"`

	// Informations familiales
	SituationMatrimoniale string `json:"situation_matrimoniale" validate:"oneof=celibataire marie divorce veuf"`
	NombreEnfants         int    `json:"nombre_enfants" gorm:"default:0"`
	PersonneContact       string `json:"personne_contact"`
	TelephoneContact      string `json:"telephone_contact"`

	// Statut migration
	StatutMigratoire string     `json:"statut_migratoire" validate:"required,oneof=regulier irregulier demandeur_asile refugie"`
	DateEntree       *time.Time `json:"date_entree"`
	PointEntree      string     `json:"point_entree"`
	PaysDestination  string     `json:"pays_destination"`

	// Relations avec autres modèles
	MotifDeplacements []MotifDeplacement `json:"motif_deplacements" gorm:"foreignKey:MigrantUUID;constraint:OnDelete:CASCADE"`
	Alertes           []Alert            `json:"alertes" gorm:"foreignKey:MigrantUUID;constraint:OnDelete:CASCADE"`
	Biometries        []Biometrie        `json:"biometries" gorm:"foreignKey:MigrantUUID;constraint:OnDelete:CASCADE"`
	Geolocalisations  []Geolocalisation  `json:"geolocalisations" gorm:"foreignKey:MigrantUUID;constraint:OnDelete:CASCADE"`

	// Métadonnées
	Actif bool `json:"actif" gorm:"default:true"`
}

func (m *Migrant) TableName() string {
	return "migrants"
}
