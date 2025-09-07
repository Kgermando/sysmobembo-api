package dashboard

import (
	"math"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/kgermando/sysmobembo-api/database"
	"github.com/kgermando/sysmobembo-api/models"
)

// =======================
// ANALYSE SPATIALE
// =======================

// Structure pour l'analyse spatiale
type SpatialCluster struct {
	CenterLatitude  float64 `json:"center_latitude"`
	CenterLongitude float64 `json:"center_longitude"`
	Radius          float64 `json:"radius_km"`
	MigrantCount    int64   `json:"migrant_count"`
	Ville           string  `json:"ville"`
	Pays            string  `json:"pays"`
	DensityScore    float64 `json:"density_score"`
}

type HeatmapPoint struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Intensity float64 `json:"intensity"`
	Count     int64   `json:"count"`
}

// Calcul de distance pour l'analyse spatiale
func calculateSpatialDistance(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371
	lat1Rad := lat1 * math.Pi / 180
	lon1Rad := lon1 * math.Pi / 180
	lat2Rad := lat2 * math.Pi / 180
	lon2Rad := lon2 * math.Pi / 180

	dlat := lat2Rad - lat1Rad
	dlon := lon2Rad - lon1Rad

	a := math.Sin(dlat/2)*math.Sin(dlat/2) + math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(dlon/2)*math.Sin(dlon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadius * c
}

// Analyse de densité spatiale
func GetSpatialDensityAnalysis(c *fiber.Ctx) error {
	db := database.DB

	radiusKm, _ := strconv.ParseFloat(c.Query("radius", "10"), 64)
	daysPeriod, _ := strconv.Atoi(c.Query("days", "30"))
	minMigrants, _ := strconv.Atoi(c.Query("min_migrants", "3"))

	// Récupérer toutes les positions uniques
	var positions []struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Ville     string  `json:"ville"`
		Pays      string  `json:"pays"`
		Count     int64   `json:"count"`
	}

	db.Raw(`
		SELECT 
			ROUND(latitude::numeric, 4)::float as latitude,
			ROUND(longitude::numeric, 4)::float as longitude,
			ville,
			pays,
			COUNT(DISTINCT migrant_uuid) as count
		FROM geolocalisations
		WHERE created_at >= NOW() - INTERVAL '%d days'
		AND latitude IS NOT NULL
		AND longitude IS NOT NULL
		AND deleted_at IS NULL
		GROUP BY ROUND(latitude::numeric, 4), ROUND(longitude::numeric, 4), ville, pays
		HAVING COUNT(DISTINCT migrant_uuid) >= %d
		ORDER BY count DESC
	`, daysPeriod, minMigrants).Scan(&positions)

	// Identifier les clusters de densité
	var clusters []SpatialCluster
	processed := make(map[string]bool)

	for _, pos := range positions {
		key := strconv.FormatFloat(pos.Latitude, 'f', 4, 64) + "," + strconv.FormatFloat(pos.Longitude, 'f', 4, 64)
		if processed[key] {
			continue
		}

		// Créer un cluster centré sur cette position
		cluster := SpatialCluster{
			CenterLatitude:  pos.Latitude,
			CenterLongitude: pos.Longitude,
			Radius:          radiusKm,
			MigrantCount:    pos.Count,
			Ville:           pos.Ville,
			Pays:            pos.Pays,
		}

		// Chercher d'autres points dans le rayon
		var totalMigrants int64 = pos.Count
		for _, otherPos := range positions {
			if otherPos.Latitude == pos.Latitude && otherPos.Longitude == pos.Longitude {
				continue
			}

			distance := calculateSpatialDistance(
				pos.Latitude, pos.Longitude,
				otherPos.Latitude, otherPos.Longitude,
			)

			if distance <= radiusKm {
				totalMigrants += otherPos.Count
				otherKey := strconv.FormatFloat(otherPos.Latitude, 'f', 4, 64) + "," + strconv.FormatFloat(otherPos.Longitude, 'f', 4, 64)
				processed[otherKey] = true
			}
		}

		cluster.MigrantCount = totalMigrants
		cluster.DensityScore = float64(totalMigrants) / (math.Pi * radiusKm * radiusKm) // migrants par km²

		clusters = append(clusters, cluster)
		processed[key] = true
	}

	// Générer des données de heatmap
	var heatmapData []HeatmapPoint
	maxIntensity := float64(0)

	for _, pos := range positions {
		intensity := float64(pos.Count)
		if intensity > maxIntensity {
			maxIntensity = intensity
		}
	}

	for _, pos := range positions {
		normalizedIntensity := float64(pos.Count) / maxIntensity
		heatmapData = append(heatmapData, HeatmapPoint{
			Latitude:  pos.Latitude,
			Longitude: pos.Longitude,
			Intensity: normalizedIntensity,
			Count:     pos.Count,
		})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data": map[string]interface{}{
			"clusters":        clusters,
			"heatmap_data":    heatmapData,
			"analysis_radius": radiusKm,
			"analysis_period": daysPeriod,
			"total_clusters":  len(clusters),
			"max_intensity":   maxIntensity,
		},
	})
}

// Analyse des corridors de migration
func GetMigrationCorridors(c *fiber.Ctx) error {
	db := database.DB

	daysPeriod, _ := strconv.Atoi(c.Query("days", "60"))
	minFlow, _ := strconv.Atoi(c.Query("min_flow", "2"))

	// Identifier les corridors principaux entre villes/pays
	var corridors []struct {
		VilleOrigine     string  `json:"ville_origine"`
		PaysOrigine      string  `json:"pays_origine"`
		VilleDestination string  `json:"ville_destination"`
		PaysDestination  string  `json:"pays_destination"`
		FlowCount        int64   `json:"flow_count"`
		LatOrigine       float64 `json:"lat_origine"`
		LonOrigine       float64 `json:"lon_origine"`
		LatDestination   float64 `json:"lat_destination"`
		LonDestination   float64 `json:"lon_destination"`
		Distance         float64 `json:"distance_km"`
	}

	db.Raw(`
		WITH movement_pairs AS (
			SELECT DISTINCT
				g1.ville as ville_origine,
				g1.pays as pays_origine,
				g2.ville as ville_destination,
				g2.pays as pays_destination,
				g1.latitude as lat_origine,
				g1.longitude as lon_origine,
				g2.latitude as lat_destination,
				g2.longitude as lon_destination,
				g1.migrant_uuid
			FROM geolocalisations g1
			JOIN geolocalisations g2 ON g1.migrant_uuid = g2.migrant_uuid
			WHERE g1.created_at < g2.created_at
			AND g2.created_at >= NOW() - INTERVAL '%d days'
			AND g1.ville IS NOT NULL AND g2.ville IS NOT NULL
			AND g1.pays IS NOT NULL AND g2.pays IS NOT NULL
			AND g1.ville != g2.ville
			AND g1.deleted_at IS NULL AND g2.deleted_at IS NULL
		)
		SELECT 
			ville_origine,
			pays_origine,
			ville_destination,
			pays_destination,
			COUNT(*) as flow_count,
			AVG(lat_origine) as lat_origine,
			AVG(lon_origine) as lon_origine,
			AVG(lat_destination) as lat_destination,
			AVG(lon_destination) as lon_destination,
			0.0 as distance_km  -- Sera calculé côté Go
		FROM movement_pairs
		GROUP BY ville_origine, pays_origine, ville_destination, pays_destination
		HAVING COUNT(*) >= %d
		ORDER BY flow_count DESC
		LIMIT 50
	`, daysPeriod, minFlow).Scan(&corridors)

	// Calculer les distances pour chaque corridor
	for i := range corridors {
		corridors[i].Distance = calculateSpatialDistance(
			corridors[i].LatOrigine, corridors[i].LonOrigine,
			corridors[i].LatDestination, corridors[i].LonDestination,
		)
	}

	// Analyse des points de transit
	var transitPoints []struct {
		Ville             string  `json:"ville"`
		Pays              string  `json:"pays"`
		Latitude          float64 `json:"latitude"`
		Longitude         float64 `json:"longitude"`
		TransitCount      int64   `json:"transit_count"`
		UniqueRoutes      int64   `json:"unique_routes"`
		AvgStayDuration   float64 `json:"avg_stay_duration_hours"`
		ConnectivityScore float64 `json:"connectivity_score"`
	}

	db.Raw(`
		SELECT 
			ville,
			pays,
			AVG(latitude) as latitude,
			AVG(longitude) as longitude,
			COUNT(*) as transit_count,
			COUNT(DISTINCT migrant_uuid) as unique_routes,
			AVG(COALESCE(duree_sejour * 24, 24)) as avg_stay_duration,
			COUNT(DISTINCT migrant_uuid) * 1.0 / COUNT(*) as connectivity_score
		FROM geolocalisations
		WHERE created_at >= NOW() - INTERVAL '%d days'
		AND type_localisation IN ('transit', 'point_passage')
		AND ville IS NOT NULL
		AND pays IS NOT NULL
		AND deleted_at IS NULL
		GROUP BY ville, pays
		HAVING COUNT(DISTINCT migrant_uuid) >= %d
		ORDER BY connectivity_score DESC, transit_count DESC
		LIMIT 30
	`, daysPeriod, minFlow).Scan(&transitPoints)

	return c.JSON(fiber.Map{
		"status": "success",
		"data": map[string]interface{}{
			"corridors":          corridors,
			"transit_points":     transitPoints,
			"analysis_period":    daysPeriod,
			"min_flow_threshold": minFlow,
		},
	})
}

// Analyse de proximité géographique
func GetProximityAnalysis(c *fiber.Ctx) error {
	db := database.DB

	targetLat, _ := strconv.ParseFloat(c.Query("latitude"), 64)
	targetLon, _ := strconv.ParseFloat(c.Query("longitude"), 64)
	searchRadius, _ := strconv.ParseFloat(c.Query("radius", "50"), 64)
	daysPeriod, _ := strconv.Atoi(c.Query("days", "30"))

	if targetLat == 0 || targetLon == 0 {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Latitude and longitude are required",
		})
	}

	// Migrants dans le rayon spécifié
	var nearbyMigrants []struct {
		MigrantUUID      string    `json:"migrant_uuid"`
		MigrantNom       string    `json:"migrant_nom"`
		MigrantPrenom    string    `json:"migrant_prenom"`
		Latitude         float64   `json:"latitude"`
		Longitude        float64   `json:"longitude"`
		Distance         float64   `json:"distance_km"`
		Ville            string    `json:"ville"`
		Pays             string    `json:"pays"`
		LastSeen         time.Time `json:"last_seen"`
		TypeLocalisation string    `json:"type_localisation"`
	}

	// Utiliser la formule de Haversine en SQL pour le filtrage initial
	db.Raw(`
		SELECT DISTINCT ON (g.migrant_uuid)
			g.migrant_uuid,
			m.nom as migrant_nom,
			m.prenom as migrant_prenom,
			g.latitude,
			g.longitude,
			0.0 as distance_km,  -- Sera calculé côté Go
			g.ville,
			g.pays,
			g.created_at as last_seen,
			g.type_localisation
		FROM geolocalisations g
		JOIN migrants m ON g.migrant_uuid = m.uuid
		WHERE g.created_at >= NOW() - INTERVAL '%d days'
		AND g.latitude IS NOT NULL
		AND g.longitude IS NOT NULL
		AND g.deleted_at IS NULL
		AND m.deleted_at IS NULL
		-- Filtrage approximatif par bounding box pour optimiser
		AND g.latitude BETWEEN %f - %f/111.0 AND %f + %f/111.0
		AND g.longitude BETWEEN %f - %f/(111.0 * COS(RADIANS(%f))) AND %f + %f/(111.0 * COS(RADIANS(%f)))
		ORDER BY g.migrant_uuid, g.created_at DESC
	`, daysPeriod, targetLat, searchRadius, targetLat, searchRadius,
		targetLon, searchRadius, targetLat, targetLon, searchRadius, targetLat).Scan(&nearbyMigrants)

	// Calculer les distances exactes et filtrer
	var filteredMigrants []struct {
		MigrantUUID      string    `json:"migrant_uuid"`
		MigrantNom       string    `json:"migrant_nom"`
		MigrantPrenom    string    `json:"migrant_prenom"`
		Latitude         float64   `json:"latitude"`
		Longitude        float64   `json:"longitude"`
		Distance         float64   `json:"distance_km"`
		Ville            string    `json:"ville"`
		Pays             string    `json:"pays"`
		LastSeen         time.Time `json:"last_seen"`
		TypeLocalisation string    `json:"type_localisation"`
	}

	for _, migrant := range nearbyMigrants {
		distance := calculateSpatialDistance(targetLat, targetLon, migrant.Latitude, migrant.Longitude)
		if distance <= searchRadius {
			migrant.Distance = distance
			filteredMigrants = append(filteredMigrants, migrant)
		}
	}

	// Analyse de densité par zones concentriques
	zones := []float64{5, 10, 25, 50}
	var densityZones []map[string]interface{}

	for _, radius := range zones {
		count := 0
		for _, migrant := range filteredMigrants {
			if migrant.Distance <= radius {
				count++
			}
		}

		area := math.Pi * radius * radius
		density := float64(count) / area

		densityZones = append(densityZones, map[string]interface{}{
			"radius_km":     radius,
			"migrant_count": count,
			"area_km2":      area,
			"density":       density,
		})
	}

	// Alertes dans la zone
	var nearbyAlerts []models.Alert
	db.Preload("Migrant").
		Where(`
			latitude IS NOT NULL AND longitude IS NOT NULL
			AND created_at >= NOW() - INTERVAL '%d days'
			AND statut = 'active'
			AND deleted_at IS NULL
		`, daysPeriod).
		Find(&nearbyAlerts)

	var filteredAlerts []models.Alert
	for _, alert := range nearbyAlerts {
		if alert.Latitude != nil && alert.Longitude != nil {
			distance := calculateSpatialDistance(targetLat, targetLon, *alert.Latitude, *alert.Longitude)
			if distance <= searchRadius {
				filteredAlerts = append(filteredAlerts, alert)
			}
		}
	}

	// Statistiques de proximité
	proximityStats := map[string]interface{}{
		"search_center": map[string]float64{
			"latitude":  targetLat,
			"longitude": targetLon,
		},
		"search_radius_km": searchRadius,
		"total_migrants":   len(filteredMigrants),
		"total_alerts":     len(filteredAlerts),
		"analysis_period":  daysPeriod,
		"area_km2":         math.Pi * searchRadius * searchRadius,
		"density":          float64(len(filteredMigrants)) / (math.Pi * searchRadius * searchRadius),
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data": map[string]interface{}{
			"nearby_migrants": filteredMigrants,
			"nearby_alerts":   filteredAlerts,
			"density_zones":   densityZones,
			"proximity_stats": proximityStats,
		},
	})
}

// Analyse des zones d'intérêt (AOI - Areas of Interest)
func GetAreasOfInterest(c *fiber.Ctx) error {
	db := database.DB

	daysPeriod, _ := strconv.Atoi(c.Query("days", "30"))
	minActivity, _ := strconv.Atoi(c.Query("min_activity", "5"))

	// Zones avec forte activité
	var hotZones []struct {
		Ville           string  `json:"ville"`
		Pays            string  `json:"pays"`
		CenterLatitude  float64 `json:"center_latitude"`
		CenterLongitude float64 `json:"center_longitude"`
		ActivityCount   int64   `json:"activity_count"`
		UniqueMigrants  int64   `json:"unique_migrants"`
		AlertCount      int64   `json:"alert_count"`
		RiskScore       float64 `json:"risk_score"`
		ActivityType    string  `json:"activity_type"`
	}

	db.Raw(`
		SELECT 
			g.ville,
			g.pays,
			AVG(g.latitude) as center_latitude,
			AVG(g.longitude) as center_longitude,
			COUNT(g.uuid) as activity_count,
			COUNT(DISTINCT g.migrant_uuid) as unique_migrants,
			COUNT(DISTINCT a.uuid) as alert_count,
			(COUNT(DISTINCT a.uuid)::float / NULLIF(COUNT(DISTINCT g.migrant_uuid), 0)) * 100 as risk_score,
			CASE 
				WHEN g.type_localisation = 'frontiere' THEN 'border_zone'
				WHEN g.type_localisation IN ('point_passage', 'transit') THEN 'transit_zone'
				WHEN g.type_localisation = 'centre_accueil' THEN 'reception_center'
				ELSE 'general_area'
			END as activity_type
		FROM geolocalisations g
		LEFT JOIN alertes a ON g.migrant_uuid = a.migrant_uuid 
			AND a.created_at >= NOW() - INTERVAL '%d days'
			AND a.deleted_at IS NULL
		WHERE g.created_at >= NOW() - INTERVAL '%d days'
		AND g.ville IS NOT NULL
		AND g.pays IS NOT NULL
		AND g.deleted_at IS NULL
		GROUP BY g.ville, g.pays, g.type_localisation
		HAVING COUNT(g.uuid) >= %d
		ORDER BY activity_count DESC, risk_score DESC
		LIMIT 25
	`, daysPeriod, daysPeriod, minActivity).Scan(&hotZones)

	// Zones frontalières critiques
	var borderZones []struct {
		Ville              string  `json:"ville"`
		Pays               string  `json:"pays"`
		Latitude           float64 `json:"latitude"`
		Longitude          float64 `json:"longitude"`
		CrossingCount      int64   `json:"crossing_count"`
		UniqueMigrants     int64   `json:"unique_migrants"`
		AvgStayHours       float64 `json:"avg_stay_hours"`
		SecurityAlerts     int64   `json:"security_alerts"`
		HumanitarianAlerts int64   `json:"humanitarian_alerts"`
	}

	db.Raw(`
		SELECT 
			g.ville,
			g.pays,
			AVG(g.latitude) as latitude,
			AVG(g.longitude) as longitude,
			COUNT(g.uuid) as crossing_count,
			COUNT(DISTINCT g.migrant_uuid) as unique_migrants,
			AVG(COALESCE(g.duree_sejour * 24, 12)) as avg_stay_hours,
			COUNT(DISTINCT CASE WHEN a.type_alerte = 'securite' THEN a.uuid END) as security_alerts,
			COUNT(DISTINCT CASE WHEN a.type_alerte = 'humanitaire' THEN a.uuid END) as humanitarian_alerts
		FROM geolocalisations g
		LEFT JOIN alertes a ON g.migrant_uuid = a.migrant_uuid 
			AND a.created_at >= NOW() - INTERVAL '%d days'
			AND a.deleted_at IS NULL
		WHERE g.created_at >= NOW() - INTERVAL '%d days'
		AND g.type_localisation = 'frontiere'
		AND g.ville IS NOT NULL
		AND g.pays IS NOT NULL
		AND g.deleted_at IS NULL
		GROUP BY g.ville, g.pays
		HAVING COUNT(DISTINCT g.migrant_uuid) >= 2
		ORDER BY crossing_count DESC
		LIMIT 15
	`, daysPeriod, daysPeriod).Scan(&borderZones)

	// Centres d'accueil et leur capacité
	var receptionCenters []struct {
		Ville            string  `json:"ville"`
		Pays             string  `json:"pays"`
		Latitude         float64 `json:"latitude"`
		Longitude        float64 `json:"longitude"`
		CurrentOccupancy int64   `json:"current_occupancy"`
		TotalVisits      int64   `json:"total_visits"`
		AvgStayDays      float64 `json:"avg_stay_days"`
		CapacityStress   float64 `json:"capacity_stress"`
	}

	db.Raw(`
		SELECT 
			g.ville,
			g.pays,
			AVG(g.latitude) as latitude,
			AVG(g.longitude) as longitude,
			COUNT(DISTINCT CASE WHEN g.created_at >= NOW() - INTERVAL '7 days' THEN g.migrant_uuid END) as current_occupancy,
			COUNT(g.uuid) as total_visits,
			AVG(COALESCE(g.duree_sejour, 7)) as avg_stay_days,
			(COUNT(DISTINCT CASE WHEN g.created_at >= NOW() - INTERVAL '7 days' THEN g.migrant_uuid END)::float / 100.0) * 100 as capacity_stress
		FROM geolocalisations g
		WHERE g.created_at >= NOW() - INTERVAL '%d days'
		AND g.type_localisation = 'centre_accueil'
		AND g.ville IS NOT NULL
		AND g.pays IS NOT NULL
		AND g.deleted_at IS NULL
		GROUP BY g.ville, g.pays
		HAVING COUNT(g.uuid) >= 3
		ORDER BY current_occupancy DESC
		LIMIT 15
	`, daysPeriod).Scan(&receptionCenters)

	return c.JSON(fiber.Map{
		"status": "success",
		"data": map[string]interface{}{
			"hot_zones":         hotZones,
			"border_zones":      borderZones,
			"reception_centers": receptionCenters,
			"analysis_period":   daysPeriod,
			"min_activity":      minActivity,
		},
	})
}
