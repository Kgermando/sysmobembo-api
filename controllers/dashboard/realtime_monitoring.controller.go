package dashboard

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/kgermando/sysmobembo-api/database"
	"github.com/kgermando/sysmobembo-api/models"
	"gorm.io/gorm"
)

// =======================
// DASHBOARD TEMPS RÉEL
// =======================

// GetRealtimeDashboard - Dashboard principal avec toutes les métriques d'alertes
func GetRealtimeDashboard(c *fiber.Ctx) error {
	db := database.DB

	dashboard := map[string]interface{}{
		"timestamp":           time.Now(),
		"general_stats":       getGeneralAlertsStats(db),
		"alerts_by_type":      getAlertsByType(db),
		"alerts_by_gravity":   getAlertsByGravity(db),
		"alerts_by_status":    getAlertsByStatus(db),
		"recent_alerts":       getRecentAlerts(db, 10),
		"critical_alerts":     getCriticalAlerts(db),
		"expired_alerts":      getExpiredAlerts(db),
		"trending_alerts":     getTrendingAlerts(db),
		"geographic_alerts":   getGeographicAlerts(db),
		"migrants_at_risk":    getMigrantsAtRisk(db),
		"resolution_metrics":  getResolutionMetrics(db),
		"alert_timeline":      getAlertTimeline(db),
		"performance_metrics": getPerformanceMetrics(db),
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Dashboard temps réel récupéré avec succès",
		"data":    dashboard,
	})
}

// =======================
// MÉTRIQUES GÉNÉRALES
// =======================

func getGeneralAlertsStats(db *gorm.DB) map[string]interface{} {
	var totalAlerts, activeAlerts, resolvedAlerts, dismissedAlerts, expiredAlerts int64
	var criticalAlerts, dangerAlerts, warningAlerts, infoAlerts int64

	// Statistiques par statut
	db.Model(&models.Alert{}).Count(&totalAlerts)
	db.Model(&models.Alert{}).Where("statut = ?", "active").Count(&activeAlerts)
	db.Model(&models.Alert{}).Where("statut = ?", "resolved").Count(&resolvedAlerts)
	db.Model(&models.Alert{}).Where("statut = ?", "dismissed").Count(&dismissedAlerts)
	db.Model(&models.Alert{}).Where("statut = ?", "expired").Count(&expiredAlerts)

	// Statistiques par gravité
	db.Model(&models.Alert{}).Where("niveau_gravite = ?", "critical").Count(&criticalAlerts)
	db.Model(&models.Alert{}).Where("niveau_gravite = ?", "danger").Count(&dangerAlerts)
	db.Model(&models.Alert{}).Where("niveau_gravite = ?", "warning").Count(&warningAlerts)
	db.Model(&models.Alert{}).Where("niveau_gravite = ?", "info").Count(&infoAlerts)

	// Calcul des pourcentages
	var resolutionRate float64
	if totalAlerts > 0 {
		resolutionRate = float64(resolvedAlerts) / float64(totalAlerts) * 100
	}

	return map[string]interface{}{
		"total_alerts":     totalAlerts,
		"active_alerts":    activeAlerts,
		"resolved_alerts":  resolvedAlerts,
		"dismissed_alerts": dismissedAlerts,
		"expired_alerts":   expiredAlerts,
		"critical_alerts":  criticalAlerts,
		"danger_alerts":    dangerAlerts,
		"warning_alerts":   warningAlerts,
		"info_alerts":      infoAlerts,
		"resolution_rate":  fmt.Sprintf("%.2f%%", resolutionRate),
	}
}

// =======================
// ALERTES PAR CATÉGORIE
// =======================

func getAlertsByType(db *gorm.DB) []map[string]interface{} {
	var results []map[string]interface{}

	query := `
		SELECT 
			type_alerte,
			COUNT(*) as total,
			COUNT(CASE WHEN statut = 'active' THEN 1 END) as active,
			COUNT(CASE WHEN statut = 'resolved' THEN 1 END) as resolved,
			COUNT(CASE WHEN niveau_gravite = 'critical' THEN 1 END) as critical_count
		FROM alertes 
		WHERE deleted_at IS NULL
		GROUP BY type_alerte
		ORDER BY total DESC
	`

	db.Raw(query).Scan(&results)
	return results
}

func getAlertsByGravity(db *gorm.DB) []map[string]interface{} {
	var results []map[string]interface{}

	query := `
		SELECT 
			niveau_gravite,
			COUNT(*) as count,
			COUNT(CASE WHEN statut = 'active' THEN 1 END) as active_count,
			ROUND(AVG(CASE WHEN date_resolution IS NOT NULL 
				THEN EXTRACT(EPOCH FROM (date_resolution - created_at))/3600 
				END), 2) as avg_resolution_hours
		FROM alertes 
		WHERE deleted_at IS NULL
		GROUP BY niveau_gravite
		ORDER BY 
			CASE niveau_gravite 
				WHEN 'critical' THEN 1 
				WHEN 'danger' THEN 2 
				WHEN 'warning' THEN 3 
				WHEN 'info' THEN 4 
			END
	`

	db.Raw(query).Scan(&results)
	return results
}

func getAlertsByStatus(db *gorm.DB) []map[string]interface{} {
	var results []map[string]interface{}

	query := `
		SELECT 
			statut,
			COUNT(*) as count,
			ROUND(COUNT(*) * 100.0 / SUM(COUNT(*)) OVER(), 2) as percentage
		FROM alertes 
		WHERE deleted_at IS NULL
		GROUP BY statut
		ORDER BY count DESC
	`

	db.Raw(query).Scan(&results)
	return results
}

// =======================
// ALERTES RÉCENTES ET CRITIQUES
// =======================

func getRecentAlerts(db *gorm.DB, limit int) []models.Alert {
	var alerts []models.Alert

	db.Preload("Migrant").
		Where("statut = ?", "active").
		Order("created_at DESC").
		Limit(limit).
		Find(&alerts)

	return alerts
}

func getCriticalAlerts(db *gorm.DB) []models.Alert {
	var alerts []models.Alert

	db.Preload("Migrant").
		Where("niveau_gravite = ? AND statut = ?", "critical", "active").
		Order("created_at DESC").
		Find(&alerts)

	return alerts
}

func getExpiredAlerts(db *gorm.DB) []models.Alert {
	var alerts []models.Alert

	db.Preload("Migrant").
		Where("date_expiration < ? AND statut = ?", time.Now(), "active").
		Order("date_expiration ASC").
		Find(&alerts)

	return alerts
}

// =======================
// ANALYSES AVANCÉES
// =======================

func getTrendingAlerts(db *gorm.DB) map[string]interface{} {
	// Alertes créées dans les dernières 24h, 7 jours, 30 jours
	var last24h, last7days, last30days int64

	now := time.Now()
	db.Model(&models.Alert{}).Where("created_at >= ?", now.Add(-24*time.Hour)).Count(&last24h)
	db.Model(&models.Alert{}).Where("created_at >= ?", now.Add(-7*24*time.Hour)).Count(&last7days)
	db.Model(&models.Alert{}).Where("created_at >= ?", now.Add(-30*24*time.Hour)).Count(&last30days)

	// Tendance par type dans les 7 derniers jours
	var typesTrends []map[string]interface{}
	query := `
		SELECT 
			type_alerte,
			COUNT(*) as count_7days,
			COUNT(CASE WHEN created_at >= ? THEN 1 END) as count_24h
		FROM alertes 
		WHERE created_at >= ? AND deleted_at IS NULL
		GROUP BY type_alerte
		ORDER BY count_7days DESC
	`
	db.Raw(query, now.Add(-24*time.Hour), now.Add(-7*24*time.Hour)).Scan(&typesTrends)

	return map[string]interface{}{
		"last_24h":     last24h,
		"last_7_days":  last7days,
		"last_30_days": last30days,
		"types_trends": typesTrends,
	}
}

func getGeographicAlerts(db *gorm.DB) []map[string]interface{} {
	var results []map[string]interface{}

	query := `
		SELECT 
			g.pays,
			g.ville,
			COUNT(a.uuid) as alert_count,
			COUNT(CASE WHEN a.niveau_gravite = 'critical' THEN 1 END) as critical_count,
			COUNT(CASE WHEN a.statut = 'active' THEN 1 END) as active_count
		FROM alertes a
		JOIN migrants m ON a.migrant_uuid = m.uuid
		JOIN geolocalisations g ON m.uuid = g.migrant_uuid
		WHERE a.deleted_at IS NULL AND g.deleted_at IS NULL
		GROUP BY g.pays, g.ville
		HAVING COUNT(a.uuid) > 0
		ORDER BY alert_count DESC
		LIMIT 20
	`

	db.Raw(query).Scan(&results)
	return results
}

func getMigrantsAtRisk(db *gorm.DB) []map[string]interface{} {
	var results []map[string]interface{}

	query := `
		SELECT 
			m.uuid,
			m.nom,
			m.prenom,
			m.numero_identifiant,
			COUNT(a.uuid) as total_alerts,
			COUNT(CASE WHEN a.statut = 'active' THEN 1 END) as active_alerts,
			COUNT(CASE WHEN a.niveau_gravite = 'critical' THEN 1 END) as critical_alerts,
			MAX(a.created_at) as last_alert_date
		FROM migrants m
		JOIN alertes a ON m.uuid = a.migrant_uuid
		WHERE a.deleted_at IS NULL AND m.deleted_at IS NULL
		GROUP BY m.uuid, m.nom, m.prenom, m.numero_identifiant
		HAVING COUNT(CASE WHEN a.statut = 'active' THEN 1 END) > 0
		ORDER BY active_alerts DESC, critical_alerts DESC
		LIMIT 15
	`

	db.Raw(query).Scan(&results)
	return results
}

// =======================
// MÉTRIQUES DE PERFORMANCE
// =======================

func getResolutionMetrics(db *gorm.DB) map[string]interface{} {
	var avgResolutionHours float64
	var totalResolved int64

	// Temps moyen de résolution
	db.Model(&models.Alert{}).
		Where("statut = ? AND date_resolution IS NOT NULL", "resolved").
		Count(&totalResolved)

	if totalResolved > 0 {
		db.Raw(`
			SELECT AVG(EXTRACT(EPOCH FROM (date_resolution - created_at))/3600) 
			FROM alertes 
			WHERE statut = 'resolved' AND date_resolution IS NOT NULL AND deleted_at IS NULL
		`).Scan(&avgResolutionHours)
	}

	// Métriques par gravité
	var gravityMetrics []map[string]interface{}
	db.Raw(`
		SELECT 
			niveau_gravite,
			COUNT(*) as total_resolved,
			ROUND(AVG(EXTRACT(EPOCH FROM (date_resolution - created_at))/3600), 2) as avg_hours
		FROM alertes 
		WHERE statut = 'resolved' AND date_resolution IS NOT NULL AND deleted_at IS NULL
		GROUP BY niveau_gravite
		ORDER BY avg_hours ASC
	`).Scan(&gravityMetrics)

	return map[string]interface{}{
		"total_resolved":       totalResolved,
		"avg_resolution_hours": fmt.Sprintf("%.2f", avgResolutionHours),
		"metrics_by_gravity":   gravityMetrics,
	}
}

func getAlertTimeline(db *gorm.DB) []map[string]interface{} {
	var results []map[string]interface{}

	// Timeline des 30 derniers jours
	query := `
		SELECT 
			DATE(created_at) as date,
			COUNT(*) as total_alerts,
			COUNT(CASE WHEN niveau_gravite = 'critical' THEN 1 END) as critical_alerts,
			COUNT(CASE WHEN type_alerte = 'securite' THEN 1 END) as security_alerts,
			COUNT(CASE WHEN type_alerte = 'sante' THEN 1 END) as health_alerts,
			COUNT(CASE WHEN type_alerte = 'juridique' THEN 1 END) as legal_alerts,
			COUNT(CASE WHEN type_alerte = 'administrative' THEN 1 END) as admin_alerts,
			COUNT(CASE WHEN type_alerte = 'humanitaire' THEN 1 END) as humanitarian_alerts
		FROM alertes 
		WHERE created_at >= ? AND deleted_at IS NULL
		GROUP BY DATE(created_at)
		ORDER BY date DESC
	`

	thirtyDaysAgo := time.Now().Add(-30 * 24 * time.Hour)
	db.Raw(query, thirtyDaysAgo).Scan(&results)

	return results
}

func getPerformanceMetrics(db *gorm.DB) map[string]interface{} {
	// Temps de réponse par type d'alerte
	var responseMetrics []map[string]interface{}
	db.Raw(`
		SELECT 
			type_alerte,
			COUNT(*) as total_alerts,
			COUNT(CASE WHEN date_resolution IS NOT NULL THEN 1 END) as resolved_count,
			ROUND(AVG(CASE WHEN date_resolution IS NOT NULL 
				THEN EXTRACT(EPOCH FROM (date_resolution - created_at))/3600 
				END), 2) as avg_response_hours,
			MIN(CASE WHEN date_resolution IS NOT NULL 
				THEN EXTRACT(EPOCH FROM (date_resolution - created_at))/3600 
				END) as min_response_hours,
			MAX(CASE WHEN date_resolution IS NOT NULL 
				THEN EXTRACT(EPOCH FROM (date_resolution - created_at))/3600 
				END) as max_response_hours
		FROM alertes 
		WHERE deleted_at IS NULL
		GROUP BY type_alerte
		ORDER BY avg_response_hours ASC
	`).Scan(&responseMetrics)

	// Alertes en attente depuis plus de 48h
	var alertesEnRetard int64
	db.Model(&models.Alert{}).
		Where("statut = ? AND created_at < ?", "active", time.Now().Add(-48*time.Hour)).
		Count(&alertesEnRetard)

	// Taux de résolution par mois
	var monthlyResolution []map[string]interface{}
	db.Raw(`
		SELECT 
			DATE_TRUNC('month', created_at) as month,
			COUNT(*) as total_alerts,
			COUNT(CASE WHEN statut = 'resolved' THEN 1 END) as resolved_alerts,
			ROUND(COUNT(CASE WHEN statut = 'resolved' THEN 1 END) * 100.0 / COUNT(*), 2) as resolution_rate
		FROM alertes 
		WHERE created_at >= ? AND deleted_at IS NULL
		GROUP BY DATE_TRUNC('month', created_at)
		ORDER BY month DESC
		LIMIT 6
	`, time.Now().Add(-6*30*24*time.Hour)).Scan(&monthlyResolution)

	return map[string]interface{}{
		"response_metrics":   responseMetrics,
		"alertes_en_retard":  alertesEnRetard,
		"monthly_resolution": monthlyResolution,
	}
}

// =======================
// FONCTIONS SPÉCIALISÉES
// =======================

// GetAlertsByDateRange - Récupérer les alertes dans une période donnée
func GetAlertsByDateRange(c *fiber.Ctx) error {
	db := database.DB

	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	if startDate == "" || endDate == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "start_date et end_date sont requis (format: YYYY-MM-DD)",
		})
	}

	var alerts []models.Alert
	query := db.Preload("Migrant").Where("created_at BETWEEN ? AND ?", startDate, endDate)

	// Filtres optionnels
	if typeAlerte := c.Query("type"); typeAlerte != "" {
		query = query.Where("type_alerte = ?", typeAlerte)
	}
	if gravite := c.Query("gravite"); gravite != "" {
		query = query.Where("niveau_gravite = ?", gravite)
	}
	if statut := c.Query("statut"); statut != "" {
		query = query.Where("statut = ?", statut)
	}

	if err := query.Order("created_at DESC").Find(&alerts).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Erreur lors de la récupération des alertes",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Alertes récupérées avec succès",
		"data":    alerts,
		"count":   len(alerts),
	})
}

// GetAlertsHeatmap - Carte de chaleur des alertes par zone géographique
func GetAlertsHeatmap(c *fiber.Ctx) error {
	db := database.DB

	var heatmapData []map[string]interface{}

	query := `
		SELECT 
			g.latitude,
			g.longitude,
			g.pays,
			g.ville,
			COUNT(a.uuid) as alert_intensity,
			COUNT(CASE WHEN a.niveau_gravite = 'critical' THEN 1 END) as critical_intensity,
			STRING_AGG(DISTINCT a.type_alerte, ', ') as alert_types
		FROM geolocalisations g
		JOIN alertes a ON g.migrant_uuid = a.migrant_uuid
		WHERE a.statut = 'active' AND a.deleted_at IS NULL AND g.deleted_at IS NULL
		GROUP BY g.latitude, g.longitude, g.pays, g.ville
		HAVING COUNT(a.uuid) > 0
		ORDER BY alert_intensity DESC
	`

	if err := db.Raw(query).Scan(&heatmapData).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Erreur lors de la génération de la heatmap",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Heatmap des alertes générée avec succès",
		"data":    heatmapData,
	})
}

// GetAlertsNotifications - Récupérer les notifications d'alertes
func GetAlertsNotifications(c *fiber.Ctx) error {
	db := database.DB

	// Récupérer seulement les alertes urgentes/nouvelles
	var notifications []map[string]interface{}

	query := `
		SELECT 
			a.uuid,
			a.titre,
			a.type_alerte,
			a.niveau_gravite,
			a.created_at,
			a.statut,
			m.nom,
			m.prenom,
			m.numero_identifiant,
			CASE 
				WHEN a.niveau_gravite = 'critical' THEN 'urgent'
				WHEN a.created_at >= ? THEN 'nouvelle'
				WHEN a.date_expiration < ? THEN 'expirée'
				ELSE 'normale'
			END as priority
		FROM alertes a
		JOIN migrants m ON a.migrant_uuid = m.uuid
		WHERE a.statut = 'active' 
			AND a.deleted_at IS NULL 
			AND (
				a.niveau_gravite = 'critical' 
				OR a.created_at >= ? 
				OR (a.date_expiration IS NOT NULL AND a.date_expiration < ?)
			)
		ORDER BY 
			CASE a.niveau_gravite 
				WHEN 'critical' THEN 1 
				WHEN 'danger' THEN 2 
				ELSE 3 
			END,
			a.created_at DESC
		LIMIT 50
	`

	now := time.Now()
	last24h := now.Add(-24 * time.Hour)

	if err := db.Raw(query, last24h, now, last24h, now).Scan(&notifications).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Erreur lors de la récupération des notifications",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":    "success",
		"message":   "Notifications récupérées avec succès",
		"data":      notifications,
		"timestamp": now,
	})
}

// BulkUpdateAlerts - Mise à jour en masse des alertes
func BulkUpdateAlerts(c *fiber.Ctx) error {
	db := database.DB

	var requestData struct {
		AlertUUIDs []string `json:"alert_uuids" validate:"required"`
		Action     string   `json:"action" validate:"required,oneof=resolve dismiss reactivate"`
		Comment    string   `json:"comment"`
	}

	if err := c.BodyParser(&requestData); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Format de requête invalide",
			"error":   err.Error(),
		})
	}

	if len(requestData.AlertUUIDs) == 0 {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Aucune alerte spécifiée",
		})
	}

	// Préparer les données de mise à jour
	updateData := map[string]interface{}{}
	now := time.Now()

	switch requestData.Action {
	case "resolve":
		updateData["statut"] = "resolved"
		updateData["date_resolution"] = &now
		if requestData.Comment != "" {
			updateData["commentaire_resolution"] = requestData.Comment
		}
	case "dismiss":
		updateData["statut"] = "dismissed"
	case "reactivate":
		updateData["statut"] = "active"
		updateData["date_resolution"] = nil
		updateData["commentaire_resolution"] = ""
	}

	// Effectuer la mise à jour
	result := db.Model(&models.Alert{}).
		Where("uuid IN ?", requestData.AlertUUIDs).
		Updates(updateData)

	if result.Error != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Erreur lors de la mise à jour des alertes",
			"error":   result.Error.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":           "success",
		"message":          fmt.Sprintf("Mise à jour effectuée avec succès pour %d alertes", result.RowsAffected),
		"updated_count":    result.RowsAffected,
		"action_performed": requestData.Action,
	})
}

// GetAlertsExport - Exporter les données d'alertes pour reporting
func GetAlertsExport(c *fiber.Ctx) error {
	db := database.DB

	// Paramètres d'export
	format := c.Query("format", "json") // json, csv
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	var alerts []map[string]interface{}

	query := `
		SELECT 
			a.uuid,
			a.created_at,
			a.type_alerte,
			a.niveau_gravite,
			a.titre,
			a.description,
			a.statut,
			a.date_resolution,
			a.commentaire_resolution,
			a.personne_responsable,
			m.numero_identifiant,
			m.nom,
			m.prenom,
			m.nationalite,
			m.statut_migratoire,
			g.pays as localisation_pays,
			g.ville as localisation_ville,
			g.latitude,
			g.longitude
		FROM alertes a
		JOIN migrants m ON a.migrant_uuid = m.uuid
		LEFT JOIN geolocalisations g ON m.uuid = g.migrant_uuid
		WHERE a.deleted_at IS NULL
	`

	// Ajouter les filtres de date si spécifiés
	if startDate != "" && endDate != "" {
		query += " AND a.created_at BETWEEN '" + startDate + "' AND '" + endDate + "'"
	}

	query += " ORDER BY a.created_at DESC"

	if err := db.Raw(query).Scan(&alerts).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Erreur lors de l'export des données",
			"error":   err.Error(),
		})
	}

	// Selon le format demandé
	if format == "csv" {
		c.Set("Content-Type", "text/csv")
		c.Set("Content-Disposition", "attachment; filename=alerts_export.csv")

		// Note: Ici vous pourriez implémenter la conversion CSV
		// Pour simplifier, on retourne du JSON
	}

	return c.JSON(fiber.Map{
		"status":    "success",
		"message":   "Export des alertes généré avec succès",
		"data":      alerts,
		"count":     len(alerts),
		"timestamp": time.Now(),
	})
}
