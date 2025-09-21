package migrants

import (
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
// CRUD OPERATIONS
// =======================

// Paginate - Récupérer les motifs avec pagination
func GetPaginatedMotifDeplacements(c *fiber.Ctx) error {
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

	search := c.Query("search", "")
	migrantUUID := c.Query("migrant_uuid", "")

	var motifs []models.MotifDeplacement
	var totalRecords int64

	query := db.Model(&models.MotifDeplacement{}).
		Preload("Migrant")

	// Filtrer par migrant si spécifié
	if migrantUUID != "" {
		query = query.Where("migrant_uuid = ?", migrantUUID)
	}

	// Recherche textuelle
	if search != "" {
		query = query.Where("type_motif ILIKE ? OR motif_principal ILIKE ? OR description ILIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	// Count total
	query.Count(&totalRecords)

	// Get paginated results
	err = query.Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&motifs).Error

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch motifs de déplacement",
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
		"message":    "Motifs de déplacement retrieved successfully",
		"data":       motifs,
		"pagination": pagination,
	})
}

// Get all motifs
func GetAllMotifDeplacements(c *fiber.Ctx) error {
	db := database.DB
	var motifs []models.MotifDeplacement

	err := db.Preload("Migrant").Find(&motifs).Error
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch motifs",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "All motifs de déplacement",
		"data":    motifs,
	})
}

// Get one motif
func GetMotifDeplacement(c *fiber.Ctx) error {
	uuid := c.Params("uuid")
	db := database.DB
	var motif models.MotifDeplacement

	err := db.Where("uuid = ?", uuid).
		Preload("Migrant").
		First(&motif).Error

	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Motif de déplacement not found",
			"data":    nil,
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Motif de déplacement found",
		"data":    motif,
	})
}

// Get motifs by migrant with pagination
func GetMotifsByMigrant(c *fiber.Ctx) error {
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

	search := c.Query("search", "")

	var motifs []models.MotifDeplacement
	var totalRecords int64

	query := db.Model(&models.MotifDeplacement{}).
		Preload("Migrant").
		Where("migrant_uuid = ?", migrantUUID)

	// Recherche textuelle
	if search != "" {
		query = query.Where("type_motif ILIKE ? OR motif_principal ILIKE ? OR description ILIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	// Count total
	query.Count(&totalRecords)

	// Get paginated results
	err = query.Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&motifs).Error

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch motifs for migrant",
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
		"message":    "Motifs for migrant retrieved successfully",
		"data":       motifs,
		"pagination": pagination,
	})
}

// Create motif
func CreateMotifDeplacement(c *fiber.Ctx) error {
	motif := &models.MotifDeplacement{}

	if err := c.BodyParser(motif); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request format",
			"error":   err.Error(),
		})
	}

	// Validation des champs requis
	if motif.MigrantUUID == "" || motif.TypeMotif == "" || motif.MotifPrincipal == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "MigrantUUID, TypeMotif, and MotifPrincipal are required",
			"data":    nil,
		})
	}

	// Vérifier que le migrant existe
	var migrant models.Migrant
	if err := database.DB.Where("uuid = ?", motif.MigrantUUID).First(&migrant).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Migrant not found",
			"data":    nil,
		})
	}

	// Générer l'UUID
	motif.UUID = utils.GenerateUUID()

	// Validation des données
	if err := utils.ValidateStruct(*motif); err != nil {
		return c.Status(400).JSON(err)
	}

	if err := database.DB.Create(motif).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to create motif de déplacement",
			"error":   err.Error(),
		})
	}

	// Recharger avec les relations
	database.DB.Preload("Migrant").First(motif, "uuid = ?", motif.UUID)

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Motif de déplacement created successfully",
		"data":    motif,
	})
}

// Update motif
func UpdateMotifDeplacement(c *fiber.Ctx) error {
	uuid := c.Params("uuid")
	db := database.DB

	var updateData models.MotifDeplacement
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Review your input",
			"error":   err.Error(),
		})
	}

	motif := new(models.MotifDeplacement)
	if err := db.Where("uuid = ?", uuid).First(&motif).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Motif de déplacement not found",
			"data":    nil,
		})
	}

	// Conserver l'UUID
	updateData.UUID = motif.UUID

	if err := db.Model(&motif).Updates(updateData).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to update motif de déplacement",
			"error":   err.Error(),
		})
	}

	// Recharger avec les relations
	db.Preload("Migrant").First(&motif)

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Motif de déplacement updated successfully",
		"data":    motif,
	})
}

// Delete motif
func DeleteMotifDeplacement(c *fiber.Ctx) error {
	uuid := c.Params("uuid")
	db := database.DB

	var motif models.MotifDeplacement
	if err := db.Where("uuid = ?", uuid).First(&motif).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Motif de déplacement not found",
			"data":    nil,
		})
	}

	if err := db.Delete(&motif).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to delete motif de déplacement",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Motif de déplacement deleted successfully",
		"data":    nil,
	})
}

// =======================
// ANALYTICS & STATISTICS
// =======================

// Get motifs statistics
func GetMotifsStats(c *fiber.Ctx) error {
	db := database.DB

	var totalMotifs int64
	var motifsVolontaires int64
	var motifsInvolontaires int64

	// Compter par type de motif
	var motifTypes []map[string]interface{}

	db.Model(&models.MotifDeplacement{}).Count(&totalMotifs)
	db.Model(&models.MotifDeplacement{}).Where("caractere_volontaire = ?", true).Count(&motifsVolontaires)
	db.Model(&models.MotifDeplacement{}).Where("caractere_volontaire = ?", false).Count(&motifsInvolontaires)

	// Statistiques par type de motif
	db.Model(&models.MotifDeplacement{}).
		Select("type_motif, COUNT(*) as count").
		Group("type_motif").
		Order("count DESC").
		Scan(&motifTypes)

	// Statistiques par niveau d'urgence
	var urgenceStats []map[string]interface{}
	db.Model(&models.MotifDeplacement{}).
		Select("urgence, COUNT(*) as count").
		Group("urgence").
		Order("count DESC").
		Scan(&urgenceStats)

	// Statistiques par facteurs externes
	var facteursExternes map[string]int64
	var conflitArme, catastrophe, persecution, violence int64

	db.Model(&models.MotifDeplacement{}).Where("conflit_arme = ?", true).Count(&conflitArme)
	db.Model(&models.MotifDeplacement{}).Where("catastrophe_naturelle = ?", true).Count(&catastrophe)
	db.Model(&models.MotifDeplacement{}).Where("persecution = ?", true).Count(&persecution)
	db.Model(&models.MotifDeplacement{}).Where("violence_generalisee = ?", true).Count(&violence)

	facteursExternes = map[string]int64{
		"conflit_arme":          conflitArme,
		"catastrophe_naturelle": catastrophe,
		"persecution":           persecution,
		"violence_generalisee":  violence,
	}

	stats := map[string]interface{}{
		"total_motifs":         totalMotifs,
		"motifs_volontaires":   motifsVolontaires,
		"motifs_involontaires": motifsInvolontaires,
		"types_motifs":         motifTypes,
		"urgence_stats":        urgenceStats,
		"facteurs_externes":    facteursExternes,
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Motifs statistics",
		"data":    stats,
	})
}

// =======================
// EXCEL EXPORT
// =======================

// ExportMotifDeplacementsToExcel - Exporter les motifs de déplacement vers Excel avec mise en forme
func ExportMotifDeplacementsToExcel(c *fiber.Ctx) error {
	db := database.DB

	// Récupérer les paramètres de filtre
	migrantUUID := c.Query("migrant_uuid", "")
	typeMotif := c.Query("type_motif", "")
	urgence := c.Query("urgence", "")
	caractereVolontaire := c.Query("caractere_volontaire", "")
	search := c.Query("search", "")

	var motifs []models.MotifDeplacement

	query := db.Model(&models.MotifDeplacement{}).Preload("Migrant")

	// Appliquer les filtres
	if migrantUUID != "" {
		query = query.Where("migrant_uuid = ?", migrantUUID)
	}
	if typeMotif != "" {
		query = query.Where("type_motif = ?", typeMotif)
	}
	if urgence != "" {
		query = query.Where("urgence = ?", urgence)
	}
	if caractereVolontaire != "" {
		if caractereVolontaire == "true" {
			query = query.Where("caractere_volontaire = ?", true)
		} else if caractereVolontaire == "false" {
			query = query.Where("caractere_volontaire = ?", false)
		}
	}
	if search != "" {
		query = query.Where("type_motif ILIKE ? OR motif_principal ILIKE ? OR description ILIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	// Récupérer toutes les données
	err := query.Order("created_at DESC").Find(&motifs).Error
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch motifs for export",
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
	index, err := f.NewSheet("Motifs de Déplacement")
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
		NumFmt: 1, // Format numérique sans décimales
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

	// ===== EN-TÊTE PRINCIPAL =====
	currentTime := time.Now().Format("02/01/2006 15:04")
	mainHeader := fmt.Sprintf("RAPPORT D'EXPORT DES MOTIFS DE DÉPLACEMENT - %s", currentTime)
	f.SetCellValue("Motifs de Déplacement", "A1", mainHeader)
	f.MergeCell("Motifs de Déplacement", "A1", "R1")
	f.SetCellStyle("Motifs de Déplacement", "A1", "R1", headerStyle)
	f.SetRowHeight("Motifs de Déplacement", 1, 30)

	// ===== INFORMATIONS DE FILTRE =====
	row := 3
	filterApplied := false
	if migrantUUID != "" || typeMotif != "" || urgence != "" || caractereVolontaire != "" || search != "" {
		f.SetCellValue("Motifs de Déplacement", "A2", "Filtres appliqués:")
		f.SetCellStyle("Motifs de Déplacement", "A2", "A2", columnHeaderStyle)
		filterApplied = true

		if migrantUUID != "" {
			f.SetCellValue("Motifs de Déplacement", fmt.Sprintf("A%d", row), fmt.Sprintf("Migrant UUID: %s", migrantUUID))
			row++
		}
		if typeMotif != "" {
			f.SetCellValue("Motifs de Déplacement", fmt.Sprintf("A%d", row), fmt.Sprintf("Type de motif: %s", typeMotif))
			row++
		}
		if urgence != "" {
			f.SetCellValue("Motifs de Déplacement", fmt.Sprintf("A%d", row), fmt.Sprintf("Niveau d'urgence: %s", urgence))
			row++
		}
		if caractereVolontaire != "" {
			f.SetCellValue("Motifs de Déplacement", fmt.Sprintf("A%d", row), fmt.Sprintf("Caractère volontaire: %s", caractereVolontaire))
			row++
		}
		if search != "" {
			f.SetCellValue("Motifs de Déplacement", fmt.Sprintf("A%d", row), fmt.Sprintf("Recherche: %s", search))
			row++
		}
		row++ // Ligne vide
	}

	if !filterApplied {
		row = 2 // Pas de filtres, commencer plus haut
	}

	// ===== EN-TÊTES DE COLONNES =====
	headers := []string{
		"UUID",
		"Migrant UUID",
		"Nom du migrant",
		"Prénom du migrant",
		"Type de motif",
		"Motif principal",
		"Motif secondaire",
		"Description",
		"Caractère volontaire",
		"Niveau d'urgence",
		"Date déclenchement",
		"Durée estimée (jours)",
		"Conflit armé",
		"Catastrophe naturelle",
		"Persécution",
		"Violence généralisée",
		"Date de création",
		"Date de MAJ",
	}

	// Écrire les en-têtes
	for i, header := range headers {
		cell := fmt.Sprintf("%c%d", 'A'+i, row)
		f.SetCellValue("Motifs de Déplacement", cell, header)
		f.SetCellStyle("Motifs de Déplacement", cell, cell, columnHeaderStyle)
	}
	f.SetRowHeight("Motifs de Déplacement", row, 25)

	// ===== DONNÉES =====
	for i, motif := range motifs {
		dataRow := row + 1 + i

		// UUID
		cell := fmt.Sprintf("A%d", dataRow)
		f.SetCellValue("Motifs de Déplacement", cell, motif.UUID)
		f.SetCellStyle("Motifs de Déplacement", cell, cell, dataStyle)

		// Migrant UUID
		cell = fmt.Sprintf("B%d", dataRow)
		f.SetCellValue("Motifs de Déplacement", cell, motif.MigrantUUID)
		f.SetCellStyle("Motifs de Déplacement", cell, cell, dataStyle)

		// Nom du migrant
		cell = fmt.Sprintf("C%d", dataRow)
		if motif.Migrant.Nom != "" {
			f.SetCellValue("Motifs de Déplacement", cell, motif.Migrant.Nom)
		} else {
			f.SetCellValue("Motifs de Déplacement", cell, "N/A")
		}
		f.SetCellStyle("Motifs de Déplacement", cell, cell, dataStyle)

		// Prénom du migrant
		cell = fmt.Sprintf("D%d", dataRow)
		if motif.Migrant.Prenom != "" {
			f.SetCellValue("Motifs de Déplacement", cell, motif.Migrant.Prenom)
		} else {
			f.SetCellValue("Motifs de Déplacement", cell, "N/A")
		}
		f.SetCellStyle("Motifs de Déplacement", cell, cell, dataStyle)

		// Type de motif
		cell = fmt.Sprintf("E%d", dataRow)
		f.SetCellValue("Motifs de Déplacement", cell, motif.TypeMotif)
		f.SetCellStyle("Motifs de Déplacement", cell, cell, dataStyle)

		// Motif principal
		cell = fmt.Sprintf("F%d", dataRow)
		f.SetCellValue("Motifs de Déplacement", cell, motif.MotifPrincipal)
		f.SetCellStyle("Motifs de Déplacement", cell, cell, dataStyle)

		// Motif secondaire
		cell = fmt.Sprintf("G%d", dataRow)
		if motif.MotifSecondaire != "" {
			f.SetCellValue("Motifs de Déplacement", cell, motif.MotifSecondaire)
		} else {
			f.SetCellValue("Motifs de Déplacement", cell, "")
		}
		f.SetCellStyle("Motifs de Déplacement", cell, cell, dataStyle)

		// Description
		cell = fmt.Sprintf("H%d", dataRow)
		if motif.Description != "" {
			f.SetCellValue("Motifs de Déplacement", cell, motif.Description)
		} else {
			f.SetCellValue("Motifs de Déplacement", cell, "")
		}
		f.SetCellStyle("Motifs de Déplacement", cell, cell, dataStyle)

		// Caractère volontaire
		cell = fmt.Sprintf("I%d", dataRow)
		if motif.CaractereVolontaire {
			f.SetCellValue("Motifs de Déplacement", cell, "OUI")
		} else {
			f.SetCellValue("Motifs de Déplacement", cell, "NON")
		}
		f.SetCellStyle("Motifs de Déplacement", cell, cell, booleanStyle)

		// Niveau d'urgence
		cell = fmt.Sprintf("J%d", dataRow)
		f.SetCellValue("Motifs de Déplacement", cell, motif.Urgence)
		f.SetCellStyle("Motifs de Déplacement", cell, cell, dataStyle)

		// Date déclenchement
		cell = fmt.Sprintf("K%d", dataRow)
		f.SetCellValue("Motifs de Déplacement", cell, motif.DateDeclenchement.Format("02/01/2006"))
		f.SetCellStyle("Motifs de Déplacement", cell, cell, dateStyle)

		// Durée estimée
		cell = fmt.Sprintf("L%d", dataRow)
		if motif.DureeEstimee > 0 {
			f.SetCellValue("Motifs de Déplacement", cell, motif.DureeEstimee)
		} else {
			f.SetCellValue("Motifs de Déplacement", cell, "")
		}
		f.SetCellStyle("Motifs de Déplacement", cell, cell, numberStyle)

		// Conflit armé
		cell = fmt.Sprintf("M%d", dataRow)
		if motif.ConflitArme {
			f.SetCellValue("Motifs de Déplacement", cell, "OUI")
		} else {
			f.SetCellValue("Motifs de Déplacement", cell, "NON")
		}
		f.SetCellStyle("Motifs de Déplacement", cell, cell, booleanStyle)

		// Catastrophe naturelle
		cell = fmt.Sprintf("N%d", dataRow)
		if motif.CatastropheNaturelle {
			f.SetCellValue("Motifs de Déplacement", cell, "OUI")
		} else {
			f.SetCellValue("Motifs de Déplacement", cell, "NON")
		}
		f.SetCellStyle("Motifs de Déplacement", cell, cell, booleanStyle)

		// Persécution
		cell = fmt.Sprintf("O%d", dataRow)
		if motif.Persecution {
			f.SetCellValue("Motifs de Déplacement", cell, "OUI")
		} else {
			f.SetCellValue("Motifs de Déplacement", cell, "NON")
		}
		f.SetCellStyle("Motifs de Déplacement", cell, cell, booleanStyle)

		// Violence généralisée
		cell = fmt.Sprintf("P%d", dataRow)
		if motif.ViolenceGeneralisee {
			f.SetCellValue("Motifs de Déplacement", cell, "OUI")
		} else {
			f.SetCellValue("Motifs de Déplacement", cell, "NON")
		}
		f.SetCellStyle("Motifs de Déplacement", cell, cell, booleanStyle)

		// Date de création
		cell = fmt.Sprintf("Q%d", dataRow)
		f.SetCellValue("Motifs de Déplacement", cell, motif.CreatedAt.Format("02/01/2006 15:04"))
		f.SetCellStyle("Motifs de Déplacement", cell, cell, dateStyle)

		// Date de MAJ
		cell = fmt.Sprintf("R%d", dataRow)
		f.SetCellValue("Motifs de Déplacement", cell, motif.UpdatedAt.Format("02/01/2006 15:04"))
		f.SetCellStyle("Motifs de Déplacement", cell, cell, dateStyle)

		// Définir la hauteur de ligne
		f.SetRowHeight("Motifs de Déplacement", dataRow, 20)
	}

	// ===== AJUSTEMENT DE LA LARGEUR DES COLONNES =====
	columnWidths := []float64{
		25, // UUID
		25, // Migrant UUID
		15, // Nom
		15, // Prénom
		15, // Type motif
		25, // Motif principal
		25, // Motif secondaire
		40, // Description
		12, // Caractère volontaire
		12, // Urgence
		15, // Date déclenchement
		12, // Durée estimée
		12, // Conflit armé
		18, // Catastrophe naturelle
		12, // Persécution
		18, // Violence généralisée
		18, // Date création
		18, // Date MAJ
	}

	columns := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R"}
	for i, width := range columnWidths {
		if i < len(columns) {
			f.SetColWidth("Motifs de Déplacement", columns[i], columns[i], width)
		}
	}

	// ===== AJOUTER UNE FEUILLE DE STATISTIQUES =====
	_, err = f.NewSheet("Statistiques")
	if err == nil {
		// Calculer les statistiques
		totalRecords := len(motifs)

		// Compter par type de motif
		typeCount := make(map[string]int)
		urgenceCount := make(map[string]int)
		volontaireCount := 0
		involontaireCount := 0
		conflitCount := 0
		catastropheCount := 0
		persecutionCount := 0
		violenceCount := 0

		for _, motif := range motifs {
			typeCount[motif.TypeMotif]++
			urgenceCount[motif.Urgence]++
			if motif.CaractereVolontaire {
				volontaireCount++
			} else {
				involontaireCount++
			}
			if motif.ConflitArme {
				conflitCount++
			}
			if motif.CatastropheNaturelle {
				catastropheCount++
			}
			if motif.Persecution {
				persecutionCount++
			}
			if motif.ViolenceGeneralisee {
				violenceCount++
			}
		}

		// En-tête de la feuille statistiques
		f.SetCellValue("Statistiques", "A1", "STATISTIQUES DES MOTIFS DE DÉPLACEMENT")
		f.MergeCell("Statistiques", "A1", "C1")
		f.SetCellStyle("Statistiques", "A1", "C1", headerStyle)

		row = 3
		f.SetCellValue("Statistiques", fmt.Sprintf("A%d", row), "Total des enregistrements:")
		f.SetCellValue("Statistiques", fmt.Sprintf("B%d", row), totalRecords)
		row += 2

		// Par caractère volontaire
		f.SetCellValue("Statistiques", fmt.Sprintf("A%d", row), "Déplacements volontaires:")
		f.SetCellValue("Statistiques", fmt.Sprintf("B%d", row), volontaireCount)
		row++
		f.SetCellValue("Statistiques", fmt.Sprintf("A%d", row), "Déplacements involontaires:")
		f.SetCellValue("Statistiques", fmt.Sprintf("B%d", row), involontaireCount)
		row += 2

		// Par type de motif
		f.SetCellValue("Statistiques", fmt.Sprintf("A%d", row), "Par type de motif:")
		f.SetCellStyle("Statistiques", fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), columnHeaderStyle)
		row++
		for typeMotif, count := range typeCount {
			f.SetCellValue("Statistiques", fmt.Sprintf("A%d", row), typeMotif)
			f.SetCellValue("Statistiques", fmt.Sprintf("B%d", row), count)
			row++
		}
		row++

		// Par niveau d'urgence
		f.SetCellValue("Statistiques", fmt.Sprintf("A%d", row), "Par niveau d'urgence:")
		f.SetCellStyle("Statistiques", fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), columnHeaderStyle)
		row++
		for urgence, count := range urgenceCount {
			f.SetCellValue("Statistiques", fmt.Sprintf("A%d", row), urgence)
			f.SetCellValue("Statistiques", fmt.Sprintf("B%d", row), count)
			row++
		}
		row++

		// Facteurs externes
		f.SetCellValue("Statistiques", fmt.Sprintf("A%d", row), "Facteurs externes:")
		f.SetCellStyle("Statistiques", fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), columnHeaderStyle)
		row++
		f.SetCellValue("Statistiques", fmt.Sprintf("A%d", row), "Conflit armé")
		f.SetCellValue("Statistiques", fmt.Sprintf("B%d", row), conflitCount)
		row++
		f.SetCellValue("Statistiques", fmt.Sprintf("A%d", row), "Catastrophe naturelle")
		f.SetCellValue("Statistiques", fmt.Sprintf("B%d", row), catastropheCount)
		row++
		f.SetCellValue("Statistiques", fmt.Sprintf("A%d", row), "Persécution")
		f.SetCellValue("Statistiques", fmt.Sprintf("B%d", row), persecutionCount)
		row++
		f.SetCellValue("Statistiques", fmt.Sprintf("A%d", row), "Violence généralisée")
		f.SetCellValue("Statistiques", fmt.Sprintf("B%d", row), violenceCount)

		f.SetColWidth("Statistiques", "A", "A", 25)
		f.SetColWidth("Statistiques", "B", "B", 15)
	}

	// ===== GÉNÉRATION DU FICHIER =====
	filename := fmt.Sprintf("motifs_deplacement_export_%s.xlsx", time.Now().Format("20060102_150405"))

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
