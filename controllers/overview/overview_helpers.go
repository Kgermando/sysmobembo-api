package overview

import (
	"strings"
	"time"

	"github.com/kgermando/sysmobembo-api/database"
	"github.com/kgermando/sysmobembo-api/models"
)

// =================== FONCTIONS HELPER ===================

// 🧍‍♂️ INDICATEURS DE VOLUME ET LOCALISATION
func getVolumeLocalisationIndicateurs(periode int, province string) VolumeLocalisationIndicateurs {
	db := database.DB
	dateDebut := time.Now().AddDate(0, -periode, 0)

	// Nombre total de migrants
	var totalMigrants int64
	query := db.Model(&models.Migrant{}).Where("created_at >= ?", dateDebut)
	if province != "" {
		query = query.Where("ville_actuelle = ? OR pays_actuel LIKE ?", province, "%"+province+"%")
	}
	query.Count(&totalMigrants)

	// Nombre de déplacés internes (migrants qui ont changé de province/ville dans le même pays)
	var deplacesInternes int64
	deplacesQuery := db.Table("migrants m").
		Joins("JOIN identites i ON m.identite_uuid = i.uuid").
		Where("m.created_at >= ? AND i.nationalite = m.pays_actuel AND i.lieu_naissance != m.ville_actuelle",
			dateDebut)
	if province != "" {
		deplacesQuery = deplacesQuery.Where("m.ville_actuelle = ? OR m.pays_actuel LIKE ?", province, "%"+province+"%")
	}
	deplacesQuery.Count(&deplacesInternes)

	// Le nombre total PDI correspond au nombre total de migrants
	totalPDI := totalMigrants

	// Personnes retournées (approximation basée sur les géolocalisations récentes)
	var personnesRetournees int64
	geoQuery := db.Table("geolocalisations g").
		Joins("JOIN migrants m ON g.identite_uuid = m.identite_uuid").
		Where("g.created_at >= ?", dateDebut)
	if province != "" {
		geoQuery = geoQuery.Where("m.ville_actuelle = ?", province)
	}
	geoQuery.Count(&personnesRetournees)

	// Répartition géographique par province
	repartitionGeo := getRepartitionGeographique(periode, province)

	// Évolution mensuelle
	evolutionMensuelle := getEvolutionMensuelle(periode, province)

	return VolumeLocalisationIndicateurs{
		NombreTotalPDI:          totalPDI,
		NombreTotalMigrants:     totalMigrants,
		NombreDeplacesInternes:  deplacesInternes,
		PersonnesRetournees:     personnesRetournees,
		RepartitionGeographique: repartitionGeo,
		EvolutionMensuelle:      evolutionMensuelle,
	}
}

func getRepartitionGeographique(periode int, province string) []RepartitionProvinceStats {
	db := database.DB
	dateDebut := time.Now().AddDate(0, -periode, 0)

	var results []struct {
		Province string `json:"province"`
		Count    int64  `json:"count"`
	}

	query := db.Table("migrants").
		Select("ville_actuelle as province, COUNT(*) as count").
		Where("created_at >= ?", dateDebut).
		Group("ville_actuelle").
		Order("count DESC")

	if province != "" {
		query = query.Where("ville_actuelle = ?", province)
	}

	query.Scan(&results)

	// Calculer le total pour les pourcentages
	var total int64
	for _, result := range results {
		total += result.Count
	}

	var repartition []RepartitionProvinceStats
	for _, result := range results {
		pourcentage := float64(0)
		if total > 0 {
			pourcentage = float64(result.Count) / float64(total) * 100
		}
		repartition = append(repartition, RepartitionProvinceStats{
			Province:    result.Province,
			NombrePDI:   result.Count,
			Pourcentage: pourcentage,
		})
	}

	return repartition
}

func getEvolutionMensuelle(periode int, province string) []EvolutionTemporelleStats {
	db := database.DB
	var evolution []EvolutionTemporelleStats

	for i := periode - 1; i >= 0; i-- {
		debutMois := time.Now().AddDate(0, -i-1, 0)
		finMois := time.Now().AddDate(0, -i, 0)

		// Nouveaux déplacés ce mois
		var nouveauxDeplaces int64
		query := db.Model(&models.Migrant{}).
			Where("created_at >= ? AND created_at < ?", debutMois, finMois)
		if province != "" {
			query = query.Where("ville_actuelle = ?", province)
		}
		query.Count(&nouveauxDeplaces)

		// Retours ce mois (approximation basée sur géolocalisations)
		var retours int64
		geoQuery := db.Table("geolocalisations g").
			Joins("JOIN migrants m ON g.identite_uuid = m.identite_uuid").
			Where("g.created_at >= ? AND g.created_at < ?", debutMois, finMois)
		if province != "" {
			geoQuery = geoQuery.Where("m.ville_actuelle = ?", province)
		}
		geoQuery.Count(&retours)

		// Total cumulé jusqu'à cette date
		var totalCumule int64
		cumulQuery := db.Model(&models.Migrant{}).
			Where("created_at < ?", finMois)
		if province != "" {
			cumulQuery = cumulQuery.Where("ville_actuelle = ?", province)
		}
		cumulQuery.Count(&totalCumule)

		periodeStr := debutMois.Format("2006-01")
		evolution = append(evolution, EvolutionTemporelleStats{
			Periode:          periodeStr,
			NouveauxDeplaces: nouveauxDeplaces,
			Retours:          retours,
			TotalCumule:      totalCumule,
		})
	}

	return evolution
}

// 🔥 INDICATEURS DES CAUSES DE DÉPLACEMENT
func getCausesDeplacementsIndicateurs(periode int, province string) CausesDeplacementsIndicateurs {
	db := database.DB
	dateDebut := time.Now().AddDate(0, -periode, 0)

	// Total des motifs de déplacement dans la période
	var totalMotifs int64
	motifQuery := db.Table("motif_deplacements md").
		Joins("JOIN migrants m ON md.migrant_uuid = m.uuid").
		Where("md.created_at >= ?", dateDebut)
	if province != "" {
		motifQuery = motifQuery.Where("m.ville_actuelle = ?", province)
	}
	motifQuery.Count(&totalMotifs)

	// Compter par type de motif
	var results []struct {
		TypeMotif       string `json:"type_motif"`
		MotifPrincipal  string `json:"motif_principal"`
		MotifSecondaire string `json:"motif_secondaire"`
		Description     string `json:"description"`
		Count           int64  `json:"count"`
	}

	detailQuery := db.Table("motif_deplacements md").
		Select("md.type_motif, md.motif_principal, md.motif_secondaire, md.description, COUNT(*) as count").
		Joins("JOIN migrants m ON md.migrant_uuid = m.uuid").
		Where("md.created_at >= ?", dateDebut).
		Group("md.type_motif, md.motif_principal, md.motif_secondaire, md.description")
	if province != "" {
		detailQuery = detailQuery.Where("m.ville_actuelle = ?", province)
	}
	detailQuery.Scan(&results)

	// Calculer les pourcentages par catégorie
	var conflitsArmes, catastrophes, persecution, violenceGen, autresCauses float64
	var detailsCauses []CauseDetailStats

	for _, result := range results {
		pourcentage := float64(0)
		if totalMotifs > 0 {
			pourcentage = float64(result.Count) / float64(totalMotifs) * 100
		}

		detailsCauses = append(detailsCauses, CauseDetailStats{
			TypeMotif:   result.TypeMotif,
			NombreCas:   result.Count,
			Pourcentage: pourcentage,
		})

		// Catégoriser selon le type de motif
		switch result.TypeMotif {
		case "conflit_arme", "guerre", "violence_politique":
			conflitsArmes += pourcentage
		case "catastrophe_naturelle", "inondation", "secheresse", "tremblement_terre":
			catastrophes += pourcentage
		case "persecution_religieuse", "persecution_ethnique", "persecution_politique":
			persecution += pourcentage
		case "violence_generalisee", "insecurite", "criminalite":
			violenceGen += pourcentage
		default:
			autresCauses += pourcentage
		}
	}

	return CausesDeplacementsIndicateurs{
		PourcentageConflitsArmes:       conflitsArmes,
		PourcentageCatastrophes:        catastrophes,
		PourcentagePersecution:         persecution,
		PourcentageViolenceGeneralisee: violenceGen,
		PourcentageAutresCauses:        autresCauses,
		DetailsCauses:                  detailsCauses,
	}
}

// 👥 INDICATEURS DE VULNÉRABILITÉ ET BESOINS
func getVulnerabiliteBesoinsIndicateurs(periode int, province string) VulnerabiliteBesoinsIndicateurs {
	db := database.DB
	dateDebut := time.Now().AddDate(0, -periode, 0)

	// Profil démographique
	profilDemo := getProfilDemographique(periode, province)

	// Accès aux services de base (données simulées car pas dans le modèle actuel)
	accesServices := AccesServicesStats{
		AccesEau:       75.5, // À remplacer par des vraies données si disponibles
		AccesSante:     68.2,
		AccesEducation: 82.3,
		AccesLogement:  58.7,
	}

	// Taux d'occupation des sites et déplacés hors sites (approximation)
	var deplacesHorsSites int64
	horsQuery := db.Table("geolocalisations g").
		Joins("JOIN migrants m ON g.identite_uuid = m.identite_uuid").
		Where("g.created_at >= ?", dateDebut)
	if province != "" {
		horsQuery = horsQuery.Where("m.ville_actuelle = ?", province)
	}
	horsQuery.Count(&deplacesHorsSites)

	var totalDansStructures int64
	structuresQuery := db.Table("geolocalisations g").
		Joins("JOIN migrants m ON g.identite_uuid = m.identite_uuid").
		Where("g.created_at >= ?", dateDebut)
	if province != "" {
		structuresQuery = structuresQuery.Where("m.ville_actuelle = ?", province)
	}
	structuresQuery.Count(&totalDansStructures)

	tauxOccupation := float64(0)
	if totalDansStructures > 0 {
		tauxOccupation = float64(totalDansStructures-deplacesHorsSites) / float64(totalDansStructures) * 100
	}

	return VulnerabiliteBesoinsIndicateurs{
		ProfilDemographique: profilDemo,
		AccesServicesBase:   accesServices,
		TauxOccupationSites: tauxOccupation,
		DeplacesHorsSites:   deplacesHorsSites,
	}
}

func getProfilDemographique(periode int, province string) ProfilDemographiqueStats {
	db := database.DB
	dateDebut := time.Now().AddDate(0, -periode, 0)

	var totalMigrants int64
	var femmes, enfants, ages int64
	var ageTotal float64

	// Requête de base avec les filtres de période et province
	baseQuery := db.Model(&models.Migrant{}).Where("created_at >= ?", dateDebut)
	if province != "" {
		baseQuery = baseQuery.Where("ville_actuelle = ?", province)
	}

	// Total des migrants
	baseQuery.Count(&totalMigrants)

	// Compter les femmes - Utiliser JOIN avec identites
	femmeQuery := db.Table("migrants m").
		Joins("JOIN identites i ON m.identite_uuid = i.uuid").
		Where("m.created_at >= ? AND i.sexe = ?", dateDebut, "F")
	if province != "" {
		femmeQuery = femmeQuery.Where("m.ville_actuelle = ?", province)
	}
	femmeQuery.Count(&femmes)

	// Compter les hommes - Utiliser JOIN avec identites
	var hommes int64
	hommeQuery := db.Table("migrants m").
		Joins("JOIN identites i ON m.identite_uuid = i.uuid").
		Where("m.created_at >= ? AND i.sexe = ?", dateDebut, "M")
	if province != "" {
		hommeQuery = hommeQuery.Where("m.ville_actuelle = ?", province)
	}
	hommeQuery.Count(&hommes)

	// Compter les enfants (moins de 18 ans)
	dateMineure := time.Now().AddDate(-18, 0, 0)
	enfantQuery := db.Table("migrants m").
		Joins("JOIN identites i ON m.identite_uuid = i.uuid").
		Where("m.created_at >= ? AND i.date_naissance > ?", dateDebut, dateMineure)
	if province != "" {
		enfantQuery = enfantQuery.Where("m.ville_actuelle = ?", province)
	}
	enfantQuery.Count(&enfants)

	// Compter les personnes âgées (plus de 65 ans)
	dateAgee := time.Now().AddDate(-65, 0, 0)
	ageQuery := db.Table("migrants m").
		Joins("JOIN identites i ON m.identite_uuid = i.uuid").
		Where("m.created_at >= ? AND i.date_naissance < ?", dateDebut, dateAgee)
	if province != "" {
		ageQuery = ageQuery.Where("m.ville_actuelle = ?", province)
	}
	ageQuery.Count(&ages)

	// Calculer l'âge moyen - Utiliser la requête de base correcte
	var migrants []models.Migrant
	migrantsQuery := db.Model(&models.Migrant{}).Where("created_at >= ?", dateDebut)
	if province != "" {
		migrantsQuery = migrantsQuery.Where("ville_actuelle = ?", province)
	}
	migrantsQuery.Preload("Identite").Find(&migrants)

	if len(migrants) > 0 {
		for _, migrant := range migrants {
			if migrant.Identite.DateNaissance.Year() > 0 {
				age := float64(time.Now().Year() - migrant.Identite.DateNaissance.Year())
				ageTotal += age
			}
		}
		ageTotal = ageTotal / float64(len(migrants))
	}

	pourcentageFemmes := float64(0)
	pourcentageHommes := float64(0)
	pourcentageEnfants := float64(0)
	pourcentageAges := float64(0)

	if totalMigrants > 0 {
		pourcentageFemmes = float64(femmes) / float64(totalMigrants) * 100
		pourcentageHommes = float64(hommes) / float64(totalMigrants) * 100
		pourcentageEnfants = float64(enfants) / float64(totalMigrants) * 100
		pourcentageAges = float64(ages) / float64(totalMigrants) * 100
	}

	return ProfilDemographiqueStats{
		PourcentageFemmes:  pourcentageFemmes,
		PourcentageHommes:  pourcentageHommes,
		PourcentageEnfants: pourcentageEnfants,
		PourcentageAges:    pourcentageAges,
		AgeMoyen:           ageTotal,
	}
} // ⚠️ INDICATEURS DYNAMIQUES ET D'ALERTE
func getDynamiquesAlerteIndicateurs(periode int, province string) DynamiquesAlerteIndicateurs {
	db := database.DB

	// Zones à haut risque
	zonesRisque := getZonesHautRisque(periode, province)

	// Tendances de retour
	tendancesRetour := getTendancesRetour(periode, province)

	// Alertes précoces
	alertesPrecoces := getAlertesPrecoces(periode, province)

	// Mouvements massifs récents (30 derniers jours)
	var mouvementsMassifs int64
	date30Jours := time.Now().AddDate(0, 0, -30)
	massifQuery := db.Model(&models.Migrant{}).
		Where("created_at >= ?", date30Jours)
	if province != "" {
		massifQuery = massifQuery.Where("ville_actuelle = ?", province)
	}
	massifQuery.Count(&mouvementsMassifs)

	return DynamiquesAlerteIndicateurs{
		ZonesHautRisque:   zonesRisque,
		TendancesRetour:   tendancesRetour,
		AlertesPrecoces:   alertesPrecoces,
		MouvementsMassifs: mouvementsMassifs,
	}
}

func getZonesHautRisque(periode int, province string) []ZoneRisqueStats {
	db := database.DB
	dateDebut := time.Now().AddDate(0, -periode, 0)

	var results []struct {
		Zone  string `json:"zone"`
		Count int64  `json:"count"`
	}

	query := db.Table("alertes a").
		Select("m.ville_actuelle as zone, COUNT(*) as count").
		Joins("JOIN migrants m ON a.migrant_uuid = m.uuid").
		Where("a.niveau_gravite IN (?) AND a.created_at >= ? AND a.statut = ?",
			[]string{"danger", "critical"}, dateDebut, "active").
		Group("zone").
		Order("count DESC").
		Limit(10)

	if province != "" {
		query = query.Where("m.ville_actuelle = ?", province)
	}

	query.Scan(&results)

	var zones []ZoneRisqueStats
	for _, result := range results {
		// Déterminer le niveau de risque selon le nombre d'alertes
		niveauRisque := "MOYEN"
		if result.Count >= 50 {
			niveauRisque = "CRITIQUE"
		} else if result.Count >= 20 {
			niveauRisque = "ÉLEVÉ"
		}

		zones = append(zones, ZoneRisqueStats{
			Zone:           result.Zone,
			NiveauRisque:   niveauRisque,
			TypeMenace:     "MULTIPLE",        // À détailler selon les alertes
			PopulationRisk: result.Count * 10, // Estimation
		})
	}

	return zones
}

func getTendancesRetour(periode int, province string) []TendanceRetourStats {
	db := database.DB
	dateDebut := time.Now().AddDate(0, -periode, 0)

	var results []struct {
		ZoneOrigine string `json:"zone_origine"`
		ZoneRetour  string `json:"zone_retour"`
		Count       int64  `json:"count"`
	}

	query := db.Table("geolocalisations g").
		Select("i.lieu_naissance as zone_origine, m.ville_actuelle as zone_retour, COUNT(*) as count").
		Joins("JOIN migrants m ON g.identite_uuid = m.identite_uuid").
		Joins("JOIN identites i ON m.identite_uuid = i.uuid").
		Where("g.created_at >= ?", dateDebut).
		Group("zone_origine, zone_retour").
		Order("count DESC").
		Limit(10)

	if province != "" {
		query = query.Where("m.ville_actuelle = ?", province)
	}

	query.Scan(&results)

	var tendances []TendanceRetourStats
	for _, result := range results {
		tendanceEvol := "STABLE"
		if result.Count >= 100 {
			tendanceEvol = "HAUSSE"
		} else if result.Count <= 10 {
			tendanceEvol = "BAISSE"
		}

		tendances = append(tendances, TendanceRetourStats{
			ZoneOrigine:   result.ZoneOrigine,
			ZoneRetour:    result.ZoneRetour,
			NombreRetours: result.Count,
			TendanceEvol:  tendanceEvol,
		})
	}

	return tendances
}

func getAlertesPrecoces(periode int, province string) []AlertePrecoceStats {
	db := database.DB
	dateDebut := time.Now().AddDate(0, -periode, 0)

	var alertes []models.Alert
	query := db.Where("created_at >= ? AND statut = ?", dateDebut, "active").
		Order("created_at DESC").
		Limit(20).
		Preload("Migrant")

	query.Find(&alertes)

	var alertesStats []AlertePrecoceStats
	for _, alerte := range alertes {
		// Filtrer par province si spécifiée
		if province != "" && alerte.Migrant.VilleActuelle != province {
			continue
		}

		alertesStats = append(alertesStats, AlertePrecoceStats{
			Zone:          alerte.Migrant.VilleActuelle,
			TypeAlerte:    alerte.TypeAlerte,
			NiveauGravite: alerte.NiveauGravite,
			DateDetection: alerte.CreatedAt,
			Description:   alerte.Description,
		})
	}

	return alertesStats
}

// Fonction pour récupérer les données du pie chart des motifs de déplacement
func getMotifsPieChartData(periode int, province string) []ChartDataPoint {
	db := database.DB
	dateDebut := time.Now().AddDate(0, -periode, 0)

	// Compter par type de motif uniquement
	var results []struct {
		TypeMotif string `json:"type_motif"`
		Count     int64  `json:"count"`
	}

	query := db.Table("motif_deplacements md").
		Select("md.type_motif, COUNT(*) as count").
		Joins("JOIN migrants m ON md.migrant_uuid = m.uuid").
		Where("md.created_at >= ?", dateDebut).
		Group("md.type_motif").
		Order("count DESC")

	if province != "" {
		query = query.Where("m.ville_actuelle = ?", province)
	}

	query.Scan(&results)

	// Transformer en ChartDataPoint avec labels en français
	var pieData []ChartDataPoint
	motifLabels := map[string]string{
		"economique":            "Économique",
		"politique":             "Politique",
		"persecution":           "Persécution",
		"naturelle":             "Catastrophe Naturelle",
		"familial":              "Familial",
		"education":             "Éducation",
		"sanitaire":             "Sanitaire",
		"conflit_arme":          "Conflit Armé",
		"catastrophe_naturelle": "Catastrophe Naturelle",
		"violence_generalisee":  "Violence Généralisée",
	}

	for _, result := range results {
		label := motifLabels[result.TypeMotif]
		if label == "" {
			label = result.TypeMotif // Utiliser la valeur brute si pas de traduction
		}

		pieData = append(pieData, ChartDataPoint{
			Name:  label,
			Value: float64(result.Count),
			Extra: result.TypeMotif, // Garder la valeur originale en extra
		})
	}

	return pieData
}

// Fonction pour récupérer les alertes récentes (utilisée par GetAlertesTempsReel)
func getAlertesRecentes(niveaux, province string, jours int) []AlertePrecoceStats {
	db := database.DB
	dateDebut := time.Now().AddDate(0, 0, -jours)

	// Parser les niveaux
	niveauxList := []string{"danger", "critical"}
	if niveaux != "" {
		// Diviser la chaîne par virgule et nettoyer
		niveauxSplit := strings.Split(niveaux, ",")
		niveauxList = []string{}
		for _, niveau := range niveauxSplit {
			niveau = strings.TrimSpace(niveau)
			if niveau != "" {
				niveauxList = append(niveauxList, niveau)
			}
		}
	}

	var alertes []models.Alert
	query := db.Where("created_at >= ? AND statut = ? AND niveau_gravite IN (?)",
		dateDebut, "active", niveauxList).
		Order("created_at DESC").
		Preload("Migrant")

	query.Find(&alertes)

	var alertesStats []AlertePrecoceStats
	for _, alerte := range alertes {
		// Filtrer par province si spécifiée
		if province != "" && alerte.Migrant.VilleActuelle != province {
			continue
		}

		alertesStats = append(alertesStats, AlertePrecoceStats{
			Zone:          alerte.Migrant.VilleActuelle,
			TypeAlerte:    alerte.TypeAlerte,
			NiveauGravite: alerte.NiveauGravite,
			DateDetection: alerte.CreatedAt,
			Description:   alerte.Description,
		})
	}

	return alertesStats
}
