package identites

import (
	"fmt"
	"strconv"

	"github.com/kgermando/sysmobembo-api/database"
	"github.com/kgermando/sysmobembo-api/models"
	"github.com/kgermando/sysmobembo-api/utils"

	"github.com/gofiber/fiber/v2"
	"github.com/xuri/excelize/v2"
)

// GetPaginatedIdentites - Récupérer toutes les identités avec pagination et recherche
func GetPaginatedIdentites(c *fiber.Ctx) error {
	db := database.DB

	// Parse query parameters for pagination
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page <= 0 {
		page = 1
	}
	limit, err := strconv.Atoi(c.Query("limit", "15"))
	if err != nil || limit <= 0 {
		limit = 15
	}
	offset := (page - 1) * limit

	// Parse search parameter
	search := c.Query("search", "")

	var identites []models.Identite
	var totalRecords int64

	// Build query with filters
	query := db.Model(&models.Identite{})

	// Search filter
	if search != "" {
		query = query.Where("nom ILIKE ? OR postnom ILIKE ? OR prenom ILIKE ? OR numero_passeport ILIKE ? OR nationalite ILIKE ? OR lieu_naissance ILIKE ? OR sexe ILIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	// Count total records with filters applied
	query.Count(&totalRecords)

	// Execute query with pagination
	err = query.
		Offset(offset).
		Limit(limit).
		Order("updated_at DESC").
		Find(&identites).Error

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch identites",
			"error":   err.Error(),
		})
	}

	// Calculate total pages
	totalPages := int((totalRecords + int64(limit) - 1) / int64(limit))

	// Prepare pagination metadata
	pagination := map[string]interface{}{
		"total_records": totalRecords,
		"total_pages":   totalPages,
		"current_page":  page,
		"page_size":     limit,
	}

	// Return response
	return c.JSON(fiber.Map{
		"status":     "success",
		"message":    "Identites retrieved successfully",
		"data":       identites,
		"pagination": pagination,
	})
}

// GetMigrantsByIdentiteUUID - Récupérer tous les migrants selon un identite_uuid avec pagination et recherche
func GetMigrantsByIdentiteUUID(c *fiber.Ctx) error {
	db := database.DB

	// Get identite_uuid from query parameter
	identiteUUID := c.Query("identite_uuid", "")
	if identiteUUID == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "identite_uuid query parameter is required",
		})
	}

	// Parse query parameters for pagination
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page <= 0 {
		page = 1
	}
	limit, err := strconv.Atoi(c.Query("limit", "15"))
	if err != nil || limit <= 0 {
		limit = 15
	}
	offset := (page - 1) * limit

	// Parse search parameter
	search := c.Query("search", "")

	var migrants []models.Migrant
	var totalRecords int64

	// Build query with filters
	query := db.Model(&models.Migrant{}).Where("migrants.identite_uuid = ?", identiteUUID)

	// Search filter
	if search != "" {
		query = query.Joins("LEFT JOIN identites ON migrants.identite_uuid = identites.uuid").
			Where("migrants.identite_uuid = ? AND (identites.nom ILIKE ? OR identites.postnom ILIKE ? OR identites.prenom ILIKE ? OR migrants.numero_identifiant ILIKE ? OR identites.nationalite ILIKE ? OR identites.numero_passeport ILIKE ? OR migrants.adresse_actuelle ILIKE ? OR migrants.ville_actuelle ILIKE ? OR migrants.pays_actuel ILIKE ? OR migrants.situation_matrimoniale ILIKE ?)",
				identiteUUID, "%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	// Count total records with filters applied
	query.Count(&totalRecords)

	// Execute query with pagination
	err = query.
		Preload("Identite").
		Preload("MotifDeplacements").
		Preload("Alertes").
		Preload("Biometries").
		Preload("Geolocalisations").
		Offset(offset).
		Limit(limit).
		Order("migrants.updated_at DESC").
		Find(&migrants).Error

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch migrants",
			"error":   err.Error(),
		})
	}

	// Calculate total pages
	totalPages := int((totalRecords + int64(limit) - 1) / int64(limit))

	// Prepare pagination metadata
	pagination := map[string]interface{}{
		"total_records": totalRecords,
		"total_pages":   totalPages,
		"current_page":  page,
		"page_size":     limit,
	}

	// Return response
	return c.JSON(fiber.Map{
		"status":     "success",
		"message":    "Migrants retrieved successfully for identite_uuid: " + identiteUUID,
		"data":       migrants,
		"pagination": pagination,
	})
}

// GetIdentite récupère une identité par UUID
func GetIdentite(c *fiber.Ctx) error {
	uuid := c.Params("uuid")
	db := database.DB

	var identite models.Identite
	err := db.Where("uuid = ?", uuid).First(&identite).Error
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Identite not found",
		})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   identite,
	})
}

// CreateIdentite crée une nouvelle identité
func CreateIdentite(c *fiber.Ctx) error {
	db := database.DB
	identite := new(models.Identite)

	if err := c.BodyParser(identite); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid input",
			"error":   err.Error(),
		})
	}

	// Générer l'UUID
	identite.UUID = utils.GenerateUUID()

	// Validation des données
	if errors := utils.ValidateStruct(*identite); len(errors) > 0 {
		var errorMessages []string
		for _, err := range errors {
			switch err.FailedField {
			case "Identite.Nom":
				errorMessages = append(errorMessages, "Le nom est requis")
			case "Identite.Postnom":
				errorMessages = append(errorMessages, "Le postnom est requis")
			case "Identite.Prenom":
				errorMessages = append(errorMessages, "Le prénom est requis")
			case "Identite.DateNaissance":
				errorMessages = append(errorMessages, "La date de naissance est requise")
			case "Identite.LieuNaissance":
				errorMessages = append(errorMessages, "Le lieu de naissance est requis")
			case "Identite.Sexe":
				if err.Tag == "oneof" {
					errorMessages = append(errorMessages, "Le sexe doit être 'M' ou 'F'")
				} else {
					errorMessages = append(errorMessages, "Le sexe est requis")
				}
			case "Identite.Nationalite":
				errorMessages = append(errorMessages, "La nationalité est requise")
			default:
				errorMessages = append(errorMessages, fmt.Sprintf("Erreur de validation pour %s: %s", err.FailedField, err.Tag))
			}
		}

		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Erreurs de validation",
			"errors":  errorMessages,
		})
	}

	// Vérifier l'unicité du numéro de passeport
	if identite.NumeroPasseport != "" {
		var existingIdentite models.Identite
		if err := db.Where("numero_passeport = ?", identite.NumeroPasseport).First(&existingIdentite).Error; err == nil {
			return c.Status(400).JSON(fiber.Map{
				"status":  "error",
				"message": "Une identité avec ce numéro de passeport existe déjà",
			})
		}
	}

	// Créer l'identité
	if err := db.Create(&identite).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Cannot create identite",
			"error":   err.Error(),
		})
	}

	return c.Status(201).JSON(fiber.Map{
		"status":  "success",
		"message": "Identite created successfully",
		"data":    identite,
	})
}

// UpdateIdentite met à jour une identité
func UpdateIdentite(c *fiber.Ctx) error {
	uuid := c.Params("uuid")
	db := database.DB

	var identite models.Identite
	err := db.Where("uuid = ?", uuid).First(&identite).Error
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Identite not found",
		})
	}

	var updateData models.Identite
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid input",
			"error":   err.Error(),
		})
	}

	// Conserver l'UUID
	updateData.UUID = identite.UUID

	// Vérifier l'unicité du numéro de passeport
	if updateData.NumeroPasseport != "" && updateData.NumeroPasseport != identite.NumeroPasseport {
		var existingIdentite models.Identite
		if err := db.Where("numero_passeport = ? AND uuid != ?", updateData.NumeroPasseport, uuid).First(&existingIdentite).Error; err == nil {
			return c.Status(400).JSON(fiber.Map{
				"status":  "error",
				"message": "Une identité avec ce numéro de passeport existe déjà",
			})
		}
	}

	// Mettre à jour
	if err := db.Model(&identite).Updates(&updateData).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Cannot update identite",
			"error":   err.Error(),
		})
	}

	// Récupérer l'identité mise à jour
	db.Where("uuid = ?", uuid).First(&identite)

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Identite updated successfully",
		"data":    identite,
	})
}

// DeleteIdentite supprime une identité (soft delete)
func DeleteIdentite(c *fiber.Ctx) error {
	uuid := c.Params("uuid")
	db := database.DB

	var identite models.Identite
	err := db.Where("uuid = ?", uuid).First(&identite).Error
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Identite not found",
		})
	}

	// Vérifier si l'identité est utilisée par un migrant
	var migrantCount int64
	db.Model(&models.Migrant{}).Where("identite_uuid = ?", uuid).Count(&migrantCount)
	if migrantCount > 0 {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Impossible de supprimer cette identité car elle est associée à un ou plusieurs migrants",
		})
	}

	// Supprimer (soft delete)
	if err := db.Delete(&identite).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Cannot delete identite",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Identite deleted successfully",
	})
}


// ExportIdentitesToExcel exporte les identités vers Excel
func ExportIdentitesToExcel(c *fiber.Ctx) error {
	db := database.DB

	// Récupérer les paramètres de plage de dates
	startDate := c.Query("start_date", "")
	endDate := c.Query("end_date", "")

	var identites []models.Identite

	query := db.Model(&models.Identite{})

	// Appliquer les filtres de dates
	if startDate != "" {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("created_at <= ?", endDate)
	}

	// Récupérer toutes les données
	err := query.Order("created_at DESC").Find(&identites).Error
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch identites for export",
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
	_, err = f.NewSheet("Identités")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to create Excel sheet",
			"error":   err.Error(),
		})
	}

	// Styles
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   12,
			Family: "Calibri",
			Color:  "FFFFFF",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"4472C4"},
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

	dataStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   11,
			Family: "Calibri",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "left",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "CCCCCC", Style: 1},
			{Type: "top", Color: "CCCCCC", Style: 1},
			{Type: "bottom", Color: "CCCCCC", Style: 1},
			{Type: "right", Color: "CCCCCC", Style: 1},
		},
	})

	// En-têtes de colonnes
	headers := []string{
		"UUID",
		"Nom",
		"Postnom",
		"Prénom",
		"Date de naissance",
		"Lieu de naissance",
		"Sexe",
		"Nationalité",
		"Adresse",
		"Profession",
		"N° Passeport",
		"Pays émetteur",
		"Autorité émetteur",
		"Date création",
	}

	row := 1
	for i, header := range headers {
		cell := fmt.Sprintf("%c%d", 'A'+i, row)
		f.SetCellValue("Identités", cell, header)
		f.SetCellStyle("Identités", cell, cell, headerStyle)
	}
	f.SetRowHeight("Identités", row, 25)

	// Données
	for i, identite := range identites {
		dataRow := row + 1 + i

		f.SetCellValue("Identités", fmt.Sprintf("A%d", dataRow), identite.UUID)
		f.SetCellValue("Identités", fmt.Sprintf("B%d", dataRow), identite.Nom)
		f.SetCellValue("Identités", fmt.Sprintf("C%d", dataRow), identite.Postnom)
		f.SetCellValue("Identités", fmt.Sprintf("D%d", dataRow), identite.Prenom)
		f.SetCellValue("Identités", fmt.Sprintf("E%d", dataRow), identite.DateNaissance.Format("02/01/2006"))
		f.SetCellValue("Identités", fmt.Sprintf("F%d", dataRow), identite.LieuNaissance)
		f.SetCellValue("Identités", fmt.Sprintf("G%d", dataRow), identite.Sexe)
		f.SetCellValue("Identités", fmt.Sprintf("H%d", dataRow), identite.Nationalite)
		f.SetCellValue("Identités", fmt.Sprintf("I%d", dataRow), identite.Adresse)
		f.SetCellValue("Identités", fmt.Sprintf("J%d", dataRow), identite.Profession)
		f.SetCellValue("Identités", fmt.Sprintf("K%d", dataRow), identite.NumeroPasseport)
		f.SetCellValue("Identités", fmt.Sprintf("L%d", dataRow), identite.PaysEmetteur)
		f.SetCellValue("Identités", fmt.Sprintf("M%d", dataRow), identite.AutoriteEmetteur)
		f.SetCellValue("Identités", fmt.Sprintf("N%d", dataRow), identite.CreatedAt.Format("02/01/2006 15:04"))

		// Appliquer le style aux données
		for col := 'A'; col <= 'N'; col++ {
			cell := fmt.Sprintf("%c%d", col, dataRow)
			f.SetCellStyle("Identités", cell, cell, dataStyle)
		}
	}

	// Largeur des colonnes
	columnWidths := map[string]float64{
		"A": 38, // UUID
		"B": 15, // Nom
		"C": 15, // Postnom
		"D": 15, // Prénom
		"E": 15, // Date naissance
		"F": 20, // Lieu naissance
		"G": 8,  // Sexe
		"H": 20, // Nationalité
		"I": 30, // Adresse
		"J": 20, // Profession
		"K": 15, // N° Passeport
		"L": 20, // Pays émetteur
		"M": 25, // Autorité émetteur
		"N": 18, // Date création
	}

	for col, width := range columnWidths {
		f.SetColWidth("Identités", col, col, width)
	}

	// Activer la feuille
	f.SetActiveSheet(0)

	// Envoyer le fichier
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", "attachment; filename=identites.xlsx")

	return f.Write(c.Response().BodyWriter())
}

// GetIdentiteStatistics retourne des statistiques sur les identités
func GetIdentiteStatistics(c *fiber.Ctx) error {
	db := database.DB

	var total int64
	db.Model(&models.Identite{}).Count(&total)

	// Statistiques par sexe
	var sexeStats []struct {
		Sexe  string
		Count int64
	}
	db.Model(&models.Identite{}).
		Select("sexe, COUNT(*) as count").
		Group("sexe").
		Scan(&sexeStats)

	// Statistiques par nationalité (top 10)
	var nationaliteStats []struct {
		Nationalite string
		Count       int64
	}
	db.Model(&models.Identite{}).
		Select("nationalite, COUNT(*) as count").
		Group("nationalite").
		Order("count DESC").
		Limit(10).
		Scan(&nationaliteStats)

	// Identités avec/sans passeport
	var withPassport int64
	var withoutPassport int64
	db.Model(&models.Identite{}).Where("numero_passeport != ''").Count(&withPassport)
	db.Model(&models.Identite{}).Where("numero_passeport = '' OR numero_passeport IS NULL").Count(&withoutPassport)

	return c.JSON(fiber.Map{
		"status": "success",
		"data": fiber.Map{
			"total":           total,
			"par_sexe":        sexeStats,
			"par_nationalite": nationaliteStats,
			"avec_passeport":  withPassport,
			"sans_passeport":  withoutPassport,
		},
	})
}
