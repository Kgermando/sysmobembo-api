package migrants

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/kgermando/sysmobembo-api/database"
	"github.com/kgermando/sysmobembo-api/models"
	"github.com/kgermando/sysmobembo-api/utils"
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
