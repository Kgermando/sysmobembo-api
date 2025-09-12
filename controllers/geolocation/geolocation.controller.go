package geolocation

import (
	"fmt"
	"math"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/kgermando/sysmobembo-api/database"
	"github.com/kgermando/sysmobembo-api/models"
	"github.com/kgermando/sysmobembo-api/utils"
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

// Vérifier si un point est dans un rayon donné
func isWithinRadius(centerLat, centerLon, pointLat, pointLon, radiusKm float64) bool {
	distance := calculateDistance(centerLat, centerLon, pointLat, pointLon)
	return distance <= radiusKm
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

	// Vérifier que le migrant existe
	var migrant models.Migrant
	if err := db.Where("uuid = ?", migrantUUID).First(&migrant).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Migrant not found",
			"data":    nil,
		})
	}

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
		"migrant":    migrant,
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
	if geolocalisation.MigrantUUID == "" || geolocalisation.TypeLocalisation == "" || geolocalisation.Pays == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "MigrantUUID, TypeLocalisation, and Pays are required",
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

	// Vérifier que le migrant existe
	var migrant models.Migrant
	if err := database.DB.Where("uuid = ?", geolocalisation.MigrantUUID).First(&migrant).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Migrant not found",
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

	// Recharger avec les relations
	database.DB.Preload("Migrant").First(geolocalisation, "uuid = ?", geolocalisation.UUID)

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

	// Recharger avec les relations
	db.Preload("Migrant").First(&geolocalisation)

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
// ANALYTICS & STATISTICS
// =======================

// Get geolocations statistics
func GetGeolocalisationsStats(c *fiber.Ctx) error {
	db := database.DB

	var totalGeolocalisations int64

	// Statistiques générales
	db.Model(&models.Geolocalisation{}).Count(&totalGeolocalisations)

	// Statistiques par type de localisation
	var localisationTypes []map[string]interface{}
	db.Model(&models.Geolocalisation{}).
		Select("type_localisation, COUNT(*) as count").
		Group("type_localisation").
		Order("count DESC").
		Scan(&localisationTypes)

	// Statistiques par pays
	var countryStats []map[string]interface{}
	db.Model(&models.Geolocalisation{}).
		Select("pays, COUNT(*) as count").
		Group("pays").
		Order("count DESC").
		Limit(10).
		Scan(&countryStats)

	// Statistiques par type de mouvement
	var movementStats []map[string]interface{}
	db.Model(&models.Geolocalisation{}).
		Where("type_mouvement IS NOT NULL AND type_mouvement != ''").
		Select("type_mouvement, COUNT(*) as count").
		Group("type_mouvement").
		Order("count DESC").
		Scan(&movementStats)

	stats := map[string]interface{}{
		"total_geolocations":   totalGeolocalisations,
		"location_types":       localisationTypes,
		"country_distribution": countryStats,
		"movement_types":       movementStats,
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Geolocations statistics",
		"data":    stats,
	})
}
