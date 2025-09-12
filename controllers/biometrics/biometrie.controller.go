package biometrics

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/kgermando/sysmobembo-api/database"
	"github.com/kgermando/sysmobembo-api/models"
	"github.com/kgermando/sysmobembo-api/utils"
)

// =======================
// SECURITY FUNCTIONS
// =======================

// Fonction pour chiffrer les données biométriques
func encryptBiometricData(data string) (string, string, error) {
	// Générer une clé de chiffrement
	key := make([]byte, 32) // 256 bits
	if _, err := rand.Read(key); err != nil {
		return "", "", err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, []byte(data), nil)

	// Retourner les données chiffrées et la clé encodées en base64
	return base64.StdEncoding.EncodeToString(ciphertext),
		base64.StdEncoding.EncodeToString(key), nil
}

// Fonction pour vérifier la qualité des données biométriques
func assessDataQuality(dataSize int, typeBiometrie string) string {
	switch typeBiometrie {
	case "empreinte_digitale":
		if dataSize > 50000 {
			return "excellente"
		} else if dataSize > 30000 {
			return "bonne"
		} else if dataSize > 15000 {
			return "moyenne"
		}
		return "faible"
	case "reconnaissance_faciale":
		if dataSize > 100000 {
			return "excellente"
		} else if dataSize > 60000 {
			return "bonne"
		} else if dataSize > 30000 {
			return "moyenne"
		}
		return "faible"
	case "iris", "scan_retine":
		if dataSize > 80000 {
			return "excellente"
		} else if dataSize > 50000 {
			return "bonne"
		} else if dataSize > 25000 {
			return "moyenne"
		}
		return "faible"
	default:
		return "moyenne"
	}
}

// =======================
// CRUD OPERATIONS
// =======================

// Paginate - Récupérer les données biométriques avec pagination
func GetPaginatedBiometries(c *fiber.Ctx) error {
	db := database.DB

	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page <= 0 {
		page = 1
	}
	limit, err := strconv.Atoi(c.Query("limit", "15"))
	if err != nil || limit <= 0 {
		limit = 15
	}
	offset := (page - 1) * limit

	migrantUUID := c.Query("migrant_uuid", "")
	typeBiometrie := c.Query("type_biometrie", "")
	verifie := c.Query("verifie", "")

	var biometries []models.Biometrie
	var totalRecords int64

	query := db.Model(&models.Biometrie{}).
		Preload("Migrant").
		Select("uuid, migrant_uuid, type_biometrie, index_doigt, qualite_donnee, algorithme_encodage, taille_fichier, date_capture, dispositif_capture, verifie, score_confiance, chiffre, created_at, updated_at") // Exclure les données sensibles

	// Filtrer par migrant si spécifié
	if migrantUUID != "" {
		query = query.Where("migrant_uuid = ?", migrantUUID)
	}

	// Filtrer par type de biométrie
	if typeBiometrie != "" {
		query = query.Where("type_biometrie = ?", typeBiometrie)
	}

	// Filtrer par statut de vérification
	if verifie != "" {
		isVerified := verifie == "true"
		query = query.Where("verifie = ?", isVerified)
	}

	// Count total
	query.Count(&totalRecords)

	// Get paginated results
	err = query.Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&biometries).Error

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch biometric data",
			"error":   err.Error(),
		})
	}

	totalPages := int((totalRecords + int64(limit) - 1) / int64(limit))

	pagination := map[string]interface{}{
		"total_records": totalRecords,
		"total_pages":   totalPages,
		"current_page":  page,
		"page_size":     limit,
	}

	return c.JSON(fiber.Map{
		"status":     "success",
		"message":    "Biometric data retrieved successfully",
		"data":       biometries,
		"pagination": pagination,
	})
}

// Get all biometries (without sensitive data)
func GetAllBiometries(c *fiber.Ctx) error {
	db := database.DB
	var biometries []models.Biometrie

	err := db.Select("uuid, migrant_uuid, type_biometrie, index_doigt, qualite_donnee, algorithme_encodage, taille_fichier, date_capture, dispositif_capture, verifie, score_confiance, chiffre, created_at, updated_at").
		Preload("Migrant").
		Find(&biometries).Error

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch biometric data",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "All biometric data",
		"data":    biometries,
	})
}

// Get one biometry (without sensitive data by default)
func GetBiometrie(c *fiber.Ctx) error {
	uuid := c.Params("uuid")
	includeSensitive := c.Query("include_sensitive", "false")

	db := database.DB
	var biometrie models.Biometrie

	query := db.Where("uuid = ?", uuid).Preload("Migrant")

	// Exclure les données sensibles par défaut
	if includeSensitive != "true" {
		query = query.Select("uuid, migrant_uuid, type_biometrie, index_doigt, qualite_donnee, algorithme_encodage, taille_fichier, date_capture, dispositif_capture, verifie, score_confiance, chiffre, created_at, updated_at")
	}

	err := query.First(&biometrie).Error
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Biometric data not found",
			"data":    nil,
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Biometric data found",
		"data":    biometrie,
	})
}

// Get biometries by migrant
func GetBiometriesByMigrant(c *fiber.Ctx) error {
	migrantUUID := c.Params("migrant_uuid")
	db := database.DB
	var biometries []models.Biometrie

	err := db.Where("migrant_uuid = ?", migrantUUID).
		Select("uuid, migrant_uuid, type_biometrie, index_doigt, qualite_donnee, algorithme_encodage, taille_fichier, date_capture, dispositif_capture, verifie, score_confiance, chiffre, created_at, updated_at").
		Preload("Migrant").
		Order("created_at DESC").
		Find(&biometries).Error

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch biometric data for migrant",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Biometric data for migrant",
		"data":    biometries,
	})
}

// Get verified biometries
func GetVerifiedBiometries(c *fiber.Ctx) error {
	db := database.DB
	var biometries []models.Biometrie

	err := db.Where("verifie = ?", true).
		Select("uuid, migrant_uuid, type_biometrie, index_doigt, qualite_donnee, algorithme_encodage, taille_fichier, date_capture, dispositif_capture, verifie, score_confiance, chiffre, created_at, updated_at").
		Preload("Migrant").
		Order("score_confiance DESC").
		Find(&biometries).Error

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch verified biometric data",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Verified biometric data",
		"data":    biometries,
	})
}

// Create biometry
func CreateBiometrie(c *fiber.Ctx) error {
	biometrie := &models.Biometrie{}

	if err := c.BodyParser(biometrie); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request format",
			"error":   err.Error(),
		})
	}

	// Validation des champs requis
	if biometrie.MigrantUUID == "" || biometrie.TypeBiometrie == "" || biometrie.DonneesBiometriques == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "MigrantUUID, TypeBiometrie, and DonneesBiometriques are required",
			"data":    nil,
		})
	}

	// Vérifier que le migrant existe
	var migrant models.Migrant
	if err := database.DB.Where("uuid = ?", biometrie.MigrantUUID).First(&migrant).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Migrant not found",
			"data":    nil,
		})
	}

	// Générer l'UUID
	biometrie.UUID = utils.GenerateUUID()

	// Chiffrer les données biométriques
	encryptedData, encryptionKey, err := encryptBiometricData(biometrie.DonneesBiometriques)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to encrypt biometric data",
			"error":   err.Error(),
		})
	}

	biometrie.DonneesBiometriques = encryptedData
	biometrie.CleChiffrement = encryptionKey
	biometrie.Chiffre = true

	// Calculer la taille des données
	biometrie.TailleFichier = len(biometrie.DonneesBiometriques)

	// Évaluer la qualité des données
	if biometrie.QualiteDonnee == "" {
		biometrie.QualiteDonnee = assessDataQuality(biometrie.TailleFichier, biometrie.TypeBiometrie)
	}

	// Validation des données
	if err := utils.ValidateStruct(*biometrie); err != nil {
		return c.Status(400).JSON(err)
	}

	// Vérifier l'unicité pour le même type de biométrie et migrant
	var existingBiometrie models.Biometrie
	query := database.DB.Where("migrant_uuid = ? AND type_biometrie = ?", biometrie.MigrantUUID, biometrie.TypeBiometrie)

	// Pour les empreintes digitales, vérifier aussi l'index du doigt
	if biometrie.TypeBiometrie == "empreinte_digitale" && biometrie.IndexDoigt != nil {
		query = query.Where("index_doigt = ?", *biometrie.IndexDoigt)
	}

	if err := query.First(&existingBiometrie).Error; err == nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Biometric data of this type already exists for this migrant",
			"data":    nil,
		})
	}

	if err := database.DB.Create(biometrie).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to create biometric data",
			"error":   err.Error(),
		})
	}

	// Recharger sans les données sensibles
	database.DB.Select("uuid, migrant_uuid, type_biometrie, index_doigt, qualite_donnee, algorithme_encodage, taille_fichier, date_capture, dispositif_capture, verifie, score_confiance, chiffre, created_at, updated_at").
		Preload("Migrant").
		First(biometrie, "uuid = ?", biometrie.UUID)

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Biometric data created successfully",
		"data":    biometrie,
	})
}

// Verify biometry
func VerifyBiometrie(c *fiber.Ctx) error {
	uuid := c.Params("uuid")
	db := database.DB

	var verificationData struct {
		ScoreConfiance        float64 `json:"score_confiance"`
		OperateurVerification string  `json:"operateur_verification"`
	}

	if err := c.BodyParser(&verificationData); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request format",
			"error":   err.Error(),
		})
	}

	var biometrie models.Biometrie
	if err := db.Where("uuid = ?", uuid).First(&biometrie).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Biometric data not found",
			"data":    nil,
		})
	}

	// Marquer comme vérifié
	now := time.Now()
	updateData := map[string]interface{}{
		"verifie":           true,
		"date_verification": &now,
		"score_confiance":   verificationData.ScoreConfiance,
	}

	if err := db.Model(&biometrie).Updates(updateData).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to verify biometric data",
			"error":   err.Error(),
		})
	}

	// Recharger sans données sensibles
	db.Select("uuid, migrant_uuid, type_biometrie, index_doigt, qualite_donnee, algorithme_encodage, taille_fichier, date_capture, dispositif_capture, verifie, score_confiance, chiffre, created_at, updated_at").
		Preload("Migrant").
		First(&biometrie)

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Biometric data verified successfully",
		"data":    biometrie,
	})
}

// Update biometry (metadata only, not the biometric data itself)
func UpdateBiometrie(c *fiber.Ctx) error {
	uuid := c.Params("uuid")
	db := database.DB

	var updateData struct {
		QualiteDonnee     string `json:"qualite_donnee"`
		DisposifCapture   string `json:"dispositif_capture"`
		ResolutionCapture string `json:"resolution_capture"`
		OperateurCapture  string `json:"operateur_capture"`
	}

	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Review your input",
			"error":   err.Error(),
		})
	}

	biometrie := new(models.Biometrie)
	if err := db.Where("uuid = ?", uuid).First(&biometrie).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Biometric data not found",
			"data":    nil,
		})
	}

	if err := db.Model(&biometrie).Updates(updateData).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to update biometric data",
			"error":   err.Error(),
		})
	}

	// Recharger sans données sensibles
	db.Select("uuid, migrant_uuid, type_biometrie, index_doigt, qualite_donnee, algorithme_encodage, taille_fichier, date_capture, dispositif_capture, verifie, score_confiance, chiffre, created_at, updated_at").
		Preload("Migrant").
		First(&biometrie)

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Biometric data updated successfully",
		"data":    biometrie,
	})
}

// Delete biometry
func DeleteBiometrie(c *fiber.Ctx) error {
	uuid := c.Params("uuid")
	db := database.DB

	var biometrie models.Biometrie
	if err := db.Where("uuid = ?", uuid).First(&biometrie).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Biometric data not found",
			"data":    nil,
		})
	}

	if err := db.Delete(&biometrie).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to delete biometric data",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Biometric data deleted successfully",
		"data":    nil,
	})
}

// =======================
// ANALYTICS & STATISTICS
// =======================

// Get biometrics statistics
func GetBiometricsStats(c *fiber.Ctx) error {
	db := database.DB

	var totalBiometrics int64
	var verifiedBiometrics int64
	var encryptedBiometrics int64

	// Statistiques générales
	db.Model(&models.Biometrie{}).Count(&totalBiometrics)
	db.Model(&models.Biometrie{}).Where("verifie = ?", true).Count(&verifiedBiometrics)
	db.Model(&models.Biometrie{}).Where("chiffre = ?", true).Count(&encryptedBiometrics)

	// Statistiques par type de biométrie
	var biometricTypes []map[string]interface{}
	db.Model(&models.Biometrie{}).
		Select("type_biometrie, COUNT(*) as count").
		Group("type_biometrie").
		Order("count DESC").
		Scan(&biometricTypes)

	// Statistiques par qualité
	var qualityStats []map[string]interface{}
	db.Model(&models.Biometrie{}).
		Select("qualite_donnee, COUNT(*) as count").
		Group("qualite_donnee").
		Order("count DESC").
		Scan(&qualityStats)

	// Score de confiance moyen
	var avgConfidenceScore float64
	db.Model(&models.Biometrie{}).
		Where("score_confiance IS NOT NULL").
		Select("AVG(score_confiance)").
		Scan(&avgConfidenceScore)

	// Dispositifs de capture les plus utilisés
	var captureDevices []map[string]interface{}
	db.Model(&models.Biometrie{}).
		Where("dispositif_capture IS NOT NULL AND dispositif_capture != ''").
		Select("dispositif_capture, COUNT(*) as count").
		Group("dispositif_capture").
		Order("count DESC").
		Limit(10).
		Scan(&captureDevices)

	stats := map[string]interface{}{
		"total_biometrics":     totalBiometrics,
		"verified_biometrics":  verifiedBiometrics,
		"encrypted_biometrics": encryptedBiometrics,
		"biometric_types":      biometricTypes,
		"quality_distribution": qualityStats,
		"avg_confidence_score": avgConfidenceScore,
		"capture_devices":      captureDevices,
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Biometrics statistics",
		"data":    stats,
	})
}
 
