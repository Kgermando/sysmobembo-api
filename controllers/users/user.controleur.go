package users

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/kgermando/sysmobembo-api/database"
	"github.com/kgermando/sysmobembo-api/models"
	"github.com/kgermando/sysmobembo-api/utils"
)

// Paginate
func GetPaginatedUsers(c *fiber.Ctx) error {
	db := database.DB

	// Parse query parameters for pagination
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page <= 0 {
		page = 1
	}
	limit, err := strconv.Atoi(c.Query("limit", "15"))
	if err != nil || limit <= 0 {
		limit = 15
	}
	offset := (page - 1) * limit

	// Parse search query
	search := c.Query("search", "")

	var users []models.User
	var totalRecords int64

	// Count total records matching the search query
	db.Model(&models.User{}).
		Where("nom ILIKE ? OR post_nom ILIKE ? OR prenom ILIKE ? OR role ILIKE ? OR matricule ILIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%").
		Count(&totalRecords)

	err = db.
		Where("nom ILIKE ? OR post_nom ILIKE ? OR prenom ILIKE ? OR role ILIKE ? OR matricule ILIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%").
		Offset(offset).
		Limit(limit).
		Order("users.updated_at DESC").
		Find(&users).Error

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch Users",
			"error":   err.Error(),
		})
	}

	// Calculate total pages
	totalPages := int((totalRecords + int64(limit) - 1) / int64(limit))

	//  Prepare pagination metadata
	pagination := map[string]interface{}{
		"total_records": totalRecords,
		"total_pages":   totalPages,
		"current_page":  page,
		"page_size":     limit,
	}

	// Return response
	return c.JSON(fiber.Map{
		"status":     "success",
		"message":    "Users retrieved successfully",
		"data":       users,
		"pagination": pagination,
	})
}

// query all data
func GetAllUsers(c *fiber.Ctx) error {
	db := database.DB
	var users []models.User
	db.Find(&users)
	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "All users",
		"data":    users,
	})
}

func GetAllUsersByUUID(c *fiber.Ctx) error {
	db := database.DB
	bayerUUID := c.Params("bayer_uuid")

	var users []models.User
	db.Where("bayer_uuid = ?", bayerUUID).Find(&users)
	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "All users",
		"data":    users,
	})
}

// Get one data
func GetUser(c *fiber.Ctx) error {
	uuid := c.Params("uuid")
	db := database.DB
	var user models.User
	db.Where("uuid = ?", uuid).First(&user)
	if user.Nom == "" {
		return c.Status(404).JSON(
			fiber.Map{
				"status":  "error",
				"message": "No User found",
				"data":    nil,
			},
		)
	}
	return c.JSON(
		fiber.Map{
			"status":  "success",
			"message": "User found",
			"data":    user,
		},
	)
}

// Create data
func CreateUser(c *fiber.Ctx) error {
	user := &models.User{}

	if err := c.BodyParser(user); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request format",
			"error":   err.Error(),
		})
	}

	if user.Nom == "" || user.PostNom == "" || user.Prenom == "" {
		return c.Status(400).JSON(
			fiber.Map{
				"status":  "error",
				"message": "Form not complete - nom, postnom and prenom are required",
				"data":    nil,
			},
		)
	}

	if user.Password != user.PasswordConfirm {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "passwords do not match",
		})
	}

	user.SetPassword(user.Password)

	if err := utils.ValidateStruct(*user); err != nil {
		return c.Status(400).JSON(err)
	}

	user.UUID = utils.GenerateUUID()

	if err := database.DB.Create(user).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to create user",
			"error":   err.Error(),
		})
	}

	return c.JSON(
		fiber.Map{
			"status":  "success",
			"message": "User created successfully",
			"data":    user,
		},
	)
}

// Update data
func UpdateUser(c *fiber.Ctx) error {
	uuid := c.Params("uuid")
	db := database.DB

	var updateData models.User

	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(
			fiber.Map{
				"status":  "error",
				"message": "Review your input",
				"error":   err.Error(),
			},
		)
	}

	user := new(models.User)

	if err := db.Where("uuid = ?", uuid).First(&user).Error; err != nil {
		return c.Status(404).JSON(
			fiber.Map{
				"status":  "error",
				"message": "User not found",
				"data":    nil,
			},
		)
	}

	// Mise à jour des champs
	user.Nom = updateData.Nom
	user.PostNom = updateData.PostNom
	user.Prenom = updateData.Prenom
	user.Sexe = updateData.Sexe
	user.DateNaissance = updateData.DateNaissance
	user.LieuNaissance = updateData.LieuNaissance
	user.EtatCivil = updateData.EtatCivil
	user.NombreEnfants = updateData.NombreEnfants
	user.Nationalite = updateData.Nationalite
	user.NumeroCNI = updateData.NumeroCNI
	user.DateEmissionCNI = updateData.DateEmissionCNI
	user.DateExpirationCNI = updateData.DateExpirationCNI
	user.LieuEmissionCNI = updateData.LieuEmissionCNI
	user.Email = updateData.Email
	user.Telephone = updateData.Telephone
	user.TelephoneUrgence = updateData.TelephoneUrgence
	user.Province = updateData.Province
	user.Ville = updateData.Ville
	user.Commune = updateData.Commune
	user.Quartier = updateData.Quartier
	user.Avenue = updateData.Avenue
	user.Numero = updateData.Numero
	user.Matricule = updateData.Matricule
	user.Grade = updateData.Grade
	user.Fonction = updateData.Fonction
	user.Service = updateData.Service
	user.Direction = updateData.Direction
	user.Ministere = updateData.Ministere
	user.DateRecrutement = updateData.DateRecrutement
	user.DatePriseService = updateData.DatePriseService
	user.TypeAgent = updateData.TypeAgent
	user.Statut = updateData.Statut
	user.NiveauEtude = updateData.NiveauEtude
	user.DiplomeBase = updateData.DiplomeBase
	user.UniversiteEcole = updateData.UniversiteEcole
	user.AnneeObtention = updateData.AnneeObtention
	user.Specialisation = updateData.Specialisation
	user.NumeroBancaire = updateData.NumeroBancaire
	user.Banque = updateData.Banque
	user.NumeroCNSS = updateData.NumeroCNSS
	user.NumeroONEM = updateData.NumeroONEM
	user.PhotoProfil = updateData.PhotoProfil
	user.CVDocument = updateData.CVDocument
	user.Role = updateData.Role
	user.Permission = updateData.Permission
	user.Status = updateData.Status
	user.Signature = updateData.Signature

	if err := db.Save(&user).Error; err != nil {
		return c.Status(500).JSON(
			fiber.Map{
				"status":  "error",
				"message": "Failed to update user",
				"error":   err.Error(),
			},
		)
	}

	return c.JSON(
		fiber.Map{
			"status":  "success",
			"message": "User updated successfully",
			"data":    user,
		},
	)
}

// Delete data
func DeleteUser(c *fiber.Ctx) error {
	uuid := c.Params("uuid")

	db := database.DB

	var User models.User
	db.Where("uuid = ?", uuid).First(&User)
	if User.Nom == "" {
		return c.Status(404).JSON(
			fiber.Map{
				"status":  "error",
				"message": "No User found",
				"data":    nil,
			},
		)
	}

	db.Delete(&User)

	return c.JSON(
		fiber.Map{
			"status":  "success",
			"message": "User deleted success",
			"data":    nil,
		},
	)
}
