package alerts

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/kgermando/sysmobembo-api/database"
	"github.com/kgermando/sysmobembo-api/models"
	"github.com/kgermando/sysmobembo-api/utils"
)

// =======================
// CRUD OPERATIONS
// =======================

// Paginate - Récupérer les alertes avec pagination
func GetPaginatedAlerts(c *fiber.Ctx) error {
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
	statut := c.Query("statut", "")
	gravite := c.Query("gravite", "")

	var alerts []models.Alert
	var totalRecords int64

	query := db.Model(&models.Alert{}).Preload("Migrant")

	// Filtrer par migrant si spécifié
	if migrantUUID != "" {
		query = query.Where("migrant_uuid = ?", migrantUUID)
	}

	// Filtrer par statut
	if statut != "" {
		query = query.Where("statut = ?", statut)
	}

	// Filtrer par gravité
	if gravite != "" {
		query = query.Where("niveau_gravite = ?", gravite)
	}

	// Recherche textuelle
	if search != "" {
		query = query.Where("titre ILIKE ? OR description ILIKE ? OR type_alerte ILIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	// Count total
	query.Count(&totalRecords)

	// Get paginated results
	err = query.Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&alerts).Error

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch alerts",
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
		"message":    "Alerts retrieved successfully",
		"data":       alerts,
		"pagination": pagination,
	})
}

// Get all alerts
func GetAllAlerts(c *fiber.Ctx) error {
	db := database.DB
	var alerts []models.Alert

	err := db.Preload("Migrant").
		Order("niveau_gravite DESC, created_at DESC").
		Find(&alerts).Error

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch alerts",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "All alerts",
		"data":    alerts,
	})
}

// Get one alert
func GetAlert(c *fiber.Ctx) error {
	uuid := c.Params("uuid")
	db := database.DB
	var alert models.Alert

	err := db.Where("uuid = ?", uuid).
		Preload("Migrant").
		First(&alert).Error

	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Alert not found",
			"data":    nil,
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Alert found",
		"data":    alert,
	})
}

// Get alerts by migrant with pagination
func GetAlertsByMigrant(c *fiber.Ctx) error {
	migrantUUID := c.Params("migrant_uuid")
	db := database.DB

	// Paramètres de pagination
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page <= 0 {
		page = 1
	}
	limit, err := strconv.Atoi(c.Query("limit", "15"))
	if err != nil || limit <= 0 {
		limit = 15
	}
	offset := (page - 1) * limit

	// Paramètres de filtrage
	search := c.Query("search", "")
	statut := c.Query("statut", "")
	gravite := c.Query("gravite", "")

	var alerts []models.Alert
	var totalRecords int64

	query := db.Model(&models.Alert{}).Where("migrant_uuid = ?", migrantUUID).Preload("Migrant")

	// Filtrer par statut
	if statut != "" {
		query = query.Where("statut = ?", statut)
	}

	// Filtrer par gravité
	if gravite != "" {
		query = query.Where("niveau_gravite = ?", gravite)
	}

	// Recherche textuelle
	if search != "" {
		query = query.Where("titre ILIKE ? OR description ILIKE ? OR type_alerte ILIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	// Count total
	query.Count(&totalRecords)

	// Get paginated results
	err = query.Offset(offset).
		Limit(limit).
		Order("niveau_gravite DESC, created_at DESC").
		Find(&alerts).Error

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch alerts for migrant",
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
		"message":    "Alerts for migrant retrieved successfully",
		"data":       alerts,
		"pagination": pagination,
	})
}

// Create alert
func CreateAlert(c *fiber.Ctx) error {
	alert := &models.Alert{}

	if err := c.BodyParser(alert); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request format",
			"error":   err.Error(),
		})
	}

	// Validation des champs requis
	if alert.MigrantUUID == "" || alert.TypeAlerte == "" || alert.Titre == "" || alert.Description == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "MigrantUUID, TypeAlerte, Titre, and Description are required",
			"data":    nil,
		})
	}

	// Vérifier que le migrant existe
	var migrant models.Migrant
	if err := database.DB.Where("uuid = ?", alert.MigrantUUID).First(&migrant).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Migrant not found",
			"data":    nil,
		})
	}

	// Générer l'UUID
	alert.UUID = utils.GenerateUUID()

	// Définir le statut par défaut si non spécifié
	if alert.Statut == "" {
		alert.Statut = "active"
	}

	// Validation des données
	if err := utils.ValidateStruct(*alert); err != nil {
		return c.Status(400).JSON(err)
	}

	if err := database.DB.Create(alert).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to create alert",
			"error":   err.Error(),
		})
	}

	// Recharger avec les relations
	database.DB.Preload("Migrant").First(alert, "uuid = ?", alert.UUID)

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Alert created successfully",
		"data":    alert,
	})
}

// Update alert
func UpdateAlert(c *fiber.Ctx) error {
	uuid := c.Params("uuid")
	db := database.DB

	var updateData models.Alert
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Review your input",
			"error":   err.Error(),
		})
	}

	alert := new(models.Alert)
	if err := db.Where("uuid = ?", uuid).First(&alert).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Alert not found",
			"data":    nil,
		})
	}

	// Conserver l'UUID
	updateData.UUID = alert.UUID

	if err := db.Model(&alert).Updates(updateData).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to update alert",
			"error":   err.Error(),
		})
	}

	// Recharger avec les relations
	db.Preload("Migrant").First(&alert)

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Alert updated successfully",
		"data":    alert,
	})
}

// Resolve alert
func ResolveAlert(c *fiber.Ctx) error {
	uuid := c.Params("uuid")
	db := database.DB

	var resolutionData struct {
		CommentaireResolution string `json:"commentaire_resolution"`
	}

	if err := c.BodyParser(&resolutionData); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request format",
			"error":   err.Error(),
		})
	}

	var alert models.Alert
	if err := db.Where("uuid = ?", uuid).First(&alert).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Alert not found",
			"data":    nil,
		})
	}

	// Marquer comme résolu
	now := time.Now()
	updateData := map[string]interface{}{
		"statut":                 "resolved",
		"date_resolution":        &now,
		"commentaire_resolution": resolutionData.CommentaireResolution,
	}

	if err := db.Model(&alert).Updates(updateData).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to resolve alert",
			"error":   err.Error(),
		})
	}

	// Recharger avec les relations
	db.Preload("Migrant").First(&alert)

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Alert resolved successfully",
		"data":    alert,
	})
}

// Delete alert
func DeleteAlert(c *fiber.Ctx) error {
	uuid := c.Params("uuid")
	db := database.DB

	var alert models.Alert
	if err := db.Where("uuid = ?", uuid).First(&alert).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Alert not found",
			"data":    nil,
		})
	}

	if err := db.Delete(&alert).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to delete alert",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Alert deleted successfully",
		"data":    nil,
	})
}

// =======================
// ANALYTICS & STATISTICS
// =======================

// Get alerts statistics
func GetAlertsStats(c *fiber.Ctx) error {
	db := database.DB

	var totalAlerts int64
	var activeAlerts int64
	var resolvedAlerts int64
	var criticalAlerts int64
	var expiredAlerts int64

	// Statistiques générales
	db.Model(&models.Alert{}).Count(&totalAlerts)
	db.Model(&models.Alert{}).Where("statut = ?", "active").Count(&activeAlerts)
	db.Model(&models.Alert{}).Where("statut = ?", "resolved").Count(&resolvedAlerts)
	db.Model(&models.Alert{}).Where("niveau_gravite = ?", "critical").Count(&criticalAlerts)
	db.Model(&models.Alert{}).Where("statut = ?", "expired").Count(&expiredAlerts)

	// Statistiques par type d'alerte
	var alertTypes []map[string]interface{}
	db.Model(&models.Alert{}).
		Select("type_alerte, COUNT(*) as count").
		Group("type_alerte").
		Order("count DESC").
		Scan(&alertTypes)

	// Statistiques par niveau de gravité
	var gravityStats []map[string]interface{}
	db.Model(&models.Alert{}).
		Select("niveau_gravite, COUNT(*) as count").
		Group("niveau_gravite").
		Order("count DESC").
		Scan(&gravityStats)

	stats := map[string]interface{}{
		"total_alerts":         totalAlerts,
		"active_alerts":        activeAlerts,
		"resolved_alerts":      resolvedAlerts,
		"critical_alerts":      criticalAlerts,
		"expired_alerts":       expiredAlerts,
		"alert_types":          alertTypes,
		"gravity_distribution": gravityStats,
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Alerts statistics",
		"data":    stats,
	})
}
