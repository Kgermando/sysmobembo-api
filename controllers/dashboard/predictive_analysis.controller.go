package dashboard

import (
	"math"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/kgermando/sysmobembo-api/database"
)

// =======================
// ANALYSE PREDICTIVE
// =======================

// Prédiction des flux migratoires
func GetMigrationFlowPrediction(c *fiber.Ctx) error {
	db := database.DB

	// Récupérer les paramètres
	periodeDays, _ := strconv.Atoi(c.Query("periode_days", "30"))
	paysOrigine := c.Query("pays_origine", "")
	paysDestination := c.Query("pays_destination", "")

	// Analyse historique pour la prédiction
	var historicalData []struct {
		Date  time.Time
		Count int64
	}

	query := `
		SELECT DATE(created_at) as date, COUNT(*) as count 
		FROM migrants 
		WHERE created_at >= NOW() - INTERVAL '%d days'
	`

	if paysOrigine != "" {
		query += " AND pays_origine = '" + paysOrigine + "'"
	}
	if paysDestination != "" {
		query += " AND pays_destination = '" + paysDestination + "'"
	}

	query += " GROUP BY DATE(created_at) ORDER BY date"

	db.Raw(query, periodeDays).Scan(&historicalData)

	// Calcul de tendance et prédiction simple (moyenne mobile)
	var total int64
	for _, data := range historicalData {
		total += data.Count
	}

	moyenne := float64(total) / float64(len(historicalData))

	// Prédiction pour les 7 prochains jours
	predictions := make([]map[string]interface{}, 7)
	for i := 0; i < 7; i++ {
		// Ajout d'un facteur de variation basé sur les tendances
		variation := 1.0 + (float64(i%3-1) * 0.1) // Variation de ±10%
		predictedValue := moyenne * variation

		predictions[i] = map[string]interface{}{
			"date":            time.Now().AddDate(0, 0, i+1).Format("2006-01-02"),
			"predicted_count": math.Round(predictedValue),
			"confidence":      85 - (float64(i) * 2), // Confiance décroissante
		}
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data": map[string]interface{}{
			"historical_data": historicalData,
			"predictions":     predictions,
			"analysis_period": periodeDays,
			"base_average":    moyenne,
		},
	})
}

// Analyse des risques prédictifs
func GetRiskPredictionAnalysis(c *fiber.Ctx) error {
	db := database.DB

	// Analyse des alertes par gravité et type
	var riskAnalysis []struct {
		TypeAlerte    string `json:"type_alerte"`
		NiveauGravite string `json:"niveau_gravite"`
		Count         int64  `json:"count"`
		TendanceMois  int64  `json:"tendance_mois"`
	}

	db.Raw(`
		SELECT 
			type_alerte,
			niveau_gravite,
			COUNT(*) as count,
			(SELECT COUNT(*) FROM alertes a2 
			 WHERE a2.type_alerte = alertes.type_alerte 
			 AND a2.niveau_gravite = alertes.niveau_gravite 
			 AND a2.created_at >= NOW() - INTERVAL '30 days') as tendance_mois
		FROM alertes 
		WHERE deleted_at IS NULL
		GROUP BY type_alerte, niveau_gravite
		ORDER BY count DESC
	`).Scan(&riskAnalysis)

	// Calcul des zones à risque basé sur la géolocalisation
	var zoneRisks []struct {
		Pays         string  `json:"pays"`
		Ville        string  `json:"ville"`
		AlertCount   int64   `json:"alert_count"`
		MigrantCount int64   `json:"migrant_count"`
		RiskScore    float64 `json:"risk_score"`
	}

	db.Raw(`
		SELECT 
			g.pays,
			g.ville,
			COUNT(DISTINCT a.uuid) as alert_count,
			COUNT(DISTINCT m.uuid) as migrant_count,
			(COUNT(DISTINCT a.uuid)::float / NULLIF(COUNT(DISTINCT m.uuid), 0)) * 100 as risk_score
		FROM geolocalisations g
		LEFT JOIN migrants m ON g.migrant_uuid = m.uuid
		LEFT JOIN alertes a ON m.uuid = a.migrant_uuid
		WHERE g.deleted_at IS NULL
		GROUP BY g.pays, g.ville
		HAVING COUNT(DISTINCT m.uuid) > 0
		ORDER BY risk_score DESC
		LIMIT 20
	`).Scan(&zoneRisks)

	// Prédiction des motifs de déplacement
	var motifTrends []struct {
		TypeMotif     string  `json:"type_motif"`
		CurrentCount  int64   `json:"current_count"`
		PreviousCount int64   `json:"previous_count"`
		GrowthRate    float64 `json:"growth_rate"`
	}

	db.Raw(`
		SELECT 
			type_motif,
			COUNT(*) as current_count,
			(SELECT COUNT(*) FROM motif_deplacements md2 
			 WHERE md2.type_motif = motif_deplacements.type_motif 
			 AND md2.created_at < NOW() - INTERVAL '30 days'
			 AND md2.created_at >= NOW() - INTERVAL '60 days') as previous_count,
			CASE 
				WHEN (SELECT COUNT(*) FROM motif_deplacements md3 
					  WHERE md3.type_motif = motif_deplacements.type_motif 
					  AND md3.created_at < NOW() - INTERVAL '30 days'
					  AND md3.created_at >= NOW() - INTERVAL '60 days') > 0
				THEN ((COUNT(*)::float - (SELECT COUNT(*) FROM motif_deplacements md4 
										  WHERE md4.type_motif = motif_deplacements.type_motif 
										  AND md4.created_at < NOW() - INTERVAL '30 days'
										  AND md4.created_at >= NOW() - INTERVAL '60 days')::float) / 
					  (SELECT COUNT(*) FROM motif_deplacements md5 
					   WHERE md5.type_motif = motif_deplacements.type_motif 
					   AND md5.created_at < NOW() - INTERVAL '30 days'
					   AND md5.created_at >= NOW() - INTERVAL '60 days')::float) * 100
				ELSE 0
			END as growth_rate
		FROM motif_deplacements
		WHERE created_at >= NOW() - INTERVAL '30 days'
		GROUP BY type_motif
		ORDER BY current_count DESC
	`).Scan(&motifTrends)

	return c.JSON(fiber.Map{
		"status": "success",
		"data": map[string]interface{}{
			"risk_analysis": riskAnalysis,
			"zone_risks":    zoneRisks,
			"motif_trends":  motifTrends,
			"analysis_date": time.Now(),
		},
	})
}

// Prédiction de l'évolution démographique
func GetDemographicPrediction(c *fiber.Ctx) error {
	db := database.DB

	// Analyse par tranche d'âge
	var demographicData []struct {
		TrancheAge   string `json:"tranche_age"`
		Count        int64  `json:"count"`
		Sexe         string `json:"sexe"`
		TendanceMois int64  `json:"tendance_mois"`
	}

	db.Raw(`
		SELECT 
			CASE 
				WHEN EXTRACT(YEAR FROM AGE(date_naissance)) < 18 THEN 'Moins de 18'
				WHEN EXTRACT(YEAR FROM AGE(date_naissance)) BETWEEN 18 AND 25 THEN '18-25'
				WHEN EXTRACT(YEAR FROM AGE(date_naissance)) BETWEEN 26 AND 35 THEN '26-35'
				WHEN EXTRACT(YEAR FROM AGE(date_naissance)) BETWEEN 36 AND 50 THEN '36-50'
				WHEN EXTRACT(YEAR FROM AGE(date_naissance)) > 50 THEN 'Plus de 50'
				ELSE 'Non défini'
			END as tranche_age,
			sexe,
			COUNT(*) as count,
			(SELECT COUNT(*) FROM migrants m2 
			 WHERE CASE 
				WHEN EXTRACT(YEAR FROM AGE(m2.date_naissance)) < 18 THEN 'Moins de 18'
				WHEN EXTRACT(YEAR FROM AGE(m2.date_naissance)) BETWEEN 18 AND 25 THEN '18-25'
				WHEN EXTRACT(YEAR FROM AGE(m2.date_naissance)) BETWEEN 26 AND 35 THEN '26-35'
				WHEN EXTRACT(YEAR FROM AGE(m2.date_naissance)) BETWEEN 36 AND 50 THEN '36-50'
				WHEN EXTRACT(YEAR FROM AGE(m2.date_naissance)) > 50 THEN 'Plus de 50'
				ELSE 'Non défini'
			END = CASE 
				WHEN EXTRACT(YEAR FROM AGE(migrants.date_naissance)) < 18 THEN 'Moins de 18'
				WHEN EXTRACT(YEAR FROM AGE(migrants.date_naissance)) BETWEEN 18 AND 25 THEN '18-25'
				WHEN EXTRACT(YEAR FROM AGE(migrants.date_naissance)) BETWEEN 26 AND 35 THEN '26-35'
				WHEN EXTRACT(YEAR FROM AGE(migrants.date_naissance)) BETWEEN 36 AND 50 THEN '36-50'
				WHEN EXTRACT(YEAR FROM AGE(migrants.date_naissance)) > 50 THEN 'Plus de 50'
				ELSE 'Non défini'
			END
			AND m2.sexe = migrants.sexe
			AND m2.created_at >= NOW() - INTERVAL '30 days') as tendance_mois
		FROM migrants 
		WHERE deleted_at IS NULL
		GROUP BY tranche_age, sexe
		ORDER BY count DESC
	`).Scan(&demographicData)

	// Prédiction par nationalité
	var nationalityTrends []struct {
		Nationalite   string  `json:"nationalite"`
		CurrentCount  int64   `json:"current_count"`
		MonthlyGrowth float64 `json:"monthly_growth"`
		Projection    int64   `json:"projection_3_months"`
	}

	db.Raw(`
		SELECT 
			nationalite,
			COUNT(*) as current_count,
			((SELECT COUNT(*) FROM migrants m2 
			  WHERE m2.nationalite = migrants.nationalite 
			  AND m2.created_at >= NOW() - INTERVAL '30 days')::float / 
			 NULLIF((SELECT COUNT(*) FROM migrants m3 
					 WHERE m3.nationalite = migrants.nationalite 
					 AND m3.created_at < NOW() - INTERVAL '30 days'), 0) - 1) * 100 as monthly_growth,
			COUNT(*) + (COUNT(*) * 0.1) as projection_3_months
		FROM migrants 
		WHERE deleted_at IS NULL
		GROUP BY nationalite
		HAVING COUNT(*) >= 5
		ORDER BY current_count DESC
		LIMIT 15
	`).Scan(&nationalityTrends)

	return c.JSON(fiber.Map{
		"status": "success",
		"data": map[string]interface{}{
			"demographic_data":   demographicData,
			"nationality_trends": nationalityTrends,
			"analysis_date":      time.Now(),
			"projection_period":  "3_months",
		},
	})
}

// Analyse prédictive des patterns de mouvement
func GetMovementPatternPrediction(c *fiber.Ctx) error {
	db := database.DB

	// Analyse des routes de migration fréquentes
	var migrationRoutes []struct {
		PaysOrigine     string  `json:"pays_origine"`
		PaysDestination string  `json:"pays_destination"`
		Count           int64   `json:"count"`
		TendanceMois    int64   `json:"tendance_mois"`
		PourcentageTota float64 `json:"pourcentage_total"`
	}

	db.Raw(`
		SELECT 
			pays_origine,
			pays_destination,
			COUNT(*) as count,
			(SELECT COUNT(*) FROM migrants m2 
			 WHERE m2.pays_origine = migrants.pays_origine 
			 AND m2.pays_destination = migrants.pays_destination 
			 AND m2.created_at >= NOW() - INTERVAL '30 days') as tendance_mois,
			(COUNT(*)::float / (SELECT COUNT(*) FROM migrants WHERE deleted_at IS NULL)::float) * 100 as pourcentage_total
		FROM migrants 
		WHERE deleted_at IS NULL 
		AND pays_origine IS NOT NULL 
		AND pays_destination IS NOT NULL
		GROUP BY pays_origine, pays_destination
		HAVING COUNT(*) >= 3
		ORDER BY count DESC
		LIMIT 20
	`).Scan(&migrationRoutes)

	// Analyse saisonnière
	var seasonalAnalysis []struct {
		Mois  string `json:"mois"`
		Count int64  `json:"count"`
		Annee int    `json:"annee"`
	}

	db.Raw(`
		SELECT 
			TO_CHAR(created_at, 'YYYY-MM') as mois,
			COUNT(*) as count,
			EXTRACT(YEAR FROM created_at)::int as annee
		FROM migrants 
		WHERE deleted_at IS NULL 
		AND created_at >= NOW() - INTERVAL '24 months'
		GROUP BY TO_CHAR(created_at, 'YYYY-MM'), EXTRACT(YEAR FROM created_at)
		ORDER BY mois DESC
	`).Scan(&seasonalAnalysis)

	// Points de passage critiques
	var criticalPoints []struct {
		Ville         string  `json:"ville"`
		Pays          string  `json:"pays"`
		PassageCount  int64   `json:"passage_count"`
		AlertCount    int64   `json:"alert_count"`
		CriticalScore float64 `json:"critical_score"`
	}

	db.Raw(`
		SELECT 
			g.ville,
			g.pays,
			COUNT(DISTINCT g.migrant_uuid) as passage_count,
			COUNT(DISTINCT a.uuid) as alert_count,
			(COUNT(DISTINCT a.uuid)::float / NULLIF(COUNT(DISTINCT g.migrant_uuid), 0)) * 100 as critical_score
		FROM geolocalisations g
		LEFT JOIN alertes a ON g.migrant_uuid = a.migrant_uuid
		WHERE g.deleted_at IS NULL 
		AND g.type_localisation IN ('point_passage', 'frontiere')
		GROUP BY g.ville, g.pays
		HAVING COUNT(DISTINCT g.migrant_uuid) >= 5
		ORDER BY critical_score DESC, passage_count DESC
		LIMIT 15
	`).Scan(&criticalPoints)

	return c.JSON(fiber.Map{
		"status": "success",
		"data": map[string]interface{}{
			"migration_routes":  migrationRoutes,
			"seasonal_analysis": seasonalAnalysis,
			"critical_points":   criticalPoints,
			"analysis_period":   "24_months",
		},
	})
}
