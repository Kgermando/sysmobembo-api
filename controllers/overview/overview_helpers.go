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

	// Nombre total de migrants (tous les migrants actifs)
	var totalMigrants int64
	query := db.Model(&models.Migrant{}).Where("actif = ? AND created_at >= ?", true, dateDebut)
	if province != "" {
		query = query.Where("ville_actuelle = ? OR pays_actuel LIKE ?", province, "%"+province+"%")
	}
	query.Count(&totalMigrants)

	// Nombre de déplacés internes (migrants qui ont changé de province/ville dans le même pays)
	var deplacesInternes int64
	deplacesQuery := db.Model(&models.Migrant{}).
		Where("actif = ? AND created_at >= ? AND pays_origine = pays_actuel AND lieu_naissance != ville_actuelle",
			true, dateDebut)
	if province != "" {
		deplacesQuery = deplacesQuery.Where("ville_actuelle = ? OR pays_actuel LIKE ?", province, "%"+province+"%")
	}
	deplacesQuery.Count(&deplacesInternes)

	// Le nombre total PDI correspond au nombre total de migrants
	totalPDI := totalMigrants

	// Personnes retournées (basé sur type_mouvement = "residence_permanente" dans geolocalisation)
	var personnesRetournees int64
	geoQuery := db.Table("geolocalisations g").
		Joins("JOIN migrants m ON g.migrant_uuid = m.uuid").
		Where("g.type_mouvement = ? AND g.created_at >= ? AND m.actif = ?", "residence_permanente", dateDebut, true)
	if province != "" {
		geoQuery = geoQuery.Where("g.ville = ? OR g.pays LIKE ?", province, "%"+province+"%")
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
		Where("actif = ? AND created_at >= ?", true, dateDebut).
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
			Where("actif = ? AND created_at >= ? AND created_at < ?", true, debutMois, finMois)
		if province != "" {
			query = query.Where("ville_actuelle = ?", province)
		}
		query.Count(&nouveauxDeplaces)

		// Retours ce mois
		var retours int64
		geoQuery := db.Table("geolocalisations g").
			Joins("JOIN migrants m ON g.migrant_uuid = m.uuid").
			Where("g.type_mouvement = ? AND g.created_at >= ? AND g.created_at < ? AND m.actif = ?",
				"residence_permanente", debutMois, finMois, true)
		if province != "" {
			geoQuery = geoQuery.Where("g.ville = ?", province)
		}
		geoQuery.Count(&retours)

		// Total cumulé jusqu'à cette date
		var totalCumule int64
		cumulQuery := db.Model(&models.Migrant{}).
			Where("actif = ? AND created_at < ?", true, finMois)
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
		Where("md.created_at >= ? AND m.actif = ?", dateDebut, true)
	if province != "" {
		motifQuery = motifQuery.Where("m.ville_actuelle = ?", province)
	}
	motifQuery.Count(&totalMotifs)

	// Compter par type de motif
	var results []struct {
		TypeMotif string `json:"type_motif"`
		Count     int64  `json:"count"`
	}

	detailQuery := db.Table("motif_deplacements md").
		Select("md.type_motif, COUNT(*) as count").
		Joins("JOIN migrants m ON md.migrant_uuid = m.uuid").
		Where("md.created_at >= ? AND m.actif = ?", dateDebut, true).
		Group("md.type_motif")
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

	// Taux d'occupation des sites et déplacés hors sites
	var deplacesHorsSites int64
	horsQuery := db.Table("geolocalisations g").
		Joins("JOIN migrants m ON g.migrant_uuid = m.uuid").
		Where("g.type_hebergement != ? AND g.created_at >= ? AND m.actif = ?", "site_officiel", dateDebut, true)
	if province != "" {
		horsQuery = horsQuery.Where("g.ville = ?", province)
	}
	horsQuery.Count(&deplacesHorsSites)

	var totalDansStructures int64
	structuresQuery := db.Table("geolocalisations g").
		Joins("JOIN migrants m ON g.migrant_uuid = m.uuid").
		Where("g.created_at >= ? AND m.actif = ?", dateDebut, true)
	if province != "" {
		structuresQuery = structuresQuery.Where("g.ville = ?", province)
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

	baseQuery := db.Model(&models.Migrant{}).Where("actif = ? AND created_at >= ?", true, dateDebut)
	if province != "" {
		baseQuery = baseQuery.Where("ville_actuelle = ?", province)
	}

	// Total des migrants
	baseQuery.Count(&totalMigrants)

	// Compter les femmes
	femmeQuery := baseQuery.Where("sexe = ?", "F")
	femmeQuery.Count(&femmes)

	// Compter les enfants (moins de 18 ans)
	dateMineure := time.Now().AddDate(-18, 0, 0)
	enfantQuery := baseQuery.Where("date_naissance > ?", dateMineure)
	enfantQuery.Count(&enfants)

	// Compter les personnes âgées (plus de 65 ans)
	dateAgee := time.Now().AddDate(-65, 0, 0)
	ageQuery := baseQuery.Where("date_naissance < ?", dateAgee)
	ageQuery.Count(&ages)

	// Calculer l'âge moyen
	var migrants []models.Migrant
	baseQuery.Select("date_naissance").Find(&migrants)

	if len(migrants) > 0 {
		for _, migrant := range migrants {
			age := float64(time.Now().Year() - migrant.DateNaissance.Year())
			ageTotal += age
		}
		ageTotal = ageTotal / float64(len(migrants))
	}

	pourcentageFemmes := float64(0)
	pourcentageEnfants := float64(0)
	pourcentageAges := float64(0)

	if totalMigrants > 0 {
		pourcentageFemmes = float64(femmes) / float64(totalMigrants) * 100
		pourcentageEnfants = float64(enfants) / float64(totalMigrants) * 100
		pourcentageAges = float64(ages) / float64(totalMigrants) * 100
	}

	return ProfilDemographiqueStats{
		PourcentageFemmes:  pourcentageFemmes,
		PourcentageEnfants: pourcentageEnfants,
		PourcentageAges:    pourcentageAges,
		AgeMoyen:           ageTotal,
	}
}

// ⚠️ INDICATEURS DYNAMIQUES ET D'ALERTE
func getDynamiquesAlerteIndicateurs(periode int, province string) DynamiquesAlerteIndicateurs {
	// Zones à haut risque
	zonesRisque := getZonesHautRisque(periode, province)

	// Tendances de retour
	tendancesRetour := getTendancesRetour(periode, province)

	// Alertes précoces
	alertesPrecoces := getAlertesPrecoces(periode, province)

	// Mouvements massifs récents (30 derniers jours)
	var mouvementsMassifs int64
	db := database.DB
	date30Jours := time.Now().AddDate(0, 0, -30)
	massifQuery := db.Model(&models.Migrant{}).
		Where("actif = ? AND created_at >= ?", true, date30Jours)
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
		Select("COALESCE(g.ville, m.ville_actuelle) as zone, COUNT(*) as count").
		Joins("JOIN migrants m ON a.migrant_uuid = m.uuid").
		Joins("LEFT JOIN geolocalisations g ON g.migrant_uuid = m.uuid").
		Where("a.niveau_gravite IN (?) AND a.created_at >= ? AND a.statut = ?",
			[]string{"danger", "critical"}, dateDebut, "active").
		Group("zone").
		Order("count DESC").
		Limit(10)

	if province != "" {
		query = query.Where("COALESCE(g.ville, m.ville_actuelle) = ?", province)
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
		Select("m.lieu_naissance as zone_origine, g.ville as zone_retour, COUNT(*) as count").
		Joins("JOIN migrants m ON g.migrant_uuid = m.uuid").
		Where("g.type_mouvement = ? AND g.created_at >= ? AND m.actif = ?",
			"residence_permanente", dateDebut, true).
		Group("zone_origine, zone_retour").
		Order("count DESC").
		Limit(10)

	if province != "" {
		query = query.Where("g.ville = ?", province)
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
