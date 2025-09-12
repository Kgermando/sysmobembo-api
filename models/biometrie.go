package models

import (
	"time"

	"gorm.io/gorm"
)

// Biometrie représente les données biométriques du migrant
type Biometrie struct {
	UUID      string         `gorm:"type:varchar(255);primary_key" json:"uuid"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deleted_at"`

	MigrantUUID string `json:"migrant_uuid" gorm:"type:varchar(255);not null"`

	// Types de données biométriques
	TypeBiometrie string `json:"type_biometrie" validate:"required,oneof=empreinte_digitale reconnaissance_faciale iris scan_retine signature_numerique"`
	IndexDoigt    *int   `json:"index_doigt"` // Pour les empreintes (1-10)
	QualiteDonnee string `json:"qualite_donnee" validate:"oneof=excellente bonne moyenne faible"`

	// Données encodées
	DonneesBiometriques string `json:"donnees_biometriques" gorm:"type:text;not null"` // Base64 ou hash
	AlgorithmeEncodage  string `json:"algorithme_encodage" validate:"required"`
	TailleFichier       int    `json:"taille_fichier"` // en bytes

	// Métadonnées de capture
	DateCapture       time.Time `json:"date_capture" validate:"required"`
	DisposifCapture   string    `json:"dispositif_capture"`
	ResolutionCapture string    `json:"resolution_capture"`
	OperateurCapture  string    `json:"operateur_capture"`

	// Sécurité et chiffrement
	Chiffre         bool   `json:"chiffre" gorm:"default:false"`          // Indique si les données sont chiffrées
	CleChiffrement  string `json:"-" gorm:"type:text"`                    // Clé de chiffrement (non exposée en JSON)

	// Validation et vérification
	Verifie          bool       `json:"verifie" gorm:"default:false"`
	DateVerification *time.Time `json:"date_verification"`
	ScoreConfiance   *float64   `json:"score_confiance"` // 0-1

	// Relation avec Migrant
	Migrant Migrant `json:"migrant" gorm:"constraint:OnDelete:CASCADE"`
}

func (b *Biometrie) TableName() string {
	return "biometries"
}
