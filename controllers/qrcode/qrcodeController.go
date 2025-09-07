package qrcode

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/kgermando/sysmobembo-api/database"
	"github.com/kgermando/sysmobembo-api/models"
	qrGen "github.com/skip2/go-qrcode"
)

// QRCodeData représente les données à encoder dans le QR code
type QRCodeData struct {
	UUID         string    `json:"uuid"`
	Matricule    string    `json:"matricule"`
	Nom          string    `json:"nom"`
	PostNom      string    `json:"postnom"`
	Prenom       string    `json:"prenom"`
	Grade        string    `json:"grade"`
	Fonction     string    `json:"fonction"`
	Service      string    `json:"service"`
	Direction    string    `json:"direction"`
	Ministere    string    `json:"ministere"`
	PhotoProfil  string    `json:"photo_profil"`
	DateEmission time.Time `json:"date_emission"`
	ValidUntil   time.Time `json:"valid_until"`
}

type QRCodeController struct{}

// GenerateQRCode génère un QR code pour un agent
func GenerateQRCode(c *fiber.Ctx) error {
	userUUID := c.Params("uuid")
	if userUUID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "UUID utilisateur requis"})
	}

	// Récupérer l'utilisateur depuis la base de données
	var user models.User
	if err := database.DB.Where("uuid = ?", userUUID).First(&user).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Utilisateur non trouvé"})
	}

	// Créer les données pour le QR code
	qrData := QRCodeData{
		UUID:         user.UUID,
		Matricule:    user.Matricule,
		Nom:          user.Nom,
		PostNom:      user.PostNom,
		Prenom:       user.Prenom,
		Grade:        user.Grade,
		Fonction:     user.Fonction,
		Service:      user.Service,
		Direction:    user.Direction,
		Ministere:    user.Ministere,
		PhotoProfil:  user.PhotoProfil,
		DateEmission: time.Now(),
		ValidUntil:   time.Now().Add(365 * 24 * time.Hour), // Validité 1 an
	}

	// Encoder les données en JSON
	jsonData, err := json.Marshal(qrData)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Erreur lors de l'encodage JSON"})
	}

	// Définir le répertoire de sortie
	outputDir := "./uploads/qrcodes"
	err = os.MkdirAll(outputDir, 0755)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Erreur lors de la création du répertoire"})
	}

	// Nom du fichier QR code
	filename := fmt.Sprintf("qr_%s_%d.png", qrData.Matricule, time.Now().Unix())
	filepath := filepath.Join(outputDir, filename)

	// Générer le QR code
	err = qrGen.WriteFile(string(jsonData), qrGen.Medium, 256, filepath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Erreur lors de la génération du QR code"})
	}

	// Mettre à jour l'utilisateur avec les informations du QR code
	user.QRCode = fmt.Sprintf("/api/qrcode/image/%s", filename)
	user.QRCodeData = string(jsonData)
	database.DB.Save(&user)

	return c.JSON(fiber.Map{
		"message":      "QR code généré avec succès",
		"qr_code_path": filepath,
		"qr_code_url":  fmt.Sprintf("/api/qrcode/image/%s", filename),
		"qr_code_data": string(jsonData),
		"validity":     qrData.ValidUntil,
		"agent": fiber.Map{
			"uuid":      user.UUID,
			"matricule": user.Matricule,
			"nom":       user.Nom,
			"postnom":   user.PostNom,
			"prenom":    user.Prenom,
		},
	})
}

// VerifyQRCode vérifie et décode un QR code
func VerifyQRCode(c *fiber.Ctx) error {
	var request struct {
		QRCodeData string `json:"qr_code_data"`
	}

	if err := c.BodyParser(&request); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Données invalides"})
	}

	if request.QRCodeData == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Données QR code requises"})
	}

	// Décoder les données JSON
	var qrData QRCodeData
	err := json.Unmarshal([]byte(request.QRCodeData), &qrData)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Format QR code invalide"})
	}

	// Vérifier la validité
	if time.Now().After(qrData.ValidUntil) {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "QR code expiré"})
	}

	return c.JSON(fiber.Map{
		"message": "QR code valide",
		"data":    qrData,
		"status":  "valid",
	})
}

// GetAgentByQR récupère les informations d'un agent via son UUID depuis un QR code
func GetAgentByQR(c *fiber.Ctx) error {
	uuid := c.Params("uuid")
	if uuid == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "UUID requis"})
	}

	// Ici vous devriez récupérer l'utilisateur depuis votre base de données
	// var user models.User
	// result := db.Where("uuid = ?", uuid).First(&user)
	// if result.Error != nil {
	// 	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Agent non trouvé"})
	// }

	// Pour l'exemple, retournons des données fictives
	response := fiber.Map{
		"uuid":      uuid,
		"matricule": "MAT001",
		"nom":       "DUPONT",
		"postnom":   "MARTIN",
		"prenom":    "Jean",
		"fullname":  "DUPONT MARTIN Jean",
		"grade":     "Agent Principal",
		"fonction":  "Administrateur",
		"service":   "Ressources Humaines",
		"direction": "Administration Générale",
		"ministere": "Ministère de la Fonction Publique",
		"photo":     "/uploads/photos/photo_001.jpg",
		"statut":    "Actif",
	}

	return c.JSON(fiber.Map{
		"message": "Agent vérifié avec succès",
		"agent":   response,
	})
}

// RefreshQRCode renouvelle le QR code d'un agent
func RefreshQRCode(c *fiber.Ctx) error {
	userUUID := c.Params("uuid")
	if userUUID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "UUID utilisateur requis"})
	}

	var request struct {
		ValidityDays int `json:"validity_days"`
	}

	// Valeur par défaut: 365 jours
	validityDays := 365
	if err := c.BodyParser(&request); err == nil && request.ValidityDays > 0 {
		validityDays = request.ValidityDays
	}

	// Générer un nouveau QR code
	validityDuration := time.Duration(validityDays) * 24 * time.Hour
	newValidUntil := time.Now().Add(validityDuration)

	qrData := QRCodeData{
		UUID:         userUUID,
		Matricule:    "MAT001",
		Nom:          "DUPONT",
		PostNom:      "MARTIN",
		Prenom:       "Jean",
		Grade:        "Agent Principal",
		Fonction:     "Administrateur",
		Service:      "Ressources Humaines",
		Direction:    "Administration Générale",
		Ministere:    "Ministère de la Fonction Publique",
		PhotoProfil:  "/uploads/photos/photo_001.jpg",
		DateEmission: time.Now(),
		ValidUntil:   newValidUntil,
	}

	jsonData, err := json.Marshal(qrData)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Erreur lors de l'encodage JSON"})
	}

	// Générer le nouveau fichier QR code
	outputDir := "./uploads/qrcodes"
	filename := fmt.Sprintf("qr_%s_%d.png", qrData.Matricule, time.Now().Unix())
	filepath := filepath.Join(outputDir, filename)

	err = qrGen.WriteFile(string(jsonData), qrGen.Medium, 256, filepath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Erreur lors de la génération du fichier QR code"})
	}

	return c.JSON(fiber.Map{
		"message":        "QR code renouvelé avec succès",
		"qr_code_path":   filepath,
		"qr_code_url":    fmt.Sprintf("/api/qrcode/image/%s", filename),
		"qr_code_data":   string(jsonData),
		"validity_until": newValidUntil,
	})
}

// ServeQRCode sert les fichiers QR code
func ServeQRCode(c *fiber.Ctx) error {
	filename := c.Params("filename")
	filepath := fmt.Sprintf("./uploads/qrcodes/%s", filename)

	// Vérifier si le fichier existe
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "QR code non trouvé"})
	}

	return c.SendFile(filepath)
}
