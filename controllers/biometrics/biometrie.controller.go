package biometrics

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/kgermando/sysmobembo-api/database"
	"github.com/kgermando/sysmobembo-api/models"
	"github.com/kgermando/sysmobembo-api/utils"
	"github.com/xuri/excelize/v2"
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

	// Récupérer le paramètre de recherche
	search := c.Query("search", "")

	var biometries []models.Biometrie
	var totalRecords int64

	query := db.Model(&models.Biometrie{}).
		Preload("Migrant").
		Joins("LEFT JOIN migrants ON migrants.uuid = biometries.migrant_uuid").
		Select("biometries.uuid, biometries.migrant_uuid, migrants.numero_identifiant, biometries.type_biometrie, biometries.index_doigt, biometries.qualite_donnee, biometries.algorithme_encodage, biometries.taille_fichier, biometries.date_capture, biometries.disposif_capture, biometries.verifie, biometries.score_confiance, biometries.chiffre, biometries.created_at, biometries.updated_at") // Exclure les données sensibles

	// Appliquer la recherche si le paramètre est fourni
	if search != "" {
		query = query.Joins("LEFT JOIN migrants ON migrants.uuid = biometries.migrant_uuid").
			Where("migrants.numero_identifiant ILIKE ? OR biometries.type_biometrie ILIKE ? OR biometries.qualite_donnee ILIKE ? OR biometries.disposif_capture ILIKE ?",
				"%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	// Count total
	query.Count(&totalRecords)

	// Get paginated results
	err = query.Offset(offset).
		Limit(limit).
		Order("biometries.created_at DESC").
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

	err := db.Model(&models.Biometrie{}).
		Joins("LEFT JOIN migrants ON migrants.uuid = biometries.migrant_uuid").
		Select("biometries.uuid, biometries.migrant_uuid, migrants.numero_identifiant, biometries.type_biometrie, biometries.index_doigt, biometries.qualite_donnee, biometries.algorithme_encodage, biometries.taille_fichier, biometries.date_capture, biometries.disposif_capture, biometries.verifie, biometries.score_confiance, biometries.chiffre, biometries.created_at, biometries.updated_at").
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
		query = query.Select("uuid, migrant_uuid, type_biometrie, index_doigt, qualite_donnee, algorithme_encodage, taille_fichier, date_capture, disposif_capture, verifie, score_confiance, chiffre, created_at, updated_at")
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

// Get biometries by migrant with pagination
func GetBiometriesByMigrant(c *fiber.Ctx) error {
	migrantUUID := c.Params("migrant_uuid")
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

	var biometries []models.Biometrie
	var totalRecords int64

	query := db.Model(&models.Biometrie{}).
		Joins("LEFT JOIN migrants ON migrants.uuid = biometries.migrant_uuid").
		Where("biometries.migrant_uuid = ?", migrantUUID).
		Select("biometries.uuid, biometries.migrant_uuid, migrants.numero_identifiant, biometries.type_biometrie, biometries.index_doigt, biometries.qualite_donnee, biometries.algorithme_encodage, biometries.taille_fichier, biometries.date_capture, biometries.disposif_capture, biometries.verifie, biometries.score_confiance, biometries.chiffre, biometries.created_at, biometries.updated_at")

	// Count total
	query.Count(&totalRecords)

	// Get paginated results
	err = query.Preload("Migrant").
		Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&biometries).Error

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch biometric data for migrant",
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
		"message":    "Biometric data for migrant",
		"data":       biometries,
		"pagination": pagination,
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

	if err := database.DB.Create(biometrie).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to create biometric data",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Biometric data created successfully",
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
	var avgConfidenceScore *float64
	db.Model(&models.Biometrie{}).
		Where("score_confiance IS NOT NULL").
		Select("AVG(score_confiance)").
		Scan(&avgConfidenceScore)

	// Dispositifs de capture les plus utilisés
	var captureDevices []map[string]interface{}
	db.Model(&models.Biometrie{}).
		Where("disposif_capture IS NOT NULL AND disposif_capture != ''").
		Select("disposif_capture, COUNT(*) as count").
		Group("disposif_capture").
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

// =======================
// EXCEL EXPORT
// =======================

// ExportBiometriesToExcel - Exporter les biométries vers Excel avec mise en forme
func ExportBiometriesToExcel(c *fiber.Ctx) error {
	db := database.DB

	// Récupérer les paramètres de date
	startDate := c.Query("start_date", "")
	endDate := c.Query("end_date", "")

	var biometries []models.Biometrie

	query := db.Model(&models.Biometrie{}).Preload("Migrant")

	// Appliquer les filtres de date
	if startDate != "" {
		parsedStartDate, err := time.Parse("2006-01-02", startDate)
		if err == nil {
			query = query.Where("biometries.created_at >= ?", parsedStartDate)
		}
	}
	if endDate != "" {
		parsedEndDate, err := time.Parse("2006-01-02", endDate)
		if err == nil {
			// Ajouter 23:59:59 pour inclure toute la journée
			parsedEndDate = parsedEndDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			query = query.Where("biometries.created_at <= ?", parsedEndDate)
		}
	}

	// Récupérer toutes les données
	err := query.Order("created_at DESC").Find(&biometries).Error
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch biometrics for export",
			"error":   err.Error(),
		})
	}

	// Créer un nouveau fichier Excel
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	// Supprimer la feuille par défaut et créer notre feuille
	f.DeleteSheet("Sheet1")
	index, err := f.NewSheet("Biométries")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to create Excel sheet",
			"error":   err.Error(),
		})
	}
	f.SetActiveSheet(index)

	// ===== STYLES =====
	// Style pour l'en-tête principal
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   16,
			Color:  "FFFFFF",
			Family: "Calibri",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#2E75B6"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to create header style",
			"error":   err.Error(),
		})
	}

	// Style pour les en-têtes de colonnes
	columnHeaderStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   12,
			Color:  "FFFFFF",
			Family: "Calibri",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#4F81BD"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to create column header style",
			"error":   err.Error(),
		})
	}

	// Style pour les cellules de données
	dataStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   11,
			Family: "Calibri",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "left",
			Vertical:   "center",
			WrapText:   true,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "CCCCCC", Style: 1},
			{Type: "top", Color: "CCCCCC", Style: 1},
			{Type: "bottom", Color: "CCCCCC", Style: 1},
			{Type: "right", Color: "CCCCCC", Style: 1},
		},
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to create data style",
			"error":   err.Error(),
		})
	}

	// Style pour les cellules numériques
	numberStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   11,
			Family: "Calibri",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "CCCCCC", Style: 1},
			{Type: "top", Color: "CCCCCC", Style: 1},
			{Type: "bottom", Color: "CCCCCC", Style: 1},
			{Type: "right", Color: "CCCCCC", Style: 1},
		},
		NumFmt: 1,
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to create number style",
			"error":   err.Error(),
		})
	}

	// Style pour les cellules de date
	dateStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   11,
			Family: "Calibri",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "CCCCCC", Style: 1},
			{Type: "top", Color: "CCCCCC", Style: 1},
			{Type: "bottom", Color: "CCCCCC", Style: 1},
			{Type: "right", Color: "CCCCCC", Style: 1},
		},
		NumFmt: 14,
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to create date style",
			"error":   err.Error(),
		})
	}

	// Style pour les cellules booléennes
	booleanStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   11,
			Family: "Calibri",
			Bold:   true,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "CCCCCC", Style: 1},
			{Type: "top", Color: "CCCCCC", Style: 1},
			{Type: "bottom", Color: "CCCCCC", Style: 1},
			{Type: "right", Color: "CCCCCC", Style: 1},
		},
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to create boolean style",
			"error":   err.Error(),
		})
	}

	// Style pour le score de confiance
	scoreStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   11,
			Family: "Calibri",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "CCCCCC", Style: 1},
			{Type: "top", Color: "CCCCCC", Style: 1},
			{Type: "bottom", Color: "CCCCCC", Style: 1},
			{Type: "right", Color: "CCCCCC", Style: 1},
		},
		NumFmt: 4,
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to create score style",
			"error":   err.Error(),
		})
	}

	// ===== EN-TÊTE PRINCIPAL =====
	currentTime := time.Now().Format("02/01/2006 15:04")
	mainHeader := fmt.Sprintf("RAPPORT D'EXPORT DES BIOMÉTRIES - %s", currentTime)
	f.SetCellValue("Biométries", "A1", mainHeader)
	f.MergeCell("Biométries", "A1", "R1")
	f.SetCellStyle("Biométries", "A1", "R1", headerStyle)
	f.SetRowHeight("Biométries", 1, 30)

	// ===== INFORMATIONS DE FILTRE =====
	row := 3
	filterApplied := false
	if startDate != "" || endDate != "" {
		f.SetCellValue("Biométries", "A2", "Filtres appliqués:")
		f.SetCellStyle("Biométries", "A2", "A2", columnHeaderStyle)
		filterApplied = true

		if startDate != "" {
			f.SetCellValue("Biométries", fmt.Sprintf("A%d", row), fmt.Sprintf("Date de début: %s", startDate))
			row++
		}
		if endDate != "" {
			f.SetCellValue("Biométries", fmt.Sprintf("A%d", row), fmt.Sprintf("Date de fin: %s", endDate))
			row++
		}
		row++
	}

	if !filterApplied {
		row = 2
	}

	// ===== EN-TÊTES DE COLONNES =====
	headers := []string{
		"UUID",
		"N° Identifiant Migrant",
		"Type de biométrie",
		"Index du doigt",
		"Qualité des données",
		"Algorithme d'encodage",
		"Taille fichier (bytes)",
		"Date de capture",
		"Dispositif de capture",
		"Résolution de capture",
		"Opérateur de capture",
		"Chiffré",
		"Vérifié",
		"Date de vérification",
		"Score de confiance",
		"Date de création",
		"Date de MAJ",
		"Données biométriques (tronquées)",
	}

	// Écrire les en-têtes
	for i, header := range headers {
		cell := fmt.Sprintf("%c%d", 'A'+i, row)
		f.SetCellValue("Biométries", cell, header)
		f.SetCellStyle("Biométries", cell, cell, columnHeaderStyle)
	}
	f.SetRowHeight("Biométries", row, 25)

	// ===== DONNÉES =====
	for i, biometrie := range biometries {
		dataRow := row + 1 + i

		// UUID
		cell := fmt.Sprintf("A%d", dataRow)
		f.SetCellValue("Biométries", cell, biometrie.UUID)
		f.SetCellStyle("Biométries", cell, cell, dataStyle)

		// N° Identifiant Migrant
		cell = fmt.Sprintf("B%d", dataRow)
		if biometrie.Migrant.NumeroIdentifiant != "" {
			f.SetCellValue("Biométries", cell, biometrie.Migrant.NumeroIdentifiant)
		} else {
			f.SetCellValue("Biométries", cell, "N/A")
		}
		f.SetCellStyle("Biométries", cell, cell, dataStyle)

		// Type de biométrie
		cell = fmt.Sprintf("C%d", dataRow)
		f.SetCellValue("Biométries", cell, biometrie.TypeBiometrie)
		f.SetCellStyle("Biométries", cell, cell, dataStyle)

		// Index du doigt
		cell = fmt.Sprintf("D%d", dataRow)
		if biometrie.IndexDoigt != nil {
			f.SetCellValue("Biométries", cell, *biometrie.IndexDoigt)
		} else {
			f.SetCellValue("Biométries", cell, "")
		}
		f.SetCellStyle("Biométries", cell, cell, numberStyle)

		// Qualité des données
		cell = fmt.Sprintf("E%d", dataRow)
		f.SetCellValue("Biométries", cell, biometrie.QualiteDonnee)
		f.SetCellStyle("Biométries", cell, cell, dataStyle)

		// Algorithme d'encodage
		cell = fmt.Sprintf("F%d", dataRow)
		f.SetCellValue("Biométries", cell, biometrie.AlgorithmeEncodage)
		f.SetCellStyle("Biométries", cell, cell, dataStyle)

		// Taille fichier
		cell = fmt.Sprintf("G%d", dataRow)
		if biometrie.TailleFichier > 0 {
			f.SetCellValue("Biométries", cell, biometrie.TailleFichier)
		} else {
			f.SetCellValue("Biométries", cell, "")
		}
		f.SetCellStyle("Biométries", cell, cell, numberStyle)

		// Date de capture
		cell = fmt.Sprintf("H%d", dataRow)
		f.SetCellValue("Biométries", cell, biometrie.DateCapture.Format("02/01/2006 15:04"))
		f.SetCellStyle("Biométries", cell, cell, dateStyle)

		// Dispositif de capture
		cell = fmt.Sprintf("I%d", dataRow)
		f.SetCellValue("Biométries", cell, biometrie.DisposifCapture)
		f.SetCellStyle("Biométries", cell, cell, dataStyle)

		// Résolution de capture
		cell = fmt.Sprintf("J%d", dataRow)
		f.SetCellValue("Biométries", cell, biometrie.ResolutionCapture)
		f.SetCellStyle("Biométries", cell, cell, dataStyle)

		// Opérateur de capture
		cell = fmt.Sprintf("K%d", dataRow)
		f.SetCellValue("Biométries", cell, biometrie.OperateurCapture)
		f.SetCellStyle("Biométries", cell, cell, dataStyle)

		// Chiffré
		cell = fmt.Sprintf("L%d", dataRow)
		if biometrie.Chiffre {
			f.SetCellValue("Biométries", cell, "OUI")
		} else {
			f.SetCellValue("Biométries", cell, "NON")
		}
		f.SetCellStyle("Biométries", cell, cell, booleanStyle)

		// Vérifié
		cell = fmt.Sprintf("M%d", dataRow)
		if biometrie.Verifie {
			f.SetCellValue("Biométries", cell, "OUI")
		} else {
			f.SetCellValue("Biométries", cell, "NON")
		}
		f.SetCellStyle("Biométries", cell, cell, booleanStyle)

		// Date de vérification
		cell = fmt.Sprintf("N%d", dataRow)
		if biometrie.DateVerification != nil {
			f.SetCellValue("Biométries", cell, biometrie.DateVerification.Format("02/01/2006 15:04"))
		} else {
			f.SetCellValue("Biométries", cell, "")
		}
		f.SetCellStyle("Biométries", cell, cell, dateStyle)

		// Score de confiance
		cell = fmt.Sprintf("O%d", dataRow)
		if biometrie.ScoreConfiance != nil {
			f.SetCellValue("Biométries", cell, *biometrie.ScoreConfiance)
		} else {
			f.SetCellValue("Biométries", cell, "")
		}
		f.SetCellStyle("Biométries", cell, cell, scoreStyle)

		// Date de création
		cell = fmt.Sprintf("P%d", dataRow)
		f.SetCellValue("Biométries", cell, biometrie.CreatedAt.Format("02/01/2006 15:04"))
		f.SetCellStyle("Biométries", cell, cell, dateStyle)

		// Date de MAJ
		cell = fmt.Sprintf("Q%d", dataRow)
		f.SetCellValue("Biométries", cell, biometrie.UpdatedAt.Format("02/01/2006 15:04"))
		f.SetCellStyle("Biométries", cell, cell, dateStyle)

		// Données biométriques (tronquées pour sécurité)
		cell = fmt.Sprintf("R%d", dataRow)
		if len(biometrie.DonneesBiometriques) > 50 {
			f.SetCellValue("Biométries", cell, biometrie.DonneesBiometriques[:50]+"...")
		} else {
			f.SetCellValue("Biométries", cell, biometrie.DonneesBiometriques)
		}
		f.SetCellStyle("Biométries", cell, cell, dataStyle)

		// Définir la hauteur de ligne
		f.SetRowHeight("Biométries", dataRow, 20)
	}

	// ===== AJUSTEMENT DE LA LARGEUR DES COLONNES =====
	columnWidths := []float64{
		25, // UUID
		20, // N° Identifiant Migrant
		20, // Type biométrie
		8,  // Index doigt
		15, // Qualité
		20, // Algorithme
		12, // Taille fichier
		18, // Date capture
		20, // Dispositif capture
		15, // Résolution
		20, // Opérateur
		8,  // Chiffré
		8,  // Vérifié
		18, // Date vérification
		12, // Score confiance
		18, // Date création
		18, // Date MAJ
		30, // Données biométriques
	}

	columns := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R"}
	for i, width := range columnWidths {
		if i < len(columns) {
			f.SetColWidth("Biométries", columns[i], columns[i], width)
		}
	}

	// ===== AJOUTER UNE FEUILLE DE STATISTIQUES =====
	_, err = f.NewSheet("Statistiques")
	if err == nil {
		// Calculer les statistiques
		totalRecords := len(biometries)

		// Compter par type de biométrie
		typeCount := make(map[string]int)
		qualiteCount := make(map[string]int)
		dispositifCount := make(map[string]int)
		verifiedCount := 0
		encryptedCount := 0
		totalScore := 0.0
		scoreCount := 0

		for _, bio := range biometries {
			typeCount[bio.TypeBiometrie]++
			qualiteCount[bio.QualiteDonnee]++
			if bio.DisposifCapture != "" {
				dispositifCount[bio.DisposifCapture]++
			}
			if bio.Verifie {
				verifiedCount++
			}
			if bio.Chiffre {
				encryptedCount++
			}
			if bio.ScoreConfiance != nil {
				totalScore += *bio.ScoreConfiance
				scoreCount++
			}
		}

		avgScore := 0.0
		if scoreCount > 0 {
			avgScore = totalScore / float64(scoreCount)
		}

		// En-tête de la feuille statistiques
		f.SetCellValue("Statistiques", "A1", "STATISTIQUES DES BIOMÉTRIES")
		f.MergeCell("Statistiques", "A1", "C1")
		f.SetCellStyle("Statistiques", "A1", "C1", headerStyle)

		row = 3
		f.SetCellValue("Statistiques", fmt.Sprintf("A%d", row), "Total des enregistrements:")
		f.SetCellValue("Statistiques", fmt.Sprintf("B%d", row), totalRecords)
		row += 2

		// Statistiques générales
		f.SetCellValue("Statistiques", fmt.Sprintf("A%d", row), "Biométries vérifiées:")
		f.SetCellValue("Statistiques", fmt.Sprintf("B%d", row), verifiedCount)
		row++
		f.SetCellValue("Statistiques", fmt.Sprintf("A%d", row), "Biométries chiffrées:")
		f.SetCellValue("Statistiques", fmt.Sprintf("B%d", row), encryptedCount)
		row++
		f.SetCellValue("Statistiques", fmt.Sprintf("A%d", row), "Score de confiance moyen:")
		f.SetCellValue("Statistiques", fmt.Sprintf("B%d", row), fmt.Sprintf("%.2f", avgScore))
		row += 2

		// Par type de biométrie
		f.SetCellValue("Statistiques", fmt.Sprintf("A%d", row), "Par type de biométrie:")
		f.SetCellStyle("Statistiques", fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), columnHeaderStyle)
		row++
		for typeBio, count := range typeCount {
			f.SetCellValue("Statistiques", fmt.Sprintf("A%d", row), typeBio)
			f.SetCellValue("Statistiques", fmt.Sprintf("B%d", row), count)
			row++
		}
		row++

		// Par qualité des données
		f.SetCellValue("Statistiques", fmt.Sprintf("A%d", row), "Par qualité des données:")
		f.SetCellStyle("Statistiques", fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), columnHeaderStyle)
		row++
		for qualite, count := range qualiteCount {
			f.SetCellValue("Statistiques", fmt.Sprintf("A%d", row), qualite)
			f.SetCellValue("Statistiques", fmt.Sprintf("B%d", row), count)
			row++
		}
		row++

		// Top 5 dispositifs de capture
		f.SetCellValue("Statistiques", fmt.Sprintf("A%d", row), "Top 5 dispositifs de capture:")
		f.SetCellStyle("Statistiques", fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), columnHeaderStyle)
		row++
		count := 0
		for dispositif, nb := range dispositifCount {
			if count >= 5 {
				break
			}
			f.SetCellValue("Statistiques", fmt.Sprintf("A%d", row), dispositif)
			f.SetCellValue("Statistiques", fmt.Sprintf("B%d", row), nb)
			row++
			count++
		}

		f.SetColWidth("Statistiques", "A", "A", 30)
		f.SetColWidth("Statistiques", "B", "B", 15)
	}

	// ===== GÉNÉRATION DU FICHIER =====
	filename := fmt.Sprintf("biometries_export_%s.xlsx", time.Now().Format("20060102_150405"))

	// Sauvegarder en mémoire
	buffer, err := f.WriteToBuffer()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to generate Excel file",
			"error":   err.Error(),
		})
	}

	// Configurer les en-têtes de réponse pour le téléchargement
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))

	return c.Send(buffer.Bytes())
}
