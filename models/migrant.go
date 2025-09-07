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

	// Informations personnelles
	Nom           string    `json:"nom" gorm:"not null" validate:"required"`
	Prenom        string    `json:"prenom" gorm:"not null" validate:"required"`
	DateNaissance time.Time `json:"date_naissance" validate:"required"`
	LieuNaissance string    `json:"lieu_naissance" validate:"required"`
	Sexe          string    `json:"sexe" gorm:"type:varchar(1)" validate:"required,oneof=M F"`
	Nationalite   string    `json:"nationalite" validate:"required"`

	// Documents d'identité
	TypeDocument      string     `json:"type_document" validate:"required,oneof=passport carte_identite permis_conduire"`
	NumeroDocument    string     `json:"numero_document" gorm:"unique" validate:"required"`
	DateEmissionDoc   *time.Time `json:"date_emission_document"`
	DateExpirationDoc *time.Time `json:"date_expiration_document"`
	AutoriteEmission  string     `json:"autorite_emission"`

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
	PaysOrigine      string     `json:"pays_origine" validate:"required"`
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
