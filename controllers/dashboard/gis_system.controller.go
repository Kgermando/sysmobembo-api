package dashboard

import (
	"math"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/kgermando/sysmobembo-api/database"
)

// =======================
// SYSTÈME D'INFORMATION GÉOGRAPHIQUE (SIG)
// =======================

// Structure pour les couches GIS
type GISLayer struct {
	LayerID     string                 `json:"layer_id"`
	LayerName   string                 `json:"layer_name"`
	LayerType   string                 `json:"layer_type"`
	Visible     bool                   `json:"visible"`
	Data        interface{}            `json:"data"`
	StyleConfig map[string]interface{} `json:"style_config"`
}

type GISFeature struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	Geometry   map[string]interface{} `json:"geometry"`
	Properties map[string]interface{} `json:"properties"`
}

// Configuration de la carte principale du SIG
func GetGISMapConfiguration(c *fiber.Ctx) error {
	// Configuration de base pour la carte
	mapConfig := map[string]interface{}{
		"default_center": map[string]float64{
			"latitude":  0.0,  // Centre par défaut (à adapter selon la région)
			"longitude": 25.0, // Centre Afrique centrale
		},
		"default_zoom": 6,
		"min_zoom":     2,
		"max_zoom":     18,
		"map_style":    "satellite", // ou "street", "terrain"
		"projection":   "EPSG:4326", // WGS84
		"bounds": map[string]interface{}{
			"north": 20.0,
			"south": -20.0,
			"east":  40.0,
			"west":  10.0,
		},
	}

	// Couches disponibles
	availableLayers := []map[string]interface{}{
		{
			"layer_id":   "migrants_current",
			"layer_name": "Positions actuelles des migrants",
			"layer_type": "point",
			"category":   "migrants",
			"visible":    true,
			"style": map[string]interface{}{
				"color":       "#FF5722",
				"radius":      8,
				"stroke":      true,
				"strokeColor": "#FFFFFF",
				"strokeWidth": 2,
			},
		},
		{
			"layer_id":   "migration_routes",
			"layer_name": "Routes de migration",
			"layer_type": "line",
			"category":   "movement",
			"visible":    true,
			"style": map[string]interface{}{
				"color":   "#2196F3",
				"weight":  3,
				"opacity": 0.7,
			},
		},
		{
			"layer_id":   "alert_zones",
			"layer_name": "Zones d'alerte",
			"layer_type": "polygon",
			"category":   "security",
			"visible":    true,
			"style": map[string]interface{}{
				"fillColor":   "#F44336",
				"fillOpacity": 0.3,
				"stroke":      true,
				"color":       "#F44336",
				"weight":      2,
			},
		},
		{
			"layer_id":   "border_points",
			"layer_name": "Points frontaliers",
			"layer_type": "point",
			"category":   "infrastructure",
			"visible":    true,
			"style": map[string]interface{}{
				"color":  "#9C27B0",
				"radius": 10,
				"icon":   "border-crossing",
			},
		},
		{
			"layer_id":   "reception_centers",
			"layer_name": "Centres d'accueil",
			"layer_type": "point",
			"category":   "infrastructure",
			"visible":    true,
			"style": map[string]interface{}{
				"color":  "#4CAF50",
				"radius": 12,
				"icon":   "home",
			},
		},
		{
			"layer_id":   "density_heatmap",
			"layer_name": "Carte de densité",
			"layer_type": "heatmap",
			"category":   "analysis",
			"visible":    false,
			"style": map[string]interface{}{
				"radius":     25,
				"blur":       15,
				"maxOpacity": 0.8,
				"gradient": map[string]string{
					"0.0": "blue",
					"0.5": "yellow",
					"1.0": "red",
				},
			},
		},
	}

	// Outils disponibles
	mapTools := []map[string]interface{}{
		{
			"tool_id":   "measure_distance",
			"tool_name": "Mesurer distance",
			"icon":      "ruler",
			"enabled":   true,
		},
		{
			"tool_id":   "measure_area",
			"tool_name": "Mesurer surface",
			"icon":      "square",
			"enabled":   true,
		},
		{
			"tool_id":   "draw_polygon",
			"tool_name": "Dessiner zone",
			"icon":      "polygon",
			"enabled":   true,
		},
		{
			"tool_id":   "identify",
			"tool_name": "Identifier",
			"icon":      "info-circle",
			"enabled":   true,
		},
		{
			"tool_id":   "export_map",
			"tool_name": "Exporter carte",
			"icon":      "download",
			"enabled":   true,
		},
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data": map[string]interface{}{
			"map_config":       mapConfig,
			"available_layers": availableLayers,
			"map_tools":        mapTools,
			"last_updated":     time.Now(),
		},
	})
}

// Données pour la couche des migrants actuels
func GetMigrantsLayer(c *fiber.Ctx) error {
	db := database.DB

	// Récupérer les positions les plus récentes de chaque migrant
	var migrantLocations []struct {
		MigrantUUID   string    `json:"migrant_uuid"`
		MigrantNom    string    `json:"migrant_nom"`
		MigrantPrenom string    `json:"migrant_prenom"`
		Latitude      float64   `json:"latitude"`
		Longitude     float64   `json:"longitude"`
		Ville         string    `json:"ville"`
		Pays          string    `json:"pays"`
		LastUpdate    time.Time `json:"last_update"`
		StatutMigrant string    `json:"statut_migrant"`
		AlerteActive  bool      `json:"alerte_active"`
		TypeLocation  string    `json:"type_location"`
	}

	db.Raw(`
		SELECT DISTINCT ON (g.migrant_uuid)
			g.migrant_uuid,
			m.nom as migrant_nom,
			m.prenom as migrant_prenom,
			g.latitude,
			g.longitude,
			g.ville,
			g.pays,
			g.created_at as last_update,
			m.statut_migratoire as statut_migrant,
			EXISTS(SELECT 1 FROM alertes a WHERE a.migrant_uuid = g.migrant_uuid AND a.statut = 'active') as alerte_active,
			g.type_localisation as type_location
		FROM geolocalisations g
		JOIN migrants m ON g.migrant_uuid = m.uuid
		WHERE g.latitude IS NOT NULL
		AND g.longitude IS NOT NULL
		AND g.deleted_at IS NULL
		AND m.deleted_at IS NULL
		ORDER BY g.migrant_uuid, g.created_at DESC
	`).Scan(&migrantLocations)

	// Convertir en format GeoJSON
	features := make([]GISFeature, 0)
	for _, location := range migrantLocations {
		feature := GISFeature{
			ID:   location.MigrantUUID,
			Type: "Feature",
			Geometry: map[string]interface{}{
				"type":        "Point",
				"coordinates": []float64{location.Longitude, location.Latitude},
			},
			Properties: map[string]interface{}{
				"migrant_uuid":   location.MigrantUUID,
				"migrant_nom":    location.MigrantNom,
				"migrant_prenom": location.MigrantPrenom,
				"ville":          location.Ville,
				"pays":           location.Pays,
				"last_update":    location.LastUpdate,
				"statut_migrant": location.StatutMigrant,
				"alerte_active":  location.AlerteActive,
				"type_location":  location.TypeLocation,
				"popup_content":  location.MigrantNom + " " + location.MigrantPrenom + " - " + location.Ville,
			},
		}
		features = append(features, feature)
	}

	geoJsonData := map[string]interface{}{
		"type":     "FeatureCollection",
		"features": features,
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data": map[string]interface{}{
			"layer_id":   "migrants_current",
			"layer_type": "point",
			"data":       geoJsonData,
			"count":      len(features),
			"timestamp":  time.Now(),
		},
	})
}

// Données pour la couche des routes de migration
func GetMigrationRoutesLayer(c *fiber.Ctx) error {
	db := database.DB

	daysPeriod, _ := strconv.Atoi(c.Query("days", "30"))
	minFlow, _ := strconv.Atoi(c.Query("min_flow", "3"))

	// Récupérer les routes de migration principales
	var routes []struct {
		RouteID         string  `json:"route_id"`
		PaysOrigine     string  `json:"pays_origine"`
		PaysDestination string  `json:"pays_destination"`
		LatOrigine      float64 `json:"lat_origine"`
		LonOrigine      float64 `json:"lon_origine"`
		LatDestination  float64 `json:"lat_destination"`
		LonDestination  float64 `json:"lon_destination"`
		FlowCount       int64   `json:"flow_count"`
		RouteType       string  `json:"route_type"`
	}

	db.Raw(`
		WITH route_analysis AS (
			SELECT 
				CONCAT(m.pays_origine, '-', m.pays_destination) as route_id,
				m.pays_origine,
				m.pays_destination,
				COUNT(DISTINCT m.uuid) as flow_count,
				'main_route' as route_type
			FROM migrants m
			JOIN geolocalisations g ON m.uuid = g.migrant_uuid
			WHERE g.created_at >= NOW() - INTERVAL '%d days'
			AND m.pays_origine IS NOT NULL
			AND m.pays_destination IS NOT NULL
			AND m.pays_origine != m.pays_destination
			AND g.deleted_at IS NULL
			AND m.deleted_at IS NULL
			GROUP BY m.pays_origine, m.pays_destination
			HAVING COUNT(DISTINCT m.uuid) >= %d
		)
		SELECT 
			route_id,
			pays_origine,
			pays_destination,
			-- Coordonnées simulées (à remplacer par vraies coordonnées des pays)
			CASE 
				WHEN pays_origine = 'RDC' THEN -4.038333
				WHEN pays_origine = 'Angola' THEN -11.202692
				WHEN pays_origine = 'Cameroun' THEN 7.369722
				ELSE 0.0
			END as lat_origine,
			CASE 
				WHEN pays_origine = 'RDC' THEN 21.758664
				WHEN pays_origine = 'Angola' THEN 17.873887
				WHEN pays_origine = 'Cameroun' THEN 12.354722
				ELSE 25.0
			END as lon_origine,
			CASE 
				WHEN pays_destination = 'RDC' THEN -4.038333
				WHEN pays_destination = 'Angola' THEN -11.202692
				WHEN pays_destination = 'Cameroun' THEN 7.369722
				ELSE 0.0
			END as lat_destination,
			CASE 
				WHEN pays_destination = 'RDC' THEN 21.758664
				WHEN pays_destination = 'Angola' THEN 17.873887
				WHEN pays_destination = 'Cameroun' THEN 12.354722
				ELSE 25.0
			END as lon_destination,
			flow_count,
			route_type
		FROM route_analysis
		ORDER BY flow_count DESC
	`, daysPeriod, minFlow).Scan(&routes)

	// Convertir en format GeoJSON LineString
	features := make([]GISFeature, 0)
	for _, route := range routes {
		feature := GISFeature{
			ID:   route.RouteID,
			Type: "Feature",
			Geometry: map[string]interface{}{
				"type": "LineString",
				"coordinates": [][]float64{
					{route.LonOrigine, route.LatOrigine},
					{route.LonDestination, route.LatDestination},
				},
			},
			Properties: map[string]interface{}{
				"route_id":         route.RouteID,
				"pays_origine":     route.PaysOrigine,
				"pays_destination": route.PaysDestination,
				"flow_count":       route.FlowCount,
				"route_type":       route.RouteType,
				"popup_content":    route.PaysOrigine + " → " + route.PaysDestination + " (" + strconv.FormatInt(route.FlowCount, 10) + " migrants)",
			},
		}
		features = append(features, feature)
	}

	geoJsonData := map[string]interface{}{
		"type":     "FeatureCollection",
		"features": features,
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data": map[string]interface{}{
			"layer_id":        "migration_routes",
			"layer_type":      "line",
			"data":            geoJsonData,
			"count":           len(features),
			"analysis_period": daysPeriod,
			"timestamp":       time.Now(),
		},
	})
}

// Données pour la couche des zones d'alerte
func GetAlertZonesLayer(c *fiber.Ctx) error {
	db := database.DB

	// Récupérer les alertes avec géolocalisation
	var alertZones []struct {
		AlertUUID     string    `json:"alert_uuid"`
		TypeAlerte    string    `json:"type_alerte"`
		NiveauGravite string    `json:"niveau_gravite"`
		Latitude      float64   `json:"latitude"`
		Longitude     float64   `json:"longitude"`
		Titre         string    `json:"titre"`
		Description   string    `json:"description"`
		DateCreation  time.Time `json:"date_creation"`
		Statut        string    `json:"statut"`
		Adresse       string    `json:"adresse"`
	}

	db.Raw(`
		SELECT 
			a.uuid as alert_uuid,
			a.type_alerte,
			a.niveau_gravite,
			COALESCE(a.latitude, g.latitude) as latitude,
			COALESCE(a.longitude, g.longitude) as longitude,
			a.titre,
			a.description,
			a.created_at as date_creation,
			a.statut,
			COALESCE(a.adresse, g.adresse) as adresse
		FROM alertes a
		LEFT JOIN geolocalisations g ON a.migrant_uuid = g.migrant_uuid
		WHERE a.statut = 'active'
		AND (a.latitude IS NOT NULL OR g.latitude IS NOT NULL)
		AND (a.longitude IS NOT NULL OR g.longitude IS NOT NULL)
		AND a.deleted_at IS NULL
		ORDER BY a.created_at DESC
	`).Scan(&alertZones)

	// Créer des zones circulaires autour des alertes
	features := make([]GISFeature, 0)
	for _, alert := range alertZones {
		// Rayon basé sur la gravité
		radius := 1000.0 // mètres
		switch alert.NiveauGravite {
		case "critical":
			radius = 5000.0
		case "danger":
			radius = 3000.0
		case "warning":
			radius = 1500.0
		default:
			radius = 1000.0
		}

		// Créer un cercle approximatif avec des points
		points := make([][]float64, 0)
		numPoints := 20
		for i := 0; i < numPoints; i++ {
			angle := float64(i) * 2.0 * 3.14159 / float64(numPoints)
			// Approximation simple pour créer un cercle
			deltaLat := (radius / 111000.0) * math.Cos(angle)
			deltaLon := (radius / (111000.0 * math.Cos(alert.Latitude*3.14159/180.0))) * math.Sin(angle)

			points = append(points, []float64{
				alert.Longitude + deltaLon,
				alert.Latitude + deltaLat,
			})
		}
		// Fermer le polygone
		points = append(points, points[0])

		feature := GISFeature{
			ID:   alert.AlertUUID,
			Type: "Feature",
			Geometry: map[string]interface{}{
				"type":        "Polygon",
				"coordinates": [][][]float64{points},
			},
			Properties: map[string]interface{}{
				"alert_uuid":     alert.AlertUUID,
				"type_alerte":    alert.TypeAlerte,
				"niveau_gravite": alert.NiveauGravite,
				"titre":          alert.Titre,
				"description":    alert.Description,
				"date_creation":  alert.DateCreation,
				"statut":         alert.Statut,
				"adresse":        alert.Adresse,
				"radius_meters":  radius,
				"popup_content":  alert.Titre + " (" + alert.NiveauGravite + ")",
			},
		}
		features = append(features, feature)
	}

	geoJsonData := map[string]interface{}{
		"type":     "FeatureCollection",
		"features": features,
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data": map[string]interface{}{
			"layer_id":   "alert_zones",
			"layer_type": "polygon",
			"data":       geoJsonData,
			"count":      len(features),
			"timestamp":  time.Now(),
		},
	})
}

// Données pour la couche des infrastructures (frontières, centres d'accueil)
func GetInfrastructureLayer(c *fiber.Ctx) error {
	db := database.DB

	layerType := c.Query("type", "all") // "border_points", "reception_centers", "all"

	var infraPoints []struct {
		PointID       string  `json:"point_id"`
		PointType     string  `json:"point_type"`
		Nom           string  `json:"nom"`
		Ville         string  `json:"ville"`
		Pays          string  `json:"pays"`
		Latitude      float64 `json:"latitude"`
		Longitude     float64 `json:"longitude"`
		ActivityCount int64   `json:"activity_count"`
		Status        string  `json:"status"`
		Description   string  `json:"description"`
	}

	query := `
		SELECT 
			CONCAT(type_localisation, '_', ville, '_', pays) as point_id,
			type_localisation as point_type,
			ville as nom,
			ville,
			pays,
			AVG(latitude) as latitude,
			AVG(longitude) as longitude,
			COUNT(*) as activity_count,
			'active' as status,
			CONCAT('Point ', type_localisation, ' - ', COUNT(DISTINCT migrant_uuid), ' migrants') as description
		FROM geolocalisations
		WHERE type_localisation IN ('frontiere', 'centre_accueil', 'point_passage')
		AND ville IS NOT NULL
		AND pays IS NOT NULL
		AND latitude IS NOT NULL
		AND longitude IS NOT NULL
		AND deleted_at IS NULL
	`

	if layerType != "all" {
		if layerType == "border_points" {
			query += " AND type_localisation IN ('frontiere', 'point_passage')"
		} else if layerType == "reception_centers" {
			query += " AND type_localisation = 'centre_accueil'"
		}
	}

	query += `
		GROUP BY type_localisation, ville, pays
		ORDER BY activity_count DESC
	`

	db.Raw(query).Scan(&infraPoints)

	// Convertir en format GeoJSON
	features := make([]GISFeature, 0)
	for _, point := range infraPoints {
		feature := GISFeature{
			ID:   point.PointID,
			Type: "Feature",
			Geometry: map[string]interface{}{
				"type":        "Point",
				"coordinates": []float64{point.Longitude, point.Latitude},
			},
			Properties: map[string]interface{}{
				"point_id":       point.PointID,
				"point_type":     point.PointType,
				"nom":            point.Nom,
				"ville":          point.Ville,
				"pays":           point.Pays,
				"activity_count": point.ActivityCount,
				"status":         point.Status,
				"description":    point.Description,
				"popup_content":  point.Nom + " - " + point.Description,
			},
		}
		features = append(features, feature)
	}

	geoJsonData := map[string]interface{}{
		"type":     "FeatureCollection",
		"features": features,
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data": map[string]interface{}{
			"layer_id":   "infrastructure_" + layerType,
			"layer_type": "point",
			"data":       geoJsonData,
			"count":      len(features),
			"filter":     layerType,
			"timestamp":  time.Now(),
		},
	})
}

// Données pour la heatmap de densité
func GetDensityHeatmapLayer(c *fiber.Ctx) error {
	db := database.DB

	daysPeriod, _ := strconv.Atoi(c.Query("days", "30"))
	intensity := c.Query("intensity", "migrant_count") // "migrant_count", "alert_count", "movement_count"

	// Récupérer les points de densité
	var densityPoints []struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		Weight    float64 `json:"weight"`
	}

	var query string
	switch intensity {
	case "alert_count":
		query = `
			SELECT 
				COALESCE(a.latitude, g.latitude) as latitude,
				COALESCE(a.longitude, g.longitude) as longitude,
				COUNT(a.uuid)::float as weight
			FROM alertes a
			LEFT JOIN geolocalisations g ON a.migrant_uuid = g.migrant_uuid
			WHERE a.created_at >= NOW() - INTERVAL '%d days'
			AND (a.latitude IS NOT NULL OR g.latitude IS NOT NULL)
			AND (a.longitude IS NOT NULL OR g.longitude IS NOT NULL)
			AND a.deleted_at IS NULL
			GROUP BY COALESCE(a.latitude, g.latitude), COALESCE(a.longitude, g.longitude)
		`
	case "movement_count":
		query = `
			SELECT 
				latitude,
				longitude,
				COUNT(*)::float as weight
			FROM geolocalisations
			WHERE created_at >= NOW() - INTERVAL '%d days'
			AND latitude IS NOT NULL
			AND longitude IS NOT NULL
			AND deleted_at IS NULL
			GROUP BY latitude, longitude
		`
	default: // migrant_count
		query = `
			SELECT 
				latitude,
				longitude,
				COUNT(DISTINCT migrant_uuid)::float as weight
			FROM geolocalisations
			WHERE created_at >= NOW() - INTERVAL '%d days'
			AND latitude IS NOT NULL
			AND longitude IS NOT NULL
			AND deleted_at IS NULL
			GROUP BY latitude, longitude
		`
	}

	db.Raw(query, daysPeriod).Scan(&densityPoints)

	// Normaliser les poids (0-1)
	maxWeight := 0.0
	for _, point := range densityPoints {
		if point.Weight > maxWeight {
			maxWeight = point.Weight
		}
	}

	for i := range densityPoints {
		densityPoints[i].Weight = densityPoints[i].Weight / maxWeight
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data": map[string]interface{}{
			"layer_id":        "density_heatmap",
			"layer_type":      "heatmap",
			"data":            densityPoints,
			"count":           len(densityPoints),
			"intensity_type":  intensity,
			"analysis_period": daysPeriod,
			"max_weight":      maxWeight,
			"timestamp":       time.Now(),
		},
	})
}

// Export des données SIG en différents formats
func ExportGISData(c *fiber.Ctx) error {
	format := c.Query("format", "geojson") // "geojson", "kml", "shapefile"
	layers := c.Query("layers", "all")     // comma-separated layer IDs

	if format != "geojson" {
		return c.Status(501).JSON(fiber.Map{
			"status":  "error",
			"message": "Format non supporté pour le moment. Utilisez 'geojson'",
		})
	}

	// Pour la démo, retourner un exemple d'export
	exportData := map[string]interface{}{
		"export_format": format,
		"export_layers": layers,
		"export_date":   time.Now(),
		"download_url":  "/api/gis/download/" + strconv.FormatInt(time.Now().Unix(), 10),
		"file_size_mb":  2.5,
		"status":        "ready",
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   exportData,
	})
}
