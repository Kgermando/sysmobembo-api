package alerts

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

// =======================
// CRUD OPERATIONS
// =======================

// Paginate - Récupérer les alertes avec pagination
func GetPaginatedAlerts(c *fiber.Ctx) error {
	db := database.DB

	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page <= 0 {
		page = 1
	}
	limit, err := strconv.Atoi(c.Query("limit", "15"))
	if err != nil || limit <= 0 {
		limit = 15
	}
	offset := (page - 1) * limit

	search := c.Query("search", "")
	migrantUUID := c.Query("migrant_uuid", "")
	statut := c.Query("statut", "")
	gravite := c.Query("gravite", "")

	var alerts []models.Alert
	var totalRecords int64

	query := db.Model(&models.Alert{}).Preload("Migrant")

	// Filtrer par migrant si spécifié
	if migrantUUID != "" {
		query = query.Where("migrant_uuid = ?", migrantUUID)
	}

	// Filtrer par statut
	if statut != "" {
		query = query.Where("statut = ?", statut)
	}

	// Filtrer par gravité
	if gravite != "" {
		query = query.Where("niveau_gravite = ?", gravite)
	}

	// Recherche textuelle
	if search != "" {
		query = query.Where("titre ILIKE ? OR description ILIKE ? OR type_alerte ILIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	// Count total
	query.Count(&totalRecords)

	// Get paginated results
	err = query.Offset(offset).
		Limit(limit).
		Order("created_at DESC").
		Find(&alerts).Error

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch alerts",
			"error":   err.Error(),
		})
	}

	totalPages := int((totalRecords + int64(limit) - 1) / int64(limit))

	pagination := map[string]interface{}{
		"total_records": totalRecords,
		"total_pages":   totalPages,
		"current_page":  page,
		"page_size":     limit,
	}

	return c.JSON(fiber.Map{
		"status":     "success",
		"message":    "Alerts retrieved successfully",
		"data":       alerts,
		"pagination": pagination,
	})
}

// Get all alerts
func GetAllAlerts(c *fiber.Ctx) error {
	db := database.DB
	var alerts []models.Alert

	err := db.Preload("Migrant").
		Order("niveau_gravite DESC, created_at DESC").
		Find(&alerts).Error

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch alerts",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "All alerts",
		"data":    alerts,
	})
}

// Get one alert
func GetAlert(c *fiber.Ctx) error {
	uuid := c.Params("uuid")
	db := database.DB
	var alert models.Alert

	err := db.Where("uuid = ?", uuid).
		Preload("Migrant").
		First(&alert).Error

	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Alert not found",
			"data":    nil,
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Alert found",
		"data":    alert,
	})
}

// Get alerts by migrant with pagination
func GetAlertsByMigrant(c *fiber.Ctx) error {
	migrantUUID := c.Params("migrant_uuid")
	db := database.DB

	// Paramètres de pagination
	page, err := strconv.Atoi(c.Query("page", "1"))
	if err != nil || page <= 0 {
		page = 1
	}
	limit, err := strconv.Atoi(c.Query("limit", "15"))
	if err != nil || limit <= 0 {
		limit = 15
	}
	offset := (page - 1) * limit

	// Paramètres de filtrage
	search := c.Query("search", "")
	statut := c.Query("statut", "")
	gravite := c.Query("gravite", "")

	var alerts []models.Alert
	var totalRecords int64

	query := db.Model(&models.Alert{}).Where("migrant_uuid = ?", migrantUUID).Preload("Migrant")

	// Filtrer par statut
	if statut != "" {
		query = query.Where("statut = ?", statut)
	}

	// Filtrer par gravité
	if gravite != "" {
		query = query.Where("niveau_gravite = ?", gravite)
	}

	// Recherche textuelle
	if search != "" {
		query = query.Where("titre ILIKE ? OR description ILIKE ? OR type_alerte ILIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	// Count total
	query.Count(&totalRecords)

	// Get paginated results
	err = query.Offset(offset).
		Limit(limit).
		Order("niveau_gravite DESC, created_at DESC").
		Find(&alerts).Error

	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch alerts for migrant",
			"error":   err.Error(),
		})
	}

	totalPages := int((totalRecords + int64(limit) - 1) / int64(limit))

	pagination := map[string]interface{}{
		"total_records": totalRecords,
		"total_pages":   totalPages,
		"current_page":  page,
		"page_size":     limit,
	}

	return c.JSON(fiber.Map{
		"status":     "success",
		"message":    "Alerts for migrant retrieved successfully",
		"data":       alerts,
		"pagination": pagination,
	})
}

// Create alert
func CreateAlert(c *fiber.Ctx) error {
	alert := &models.Alert{}

	if err := c.BodyParser(alert); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request format",
			"error":   err.Error(),
		})
	}

	// Validation des champs requis
	if alert.MigrantUUID == "" || alert.TypeAlerte == "" || alert.Titre == "" || alert.Description == "" {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "MigrantUUID, TypeAlerte, Titre, and Description are required",
			"data":    nil,
		})
	}

	// Vérifier que le migrant existe
	var migrant models.Migrant
	if err := database.DB.Where("uuid = ?", alert.MigrantUUID).First(&migrant).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Migrant not found",
			"data":    nil,
		})
	}

	// Générer l'UUID
	alert.UUID = utils.GenerateUUID()

	// Définir le statut par défaut si non spécifié
	if alert.Statut == "" {
		alert.Statut = "active"
	}

	// Validation des données
	if err := utils.ValidateStruct(*alert); err != nil {
		return c.Status(400).JSON(err)
	}

	if err := database.DB.Create(alert).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to create alert",
			"error":   err.Error(),
		})
	}

	// Recharger avec les relations
	database.DB.Preload("Migrant").First(alert, "uuid = ?", alert.UUID)

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Alert created successfully",
		"data":    alert,
	})
}

// Update alert
func UpdateAlert(c *fiber.Ctx) error {
	uuid := c.Params("uuid")
	db := database.DB

	var updateData models.Alert
	if err := c.BodyParser(&updateData); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Review your input",
			"error":   err.Error(),
		})
	}

	alert := new(models.Alert)
	if err := db.Where("uuid = ?", uuid).First(&alert).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Alert not found",
			"data":    nil,
		})
	}

	// Conserver l'UUID
	updateData.UUID = alert.UUID

	if err := db.Model(&alert).Updates(updateData).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to update alert",
			"error":   err.Error(),
		})
	}

	// Recharger avec les relations
	db.Preload("Migrant").First(&alert)

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Alert updated successfully",
		"data":    alert,
	})
}

// Resolve alert
func ResolveAlert(c *fiber.Ctx) error {
	uuid := c.Params("uuid")
	db := database.DB

	var resolutionData struct {
		CommentaireResolution string `json:"commentaire_resolution"`
	}

	if err := c.BodyParser(&resolutionData); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"status":  "error",
			"message": "Invalid request format",
			"error":   err.Error(),
		})
	}

	var alert models.Alert
	if err := db.Where("uuid = ?", uuid).First(&alert).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Alert not found",
			"data":    nil,
		})
	}

	// Marquer comme résolu
	now := time.Now()
	updateData := map[string]interface{}{
		"statut":                 "resolved",
		"date_resolution":        &now,
		"commentaire_resolution": resolutionData.CommentaireResolution,
	}

	if err := db.Model(&alert).Updates(updateData).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to resolve alert",
			"error":   err.Error(),
		})
	}

	// Recharger avec les relations
	db.Preload("Migrant").First(&alert)

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Alert resolved successfully",
		"data":    alert,
	})
}

// Delete alert
func DeleteAlert(c *fiber.Ctx) error {
	uuid := c.Params("uuid")
	db := database.DB

	var alert models.Alert
	if err := db.Where("uuid = ?", uuid).First(&alert).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"status":  "error",
			"message": "Alert not found",
			"data":    nil,
		})
	}

	if err := db.Delete(&alert).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to delete alert",
			"error":   err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Alert deleted successfully",
		"data":    nil,
	})
}

// =======================
// ANALYTICS & STATISTICS
// =======================

// Get alerts statistics
func GetAlertsStats(c *fiber.Ctx) error {
	db := database.DB

	var totalAlerts int64
	var activeAlerts int64
	var resolvedAlerts int64
	var criticalAlerts int64
	var expiredAlerts int64

	// Statistiques générales
	db.Model(&models.Alert{}).Count(&totalAlerts)
	db.Model(&models.Alert{}).Where("statut = ?", "active").Count(&activeAlerts)
	db.Model(&models.Alert{}).Where("statut = ?", "resolved").Count(&resolvedAlerts)
	db.Model(&models.Alert{}).Where("niveau_gravite = ?", "critical").Count(&criticalAlerts)
	db.Model(&models.Alert{}).Where("statut = ?", "expired").Count(&expiredAlerts)

	// Statistiques par type d'alerte
	var alertTypes []map[string]interface{}
	db.Model(&models.Alert{}).
		Select("type_alerte, COUNT(*) as count").
		Group("type_alerte").
		Order("count DESC").
		Scan(&alertTypes)

	// Statistiques par niveau de gravité
	var gravityStats []map[string]interface{}
	db.Model(&models.Alert{}).
		Select("niveau_gravite, COUNT(*) as count").
		Group("niveau_gravite").
		Order("count DESC").
		Scan(&gravityStats)

	stats := map[string]interface{}{
		"total_alerts":         totalAlerts,
		"active_alerts":        activeAlerts,
		"resolved_alerts":      resolvedAlerts,
		"critical_alerts":      criticalAlerts,
		"expired_alerts":       expiredAlerts,
		"alert_types":          alertTypes,
		"gravity_distribution": gravityStats,
	}

	return c.JSON(fiber.Map{
		"status":  "success",
		"message": "Alerts statistics",
		"data":    stats,
	})
}

// =======================
// EXCEL EXPORT
// =======================

// ExportAlertsToExcel - Exporter les alertes vers Excel avec mise en forme
func ExportAlertsToExcel(c *fiber.Ctx) error {
	db := database.DB

	// Récupérer les paramètres de filtre
	migrantUUID := c.Query("migrant_uuid", "")
	typeAlerte := c.Query("type_alerte", "")
	niveauGravite := c.Query("niveau_gravite", "")
	statut := c.Query("statut", "")
	search := c.Query("search", "")
	dateDebut := c.Query("date_debut", "")
	dateFin := c.Query("date_fin", "")

	var alerts []models.Alert

	query := db.Model(&models.Alert{}).Preload("Migrant")

	// Appliquer les filtres
	if migrantUUID != "" {
		query = query.Where("migrant_uuid = ?", migrantUUID)
	}
	if typeAlerte != "" {
		query = query.Where("type_alerte = ?", typeAlerte)
	}
	if niveauGravite != "" {
		query = query.Where("niveau_gravite = ?", niveauGravite)
	}
	if statut != "" {
		query = query.Where("statut = ?", statut)
	}
	if search != "" {
		query = query.Where("titre ILIKE ? OR description ILIKE ? OR action_requise ILIKE ?",
			"%"+search+"%", "%"+search+"%", "%"+search+"%")
	}
	if dateDebut != "" && dateFin != "" {
		query = query.Where("created_at BETWEEN ? AND ?", dateDebut, dateFin)
	} else if dateDebut != "" {
		query = query.Where("created_at >= ?", dateDebut)
	} else if dateFin != "" {
		query = query.Where("created_at <= ?", dateFin)
	}

	// Récupérer toutes les données
	err := query.Order("created_at DESC").Find(&alerts).Error
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to fetch alerts for export",
			"error":   err.Error(),
		})
	}

	// Créer un nouveau fichier Excel
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	// Supprimer la feuille par défaut et créer notre feuille
	f.DeleteSheet("Sheet1")
	index, err := f.NewSheet("Alertes")
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to create Excel sheet",
			"error":   err.Error(),
		})
	}
	f.SetActiveSheet(index)

	// ===== STYLES =====
	// Style pour l'en-tête principal
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   16,
			Color:  "FFFFFF",
			Family: "Calibri",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#2E75B6"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to create header style",
			"error":   err.Error(),
		})
	}

	// Style pour les en-têtes de colonnes
	columnHeaderStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   12,
			Color:  "FFFFFF",
			Family: "Calibri",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#4F81BD"},
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 1},
			{Type: "top", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
		},
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to create column header style",
			"error":   err.Error(),
		})
	}

	// Style pour les cellules de données
	dataStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   11,
			Family: "Calibri",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "left",
			Vertical:   "center",
			WrapText:   true,
		},
		Border: []excelize.Border{
			{Type: "left", Color: "CCCCCC", Style: 1},
			{Type: "top", Color: "CCCCCC", Style: 1},
			{Type: "bottom", Color: "CCCCCC", Style: 1},
			{Type: "right", Color: "CCCCCC", Style: 1},
		},
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to create data style",
			"error":   err.Error(),
		})
	}

	// Style pour les cellules de date
	dateStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   11,
			Family: "Calibri",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "CCCCCC", Style: 1},
			{Type: "top", Color: "CCCCCC", Style: 1},
			{Type: "bottom", Color: "CCCCCC", Style: 1},
			{Type: "right", Color: "CCCCCC", Style: 1},
		},
		NumFmt: 14, // Format de date mm/dd/yyyy
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to create date style",
			"error":   err.Error(),
		})
	}

	// Style pour les niveaux de gravité avec couleurs
	criticalStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   11,
			Family: "Calibri",
			Bold:   true,
			Color:  "FFFFFF",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#DC2626"}, // Rouge pour critique
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "CCCCCC", Style: 1},
			{Type: "top", Color: "CCCCCC", Style: 1},
			{Type: "bottom", Color: "CCCCCC", Style: 1},
			{Type: "right", Color: "CCCCCC", Style: 1},
		},
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to create critical style",
			"error":   err.Error(),
		})
	}

	dangerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   11,
			Family: "Calibri",
			Bold:   true,
			Color:  "FFFFFF",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#EA580C"}, // Orange pour danger
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "CCCCCC", Style: 1},
			{Type: "top", Color: "CCCCCC", Style: 1},
			{Type: "bottom", Color: "CCCCCC", Style: 1},
			{Type: "right", Color: "CCCCCC", Style: 1},
		},
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to create danger style",
			"error":   err.Error(),
		})
	}

	warningStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   11,
			Family: "Calibri",
			Bold:   true,
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#FBBF24"}, // Jaune pour warning
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "CCCCCC", Style: 1},
			{Type: "top", Color: "CCCCCC", Style: 1},
			{Type: "bottom", Color: "CCCCCC", Style: 1},
			{Type: "right", Color: "CCCCCC", Style: 1},
		},
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to create warning style",
			"error":   err.Error(),
		})
	}

	infoStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   11,
			Family: "Calibri",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#60A5FA"}, // Bleu pour info
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "CCCCCC", Style: 1},
			{Type: "top", Color: "CCCCCC", Style: 1},
			{Type: "bottom", Color: "CCCCCC", Style: 1},
			{Type: "right", Color: "CCCCCC", Style: 1},
		},
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to create info style",
			"error":   err.Error(),
		})
	}

	// Style pour les statuts
	activeStatusStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:   11,
			Family: "Calibri",
			Bold:   true,
			Color:  "FFFFFF",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#16A34A"}, // Vert pour actif
			Pattern: 1,
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
			Vertical:   "center",
		},
		Border: []excelize.Border{
			{Type: "left", Color: "CCCCCC", Style: 1},
			{Type: "top", Color: "CCCCCC", Style: 1},
			{Type: "bottom", Color: "CCCCCC", Style: 1},
			{Type: "right", Color: "CCCCCC", Style: 1},
		},
	})
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to create active status style",
			"error":   err.Error(),
		})
	}

	// ===== EN-TÊTE PRINCIPAL =====
	currentTime := time.Now().Format("02/01/2006 15:04")
	mainHeader := fmt.Sprintf("RAPPORT D'EXPORT DES ALERTES - %s", currentTime)
	f.SetCellValue("Alertes", "A1", mainHeader)
	f.MergeCell("Alertes", "A1", "P1")
	f.SetCellStyle("Alertes", "A1", "P1", headerStyle)
	f.SetRowHeight("Alertes", 1, 30)

	// ===== INFORMATIONS DE FILTRE =====
	row := 3
	filterApplied := false
	if migrantUUID != "" || typeAlerte != "" || niveauGravite != "" || statut != "" || search != "" || dateDebut != "" || dateFin != "" {
		f.SetCellValue("Alertes", "A2", "Filtres appliqués:")
		f.SetCellStyle("Alertes", "A2", "A2", columnHeaderStyle)
		filterApplied = true

		if migrantUUID != "" {
			f.SetCellValue("Alertes", fmt.Sprintf("A%d", row), fmt.Sprintf("Migrant UUID: %s", migrantUUID))
			row++
		}
		if typeAlerte != "" {
			f.SetCellValue("Alertes", fmt.Sprintf("A%d", row), fmt.Sprintf("Type d'alerte: %s", typeAlerte))
			row++
		}
		if niveauGravite != "" {
			f.SetCellValue("Alertes", fmt.Sprintf("A%d", row), fmt.Sprintf("Niveau de gravité: %s", niveauGravite))
			row++
		}
		if statut != "" {
			f.SetCellValue("Alertes", fmt.Sprintf("A%d", row), fmt.Sprintf("Statut: %s", statut))
			row++
		}
		if search != "" {
			f.SetCellValue("Alertes", fmt.Sprintf("A%d", row), fmt.Sprintf("Recherche: %s", search))
			row++
		}
		if dateDebut != "" || dateFin != "" {
			dateFilter := "Période: "
			if dateDebut != "" && dateFin != "" {
				dateFilter += fmt.Sprintf("du %s au %s", dateDebut, dateFin)
			} else if dateDebut != "" {
				dateFilter += fmt.Sprintf("à partir du %s", dateDebut)
			} else {
				dateFilter += fmt.Sprintf("jusqu'au %s", dateFin)
			}
			f.SetCellValue("Alertes", fmt.Sprintf("A%d", row), dateFilter)
			row++
		}
		row++ // Ligne vide
	}

	if !filterApplied {
		row = 2 // Pas de filtres, commencer plus haut
	}

	// ===== EN-TÊTES DE COLONNES =====
	headers := []string{
		"UUID",
		"Migrant UUID",
		"Nom du migrant",
		"Prénom du migrant",
		"Type d'alerte",
		"Niveau de gravité",
		"Titre",
		"Description",
		"Statut",
		"Date d'expiration",
		"Action requise",
		"Personne responsable",
		"Date de résolution",
		"Commentaire de résolution",
		"Date de création",
		"Date de MAJ",
	}

	// Écrire les en-têtes
	for i, header := range headers {
		cell := fmt.Sprintf("%c%d", 'A'+i, row)
		f.SetCellValue("Alertes", cell, header)
		f.SetCellStyle("Alertes", cell, cell, columnHeaderStyle)
	}
	f.SetRowHeight("Alertes", row, 25)

	// ===== DONNÉES =====
	for i, alert := range alerts {
		dataRow := row + 1 + i

		// UUID
		cell := fmt.Sprintf("A%d", dataRow)
		f.SetCellValue("Alertes", cell, alert.UUID)
		f.SetCellStyle("Alertes", cell, cell, dataStyle)

		// Migrant UUID
		cell = fmt.Sprintf("B%d", dataRow)
		f.SetCellValue("Alertes", cell, alert.MigrantUUID)
		f.SetCellStyle("Alertes", cell, cell, dataStyle)

		// Nom du migrant
		cell = fmt.Sprintf("C%d", dataRow)
		if alert.Migrant.Identite.UUID != "" {
			f.SetCellValue("Alertes", cell, alert.Migrant.Identite.Nom)
		} else {
			f.SetCellValue("Alertes", cell, "N/A")
		}
		f.SetCellStyle("Alertes", cell, cell, dataStyle)

		// Prénom du migrant
		cell = fmt.Sprintf("D%d", dataRow)
		if alert.Migrant.Identite.UUID != "" {
			f.SetCellValue("Alertes", cell, alert.Migrant.Identite.Prenom)
		} else {
			f.SetCellValue("Alertes", cell, "N/A")
		}
		f.SetCellStyle("Alertes", cell, cell, dataStyle)

		// Type d'alerte
		cell = fmt.Sprintf("E%d", dataRow)
		f.SetCellValue("Alertes", cell, alert.TypeAlerte)
		f.SetCellStyle("Alertes", cell, cell, dataStyle)

		// Niveau de gravité avec couleur
		cell = fmt.Sprintf("F%d", dataRow)
		f.SetCellValue("Alertes", cell, alert.NiveauGravite)
		switch alert.NiveauGravite {
		case "critical":
			f.SetCellStyle("Alertes", cell, cell, criticalStyle)
		case "danger":
			f.SetCellStyle("Alertes", cell, cell, dangerStyle)
		case "warning":
			f.SetCellStyle("Alertes", cell, cell, warningStyle)
		case "info":
			f.SetCellStyle("Alertes", cell, cell, infoStyle)
		default:
			f.SetCellStyle("Alertes", cell, cell, dataStyle)
		}

		// Titre
		cell = fmt.Sprintf("G%d", dataRow)
		f.SetCellValue("Alertes", cell, alert.Titre)
		f.SetCellStyle("Alertes", cell, cell, dataStyle)

		// Description
		cell = fmt.Sprintf("H%d", dataRow)
		f.SetCellValue("Alertes", cell, alert.Description)
		f.SetCellStyle("Alertes", cell, cell, dataStyle)

		// Statut avec couleur
		cell = fmt.Sprintf("I%d", dataRow)
		f.SetCellValue("Alertes", cell, alert.Statut)
		if alert.Statut == "active" {
			f.SetCellStyle("Alertes", cell, cell, activeStatusStyle)
		} else {
			f.SetCellStyle("Alertes", cell, cell, dataStyle)
		}

		// Date d'expiration
		cell = fmt.Sprintf("J%d", dataRow)
		if alert.DateExpiration != nil {
			f.SetCellValue("Alertes", cell, alert.DateExpiration.Format("02/01/2006"))
		} else {
			f.SetCellValue("Alertes", cell, "")
		}
		f.SetCellStyle("Alertes", cell, cell, dateStyle)

		// Action requise
		cell = fmt.Sprintf("K%d", dataRow)
		f.SetCellValue("Alertes", cell, alert.ActionRequise)
		f.SetCellStyle("Alertes", cell, cell, dataStyle)

		// Personne responsable
		cell = fmt.Sprintf("L%d", dataRow)
		f.SetCellValue("Alertes", cell, alert.PersonneResponsable)
		f.SetCellStyle("Alertes", cell, cell, dataStyle)

		// Date de résolution
		cell = fmt.Sprintf("M%d", dataRow)
		if alert.DateResolution != nil {
			f.SetCellValue("Alertes", cell, alert.DateResolution.Format("02/01/2006 15:04"))
		} else {
			f.SetCellValue("Alertes", cell, "")
		}
		f.SetCellStyle("Alertes", cell, cell, dateStyle)

		// Commentaire de résolution
		cell = fmt.Sprintf("N%d", dataRow)
		f.SetCellValue("Alertes", cell, alert.CommentaireResolution)
		f.SetCellStyle("Alertes", cell, cell, dataStyle)

		// Date de création
		cell = fmt.Sprintf("O%d", dataRow)
		f.SetCellValue("Alertes", cell, alert.CreatedAt.Format("02/01/2006 15:04"))
		f.SetCellStyle("Alertes", cell, cell, dateStyle)

		// Date de MAJ
		cell = fmt.Sprintf("P%d", dataRow)
		f.SetCellValue("Alertes", cell, alert.UpdatedAt.Format("02/01/2006 15:04"))
		f.SetCellStyle("Alertes", cell, cell, dateStyle)

		// Définir la hauteur de ligne
		f.SetRowHeight("Alertes", dataRow, 25)
	}

	// ===== AJUSTEMENT DE LA LARGEUR DES COLONNES =====
	columnWidths := []float64{
		25, // UUID
		25, // Migrant UUID
		15, // Nom
		15, // Prénom
		15, // Type alerte
		15, // Niveau gravité
		30, // Titre
		50, // Description
		12, // Statut
		15, // Date expiration
		40, // Action requise
		20, // Personne responsable
		18, // Date résolution
		40, // Commentaire résolution
		18, // Date création
		18, // Date MAJ
	}

	columns := []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P"}
	for i, width := range columnWidths {
		if i < len(columns) {
			f.SetColWidth("Alertes", columns[i], columns[i], width)
		}
	}

	// ===== AJOUTER UNE FEUILLE DE STATISTIQUES =====
	_, err = f.NewSheet("Statistiques")
	if err == nil {
		// Calculer les statistiques
		totalRecords := len(alerts)

		// Compter par type d'alerte
		typeCount := make(map[string]int)
		graviteCount := make(map[string]int)
		statutCount := make(map[string]int)
		responsableCount := make(map[string]int)
		alertesExpireesCount := 0
		alertesResoluesCount := 0

		for _, alert := range alerts {
			typeCount[alert.TypeAlerte]++
			graviteCount[alert.NiveauGravite]++
			statutCount[alert.Statut]++
			if alert.PersonneResponsable != "" {
				responsableCount[alert.PersonneResponsable]++
			}
			if alert.DateExpiration != nil && alert.DateExpiration.Before(time.Now()) {
				alertesExpireesCount++
			}
			if alert.DateResolution != nil {
				alertesResoluesCount++
			}
		}

		// En-tête de la feuille statistiques
		f.SetCellValue("Statistiques", "A1", "STATISTIQUES DES ALERTES")
		f.MergeCell("Statistiques", "A1", "C1")
		f.SetCellStyle("Statistiques", "A1", "C1", headerStyle)

		row = 3
		f.SetCellValue("Statistiques", fmt.Sprintf("A%d", row), "Total des enregistrements:")
		f.SetCellValue("Statistiques", fmt.Sprintf("B%d", row), totalRecords)
		row += 2

		// Statistiques générales
		f.SetCellValue("Statistiques", fmt.Sprintf("A%d", row), "Alertes résolues:")
		f.SetCellValue("Statistiques", fmt.Sprintf("B%d", row), alertesResoluesCount)
		row++
		f.SetCellValue("Statistiques", fmt.Sprintf("A%d", row), "Alertes expirées:")
		f.SetCellValue("Statistiques", fmt.Sprintf("B%d", row), alertesExpireesCount)
		row += 2

		// Par type d'alerte
		f.SetCellValue("Statistiques", fmt.Sprintf("A%d", row), "Par type d'alerte:")
		f.SetCellStyle("Statistiques", fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), columnHeaderStyle)
		row++
		for typeAlert, count := range typeCount {
			f.SetCellValue("Statistiques", fmt.Sprintf("A%d", row), typeAlert)
			f.SetCellValue("Statistiques", fmt.Sprintf("B%d", row), count)
			row++
		}
		row++

		// Par niveau de gravité
		f.SetCellValue("Statistiques", fmt.Sprintf("A%d", row), "Par niveau de gravité:")
		f.SetCellStyle("Statistiques", fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), columnHeaderStyle)
		row++
		for gravite, count := range graviteCount {
			f.SetCellValue("Statistiques", fmt.Sprintf("A%d", row), gravite)
			f.SetCellValue("Statistiques", fmt.Sprintf("B%d", row), count)
			row++
		}
		row++

		// Par statut
		f.SetCellValue("Statistiques", fmt.Sprintf("A%d", row), "Par statut:")
		f.SetCellStyle("Statistiques", fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), columnHeaderStyle)
		row++
		for statut, count := range statutCount {
			f.SetCellValue("Statistiques", fmt.Sprintf("A%d", row), statut)
			f.SetCellValue("Statistiques", fmt.Sprintf("B%d", row), count)
			row++
		}
		row++

		// Top 10 personnes responsables
		f.SetCellValue("Statistiques", fmt.Sprintf("A%d", row), "Top 10 personnes responsables:")
		f.SetCellStyle("Statistiques", fmt.Sprintf("A%d", row), fmt.Sprintf("A%d", row), columnHeaderStyle)
		row++
		count := 0
		for responsable, nb := range responsableCount {
			if count >= 10 {
				break
			}
			f.SetCellValue("Statistiques", fmt.Sprintf("A%d", row), responsable)
			f.SetCellValue("Statistiques", fmt.Sprintf("B%d", row), nb)
			row++
			count++
		}

		f.SetColWidth("Statistiques", "A", "A", 30)
		f.SetColWidth("Statistiques", "B", "B", 15)
	}

	// ===== GÉNÉRATION DU FICHIER =====
	filename := fmt.Sprintf("alertes_export_%s.xlsx", time.Now().Format("20060102_150405"))

	// Sauvegarder en mémoire
	buffer, err := f.WriteToBuffer()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to generate Excel file",
			"error":   err.Error(),
		})
	}

	// Configurer les en-têtes de réponse pour le téléchargement
	c.Set("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
	c.Set("Content-Length", strconv.Itoa(len(buffer.Bytes())))

	return c.Send(buffer.Bytes())
}
