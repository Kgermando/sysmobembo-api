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
	actif := c.Query("actif", "")

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

	// Filtrer par statut actif
	if actif != "" {
		isActive := actif == "true"
		query = query.Where("actif = ?", isActive)
	}

	// Count total
	query.Count(&totalRecords)

	// Get paginated results
	err = query.Offset(offset).
		Limit(limit).
		Order("date_enregistrement DESC").
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
		Order("date_enregistrement DESC").
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

// Get geolocations by migrant
func GetGeolocalisationsByMigrant(c *fiber.Ctx) error {
	migrantUUID := c.Params("migrant_uuid")
	db := database.DB
	var geolocalisations []models.Geolocalisation

	err := db.Where("migrant_uuid = ?", migrantUUID).
		Preload("Migrant").
		Order("date_enregistrement DESC").
		Find(&geolocalisations).Error

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch geolocations for migrant",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Geolocations for migrant",
		"data":    geolocalisations,
	})
}

// Get active geolocations
func GetActiveGeolocalisations(c *fiber.Ctx) error {
	db := database.DB
	var geolocalisations []models.Geolocalisation

	err := db.Where("actif = ?", true).
		Preload("Migrant").
		Order("date_enregistrement DESC").
		Find(&geolocalisations).Error

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch active geolocations",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Active geolocations",
		"data":    geolocalisations,
	})
}

// Get geolocations within radius
func GetGeolocalisationsWithinRadius(c *fiber.Ctx) error {
	db := database.DB

	// Parse parameters
	latStr := c.Query("latitude")
	lonStr := c.Query("longitude")
	radiusStr := c.Query("radius", "10") // 10km par défaut

	if latStr == "" || lonStr == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Latitude and longitude are required",
			"data":    nil,
		})
	}

	centerLat, err := strconv.ParseFloat(latStr, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid latitude format",
			"data":    nil,
		})
	}

	centerLon, err := strconv.ParseFloat(lonStr, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid longitude format",
			"data":    nil,
		})
	}

	radius, err := strconv.ParseFloat(radiusStr, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid radius format",
			"data":    nil,
		})
	}

	if err := validateCoordinates(centerLat, centerLon); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": err.Error(),
			"data":    nil,
		})
	}

	var allGeolocalisations []models.Geolocalisation
	err = db.Preload("Migrant").Find(&allGeolocalisations).Error
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch geolocations",
			"error":   err.Error(),
		})
	}

	// Filtrer par distance
	var nearbyGeolocalisations []models.Geolocalisation
	for _, geo := range allGeolocalisations {
		if isWithinRadius(centerLat, centerLon, geo.Latitude, geo.Longitude, radius) {
			nearbyGeolocalisations = append(nearbyGeolocalisations, geo)
		}
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": fmt.Sprintf("Geolocations within %.1f km", radius),
		"data":    nearbyGeolocalisations,
		"center":  map[string]float64{"latitude": centerLat, "longitude": centerLon},
		"radius":  radius,
		"count":   len(nearbyGeolocalisations),
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

	// Définir des valeurs par défaut
	if geolocalisation.FiabiliteSource == "" {
		geolocalisation.FiabiliteSource = "moyenne"
	}
	if geolocalisation.MethodeCapture == "" {
		geolocalisation.MethodeCapture = "manuel"
	}

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

// Validate geolocation
func ValidateGeolocalisation(c *fiber.Ctx) error {
	uuid := c.Params("uuid")
	db := database.DB

	var validationData struct {
		ValidePar   string `json:"valide_par"`
		Commentaire string `json:"commentaire"`
	}

	if err := c.BodyParser(&validationData); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request format",
			"error":   err.Error(),
		})
	}

	var geolocalisation models.Geolocalisation
	if err := db.Where("uuid = ?", uuid).First(&geolocalisation).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Geolocation not found",
			"data":    nil,
		})
	}

	// Marquer comme validé
	now := time.Now()
	updateData := map[string]interface{}{
		"date_validation": &now,
		"valide_par":      validationData.ValidePar,
		"commentaire":     validationData.Commentaire,
	}

	if err := db.Model(&geolocalisation).Updates(updateData).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to validate geolocation",
			"error":   err.Error(),
		})
	}

	// Recharger avec les relations
	db.Preload("Migrant").First(&geolocalisation)

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Geolocation validated successfully",
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
	var activeGeolocalisations int64
	var validatedGeolocalisations int64

	// Statistiques générales
	db.Model(&models.Geolocalisation{}).Count(&totalGeolocalisations)
	db.Model(&models.Geolocalisation{}).Where("actif = ?", true).Count(&activeGeolocalisations)
	db.Model(&models.Geolocalisation{}).Where("date_validation IS NOT NULL").Count(&validatedGeolocalisations)

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

	// Statistiques par méthode de capture
	var captureMethodStats []map[string]interface{}
	db.Model(&models.Geolocalisation{}).
		Select("methode_capture, COUNT(*) as count").
		Group("methode_capture").
		Order("count DESC").
		Scan(&captureMethodStats)

	// Statistiques par fiabilité
	var reliabilityStats []map[string]interface{}
	db.Model(&models.Geolocalisation{}).
		Select("fiabilite_source, COUNT(*) as count").
		Group("fiabilite_source").
		Order("count DESC").
		Scan(&reliabilityStats)

	// Statistiques par type de mouvement
	var movementStats []map[string]interface{}
	db.Model(&models.Geolocalisation{}).
		Where("type_mouvement IS NOT NULL AND type_mouvement != ''").
		Select("type_mouvement, COUNT(*) as count").
		Group("type_mouvement").
		Order("count DESC").
		Scan(&movementStats)

	stats := map[string]interface{}{
		"total_geolocations":       totalGeolocalisations,
		"active_geolocations":      activeGeolocalisations,
		"validated_geolocations":   validatedGeolocalisations,
		"location_types":           localisationTypes,
		"country_distribution":     countryStats,
		"capture_methods":          captureMethodStats,
		"reliability_distribution": reliabilityStats,
		"movement_types":           movementStats,
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Geolocations statistics",
		"data":    stats,
	})
}

// Get migration routes analysis
func GetMigrationRoutes(c *fiber.Ctx) error {
	db := database.DB

	// Analyser les routes de migration en regroupant par migrant et en ordonnant par date
	var routes []map[string]interface{}

	// Cette requête complexe groupe les geolocalisations par migrant et les ordonne par date
	// pour identifier les routes de migration
	err := db.Raw(`
		SELECT 
			m.uuid as migrant_uuid,
			m.nom,
			m.prenom,
			m.nationalite,
			COUNT(g.uuid) as locations_count,
			string_agg(DISTINCT g.pays, ' -> ' ORDER BY g.pays) as countries_visited,
			MIN(g.date_enregistrement) as first_location,
			MAX(g.date_enregistrement) as last_location
		FROM migrants m
		JOIN geolocalisations g ON m.uuid = g.migrant_uuid
		WHERE g.actif = true
		GROUP BY m.uuid, m.nom, m.prenom, m.nationalite
		HAVING COUNT(g.uuid) > 1
		ORDER BY locations_count DESC
		LIMIT 50
	`).Scan(&routes).Error

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to analyze migration routes",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Migration routes analysis",
		"data":    routes,
	})
}

// Get hotspots analysis
func GetGeolocationHotspots(c *fiber.Ctx) error {
	db := database.DB

	// Analyser les points chauds basés sur la densité de géolocalisations
	var hotspots []map[string]interface{}

	err := db.Model(&models.Geolocalisation{}).
		Select("ville, pays, COUNT(*) as density, AVG(latitude) as avg_latitude, AVG(longitude) as avg_longitude").
		Where("ville IS NOT NULL AND ville != ''").
		Group("ville, pays").
		Having("COUNT(*) >= 3"). // Minimum 3 localisations pour être considéré comme hotspot
		Order("density DESC").
		Limit(20).
		Scan(&hotspots).Error

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to analyze geolocation hotspots",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Geolocation hotspots analysis",
		"data":    hotspots,
	})
}

// Search with advanced filters
func SearchGeolocalisations(c *fiber.Ctx) error {
	db := database.DB

	typeLocalisation := c.Query("type_localisation")
	pays := c.Query("pays")
	ville := c.Query("ville")
	methodeCapture := c.Query("methode_capture")
	fiabilite := c.Query("fiabilite")
	typeMouvement := c.Query("type_mouvement")
	actif := c.Query("actif")
	dateFrom := c.Query("date_from")
	dateTo := c.Query("date_to")

	var geolocalisations []models.Geolocalisation
	query := db.Model(&models.Geolocalisation{}).Preload("Migrant")

	if typeLocalisation != "" {
		query = query.Where("type_localisation = ?", typeLocalisation)
	}
	if pays != "" {
		query = query.Where("pays ILIKE ?", "%"+pays+"%")
	}
	if ville != "" {
		query = query.Where("ville ILIKE ?", "%"+ville+"%")
	}
	if methodeCapture != "" {
		query = query.Where("methode_capture = ?", methodeCapture)
	}
	if fiabilite != "" {
		query = query.Where("fiabilite_source = ?", fiabilite)
	}
	if typeMouvement != "" {
		query = query.Where("type_mouvement = ?", typeMouvement)
	}
	if actif != "" {
		isActive := actif == "true"
		query = query.Where("actif = ?", isActive)
	}
	if dateFrom != "" {
		query = query.Where("date_enregistrement >= ?", dateFrom)
	}
	if dateTo != "" {
		query = query.Where("date_enregistrement <= ?", dateTo)
	}

	err := query.Order("date_enregistrement DESC").Find(&geolocalisations).Error
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to search geolocations",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Search results for geolocations",
		"data":    geolocalisations,
	})
}
