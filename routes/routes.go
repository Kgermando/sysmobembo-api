package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kgermando/sysmobembo-api/controllers/auth"
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

}
