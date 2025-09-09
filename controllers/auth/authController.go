package auth

import (
	"fmt"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/kgermando/sysmobembo-api/database"
	"github.com/kgermando/sysmobembo-api/models"
	"github.com/kgermando/sysmobembo-api/utils"
)

var SECRET_KEY string = os.Getenv("SECRET_KEY")

func Register(c *fiber.Ctx) error {

	nu := new(models.User)

	if err := c.BodyParser(&nu); err != nil {
		c.Status(400)
		return c.JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	if nu.Password != nu.PasswordConfirm {
		c.Status(400)
		return c.JSON(fiber.Map{
			"message": "passwords do not match",
		})
	}

	u := &models.User{
		UUID: uuid.New().String(),

		// Informations personnelles de base
		Nom:           nu.Nom,
		PostNom:       nu.PostNom,
		Prenom:        nu.Prenom,
		Sexe:          nu.Sexe,
		DateNaissance: nu.DateNaissance,
		LieuNaissance: nu.LieuNaissance,

		// État civil
		EtatCivil:     nu.EtatCivil,
		NombreEnfants: nu.NombreEnfants,

		// Nationalité et documents d'identité
		Nationalite:       nu.Nationalite,
		NumeroCNI:         nu.NumeroCNI,
		DateEmissionCNI:   nu.DateEmissionCNI,
		DateExpirationCNI: nu.DateExpirationCNI,
		LieuEmissionCNI:   nu.LieuEmissionCNI,

		// Contacts
		Email:            nu.Email,
		Telephone:        nu.Telephone,
		TelephoneUrgence: nu.TelephoneUrgence,

		// Adresse
		Province: nu.Province,
		Ville:    nu.Ville,
		Commune:  nu.Commune,
		Quartier: nu.Quartier,
		Avenue:   nu.Avenue,
		Numero:   nu.Numero,

		// Informations professionnelles
		Matricule:        nu.Matricule,
		Grade:            nu.Grade,
		Fonction:         nu.Fonction,
		Service:          nu.Service,
		Direction:        nu.Direction,
		Ministere:        nu.Ministere,
		DateRecrutement:  nu.DateRecrutement,
		DatePriseService: nu.DatePriseService,
		TypeAgent:        nu.TypeAgent,
		Statut:           nu.Statut,

		// Formation et éducation
		NiveauEtude:     nu.NiveauEtude,
		DiplomeBase:     nu.DiplomeBase,
		UniversiteEcole: nu.UniversiteEcole,
		AnneeObtention:  nu.AnneeObtention,
		Specialisation:  nu.Specialisation,

		// Informations bancaires
		NumeroBancaire: nu.NumeroBancaire,
		Banque:         nu.Banque,

		// Informations de sécurité sociale
		NumeroCNSS: nu.NumeroCNSS,
		NumeroONEM: nu.NumeroONEM,

		// Documents et photos
		PhotoProfil: nu.PhotoProfil,
		CVDocument:  nu.CVDocument,

		// Informations système
		Role:       nu.Role,
		Permission: nu.Permission,
		Status:     nu.Status,
		Signature:  nu.Signature,
	}

	u.SetPassword(nu.Password)

	if err := utils.ValidateStruct(*u); err != nil {
		c.Status(400)
		return c.JSON(err)
	}

	database.DB.Create(u)

	return c.JSON(fiber.Map{
		"message": "user account created",
		"data":    u,
	})
}

func Login(c *fiber.Ctx) error {

	err := CreateAdminUser()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Erreur lors de la création de l'utilisateur admin",
			"error":   err.Error(),
		})
	}

	lu := new(models.Login)

	if err := c.BodyParser(&lu); err != nil {
		c.Status(400)
		return c.JSON(fiber.Map{
			"message": err.Error(),
		})
	}

	if err := utils.ValidateStruct(*lu); err != nil {
		c.Status(400)
		return c.JSON(err)
	}

	u := &models.User{}

	result := database.DB.Where("email = ? OR telephone = ?", lu.Identifier, lu.Identifier).
		First(&u)

	if result.Error != nil {
		c.Status(404)
		return c.JSON(fiber.Map{
			"message": "invalid email or telephone 😰",
		})
	}

	if err := u.ComparePassword(lu.Password); err != nil {
		c.Status(400)
		return c.JSON(fiber.Map{
			"message": "mot de passe incorrect! 😰",
		})
	}

	if !u.Status {
		c.Status(400)
		return c.JSON(fiber.Map{
			"message": "vous n'êtes pas autorisé de se connecter 😰",
		})
	}

	// Mettre à jour les informations de connexion
	u.DernierAcces = time.Now()
	u.NombreConnexions++
	database.DB.Save(&u)

	token, err := utils.GenerateJwt(u.UUID)
	if err != nil {
		return c.SendStatus(fiber.StatusInternalServerError)
	}

	return c.JSON(fiber.Map{
		"message": "success",
		"data":    token,
	})

}

func AuthUser(c *fiber.Ctx) error {

	token := c.Query("token")

	fmt.Println("token", token)

	// cookie := c.Cookies("token")
	UserUUID, _ := utils.VerifyJwt(token)

	fmt.Println("UserUUID", UserUUID)

	u := models.User{}

	database.DB.Where("users.uuid = ?", UserUUID).
		First(&u)
	r := &models.UserResponse{
		UUID: u.UUID,

		// Informations personnelles de base
		Nom:           u.Nom,
		PostNom:       u.PostNom,
		Prenom:        u.Prenom,
		Fullname:      u.GetFullName(),
		Sexe:          u.Sexe,
		DateNaissance: u.DateNaissance,
		LieuNaissance: u.LieuNaissance,

		// État civil
		EtatCivil:     u.EtatCivil,
		NombreEnfants: u.NombreEnfants,

		// Nationalité et documents d'identité
		Nationalite:       u.Nationalite,
		NumeroCNI:         u.NumeroCNI,
		DateEmissionCNI:   u.DateEmissionCNI,
		DateExpirationCNI: u.DateExpirationCNI,
		LieuEmissionCNI:   u.LieuEmissionCNI,

		// Contacts
		Email:            u.Email,
		Telephone:        u.Telephone,
		TelephoneUrgence: u.TelephoneUrgence,

		// Adresse
		Province:        u.Province,
		Ville:           u.Ville,
		Commune:         u.Commune,
		Quartier:        u.Quartier,
		Avenue:          u.Avenue,
		Numero:          u.Numero,
		AdresseComplete: fmt.Sprintf("%s, %s, %s, %s", u.Avenue, u.Quartier, u.Commune, u.Ville),

		// Informations professionnelles
		Matricule:        u.Matricule,
		Grade:            u.Grade,
		Fonction:         u.Fonction,
		Service:          u.Service,
		Direction:        u.Direction,
		Ministere:        u.Ministere,
		DateRecrutement:  u.DateRecrutement,
		DatePriseService: u.DatePriseService,
		TypeAgent:        u.TypeAgent,
		Statut:           u.Statut,

		// Formation et éducation
		NiveauEtude:     u.NiveauEtude,
		DiplomeBase:     u.DiplomeBase,
		UniversiteEcole: u.UniversiteEcole,
		AnneeObtention:  u.AnneeObtention,
		Specialisation:  u.Specialisation,

		// Informations bancaires
		NumeroBancaire: u.NumeroBancaire,
		Banque:         u.Banque,

		// Informations de sécurité sociale
		NumeroINSS: u.NumeroCNSS, // Mapping de NumeroCNSS vers NumeroINSS
		NumeroONEM: u.NumeroONEM,

		// Documents et photos
		PhotoProfil: u.PhotoProfil,
		CVDocument:  u.CVDocument,

		// QR Code
		QRCode:     u.QRCode,
		QRCodeData: u.QRCodeData,

		// Système
		Role:       u.Role,
		Permission: u.Permission,
		Status:     u.Status,
		Signature:  u.Signature,

		// Audit
		CreatedAt:        u.CreatedAt,
		UpdatedAt:        u.UpdatedAt,
		DernierAcces:     u.DernierAcces,
		NombreConnexions: u.NombreConnexions,
	}
	return c.JSON(r)
}

func Logout(c *fiber.Ctx) error {
	cookie := fiber.Cookie{
		Name:     "token",
		Value:    "",
		Expires:  time.Now().Add(-time.Hour), // 1 day ,
		HTTPOnly: true,
	}
	c.Cookie(&cookie)

	return c.JSON(fiber.Map{
		"message": "success",
		"Logout":  "success",
	})

}

// User bioprofile
func UpdateInfo(c *fiber.Ctx) error {
	type UpdateDataInput struct {
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
		Province string `json:"province"`
		Ville    string `json:"ville"`
		Commune  string `json:"commune"`
		Quartier string `json:"quartier"`
		Avenue   string `json:"avenue"`
		Numero   string `json:"numero"`

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
		NumeroCNSS string `json:"numero_cnss"`
		NumeroONEM string `json:"numero_onem"`

		// Documents et photos
		PhotoProfil string `json:"photo_profil"`
		CVDocument  string `json:"cv_document"`

		// Signature
		Signature string `json:"signature"`
	}
	var updateData UpdateDataInput

	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Review your input",
			"errors":  err.Error(),
		})
	}

	cookie := c.Cookies("token")

	UserUUID, _ := utils.VerifyJwt(cookie)

	user := new(models.User)

	db := database.DB

	// Utiliser UUID au lieu de convertir en int
	result := db.Where("uuid = ?", UserUUID).First(&user)

	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Utilisateur non trouvé",
		})
	}

	// Mettre à jour tous les champs
	// Informations personnelles de base
	user.Nom = updateData.Nom
	user.PostNom = updateData.PostNom
	user.Prenom = updateData.Prenom
	user.Sexe = updateData.Sexe
	user.DateNaissance = updateData.DateNaissance
	user.LieuNaissance = updateData.LieuNaissance

	// État civil
	user.EtatCivil = updateData.EtatCivil
	user.NombreEnfants = updateData.NombreEnfants

	// Nationalité et documents d'identité
	user.Nationalite = updateData.Nationalite
	user.NumeroCNI = updateData.NumeroCNI
	user.DateEmissionCNI = updateData.DateEmissionCNI
	user.DateExpirationCNI = updateData.DateExpirationCNI
	user.LieuEmissionCNI = updateData.LieuEmissionCNI

	// Contacts
	user.Email = updateData.Email
	user.Telephone = updateData.Telephone
	user.TelephoneUrgence = updateData.TelephoneUrgence

	// Adresse
	user.Province = updateData.Province
	user.Ville = updateData.Ville
	user.Commune = updateData.Commune
	user.Quartier = updateData.Quartier
	user.Avenue = updateData.Avenue
	user.Numero = updateData.Numero

	// Informations professionnelles
	user.Matricule = updateData.Matricule
	user.Grade = updateData.Grade
	user.Fonction = updateData.Fonction
	user.Service = updateData.Service
	user.Direction = updateData.Direction
	user.Ministere = updateData.Ministere
	user.DateRecrutement = updateData.DateRecrutement
	user.DatePriseService = updateData.DatePriseService
	user.TypeAgent = updateData.TypeAgent
	user.Statut = updateData.Statut

	// Formation et éducation
	user.NiveauEtude = updateData.NiveauEtude
	user.DiplomeBase = updateData.DiplomeBase
	user.UniversiteEcole = updateData.UniversiteEcole
	user.AnneeObtention = updateData.AnneeObtention
	user.Specialisation = updateData.Specialisation

	// Informations bancaires
	user.NumeroBancaire = updateData.NumeroBancaire
	user.Banque = updateData.Banque

	// Informations de sécurité sociale
	user.NumeroCNSS = updateData.NumeroCNSS
	user.NumeroONEM = updateData.NumeroONEM

	// Documents et photos
	user.PhotoProfil = updateData.PhotoProfil
	user.CVDocument = updateData.CVDocument

	// Signature
	user.Signature = updateData.Signature

	db.Save(&user)

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "User successfully updated",
		"data":    user,
	})

}

func ChangePassword(c *fiber.Ctx) error {
	type UpdateDataInput struct {
		OldPassword     string `json:"old_password"`
		Password        string `json:"password"`
		PasswordConfirm string `json:"password_confirm"`
	}
	var updateData UpdateDataInput

	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Review your input",
			"errors":  err.Error(),
		})
	}

	// Utiliser la même logique que AuthUser - récupérer le token depuis les query params
	token := c.Query("token")

	fmt.Println("token", token)

	UserUUID, err := utils.VerifyJwt(token)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"status":  "error",
			"message": "Token invalide ou expiré",
		})
	}

	fmt.Println("UserUUID", UserUUID)

	user := new(models.User)

	// Utiliser UUID au lieu de id car c'est la clé primaire du modèle User
	result := database.DB.Where("uuid = ?", UserUUID).First(&user)

	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Utilisateur non trouvé",
		})
	}

	if err := user.ComparePassword(updateData.OldPassword); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "votre mot de passe n'est pas correct! 😰",
		})
	}

	if updateData.Password != updateData.PasswordConfirm {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "passwords do not match",
		})
	}

	// Utiliser la méthode SetPassword du modèle au lieu de utils.HashPassword
	user.SetPassword(updateData.Password)

	db := database.DB
	db.Save(&user)

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Mot de passe modifié avec succès",
	})
}

// GenerateUserQRCode génère ou met à jour le QR code de l'utilisateur
func GenerateUserQRCode(c *fiber.Ctx) error {
	token := c.Query("token")

	UserUUID, err := utils.VerifyJwt(token)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"status":  "error",
			"message": "Token invalide ou expiré",
		})
	}

	user := new(models.User)
	result := database.DB.Where("uuid = ?", UserUUID).First(&user)

	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Utilisateur non trouvé",
		})
	}

	// Générer les données du QR code (validité de 1 an)
	qrData, err := user.GenerateQRCodeData(365 * 24 * time.Hour)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Erreur lors de la génération des données QR",
			"error":   err.Error(),
		})
	}

	// Sauvegarder les données QR dans la base de données
	user.QRCodeData = qrData
	database.DB.Save(&user)

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "QR code généré avec succès",
		"data": fiber.Map{
			"qr_code_data": qrData,
			"user_uuid":    user.UUID,
		},
	})
}

// GetUserQRCodeInfo récupère les informations du QR code de l'utilisateur
func GetUserQRCodeInfo(c *fiber.Ctx) error {
	token := c.Query("token")

	UserUUID, err := utils.VerifyJwt(token)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"status":  "error",
			"message": "Token invalide ou expiré",
		})
	}

	user := new(models.User)
	result := database.DB.Where("uuid = ?", UserUUID).First(&user)

	if result.Error != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Utilisateur non trouvé",
		})
	}

	qrInfo, err := user.GetQRCodeInfo()
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Aucun QR code trouvé",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":   "success",
		"message":  "Informations QR code récupérées",
		"data":     qrInfo,
		"is_valid": user.IsQRCodeValid(),
	})
}

// CreateAdminUser crée un utilisateur administrateur avec mot de passe hashé
func CreateAdminUser() error {
	// Vérifier si un admin existe déjà
	var existingAdmin models.User
	result := database.DB.Where("role = ?", "Admin").First(&existingAdmin)

	if result.Error == nil {
		fmt.Println("Un utilisateur admin existe déjà dans la base de données")
		return nil
	}

	// Créer un nouvel utilisateur admin
	adminUser := &models.User{
		UUID: uuid.New().String(),

		// Informations personnelles de base
		Nom:           "Mutala",
		PostNom:       "Tshibangu",
		Prenom:        "Leon",
		Sexe:          "M",
		DateNaissance: time.Date(1950, 1, 1, 0, 0, 0, 0, time.UTC),
		LieuNaissance: "Kinshasa",

		// État civil
		EtatCivil:     "Marié",
		NombreEnfants: 5,

		// Nationalité et documents d'identité
		Nationalite:       "Congolaise",
		NumeroCNI:         "CNI001ADMIN",
		DateEmissionCNI:   time.Now(),
		DateExpirationCNI: time.Now().AddDate(10, 0, 0),
		LieuEmissionCNI:   "Kinshasa",

		// Contacts
		Email:            "admin@sysmobembo.cd",
		Telephone:        "+243000000000",
		TelephoneUrgence: "+243000000001",

		// Adresse
		Province: "Kinshasa",
		Ville:    "Kinshasa",
		Commune:  "Gombe",
		Quartier: "Centre-ville",
		Avenue:   "Av. Kasa-Vubu",
		Numero:   "1",

		// Informations professionnelles
		Matricule:        "ADM001",
		Grade:            "Directeur Général",
		Fonction:         "Administrateur Système",
		Service:          "Informatique",
		Direction:        "Direction Générale",
		Ministere:        "Ministère de l'Intérieur",
		DateRecrutement:  time.Now(),
		DatePriseService: time.Now(),
		TypeAgent:        "Fonctionnaire",
		Statut:           "Actif",

		// Formation et éducation
		NiveauEtude:     "Universitaire",
		DiplomeBase:     "Licence en Informatique",
		UniversiteEcole: "Université de Kinshasa",
		AnneeObtention:  2015,
		Specialisation:  "Génie Logiciel",

		// Informations bancaires
		NumeroBancaire: "1234567890",
		Banque:         "BCDC",

		// Informations de sécurité sociale
		NumeroCNSS: "CNSS001ADMIN",
		NumeroONEM: "ONEM001ADMIN",

		// Documents et photos
		PhotoProfil: "",
		CVDocument:  "",

		// Informations système
		Role:       "Admin",
		Permission: "ALL",
		Status:     true,
		Signature:  "ADMIN_SIGNATURE",
	}

	// Définir le mot de passe et le hasher
	adminUser.SetPassword("Admin@2024!")

	// Valider la structure
	if err := utils.ValidateStruct(*adminUser); err != nil {
		fmt.Printf("Erreur de validation: %v\n", err)
		return fmt.Errorf("erreur de validation: %v", err)
	}

	// Sauvegarder dans la base de données
	if err := database.DB.Create(adminUser).Error; err != nil {
		fmt.Printf("Erreur lors de la création de l'admin: %v\n", err)
		return fmt.Errorf("erreur lors de la création de l'admin: %v", err)
	}

	fmt.Printf("Utilisateur admin créé avec succès!\n")
	fmt.Printf("Email: %s\n", adminUser.Email)
	fmt.Printf("Mot de passe: Admin@2024!\n")
	fmt.Printf("Rôle: %s\n", adminUser.Role)

	return nil
}

// CreateAdminHandler endpoint pour créer un utilisateur admin via HTTP
func CreateAdminHandler(c *fiber.Ctx) error {
	// Optionnel: Ajouter une vérification de sécurité ici
	// Par exemple, vérifier un token spécial ou une clé API

	err := CreateAdminUser()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Erreur lors de la création de l'utilisateur admin",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Utilisateur administrateur créé avec succès",
		"data": fiber.Map{
			"email":    "admin@sysmobembo.cd",
			"password": "Admin@2024!",
			"role":     "Admin",
		},
	})
}
