package users

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/kgermando/sysmobembo-api/database"
	"github.com/kgermando/sysmobembo-api/models"
	"github.com/kgermando/sysmobembo-api/utils"
	"github.com/xuri/excelize/v2"
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

// Export Users to Excel with styling
func ExportUsersToExcel(c *fiber.Ctx) error {
	db := database.DB

	// Parse query parameters for filtering
	role := c.Query("role", "")
	status := c.Query("status", "")
	search := c.Query("search", "")
	startDate := c.Query("start_date", "")
	endDate := c.Query("end_date", "")

	// Build query
	query := db.Model(&models.User{})

	// Apply filters
	if role != "" {
		query = query.Where("role = ?", role)
	}
	if status != "" {
		if status == "true" || status == "Actif" {
			query = query.Where("status = ?", true)
		} else {
			query = query.Where("status = ?", false)
		}
	}
	if search != "" {
		query = query.Where("nom ILIKE ? OR post_nom ILIKE ? OR prenom ILIKE ? OR email ILIKE ? OR matricule ILIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}
	if startDate != "" {
		query = query.Where("created_at >= ?", startDate)
	}
	if endDate != "" {
		query = query.Where("created_at <= ?", endDate)
	}

	var users []models.User
	if err := query.Order("created_at DESC").Find(&users).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch users",
			"error":   err.Error(),
		})
	}

	// Create new Excel file
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	// Create sheets
	f.NewSheet("Utilisateurs")
	f.NewSheet("Statistiques")
	f.DeleteSheet("Sheet1")

	// Define styles
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 12, Color: "FFFFFF"},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"366092"}, Pattern: 1},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{Horizontal: "center", Vertical: "center"},
	})

	dataStyle, _ := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})

	activeStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Color: "008000", Bold: true},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})

	inactiveStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Color: "FF0000", Bold: true},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})

	adminStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Color: "FF6600", Bold: true},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})

	userStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Color: "0066CC", Bold: true},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{Horizontal: "left", Vertical: "center"},
	})

	// Users sheet
	sheet := "Utilisateurs"
	headers := []string{
		"UUID", "Matricule", "Nom", "Post-Nom", "Prénom", "Email", "Téléphone",
		"Province", "Ville", "Commune", "Quartier", "Avenue", "Numéro",
		"Sexe", "État Civil", "Date de Naissance", "Lieu de Naissance",
		"Nationalité", "Numéro CNI", "Date Émission CNI", "Date Expiration CNI",
		"Grade", "Fonction", "Service", "Direction", "Ministère", "Type Agent",
		"Statut", "Rôle", "Permission", "Status", "Photo Profil", "CV Document",
		"Signature", "Créé le", "Mis à jour le",
	}

	// Set headers
	for i, header := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, header)
		f.SetCellStyle(sheet, cell, cell, headerStyle)
	}

	// Set data
	for i, user := range users {
		row := i + 2

		// Basic info
		f.SetCellValue(sheet, "A"+strconv.Itoa(row), user.UUID)
		f.SetCellValue(sheet, "B"+strconv.Itoa(row), user.Matricule)
		f.SetCellValue(sheet, "C"+strconv.Itoa(row), user.Nom)
		f.SetCellValue(sheet, "D"+strconv.Itoa(row), user.PostNom)
		f.SetCellValue(sheet, "E"+strconv.Itoa(row), user.Prenom)
		f.SetCellValue(sheet, "F"+strconv.Itoa(row), user.Email)
		f.SetCellValue(sheet, "G"+strconv.Itoa(row), user.Telephone)

		// Address
		f.SetCellValue(sheet, "H"+strconv.Itoa(row), user.Province)
		f.SetCellValue(sheet, "I"+strconv.Itoa(row), user.Ville)
		f.SetCellValue(sheet, "J"+strconv.Itoa(row), user.Commune)
		f.SetCellValue(sheet, "K"+strconv.Itoa(row), user.Quartier)
		f.SetCellValue(sheet, "L"+strconv.Itoa(row), user.Avenue)
		f.SetCellValue(sheet, "M"+strconv.Itoa(row), user.Numero)

		// Personal info
		f.SetCellValue(sheet, "N"+strconv.Itoa(row), user.Sexe)
		f.SetCellValue(sheet, "O"+strconv.Itoa(row), user.EtatCivil)

		// Dates
		if !user.DateNaissance.IsZero() {
			f.SetCellValue(sheet, "P"+strconv.Itoa(row), user.DateNaissance.Format("2006-01-02"))
		}
		f.SetCellValue(sheet, "Q"+strconv.Itoa(row), user.LieuNaissance)
		f.SetCellValue(sheet, "R"+strconv.Itoa(row), user.Nationalite)
		f.SetCellValue(sheet, "S"+strconv.Itoa(row), user.NumeroCNI)

		if !user.DateEmissionCNI.IsZero() {
			f.SetCellValue(sheet, "T"+strconv.Itoa(row), user.DateEmissionCNI.Format("2006-01-02"))
		}
		if !user.DateExpirationCNI.IsZero() {
			f.SetCellValue(sheet, "U"+strconv.Itoa(row), user.DateExpirationCNI.Format("2006-01-02"))
		}

		// Professional info
		f.SetCellValue(sheet, "V"+strconv.Itoa(row), user.Grade)
		f.SetCellValue(sheet, "W"+strconv.Itoa(row), user.Fonction)
		f.SetCellValue(sheet, "X"+strconv.Itoa(row), user.Service)
		f.SetCellValue(sheet, "Y"+strconv.Itoa(row), user.Direction)
		f.SetCellValue(sheet, "Z"+strconv.Itoa(row), user.Ministere)
		f.SetCellValue(sheet, "AA"+strconv.Itoa(row), user.TypeAgent)
		f.SetCellValue(sheet, "AB"+strconv.Itoa(row), user.Statut)

		// Role with styling
		roleCell := "AC" + strconv.Itoa(row)
		f.SetCellValue(sheet, roleCell, user.Role)
		if user.Role == "Administrator" || user.Role == "Supervisor" {
			f.SetCellStyle(sheet, roleCell, roleCell, adminStyle)
		} else {
			f.SetCellStyle(sheet, roleCell, roleCell, userStyle)
		}

		f.SetCellValue(sheet, "AD"+strconv.Itoa(row), user.Permission)

		// Status with styling
		statusCell := "AE" + strconv.Itoa(row)
		statusText := "Inactif"
		if user.Status {
			statusText = "Actif"
		}
		f.SetCellValue(sheet, statusCell, statusText)
		if user.Status {
			f.SetCellStyle(sheet, statusCell, statusCell, activeStyle)
		} else {
			f.SetCellStyle(sheet, statusCell, statusCell, inactiveStyle)
		}

		// Additional info
		f.SetCellValue(sheet, "AF"+strconv.Itoa(row), user.PhotoProfil)
		f.SetCellValue(sheet, "AG"+strconv.Itoa(row), user.CVDocument)
		f.SetCellValue(sheet, "AH"+strconv.Itoa(row), user.Signature)
		f.SetCellValue(sheet, "AI"+strconv.Itoa(row), user.CreatedAt.Format("2006-01-02 15:04:05"))
		f.SetCellValue(sheet, "AJ"+strconv.Itoa(row), user.UpdatedAt.Format("2006-01-02 15:04:05"))

		// Apply data style to all cells except role and status
		for col := 1; col <= 35; col++ {
			if col != 29 && col != 31 { // Skip role and status columns
				cell, _ := excelize.CoordinatesToCellName(col, row)
				f.SetCellStyle(sheet, cell, cell, dataStyle)
			}
		}
	}

	// Auto-fit columns
	for i := 1; i <= 35; i++ {
		col, _ := excelize.ColumnNumberToName(i)
		f.SetColWidth(sheet, col, col, 15)
	}

	// Statistics sheet
	statsSheet := "Statistiques"

	// Calculate statistics
	totalUsers := len(users)
	activeUsers := 0
	inactiveUsers := 0
	adminUsers := 0
	regularUsers := 0
	maleUsers := 0
	femaleUsers := 0

	for _, user := range users {
		if user.Status {
			activeUsers++
		} else {
			inactiveUsers++
		}

		if user.Role == "Administrator" || user.Role == "Supervisor" {
			adminUsers++
		} else {
			regularUsers++
		}

		if user.Sexe == "M" || user.Sexe == "Masculin" {
			maleUsers++
		} else if user.Sexe == "F" || user.Sexe == "Féminin" {
			femaleUsers++
		}
	}

	// Statistics headers
	statsHeaders := [][]interface{}{
		{"Statistiques des Utilisateurs", ""},
		{"", ""},
		{"Total des utilisateurs", totalUsers},
		{"Utilisateurs actifs", activeUsers},
		{"Utilisateurs inactifs", inactiveUsers},
		{"", ""},
		{"Administrateurs/Superviseurs", adminUsers},
		{"Utilisateurs réguliers", regularUsers},
		{"", ""},
		{"Utilisateurs masculins", maleUsers},
		{"Utilisateurs féminins", femaleUsers},
		{"", ""},
		{"Date d'export", time.Now().Format("2006-01-02 15:04:05")},
	}

	// Set statistics data
	for i, row := range statsHeaders {
		f.SetCellValue(statsSheet, "A"+strconv.Itoa(i+1), row[0])
		f.SetCellValue(statsSheet, "B"+strconv.Itoa(i+1), row[1])
	}

	// Style statistics
	f.SetCellStyle(statsSheet, "A1", "B1", headerStyle)
	f.SetColWidth(statsSheet, "A", "A", 25)
	f.SetColWidth(statsSheet, "B", "B", 15)

	// Set active sheet
	sheetIndex, _ := f.GetSheetIndex(sheet)
	f.SetActiveSheet(sheetIndex)

	// Prepare response
	filename := fmt.Sprintf("export_utilisateurs_%s.xlsx", time.Now().Format("20060102_150405"))

	// Save to buffer and return
	buffer, err := f.WriteToBuffer()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to generate Excel file",
			"error":   err.Error(),
		})
	}

	// Set headers for file download
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, filename))
	c.Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))

	return c.Send(buffer.Bytes())
}
