package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kgermando/sysmobembo-api/controllers/alerts"
	"github.com/kgermando/sysmobembo-api/controllers/auth"
	"github.com/kgermando/sysmobembo-api/controllers/biometrics"
	"github.com/kgermando/sysmobembo-api/controllers/dashboard"
	"github.com/kgermando/sysmobembo-api/controllers/geolocation"
	"github.com/kgermando/sysmobembo-api/controllers/migrants"
	motifDeplacement "github.com/kgermando/sysmobembo-api/controllers/motifDeplacement"
	"github.com/kgermando/sysmobembo-api/controllers/overview"
	"github.com/kgermando/sysmobembo-api/controllers/users"

	"github.com/gofiber/fiber/v2/middleware/logger"
)

func Setup(app *fiber.App) {

	api := app.Group("/api", logger.New())

	// Authentification controller
	a := api.Group("/auth")
	a.Post("/register", auth.Register)
	a.Post("/login", auth.Login)
	a.Post("/forgot-password", auth.ForgotPassword)
	a.Post("/reset/:token", auth.ResetPassword)
	a.Post("/create-admin", auth.CreateAdminHandler) // Endpoint pour créer un admin

	// app.Use(middlewares.IsAuthenticated)

	a.Get("/user", auth.AuthUser)
	a.Put("/profil/info", auth.UpdateInfo)
	a.Put("/change-password", auth.ChangePassword)
	a.Post("/logout", auth.Logout)

	// Users controller
	u := api.Group("/users")
	u.Get("/all", users.GetAllUsers)
	u.Get("/all/paginate", users.GetPaginatedUsers)
	u.Get("/all/:uuid", users.GetAllUsersByUUID)
	u.Get("/get/:uuid", users.GetUser)
	u.Post("/create", users.CreateUser)
	u.Put("/update/:uuid", users.UpdateUser)
	u.Delete("/delete/:uuid", users.DeleteUser)
	u.Get("/export/excel", users.ExportUsersToExcel)

	// Alerts controller
	alertsGroup := api.Group("/alerts")
	alertsGroup.Get("/paginate", alerts.GetPaginatedAlerts)
	alertsGroup.Get("/all", alerts.GetAllAlerts)
	alertsGroup.Get("/get/:uuid", alerts.GetAlert)
	alertsGroup.Get("/migrant/:uuid", alerts.GetAlertsByMigrant)
	alertsGroup.Post("/create", alerts.CreateAlert)
	alertsGroup.Put("/update/:uuid", alerts.UpdateAlert)
	alertsGroup.Put("/resolve/:uuid", alerts.ResolveAlert)
	alertsGroup.Delete("/delete/:uuid", alerts.DeleteAlert)
	alertsGroup.Get("/stats", alerts.GetAlertsStats)
	alertsGroup.Get("/export/excel", alerts.ExportAlertsToExcel)

	// Biometrics controller
	bio := api.Group("/biometrics")
	bio.Get("/paginate", biometrics.GetPaginatedBiometries)
	bio.Get("/all", biometrics.GetAllBiometries)
	bio.Get("/get/:uuid", biometrics.GetBiometrie)
	bio.Get("/migrant/:uuid", biometrics.GetBiometriesByMigrant)
	bio.Get("/verified", biometrics.GetVerifiedBiometries)
	bio.Post("/create", biometrics.CreateBiometrie)
	bio.Post("/verify/:uuid", biometrics.VerifyBiometrie)
	bio.Put("/update/:uuid", biometrics.UpdateBiometrie)
	bio.Delete("/delete/:uuid", biometrics.DeleteBiometrie)
	bio.Get("/stats", biometrics.GetBiometricsStats)
	bio.Get("/export/excel", biometrics.ExportBiometriesToExcel)

	// Geolocation controller
	geo := api.Group("/geolocations")
	geo.Get("/paginate", geolocation.GetPaginatedGeolocalisations)
	geo.Get("/all", geolocation.GetAllGeolocalisations)
	geo.Get("/get/:uuid", geolocation.GetGeolocalisation)
	geo.Get("/migrant/:migrant_uuid", geolocation.GetGeolocalisationsByMigrant)
	geo.Post("/create", geolocation.CreateGeolocalisation)
	geo.Put("/update/:uuid", geolocation.UpdateGeolocalisation)
	geo.Delete("/delete/:uuid", geolocation.DeleteGeolocalisation)
	geo.Get("/stats", geolocation.GetGeolocalisationsStats)
	geo.Get("/export/excel", geolocation.ExportGeolocalisationsToExcel)

	// Migrants controller
	migrant := api.Group("/migrants")
	migrant.Get("/paginate", migrants.GetPaginatedMigrants)
	migrant.Get("/all", migrants.GetAllMigrants)
	migrant.Get("/get/:uuid", migrants.GetMigrant)
	migrant.Post("/create", migrants.CreateMigrant)
	migrant.Put("/update/:uuid", migrants.UpdateMigrant)
	migrant.Delete("/delete/:uuid", migrants.DeleteMigrant)
	migrant.Get("/stats", migrants.GetMigrantsStats)
	migrant.Get("/export/excel", migrants.ExportMigrantsToExcel)

	// Motif Deplacement controller
	motif := api.Group("/motif-deplacements")
	motif.Get("/paginate", motifDeplacement.GetPaginatedMotifDeplacements)
	motif.Get("/all", motifDeplacement.GetAllMotifDeplacements)
	motif.Get("/get/:uuid", motifDeplacement.GetMotifDeplacement)
	motif.Get("/migrant/:uuid", motifDeplacement.GetMotifsByMigrant)
	motif.Post("/create", motifDeplacement.CreateMotifDeplacement)
	motif.Put("/update/:uuid", motifDeplacement.UpdateMotifDeplacement)
	motif.Delete("/delete/:uuid", motifDeplacement.DeleteMotifDeplacement)
	motif.Get("/stats", motifDeplacement.GetMotifsStats)
	motif.Get("/export/excel", motifDeplacement.ExportMotifDeplacementsToExcel)

	// Dashboard GIS System controller
	dash := api.Group("/dashboard")
	gis := dash.Group("/gis")
	gis.Get("/statistics", dashboard.GetGISStatistics)
	gis.Get("/heatmap", dashboard.GetMigrationHeatmap)
	gis.Get("/live-data", dashboard.GetLiveMigrationData)
	gis.Get("/predictive-insights", dashboard.GetPredictiveInsights)
	gis.Get("/interactive-map", dashboard.GetInteractiveMap)

	// Dashboard Advanced Predictive Analysis controller (nouvelles fonctions)
	predic := dash.Group("/predictive")
	predic.Get("/stats", dashboard.GetAdvancedMigrationStats)
	predic.Get("/predictive", dashboard.GetAdvancedPredictiveAnalysis)
	predic.Get("/trends", dashboard.GetAdvancedMigrationTrends)
	predic.Get("/risk", dashboard.GetAdvancedRiskAnalysis)
	predic.Get("/models-performance", dashboard.GetPredictiveModelsPerformance)

	// Nouvelles routes pour les endpoints manquants
	predic.Get("/migration-flow", dashboard.GetMigrationFlow)
	predic.Get("/risk-analysis", dashboard.GetAdvancedRiskAnalysisData)
	predic.Get("/demographic", dashboard.GetDemographicAnalysis)
	predic.Get("/movement-patterns", dashboard.GetMovementPatterns)

	// Dashboard Alertes - Système de monitoring avancé
	alertsDash := dash.Group("/alerts")
	alertsDash.Get("/realtime", dashboard.GetRealtimeDashboard)
	alertsDash.Get("/date-range", dashboard.GetAlertsByDateRange)
	alertsDash.Get("/heatmap", dashboard.GetAlertsHeatmap)
	alertsDash.Get("/notifications", dashboard.GetAlertsNotifications)
	alertsDash.Put("/bulk-update", dashboard.BulkUpdateAlerts)
	alertsDash.Get("/export", dashboard.GetAlertsExport)

	// Dashboard Analyse des Déplacements - Indicateurs RDC
	deplacementDash := dash.Group("/deplacement")
	deplacementDash.Get("/analyse", dashboard.AnalyseDeplacement)
	deplacementDash.Get("/province/:province", dashboard.AnalyseDeplacementParProvince)
	deplacementDash.Get("/tendances", dashboard.TendancesEvolution)
	deplacementDash.Get("/alertes-temps-reel", dashboard.AlertesTempsReel)
	deplacementDash.Get("/repartition-geo", dashboard.RepartitionGeographiqueDetaillee)
	deplacementDash.Get("/causes", dashboard.AnalyseCausesDetaillees)

	// Dashboard Overview - APIs spécifiques pour le composant Angular overview
	overviewDash := dash.Group("/overview")
	overviewDash.Get("/indicateurs", overview.GetIndicateursGeneraux)
	overviewDash.Get("/alertes", overview.GetAlertesTempsReel)
	overviewDash.Get("/repartition", overview.GetRepartitionGeographique)

}
