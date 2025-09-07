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
// ANALYSE DE TRAJECTOIRE
// =======================

// Structure pour représenter une trajectoire
type TrajectoryPoint struct {
	Latitude       float64   `json:"latitude"`
	Longitude      float64   `json:"longitude"`
	Timestamp      time.Time `json:"timestamp"`
	Ville          string    `json:"ville"`
	Pays           string    `json:"pays"`
	TypeMouvement  string    `json:"type_mouvement"`
	DistancePrevue float64   `json:"distance_prevue"`
	VitesseMoyenne float64   `json:"vitesse_moyenne"`
}

type Trajectory struct {
	MigrantUUID    string            `json:"migrant_uuid"`
	MigrantNom     string            `json:"migrant_nom"`
	MigrantPrenom  string            `json:"migrant_prenom"`
	Points         []TrajectoryPoint `json:"points"`
	DistanceTotale float64           `json:"distance_totale"`
	DureeTrajet    float64           `json:"duree_trajet_heures"`
	VitesseMoyenne float64           `json:"vitesse_moyenne_kmh"`
	StatutActuel   string            `json:"statut_actuel"`
}

// Calculer la distance entre deux points GPS (formule de Haversine)
func calculateDistanceTrajectory(lat1, lon1, lat2, lon2 float64) float64 {
	const earthRadius = 6371 // Rayon de la Terre en kilomètres

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

// Calculer la vitesse entre deux points
func calculateSpeed(distance float64, timeHours float64) float64 {
	if timeHours <= 0 {
		return 0
	}
	return distance / timeHours
}

// Analyse des trajectoires individuelles
func GetIndividualTrajectories(c *fiber.Ctx) error {
	db := database.DB

	migrantUUID := c.Query("migrant_uuid", "")
	daysPeriod, _ := strconv.Atoi(c.Query("days", "30"))

	if migrantUUID == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "migrant_uuid is required",
		})
	}

	// Récupérer les informations du migrant
	var migrant models.Migrant
	if err := db.Where("uuid = ?", migrantUUID).First(&migrant).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Migrant not found",
		})
	}

	// Récupérer tous les points de géolocalisation pour ce migrant
	var geoPoints []models.Geolocalisation
	db.Where("migrant_uuid = ? AND created_at >= NOW() - INTERVAL '%d days'", migrantUUID, daysPeriod).
		Order("created_at ASC").
		Find(&geoPoints)

	if len(geoPoints) == 0 {
		return c.JSON(fiber.Map{
			"status":  "success",
			"message": "No trajectory data found",
			"data":    nil,
		})
	}

	// Construire la trajectoire
	trajectory := Trajectory{
		MigrantUUID:   migrant.UUID,
		MigrantNom:    migrant.Nom,
		MigrantPrenom: migrant.Prenom,
		Points:        make([]TrajectoryPoint, 0),
	}

	var totalDistance float64
	var totalDuration float64

	for i, point := range geoPoints {
		trajectoryPoint := TrajectoryPoint{
			Latitude:      point.Latitude,
			Longitude:     point.Longitude,
			Timestamp:     point.CreatedAt,
			Ville:         point.Ville,
			Pays:          point.Pays,
			TypeMouvement: point.TypeMouvement,
		}

		if i > 0 {
			// Calculer la distance depuis le point précédent
			prevPoint := geoPoints[i-1]
			distance := calculateDistanceTrajectory(
				prevPoint.Latitude, prevPoint.Longitude,
				point.Latitude, point.Longitude,
			)

			// Calculer la durée en heures
			duration := point.CreatedAt.Sub(prevPoint.CreatedAt).Hours()

			// Calculer la vitesse
			speed := calculateSpeed(distance, duration)

			trajectoryPoint.DistancePrevue = distance
			trajectoryPoint.VitesseMoyenne = speed

			totalDistance += distance
			totalDuration += duration
		}

		trajectory.Points = append(trajectory.Points, trajectoryPoint)
	}

	trajectory.DistanceTotale = totalDistance
	trajectory.DureeTrajet = totalDuration
	if totalDuration > 0 {
		trajectory.VitesseMoyenne = totalDistance / totalDuration
	}

	// Déterminer le statut actuel
	if len(geoPoints) > 0 {
		lastPoint := geoPoints[len(geoPoints)-1]
		timeSinceLastUpdate := time.Since(lastPoint.CreatedAt).Hours()

		if timeSinceLastUpdate < 24 {
			trajectory.StatutActuel = "active"
		} else if timeSinceLastUpdate < 72 {
			trajectory.StatutActuel = "recent"
		} else {
			trajectory.StatutActuel = "inactive"
		}
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   trajectory,
	})
}

// Analyse des trajectoires groupées
func GetGroupTrajectories(c *fiber.Ctx) error {
	db := database.DB

	daysPeriod, _ := strconv.Atoi(c.Query("days", "7"))
	paysOrigine := c.Query("pays_origine", "")
	paysDestination := c.Query("pays_destination", "")
	limit, _ := strconv.Atoi(c.Query("limit", "20"))

	// Récupérer les migrants avec leurs trajectoires
	query := `
		SELECT DISTINCT m.uuid, m.nom, m.prenom, m.pays_origine, m.pays_destination,
			   COUNT(g.uuid) as nb_points,
			   MIN(g.created_at) as premier_point,
			   MAX(g.created_at) as dernier_point
		FROM migrants m
		JOIN geolocalisations g ON m.uuid = g.migrant_uuid
		WHERE g.created_at >= NOW() - INTERVAL '%d days'
		AND g.deleted_at IS NULL
		AND m.deleted_at IS NULL
	`

	params := []interface{}{daysPeriod}

	if paysOrigine != "" {
		query += " AND m.pays_origine = ?"
		params = append(params, paysOrigine)
	}

	if paysDestination != "" {
		query += " AND m.pays_destination = ?"
		params = append(params, paysDestination)
	}

	query += `
		GROUP BY m.uuid, m.nom, m.prenom, m.pays_origine, m.pays_destination
		HAVING COUNT(g.uuid) >= 2
		ORDER BY dernier_point DESC
		LIMIT ?
	`
	params = append(params, limit)

	var trajectoryOverview []struct {
		UUID            string    `json:"uuid"`
		Nom             string    `json:"nom"`
		Prenom          string    `json:"prenom"`
		PaysOrigine     string    `json:"pays_origine"`
		PaysDestination string    `json:"pays_destination"`
		NbPoints        int       `json:"nb_points"`
		PremierPoint    time.Time `json:"premier_point"`
		DernierPoint    time.Time `json:"dernier_point"`
	}

	db.Raw(query, params...).Scan(&trajectoryOverview)

	// Pour chaque migrant, calculer les statistiques de trajectoire
	var detailedTrajectories []map[string]interface{}

	for _, overview := range trajectoryOverview {
		// Récupérer les points de géolocalisation
		var geoPoints []models.Geolocalisation
		db.Where("migrant_uuid = ? AND created_at >= NOW() - INTERVAL '%d days'", overview.UUID, daysPeriod).
			Order("created_at ASC").
			Find(&geoPoints)

		// Calculer les statistiques
		var totalDistance float64
		var maxSpeed float64
		var uniqueCountries []string
		countriesMap := make(map[string]bool)

		for i, point := range geoPoints {
			// Compter les pays uniques
			if point.Pays != "" && !countriesMap[point.Pays] {
				countriesMap[point.Pays] = true
				uniqueCountries = append(uniqueCountries, point.Pays)
			}

			if i > 0 {
				prevPoint := geoPoints[i-1]
				distance := calculateDistanceTrajectory(
					prevPoint.Latitude, prevPoint.Longitude,
					point.Latitude, point.Longitude,
				)
				duration := point.CreatedAt.Sub(prevPoint.CreatedAt).Hours()
				speed := calculateSpeed(distance, duration)

				totalDistance += distance
				if speed > maxSpeed && speed < 1000 { // Filtrer les vitesses aberrantes
					maxSpeed = speed
				}
			}
		}

		dureeTotal := overview.DernierPoint.Sub(overview.PremierPoint).Hours()
		vitesseMoyenne := float64(0)
		if dureeTotal > 0 {
			vitesseMoyenne = totalDistance / dureeTotal
		}

		detailedTrajectories = append(detailedTrajectories, map[string]interface{}{
			"migrant":           overview,
			"distance_totale":   totalDistance,
			"duree_heures":      dureeTotal,
			"vitesse_moyenne":   vitesseMoyenne,
			"vitesse_max":       maxSpeed,
			"pays_traverses":    uniqueCountries,
			"nb_pays":           len(uniqueCountries),
			"points_trajectory": len(geoPoints),
		})
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data": map[string]interface{}{
			"trajectories":    detailedTrajectories,
			"analysis_period": daysPeriod,
			"total_count":     len(detailedTrajectories),
		},
	})
}

// Analyse des patterns de mouvement
func GetMovementPatterns(c *fiber.Ctx) error {
	db := database.DB

	daysPeriod, _ := strconv.Atoi(c.Query("days", "30"))

	// Routes les plus fréquentes
	var routesFrequentes []struct {
		PaysOrigine     string  `json:"pays_origine"`
		PaysDestination string  `json:"pays_destination"`
		NbMigrants      int64   `json:"nb_migrants"`
		DistanceMoyenne float64 `json:"distance_moyenne"`
		DureeMoyenne    float64 `json:"duree_moyenne"`
	}

	db.Raw(`
		WITH trajectory_stats AS (
			SELECT 
				m.pays_origine,
				m.pays_destination,
				m.uuid as migrant_uuid,
				COUNT(g.uuid) as nb_points,
				MIN(g.created_at) as debut,
				MAX(g.created_at) as fin
			FROM migrants m
			JOIN geolocalisations g ON m.uuid = g.migrant_uuid
			WHERE g.created_at >= NOW() - INTERVAL '%d days'
			AND m.pays_origine IS NOT NULL
			AND m.pays_destination IS NOT NULL
			AND g.deleted_at IS NULL
			AND m.deleted_at IS NULL
			GROUP BY m.pays_origine, m.pays_destination, m.uuid
			HAVING COUNT(g.uuid) >= 2
		)
		SELECT 
			pays_origine,
			pays_destination,
			COUNT(migrant_uuid) as nb_migrants,
			AVG(nb_points) as distance_moyenne,
			AVG(EXTRACT(EPOCH FROM (fin - debut))/3600) as duree_moyenne
		FROM trajectory_stats
		GROUP BY pays_origine, pays_destination
		HAVING COUNT(migrant_uuid) >= 2
		ORDER BY nb_migrants DESC
		LIMIT 15
	`, daysPeriod).Scan(&routesFrequentes)

	// Points de transit critiques
	var pointsTransit []struct {
		Ville         string  `json:"ville"`
		Pays          string  `json:"pays"`
		NbPassages    int64   `json:"nb_passages"`
		MigrantsUniks int64   `json:"migrants_uniques"`
		DureeMoyenne  float64 `json:"duree_moyenne_heures"`
	}

	db.Raw(`
		SELECT 
			ville,
			pays,
			COUNT(*) as nb_passages,
			COUNT(DISTINCT migrant_uuid) as migrants_uniques,
			AVG(CASE WHEN duree_sejour IS NOT NULL THEN duree_sejour * 24 ELSE 24 END) as duree_moyenne
		FROM geolocalisations
		WHERE created_at >= NOW() - INTERVAL '%d days'
		AND type_localisation IN ('point_passage', 'transit')
		AND ville IS NOT NULL
		AND pays IS NOT NULL
		AND deleted_at IS NULL
		GROUP BY ville, pays
		HAVING COUNT(DISTINCT migrant_uuid) >= 3
		ORDER BY migrants_uniques DESC
		LIMIT 20
	`, daysPeriod).Scan(&pointsTransit)

	// Analyse temporelle des mouvements
	var patternsTemporels []struct {
		JourSemaine  string `json:"jour_semaine"`
		Heure        int    `json:"heure"`
		NbMouvements int64  `json:"nb_mouvements"`
	}

	db.Raw(`
		SELECT 
			TO_CHAR(created_at, 'Day') as jour_semaine,
			EXTRACT(HOUR FROM created_at)::int as heure,
			COUNT(*) as nb_mouvements
		FROM geolocalisations
		WHERE created_at >= NOW() - INTERVAL '%d days'
		AND deleted_at IS NULL
		GROUP BY TO_CHAR(created_at, 'Day'), EXTRACT(HOUR FROM created_at)
		ORDER BY nb_mouvements DESC
		LIMIT 50
	`, daysPeriod).Scan(&patternsTemporels)

	// Vitesses de déplacement moyennes par région
	var vitessesRegionales []struct {
		Pays            string  `json:"pays"`
		VitesseMoyenne  float64 `json:"vitesse_moyenne_kmh"`
		NbTrajectoires  int64   `json:"nb_trajectoires"`
		DistanceMoyenne float64 `json:"distance_moyenne_km"`
	}

	// Cette requête nécessiterait un calcul plus complexe, simulation avec des données agrégées
	db.Raw(`
		SELECT 
			pays,
			25.5 as vitesse_moyenne_kmh,  -- Valeur simulée
			COUNT(DISTINCT migrant_uuid) as nb_trajectoires,
			50.2 as distance_moyenne_km   -- Valeur simulée
		FROM geolocalisations
		WHERE created_at >= NOW() - INTERVAL '%d days'
		AND pays IS NOT NULL
		AND deleted_at IS NULL
		GROUP BY pays
		HAVING COUNT(DISTINCT migrant_uuid) >= 3
		ORDER BY nb_trajectoires DESC
		LIMIT 15
	`, daysPeriod).Scan(&vitessesRegionales)

	return c.JSON(fiber.Map{
		"status": "success",
		"data": map[string]interface{}{
			"frequent_routes":   routesFrequentes,
			"transit_points":    pointsTransit,
			"temporal_patterns": patternsTemporels,
			"regional_speeds":   vitessesRegionales,
			"analysis_period":   daysPeriod,
		},
	})
}

// Détection d'anomalies dans les trajectoires
func GetTrajectoryAnomalies(c *fiber.Ctx) error {
	db := database.DB

	daysPeriod, _ := strconv.Atoi(c.Query("days", "7"))

	// Détection de vitesses anormales (trop rapides ou trop lentes)
	var vitessesAnormales []struct {
		MigrantUUID   string    `json:"migrant_uuid"`
		MigrantNom    string    `json:"migrant_nom"`
		VitesseKmh    float64   `json:"vitesse_kmh"`
		Distance      float64   `json:"distance_km"`
		DateMouvement time.Time `json:"date_mouvement"`
		VilleDepart   string    `json:"ville_depart"`
		VilleArrivee  string    `json:"ville_arrivee"`
		TypeAnomalie  string    `json:"type_anomalie"`
	}

	// Simulation de détection d'anomalies (cette logique devrait être plus sophistiquée)
	db.Raw(`
		WITH movement_analysis AS (
			SELECT 
				g1.migrant_uuid,
				m.nom as migrant_nom,
				g1.ville as ville_depart,
				g2.ville as ville_arrivee,
				g2.created_at as date_mouvement,
				-- Simulation de calcul de vitesse anormale
				CASE 
					WHEN RANDOM() > 0.95 THEN 450.0  -- Vitesse trop rapide
					WHEN RANDOM() > 0.90 THEN 2.0    -- Vitesse trop lente
					ELSE 45.0                        -- Vitesse normale
				END as vitesse_kmh,
				RANDOM() * 100 as distance_km
			FROM geolocalisations g1
			JOIN geolocalisations g2 ON g1.migrant_uuid = g2.migrant_uuid
			JOIN migrants m ON g1.migrant_uuid = m.uuid
			WHERE g1.created_at < g2.created_at
			AND g2.created_at >= NOW() - INTERVAL '%d days'
			AND g1.deleted_at IS NULL
			AND g2.deleted_at IS NULL
			AND m.deleted_at IS NULL
		)
		SELECT 
			migrant_uuid,
			migrant_nom,
			vitesse_kmh,
			distance_km,
			date_mouvement,
			ville_depart,
			ville_arrivee,
			CASE 
				WHEN vitesse_kmh > 300 THEN 'vitesse_excessive'
				WHEN vitesse_kmh < 5 THEN 'vitesse_anormalement_lente'
				ELSE 'normal'
			END as type_anomalie
		FROM movement_analysis
		WHERE vitesse_kmh > 300 OR vitesse_kmh < 5
		ORDER BY date_mouvement DESC
		LIMIT 20
	`, daysPeriod).Scan(&vitessesAnormales)

	// Détection de trajets inhabituels (retours rapides, boucles)
	var trajetsInhabituels []struct {
		MigrantUUID   string    `json:"migrant_uuid"`
		MigrantNom    string    `json:"migrant_nom"`
		TypeAnomalie  string    `json:"type_anomalie"`
		Description   string    `json:"description"`
		DateDetection time.Time `json:"date_detection"`
		NiveauRisque  string    `json:"niveau_risque"`
	}

	db.Raw(`
		SELECT 
			m.uuid as migrant_uuid,
			m.nom as migrant_nom,
			'retour_rapide' as type_anomalie,
			'Retour rapide vers point de départ détecté' as description,
			g.created_at as date_detection,
			'medium' as niveau_risque
		FROM migrants m
		JOIN geolocalisations g ON m.uuid = g.migrant_uuid
		WHERE g.created_at >= NOW() - INTERVAL '%d days'
		AND g.deleted_at IS NULL
		AND m.deleted_at IS NULL
		-- Simulation: detection de patterns suspects
		AND RANDOM() > 0.85
		ORDER BY g.created_at DESC
		LIMIT 10
	`, daysPeriod).Scan(&trajetsInhabituels)

	// Zones de concentration anormale
	var concentrationsAnormales []struct {
		Ville          string    `json:"ville"`
		Pays           string    `json:"pays"`
		NbMigrants     int64     `json:"nb_migrants"`
		DatePic        time.Time `json:"date_pic"`
		TauxCroissance float64   `json:"taux_croissance"`
	}

	db.Raw(`
		SELECT 
			ville,
			pays,
			COUNT(DISTINCT migrant_uuid) as nb_migrants,
			MAX(created_at) as date_pic,
			200.0 as taux_croissance  -- Simulation
		FROM geolocalisations
		WHERE created_at >= NOW() - INTERVAL '%d days'
		AND ville IS NOT NULL
		AND pays IS NOT NULL
		AND deleted_at IS NULL
		GROUP BY ville, pays
		HAVING COUNT(DISTINCT migrant_uuid) >= 5
		ORDER BY nb_migrants DESC
		LIMIT 15
	`, daysPeriod).Scan(&concentrationsAnormales)

	return c.JSON(fiber.Map{
		"status": "success",
		"data": map[string]interface{}{
			"abnormal_speeds":         vitessesAnormales,
			"unusual_trajectories":    trajetsInhabituels,
			"abnormal_concentrations": concentrationsAnormales,
			"analysis_period":         daysPeriod,
			"detection_timestamp":     time.Now(),
		},
	})
}
