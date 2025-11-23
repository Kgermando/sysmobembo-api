package geolocation

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/kgermando/sysmobembo-api/database"
	"github.com/kgermando/sysmobembo-api/models"
	"github.com/kgermando/sysmobembo-api/utils"
	"github.com/xuri/excelize/v2"
)

// =======================
// GIS UTILITY FUNCTIONS
// =======================

// Calculer la distance entre deux points (formule de Haversine)
func calculateDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371 // Rayon de la Terre en kilomètres

	// Convertir en radians
	lat1Rad := lat1 * math.Pi / 180
	lon1Rad := lon1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lon2Rad := lon2 * math.Pi / 180

	// Différences
	dlat := lat2Rad - lat1Rad
	dlon := lon2Rad - lon1Rad

	// Formule de Haversine
	a := math.Sin(dlat/2)*math.Sin(dlat/2) + math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(dlon/2)*math.Sin(dlon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

// Valider les coordonnées GPS
func validateCoordinates(lat, lon float64) error {
	if lat < -90 || lat > 90 {
		return fmt.Errorf("latitude must be between -90 and 90 degrees")
	}
	if lon < -180 || lon > 180 {
		return fmt.Errorf("longitude must be between -180 and 180 degrees")
	}
	return nil
}

// =======================
// CRUD OPERATIONS
// =======================

// Paginate - Récupérer les géolocalisations avec pagination
func GetPaginatedGeolocalisations(c *fiber.Ctx) error {
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
	typeLocalisation := c.Query("type_localisation", "")
	pays := c.Query("pays", "")

	var geolocalisations []models.Geolocalisation
	var totalRecords int64

	query := db.Model(&models.Geolocalisation{}).Preload("Migrant")

	// Filtrer par migrant si spécifié
	if migrantUUID != "" {
		query = query.Where("migrant_uuid = ?", migrantUUID)
	}

	// Filtrer par type de localisation
	if typeLocalisation != "" {
		query = query.Where("type_localisation = ?", typeLocalisation)
	}

	// Filtrer par pays
	if pays != "" {
		query = query.Where("pays ILIKE ?", "%"+pays+"%")
	}

	// Count total
	query.Count(&totalRecords)

	// Get paginated results
	err = query.Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&geolocalisations).Error

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch geolocations",
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
		"message":    "Geolocations retrieved successfully",
		"data":       geolocalisations,
		"pagination": pagination,
	})
}

// Get all geolocations
func GetAllGeolocalisations(c *fiber.Ctx) error {
	db := database.DB
	var geolocalisations []models.Geolocalisation

	err := db.Preload("Migrant").
		Order("created_at DESC").
		Find(&geolocalisations).Error

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch geolocations",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "All geolocations",
		"data":    geolocalisations,
	})
}

// Get one geolocation
func GetGeolocalisation(c *fiber.Ctx) error {
	uuid := c.Params("uuid")
	db := database.DB
	var geolocalisation models.Geolocalisation

	err := db.Where("uuid = ?", uuid).
		Preload("Migrant").
		First(&geolocalisation).Error

	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Geolocation not found",
			"data":    nil,
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Geolocation found",
		"data":    geolocalisation,
	})
}

// Get geolocations by migrant with pagination
func GetGeolocalisationsByMigrant(c *fiber.Ctx) error {
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

	var geolocalisations []models.Geolocalisation
	var totalRecords int64

	query := db.Model(&models.Geolocalisation{}).
		Where("migrant_uuid = ?", migrantUUID).
		Preload("Migrant")

	// Count total
	query.Count(&totalRecords)

	// Get paginated results
	err = query.Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&geolocalisations).Error

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch geolocations for migrant",
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
		"message":    "Geolocations for migrant retrieved successfully",
		"data":       geolocalisations,
		"pagination": pagination,
	})
}

// Create geolocation
func CreateGeolocalisation(c *fiber.Ctx) error {
	geolocalisation := &models.Geolocalisation{}

	if err := c.BodyParser(geolocalisation); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request format",
			"error":   err.Error(),
		})
	}

	// Validation des champs requis
	if geolocalisation.MigrantUUID == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "MigrantUUID is required",
			"data":    nil,
		})
	}

	// Valider les coordonnées
	if err := validateCoordinates(geolocalisation.Latitude, geolocalisation.Longitude); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
			"data":    nil,
		})
	}

	// Générer l'UUID
	geolocalisation.UUID = utils.GenerateUUID()

	// Validation des données
	if err := utils.ValidateStruct(*geolocalisation); err != nil {
		return c.Status(400).JSON(err)
	}

	if err := database.DB.Create(geolocalisation).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to create geolocation",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Geolocation created successfully",
		"data":    geolocalisation,
	})
}

// Update geolocation
func UpdateGeolocalisation(c *fiber.Ctx) error {
	uuid := c.Params("uuid")
	db := database.DB

	var updateData models.Geolocalisation
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Review your input",
			"error":   err.Error(),
		})
	}

	geolocalisation := new(models.Geolocalisation)
	if err := db.Where("uuid = ?", uuid).First(&geolocalisation).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Geolocation not found",
			"data":    nil,
		})
	}

	// Valider les nouvelles coordonnées si elles sont fournies
	if updateData.Latitude != 0 || updateData.Longitude != 0 {
		if err := validateCoordinates(updateData.Latitude, updateData.Longitude); err != nil {
			return c.Status(400).JSON(fiber.Map{
				"status":  "error",
				"message": err.Error(),
				"data":    nil,
			})
		}
	}

	// Conserver l'UUID
	updateData.UUID = geolocalisation.UUID

	if err := db.Model(&geolocalisation).Updates(updateData).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to update geolocation",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Geolocation updated successfully",
		"data":    geolocalisation,
	})
}

// Delete geolocation
func DeleteGeolocalisation(c *fiber.Ctx) error {
	uuid := c.Params("uuid")
	db := database.DB

	var geolocalisation models.Geolocalisation
	if err := db.Where("uuid = ?", uuid).First(&geolocalisation).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Geolocation not found",
			"data":    nil,
		})
	}

	if err := db.Delete(&geolocalisation).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to delete geolocation",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Geolocation deleted successfully",
		"data":    nil,
	})
}

// =======================
// EXCEL EXPORT
// =======================

// ExportGeolocalisationsToExcel - Exporter les géolocalisations vers Excel avec mise en forme
func ExportGeolocalisationsToExcel(c *fiber.Ctx) error {
	db := database.DB

	// Récupérer les paramètres de filtre de date
	startDateStr := c.Query("start_date", "")
	endDateStr := c.Query("end_date", "")

	var geolocalisations []models.Geolocalisation

	query := db.Model(&models.Geolocalisation{}).Preload("Migrant").Preload("Migrant.Identite")

	// Appliquer les filtres de date
	if startDateStr != "" {
		startDate, err := time.Parse("2006-01-02", startDateStr)
		if err == nil {
			query = query.Where("created_at >= ?", startDate)
		}
	}
	if endDateStr != "" {
		endDate, err := time.Parse("2006-01-02", endDateStr)
		if err == nil {
			// Ajouter 23:59:59 pour inclure toute la journée
			endDate = endDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			query = query.Where("created_at <= ?", endDate)
		}
	}

	// Récupérer toutes les données
	err := query.Order("created_at DESC").Find(&geolocalisations).Error
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch geolocations for export",
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
	index, err := f.NewSheet("Géolocalisations")
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
		NumFmt: 4, // Format numérique avec 2 décimales
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
		NumFmt: 14, // Format de date mm/dd/yyyy
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to create date style",
			"error":   err.Error(),
		})
	}

	// ===== EN-TÊTE PRINCIPAL =====
	currentTime := time.Now().Format("02/01/2006 15:04")
	mainHeader := fmt.Sprintf("RAPPORT D'EXPORT DES GÉOLOCALISATIONS - %s", currentTime)
	f.SetCellValue("Géolocalisations", "A1", mainHeader)
	f.MergeCell("Géolocalisations", "A1", "E1")
	f.SetCellStyle("Géolocalisations", "A1", "E1", headerStyle)
	f.SetRowHeight("Géolocalisations", 1, 30)

	// ===== INFORMATIONS DE FILTRE =====
	row := 3
	if startDateStr != "" || endDateStr != "" {
		f.SetCellValue("Géolocalisations", "A2", "Filtres appliqués:")
		f.SetCellStyle("Géolocalisations", "A2", "A2", columnHeaderStyle)

		if startDateStr != "" {
			f.SetCellValue("Géolocalisations", fmt.Sprintf("A%d", row), fmt.Sprintf("Date de début: %s", startDateStr))
			row++
		}
		if endDateStr != "" {
			f.SetCellValue("Géolocalisations", fmt.Sprintf("A%d", row), fmt.Sprintf("Date de fin: %s", endDateStr))
			row++
		}
		row++ // Ligne vide
	}

	// ===== EN-TÊTES DE COLONNES =====
	headers := []string{
		"UUID",
		"Date de création",
		"Numéro d'identifiant",
		"Latitude",
		"Longitude",
	}

	// Écrire les en-têtes
	for i, header := range headers {
		cell := fmt.Sprintf("%c%d", 'A'+i, row)
		f.SetCellValue("Géolocalisations", cell, header)
		f.SetCellStyle("Géolocalisations", cell, cell, columnHeaderStyle)
	}
	f.SetRowHeight("Géolocalisations", row, 25)

	// ===== DONNÉES =====
	for i, geo := range geolocalisations {
		dataRow := row + 1 + i

		// UUID
		cell := fmt.Sprintf("A%d", dataRow)
		f.SetCellValue("Géolocalisations", cell, geo.UUID)
		f.SetCellStyle("Géolocalisations", cell, cell, dataStyle)

		// Date de création
		cell = fmt.Sprintf("B%d", dataRow)
		f.SetCellValue("Géolocalisations", cell, geo.CreatedAt.Format("02/01/2006 15:04"))
		f.SetCellStyle("Géolocalisations", cell, cell, dateStyle)

		// Numéro d'identifiant du migrant
		cell = fmt.Sprintf("C%d", dataRow)
		if geo.Migrant.Identite.UUID != "" {
			f.SetCellValue("Géolocalisations", cell, geo.Migrant.NumeroIdentifiant)
		} else {
			f.SetCellValue("Géolocalisations", cell, "N/A")
		}
		f.SetCellStyle("Géolocalisations", cell, cell, dataStyle)

		// Latitude
		cell = fmt.Sprintf("D%d", dataRow)
		f.SetCellValue("Géolocalisations", cell, geo.Latitude)
		f.SetCellStyle("Géolocalisations", cell, cell, numberStyle)

		// Longitude
		cell = fmt.Sprintf("E%d", dataRow)
		f.SetCellValue("Géolocalisations", cell, geo.Longitude)
		f.SetCellStyle("Géolocalisations", cell, cell, numberStyle)

		// Définir la hauteur de ligne
		f.SetRowHeight("Géolocalisations", dataRow, 20)
	}

	// ===== AJUSTEMENT DE LA LARGEUR DES COLONNES =====
	columnWidths := []float64{
		25, // UUID
		18, // Date création
		20, // Numéro identifiant
		12, // Latitude
		12, // Longitude
	}

	for i, width := range columnWidths {
		col := string(rune('A' + i))
		f.SetColWidth("Géolocalisations", col, col, width)
	}

	// ===== AJOUTER UNE FEUILLE DE STATISTIQUES =====
	_, err = f.NewSheet("Statistiques")
	if err == nil {
		// Calculer les statistiques
		totalRecords := len(geolocalisations)

		// En-tête de la feuille statistiques
		f.SetCellValue("Statistiques", "A1", "STATISTIQUES DES GÉOLOCALISATIONS")
		f.MergeCell("Statistiques", "A1", "C1")
		f.SetCellStyle("Statistiques", "A1", "C1", headerStyle)

		row = 3
		f.SetCellValue("Statistiques", fmt.Sprintf("A%d", row), "Total des enregistrements:")
		f.SetCellValue("Statistiques", fmt.Sprintf("B%d", row), totalRecords)

		f.SetColWidth("Statistiques", "A", "A", 25)
		f.SetColWidth("Statistiques", "B", "B", 15)
	}

	// ===== GÉNÉRATION DU FICHIER =====
	filename := fmt.Sprintf("geolocalisations_export_%s.xlsx", time.Now().Format("20060102_150405"))

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
