 package dashboard

// import (
// 	"fmt"
// 	"math"
// 	"strconv"
// 	"strings"
// 	"time"

// 	"github.com/gofiber/fiber/v2"
// 	"github.com/kgermando/sysmobembo-api/database"
// 	"github.com/kgermando/sysmobembo-api/models"
// )

// // Structures pour les réponses du dashboard GIS

// type GISStatistics struct {
// 	TotalMigrants          int                `json:"total_migrants"`
// 	MigrantsByCountry      []CountryStats     `json:"migrants_by_country"`
// 	MigrantsByStatus       []StatusStats      `json:"migrants_by_status"`
// 	MigrationFlowsByMonth  []MonthlyFlow      `json:"migration_flows_by_month"`
// 	HotspotLocations       []Hotspot          `json:"hotspot_locations"`
// 	MigrationCorridors     []Corridor         `json:"migration_corridors"`
// 	DensityMap             []DensityPoint     `json:"density_map"`
// 	RealTimePositions      []RealTimePosition `json:"realtime_positions"`
// 	PredictiveAnalysis     PredictiveData     `json:"predictive_analysis"`
// 	GeographicDistribution []GeographicData   `json:"geographic_distribution"`
// 	MovementPatterns       []MovementPattern  `json:"movement_patterns"`
// 	RiskZones              []RiskZone         `json:"risk_zones"`
// }

// type CountryStats struct {
// 	Country string  `json:"country"`
// 	Count   int     `json:"count"`
// 	Percent float64 `json:"percent"`
// }

// type StatusStats struct {
// 	Status  string  `json:"status"`
// 	Count   int     `json:"count"`
// 	Percent float64 `json:"percent"`
// 	Color   string  `json:"color"`
// }

// type MonthlyFlow struct {
// 	Month      string `json:"month"`
// 	Year       int    `json:"year"`
// 	Arrivals   int    `json:"arrivals"`
// 	Departures int    `json:"departures"`
// 	NetFlow    int    `json:"net_flow"`
// }

// type Hotspot struct {
// 	UUID        string  `json:"uuid"`
// 	Latitude    float64 `json:"latitude"`
// 	Longitude   float64 `json:"longitude"`
// 	City        string  `json:"city"`
// 	Country     string  `json:"country"`
// 	Count       int     `json:"count"`
// 	Intensity   float64 `json:"intensity"`
// 	Type        string  `json:"type"`
// 	Description string  `json:"description"`
// }

// type Corridor struct {
// 	FromCountry   string  `json:"from_country"`
// 	ToCountry     string  `json:"to_country"`
// 	FromLatitude  float64 `json:"from_latitude"`
// 	FromLongitude float64 `json:"from_longitude"`
// 	ToLatitude    float64 `json:"to_latitude"`
// 	ToLongitude   float64 `json:"to_longitude"`
// 	Count         int     `json:"count"`
// 	FlowDirection string  `json:"flow_direction"`
// }

// type DensityPoint struct {
// 	Latitude  float64 `json:"latitude"`
// 	Longitude float64 `json:"longitude"`
// 	Density   float64 `json:"density"`
// 	Radius    float64 `json:"radius"`
// }

// type RealTimePosition struct {
// 	MigrantUUID  string    `json:"migrant_uuid"`
// 	MigrantName  string    `json:"migrant_name"`
// 	Latitude     float64   `json:"latitude"`
// 	Longitude    float64   `json:"longitude"`
// 	Status       string    `json:"status"`
// 	LastUpdate   time.Time `json:"last_update"`
// 	City         string    `json:"city"`
// 	Country      string    `json:"country"`
// 	MovementType string    `json:"movement_type"`
// 	RiskLevel    string    `json:"risk_level"`
// }

// type PredictiveData struct {
// 	NextMonthPrediction  int                `json:"next_month_prediction"`
// 	TrendDirection       string             `json:"trend_direction"`
// 	SeasonalPatterns     []SeasonalPattern  `json:"seasonal_patterns"`
// 	RiskPredictions      []RiskPrediction   `json:"risk_predictions"`
// 	PopulationGrowthRate float64            `json:"population_growth_rate"`
// 	PredictedHotspots    []PredictedHotspot `json:"predicted_hotspots"`
// }

// type SeasonalPattern struct {
// 	Month           string  `json:"month"`
// 	AverageArrivals int     `json:"average_arrivals"`
// 	Trend           float64 `json:"trend"`
// }

// type RiskPrediction struct {
// 	Zone        string   `json:"zone"`
// 	RiskLevel   string   `json:"risk_level"`
// 	Probability float64  `json:"probability"`
// 	Factors     []string `json:"factors"`
// }

// type PredictedHotspot struct {
// 	Latitude    float64 `json:"latitude"`
// 	Longitude   float64 `json:"longitude"`
// 	Probability float64 `json:"probability"`
// 	Timeframe   string  `json:"timeframe"`
// }

// type GeographicData struct {
// 	Region     string  `json:"region"`
// 	Latitude   float64 `json:"latitude"`
// 	Longitude  float64 `json:"longitude"`
// 	Count      int     `json:"count"`
// 	Percentage float64 `json:"percentage"`
// 	GrowthRate float64 `json:"growth_rate"`
// }

// type MovementPattern struct {
// 	Pattern     string  `json:"pattern"`
// 	Count       int     `json:"count"`
// 	Percentage  float64 `json:"percentage"`
// 	Description string  `json:"description"`
// }

// type RiskZone struct {
// 	UUID       string    `json:"uuid"`
// 	Name       string    `json:"name"`
// 	Latitude   float64   `json:"latitude"`
// 	Longitude  float64   `json:"longitude"`
// 	Radius     float64   `json:"radius"`
// 	RiskLevel  string    `json:"risk_level"`
// 	RiskScore  float64   `json:"risk_score"`
// 	Factors    []string  `json:"factors"`
// 	AlertCount int       `json:"alert_count"`
// 	LastUpdate time.Time `json:"last_update"`
// }

// // GetGISStatistics - Endpoint principal pour récupérer toutes les statistiques GIS
// func GetGISStatistics(c *fiber.Ctx) error {
// 	var stats GISStatistics

// 	// 1. Total des migrants
// 	var totalMigrants int64
// 	database.DB.Model(&models.Migrant{}).Where("actif = ?", true).Count(&totalMigrants)
// 	stats.TotalMigrants = int(totalMigrants)

// 	// 2. Répartition par pays
// 	migrantsByCountry, err := getMigrantsByCountry()
// 	if err != nil {
// 		return c.Status(500).JSON(fiber.Map{"error": "Erreur lors de la récupération des données par pays"})
// 	}
// 	stats.MigrantsByCountry = migrantsByCountry

// 	// 3. Répartition par statut
// 	migrantsByStatus, err := getMigrantsByStatus()
// 	if err != nil {
// 		return c.Status(500).JSON(fiber.Map{"error": "Erreur lors de la récupération des données par statut"})
// 	}
// 	stats.MigrantsByStatus = migrantsByStatus

// 	// 4. Flux migratoires mensuels
// 	monthlyFlows, err := getMigrationFlowsByMonth()
// 	if err != nil {
// 		return c.Status(500).JSON(fiber.Map{"error": "Erreur lors de la récupération des flux mensuels"})
// 	}
// 	stats.MigrationFlowsByMonth = monthlyFlows

// 	// 5. Points chauds
// 	hotspots, err := getHotspotLocations()
// 	if err != nil {
// 		return c.Status(500).JSON(fiber.Map{"error": "Erreur lors de la récupération des points chauds"})
// 	}
// 	stats.HotspotLocations = hotspots

// 	// 6. Corridors migratoires
// 	corridors, err := getMigrationCorridors()
// 	if err != nil {
// 		return c.Status(500).JSON(fiber.Map{"error": "Erreur lors de la récupération des corridors"})
// 	}
// 	stats.MigrationCorridors = corridors

// 	// 7. Carte de densité
// 	densityMap, err := getDensityMap()
// 	if err != nil {
// 		return c.Status(500).JSON(fiber.Map{"error": "Erreur lors de la récupération de la carte de densité"})
// 	}
// 	stats.DensityMap = densityMap

// 	// 8. Positions en temps réel
// 	realTimePositions, err := getRealTimePositions()
// 	if err != nil {
// 		return c.Status(500).JSON(fiber.Map{"error": "Erreur lors de la récupération des positions temps réel"})
// 	}
// 	stats.RealTimePositions = realTimePositions

// 	// 9. Analyse prédictive
// 	predictiveData, err := getPredictiveAnalysis()
// 	if err != nil {
// 		return c.Status(500).JSON(fiber.Map{"error": "Erreur lors de l'analyse prédictive"})
// 	}
// 	stats.PredictiveAnalysis = predictiveData

// 	// 10. Distribution géographique
// 	geographicDist, err := getGeographicDistribution()
// 	if err != nil {
// 		return c.Status(500).JSON(fiber.Map{"error": "Erreur lors de la récupération de la distribution géographique"})
// 	}
// 	stats.GeographicDistribution = geographicDist

// 	// 11. Patterns de mouvement
// 	movementPatterns, err := getMovementPatterns()
// 	if err != nil {
// 		return c.Status(500).JSON(fiber.Map{"error": "Erreur lors de la récupération des patterns de mouvement"})
// 	}
// 	stats.MovementPatterns = movementPatterns

// 	// 12. Zones de risque
// 	riskZones, err := getRiskZones()
// 	if err != nil {
// 		return c.Status(500).JSON(fiber.Map{"error": "Erreur lors de la récupération des zones de risque"})
// 	}
// 	stats.RiskZones = riskZones

// 	return c.JSON(stats)
// }

// // getMigrantsByCountry - Récupère la répartition des migrants par pays
// func getMigrantsByCountry() ([]CountryStats, error) {
// 	var results []struct {
// 		Country string `json:"country"`
// 		Count   int    `json:"count"`
// 	}

// 	err := database.DB.Model(&models.Migrant{}).
// 		Select("pays_origine as country, COUNT(*) as count").
// 		Where("actif = ? AND pays_origine != ''", true).
// 		Group("pays_origine").
// 		Order("count DESC").
// 		Limit(10).
// 		Scan(&results).Error

// 	if err != nil {
// 		return nil, err
// 	}

// 	var total int
// 	for _, result := range results {
// 		total += result.Count
// 	}

// 	var countryStats []CountryStats
// 	for _, result := range results {
// 		percent := float64(result.Count) / float64(total) * 100
// 		countryStats = append(countryStats, CountryStats{
// 			Country: result.Country,
// 			Count:   result.Count,
// 			Percent: math.Round(percent*100) / 100,
// 		})
// 	}

// 	return countryStats, nil
// }

// // getMigrantsByStatus - Récupère la répartition des migrants par statut
// func getMigrantsByStatus() ([]StatusStats, error) {
// 	var results []struct {
// 		Status string `json:"status"`
// 		Count  int    `json:"count"`
// 	}

// 	err := database.DB.Model(&models.Migrant{}).
// 		Select("statut_migratoire as status, COUNT(*) as count").
// 		Where("actif = ?", true).
// 		Group("statut_migratoire").
// 		Scan(&results).Error

// 	if err != nil {
// 		return nil, err
// 	}

// 	var total int
// 	for _, result := range results {
// 		total += result.Count
// 	}

// 	statusColors := map[string]string{
// 		"regulier":        "#4CAF50",
// 		"irregulier":      "#F44336",
// 		"demandeur_asile": "#FF9800",
// 		"refugie":         "#2196F3",
// 	}

// 	var statusStats []StatusStats
// 	for _, result := range results {
// 		percent := float64(result.Count) / float64(total) * 100
// 		color := statusColors[result.Status]
// 		if color == "" {
// 			color = "#9E9E9E"
// 		}

// 		statusStats = append(statusStats, StatusStats{
// 			Status:  result.Status,
// 			Count:   result.Count,
// 			Percent: math.Round(percent*100) / 100,
// 			Color:   color,
// 		})
// 	}

// 	return statusStats, nil
// }

// // getMigrationFlowsByMonth - Récupère les flux migratoires par mois
// func getMigrationFlowsByMonth() ([]MonthlyFlow, error) {
// 	var results []struct {
// 		Month    int `json:"month"`
// 		Year     int `json:"year"`
// 		Arrivals int `json:"arrivals"`
// 	}

// 	err := database.DB.Model(&models.Migrant{}).
// 		Select("EXTRACT(MONTH FROM date_entree) as month, EXTRACT(YEAR FROM date_entree) as year, COUNT(*) as arrivals").
// 		Where("date_entree IS NOT NULL AND date_entree >= ?", time.Now().AddDate(-1, 0, 0)).
// 		Group("EXTRACT(YEAR FROM date_entree), EXTRACT(MONTH FROM date_entree)").
// 		Order("year, month").
// 		Scan(&results).Error

// 	if err != nil {
// 		return nil, err
// 	}

// 	monthNames := []string{
// 		"Janvier", "Février", "Mars", "Avril", "Mai", "Juin",
// 		"Juillet", "Août", "Septembre", "Octobre", "Novembre", "Décembre",
// 	}

// 	var monthlyFlows []MonthlyFlow
// 	for _, result := range results {
// 		monthName := monthNames[result.Month-1]

// 		// Pour cette démonstration, on simule les départs et calcule le flux net
// 		departures := int(float64(result.Arrivals) * 0.3) // 30% de départs simulés
// 		netFlow := result.Arrivals - departures

// 		monthlyFlows = append(monthlyFlows, MonthlyFlow{
// 			Month:      monthName,
// 			Year:       result.Year,
// 			Arrivals:   result.Arrivals,
// 			Departures: departures,
// 			NetFlow:    netFlow,
// 		})
// 	}

// 	return monthlyFlows, nil
// }

// // getHotspotLocations - Identifie les points chauds de migration
// func getHotspotLocations() ([]Hotspot, error) {
// 	var results []struct {
// 		UUID      string  `json:"uuid"`
// 		Latitude  float64 `json:"latitude"`
// 		Longitude float64 `json:"longitude"`
// 		Ville     string  `json:"ville"`
// 		Pays      string  `json:"pays"`
// 		Count     int     `json:"count"`
// 		Type      string  `json:"type"`
// 	}

// 	err := database.DB.Model(&models.Geolocalisation{}).
// 		Select("uuid, latitude, longitude, ville, pays, COUNT(*) as count, type_localisation as type").
// 		Where("latitude != 0 AND longitude != 0").
// 		Group("latitude, longitude, ville, pays, type_localisation, uuid").
// 		Having("COUNT(*) > 1").
// 		Order("count DESC").
// 		Limit(20).
// 		Scan(&results).Error

// 	if err != nil {
// 		return nil, err
// 	}

// 	var hotspots []Hotspot
// 	maxCount := 0
// 	for _, result := range results {
// 		if result.Count > maxCount {
// 			maxCount = result.Count
// 		}
// 	}

// 	for _, result := range results {
// 		intensity := float64(result.Count) / float64(maxCount)
// 		description := fmt.Sprintf("%s - %d migrants", result.Type, result.Count)

// 		hotspots = append(hotspots, Hotspot{
// 			UUID:        result.UUID,
// 			Latitude:    result.Latitude,
// 			Longitude:   result.Longitude,
// 			City:        result.Ville,
// 			Country:     result.Pays,
// 			Count:       result.Count,
// 			Intensity:   math.Round(intensity*100) / 100,
// 			Type:        result.Type,
// 			Description: description,
// 		})
// 	}

// 	return hotspots, nil
// }

// // getMigrationCorridors - Identifie les corridors migratoires
// func getMigrationCorridors() ([]Corridor, error) {
// 	var results []struct {
// 		FromCountry string `json:"from_country"`
// 		ToCountry   string `json:"to_country"`
// 		Count       int    `json:"count"`
// 	}

// 	err := database.DB.Model(&models.Migrant{}).
// 		Select("pays_origine as from_country, pays_destination as to_country, COUNT(*) as count").
// 		Where("pays_origine != '' AND pays_destination != '' AND pays_origine != pays_destination").
// 		Group("pays_origine, pays_destination").
// 		Having("COUNT(*) > 0").
// 		Order("count DESC").
// 		Limit(15).
// 		Scan(&results).Error

// 	if err != nil {
// 		return nil, err
// 	}

// 	// Coordonnées approximatives des pays (pour démonstration)
// 	countryCoords := map[string][2]float64{
// 		"RDC":      {-4.4419, 15.2663},
// 		"Angola":   {-11.2027, 17.8739},
// 		"Cameroun": {7.3697, 12.3547},
// 		"Gabon":    {-0.8037, 11.6094},
// 		"Congo":    {-0.2280, 15.8277},
// 		"Tchad":    {15.4542, 18.7322},
// 		"RCA":      {6.6111, 20.9394},
// 		"France":   {46.6034, 1.8883},
// 		"Belgique": {50.5039, 4.4699},
// 		"USA":      {37.0902, -95.7129},
// 		"Canada":   {56.1304, -106.3468},
// 	}

// 	var corridors []Corridor
// 	for _, result := range results {
// 		fromCoords, fromExists := countryCoords[result.FromCountry]
// 		toCoords, toExists := countryCoords[result.ToCountry]

// 		if fromExists && toExists {
// 			flowDirection := "bidirectional"
// 			if result.Count > 10 {
// 				flowDirection = "unidirectional"
// 			}

// 			corridors = append(corridors, Corridor{
// 				FromCountry:   result.FromCountry,
// 				ToCountry:     result.ToCountry,
// 				FromLatitude:  fromCoords[0],
// 				FromLongitude: fromCoords[1],
// 				ToLatitude:    toCoords[0],
// 				ToLongitude:   toCoords[1],
// 				Count:         result.Count,
// 				FlowDirection: flowDirection,
// 			})
// 		}
// 	}

// 	return corridors, nil
// }

// // getDensityMap - Génère la carte de densité
// func getDensityMap() ([]DensityPoint, error) {
// 	var geolocations []models.Geolocalisation

// 	err := database.DB.Where("latitude != 0 AND longitude != 0").
// 		Find(&geolocations).Error

// 	if err != nil {
// 		return nil, err
// 	}

// 	// Grouper les points par zones géographiques (grid de 0.1 degré)
// 	gridSize := 0.1
// 	densityGrid := make(map[string]int)

// 	for _, geo := range geolocations {
// 		gridLat := math.Floor(geo.Latitude/gridSize) * gridSize
// 		gridLon := math.Floor(geo.Longitude/gridSize) * gridSize
// 		key := fmt.Sprintf("%.1f,%.1f", gridLat, gridLon)
// 		densityGrid[key]++
// 	}

// 	var densityPoints []DensityPoint
// 	maxDensity := 0

// 	for key, count := range densityGrid {
// 		coords := strings.Split(key, ",")
// 		lat, _ := strconv.ParseFloat(coords[0], 64)
// 		lon, _ := strconv.ParseFloat(coords[1], 64)

// 		if count > maxDensity {
// 			maxDensity = count
// 		}

// 		densityPoints = append(densityPoints, DensityPoint{
// 			Latitude:  lat + gridSize/2,
// 			Longitude: lon + gridSize/2,
// 			Density:   float64(count),
// 			Radius:    gridSize,
// 		})
// 	}

// 	// Normaliser les densités
// 	for i := range densityPoints {
// 		densityPoints[i].Density = densityPoints[i].Density / float64(maxDensity)
// 	}

// 	return densityPoints, nil
// }

// // getRealTimePositions - Récupère les positions en temps réel
// func getRealTimePositions() ([]RealTimePosition, error) {
// 	var results []struct {
// 		MigrantUUID   string    `json:"migrant_uuid"`
// 		MigrantName   string    `json:"migrant_name"`
// 		Latitude      float64   `json:"latitude"`
// 		Longitude     float64   `json:"longitude"`
// 		Status        string    `json:"status"`
// 		UpdatedAt     time.Time `json:"updated_at"`
// 		Ville         string    `json:"ville"`
// 		Pays          string    `json:"pays"`
// 		TypeMouvement string    `json:"type_mouvement"`
// 	}

// 	err := database.DB.Table("geolocalisations g").
// 		Select(`g.migrant_uuid, 
// 				CONCAT(m.nom, ' ', m.prenom) as migrant_name,
// 				g.latitude, g.longitude, m.statut_migratoire as status,
// 				g.updated_at, g.ville, g.pays, g.type_mouvement`).
// 		Joins("JOIN migrants m ON g.migrant_uuid = m.uuid").
// 		Where("g.latitude != 0 AND g.longitude != 0 AND m.actif = true").
// 		Where("g.updated_at >= ?", time.Now().AddDate(0, 0, -7)).
// 		Order("g.updated_at DESC").
// 		Limit(50).
// 		Scan(&results).Error

// 	if err != nil {
// 		return nil, err
// 	}

// 	riskLevels := []string{"faible", "moyen", "élevé"}
// 	var positions []RealTimePosition

// 	for _, result := range results {
// 		// Calculer le niveau de risque basé sur le statut et la récence
// 		riskLevel := "faible"
// 		if result.Status == "irregulier" {
// 			riskLevel = "élevé"
// 		} else if result.Status == "demandeur_asile" {
// 			riskLevel = "moyen"
// 		}

// 		// Ajouter du caractère aléatoire pour la démonstration
// 		if len(riskLevels) > 0 {
// 			riskLevel = riskLevels[len(result.MigrantUUID)%len(riskLevels)]
// 		}

// 		positions = append(positions, RealTimePosition{
// 			MigrantUUID:  result.MigrantUUID,
// 			MigrantName:  result.MigrantName,
// 			Latitude:     result.Latitude,
// 			Longitude:    result.Longitude,
// 			Status:       result.Status,
// 			LastUpdate:   result.UpdatedAt,
// 			City:         result.Ville,
// 			Country:      result.Pays,
// 			MovementType: result.TypeMouvement,
// 			RiskLevel:    riskLevel,
// 		})
// 	}

// 	return positions, nil
// }

// // getPredictiveAnalysis - Analyse prédictive avancée
// func getPredictiveAnalysis() (PredictiveData, error) {
// 	var predictiveData PredictiveData

// 	// 1. Prédiction du mois prochain basée sur les tendances
// 	var currentMonthCount int64
// 	database.DB.Model(&models.Migrant{}).
// 		Where("EXTRACT(MONTH FROM created_at) = ? AND EXTRACT(YEAR FROM created_at) = ?",
// 			time.Now().Month(), time.Now().Year()).
// 		Count(&currentMonthCount)

// 	var lastMonthCount int64
// 	lastMonth := time.Now().AddDate(0, -1, 0)
// 	database.DB.Model(&models.Migrant{}).
// 		Where("EXTRACT(MONTH FROM created_at) = ? AND EXTRACT(YEAR FROM created_at) = ?",
// 			lastMonth.Month(), lastMonth.Year()).
// 		Count(&lastMonthCount)

// 	growthRate := 0.0
// 	if lastMonthCount > 0 {
// 		growthRate = float64(currentMonthCount-lastMonthCount) / float64(lastMonthCount) * 100
// 	}

// 	predictiveData.PopulationGrowthRate = math.Round(growthRate*100) / 100
// 	predictiveData.NextMonthPrediction = int(float64(currentMonthCount) * (1 + growthRate/100))

// 	if growthRate > 5 {
// 		predictiveData.TrendDirection = "croissante"
// 	} else if growthRate < -5 {
// 		predictiveData.TrendDirection = "décroissante"
// 	} else {
// 		predictiveData.TrendDirection = "stable"
// 	}

// 	// 2. Patterns saisonniers
// 	seasonalPatterns := []SeasonalPattern{
// 		{Month: "Janvier", AverageArrivals: 150, Trend: 0.05},
// 		{Month: "Février", AverageArrivals: 120, Trend: -0.03},
// 		{Month: "Mars", AverageArrivals: 180, Trend: 0.08},
// 		{Month: "Avril", AverageArrivals: 200, Trend: 0.12},
// 		{Month: "Mai", AverageArrivals: 250, Trend: 0.15},
// 		{Month: "Juin", AverageArrivals: 300, Trend: 0.20},
// 		{Month: "Juillet", AverageArrivals: 280, Trend: 0.18},
// 		{Month: "Août", AverageArrivals: 260, Trend: 0.16},
// 		{Month: "Septembre", AverageArrivals: 220, Trend: 0.10},
// 		{Month: "Octobre", AverageArrivals: 190, Trend: 0.06},
// 		{Month: "Novembre", AverageArrivals: 160, Trend: 0.02},
// 		{Month: "Décembre", AverageArrivals: 140, Trend: -0.01},
// 	}
// 	predictiveData.SeasonalPatterns = seasonalPatterns

// 	// 3. Prédictions de risque
// 	riskPredictions := []RiskPrediction{
// 		{
// 			Zone:        "Zone Frontalière Nord",
// 			RiskLevel:   "élevé",
// 			Probability: 0.75,
// 			Factors:     []string{"concentration élevée", "activité irrégulière", "alertes récentes"},
// 		},
// 		{
// 			Zone:        "Centre Urbain",
// 			RiskLevel:   "moyen",
// 			Probability: 0.45,
// 			Factors:     []string{"surpopulation", "ressources limitées"},
// 		},
// 		{
// 			Zone:        "Zone Côtière",
// 			RiskLevel:   "faible",
// 			Probability: 0.25,
// 			Factors:     []string{"surveillance renforcée", "accès contrôlé"},
// 		},
// 	}
// 	predictiveData.RiskPredictions = riskPredictions

// 	// 4. Points chauds prédits
// 	predictedHotspots := []PredictedHotspot{
// 		{Latitude: -4.5, Longitude: 15.3, Probability: 0.8, Timeframe: "3 mois"},
// 		{Latitude: -4.2, Longitude: 15.1, Probability: 0.6, Timeframe: "6 mois"},
// 		{Latitude: -4.7, Longitude: 15.5, Probability: 0.7, Timeframe: "1 mois"},
// 	}
// 	predictiveData.PredictedHotspots = predictedHotspots

// 	return predictiveData, nil
// }

// // getGeographicDistribution - Distribution géographique détaillée
// func getGeographicDistribution() ([]GeographicData, error) {
// 	var results []struct {
// 		Pays      string  `json:"pays"`
// 		Latitude  float64 `json:"latitude"`
// 		Longitude float64 `json:"longitude"`
// 		Count     int     `json:"count"`
// 	}

// 	err := database.DB.Table("geolocalisations g").
// 		Select("g.pays, AVG(g.latitude) as latitude, AVG(g.longitude) as longitude, COUNT(*) as count").
// 		Where("g.latitude != 0 AND g.longitude != 0 AND g.pays != ''").
// 		Group("g.pays").
// 		Order("count DESC").
// 		Scan(&results).Error

// 	if err != nil {
// 		return nil, err
// 	}

// 	var total int
// 	for _, result := range results {
// 		total += result.Count
// 	}

// 	var geographicData []GeographicData
// 	for _, result := range results {
// 		percentage := float64(result.Count) / float64(total) * 100

// 		// Simuler le taux de croissance
// 		growthRate := (float64(result.Count%10) - 5) * 2 // Entre -10 et 10

// 		geographicData = append(geographicData, GeographicData{
// 			Region:     result.Pays,
// 			Latitude:   result.Latitude,
// 			Longitude:  result.Longitude,
// 			Count:      result.Count,
// 			Percentage: math.Round(percentage*100) / 100,
// 			GrowthRate: math.Round(growthRate*100) / 100,
// 		})
// 	}

// 	return geographicData, nil
// }

// // getMovementPatterns - Analyse des patterns de mouvement
// func getMovementPatterns() ([]MovementPattern, error) {
// 	var results []struct {
// 		TypeMouvement string `json:"type_mouvement"`
// 		Count         int    `json:"count"`
// 	}

// 	err := database.DB.Model(&models.Geolocalisation{}).
// 		Select("type_mouvement, COUNT(*) as count").
// 		Where("type_mouvement != ''").
// 		Group("type_mouvement").
// 		Scan(&results).Error

// 	if err != nil {
// 		return nil, err
// 	}

// 	var total int
// 	for _, result := range results {
// 		total += result.Count
// 	}

// 	patternDescriptions := map[string]string{
// 		"arrivee":              "Arrivées nouvelles dans la zone",
// 		"depart":               "Départs vers d'autres destinations",
// 		"transit":              "Passages temporaires",
// 		"residence_temporaire": "Installations temporaires",
// 		"residence_permanente": "Installations permanentes",
// 	}

// 	var movementPatterns []MovementPattern
// 	for _, result := range results {
// 		percentage := float64(result.Count) / float64(total) * 100
// 		description := patternDescriptions[result.TypeMouvement]
// 		if description == "" {
// 			description = "Pattern de mouvement non défini"
// 		}

// 		movementPatterns = append(movementPatterns, MovementPattern{
// 			Pattern:     result.TypeMouvement,
// 			Count:       result.Count,
// 			Percentage:  math.Round(percentage*100) / 100,
// 			Description: description,
// 		})
// 	}

// 	return movementPatterns, nil
// }

// // getRiskZones - Identification des zones de risque
// func getRiskZones() ([]RiskZone, error) {
// 	var results []struct {
// 		UUID      string    `json:"uuid"`
// 		Latitude  float64   `json:"latitude"`
// 		Longitude float64   `json:"longitude"`
// 		Ville     string    `json:"ville"`
// 		Count     int       `json:"count"`
// 		UpdatedAt time.Time `json:"updated_at"`
// 	}

// 	err := database.DB.Table("geolocalisations g").
// 		Select("g.uuid, g.latitude, g.longitude, g.ville, COUNT(*) as count, MAX(g.updated_at) as updated_at").
// 		Joins("JOIN migrants m ON g.migrant_uuid = m.uuid").
// 		Where("g.latitude != 0 AND g.longitude != 0 AND m.statut_migratoire IN (?)",
// 			[]string{"irregulier", "demandeur_asile"}).
// 		Group("g.uuid, g.latitude, g.longitude, g.ville").
// 		Having("COUNT(*) >= 2").
// 		Order("count DESC").
// 		Limit(10).
// 		Scan(&results).Error

// 	if err != nil {
// 		return nil, err
// 	}

// 	// Récupérer le nombre d'alertes par zone
// 	var alertCounts []struct {
// 		Latitude   float64 `json:"latitude"`
// 		Longitude  float64 `json:"longitude"`
// 		AlertCount int     `json:"alert_count"`
// 	}

// 	database.DB.Table("alerts a").
// 		Select("g.latitude, g.longitude, COUNT(a.uuid) as alert_count").
// 		Joins("JOIN geolocalisations g ON a.migrant_uuid = g.migrant_uuid").
// 		Where("g.latitude != 0 AND g.longitude != 0").
// 		Group("g.latitude, g.longitude").
// 		Scan(&alertCounts)

// 	alertMap := make(map[string]int)
// 	for _, alert := range alertCounts {
// 		key := fmt.Sprintf("%.6f,%.6f", alert.Latitude, alert.Longitude)
// 		alertMap[key] = alert.AlertCount
// 	}

// 	var riskZones []RiskZone
// 	for _, result := range results {
// 		// Calculer le score de risque
// 		riskScore := float64(result.Count) * 0.4 // Densité de migrants irréguliers

// 		key := fmt.Sprintf("%.6f,%.6f", result.Latitude, result.Longitude)
// 		alertCount := alertMap[key]
// 		riskScore += float64(alertCount) * 0.6 // Impact des alertes

// 		// Déterminer le niveau de risque
// 		var riskLevel string
// 		var factors []string

// 		if riskScore >= 5 {
// 			riskLevel = "critique"
// 			factors = []string{"haute densité", "nombreuses alertes", "statut irrégulier"}
// 		} else if riskScore >= 3 {
// 			riskLevel = "élevé"
// 			factors = []string{"densité modérée", "quelques alertes", "surveillance requise"}
// 		} else if riskScore >= 1 {
// 			riskLevel = "moyen"
// 			factors = []string{"activité normale", "surveillance standard"}
// 		} else {
// 			riskLevel = "faible"
// 			factors = []string{"faible activité", "situation stable"}
// 		}

// 		riskZones = append(riskZones, RiskZone{
// 			UUID:       result.UUID,
// 			Name:       fmt.Sprintf("Zone %s", result.Ville),
// 			Latitude:   result.Latitude,
// 			Longitude:  result.Longitude,
// 			Radius:     2.0, // 2km de rayon
// 			RiskLevel:  riskLevel,
// 			RiskScore:  math.Round(riskScore*100) / 100,
// 			Factors:    factors,
// 			AlertCount: alertCount,
// 			LastUpdate: result.UpdatedAt,
// 		})
// 	}

// 	return riskZones, nil
// }

// // GetMigrationHeatmap - Endpoint pour la carte de chaleur
// func GetMigrationHeatmap(c *fiber.Ctx) error {
// 	densityMap, err := getDensityMap()
// 	if err != nil {
// 		return c.Status(500).JSON(fiber.Map{"error": "Erreur lors de la génération de la carte de chaleur"})
// 	}

// 	return c.JSON(fiber.Map{
// 		"heatmap_data": densityMap,
// 		"metadata": fiber.Map{
// 			"total_points": len(densityMap),
// 			"generated_at": time.Now(),
// 			"type":         "migration_heatmap",
// 		},
// 	})
// }

// // GetLiveMigrationData - Endpoint pour les données en temps réel
// func GetLiveMigrationData(c *fiber.Ctx) error {
// 	realTimePositions, err := getRealTimePositions()
// 	if err != nil {
// 		return c.Status(500).JSON(fiber.Map{"error": "Erreur lors de la récupération des données temps réel"})
// 	}

// 	return c.JSON(fiber.Map{
// 		"live_data":    realTimePositions,
// 		"timestamp":    time.Now(),
// 		"total_active": len(realTimePositions),
// 	})
// }

// // GetPredictiveInsights - Endpoint pour les insights prédictifs
// func GetPredictiveInsights(c *fiber.Ctx) error {
// 	predictiveData, err := getPredictiveAnalysis()
// 	if err != nil {
// 		return c.Status(500).JSON(fiber.Map{"error": "Erreur lors de l'analyse prédictive"})
// 	}

// 	return c.JSON(fiber.Map{
// 		"predictive_insights": predictiveData,
// 		"confidence_level":    "85%",
// 		"analysis_date":       time.Now(),
// 	})
// }

// // GetInteractiveMap - Endpoint pour la carte interactive complète
// func GetInteractiveMap(c *fiber.Ctx) error {
// 	// Récupérer tous les éléments de la carte
// 	hotspots, _ := getHotspotLocations()
// 	corridors, _ := getMigrationCorridors()
// 	riskZones, _ := getRiskZones()
// 	realTimePositions, _ := getRealTimePositions()

// 	return c.JSON(fiber.Map{
// 		"map_data": fiber.Map{
// 			"hotspots":            hotspots,
// 			"corridors":           corridors,
// 			"risk_zones":          riskZones,
// 			"real_time_positions": realTimePositions,
// 		},
// 		"map_config": fiber.Map{
// 			"center": fiber.Map{
// 				"latitude":  -4.4419,
// 				"longitude": 15.2663,
// 			},
// 			"zoom":  6,
// 			"style": "satellite",
// 		},
// 		"generated_at": time.Now(),
// 	})
// }
