package dashboard

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/kgermando/sysmobembo-api/database"
	"github.com/kgermando/sysmobembo-api/models"
)

// Structure pour les indicateurs de déplacement
type IndicateursDeplacementResponse struct {
	// Indicateurs de volume et localisation
	VolumeLocalisation VolumeLocalisationIndicateurs `json:"volume_localisation"`

	// Indicateurs des causes de déplacement
	CausesDeplacements CausesDeplacementsIndicateurs `json:"causes_deplacements"`

	// Indicateurs de vulnérabilité et besoins
	VulnerabiliteBesoins VulnerabiliteBesoinsIndicateurs `json:"vulnerabilite_besoins"`

	// Indicateurs dynamiques et d'alerte
	DynamiquesAlerte DynamiquesAlerteIndicateurs `json:"dynamiques_alerte"`

	// Métadonnées
	DateGeneration time.Time `json:"date_generation"`
	PeriodeAnalyse string    `json:"periode_analyse"`
}

type VolumeLocalisationIndicateurs struct {
	NombreTotalPDI          int64                      `json:"nombre_total_pdi"`
	PersonnesRetournees     int64                      `json:"personnes_retournees"`
	RepartitionGeographique []RepartitionProvinceStats `json:"repartition_geographique"`
	EvolutionMensuelle      []EvolutionTemporelleStats `json:"evolution_mensuelle"`
}

type CausesDeplacementsIndicateurs struct {
	PourcentageConflitsArmes float64            `json:"pourcentage_conflits_armes"`
	PourcentageCatastrophes  float64            `json:"pourcentage_catastrophes"`
	PourcentagePersecution   float64            `json:"pourcentage_persecution"`
	PourcentageViolenceGen   float64            `json:"pourcentage_violence_generalisee"`
	PourcentageAutresCauses  float64            `json:"pourcentage_autres_causes"`
	DetailsCauses            []CauseDetailStats `json:"details_causes"`
}

type VulnerabiliteBesoinsIndicateurs struct {
	ProfilDemographique ProfilDemographiqueStats `json:"profil_demographique"`
	AccesServicesBase   AccesServicesStats       `json:"acces_services_base"`
	TauxOccupationSites float64                  `json:"taux_occupation_sites"`
	DeplacesHorsSites   int64                    `json:"deplaces_hors_sites"`
}

type DynamiquesAlerteIndicateurs struct {
	ZonesHautRisque   []ZoneRisqueStats     `json:"zones_haut_risque"`
	TendancesRetour   []TendanceRetourStats `json:"tendances_retour"`
	AlertesPrecoces   []AlertePrecoceStats  `json:"alertes_precoces"`
	MouvementsMassifs int64                 `json:"mouvements_massifs_recent"`
}

// Structures de support
type RepartitionProvinceStats struct {
	Province    string  `json:"province"`
	NombrePDI   int64   `json:"nombre_pdi"`
	Pourcentage float64 `json:"pourcentage"`
}

type EvolutionTemporelleStats struct {
	Periode          string `json:"periode"`
	NouveauxDeplaces int64  `json:"nouveaux_deplaces"`
	Retours          int64  `json:"retours"`
	TotalCumule      int64  `json:"total_cumule"`
}

type CauseDetailStats struct {
	TypeMotif   string  `json:"type_motif"`
	NombreCas   int64   `json:"nombre_cas"`
	Pourcentage float64 `json:"pourcentage"`
}

type ProfilDemographiqueStats struct {
	PourcentageFemmes  float64 `json:"pourcentage_femmes"`
	PourcentageEnfants float64 `json:"pourcentage_enfants"`
	PourcentageAges    float64 `json:"pourcentage_ages"`
	AgeMoyen           float64 `json:"age_moyen"`
}

type AccesServicesStats struct {
	AccesEau       float64 `json:"acces_eau"`
	AccesSante     float64 `json:"acces_sante"`
	AccesEducation float64 `json:"acces_education"`
	AccesLogement  float64 `json:"acces_logement"`
}

type ZoneRisqueStats struct {
	Zone           string `json:"zone"`
	NiveauRisque   string `json:"niveau_risque"`
	TypeMenace     string `json:"type_menace"`
	PopulationRisk int64  `json:"population_risque"`
}

type TendanceRetourStats struct {
	ZoneOrigine   string `json:"zone_origine"`
	ZoneRetour    string `json:"zone_retour"`
	NombreRetours int64  `json:"nombre_retours"`
	TendanceEvol  string `json:"tendance_evolution"`
}

type AlertePrecoceStats struct {
	Zone          string    `json:"zone"`
	TypeAlerte    string    `json:"type_alerte"`
	NiveauGravite string    `json:"niveau_gravite"`
	DateDetection time.Time `json:"date_detection"`
	Description   string    `json:"description"`
}

// Fonction principale pour récupérer tous les indicateurs
func AnalyseDeplacement(c *fiber.Ctx) error {
	// Paramètres optionnels
	periode := c.Query("periode", "12") // 12 derniers mois par défaut
	province := c.Query("province")     // Province spécifique si demandée

	periodeInt, err := strconv.Atoi(periode)
	if err != nil {
		periodeInt = 12
	}

	// Générer tous les indicateurs
	volumeLocalisation := getVolumeLocalisationIndicateurs(periodeInt, province)
	causesDeplacements := getCausesDeplacementsIndicateurs(periodeInt, province)
	vulnerabiliteBesoins := getVulnerabiliteBesoinsIndicateurs(periodeInt, province)
	dynamiquesAlerte := getDynamiquesAlerteIndicateurs(periodeInt, province)

	response := IndicateursDeplacementResponse{
		VolumeLocalisation:   volumeLocalisation,
		CausesDeplacements:   causesDeplacements,
		VulnerabiliteBesoins: vulnerabiliteBesoins,
		DynamiquesAlerte:     dynamiquesAlerte,
		DateGeneration:       time.Now(),
		PeriodeAnalyse:       strconv.Itoa(periodeInt) + " derniers mois",
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// 🧍‍♂️ INDICATEURS DE VOLUME ET LOCALISATION
func getVolumeLocalisationIndicateurs(periode int, province string) VolumeLocalisationIndicateurs {
	db := database.DB
	dateDebut := time.Now().AddDate(0, -periode, 0)

	// Nombre total de PDI (personnes déplacées internes)
	var totalPDI int64
	query := db.Model(&models.Migrant{}).Where("actif = ? AND created_at >= ?", true, dateDebut)
	if province != "" {
		query = query.Where("ville_actuelle = ? OR pays_actuel LIKE ?", province, "%"+province+"%")
	}
	query.Count(&totalPDI)

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
		pourcentage := float64(result.Count) / float64(total) * 100
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
		debutMois := time.Now().AddDate(0, -i-1, 0).Format("2006-01-02")
		finMois := time.Now().AddDate(0, -i, 0).Format("2006-01-02")

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

		periodeStr := time.Now().AddDate(0, -i-1, 0).Format("2006-01")
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

	if totalMotifs == 0 {
		return CausesDeplacementsIndicateurs{}
	}

	// Conflits armés
	var conflitsArmes int64
	db.Table("motif_deplacements md").
		Joins("JOIN migrants m ON md.migrant_uuid = m.uuid").
		Where("md.conflit_arme = ? AND md.created_at >= ? AND m.actif = ?", true, dateDebut, true).
		Count(&conflitsArmes)

	// Catastrophes naturelles
	var catastrophes int64
	db.Table("motif_deplacements md").
		Joins("JOIN migrants m ON md.migrant_uuid = m.uuid").
		Where("md.catastrophe_naturelle = ? AND md.created_at >= ? AND m.actif = ?", true, dateDebut, true).
		Count(&catastrophes)

	// Persécution
	var persecution int64
	db.Table("motif_deplacements md").
		Joins("JOIN migrants m ON md.migrant_uuid = m.uuid").
		Where("md.persecution = ? AND md.created_at >= ? AND m.actif = ?", true, dateDebut, true).
		Count(&persecution)

	// Violence généralisée
	var violenceGen int64
	db.Table("motif_deplacements md").
		Joins("JOIN migrants m ON md.migrant_uuid = m.uuid").
		Where("md.violence_generalisee = ? AND md.created_at >= ? AND m.actif = ?", true, dateDebut, true).
		Count(&violenceGen)

	// Autres causes (total - causes spécifiques)
	autresCauses := totalMotifs - conflitsArmes - catastrophes - persecution - violenceGen
	if autresCauses < 0 {
		autresCauses = 0
	}

	// Détails par type de motif
	detailsCauses := getDetailsCauses(periode, province)

	return CausesDeplacementsIndicateurs{
		PourcentageConflitsArmes: float64(conflitsArmes) / float64(totalMotifs) * 100,
		PourcentageCatastrophes:  float64(catastrophes) / float64(totalMotifs) * 100,
		PourcentagePersecution:   float64(persecution) / float64(totalMotifs) * 100,
		PourcentageViolenceGen:   float64(violenceGen) / float64(totalMotifs) * 100,
		PourcentageAutresCauses:  float64(autresCauses) / float64(totalMotifs) * 100,
		DetailsCauses:            detailsCauses,
	}
}

func getDetailsCauses(periode int, province string) []CauseDetailStats {
	db := database.DB
	dateDebut := time.Now().AddDate(0, -periode, 0)

	var results []struct {
		TypeMotif string `json:"type_motif"`
		Count     int64  `json:"count"`
	}

	query := db.Table("motif_deplacements md").
		Select("md.type_motif, COUNT(*) as count").
		Joins("JOIN migrants m ON md.migrant_uuid = m.uuid").
		Where("md.created_at >= ? AND m.actif = ?", dateDebut, true).
		Group("md.type_motif").
		Order("count DESC")

	if province != "" {
		query = query.Where("m.ville_actuelle = ?", province)
	}

	query.Scan(&results)

	// Total pour calculer les pourcentages
	var total int64
	for _, result := range results {
		total += result.Count
	}

	var detailsCauses []CauseDetailStats
	for _, result := range results {
		pourcentage := float64(result.Count) / float64(total) * 100
		detailsCauses = append(detailsCauses, CauseDetailStats{
			TypeMotif:   result.TypeMotif,
			NombreCas:   result.Count,
			Pourcentage: pourcentage,
		})
	}

	return detailsCauses
}

// 🏘️ INDICATEURS DE VULNÉRABILITÉ ET BESOINS
func getVulnerabiliteBesoinsIndicateurs(periode int, province string) VulnerabiliteBesoinsIndicateurs {
	db := database.DB
	dateDebut := time.Now().AddDate(0, -periode, 0)

	// Profil démographique
	profilDemo := getProfilDemographique(periode, province)

	// Accès aux services de base (simulé car pas de champs spécifiques dans le modèle)
	accesServices := getAccesServicesBase(periode, province)

	// Taux d'occupation des sites (basé sur type_localisation = "centre_accueil")
	var totalSites int64
	var occupesSites int64

	db.Table("geolocalisations g").
		Joins("JOIN migrants m ON g.migrant_uuid = m.uuid").
		Where("g.type_localisation = ? AND g.created_at >= ? AND m.actif = ?", "centre_accueil", dateDebut, true).
		Count(&occupesSites)

	// Sites disponibles (estimation basée sur les centres d'accueil existants)
	db.Table("geolocalisations").
		Where("type_localisation = ? AND created_at >= ?", "centre_accueil", dateDebut).
		Distinct("adresse").
		Count(&totalSites)

	var tauxOccupation float64
	if totalSites > 0 {
		tauxOccupation = float64(occupesSites) / float64(totalSites) * 100
	}

	// Déplacés hors sites (non hébergés dans centres d'accueil)
	var totalDeplaces int64
	migrantQuery := db.Model(&models.Migrant{}).
		Where("actif = ? AND created_at >= ?", true, dateDebut)
	if province != "" {
		migrantQuery = migrantQuery.Where("ville_actuelle = ?", province)
	}
	migrantQuery.Count(&totalDeplaces)

	deplacesHorsSites := totalDeplaces - occupesSites
	if deplacesHorsSites < 0 {
		deplacesHorsSites = 0
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

	// Total des migrants
	var totalMigrants int64
	migrantQuery := db.Model(&models.Migrant{}).
		Where("actif = ? AND created_at >= ?", true, dateDebut)
	if province != "" {
		migrantQuery = migrantQuery.Where("ville_actuelle = ?", province)
	}
	migrantQuery.Count(&totalMigrants)

	if totalMigrants == 0 {
		return ProfilDemographiqueStats{}
	}

	// Pourcentage de femmes
	var femmes int64
	femmeQuery := db.Model(&models.Migrant{}).
		Where("sexe = ? AND actif = ? AND created_at >= ?", "F", true, dateDebut)
	if province != "" {
		femmeQuery = femmeQuery.Where("ville_actuelle = ?", province)
	}
	femmeQuery.Count(&femmes)

	// Calcul de l'âge et profil démographique
	var results []struct {
		DateNaissance time.Time `json:"date_naissance"`
	}

	ageQuery := db.Model(&models.Migrant{}).
		Select("date_naissance").
		Where("actif = ? AND created_at >= ?", true, dateDebut)
	if province != "" {
		ageQuery = ageQuery.Where("ville_actuelle = ?", province)
	}
	ageQuery.Scan(&results)

	var totalAge float64
	var enfants int64 // < 18 ans
	var ages int64    // > 60 ans

	maintenant := time.Now()
	for _, result := range results {
		age := maintenant.Sub(result.DateNaissance).Hours() / 24 / 365.25
		totalAge += age

		if age < 18 {
			enfants++
		} else if age > 60 {
			ages++
		}
	}

	var ageMoyen float64
	if len(results) > 0 {
		ageMoyen = totalAge / float64(len(results))
	}

	return ProfilDemographiqueStats{
		PourcentageFemmes:  float64(femmes) / float64(totalMigrants) * 100,
		PourcentageEnfants: float64(enfants) / float64(totalMigrants) * 100,
		PourcentageAges:    float64(ages) / float64(totalMigrants) * 100,
		AgeMoyen:           ageMoyen,
	}
}

func getAccesServicesBase(periode int, province string) AccesServicesStats {
	// Cette fonction simule l'accès aux services basé sur les alertes et géolocalisation
	// Dans un système réel, ces données proviendraient d'enquêtes spécifiques

	db := database.DB
	dateDebut := time.Now().AddDate(0, -periode, 0)

	// Estimation basée sur les alertes de santé, éducation, etc.
	var alertesSante int64
	var alertesTotal int64

	alerteQuery := db.Table("alertes a").
		Joins("JOIN migrants m ON a.migrant_uuid = m.uuid").
		Where("a.created_at >= ? AND m.actif = ?", dateDebut, true)
	if province != "" {
		alerteQuery = alerteQuery.Where("m.ville_actuelle = ?", province)
	}

	alerteQuery.Count(&alertesTotal)
	alerteQuery.Where("a.type_alerte = ?", "sante").Count(&alertesSante)

	// Estimation simplifiée - dans un vrai système, il faudrait des données d'enquête
	var accesEau, accesSante, accesEducation, accesLogement float64 = 70.0, 65.0, 60.0, 55.0

	// Ajustement basé sur les alertes (plus d'alertes = moins d'accès)
	if alertesTotal > 0 {
		facteurAjustement := float64(alertesSante) / float64(alertesTotal)
		accesSante -= facteurAjustement * 20 // Réduction max de 20%
	}

	return AccesServicesStats{
		AccesEau:       accesEau,
		AccesSante:     accesSante,
		AccesEducation: accesEducation,
		AccesLogement:  accesLogement,
	}
}

// 📈 INDICATEURS DYNAMIQUES ET D'ALERTE
func getDynamiquesAlerteIndicateurs(periode int, province string) DynamiquesAlerteIndicateurs {
	// Zones à haut risque
	zonesRisque := getZonesHautRisque(periode, province)

	// Tendances de retour
	tendancesRetour := getTendancesRetour(periode, province)

	// Alertes précoces
	alertesPrecoces := getAlertesPrecoces(periode, province)

	// Mouvements massifs récents (> 1000 personnes en 30 jours)
	var mouvementsMassifs int64
	dateRecentDebut := time.Now().AddDate(0, 0, -30)

	db := database.DB
	massiveQuery := db.Table("migrants").
		Select("ville_actuelle, COUNT(*) as count").
		Where("actif = ? AND created_at >= ?", true, dateRecentDebut).
		Group("ville_actuelle").
		Having("COUNT(*) > ?", 1000)

	if province != "" {
		massiveQuery = massiveQuery.Where("ville_actuelle = ?", province)
	}

	var massiveResults []struct {
		Count int64 `json:"count"`
	}
	massiveQuery.Scan(&massiveResults)

	for _, result := range massiveResults {
		mouvementsMassifs += result.Count
	}

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

	// Identifier les zones avec beaucoup d'alertes critiques
	var results []struct {
		Ville         string `json:"ville"`
		CountAlertes  int64  `json:"count_alertes"`
		CountMigrants int64  `json:"count_migrants"`
	}

	query := db.Table("alertes a").
		Select("m.ville_actuelle as ville, COUNT(DISTINCT a.uuid) as count_alertes, COUNT(DISTINCT m.uuid) as count_migrants").
		Joins("JOIN migrants m ON a.migrant_uuid = m.uuid").
		Where("a.niveau_gravite IN (?, ?) AND a.created_at >= ? AND m.actif = ?", "danger", "critical", dateDebut, true).
		Group("m.ville_actuelle").
		Having("COUNT(DISTINCT a.uuid) >= ?", 5) // Au moins 5 alertes critiques

	if province != "" {
		query = query.Where("m.ville_actuelle = ?", province)
	}

	query.Order("count_alertes DESC").Limit(10).Scan(&results)

	var zonesRisque []ZoneRisqueStats
	for _, result := range results {
		niveauRisque := "MOYEN"
		if result.CountAlertes >= 20 {
			niveauRisque = "CRITIQUE"
		} else if result.CountAlertes >= 10 {
			niveauRisque = "ÉLEVÉ"
		}

		zonesRisque = append(zonesRisque, ZoneRisqueStats{
			Zone:           result.Ville,
			NiveauRisque:   niveauRisque,
			TypeMenace:     "MULTIPLE", // Basé sur les alertes diverses
			PopulationRisk: result.CountMigrants,
		})
	}

	return zonesRisque
}

func getTendancesRetour(periode int, province string) []TendanceRetourStats {
	db := database.DB
	dateDebut := time.Now().AddDate(0, -periode, 0)

	// Analyser les mouvements vers residence_permanente
	var results []struct {
		ZoneOrigine   string `json:"zone_origine"`
		ZoneRetour    string `json:"zone_retour"`
		NombreRetours int64  `json:"nombre_retours"`
	}

	query := db.Table("geolocalisations g").
		Select("m.ville_actuelle as zone_origine, g.ville as zone_retour, COUNT(*) as nombre_retours").
		Joins("JOIN migrants m ON g.migrant_uuid = m.uuid").
		Where("g.type_mouvement = ? AND g.created_at >= ? AND m.actif = ?", "residence_permanente", dateDebut, true).
		Group("m.ville_actuelle, g.ville").
		Having("COUNT(*) >= ?", 10) // Au moins 10 retours

	if province != "" {
		query = query.Where("m.ville_actuelle = ? OR g.ville = ?", province, province)
	}

	query.Order("nombre_retours DESC").Limit(15).Scan(&results)

	var tendances []TendanceRetourStats
	for _, result := range results {
		tendanceEvol := "STABLE"
		if result.NombreRetours >= 50 {
			tendanceEvol = "CROISSANT"
		} else if result.NombreRetours >= 25 {
			tendanceEvol = "MODÉRÉ"
		}

		tendances = append(tendances, TendanceRetourStats{
			ZoneOrigine:   result.ZoneOrigine,
			ZoneRetour:    result.ZoneRetour,
			NombreRetours: result.NombreRetours,
			TendanceEvol:  tendanceEvol,
		})
	}

	return tendances
}

func getAlertesPrecoces(periode int, province string) []AlertePrecoceStats {
	db := database.DB
	dateDebut := time.Now().AddDate(0, 0, -7) // 7 derniers jours pour alertes précoces

	var results []struct {
		Ville         string    `json:"ville"`
		TypeAlerte    string    `json:"type_alerte"`
		NiveauGravite string    `json:"niveau_gravite"`
		DateCreation  time.Time `json:"date_creation"`
		Description   string    `json:"description"`
	}

	query := db.Table("alertes a").
		Select("m.ville_actuelle as ville, a.type_alerte, a.niveau_gravite, a.created_at as date_creation, a.description").
		Joins("JOIN migrants m ON a.migrant_uuid = m.uuid").
		Where("a.niveau_gravite IN (?, ?) AND a.statut = ? AND a.created_at >= ? AND m.actif = ?",
			"danger", "critical", "active", dateDebut, true)

	if province != "" {
		query = query.Where("m.ville_actuelle = ?", province)
	}

	query.Order("a.created_at DESC").Limit(20).Scan(&results)

	var alertes []AlertePrecoceStats
	for _, result := range results {
		alertes = append(alertes, AlertePrecoceStats{
			Zone:          result.Ville,
			TypeAlerte:    result.TypeAlerte,
			NiveauGravite: result.NiveauGravite,
			DateDetection: result.DateCreation,
			Description:   result.Description,
		})
	}

	return alertes
}

// ENDPOINTS SPÉCIFIQUES POUR DES ANALYSES DÉTAILLÉES

// Indicateurs par province spécifique
func AnalyseDeplacementParProvince(c *fiber.Ctx) error {
	province := c.Params("province")
	periode := c.Query("periode", "12")

	periodeInt, err := strconv.Atoi(periode)
	if err != nil {
		periodeInt = 12
	}

	if province == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Province requise"})
	}

	// Générer les indicateurs pour la province spécifique
	volumeLocalisation := getVolumeLocalisationIndicateurs(periodeInt, province)
	causesDeplacements := getCausesDeplacementsIndicateurs(periodeInt, province)
	vulnerabiliteBesoins := getVulnerabiliteBesoinsIndicateurs(periodeInt, province)
	dynamiquesAlerte := getDynamiquesAlerteIndicateurs(periodeInt, province)

	response := IndicateursDeplacementResponse{
		VolumeLocalisation:   volumeLocalisation,
		CausesDeplacements:   causesDeplacements,
		VulnerabiliteBesoins: vulnerabiliteBesoins,
		DynamiquesAlerte:     dynamiquesAlerte,
		DateGeneration:       time.Now(),
		PeriodeAnalyse:       strconv.Itoa(periodeInt) + " derniers mois - Province: " + province,
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// Tendances évolutives des déplacements
func TendancesEvolution(c *fiber.Ctx) error {
	periode := c.Query("periode", "24") // 24 mois par défaut
	province := c.Query("province")

	periodeInt, err := strconv.Atoi(periode)
	if err != nil {
		periodeInt = 24
	}

	evolution := getEvolutionMensuelle(periodeInt, province)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"evolution_mensuelle": evolution,
		"periode_analyse":     strconv.Itoa(periodeInt) + " derniers mois",
		"province":            province,
		"date_generation":     time.Now(),
	})
}

// Alertes en temps réel
func AlertesTempsReel(c *fiber.Ctx) error {
	niveau := c.Query("niveau", "danger,critical")
	province := c.Query("province")
	jours := c.Query("jours", "7")

	joursInt, err := strconv.Atoi(jours)
	if err != nil {
		joursInt = 7
	}

	alertes := getAlertesPrecoces(joursInt, province)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"alertes_actives": alertes,
		"niveau_filtre":   niveau,
		"periode_jours":   joursInt,
		"province":        province,
		"date_generation": time.Now(),
	})
}

// Répartition géographique détaillée
func RepartitionGeographiqueDetaillee(c *fiber.Ctx) error {
	periode := c.Query("periode", "12")

	periodeInt, err := strconv.Atoi(periode)
	if err != nil {
		periodeInt = 12
	}

	repartition := getRepartitionGeographique(periodeInt, "")

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"repartition_provinces": repartition,
		"periode_analyse":       strconv.Itoa(periodeInt) + " derniers mois",
		"date_generation":       time.Now(),
	})
}

// Analyse des causes de déplacement
func AnalyseCausesDetaillees(c *fiber.Ctx) error {
	periode := c.Query("periode", "12")
	province := c.Query("province")

	periodeInt, err := strconv.Atoi(periode)
	if err != nil {
		periodeInt = 12
	}

	causes := getCausesDeplacementsIndicateurs(periodeInt, province)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"causes_deplacements": causes,
		"periode_analyse":     strconv.Itoa(periodeInt) + " derniers mois",
		"province":            province,
		"date_generation":     time.Now(),
	})
}
