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
	"github.com/kgermando/sysmobembo-api/controllers/qrcode"
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

	// QR Code controller
	qr := api.Group("/qrcode")
	qr.Post("/generate/:uuid", qrcode.GenerateQRCode) // Générer un QR code pour un agent
	qr.Post("/verify", qrcode.VerifyQRCode)           // Vérifier un QR code
	qr.Get("/agent/:uuid", qrcode.GetAgentByQR)       // Obtenir les infos d'un agent via QR
	qr.Put("/refresh/:uuid", qrcode.RefreshQRCode)    // Renouveler un QR code
	qr.Get("/image/:filename", qrcode.ServeQRCode)    // Servir les images QR code

	// Alerts controller
	alertsGroup := api.Group("/alerts")
	alertsGroup.Get("/paginate", alerts.GetPaginatedAlerts)
	alertsGroup.Get("/all", alerts.GetAllAlerts)
	alertsGroup.Get("/get/:uuid", alerts.GetAlert)
	alertsGroup.Get("/migrant/:uuid", alerts.GetAlertsByMigrant)
	alertsGroup.Get("/active", alerts.GetActiveAlerts)
	alertsGroup.Get("/critical", alerts.GetCriticalAlerts)
	alertsGroup.Post("/create", alerts.CreateAlert)
	alertsGroup.Put("/update/:uuid", alerts.UpdateAlert)
	alertsGroup.Put("/resolve/:uuid", alerts.ResolveAlert)
	alertsGroup.Delete("/delete/:uuid", alerts.DeleteAlert)
	alertsGroup.Get("/stats", alerts.GetAlertsStats)
	alertsGroup.Get("/dashboard", alerts.GetAlertsDashboard)
	alertsGroup.Get("/search", alerts.SearchAlerts)
	alertsGroup.Post("/auto-expire", alerts.AutoExpireAlerts)

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
	bio.Get("/search", biometrics.SearchBiometries)

	// Geolocation controller
	geo := api.Group("/geolocations")
	geo.Get("/paginate", geolocation.GetPaginatedGeolocalisations)
	geo.Get("/all", geolocation.GetAllGeolocalisations)
	geo.Get("/get/:uuid", geolocation.GetGeolocalisation)
	geo.Get("/migrant/:uuid", geolocation.GetGeolocalisationsByMigrant)
	geo.Get("/active", geolocation.GetActiveGeolocalisations)
	geo.Get("/radius", geolocation.GetGeolocalisationsWithinRadius)
	geo.Post("/create", geolocation.CreateGeolocalisation)
	geo.Put("/update/:uuid", geolocation.UpdateGeolocalisation)
	geo.Put("/validate/:uuid", geolocation.ValidateGeolocalisation)
	geo.Delete("/delete/:uuid", geolocation.DeleteGeolocalisation)
	geo.Get("/stats", geolocation.GetGeolocalisationsStats)
	geo.Get("/migration-routes", geolocation.GetMigrationRoutes)
	geo.Get("/hotspots", geolocation.GetGeolocationHotspots)
	geo.Get("/search", geolocation.SearchGeolocalisations)

	// Migrants controller
	migrant := api.Group("/migrants")
	migrant.Get("/paginate", migrants.GetPaginatedMigrants)
	migrant.Get("/all", migrants.GetAllMigrants)
	migrant.Get("/get/:uuid", migrants.GetMigrant)
	migrant.Post("/create", migrants.CreateMigrant)
	migrant.Put("/update/:uuid", migrants.UpdateMigrant)
	migrant.Delete("/delete/:uuid", migrants.DeleteMigrant)
	migrant.Get("/stats", migrants.GetMigrantsStats)

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
	motif.Get("/urgency-analysis", motifDeplacement.GetUrgencyAnalysis)
	motif.Get("/temporal-analysis", motifDeplacement.GetTemporalAnalysis)
	motif.Get("/search", motifDeplacement.SearchMotifDeplacements)

	// Dashboard GIS System controller
	dash := api.Group("/dashboard")
	gis := dash.Group("/gis")
	gis.Get("/map-config", dashboard.GetGISMapConfiguration)
	gis.Get("/migrants-layer", dashboard.GetMigrantsLayer)
	gis.Get("/routes-layer", dashboard.GetMigrationRoutesLayer)
	gis.Get("/alert-zones-layer", dashboard.GetAlertZonesLayer)

	// Dashboard Predictive Analysis controller
	predict := dash.Group("/predictive")
	predict.Get("/migration-flow", dashboard.GetMigrationFlowPrediction)
	predict.Get("/risk-analysis", dashboard.GetRiskPredictionAnalysis)
	predict.Get("/demographic", dashboard.GetDemographicPrediction)
	predict.Get("/movement-patterns", dashboard.GetMovementPatternPrediction)

	// Dashboard Realtime Monitoring controller
	realtime := dash.Group("/realtime")
	realtime.Get("/dashboard", dashboard.GetRealtimeDashboard)
	realtime.Get("/alerts", dashboard.GetRealtimeAlerts)
	realtime.Get("/movements", dashboard.GetRealtimeMovements)
	realtime.Get("/status", dashboard.GetRealtimeStatus)
	realtime.Get("/updates", dashboard.GetRealtimeUpdates)

	// Dashboard Spatial Analysis controller
	spatial := dash.Group("/spatial")
	spatial.Get("/density", dashboard.GetSpatialDensityAnalysis)
	spatial.Get("/corridors", dashboard.GetMigrationCorridors)
	spatial.Get("/proximity", dashboard.GetProximityAnalysis)
	spatial.Get("/areas-of-interest", dashboard.GetAreasOfInterest)

	// Dashboard Trajectory Analysis controller
	trajectory := dash.Group("/trajectory")
	trajectory.Get("/individual", dashboard.GetIndividualTrajectories)
	trajectory.Get("/group", dashboard.GetGroupTrajectories)
	trajectory.Get("/patterns", dashboard.GetMovementPatterns)
	trajectory.Get("/anomalies", dashboard.GetTrajectoryAnomalies)

}
