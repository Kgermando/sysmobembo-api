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

// Fonction pour générer automatiquement le NumeroIdentifiant
func generateNumeroIdentifiant() string {
	year := time.Now().Year()

	// Compter le nombre de migrants créés cette année
	var count int64
	database.DB.Model(&models.Migrant{}).
		Where("EXTRACT(YEAR FROM created_at) = ?", year).
		Count(&count)

	// Incrémenter pour le nouveau migrant
	sequence := count + 1

	return fmt.Sprintf("MIG-%d-%06d", year, sequence)
}

// Paginate - Récupérer les migrants avec pagination et filtres
func GetPaginatedMigrants(c *fiber.Ctx) error {
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

	// Parse search and filter parameters
	search := c.Query("search", "")

	var migrants []models.Migrant
	var totalRecords int64

	// Build query with filters
	query := db.Model(&models.Migrant{})

	// Search filter
	if search != "" {
		query = query.Joins("LEFT JOIN identites ON migrants.identite_uuid = identites.uuid").
			Where("identites.nom ILIKE ? OR identites.postnom ILIKE ? OR identites.prenom ILIKE ? OR migrants.numero_identifiant ILIKE ? OR identites.nationalite ILIKE ? OR identites.numero_passeport ILIKE ? OR migrants.adresse_actuelle ILIKE ? OR migrants.ville_actuelle ILIKE ? OR migrants.pays_actuel ILIKE ? OR migrants.situation_matrimoniale ILIKE ?",
				"%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	// Count total records with filters applied
	query.Count(&totalRecords)

	// Execute query with pagination
	err = query.
		Preload("Identite").
		Preload("MotifDeplacements").
		Preload("Alertes").
		Preload("Biometries").
		Offset(offset).
		Limit(limit).
		Order("migrants.updated_at DESC").
		Find(&migrants).Error

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch Migrants",
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
		"message":    "Migrants retrieved successfully",
		"data":       migrants,
		"pagination": pagination,
	})
}

// Query all data
func GetAllMigrants(c *fiber.Ctx) error {
	db := database.DB
	var migrants []models.Migrant

	err := db.
		Preload("Identite").
		Find(&migrants).Error

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch migrants",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "All migrants",
		"data":    migrants,
	})
}

// Get one data
func GetMigrant(c *fiber.Ctx) error {
	uuid := c.Params("uuid")
	db := database.DB
	var migrant models.Migrant

	err := db.Where("uuid = ?", uuid).
		Preload("Identite").
		Preload("MotifDeplacements").
		Preload("Alertes").
		Preload("Biometries").
		Preload("Geolocalisations").
		First(&migrant).Error

	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Migrant not found",
			"data":    nil,
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Migrant found",
		"data":    migrant,
	})
}

// Create data
func CreateMigrant(c *fiber.Ctx) error {
	migrant := &models.Migrant{}

	if err := c.BodyParser(migrant); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request format",
			"error":   err.Error(),
		})
	}

	// Générer automatiquement l'UUID et le NumeroIdentifiant
	migrant.UUID = utils.GenerateUUID()
	migrant.NumeroIdentifiant = generateNumeroIdentifiant()

	if err := database.DB.Create(migrant).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to create migrant",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Migrant created successfully",
		"data":    migrant,
	})
}

// Update data
func UpdateMigrant(c *fiber.Ctx) error {
	uuid := c.Params("uuid")
	db := database.DB

	var updateData models.Migrant

	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Review your input",
			"error":   err.Error(),
		})
	}

	migrant := new(models.Migrant)

	if err := db.Where("uuid = ?", uuid).First(&migrant).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Migrant not found",
			"data":    nil,
		})
	}

	// Conserver l'UUID et le NumeroIdentifiant existants
	updateData.UUID = migrant.UUID
	updateData.NumeroIdentifiant = migrant.NumeroIdentifiant

	// migrant.Telephone = updateData.Telephone
	// migrant.Email = updateData.Email
	// migrant.AdresseActuelle = updateData.AdresseActuelle
	// migrant.VilleActuelle = updateData.VilleActuelle
	// migrant.PaysActuel = updateData.PaysActuel
	// migrant.SituationMatrimoniale = updateData.SituationMatrimoniale
	// migrant.NombreEnfants = updateData.NombreEnfants
	// migrant.PersonneContact = updateData.PersonneContact
	// migrant.TelephoneContact = updateData.TelephoneContact
	// migrant.StatutMigratoire = updateData.StatutMigratoire
	// migrant.DateEntree = updateData.DateEntree
	// migrant.PointEntree = updateData.PointEntree
	// migrant.PaysDestination = updateData.PaysDestination

	if err := db.Model(&migrant).Updates(updateData).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to update migrant",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Migrant updated successfully",
		"data":    migrant,
	})
}

// Delete data
func DeleteMigrant(c *fiber.Ctx) error {
	uuid := c.Params("uuid")
	db := database.DB

	var migrant models.Migrant
	if err := db.Where("uuid = ?", uuid).First(&migrant).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Migrant not found",
			"data":    nil,
		})
	}

	// Soft delete - les relations seront également supprimées grâce à OnDelete:CASCADE
	if err := db.Delete(&migrant).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to delete migrant",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Migrant deleted successfully",
		"data":    nil,
	})
}

// Get migrants statistics
func GetMigrantsStats(c *fiber.Ctx) error {
	db := database.DB

	var totalMigrants int64
	var regularMigrants int64
	var irregularMigrants int64
	var refugeeMigrants int64
	var asylumSeekers int64

	// Total migrants
	db.Model(&models.Migrant{}).Count(&totalMigrants)

	// Par statut migratoire
	db.Model(&models.Migrant{}).Where("statut_migratoire = ?", "regulier").Count(&regularMigrants)
	db.Model(&models.Migrant{}).Where("statut_migratoire = ?", "irregulier").Count(&irregularMigrants)
	db.Model(&models.Migrant{}).Where("statut_migratoire = ?", "refugie").Count(&refugeeMigrants)
	db.Model(&models.Migrant{}).Where("statut_migratoire = ?", "demandeur_asile").Count(&asylumSeekers)

	stats := map[string]interface{}{
		"total_migrants":     totalMigrants,
		"regular_migrants":   regularMigrants,
		"irregular_migrants": irregularMigrants,
		"refugee_migrants":   refugeeMigrants,
		"asylum_seekers":     asylumSeekers,
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Migrants statistics",
		"data":    stats,
	})
}

// =======================
// EXCEL EXPORT
// =======================

// ExportMigrantsToExcel - Exporter les migrants vers Excel avec mise en forme
func ExportMigrantsToExcel(c *fiber.Ctx) error {
	db := database.DB

	// Récupérer les paramètres de filtre pour les dates
	startDate := c.Query("start_date", "")
	endDate := c.Query("end_date", "")

	var migrants []models.Migrant

	query := db.Model(&models.Migrant{}).Preload("Identite")

	// Appliquer le filtre de plage de dates sur created_at
	if startDate != "" {
		parsedStartDate, err := time.Parse("2006-01-02", startDate)
		if err == nil {
			query = query.Where("migrants.created_at >= ?", parsedStartDate)
		}
	}
	if endDate != "" {
		parsedEndDate, err := time.Parse("2006-01-02", endDate)
		if err == nil {
			// Ajouter 23h59m59s pour inclure toute la journée de fin
			parsedEndDate = parsedEndDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			query = query.Where("migrants.created_at <= ?", parsedEndDate)
		}
	}

	// Récupérer toutes les données
	err := query.Order("created_at DESC").Find(&migrants).Error
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch migrants for export",
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
	index, err := f.NewSheet("Migrants")
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

	// ===== EN-TÊTE PRINCIPAL =====
	currentTime := time.Now().Format("02/01/2006 15:04")
	mainHeader := fmt.Sprintf("RAPPORT D'EXPORT DES MIGRANTS - %s", currentTime)
	f.SetCellValue("Migrants", "A1", mainHeader)
	f.MergeCell("Migrants", "A1", "AA1")
	f.SetCellStyle("Migrants", "A1", "AA1", headerStyle)
	f.SetRowHeight("Migrants", 1, 30)

	// ===== INFORMATIONS DE FILTRE =====
	row := 3
	filterApplied := false
	if startDate != "" || endDate != "" {
		f.SetCellValue("Migrants", "A2", "Filtres appliqués:")
		f.SetCellStyle("Migrants", "A2", "A2", columnHeaderStyle)
		filterApplied = true

		if startDate != "" {
			f.SetCellValue("Migrants", fmt.Sprintf("A%d", row), fmt.Sprintf("Date de début: %s", startDate))
			row++
		}
		if endDate != "" {
			f.SetCellValue("Migrants", fmt.Sprintf("A%d", row), fmt.Sprintf("Date de fin: %s", endDate))
			row++
		}
		row++ // Ligne vide
	}

	if !filterApplied {
		row = 2 // Pas de filtres, commencer plus haut
	}

	// ===== EN-TÊTES DE COLONNES =====
	headers := []string{
		"N° Identifiant",
		"Nom",
		"Prénom",
		"Date de naissance",
		"Lieu de naissance",
		"Sexe",
		"Nationalité",
		"Type de document",
		"N° Document",
		"Date émission doc",
		"Date expiration doc",
		"Autorité émission",
		"Téléphone",
		"Email",
		"Adresse actuelle",
		"Ville actuelle",
		"Pays actuel",
		"Situation matrimoniale",
		"Nombre enfants",
		"Personne contact",
		"Téléphone contact",
		"Statut migratoire",
		"Date d'entrée",
		"Point d'entrée",
		"Pays destination",
		"Date création",
		"Date MAJ",
	} // Écrire les en-têtes
	for i, header := range headers {
		cell := fmt.Sprintf("%c%d", 'A'+i, row)
		f.SetCellValue("Migrants", cell, header)
		f.SetCellStyle("Migrants", cell, cell, columnHeaderStyle)
	}
	f.SetRowHeight("Migrants", row, 25)

	// ===== DONNÉES =====
	for i, migrant := range migrants {
		dataRow := row + 1 + i

		// N° Identifiant
		cell := fmt.Sprintf("A%d", dataRow)
		f.SetCellValue("Migrants", cell, migrant.NumeroIdentifiant)
		f.SetCellStyle("Migrants", cell, cell, dataStyle)

		// Nom
		cell = fmt.Sprintf("B%d", dataRow)
		if migrant.Identite.UUID != "" {
			f.SetCellValue("Migrants", cell, migrant.Identite.Nom)
		} else {
			f.SetCellValue("Migrants", cell, "")
		}
		f.SetCellStyle("Migrants", cell, cell, dataStyle)

		// Prénom
		cell = fmt.Sprintf("C%d", dataRow)
		if migrant.Identite.UUID != "" {
			f.SetCellValue("Migrants", cell, migrant.Identite.Prenom)
		} else {
			f.SetCellValue("Migrants", cell, "")
		}
		f.SetCellStyle("Migrants", cell, cell, dataStyle)

		// Date de naissance
		cell = fmt.Sprintf("D%d", dataRow)
		if migrant.Identite.UUID != "" {
			f.SetCellValue("Migrants", cell, migrant.Identite.DateNaissance.Format("02/01/2006"))
		} else {
			f.SetCellValue("Migrants", cell, "")
		}
		f.SetCellStyle("Migrants", cell, cell, dateStyle)

		// Lieu de naissance
		cell = fmt.Sprintf("E%d", dataRow)
		if migrant.Identite.UUID != "" {
			f.SetCellValue("Migrants", cell, migrant.Identite.LieuNaissance)
		} else {
			f.SetCellValue("Migrants", cell, "")
		}
		f.SetCellStyle("Migrants", cell, cell, dataStyle)

		// Sexe
		cell = fmt.Sprintf("F%d", dataRow)
		if migrant.Identite.UUID != "" {
			f.SetCellValue("Migrants", cell, migrant.Identite.Sexe)
		} else {
			f.SetCellValue("Migrants", cell, "")
		}
		f.SetCellStyle("Migrants", cell, cell, dataStyle)

		// Nationalité
		cell = fmt.Sprintf("G%d", dataRow)
		if migrant.Identite.UUID != "" {
			f.SetCellValue("Migrants", cell, migrant.Identite.Nationalite)
		} else {
			f.SetCellValue("Migrants", cell, "")
		}
		f.SetCellStyle("Migrants", cell, cell, dataStyle)

		// Type de document
		cell = fmt.Sprintf("H%d", dataRow)
		f.SetCellValue("Migrants", cell, "Passeport")
		f.SetCellStyle("Migrants", cell, cell, dataStyle)

		// N° Document
		cell = fmt.Sprintf("I%d", dataRow)
		if migrant.Identite.UUID != "" {
			f.SetCellValue("Migrants", cell, migrant.Identite.NumeroPasseport)
		} else {
			f.SetCellValue("Migrants", cell, "")
		}
		f.SetCellStyle("Migrants", cell, cell, dataStyle)

		// Date émission doc
		cell = fmt.Sprintf("J%d", dataRow)
		f.SetCellValue("Migrants", cell, "")
		f.SetCellStyle("Migrants", cell, cell, dateStyle)

		// Date expiration doc
		cell = fmt.Sprintf("K%d", dataRow)
		f.SetCellValue("Migrants", cell, "")
		f.SetCellStyle("Migrants", cell, cell, dateStyle)

		// Autorité émission
		cell = fmt.Sprintf("L%d", dataRow)
		if migrant.Identite.UUID != "" {
			f.SetCellValue("Migrants", cell, migrant.Identite.AutoriteEmetteur)
		} else {
			f.SetCellValue("Migrants", cell, "")
		}
		f.SetCellStyle("Migrants", cell, cell, dataStyle)

		// Téléphone
		cell = fmt.Sprintf("M%d", dataRow)
		f.SetCellValue("Migrants", cell, migrant.Telephone)
		f.SetCellStyle("Migrants", cell, cell, dataStyle)

		// Email
		cell = fmt.Sprintf("N%d", dataRow)
		f.SetCellValue("Migrants", cell, migrant.Email)
		f.SetCellStyle("Migrants", cell, cell, dataStyle)

		// Adresse actuelle
		cell = fmt.Sprintf("O%d", dataRow)
		f.SetCellValue("Migrants", cell, migrant.AdresseActuelle)
		f.SetCellStyle("Migrants", cell, cell, dataStyle)

		// Ville actuelle
		cell = fmt.Sprintf("P%d", dataRow)
		f.SetCellValue("Migrants", cell, migrant.VilleActuelle)
		f.SetCellStyle("Migrants", cell, cell, dataStyle)

		// Pays actuel
		cell = fmt.Sprintf("Q%d", dataRow)
		f.SetCellValue("Migrants", cell, migrant.PaysActuel)
		f.SetCellStyle("Migrants", cell, cell, dataStyle)

		// Situation matrimoniale
		cell = fmt.Sprintf("R%d", dataRow)
		f.SetCellValue("Migrants", cell, migrant.SituationMatrimoniale)
		f.SetCellStyle("Migrants", cell, cell, dataStyle)

		// Nombre enfants
		cell = fmt.Sprintf("S%d", dataRow)
		f.SetCellValue("Migrants", cell, migrant.NombreEnfants)
		f.SetCellStyle("Migrants", cell, cell, numberStyle)

		// Personne contact
		cell = fmt.Sprintf("T%d", dataRow)
		f.SetCellValue("Migrants", cell, migrant.PersonneContact)
		f.SetCellStyle("Migrants", cell, cell, dataStyle)

		// Téléphone contact
		cell = fmt.Sprintf("U%d", dataRow)
		f.SetCellValue("Migrants", cell, migrant.TelephoneContact)
		f.SetCellStyle("Migrants", cell, cell, dataStyle)

		// Statut migratoire
		cell = fmt.Sprintf("V%d", dataRow)
		f.SetCellValue("Migrants", cell, migrant.StatutMigratoire)
		f.SetCellStyle("Migrants", cell, cell, dataStyle)

		// Date d'entrée
		cell = fmt.Sprintf("W%d", dataRow)
		if migrant.DateEntree != nil {
			f.SetCellValue("Migrants", cell, migrant.DateEntree.Format("02/01/2006"))
		} else {
			f.SetCellValue("Migrants", cell, "")
		}
		f.SetCellStyle("Migrants", cell, cell, dateStyle)

		// Point d'entrée
		cell = fmt.Sprintf("X%d", dataRow)
		f.SetCellValue("Migrants", cell, migrant.PointEntree)
		f.SetCellStyle("Migrants", cell, cell, dataStyle)

		// Pays destination
		cell = fmt.Sprintf("Y%d", dataRow)
		f.SetCellValue("Migrants", cell, migrant.PaysDestination)
		f.SetCellStyle("Migrants", cell, cell, dataStyle)

		// Date création
		cell = fmt.Sprintf("Z%d", dataRow)
		f.SetCellValue("Migrants", cell, migrant.CreatedAt.Format("02/01/2006 15:04"))
		f.SetCellStyle("Migrants", cell, cell, dateStyle)

		// Date MAJ
		cell = fmt.Sprintf("AA%d", dataRow)
		f.SetCellValue("Migrants", cell, migrant.UpdatedAt.Format("02/01/2006 15:04"))
		f.SetCellStyle("Migrants", cell, cell, dateStyle)

		// Définir la hauteur de ligne
		f.SetRowHeight("Migrants", dataRow, 20)
	}

	// ===== AJUSTEMENT DE LA LARGEUR DES COLONNES =====
	columnWidths := []float64{
		18, // N° Identifiant
		15, // Nom
		15, // Prénom
		12, // Date naissance
		20, // Lieu naissance
		6,  // Sexe
		15, // Nationalité
		18, // Type document
		18, // N° Document
		12, // Date émission
		12, // Date expiration
		20, // Autorité émission
		15, // Téléphone
		25, // Email
		30, // Adresse
		15, // Ville
		15, // Pays actuel
		18, // Situation matrimoniale
		8,  // Nombre enfants
		20, // Personne contact
		15, // Téléphone contact
		18, // Statut migratoire
		12, // Date entrée
		20, // Point entrée
		15, // Pays destination
		18, // Date création
		18, // Date MAJ
	}

	columns := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z", "AA"}
	for i, width := range columnWidths {
		if i < len(columns) {
			f.SetColWidth("Migrants", columns[i], columns[i], width)
		}
	}

	// ===== AJOUTER UNE FEUILLE DE STATISTIQUES =====
	_, err = f.NewSheet("Statistiques")
	if err == nil {
		// Calculer les statistiques
		totalRecords := len(migrants)

		// Compter par statut migratoire
		statutCount := make(map[string]int)
		nationaliteCount := make(map[string]int)
		sexeCount := make(map[string]int)

		for _, migrant := range migrants {
			statutCount[migrant.StatutMigratoire]++
			if migrant.Identite.UUID != "" {
				nationaliteCount[migrant.Identite.Nationalite]++
				sexeCount[migrant.Identite.Sexe]++
			}
		}

		// En-tête de la feuille statistiques
		f.SetCellValue("Statistiques", "A1", "STATISTIQUES DES MIGRANTS")
		f.MergeCell("Statistiques", "A1", "C1")
		f.SetCellStyle("Statistiques", "A1", "C1", headerStyle)

		row = 3
		f.SetCellValue("Statistiques", fmt.Sprintf("A%d", row), "Total des enregistrements:")
		f.SetCellValue("Statistiques", fmt.Sprintf("B%d", row), totalRecords)
		row += 2

		// Par statut migratoire
		f.SetCellValue("Statistiques", fmt.Sprintf("A%d", row), "Par statut migratoire:")
		f.SetCellStyle("Statistiques", fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), columnHeaderStyle)
		row++
		for statut, count := range statutCount {
			f.SetCellValue("Statistiques", fmt.Sprintf("A%d", row), statut)
			f.SetCellValue("Statistiques", fmt.Sprintf("B%d", row), count)
			row++
		}
		row++

		// Par sexe
		f.SetCellValue("Statistiques", fmt.Sprintf("A%d", row), "Par sexe:")
		f.SetCellStyle("Statistiques", fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), columnHeaderStyle)
		row++
		for sexe, count := range sexeCount {
			f.SetCellValue("Statistiques", fmt.Sprintf("A%d", row), sexe)
			f.SetCellValue("Statistiques", fmt.Sprintf("B%d", row), count)
			row++
		}
		row++

		// Top 10 nationalités
		f.SetCellValue("Statistiques", fmt.Sprintf("A%d", row), "Top 10 nationalités:")
		f.SetCellStyle("Statistiques", fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), columnHeaderStyle)
		row++
		count := 0
		for nationalite, nb := range nationaliteCount {
			if count >= 10 {
				break
			}
			f.SetCellValue("Statistiques", fmt.Sprintf("A%d", row), nationalite)
			f.SetCellValue("Statistiques", fmt.Sprintf("B%d", row), nb)
			row++
			count++
		}

		f.SetColWidth("Statistiques", "A", "A", 25)
		f.SetColWidth("Statistiques", "B", "B", 15)
	}

	// ===== GÉNÉRATION DU FICHIER =====
	filename := fmt.Sprintf("migrants_export_%s.xlsx", time.Now().Format("20060102_150405"))

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
