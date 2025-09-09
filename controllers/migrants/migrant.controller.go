package migrants

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/kgermando/sysmobembo-api/database"
	"github.com/kgermando/sysmobembo-api/models"
	"github.com/kgermando/sysmobembo-api/utils"
)

// Fonction pour générer automatiquement le NumeroIdentifiant
func generateNumeroIdentifiant() string {
	// Format: MIG-YYYY-XXXXXX (où YYYY = année, XXXXXX = numéro séquentiel)
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
	statutMigratoire := c.Query("statut_migratoire", "")
	nationalite := c.Query("nationalite", "")
	paysOrigine := c.Query("pays_origine", "")
	genre := c.Query("genre", "")
	actif := c.Query("actif", "")
	typeDocument := c.Query("type_document", "")

	// Date filters
	dateCreationDebut := c.Query("date_creation_debut", "")
	dateCreationFin := c.Query("date_creation_fin", "")
	dateNaissanceDebut := c.Query("date_naissance_debut", "")
	dateNaissanceFin := c.Query("date_naissance_fin", "")

	var migrants []models.Migrant
	var totalRecords int64

	// Build query with filters
	query := db.Model(&models.Migrant{})

	// Search filter
	if search != "" {
		query = query.Where("nom ILIKE ? OR prenom ILIKE ? OR numero_identifiant ILIKE ? OR nationalite ILIKE ? OR numero_document ILIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	// Status filters
	if statutMigratoire != "" {
		query = query.Where("statut_migratoire = ?", statutMigratoire)
	}

	if nationalite != "" {
		query = query.Where("nationalite ILIKE ?", "%"+nationalite+"%")
	}

	if paysOrigine != "" {
		query = query.Where("pays_origine ILIKE ?", "%"+paysOrigine+"%")
	}

	if genre != "" {
		query = query.Where("genre = ?", genre)
	}

	if actif != "" {
		if actif == "true" {
			query = query.Where("actif = ?", true)
		} else if actif == "false" {
			query = query.Where("actif = ?", false)
		}
	}

	if typeDocument != "" {
		query = query.Where("type_document = ?", typeDocument)
	}

	// Date filters
	if dateCreationDebut != "" {
		if parsedDate, err := time.Parse("2006-01-02", dateCreationDebut); err == nil {
			query = query.Where("created_at >= ?", parsedDate)
		}
	}

	if dateCreationFin != "" {
		if parsedDate, err := time.Parse("2006-01-02", dateCreationFin); err == nil {
			// Ajouter 23:59:59 pour inclure toute la journée
			parsedDate = parsedDate.Add(23*time.Hour + 59*time.Minute + 59*time.Second)
			query = query.Where("created_at <= ?", parsedDate)
		}
	}

	if dateNaissanceDebut != "" {
		if parsedDate, err := time.Parse("2006-01-02", dateNaissanceDebut); err == nil {
			query = query.Where("date_naissance >= ?", parsedDate)
		}
	}

	if dateNaissanceFin != "" {
		if parsedDate, err := time.Parse("2006-01-02", dateNaissanceFin); err == nil {
			query = query.Where("date_naissance <= ?", parsedDate)
		}
	}

	// Count total records with filters applied
	query.Count(&totalRecords)

	// Execute query with pagination
	err = query.
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

	// Prepare applied filters for response
	appliedFilters := map[string]interface{}{
		"search":               search,
		"statut_migratoire":    statutMigratoire,
		"nationalite":          nationalite,
		"pays_origine":         paysOrigine,
		"genre":                genre,
		"actif":                actif,
		"type_document":        typeDocument,
		"date_creation_debut":  dateCreationDebut,
		"date_creation_fin":    dateCreationFin,
		"date_naissance_debut": dateNaissanceDebut,
		"date_naissance_fin":   dateNaissanceFin,
	}

	// Return response
	return c.JSON(fiber.Map{
		"status":          "success",
		"message":         "Migrants retrieved successfully",
		"data":            migrants,
		"pagination":      pagination,
		"applied_filters": appliedFilters,
	})
}

// Query all data
func GetAllMigrants(c *fiber.Ctx) error {
	db := database.DB
	var migrants []models.Migrant

	err := db.Preload("MotifDeplacements").
		Preload("Alertes").
		Preload("Biometries").
		Preload("Geolocalisations").
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

	// Validation des champs requis
	if migrant.Nom == "" || migrant.Prenom == "" || migrant.PaysOrigine == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Form not complete - nom, prenom, and pays_origine are required",
			"data":    nil,
		})
	}

	// Générer automatiquement l'UUID et le NumeroIdentifiant
	migrant.UUID = utils.GenerateUUID()
	migrant.NumeroIdentifiant = generateNumeroIdentifiant()

	// Validation des données
	if err := utils.ValidateStruct(*migrant); err != nil {
		return c.Status(400).JSON(err)
	}

	// Vérifier l'unicité du numéro de document s'il est fourni
	if migrant.NumeroDocument != "" {
		var existingMigrant models.Migrant
		if err := database.DB.Where("numero_document = ?", migrant.NumeroDocument).First(&existingMigrant).Error; err == nil {
			return c.Status(400).JSON(fiber.Map{
				"status":  "error",
				"message": "A migrant with this document number already exists",
				"data":    nil,
			})
		}
	}

	// Vérifier l'unicité de l'email s'il est fourni
	if migrant.Email != "" {
		var existingMigrant models.Migrant
		if err := database.DB.Where("email = ?", migrant.Email).First(&existingMigrant).Error; err == nil {
			return c.Status(400).JSON(fiber.Map{
				"status":  "error",
				"message": "A migrant with this email already exists",
				"data":    nil,
			})
		}
	}

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

	// Vérifier l'unicité du numéro de document s'il est modifié
	if updateData.NumeroDocument != "" && updateData.NumeroDocument != migrant.NumeroDocument {
		var existingMigrant models.Migrant
		if err := database.DB.Where("numero_document = ? AND uuid != ?", updateData.NumeroDocument, uuid).First(&existingMigrant).Error; err == nil {
			return c.Status(400).JSON(fiber.Map{
				"status":  "error",
				"message": "A migrant with this document number already exists",
				"data":    nil,
			})
		}
	}

	// Vérifier l'unicité de l'email s'il est modifié
	if updateData.Email != "" && updateData.Email != migrant.Email {
		var existingMigrant models.Migrant
		if err := database.DB.Where("email = ? AND uuid != ?", updateData.Email, uuid).First(&existingMigrant).Error; err == nil {
			return c.Status(400).JSON(fiber.Map{
				"status":  "error",
				"message": "A migrant with this email already exists",
				"data":    nil,
			})
		}
	}

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
	var activeMigrants int64
	var regularMigrants int64
	var irregularMigrants int64
	var refugeeMigrants int64
	var asylumSeekers int64

	// Total migrants
	db.Model(&models.Migrant{}).Count(&totalMigrants)

	// Active migrants
	db.Model(&models.Migrant{}).Where("actif = ?", true).Count(&activeMigrants)

	// Par statut migratoire
	db.Model(&models.Migrant{}).Where("statut_migratoire = ?", "regulier").Count(&regularMigrants)
	db.Model(&models.Migrant{}).Where("statut_migratoire = ?", "irregulier").Count(&irregularMigrants)
	db.Model(&models.Migrant{}).Where("statut_migratoire = ?", "refugie").Count(&refugeeMigrants)
	db.Model(&models.Migrant{}).Where("statut_migratoire = ?", "demandeur_asile").Count(&asylumSeekers)

	stats := map[string]interface{}{
		"total_migrants":     totalMigrants,
		"active_migrants":    activeMigrants,
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
