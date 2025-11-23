package overview

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

// =================== INTERFACES POUR LE FRONTEND ===================

// Structure principale de réponse pour les indicateurs
type IndicateursDeplacementResponse struct {
	VolumeLocalisation   VolumeLocalisationIndicateurs   `json:"volume_localisation"`
	CausesDeplacements   CausesDeplacementsIndicateurs   `json:"causes_deplacements"`
	VulnerabiliteBesoins VulnerabiliteBesoinsIndicateurs `json:"vulnerabilite_besoins"`
	DynamiquesAlerte     DynamiquesAlerteIndicateurs     `json:"dynamiques_alerte"`
	DateGeneration       time.Time                       `json:"date_generation"`
	PeriodeAnalyse       string                          `json:"periode_analyse"`
}

// Structures pour les données des graphiques
type ChartDataPoint struct {
	Name  string      `json:"name"`
	Value float64     `json:"value"`
	Extra interface{} `json:"extra,omitempty"`
}

type ChartSeries struct {
	Name   string           `json:"name"`
	Series []ChartDataPoint `json:"series"`
}

// Indicateurs de volume et localisation
type VolumeLocalisationIndicateurs struct {
	NombreTotalPDI          int64                      `json:"nombre_total_pdi"`
	NombreTotalMigrants     int64                      `json:"nombre_total_migrants"`
	NombreDeplacesInternes  int64                      `json:"nombre_deplaces_internes"`
	PersonnesRetournees     int64                      `json:"personnes_retournees"`
	RepartitionGeographique []RepartitionProvinceStats `json:"repartition_geographique"`
	EvolutionMensuelle      []EvolutionTemporelleStats `json:"evolution_mensuelle"`
}

// Causes de déplacements
type CausesDeplacementsIndicateurs struct {
	PourcentageConflitsArmes       float64            `json:"pourcentage_conflits_armes"`
	PourcentageCatastrophes        float64            `json:"pourcentage_catastrophes"`
	PourcentagePersecution         float64            `json:"pourcentage_persecution"`
	PourcentageViolenceGeneralisee float64            `json:"pourcentage_violence_generalisee"`
	PourcentageAutresCauses        float64            `json:"pourcentage_autres_causes"`
	DetailsCauses                  []CauseDetailStats `json:"details_causes"`
}

// Vulnérabilité et besoins
type VulnerabiliteBesoinsIndicateurs struct {
	ProfilDemographique ProfilDemographiqueStats `json:"profil_demographique"`
	AccesServicesBase   AccesServicesStats       `json:"acces_services_base"`
	TauxOccupationSites float64                  `json:"taux_occupation_sites"`
	DeplacesHorsSites   int64                    `json:"deplaces_hors_sites"`
}

// Dynamiques et alertes
type DynamiquesAlerteIndicateurs struct {
	ZonesHautRisque   []ZoneRisqueStats     `json:"zones_haut_risque"`
	TendancesRetour   []TendanceRetourStats `json:"tendances_retour"`
	AlertesPrecoces   []AlertePrecoceStats  `json:"alertes_precoces"`
	MouvementsMassifs int64                 `json:"mouvements_massifs_recent"`
}

// =================== STRUCTURES DE SUPPORT ===================

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

// Structure pour les alertes temps réel
type AlertesTempsReelResponse struct {
	AlertesActives []AlertePrecoceStats `json:"alertes_actives"`
	NombreTotal    int64                `json:"nombre_total"`
	DateMiseAJour  time.Time            `json:"date_mise_a_jour"`
}

// Structure pour le pie chart des motifs de déplacement
type MotifPieChartResponse struct {
	Data           []ChartDataPoint `json:"data"`
	Total          int64            `json:"total"`
	DateMiseAJour  time.Time        `json:"date_mise_a_jour"`
	PeriodeAnalyse string           `json:"periode_analyse"`
}

// =================== ENDPOINTS PRINCIPAUX ===================

// GetIndicateursGeneraux - Endpoint principal pour récupérer tous les indicateurs
// GET /api/overview/indicateurs?periode=12&province=
func GetIndicateursGeneraux(c *fiber.Ctx) error {
	// Paramètres optionnels
	periode := c.Query("periode", "12") // 12 derniers mois par défaut
	province := c.Query("province", "") // Province spécifique si demandée

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

// GetAlertesTempsReel - Endpoint pour récupérer les alertes en temps réel
// GET /api/overview/alertes?niveaux=danger,critical&province=&jours=7
func GetAlertesTempsReel(c *fiber.Ctx) error {
	niveaux := c.Query("niveaux", "danger,critical") // Niveaux de gravité
	province := c.Query("province", "")              // Province spécifique
	jours := c.Query("jours", "7")                   // Nombre de jours

	joursInt, err := strconv.Atoi(jours)
	if err != nil {
		joursInt = 7
	}

	alertes := getAlertesRecentes(niveaux, province, joursInt)

	response := AlertesTempsReelResponse{
		AlertesActives: alertes,
		NombreTotal:    int64(len(alertes)),
		DateMiseAJour:  time.Now(),
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// GetRepartitionGeographique - Endpoint pour récupérer la répartition géographique séparément
// GET /api/overview/repartition?periode=12
func GetRepartitionGeographique(c *fiber.Ctx) error {
	periode := c.Query("periode", "12")

	periodeInt, err := strconv.Atoi(periode)
	if err != nil {
		periodeInt = 12
	}

	// Récupérer la répartition sans filtre de province pour avoir une vue globale
	repartition := getRepartitionGeographique(periodeInt, "")

	response := struct {
		RepartitionProvinces []RepartitionProvinceStats `json:"repartition_provinces"`
		DateMiseAJour        time.Time                  `json:"date_mise_a_jour"`
		PeriodeAnalyse       string                     `json:"periode_analyse"`
	}{
		RepartitionProvinces: repartition,
		DateMiseAJour:        time.Now(),
		PeriodeAnalyse:       strconv.Itoa(periodeInt) + " derniers mois",
	}

	return c.Status(fiber.StatusOK).JSON(response)
}

// GetMotifsPieChart - Endpoint pour récupérer les données du pie chart des motifs de déplacement
// GET /api/overview/motifs-pie?periode=12&province=
func GetMotifsPieChart(c *fiber.Ctx) error {
	periode := c.Query("periode", "12")
	province := c.Query("province", "")

	periodeInt, err := strconv.Atoi(periode)
	if err != nil {
		periodeInt = 12
	}

	// Récupérer les données du pie chart
	pieData := getMotifsPieChartData(periodeInt, province)

	// Calculer le total
	var total int64
	for _, data := range pieData {
		total += int64(data.Value)
	}

	response := MotifPieChartResponse{
		Data:           pieData,
		Total:          total,
		DateMiseAJour:  time.Now(),
		PeriodeAnalyse: strconv.Itoa(periodeInt) + " derniers mois",
	}

	return c.Status(fiber.StatusOK).JSON(response)
}
