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

	var motifs []models.MotifDeplacement
	var totalRecords int64

	query := db.Model(&models.MotifDeplacement{}).
		Preload("Migrant")

	// Recherche textuelle
	if search != "" {
		query = query.Joins("LEFT JOIN migrants ON migrants.uuid = motif_deplacements.migrant_uuid").
			Where("type_motif ILIKE ? OR motif_principal ILIKE ? OR description ILIKE ? OR migrants.numero_identifiant ILIKE ?",
				"%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	// Count total
	query.Count(&totalRecords)

	// Get paginated results
	err = query.
		Preload("Migrant").
		Offset(offset).
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
	if motif.MigrantUUID == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "MigrantUUID is required",
			"data":    nil,
		})
	}

	// Vérifier que le migrant existe
	// var migrant models.Migrant
	// if err := database.DB.Where("uuid = ?", motif.MigrantUUID).First(&migrant).Error; err != nil {
	// 	return c.Status(404).JSON(fiber.Map{
	// 		"status":  "error",
	// 		"message": "Migrant not found",
	// 		"data":    nil,
	// 	})
	// }

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
 

	stats := map[string]interface{}{
		"total_motifs":         totalMotifs,
		"motifs_volontaires":   motifsVolontaires,
		"motifs_involontaires": motifsInvolontaires,
		"types_motifs":         motifTypes,
		"urgence_stats":        urgenceStats, 
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
	startDate := c.Query("start_date", "")
	endDate := c.Query("end_date", "")

	var motifs []models.MotifDeplacement

	query := db.Model(&models.MotifDeplacement{}).Preload("Migrant")

	// Appliquer les filtres de date
	if startDate != "" {
		parsedStartDate, err := time.Parse("2006-01-02", startDate)
		if err == nil {
			query = query.Where("created_at >= ?", parsedStartDate)
		}
	}
	if endDate != "" {
		parsedEndDate, err := time.Parse("2006-01-02", endDate)
		if err == nil {
			query = query.Where("created_at <= ?", parsedEndDate)
		}
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
	mainHeader := fmt.Sprintf("RAPPORT D'EXPORT DES MOTIFS DE DÉPLACEMENT - %s", currentTime)
	f.SetCellValue("Motifs de Déplacement", "A1", mainHeader)
	f.MergeCell("Motifs de Déplacement", "A1", "G1")
	f.SetCellStyle("Motifs de Déplacement", "A1", "G1", headerStyle)
	f.SetRowHeight("Motifs de Déplacement", 1, 30)

	// ===== INFORMATIONS DE FILTRE =====
	row := 3
	filterApplied := false
	if startDate != "" || endDate != "" {
		f.SetCellValue("Motifs de Déplacement", "A2", "Filtres appliqués:")
		f.SetCellStyle("Motifs de Déplacement", "A2", "A2", columnHeaderStyle)
		filterApplied = true

		if startDate != "" {
			f.SetCellValue("Motifs de Déplacement", fmt.Sprintf("A%d", row), fmt.Sprintf("Date de début: %s", startDate))
			row++
		}
		if endDate != "" {
			f.SetCellValue("Motifs de Déplacement", fmt.Sprintf("A%d", row), fmt.Sprintf("Date de fin: %s", endDate))
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
		"Numéro Identifiant Migrant",
		"Type de motif",
		"Motif principal",
		"Description",
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

		// Numéro Identifiant Migrant
		cell = fmt.Sprintf("B%d", dataRow)
		if motif.Migrant.UUID != "" {
			f.SetCellValue("Motifs de Déplacement", cell, motif.Migrant.NumeroIdentifiant)
		} else {
			f.SetCellValue("Motifs de Déplacement", cell, "N/A")
		}
		f.SetCellStyle("Motifs de Déplacement", cell, cell, dataStyle)

		// Type de motif
		cell = fmt.Sprintf("C%d", dataRow)
		f.SetCellValue("Motifs de Déplacement", cell, motif.TypeMotif)
		f.SetCellStyle("Motifs de Déplacement", cell, cell, dataStyle)

		// Motif principal
		cell = fmt.Sprintf("D%d", dataRow)
		f.SetCellValue("Motifs de Déplacement", cell, motif.MotifPrincipal)
		f.SetCellStyle("Motifs de Déplacement", cell, cell, dataStyle)

		// Description
		cell = fmt.Sprintf("E%d", dataRow)
		if motif.Description != "" {
			f.SetCellValue("Motifs de Déplacement", cell, motif.Description)
		} else {
			f.SetCellValue("Motifs de Déplacement", cell, "")
		}
		f.SetCellStyle("Motifs de Déplacement", cell, cell, dataStyle)

		// Date de création
		cell = fmt.Sprintf("F%d", dataRow)
		f.SetCellValue("Motifs de Déplacement", cell, motif.CreatedAt.Format("02/01/2006 15:04"))
		f.SetCellStyle("Motifs de Déplacement", cell, cell, dateStyle)

		// Date de MAJ
		cell = fmt.Sprintf("G%d", dataRow)
		f.SetCellValue("Motifs de Déplacement", cell, motif.UpdatedAt.Format("02/01/2006 15:04"))
		f.SetCellStyle("Motifs de Déplacement", cell, cell, dateStyle)

		// Définir la hauteur de ligne
		f.SetRowHeight("Motifs de Déplacement", dataRow, 20)
	}

	// ===== AJUSTEMENT DE LA LARGEUR DES COLONNES =====
	columnWidths := []float64{
		25, // UUID
		25, // Numéro Identifiant
		20, // Type motif
		25, // Motif principal
		40, // Description
		18, // Date création
		18, // Date MAJ
	}

	columns := []string{"A", "B", "C", "D", "E", "F", "G"}
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

		for _, motif := range motifs {
			typeCount[motif.TypeMotif]++
		}

		// En-tête de la feuille statistiques
		f.SetCellValue("Statistiques", "A1", "STATISTIQUES DES MOTIFS DE DÉPLACEMENT")
		f.MergeCell("Statistiques", "A1", "C1")
		f.SetCellStyle("Statistiques", "A1", "C1", headerStyle)

		row = 3
		f.SetCellValue("Statistiques", fmt.Sprintf("A%d", row), "Total des enregistrements:")
		f.SetCellValue("Statistiques", fmt.Sprintf("B%d", row), totalRecords)
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
