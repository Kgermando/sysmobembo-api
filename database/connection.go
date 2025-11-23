package database

import (
	"encoding/base64"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"time"

	"github.com/kgermando/sysmobembo-api/models"
	"github.com/kgermando/sysmobembo-api/utils"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Connect() {
	p := utils.Env("DB_PORT")
	port, err := strconv.ParseUint(p, 10, 32)
	if err != nil {
		panic("failed to parse database port 😵!")
	}

	DNS := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", utils.Env("DB_HOST"), port, utils.Env("DB_USER"), utils.Env("DB_PASSWORD"), utils.Env("DB_NAME"))
	connection, err := gorm.Open(postgres.Open(DNS), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		panic("Could not connect to the database 😰!")
	}

	DB = connection
	fmt.Println("Database Connected 🎉!")

	// Migration automatique des modèles
	err = connection.AutoMigrate(
		// Modèles de base
		&models.User{},
		&models.PasswordReset{},

		// Modèles d'identité
		&models.Identite{},

		// Modèles migrants et entités associées
		&models.Migrant{},
		&models.MotifDeplacement{},
		&models.Alert{},
		&models.Biometrie{},
		&models.Geolocalisation{},
	)

	if err != nil {
		panic("Failed to migrate database models 😵!")
	}

	fmt.Println("Database Models Migrated Successfully ✅!")

	// Initialiser les données simulées si la base est vide
	initializeSampleDataIfEmpty(connection)
}

// initializeSampleDataIfEmpty vérifie si la base est vide et initialise les données simulées
func initializeSampleDataIfEmpty(db *gorm.DB) {
	var userCount, migrantCount int64
	db.Model(&models.User{}).Count(&userCount)
	db.Model(&models.Migrant{}).Count(&migrantCount)

	// Si aucun utilisateur et aucun migrant n'existent, initialiser les données
	if userCount == 0 && migrantCount == 0 {
		log.Println("🎯 Base de données vide détectée. Initialisation des données simulées...")

		if err := runAllSimulators(db); err != nil {
			log.Printf("❌ Erreur lors de l'initialisation des données simulées: %v", err)
		} else {
			log.Println("✅ Données simulées initialisées avec succès!")
		}
	} else {
		log.Printf("📊 Base de données existante détectée (%d utilisateurs, %d migrants)", userCount, migrantCount)
	}
}

// runAllSimulators exécute tous les simulateurs dans l'ordre approprié
func runAllSimulators(db *gorm.DB) error {
	log.Println("=== DÉBUT DE LA SIMULATION DE DONNÉES ===")

	// 1. Créer les utilisateurs en premier
	log.Println("1. Création des utilisateurs...")
	if err := simulateUsers(db); err != nil {
		return fmt.Errorf("erreur lors de la simulation des utilisateurs: %v", err)
	}

	// 2. Créer les identités d'abord
	log.Println("2. Création des identités...")
	identiteMap, err := simulateIdentites(db)
	if err != nil {
		return fmt.Errorf("erreur lors de la simulation des identités: %v", err)
	}

	// 3. Créer les migrants (dépendent des identités)
	log.Println("3. Création des migrants...")
	if err := simulateMigrants(db, identiteMap); err != nil {
		return fmt.Errorf("erreur lors de la simulation des migrants: %v", err)
	}

	// 4. Créer les géolocalisations (dépendent des identités)
	log.Println("4. Création des géolocalisations...")
	if err := simulateGeolocalisations(db, identiteMap); err != nil {
		return fmt.Errorf("erreur lors de la simulation des géolocalisations: %v", err)
	}

	// 5. Créer les motifs de déplacement (dépendent des migrants)
	log.Println("5. Création des motifs de déplacement...")
	if err := simulateMotifDeplacements(db); err != nil {
		return fmt.Errorf("erreur lors de la simulation des motifs de déplacement: %v", err)
	}

	// 6. Créer les données biométriques (dépendent des migrants)
	log.Println("6. Création des données biométriques...")
	if err := simulateBiometries(db); err != nil {
		return fmt.Errorf("erreur lors de la simulation des biométries: %v", err)
	}

	// 7. Créer les alertes (dépendent des migrants)
	log.Println("7. Création des alertes...")
	if err := simulateAlerts(db); err != nil {
		return fmt.Errorf("erreur lors de la simulation des alertes: %v", err)
	}

	log.Println("=== SIMULATION TERMINÉE AVEC SUCCÈS ===")
	log.Println("📊 Statistiques des données créées:")
	log.Println("✅ 3 utilisateurs du système (Administrateurs DGM)")
	log.Println("✅ 50 identités de migrants (réparties sur 6 mois)")
	log.Println("✅ 50 migrants avec statuts variés")
	log.Println("   - Distribution géographique réaliste à travers la RDC")
	log.Println("   - Kinshasa (35%), Goma (20%), Lubumbashi (15%), etc.")
	log.Println("✅ ~150-200 géolocalisations (2-4 par migrant)")
	log.Println("✅ 50 motifs de déplacement")
	log.Println("   - Économiques, politiques, sécuritaires, etc.")
	log.Println("✅ ~100-150 données biométriques")
	log.Println("   - Empreintes digitales et reconnaissance faciale")
	log.Println("✅ ~75-125 alertes de suivi")
	log.Println("   - Sécurité, santé, administrative, sociale, juridique")
	log.Println("⏰ Données étalées de janvier à juin 2025")
	log.Println("🗺️  Coordonnées GPS réelles des villes de la RDC")

	return nil
}

// simulateUsers crée des utilisateurs simulés
func simulateUsers(db *gorm.DB) error {
	users := []models.User{
		{
			UUID:              utils.GenerateUUID(),
			Nom:               "MBEKO",
			PostNom:           "NGOLA",
			Prenom:            "Jean-Claude",
			Sexe:              "M",
			DateNaissance:     time.Date(1985, 3, 15, 0, 0, 0, 0, time.UTC),
			LieuNaissance:     "Kinshasa",
			EtatCivil:         "Marié(e)",
			NombreEnfants:     2,
			Nationalite:       "Congolaise (RDC)",
			NumeroCNI:         "1234567890123456",
			DateEmissionCNI:   time.Date(2020, 1, 10, 0, 0, 0, 0, time.UTC),
			DateExpirationCNI: time.Date(2030, 1, 10, 0, 0, 0, 0, time.UTC),
			LieuEmissionCNI:   "Kinshasa",
			Email:             "jean.mbeko@dgm.cd",
			Telephone:         "+243815234567",
			TelephoneUrgence:  "+243987654321",
			Province:          "Kinshasa",
			Ville:             "Kinshasa",
			Commune:           "Gombe",
			Quartier:          "Centre-ville",
			Avenue:            "Boulevard du 30 juin",
			Numero:            "123",
			Matricule:         "DGM001",
			Grade:             "Administrateur Principal",
			Fonction:          "Directeur des Migrations",
			Service:           "Direction Générale",
			Direction:         "Direction Générale des Migrations",
			Ministere:         "Ministère de l'Intérieur",
			DateRecrutement:   time.Date(2010, 6, 1, 0, 0, 0, 0, time.UTC),
			DatePriseService:  time.Date(2010, 6, 15, 0, 0, 0, 0, time.UTC),
			TypeAgent:         "Fonctionnaire",
			Statut:            "Actif",
			NiveauEtude:       "Universitaire",
			DiplomeBase:       "Master en Administration Publique",
			UniversiteEcole:   "Université de Kinshasa",
			AnneeObtention:    2008,
			Specialisation:    "Gestion des Migrations",
			Role:              "Administrator",
			Permission:        "full_access",
			Status:            true,
			DernierAcces:      time.Now(),
			NombreConnexions:  rand.Intn(50) + 10,
		},
		{
			UUID:              utils.GenerateUUID(),
			Nom:               "KASONGO",
			PostNom:           "MWAMBA",
			Prenom:            "Marie-Claire",
			Sexe:              "F",
			DateNaissance:     time.Date(1990, 7, 22, 0, 0, 0, 0, time.UTC),
			LieuNaissance:     "Lubumbashi",
			EtatCivil:         "Célibataire",
			NombreEnfants:     0,
			Nationalite:       "Congolaise (RDC)",
			NumeroCNI:         "2345678901234567",
			DateEmissionCNI:   time.Date(2021, 3, 5, 0, 0, 0, 0, time.UTC),
			DateExpirationCNI: time.Date(2031, 3, 5, 0, 0, 0, 0, time.UTC),
			LieuEmissionCNI:   "Lubumbashi",
			Email:             "marie.kasongo@dgm.cd",
			Telephone:         "+243976543210",
			Province:          "Haut-Katanga",
			Ville:             "Lubumbashi",
			Commune:           "Lubumbashi",
			Quartier:          "Kenya",
			Matricule:         "DGM002",
			Grade:             "Attaché",
			Fonction:          "Agent des Migrations",
			Service:           "Service de Contrôle",
			Direction:         "Direction des Contrôles Migratoires",
			Ministere:         "Ministère de l'Intérieur",
			DateRecrutement:   time.Date(2015, 9, 1, 0, 0, 0, 0, time.UTC),
			DatePriseService:  time.Date(2015, 9, 15, 0, 0, 0, 0, time.UTC),
			TypeAgent:         "Contractuel",
			Statut:            "Actif",
			Role:              "Manager",
			Permission:        "migration_management",
			Status:            true,
			DernierAcces:      time.Now().Add(-time.Hour * 2),
			NombreConnexions:  rand.Intn(30) + 5,
		},
		{
			UUID:             utils.GenerateUUID(),
			Nom:              "TSHISEKEDI",
			PostNom:          "KABONGO",
			Prenom:           "Joseph",
			Sexe:             "M",
			DateNaissance:    time.Date(1988, 11, 10, 0, 0, 0, 0, time.UTC),
			LieuNaissance:    "Mbuji-Mayi",
			EtatCivil:        "Marié(e)",
			NombreEnfants:    1,
			Nationalite:      "Congolaise (RDC)",
			Email:            "joseph.tshisekedi@dgm.cd",
			Telephone:        "+243898765432",
			Province:         "Kasaï-Oriental",
			Ville:            "Mbuji-Mayi",
			Matricule:        "DGM003",
			Grade:            "Conseiller",
			Fonction:         "Superviseur Régional",
			Service:          "Service Régional Kasaï",
			Direction:        "Direction Régionale",
			Ministere:        "Ministère de l'Intérieur",
			DateRecrutement:  time.Date(2012, 4, 1, 0, 0, 0, 0, time.UTC),
			DatePriseService: time.Date(2012, 4, 15, 0, 0, 0, 0, time.UTC),
			TypeAgent:        "Fonctionnaire",
			Statut:           "Actif",
			Role:             "Supervisor",
			Permission:       "regional_supervision",
			Status:           true,
			DernierAcces:     time.Now().Add(-time.Hour * 4),
			NombreConnexions: rand.Intn(40) + 8,
		},
	}

	// Hasher les mots de passe
	for i := range users {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("erreur lors du hashage du mot de passe: %v", err)
		}
		users[i].Password = string(hashedPassword)
		users[i].CreatedAt = time.Now()
		users[i].UpdatedAt = time.Now()
	}

	// Insérer en base
	for _, user := range users {
		if err := db.Create(&user).Error; err != nil {
			log.Printf("Erreur lors de la création de l'utilisateur %s: %v", user.Email, err)
			continue
		}
	}

	log.Printf("✅ %d utilisateurs créés", len(users))
	return nil
}

// simulateIdentites crée les identités et retourne un map[NumeroIdentifiant]IdentiteUUID
func simulateIdentites(db *gorm.DB) (map[string]string, error) {
	identiteMap := make(map[string]string)

	// Noms et prénoms réalistes pour la RDC et pays voisins
	noms := []string{"KABILA", "TSHISEKEDI", "MBUYI", "MUKENDI", "KASONGO", "NGOY", "MULAMBA", "ILUNGA",
		"KALALA", "KILOLO", "LUBOYA", "MATANDA", "NDALA", "NKULU", "MUTOMBO", "BANZA", "KALONJI",
		"KAMBALE", "KASEREKA", "MUHINDO", "SIVIHWA", "PALUKU", "MBUSA", "KAVIRA"}

	prenoms := []string{"Jean-Pierre", "Marie", "Joseph", "Grace", "Patient", "Espérance", "Emmanuel",
		"Jeanne", "David", "Sarah", "Daniel", "Rebecca", "Samuel", "Ruth", "Isaac", "Esther"}

	// Villes de la RDC avec leurs coordonnées GPS
	villes := []struct {
		Nom string
		Lat float64
		Lng float64
	}{
		{"Kinshasa", -4.3317, 15.3139},
		{"Lubumbashi", -11.6792, 27.4847},
		{"Goma", -1.6792, 29.2228},
		{"Bukavu", -2.5078, 28.8617},
		{"Bunia", 1.5593, 30.0944},
		{"Matadi", -5.8386, 13.4644},
		{"Kasumbalesa", -10.3667, 28.0167},
	}

	nationalites := []struct {
		Pays             string
		AutoriteEmetteur string
		PrefixePasseport string
		LieuxNaissance   []string
	}{
		{"Congolaise (RDC)", "République Démocratique du Congo", "CD", []string{"Kinshasa", "Goma", "Lubumbashi", "Bukavu", "Bunia"}},
		{"Rwandaise", "République du Rwanda", "RW", []string{"Kigali", "Butare", "Gisenyi"}},
		{"Burundaise", "République du Burundi", "BI", []string{"Bujumbura", "Gitega", "Ngozi"}},
		{"Ougandaise", "République de l'Ouganda", "UG", []string{"Kampala", "Entebbe", "Gulu"}},
		{"Sud-Soudanaise", "République du Soudan du Sud", "SS", []string{"Djouba", "Wau", "Malakal"}},
	}

	professions := []string{"Commerçant(e)", "Agriculteur", "Enseignant(e)", "Infirmier(ère)",
		"Mécanicien", "Chauffeur", "Couturier(ère)", "Menuisier", "Cultivateur", "Éleveur",
		"Pêcheur", "Artisan", "Ouvrier", "Vendeur(se)"}

	// Générer 50 identités réparties sur 6 mois (janvier à juin 2025)
	baseDate := time.Date(2025, 1, 1, 8, 0, 0, 0, time.UTC)

	for i := 1; i <= 50; i++ {
		nat := nationalites[rand.Intn(len(nationalites))]
		ville := villes[rand.Intn(len(villes))]

		// Distribution temporelle réaliste (plus de migrants récents)
		daysOffset := rand.Intn(180)  // 6 mois
		heuresOffset := rand.Intn(10) // heures de bureau 8h-18h
		createdAt := baseDate.AddDate(0, 0, daysOffset).Add(time.Hour * time.Duration(heuresOffset))

		numeroIdentifiant := fmt.Sprintf("MIG2025%03d", i)

		identite := models.Identite{
			UUID:             utils.GenerateUUID(),
			Nom:              noms[rand.Intn(len(noms))],
			Prenom:           prenoms[rand.Intn(len(prenoms))],
			DateNaissance:    time.Date(1970+rand.Intn(35), time.Month(rand.Intn(12)+1), rand.Intn(28)+1, 0, 0, 0, 0, time.UTC),
			LieuNaissance:    nat.LieuxNaissance[rand.Intn(len(nat.LieuxNaissance))],
			Sexe:             []string{"M", "F"}[rand.Intn(2)],
			Nationalite:      nat.Pays,
			Adresse:          fmt.Sprintf("Avenue %s, N°%d, %s", []string{"Kasavubu", "Lumumba", "Mobutu", "de la Libération"}[rand.Intn(4)], rand.Intn(200)+1, ville.Nom),
			Profession:       professions[rand.Intn(len(professions))],
			PaysEmetteur:     nat.AutoriteEmetteur,
			AutoriteEmetteur: nat.AutoriteEmetteur,
			NumeroPasseport:  fmt.Sprintf("%s%07d", nat.PrefixePasseport, rand.Intn(9999999)),
			CreatedAt:        createdAt,
			UpdatedAt:        createdAt,
		}

		if err := db.Create(&identite).Error; err != nil {
			log.Printf("Erreur lors de la création de l'identité %s: %v", numeroIdentifiant, err)
			continue
		}
		identiteMap[numeroIdentifiant] = identite.UUID
	}

	log.Printf("✅ %d identités créées sur 6 mois", len(identiteMap))
	return identiteMap, nil
}

// simulateMigrants crée des migrants simulés et les associe aux identités
func simulateMigrants(db *gorm.DB, identiteMap map[string]string) error {
	// Récupérer toutes les identités créées
	var identites []models.Identite
	if err := db.Find(&identites).Error; err != nil {
		return err
	}

	villes := []struct {
		Nom         string
		PointEntree string
	}{
		{"Kinshasa", "Aéroport de N'djili"},
		{"Lubumbashi", "Aéroport de Luano"},
		{"Goma", "Frontière de Gisenyi (Rwanda)"},
		{"Bukavu", "Frontière de Cyangugu (Rwanda)"},
		{"Bunia", "Frontière de Mahagi (Ouganda)"},
		{"Matadi", "Port de Matadi"},
		{"Kasumbalesa", "Frontière de Kasumbalesa (Zambie)"},
	}

	statutsMigratoires := []string{"regulier", "irregulier", "demandeur_asile", "refugie", "deplace_interne"}
	situationsMatrimoniales := []string{"celibataire", "marie", "divorce", "veuf"}

	var migrants []models.Migrant

	for i, identite := range identites {
		numeroIdentifiant := fmt.Sprintf("MIG2025%03d", i+1)
		ville := villes[rand.Intn(len(villes))]

		// Date d'entrée quelques jours avant la création de l'identité
		dateEntree := identite.CreatedAt.AddDate(0, 0, -rand.Intn(30))

		migrant := models.Migrant{
			UUID:                  utils.GenerateUUID(),
			NumeroIdentifiant:     numeroIdentifiant,
			IdentiteUUID:          identite.UUID,
			Telephone:             fmt.Sprintf("+243%d%08d", rand.Intn(2)+8, rand.Intn(99999999)),
			Email:                 fmt.Sprintf("%s.%s@email.com", identite.Prenom, identite.Nom),
			AdresseActuelle:       identite.Adresse,
			VilleActuelle:         ville.Nom,
			PaysActuel:            "République Démocratique du Congo",
			SituationMatrimoniale: situationsMatrimoniales[rand.Intn(len(situationsMatrimoniales))],
			NombreEnfants:         rand.Intn(6),
			StatutMigratoire:      statutsMigratoires[rand.Intn(len(statutsMigratoires))],
			DateEntree:            &dateEntree,
			PointEntree:           ville.PointEntree,
			PaysDestination:       "République Démocratique du Congo",
			CreatedAt:             identite.CreatedAt,
			UpdatedAt:             identite.UpdatedAt,
		}

		// Ajouter contact pour les déplacés internes
		if migrant.StatutMigratoire == "deplace_interne" {
			migrant.PersonneContact = fmt.Sprintf("%s Contact", identite.Nom)
			migrant.TelephoneContact = fmt.Sprintf("+243%d%08d", rand.Intn(2)+8, rand.Intn(99999999))
		}

		migrants = append(migrants, migrant)
	}

	// Créer les migrants en base
	for _, migrant := range migrants {
		if err := db.Create(&migrant).Error; err != nil {
			log.Printf("Erreur lors de la création du migrant %s: %v", migrant.NumeroIdentifiant, err)
			continue
		}
	}

	log.Printf("✅ %d migrants créés et associés aux identités", len(migrants))
	return nil
}

// simulateGeolocalisations crée des géolocalisations simulées avec coordonnées GPS réelles de la RDC
func simulateGeolocalisations(db *gorm.DB, identiteMap map[string]string) error {
	// Récupérer toutes les identités
	var identites []models.Identite
	if err := db.Find(&identites).Error; err != nil {
		return err
	}

	// Villes de la RDC avec coordonnées GPS réelles et variations
	villes := []struct {
		Nom        string
		LatBase    float64
		LngBase    float64
		LatRadius  float64 // Rayon pour variation de latitude
		LngRadius  float64 // Rayon pour variation de longitude
		Proportion float64 // Proportion de migrants dans cette ville
	}{
		{"Kinshasa", -4.3317, 15.3139, 0.15, 0.15, 0.35},     // 35% - Capitale
		{"Goma", -1.6792, 29.2228, 0.05, 0.05, 0.20},         // 20% - Zone de conflit
		{"Lubumbashi", -11.6792, 27.4847, 0.10, 0.10, 0.15},  // 15% - Centre minier
		{"Bukavu", -2.5078, 28.8617, 0.05, 0.05, 0.12},       // 12% - Frontière Rwanda
		{"Bunia", 1.5593, 30.0944, 0.03, 0.03, 0.10},         // 10% - Ituri
		{"Kasumbalesa", -10.3667, 28.0167, 0.02, 0.02, 0.05}, // 5% - Frontière Zambie
		{"Matadi", -5.8386, 13.4644, 0.04, 0.04, 0.03},       // 3% - Port
	}

	var geolocalisations []models.Geolocalisation

	// Attribution des villes basée sur les proportions
	villeIndex := 0
	cumul := 0.0

	for _, identite := range identites {
		// Sélectionner une ville selon la proportion
		randValue := rand.Float64()
		cumul = 0.0
		for i, v := range villes {
			cumul += v.Proportion
			if randValue <= cumul {
				villeIndex = i
				break
			}
		}

		ville := villes[villeIndex]

		// Générer 2-4 positions de géolocalisation par identité pour montrer les déplacements
		numPositions := rand.Intn(3) + 2

		for i := 0; i < numPositions; i++ {
			// Variation aléatoire autour du centre de la ville
			latVariation := (rand.Float64()*2 - 1) * ville.LatRadius
			lngVariation := (rand.Float64()*2 - 1) * ville.LngRadius

			// Date de capture étalée sur plusieurs semaines
			dateCapture := identite.CreatedAt.AddDate(0, 0, i*rand.Intn(15)+1)

			geo := models.Geolocalisation{
				UUID:         utils.GenerateUUID(),
				IdentiteUUID: identite.UUID,
				Latitude:     ville.LatBase + latVariation,
				Longitude:    ville.LngBase + lngVariation,
				CreatedAt:    dateCapture,
				UpdatedAt:    dateCapture,
			}

			geolocalisations = append(geolocalisations, geo)
		}
	}

	// Insérer en base
	for _, geo := range geolocalisations {
		if err := db.Create(&geo).Error; err != nil {
			log.Printf("Erreur lors de la création de la géolocalisation: %v", err)
			continue
		}
	}

	log.Printf("✅ %d géolocalisations créées à travers la RDC", len(geolocalisations))
	log.Println("📍 Distribution géographique:")
	for _, v := range villes {
		log.Printf("   - %s: %.0f%%", v.Nom, v.Proportion*100)
	}
	return nil
}

// simulateMotifDeplacements crée des motifs de déplacement simulés réalistes
func simulateMotifDeplacements(db *gorm.DB) error {
	// Récupérer les migrants existants
	var migrants []models.Migrant
	if err := db.Find(&migrants).Error; err != nil {
		return err
	}

	if len(migrants) == 0 {
		return nil
	}

	// Motifs réalistes par type
	motifsParType := map[string][]struct {
		Principal   string
		Secondaire  string
		Description string
		Volontaire  bool
		Urgence     string
		DureeJours  int
	}{
		"economique": {
			{"Recherche d'opportunités d'emploi", "Amélioration des conditions de vie", "Migration économique vers les centres urbains pour trouver du travail dans le secteur formel ou informel.", true, "moyenne", 730},
			{"Activités commerciales transfrontalières", "Commerce et négoce", "Commerçant effectuant des va-et-vient pour activités commerciales entre pays limitrophes.", true, "faible", 365},
			{"Formation professionnelle", "Développement des compétences", "Migration pour suivre une formation ou des études supérieures.", true, "faible", 1095},
		},
		"politique": {
			{"Conflits armés et violences", "Protection de la vie et de la famille", "Fuite des zones de conflit armé impliquant des groupes rebelles, violence contre les civils.", false, "critique", 1460},
			{"Violences intercommunautaires", "Tensions ethniques", "Déplacement forcé suite à des affrontements entre communautés ethniques.", false, "elevee", 1095},
			{"Persécutions politiques", "Activisme et opinions politiques", "Menaces liées aux opinions politiques ou à l'activisme.", false, "elevee", 1825},
		},
		"securite": {
			{"Attaques de groupes armés", "Violences et pillages", "Attaques répétées par des groupes armés non étatiques, massacres de civils.", false, "critique", 1460},
			{"Insécurité généralisée", "Crimes et violences", "Zone devenue trop dangereuse pour y vivre en sécurité.", false, "elevee", 1095},
			{"Enlèvements et kidnappings", "Menaces directes", "Vague d'enlèvements ciblant certaines communautés.", false, "critique", 730},
		},
		"environnement": {
			{"Catastrophes naturelles", "Inondations et érosions", "Déplacement suite à des inondations, glissements de terrain ou érosions massives.", false, "elevee", 365},
			{"Éruptions volcaniques", "Catastrophe naturelle", "Fuite suite à l'éruption du volcan Nyiragongo.", false, "critique", 545},
		},
		"sante": {
			{"Épidémies", "Accès aux soins médicaux", "Recherche de meilleurs soins suite à épidémie (Ebola, choléra).", true, "elevee", 180},
			{"Soins médicaux spécialisés", "Traitement médical", "Migration temporaire pour accès à des soins spécialisés.", true, "moyenne", 90},
		},
		"familial": {
			{"Regroupement familial", "Réunification avec la famille", "Migration pour rejoindre des membres de la famille déjà installés.", true, "faible", 365},
			{"Mariage", "Union matrimoniale", "Migration suite à un mariage dans une autre ville ou pays.", true, "faible", 730},
		},
	}

	var motifDeplacements []models.MotifDeplacement

	typesMotifs := []string{"economique", "politique", "securite", "environnement", "sante", "familial"}

	// Créer des motifs variés pour chaque migrant
	for _, migrant := range migrants {
		// Sélection du type de motif selon le statut migratoire
		var typeMotif string
		switch migrant.StatutMigratoire {
		case "deplace_interne", "refugie", "demandeur_asile":
			// Plus de motifs politiques et de sécurité
			typeMotif = []string{"politique", "politique", "securite", "securite", "environnement"}[rand.Intn(5)]
		case "irregulier":
			// Plus de motifs économiques
			typeMotif = []string{"economique", "economique", "economique", "familial"}[rand.Intn(4)]
		default: // regulier
			typeMotif = typesMotifs[rand.Intn(len(typesMotifs))]
		}

		motifs := motifsParType[typeMotif]
		motif := motifs[rand.Intn(len(motifs))]

		// Date de déclenchement avant la date d'entrée
		var dateDeclenchement time.Time
		if migrant.DateEntree != nil {
			dateDeclenchement = migrant.DateEntree.AddDate(0, 0, -rand.Intn(60)-30) // 1-3 mois avant
		} else {
			dateDeclenchement = migrant.CreatedAt.AddDate(0, 0, -rand.Intn(90))
		}

		motifDeplacement := models.MotifDeplacement{
			UUID:                utils.GenerateUUID(),
			MigrantUUID:         migrant.UUID,
			TypeMotif:           typeMotif,
			MotifPrincipal:      motif.Principal,
			MotifSecondaire:     motif.Secondaire,
			Description:         motif.Description,
			CaractereVolontaire: motif.Volontaire,
			Urgence:             motif.Urgence,
			DateDeclenchement:   dateDeclenchement,
			DureeEstimee:        motif.DureeJours + rand.Intn(365), // +/- 1 an de variation
			CreatedAt:           migrant.CreatedAt,
			UpdatedAt:           migrant.UpdatedAt,
		}

		motifDeplacements = append(motifDeplacements, motifDeplacement)
	}

	// Insérer en base
	for _, motif := range motifDeplacements {
		if err := db.Create(&motif).Error; err != nil {
			log.Printf("Erreur lors de la création du motif de déplacement: %v", err)
			continue
		}
	}

	log.Printf("✅ %d motifs de déplacement créés", len(motifDeplacements))

	// Statistiques par type
	stats := make(map[string]int)
	for _, m := range motifDeplacements {
		stats[m.TypeMotif]++
	}
	log.Println("📊 Distribution par type de motif:")
	for type_, count := range stats {
		log.Printf("   - %s: %d (%.1f%%)", type_, count, float64(count)/float64(len(motifDeplacements))*100)
	}

	return nil
}

// simulateBiometries crée des données biométriques simulées réalistes
func simulateBiometries(db *gorm.DB) error {
	// Récupérer les migrants existants
	var migrants []models.Migrant
	if err := db.Find(&migrants).Error; err != nil {
		return err
	}

	if len(migrants) == 0 {
		return nil
	}

	dispositifs := []string{
		"Scanner biométrique SecuGen Hamster Pro 20",
		"Lecteur d'empreintes digitales Morpho MSO 1300 E3",
		"Caméra de reconnaissance faciale HikVision DeepinMind",
		"Scanner iris IrisGuard IG-AD100",
	}

	qualites := []string{"excellente", "bonne", "moyenne"}

	var biometries []models.Biometrie

	// Créer des données biométriques pour chaque migrant
	for i, migrant := range migrants {
		// Nombre de captures biométriques par migrant (2-3)
		numCaptures := rand.Intn(2) + 2

		for capture := 0; capture < numCaptures; capture++ {
			var typeBio string
			var indexDoigt *int
			var tailleFichier int
			var resolution string
			var algorithme string

			// Alternance entre empreintes et reconnaissance faciale
			if capture%2 == 0 {
				typeBio = "empreinte_digitale"
				doigt := rand.Intn(10) + 1 // Doigts 1-10
				indexDoigt = &doigt
				tailleFichier = rand.Intn(3000) + 2000 // 2-5 KB
				resolution = []string{"500 DPI", "1000 DPI"}[rand.Intn(2)]
				algorithme = "WSQ (Wavelet Scalar Quantization)"
			} else {
				typeBio = "reconnaissance_faciale"
				tailleFichier = rand.Intn(10000) + 5000 // 5-15 KB
				resolution = []string{"1920x1080", "1280x720", "640x480"}[rand.Intn(3)]
				algorithme = "CNN-DeepFace"
			}

			// Date de capture quelques jours après la création du migrant
			dateCapture := migrant.CreatedAt.AddDate(0, 0, rand.Intn(7)+1)
			dateVerification := dateCapture.Add(time.Hour * time.Duration(rand.Intn(4)+1))

			// Qualité basée sur le type de dispositif et l'âge de la capture
			qualite := qualites[rand.Intn(len(qualites))]

			// Score de confiance basé sur la qualité
			var scoreConfiance float64
			switch qualite {
			case "excellente":
				scoreConfiance = 0.90 + rand.Float64()*0.10 // 0.90-1.00
			case "bonne":
				scoreConfiance = 0.80 + rand.Float64()*0.10 // 0.80-0.90
			default: // moyenne
				scoreConfiance = 0.70 + rand.Float64()*0.10 // 0.70-0.80
			}

			// Données biométriques simulées (encodées en base64)
			data := fmt.Sprintf("%s_DATA_%s_%d_%d",
				typeBio,
				migrant.UUID[:8],
				capture,
				rand.Intn(100000))
			donneesBiometriques := base64.StdEncoding.EncodeToString([]byte(data))

			bio := models.Biometrie{
				UUID:                utils.GenerateUUID(),
				MigrantUUID:         migrant.UUID,
				TypeBiometrie:       typeBio,
				IndexDoigt:          indexDoigt,
				QualiteDonnee:       qualite,
				DonneesBiometriques: donneesBiometriques,
				AlgorithmeEncodage:  algorithme,
				TailleFichier:       tailleFichier,
				DateCapture:         dateCapture,
				DisposifCapture:     dispositifs[rand.Intn(len(dispositifs))],
				ResolutionCapture:   resolution,
				OperateurCapture:    fmt.Sprintf("Agent DGM%03d", (i%5)+1),
				Verifie:             scoreConfiance >= 0.75, // Vérifié si score >= 75%
				DateVerification:    &dateVerification,
				ScoreConfiance:      &scoreConfiance,
				Chiffre:             true,
				CleChiffrement:      fmt.Sprintf("AES256_KEY_%s", utils.GenerateUUID()[:16]),
				CreatedAt:           dateCapture,
				UpdatedAt:           dateVerification,
			}

			biometries = append(biometries, bio)
		}
	}

	// Insérer en base
	for _, bio := range biometries {
		if err := db.Create(&bio).Error; err != nil {
			log.Printf("Erreur lors de la création des données biométriques: %v", err)
			continue
		}
	}

	log.Printf("✅ %d données biométriques créées", len(biometries))

	// Statistiques
	statsType := make(map[string]int)
	statsQualite := make(map[string]int)
	totalVerifie := 0

	for _, bio := range biometries {
		statsType[bio.TypeBiometrie]++
		statsQualite[bio.QualiteDonnee]++
		if bio.Verifie {
			totalVerifie++
		}
	}

	log.Println("📊 Distribution des données biométriques:")
	for type_, count := range statsType {
		log.Printf("   - %s: %d", type_, count)
	}
	log.Println("📊 Qualité des captures:")
	for qualite, count := range statsQualite {
		log.Printf("   - %s: %d (%.1f%%)", qualite, count, float64(count)/float64(len(biometries))*100)
	}
	log.Printf("✅ Taux de vérification: %.1f%%", float64(totalVerifie)/float64(len(biometries))*100)

	return nil
}

// simulateAlerts crée des alertes simulées réalistes
func simulateAlerts(db *gorm.DB) error {
	// Récupérer les migrants existants
	var migrants []models.Migrant
	if err := db.Find(&migrants).Error; err != nil {
		return err
	}

	if len(migrants) == 0 {
		return nil
	}

	// Modèles d'alertes par type
	alertesModeles := map[string][]struct {
		Titre               string
		DescriptionTemplate string
		Gravite             string
		JoursExpiration     int
		ActionRequise       string
	}{
		"securite": {
			{"Document d'identité expirant", "Le passeport expire dans %d jours. Renouvellement urgent requis.", "warning", 45, "Contacter l'ambassade pour renouvellement"},
			{"Zone d'origine instable", "La zone d'origine reste instable avec des combats sporadiques. Retour non recommandé.", "danger", 90, "Maintenir en zone sécurisée, surveiller évolution"},
			{"Signalement suspect", "Activité suspecte détectée nécessitant vérification.", "warning", 30, "Enquête de vérification à mener"},
		},
		"sante": {
			{"Suivi médical urgent", "Suivi médical urgent requis suite à symptômes détectés.", "danger", 15, "Orientation vers centre médical MSF ou Croix-Rouge"},
			{"Vaccination incomplète", "Carnet de vaccination incomplet. Mise à jour nécessaire.", "warning", 60, "Compléter le programme de vaccination"},
			{"Dépistage sanitaire", "Dépistage sanitaire de routine à effectuer.", "info", 30, "Planifier rendez-vous médical"},
		},
		"administrative": {
			{"Renouvellement permis de séjour", "Le permis de séjour expire dans %d jours. Renouvellement à entamer.", "warning", 60, "Accompagner dans les démarches administratives"},
			{"Documents manquants", "Dossier incomplet. Documents administratifs manquants.", "warning", 45, "Compléter le dossier avec pièces manquantes"},
			{"Enregistrement biométrique", "Enregistrement biométrique incomplet ou à renouveler.", "info", 90, "Planifier session de capture biométrique"},
		},
		"social": {
			{"Recherche d'opportunités d'emploi", "Demande d'assistance pour formation professionnelle ou recherche d'emploi.", "info", 90, "Orientation vers programmes de formation"},
			{"Assistance humanitaire", "Besoin d'assistance alimentaire ou matérielle urgente.", "danger", 15, "Coordination avec ONG partenaires (HCR, PAM)"},
			{"Recherche de membres de famille", "Recherche active de membres de famille séparés.", "warning", 120, "Inscription au programme Croix-Rouge"},
			{"Scolarisation des enfants", "Enfants non scolarisés nécessitant inscription.", "warning", 60, "Contact avec établissements scolaires locaux"},
		},
		"juridique": {
			{"Procédure d'asile en cours", "Demande d'asile en cours d'examen. Suivi requis.", "info", 180, "Suivi régulier du dossier avec autorités"},
			{"Régularisation statut", "Procédure de régularisation du statut migratoire à initier.", "warning", 90, "Entamer démarches de régularisation"},
		},
	}

	var alerts []models.Alert

	typesAlertes := []string{"securite", "sante", "administrative", "social", "juridique"}
	responsables := []string{"Agent DGM001", "Agent DGM002", "Agent DGM003", "Coordinateur UNHCR", "MSF Médecin", "Croix-Rouge RDC"}

	// Créer 1-3 alertes par migrant selon leur profil
	for _, migrant := range migrants {
		numAlertes := rand.Intn(3) + 1

		// Plus d'alertes pour les déplacés internes et demandeurs d'asile
		if migrant.StatutMigratoire == "deplace_interne" || migrant.StatutMigratoire == "demandeur_asile" {
			numAlertes = rand.Intn(2) + 2 // 2-3 alertes
		}

		for i := 0; i < numAlertes; i++ {
			typeAlerte := typesAlertes[rand.Intn(len(typesAlertes))]
			modeles := alertesModeles[typeAlerte]
			modele := modeles[rand.Intn(len(modeles))]

			// Date de création de l'alerte (après création du migrant)
			joursDepuisMigrant := rand.Intn(60) + 5
			dateCreation := migrant.CreatedAt.AddDate(0, 0, joursDepuisMigrant)
			dateExpiration := dateCreation.AddDate(0, 0, modele.JoursExpiration)

			// Description personnalisée
			description := modele.DescriptionTemplate
			if typeAlerte == "securite" && modele.Titre == "Document d'identité expirant" {
				description = fmt.Sprintf(modele.DescriptionTemplate, modele.JoursExpiration)
			}

			// Statut de l'alerte (80% actives, 20% résolues)
			statut := "active"
			var dateResolution *time.Time
			if rand.Float64() < 0.20 {
				statut = "resolved"
				dateRes := dateCreation.AddDate(0, 0, rand.Intn(modele.JoursExpiration/2))
				dateResolution = &dateRes
			}

			alert := models.Alert{
				UUID:                utils.GenerateUUID(),
				MigrantUUID:         migrant.UUID,
				TypeAlerte:          typeAlerte,
				NiveauGravite:       modele.Gravite,
				Titre:               modele.Titre,
				Description:         description,
				Statut:              statut,
				DateExpiration:      &dateExpiration,
				ActionRequise:       modele.ActionRequise,
				PersonneResponsable: responsables[rand.Intn(len(responsables))],
				DateResolution:      dateResolution,
				CreatedAt:           dateCreation,
				UpdatedAt:           dateCreation,
			}

			alerts = append(alerts, alert)
		}
	}

	// Insérer en base
	for _, alert := range alerts {
		if err := db.Create(&alert).Error; err != nil {
			log.Printf("Erreur lors de la création de l'alerte: %v", err)
			continue
		}
	}

	log.Printf("✅ %d alertes créées", len(alerts))

	// Statistiques
	statsType := make(map[string]int)
	statsGravite := make(map[string]int)
	statsStatut := make(map[string]int)

	for _, alert := range alerts {
		statsType[alert.TypeAlerte]++
		statsGravite[alert.NiveauGravite]++
		statsStatut[alert.Statut]++
	}

	log.Println("📊 Distribution des alertes par type:")
	for type_, count := range statsType {
		log.Printf("   - %s: %d (%.1f%%)", type_, count, float64(count)/float64(len(alerts))*100)
	}
	log.Println("📊 Niveau de gravité:")
	for gravite, count := range statsGravite {
		log.Printf("   - %s: %d", gravite, count)
	}
	log.Println("📊 Statut des alertes:")
	for statut, count := range statsStatut {
		log.Printf("   - %s: %d (%.1f%%)", statut, count, float64(count)/float64(len(alerts))*100)
	}

	return nil
}
