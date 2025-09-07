package dashboard

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kgermando/sysmobembo-api/database"
)

// SetupDashboardRoutes configure toutes les routes pour les dashboards
func SetupDashboardRoutes(app *fiber.App) {
	// Groupe principal pour les dashboards
	dashboardGroup := app.Group("/api/dashboard")

	// ===========================
	// ANALYSE PREDICTIVE
	// ===========================
	predictiveGroup := dashboardGroup.Group("/predictive")
	predictiveGroup.Get("/migration-flow", GetMigrationFlowPrediction)
	predictiveGroup.Get("/risk-analysis", GetRiskPredictionAnalysis)
	predictiveGroup.Get("/demographic-prediction", GetDemographicPrediction)
	predictiveGroup.Get("/movement-patterns", GetMovementPatternPrediction)

	// ===========================
	// SUIVI TEMPS REEL
	// ===========================
	realtimeGroup := dashboardGroup.Group("/realtime")
	realtimeGroup.Get("/dashboard", GetRealtimeDashboard)
	realtimeGroup.Get("/alerts", GetRealtimeAlerts)
	realtimeGroup.Get("/movements", GetRealtimeMovements)
	realtimeGroup.Get("/status", GetRealtimeStatus)
	realtimeGroup.Get("/updates", GetRealtimeUpdates)

	// ===========================
	// ANALYSE DE TRAJECTOIRE
	// ===========================
	trajectoryGroup := dashboardGroup.Group("/trajectory")
	trajectoryGroup.Get("/individual", GetIndividualTrajectories)
	trajectoryGroup.Get("/group", GetGroupTrajectories)
	trajectoryGroup.Get("/patterns", GetMovementPatterns)
	trajectoryGroup.Get("/anomalies", GetTrajectoryAnomalies)

	// ===========================
	// ANALYSE SPATIALE
	// ===========================
	spatialGroup := dashboardGroup.Group("/spatial")
	spatialGroup.Get("/density", GetSpatialDensityAnalysis)
	spatialGroup.Get("/corridors", GetMigrationCorridors)
	spatialGroup.Get("/proximity", GetProximityAnalysis)
	spatialGroup.Get("/areas-of-interest", GetAreasOfInterest)

	// ===========================
	// SYSTEME D'INFORMATION GEOGRAPHIQUE (SIG)
	// ===========================
	gisGroup := dashboardGroup.Group("/gis")
	gisGroup.Get("/config", GetGISMapConfiguration)
	gisGroup.Get("/layers/migrants", GetMigrantsLayer)
	gisGroup.Get("/layers/routes", GetMigrationRoutesLayer)
	gisGroup.Get("/layers/alerts", GetAlertZonesLayer)
	gisGroup.Get("/layers/infrastructure", GetInfrastructureLayer)
	gisGroup.Get("/layers/heatmap", GetDensityHeatmapLayer)
	gisGroup.Get("/export", ExportGISData)

	// ===========================
	// ROUTES COMBINÉES POUR VUE D'ENSEMBLE
	// ===========================
	overviewGroup := dashboardGroup.Group("/overview")
	overviewGroup.Get("/summary", GetDashboardSummary)
	overviewGroup.Get("/statistics", GetGlobalStatistics)
}

// GetDashboardSummary - Vue d'ensemble complète du dashboard
func GetDashboardSummary(c *fiber.Ctx) error {
	// Cette fonction pourrait combiner des données de plusieurs dashboards
	// pour fournir une vue d'ensemble rapide

	summary := map[string]interface{}{
		"dashboard_modules": []string{
			"predictive_analysis",
			"realtime_monitoring",
			"trajectory_analysis",
			"spatial_analysis",
			"gis_system",
		},
		"last_updated":  "2025-09-07T00:00:00Z",
		"system_status": "operational",
		"available_endpoints": map[string][]string{
			"predictive": {
				"/api/dashboard/predictive/migration-flow",
				"/api/dashboard/predictive/risk-analysis",
				"/api/dashboard/predictive/demographic-prediction",
				"/api/dashboard/predictive/movement-patterns",
			},
			"realtime": {
				"/api/dashboard/realtime/dashboard",
				"/api/dashboard/realtime/alerts",
				"/api/dashboard/realtime/movements",
				"/api/dashboard/realtime/status",
				"/api/dashboard/realtime/updates",
			},
			"trajectory": {
				"/api/dashboard/trajectory/individual",
				"/api/dashboard/trajectory/group",
				"/api/dashboard/trajectory/patterns",
				"/api/dashboard/trajectory/anomalies",
			},
			"spatial": {
				"/api/dashboard/spatial/density",
				"/api/dashboard/spatial/corridors",
				"/api/dashboard/spatial/proximity",
				"/api/dashboard/spatial/areas-of-interest",
			},
			"gis": {
				"/api/dashboard/gis/config",
				"/api/dashboard/gis/layers/migrants",
				"/api/dashboard/gis/layers/routes",
				"/api/dashboard/gis/layers/alerts",
				"/api/dashboard/gis/layers/infrastructure",
				"/api/dashboard/gis/layers/heatmap",
				"/api/dashboard/gis/export",
			},
		},
	}

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   summary,
	})
}

// GetGlobalStatistics - Statistiques globales rapides
func GetGlobalStatistics(c *fiber.Ctx) error {
	db := database.DB

	var stats struct {
		TotalMigrants      int64 `json:"total_migrants"`
		TotalAlertes       int64 `json:"total_alertes"`
		TotalLocalisations int64 `json:"total_localisations"`
		TotalBiometries    int64 `json:"total_biometries"`
		PaysCouverture     int64 `json:"pays_couverture"`
		VillesCouverture   int64 `json:"villes_couverture"`
	}

	// Compter les totaux
	db.Raw("SELECT COUNT(*) FROM migrants WHERE deleted_at IS NULL").Scan(&stats.TotalMigrants)
	db.Raw("SELECT COUNT(*) FROM alertes WHERE deleted_at IS NULL").Scan(&stats.TotalAlertes)
	db.Raw("SELECT COUNT(*) FROM geolocalisations WHERE deleted_at IS NULL").Scan(&stats.TotalLocalisations)
	db.Raw("SELECT COUNT(*) FROM biometries WHERE deleted_at IS NULL").Scan(&stats.TotalBiometries)

	// Compter la couverture géographique
	db.Raw("SELECT COUNT(DISTINCT pays) FROM geolocalisations WHERE pays IS NOT NULL AND deleted_at IS NULL").Scan(&stats.PaysCouverture)
	db.Raw("SELECT COUNT(DISTINCT ville) FROM geolocalisations WHERE ville IS NOT NULL AND deleted_at IS NULL").Scan(&stats.VillesCouverture)

	return c.JSON(fiber.Map{
		"status": "success",
		"data":   stats,
	})
}
