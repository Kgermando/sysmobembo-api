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

// Paginate - Récupérer les migrants avec pagination
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

	// Parse search query
	search := c.Query("search", "")

	var migrants []models.Migrant
	var totalRecords int64

	// Count total records matching the search query
	db.Model(&models.Migrant{}).
		Where("nom ILIKE ? OR prenom ILIKE ? OR numero_identifiant ILIKE ? OR nationalite ILIKE ? OR numero_document ILIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%").
		Count(&totalRecords)

	err = db.
		Where("nom ILIKE ? OR prenom ILIKE ? OR numero_identifiant ILIKE ? OR nationalite ILIKE ? OR numero_document ILIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%").
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

// Get migrant by NumeroIdentifiant
func GetMigrantByNumero(c *fiber.Ctx) error {
	numeroIdentifiant := c.Params("numero")
	db := database.DB
	var migrant models.Migrant

	err := db.Where("numero_identifiant = ?", numeroIdentifiant).
		Preload("MotifDeplacements").
		Preload("Alertes").
		Preload("Biometries").
		Preload("Geolocalisations").
		First(&migrant).Error

	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Migrant not found with this numero",
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

// Get migrants by nationality
func GetMigrantsByNationality(c *fiber.Ctx) error {
	db := database.DB

	var results []map[string]interface{}

	err := db.Model(&models.Migrant{}).
		Select("nationalite, COUNT(*) as count").
		Group("nationalite").
		Order("count DESC").
		Scan(&results).Error

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch nationality statistics",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Migrants by nationality",
		"data":    results,
	})
}

// Search migrants with advanced filters
func SearchMigrants(c *fiber.Ctx) error {
	db := database.DB

	// Parse query parameters
	nationalite := c.Query("nationalite")
	statut := c.Query("statut")
	sexe := c.Query("sexe")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")

	var migrants []models.Migrant
	query := db.Model(&models.Migrant{})

	// Apply filters
	if nationalite != "" {
		query = query.Where("nationalite = ?", nationalite)
	}
	if statut != "" {
		query = query.Where("statut_migratoire = ?", statut)
	}
	if sexe != "" {
		query = query.Where("sexe = ?", sexe)
	}
	if dateFrom != "" {
		query = query.Where("created_at >= ?", dateFrom)
	}
	if dateTo != "" {
		query = query.Where("created_at <= ?", dateTo)
	}

	err := query.Preload("MotifDeplacements").
		Preload("Alertes").
		Preload("Biometries").
		Preload("Geolocalisations").
		Find(&migrants).Error

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to search migrants",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Search results",
		"data":    migrants,
	})
}
