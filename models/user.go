package models

import (
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	UUID      string `gorm:"type:varchar(255);primary_key" json:"uuid"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	// Informations personnelles de base
	Nom           string    `gorm:"not null" json:"nom"`
	PostNom       string    `gorm:"not null" json:"postnom"`
	Prenom        string    `gorm:"not null" json:"prenom"`
	Sexe          string    `gorm:"not null" json:"sexe"` // M/F
	DateNaissance time.Time `gorm:"not null" json:"date_naissance"`
	LieuNaissance string    `gorm:"not null" json:"lieu_naissance"`

	// État civil
	EtatCivil     string `json:"etat_civil"` // Célibataire, Marié(e), Divorcé(e), Veuf(ve)
	NombreEnfants int    `gorm:"default:0" json:"nombre_enfants"`

	// Nationalité et documents d'identité
	Nationalite       string    `gorm:"not null" json:"nationalite"`
	NumeroCNI         string    `gorm:"unique" json:"numero_cni"` // Carte Nationale d'Identité
	DateEmissionCNI   time.Time `json:"date_emission_cni"`
	DateExpirationCNI time.Time `json:"date_expiration_cni"`
	LieuEmissionCNI   string    `json:"lieu_emission_cni"`

	// Contacts
	Email            string `gorm:"unique; not null" json:"email"`
	Telephone        string `gorm:"unique; not null" json:"telephone"`
	TelephoneUrgence string `json:"telephone_urgence"`

	// Adresse
	Province string `gorm:"not null" json:"province"`
	Ville    string `gorm:"not null" json:"ville"`
	Commune  string `gorm:"not null" json:"commune"`
	Quartier string `gorm:"not null" json:"quartier"`
	Avenue   string `json:"avenue"`
	Numero   string `json:"numero"`

	// Informations professionnelles
	Matricule        string    `gorm:"unique; not null" json:"matricule"`
	Grade            string    `gorm:"not null" json:"grade"`
	Fonction         string    `gorm:"not null" json:"fonction"`
	Service          string    `gorm:"not null" json:"service"`
	Direction        string    `gorm:"not null" json:"direction"`
	Ministere        string    `gorm:"not null" json:"ministere"`
	DateRecrutement  time.Time `gorm:"not null" json:"date_recrutement"`
	DatePriseService time.Time `gorm:"not null" json:"date_prise_service"`
	TypeAgent        string    `gorm:"not null" json:"type_agent"` // Fonctionnaire, Contractuel, Stagiaire
	Statut           string    `gorm:"not null" json:"statut"`     // Actif, Retraité, Suspendu, Révoqué

	// Formation et éducation
	NiveauEtude     string `json:"niveau_etude"` // Primaire, Secondaire, Universitaire, Post-universitaire
	DiplomeBase     string `json:"diplome_base"`
	UniversiteEcole string `json:"universite_ecole"`
	AnneeObtention  int    `json:"annee_obtention"`
	Specialisation  string `json:"specialisation"`

	// Informations bancaires
	NumeroBancaire string `json:"numero_bancaire"`
	Banque         string `json:"banque"`

	// Informations de sécurité sociale
	NumeroCNSS string `gorm:"unique" json:"numero_cnss"` // Institut National de Sécurité Sociale
	NumeroONEM string `json:"numero_onem"`               // Office National de l'Emploi

	// Documents et photos
	PhotoProfil string `json:"photo_profil"` // URL ou chemin vers la photo
	CVDocument  string `json:"cv_document"`  // URL ou chemin vers le CV

	// QR Code
	QRCode     string `json:"qr_code"`      // URL ou chemin vers l'image du QR code
	QRCodeData string `json:"qr_code_data"` // Données encodées dans le QR code (JSON avec infos de base)

	// Informations système
	Password        string `json:"password" validate:"required"`
	PasswordConfirm string `json:"password_confirm" gorm:"-"`
	Role            string `json:"role"` // Agent, Manager, Supervisor, Administrator
	Permission      string `json:"permission"`
	Status          bool   `gorm:"default:false" json:"status"`
	Signature       string `json:"signature"`

	// Audit et suivi
	DernierAcces     time.Time `json:"dernier_acces"`
	NombreConnexions int       `gorm:"default:0" json:"nombre_connexions"`
}

type UserResponse struct {
	UUID string `json:"uuid"`

	// Informations personnelles de base
	Nom           string    `json:"nom"`
	PostNom       string    `json:"postnom"`
	Prenom        string    `json:"prenom"`
	Sexe          string    `json:"sexe"`
	DateNaissance time.Time `json:"date_naissance"`
	LieuNaissance string    `json:"lieu_naissance"`

	// État civil
	EtatCivil     string `json:"etat_civil"`
	NombreEnfants int    `json:"nombre_enfants"`

	// Nationalité et documents d'identité
	Nationalite       string    `json:"nationalite"`
	NumeroCNI         string    `json:"numero_cni"`
	DateEmissionCNI   time.Time `json:"date_emission_cni"`
	DateExpirationCNI time.Time `json:"date_expiration_cni"`
	LieuEmissionCNI   string    `json:"lieu_emission_cni"`

	// Contacts
	Email            string `json:"email"`
	Telephone        string `json:"telephone"`
	TelephoneUrgence string `json:"telephone_urgence"`

	// Adresse
	Province        string `json:"province"`
	Ville           string `json:"ville"`
	Commune         string `json:"commune"`
	Quartier        string `json:"quartier"`
	Avenue          string `json:"avenue"`
	Numero          string `json:"numero"`
	AdresseComplete string `json:"adresse_complete"`

	// Informations professionnelles
	Matricule        string    `json:"matricule"`
	Grade            string    `json:"grade"`
	Fonction         string    `json:"fonction"`
	Service          string    `json:"service"`
	Direction        string    `json:"direction"`
	Ministere        string    `json:"ministere"`
	DateRecrutement  time.Time `json:"date_recrutement"`
	DatePriseService time.Time `json:"date_prise_service"`
	TypeAgent        string    `json:"type_agent"`
	Statut           string    `json:"statut"`

	// Formation et éducation
	NiveauEtude     string `json:"niveau_etude"`
	DiplomeBase     string `json:"diplome_base"`
	UniversiteEcole string `json:"universite_ecole"`
	AnneeObtention  int    `json:"annee_obtention"`
	Specialisation  string `json:"specialisation"`

	// Informations bancaires
	NumeroBancaire string `json:"numero_bancaire"`
	Banque         string `json:"banque"`

	// Informations de sécurité sociale
	NumeroINSS string `json:"numero_inss"`
	NumeroONEM string `json:"numero_onem"`

	// Documents et photos
	PhotoProfil string `json:"photo_profil"`
	PhotoCNI    string `json:"photo_cni"`
	CVDocument  string `json:"cv_document"`

	// Système
	Role       string `json:"role"`
	Permission string `json:"permission"`
	Status     bool   `json:"status"`
	Signature  string `json:"signature"`

	// Audit
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	DernierAcces     time.Time `json:"dernier_acces"`
	NombreConnexions int       `json:"nombre_connexions"`
}

type Login struct {
	Identifier string `json:"identifier" validate:"required"`
	Password   string `json:"password" validate:"required"`
}

func (u *User) SetPassword(p string) {
	hp, _ := bcrypt.GenerateFromPassword([]byte(p), 14)
	u.Password = string(hp)
}

func (u *User) ComparePassword(p string) error {
	err := bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(p))
	return err
}
