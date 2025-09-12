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

	// 2. Créer les migrants
	log.Println("2. Création des migrants...")
	if err := simulateMigrants(db); err != nil {
		return fmt.Errorf("erreur lors de la simulation des migrants: %v", err)
	}

	// 3. Créer les géolocalisations (dépendent des migrants)
	log.Println("3. Création des géolocalisations...")
	if err := simulateGeolocalisations(db); err != nil {
		return fmt.Errorf("erreur lors de la simulation des géolocalisations: %v", err)
	}

	// 4. Créer les motifs de déplacement (dépendent des migrants)
	log.Println("4. Création des motifs de déplacement...")
	if err := simulateMotifDeplacements(db); err != nil {
		return fmt.Errorf("erreur lors de la simulation des motifs de déplacement: %v", err)
	}

	// 5. Créer les données biométriques (dépendent des migrants)
	log.Println("5. Création des données biométriques...")
	if err := simulateBiometries(db); err != nil {
		return fmt.Errorf("erreur lors de la simulation des biométries: %v", err)
	}

	// 6. Créer les alertes (dépendent des migrants)
	log.Println("6. Création des alertes...")
	if err := simulateAlerts(db); err != nil {
		return fmt.Errorf("erreur lors de la simulation des alertes: %v", err)
	}

	log.Println("=== SIMULATION TERMINÉE AVEC SUCCÈS ===")
	log.Println("Données créées:")
	log.Println("- 5 utilisateurs du système")
	log.Println("- 8 migrants de différentes nationalités")
	log.Println("- Multiple géolocalisations par migrant")
	log.Println("- Motifs de déplacement variés")
	log.Println("- Données biométriques complètes")
	log.Println("- 8 alertes de différents types")

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

// simulateMigrants crée des migrants simulés étalés sur 3 mois
func simulateMigrants(db *gorm.DB) error {
	migrants := []models.Migrant{
		// === MIGRANTS INTERNATIONAUX - JUIN 2025 ===
		{
			UUID:                  utils.GenerateUUID(),
			NumeroIdentifiant:     "MIG2025001",
			Nom:                   "OUEDRAOGO",
			Prenom:                "Amadou",
			DateNaissance:         time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC),
			LieuNaissance:         "Ouagadougou",
			Sexe:                  "M",
			Nationalite:           "Burkinabè",
			TypeDocument:          "passport",
			NumeroDocument:        "BF1234567",
			DateEmissionDoc:       &[]time.Time{time.Date(2020, 3, 10, 0, 0, 0, 0, time.UTC)}[0],
			DateExpirationDoc:     &[]time.Time{time.Date(2030, 3, 10, 0, 0, 0, 0, time.UTC)}[0],
			AutoriteEmission:      "République du Burkina Faso",
			Telephone:             "+22670123456",
			Email:                 "amadou.ouedraogo@email.com",
			AdresseActuelle:       "Avenue Kasavubu, N°45",
			VilleActuelle:         "Kinshasa",
			PaysActuel:            "République Démocratique du Congo",
			SituationMatrimoniale: "marie",
			NombreEnfants:         2,
			StatutMigratoire:      "regulier",
			DateEntree:            &[]time.Time{time.Date(2025, 6, 5, 0, 0, 0, 0, time.UTC)}[0],
			PointEntree:           "Aéroport de N'djili",
			PaysOrigine:           "Burkina Faso",
			CreatedAt:             time.Date(2025, 6, 5, 10, 30, 0, 0, time.UTC),
			UpdatedAt:             time.Date(2025, 6, 5, 10, 30, 0, 0, time.UTC),
		},
		{
			UUID:                  utils.GenerateUUID(),
			NumeroIdentifiant:     "MIG2025002",
			Nom:                   "SANKARA",
			Prenom:                "Fanta",
			DateNaissance:         time.Date(1987, 9, 12, 0, 0, 0, 0, time.UTC),
			LieuNaissance:         "Bobo-Dioulasso",
			Sexe:                  "F",
			Nationalite:           "Burkinabè",
			TypeDocument:          "passport",
			NumeroDocument:        "BF9876543",
			DateEmissionDoc:       &[]time.Time{time.Date(2021, 7, 15, 0, 0, 0, 0, time.UTC)}[0],
			DateExpirationDoc:     &[]time.Time{time.Date(2031, 7, 15, 0, 0, 0, 0, time.UTC)}[0],
			AutoriteEmission:      "République du Burkina Faso",
			Telephone:             "+22675789123",
			Email:                 "fanta.sankara@email.com",
			AdresseActuelle:       "Avenue de la Justice, N°78",
			VilleActuelle:         "Kinshasa",
			PaysActuel:            "République Démocratique du Congo",
			SituationMatrimoniale: "celibataire",
			NombreEnfants:         1,
			StatutMigratoire:      "regulier",
			DateEntree:            &[]time.Time{time.Date(2025, 6, 12, 0, 0, 0, 0, time.UTC)}[0],
			PointEntree:           "Aéroport de N'djili",
			PaysOrigine:           "Burkina Faso",
			CreatedAt:             time.Date(2025, 6, 12, 14, 15, 0, 0, time.UTC),
			UpdatedAt:             time.Date(2025, 6, 12, 14, 15, 0, 0, time.UTC),
		},
		{
			UUID:                  utils.GenerateUUID(),
			NumeroIdentifiant:     "MIG2025003",
			Nom:                   "ZONGO",
			Prenom:                "Rasmané",
			DateNaissance:         time.Date(1993, 2, 28, 0, 0, 0, 0, time.UTC),
			LieuNaissance:         "Ouahigouya",
			Sexe:                  "M",
			Nationalite:           "Burkinabè",
			TypeDocument:          "carte_identite",
			NumeroDocument:        "BF5432109",
			Telephone:             "+22678345612",
			Email:                 "rasmane.zongo@email.com",
			AdresseActuelle:       "Commune de Bandalungwa, Rue 15",
			VilleActuelle:         "Kinshasa",
			PaysActuel:            "République Démocratique du Congo",
			SituationMatrimoniale: "marie",
			NombreEnfants:         3,
			StatutMigratoire:      "demandeur_asile",
			DateEntree:            &[]time.Time{time.Date(2025, 6, 18, 0, 0, 0, 0, time.UTC)}[0],
			PointEntree:           "Frontière de Bangui",
			PaysOrigine:           "Burkina Faso",
			CreatedAt:             time.Date(2025, 6, 18, 9, 45, 0, 0, time.UTC),
			UpdatedAt:             time.Date(2025, 6, 18, 9, 45, 0, 0, time.UTC),
		},
		{
			UUID:                  utils.GenerateUUID(),
			NumeroIdentifiant:     "MIG2025004",
			Nom:                   "TRAORE",
			Prenom:                "Aïssata",
			DateNaissance:         time.Date(1985, 8, 22, 0, 0, 0, 0, time.UTC),
			LieuNaissance:         "Bamako",
			Sexe:                  "F",
			Nationalite:           "Malienne",
			TypeDocument:          "carte_identite",
			NumeroDocument:        "ML9876543",
			Telephone:             "+22365123456",
			Email:                 "aissata.traore@email.com",
			AdresseActuelle:       "Boulevard Triomphal, N°78",
			VilleActuelle:         "Kinshasa",
			PaysActuel:            "République Démocratique du Congo",
			SituationMatrimoniale: "celibataire",
			NombreEnfants:         0,
			StatutMigratoire:      "demandeur_asile",
			DateEntree:            &[]time.Time{time.Date(2025, 6, 25, 0, 0, 0, 0, time.UTC)}[0],
			PointEntree:           "Frontière de Bangui",
			PaysOrigine:           "Mali",
			CreatedAt:             time.Date(2025, 6, 25, 16, 20, 0, 0, time.UTC),
			UpdatedAt:             time.Date(2025, 6, 25, 16, 20, 0, 0, time.UTC),
		},
		{
			UUID:                  utils.GenerateUUID(),
			NumeroIdentifiant:     "MIG2025005",
			Nom:                   "KEITA",
			Prenom:                "Moussa",
			DateNaissance:         time.Date(1991, 11, 7, 0, 0, 0, 0, time.UTC),
			LieuNaissance:         "Kayes",
			Sexe:                  "M",
			Nationalite:           "Malienne",
			TypeDocument:          "passport",
			NumeroDocument:        "ML7654321",
			DateEmissionDoc:       &[]time.Time{time.Date(2022, 4, 8, 0, 0, 0, 0, time.UTC)}[0],
			DateExpirationDoc:     &[]time.Time{time.Date(2032, 4, 8, 0, 0, 0, 0, time.UTC)}[0],
			AutoriteEmission:      "République du Mali",
			Telephone:             "+22367891234",
			Email:                 "moussa.keita@email.com",
			AdresseActuelle:       "Avenue Kasa-Vubu, N°156",
			VilleActuelle:         "Kinshasa",
			PaysActuel:            "République Démocratique du Congo",
			SituationMatrimoniale: "divorce",
			NombreEnfants:         2,
			StatutMigratoire:      "refugie",
			DateEntree:            &[]time.Time{time.Date(2025, 6, 30, 0, 0, 0, 0, time.UTC)}[0],
			PointEntree:           "Frontière de Matadi",
			PaysOrigine:           "Mali",
			CreatedAt:             time.Date(2025, 6, 30, 11, 0, 0, 0, time.UTC),
			UpdatedAt:             time.Date(2025, 6, 30, 11, 0, 0, 0, time.UTC),
		},

		// === MIGRANTS INTERNATIONAUX - JUILLET 2025 ===
		{
			UUID:                  utils.GenerateUUID(),
			NumeroIdentifiant:     "MIG2025006",
			Nom:                   "KONE",
			Prenom:                "Ibrahim",
			DateNaissance:         time.Date(1988, 12, 3, 0, 0, 0, 0, time.UTC),
			LieuNaissance:         "Abidjan",
			Sexe:                  "M",
			Nationalite:           "Ivoirienne",
			TypeDocument:          "passport",
			NumeroDocument:        "CI5551234",
			Telephone:             "+22507123456",
			Email:                 "ibrahim.kone@email.com",
			AdresseActuelle:       "Quartier Matonge, Rue 12",
			VilleActuelle:         "Kinshasa",
			PaysActuel:            "République Démocratique du Congo",
			SituationMatrimoniale: "divorce",
			NombreEnfants:         1,
			StatutMigratoire:      "regulier",
			DateEntree:            &[]time.Time{time.Date(2025, 7, 3, 0, 0, 0, 0, time.UTC)}[0],
			PointEntree:           "Aéroport de N'djili",
			PaysOrigine:           "Côte d'Ivoire",
			CreatedAt:             time.Date(2025, 7, 3, 13, 30, 0, 0, time.UTC),
			UpdatedAt:             time.Date(2025, 7, 3, 13, 30, 0, 0, time.UTC),
		},
		{
			UUID:                  utils.GenerateUUID(),
			NumeroIdentifiant:     "MIG2025007",
			Nom:                   "DIABATE",
			Prenom:                "Mariame",
			DateNaissance:         time.Date(1994, 6, 14, 0, 0, 0, 0, time.UTC),
			LieuNaissance:         "Bouaké",
			Sexe:                  "F",
			Nationalite:           "Ivoirienne",
			TypeDocument:          "carte_identite",
			NumeroDocument:        "CI9988776",
			Telephone:             "+22509876543",
			Email:                 "mariame.diabate@email.com",
			AdresseActuelle:       "Commune de Ngaliema, Avenue des Forces",
			VilleActuelle:         "Kinshasa",
			PaysActuel:            "République Démocratique du Congo",
			SituationMatrimoniale: "marie",
			NombreEnfants:         2,
			StatutMigratoire:      "demandeur_asile",
			DateEntree:            &[]time.Time{time.Date(2025, 7, 10, 0, 0, 0, 0, time.UTC)}[0],
			PointEntree:           "Frontière de Matadi",
			PaysOrigine:           "Côte d'Ivoire",
			CreatedAt:             time.Date(2025, 7, 10, 8, 45, 0, 0, time.UTC),
			UpdatedAt:             time.Date(2025, 7, 10, 8, 45, 0, 0, time.UTC),
		},
		{
			UUID:                  utils.GenerateUUID(),
			NumeroIdentifiant:     "MIG2025008",
			Nom:                   "DIALLO",
			Prenom:                "Fatima",
			DateNaissance:         time.Date(1992, 4, 18, 0, 0, 0, 0, time.UTC),
			LieuNaissance:         "Conakry",
			Sexe:                  "F",
			Nationalite:           "Guinéenne",
			TypeDocument:          "carte_identite",
			NumeroDocument:        "GN2468135",
			Telephone:             "+22462123456",
			Email:                 "fatima.diallo@email.com",
			AdresseActuelle:       "Commune de Kalamu, Avenue Victoire",
			VilleActuelle:         "Kinshasa",
			PaysActuel:            "République Démocratique du Congo",
			SituationMatrimoniale: "marie",
			NombreEnfants:         3,
			StatutMigratoire:      "refugie",
			DateEntree:            &[]time.Time{time.Date(2025, 7, 17, 0, 0, 0, 0, time.UTC)}[0],
			PointEntree:           "Frontière de Matadi",
			PaysOrigine:           "Guinée",
			CreatedAt:             time.Date(2025, 7, 17, 15, 10, 0, 0, time.UTC),
			UpdatedAt:             time.Date(2025, 7, 17, 15, 10, 0, 0, time.UTC),
		},
		{
			UUID:                  utils.GenerateUUID(),
			NumeroIdentifiant:     "MIG2025009",
			Nom:                   "CONDE",
			Prenom:                "Alpha",
			DateNaissance:         time.Date(1989, 1, 30, 0, 0, 0, 0, time.UTC),
			LieuNaissance:         "Labé",
			Sexe:                  "M",
			Nationalite:           "Guinéenne",
			TypeDocument:          "passport",
			NumeroDocument:        "GN1357924",
			DateEmissionDoc:       &[]time.Time{time.Date(2023, 2, 5, 0, 0, 0, 0, time.UTC)}[0],
			DateExpirationDoc:     &[]time.Time{time.Date(2033, 2, 5, 0, 0, 0, 0, time.UTC)}[0],
			AutoriteEmission:      "République de Guinée",
			Telephone:             "+22464789123",
			Email:                 "alpha.conde@email.com",
			AdresseActuelle:       "Quartier Binza, Rue 8",
			VilleActuelle:         "Kinshasa",
			PaysActuel:            "République Démocratique du Congo",
			SituationMatrimoniale: "celibataire",
			NombreEnfants:         0,
			StatutMigratoire:      "regulier",
			DateEntree:            &[]time.Time{time.Date(2025, 7, 24, 0, 0, 0, 0, time.UTC)}[0],
			PointEntree:           "Aéroport de N'djili",
			PaysOrigine:           "Guinée",
			CreatedAt:             time.Date(2025, 7, 24, 12, 0, 0, 0, time.UTC),
			UpdatedAt:             time.Date(2025, 7, 24, 12, 0, 0, 0, time.UTC),
		},
		{
			UUID:                  utils.GenerateUUID(),
			NumeroIdentifiant:     "MIG2025010",
			Nom:                   "SOW",
			Prenom:                "Hadja",
			DateNaissance:         time.Date(1996, 10, 5, 0, 0, 0, 0, time.UTC),
			LieuNaissance:         "Kankan",
			Sexe:                  "F",
			Nationalite:           "Guinéenne",
			TypeDocument:          "carte_identite",
			NumeroDocument:        "GN8642097",
			Telephone:             "+22461357890",
			Email:                 "hadja.sow@email.com",
			AdresseActuelle:       "Avenue Lumumba, N°235",
			VilleActuelle:         "Kinshasa",
			PaysActuel:            "République Démocratique du Congo",
			SituationMatrimoniale: "marie",
			NombreEnfants:         1,
			StatutMigratoire:      "demandeur_asile",
			DateEntree:            &[]time.Time{time.Date(2025, 7, 31, 0, 0, 0, 0, time.UTC)}[0],
			PointEntree:           "Frontière de Bangui",
			PaysOrigine:           "Guinée",
			CreatedAt:             time.Date(2025, 7, 31, 17, 45, 0, 0, time.UTC),
			UpdatedAt:             time.Date(2025, 7, 31, 17, 45, 0, 0, time.UTC),
		},

		// === MIGRANTS INTERNATIONAUX - AOÛT 2025 ===
		{
			UUID:                  utils.GenerateUUID(),
			NumeroIdentifiant:     "MIG2025011",
			Nom:                   "SAWADOGO",
			Prenom:                "Paul",
			DateNaissance:         time.Date(1986, 3, 20, 0, 0, 0, 0, time.UTC),
			LieuNaissance:         "Koudougou",
			Sexe:                  "M",
			Nationalite:           "Burkinabè",
			TypeDocument:          "passport",
			NumeroDocument:        "BF1928374",
			DateEmissionDoc:       &[]time.Time{time.Date(2022, 11, 12, 0, 0, 0, 0, time.UTC)}[0],
			DateExpirationDoc:     &[]time.Time{time.Date(2032, 11, 12, 0, 0, 0, 0, time.UTC)}[0],
			AutoriteEmission:      "République du Burkina Faso",
			Telephone:             "+22679135468",
			Email:                 "paul.sawadogo@email.com",
			AdresseActuelle:       "Commune de Lemba, Route de Matadi",
			VilleActuelle:         "Kinshasa",
			PaysActuel:            "République Démocratique du Congo",
			SituationMatrimoniale: "marie",
			NombreEnfants:         4,
			StatutMigratoire:      "refugie",
			DateEntree:            &[]time.Time{time.Date(2025, 8, 2, 0, 0, 0, 0, time.UTC)}[0],
			PointEntree:           "Frontière de Bangui",
			PaysOrigine:           "Burkina Faso",
			CreatedAt:             time.Date(2025, 8, 2, 9, 20, 0, 0, time.UTC),
			UpdatedAt:             time.Date(2025, 8, 2, 9, 20, 0, 0, time.UTC),
		},
		{
			UUID:                  utils.GenerateUUID(),
			NumeroIdentifiant:     "MIG2025012",
			Nom:                   "BAGAYOKO",
			Prenom:                "Aminata",
			DateNaissance:         time.Date(1990, 7, 11, 0, 0, 0, 0, time.UTC),
			LieuNaissance:         "Sikasso",
			Sexe:                  "F",
			Nationalite:           "Malienne",
			TypeDocument:          "carte_identite",
			NumeroDocument:        "ML3698741",
			Telephone:             "+22368521479",
			Email:                 "aminata.bagayoko@email.com",
			AdresseActuelle:       "Quartier Masina, Rue 25",
			VilleActuelle:         "Kinshasa",
			PaysActuel:            "République Démocratique du Congo",
			SituationMatrimoniale: "veuf",
			NombreEnfants:         2,
			StatutMigratoire:      "demandeur_asile",
			DateEntree:            &[]time.Time{time.Date(2025, 8, 8, 0, 0, 0, 0, time.UTC)}[0],
			PointEntree:           "Frontière de Matadi",
			PaysOrigine:           "Mali",
			CreatedAt:             time.Date(2025, 8, 8, 14, 30, 0, 0, time.UTC),
			UpdatedAt:             time.Date(2025, 8, 8, 14, 30, 0, 0, time.UTC),
		},
		{
			UUID:                  utils.GenerateUUID(),
			NumeroIdentifiant:     "MIG2025013",
			Nom:                   "YAO",
			Prenom:                "Kouadio",
			DateNaissance:         time.Date(1984, 12, 25, 0, 0, 0, 0, time.UTC),
			LieuNaissance:         "Yamoussoukro",
			Sexe:                  "M",
			Nationalite:           "Ivoirienne",
			TypeDocument:          "passport",
			NumeroDocument:        "CI7419630",
			DateEmissionDoc:       &[]time.Time{time.Date(2021, 9, 18, 0, 0, 0, 0, time.UTC)}[0],
			DateExpirationDoc:     &[]time.Time{time.Date(2031, 9, 18, 0, 0, 0, 0, time.UTC)}[0],
			AutoriteEmission:      "République de Côte d'Ivoire",
			Telephone:             "+22502741963",
			Email:                 "kouadio.yao@email.com",
			AdresseActuelle:       "Boulevard du 30 Juin, N°89",
			VilleActuelle:         "Kinshasa",
			PaysActuel:            "République Démocratique du Congo",
			SituationMatrimoniale: "marie",
			NombreEnfants:         3,
			StatutMigratoire:      "regulier",
			DateEntree:            &[]time.Time{time.Date(2025, 8, 15, 0, 0, 0, 0, time.UTC)}[0],
			PointEntree:           "Aéroport de N'djili",
			PaysOrigine:           "Côte d'Ivoire",
			CreatedAt:             time.Date(2025, 8, 15, 11, 15, 0, 0, time.UTC),
			UpdatedAt:             time.Date(2025, 8, 15, 11, 15, 0, 0, time.UTC),
		},
		{
			UUID:                  utils.GenerateUUID(),
			NumeroIdentifiant:     "MIG2025014",
			Nom:                   "BARRY",
			Prenom:                "Mamadou",
			DateNaissance:         time.Date(1987, 5, 8, 0, 0, 0, 0, time.UTC),
			LieuNaissance:         "Boké",
			Sexe:                  "M",
			Nationalite:           "Guinéenne",
			TypeDocument:          "carte_identite",
			NumeroDocument:        "GN9517534",
			Telephone:             "+22463852741",
			Email:                 "mamadou.barry@email.com",
			AdresseActuelle:       "Commune de Ndjili, Avenue de l'Aéroport",
			VilleActuelle:         "Kinshasa",
			PaysActuel:            "République Démocratique du Congo",
			SituationMatrimoniale: "divorce",
			NombreEnfants:         1,
			StatutMigratoire:      "refugie",
			DateEntree:            &[]time.Time{time.Date(2025, 8, 22, 0, 0, 0, 0, time.UTC)}[0],
			PointEntree:           "Frontière de Bangui",
			PaysOrigine:           "Guinée",
			CreatedAt:             time.Date(2025, 8, 22, 16, 40, 0, 0, time.UTC),
			UpdatedAt:             time.Date(2025, 8, 22, 16, 40, 0, 0, time.UTC),
		},
		{
			UUID:                  utils.GenerateUUID(),
			NumeroIdentifiant:     "MIG2025015",
			Nom:                   "OUATTARA",
			Prenom:                "Salimata",
			DateNaissance:         time.Date(1995, 2, 14, 0, 0, 0, 0, time.UTC),
			LieuNaissance:         "Korhogo",
			Sexe:                  "F",
			Nationalite:           "Ivoirienne",
			TypeDocument:          "carte_identite",
			NumeroDocument:        "CI8520741",
			Telephone:             "+22508639517",
			Email:                 "salimata.ouattara@email.com",
			AdresseActuelle:       "Quartier Righini, Rue 7",
			VilleActuelle:         "Kinshasa",
			PaysActuel:            "République Démocratique du Congo",
			SituationMatrimoniale: "celibataire",
			NombreEnfants:         0,
			StatutMigratoire:      "demandeur_asile",
			DateEntree:            &[]time.Time{time.Date(2025, 8, 29, 0, 0, 0, 0, time.UTC)}[0],
			PointEntree:           "Frontière de Matadi",
			PaysOrigine:           "Côte d'Ivoire",
			CreatedAt:             time.Date(2025, 8, 29, 10, 5, 0, 0, time.UTC),
			UpdatedAt:             time.Date(2025, 8, 29, 10, 5, 0, 0, time.UTC),
		},

		// === DÉPLACÉS INTERNES RDC - JUIN 2025 ===
		{
			UUID:                  utils.GenerateUUID(),
			NumeroIdentifiant:     "DPI2025001",
			Nom:                   "KABILA",
			Prenom:                "Jeanne",
			DateNaissance:         time.Date(1995, 2, 28, 0, 0, 0, 0, time.UTC),
			LieuNaissance:         "Goma",
			Sexe:                  "F",
			Nationalite:           "Congolaise (RDC)",
			TypeDocument:          "carte_identite",
			NumeroDocument:        "CD1234567890",
			DateEmissionDoc:       &[]time.Time{time.Date(2022, 5, 15, 0, 0, 0, 0, time.UTC)}[0],
			DateExpirationDoc:     &[]time.Time{time.Date(2032, 5, 15, 0, 0, 0, 0, time.UTC)}[0],
			AutoriteEmission:      "République Démocratique du Congo",
			Telephone:             "+243987123456",
			Email:                 "jeanne.kabila@email.cd",
			AdresseActuelle:       "Camp de déplacés de Mugunga, Goma",
			VilleActuelle:         "Goma",
			PaysActuel:            "République Démocratique du Congo",
			SituationMatrimoniale: "veuf",
			NombreEnfants:         2,
			PersonneContact:       "KABILA Pierre",
			TelephoneContact:      "+243876543210",
			StatutMigratoire:      "deplace_interne",
			DateEntree:            &[]time.Time{time.Date(2025, 6, 8, 0, 0, 0, 0, time.UTC)}[0],
			PointEntree:           "Déplacement depuis Rutshuru",
			PaysOrigine:           "République Démocratique du Congo",
			CreatedAt:             time.Date(2025, 6, 8, 7, 30, 0, 0, time.UTC),
			UpdatedAt:             time.Date(2025, 6, 8, 7, 30, 0, 0, time.UTC),
		},
		{
			UUID:                  utils.GenerateUUID(),
			NumeroIdentifiant:     "DPI2025002",
			Nom:                   "MBUYI",
			Prenom:                "Jean-Baptiste",
			DateNaissance:         time.Date(1982, 11, 12, 0, 0, 0, 0, time.UTC),
			LieuNaissance:         "Kananga",
			Sexe:                  "M",
			Nationalite:           "Congolaise (RDC)",
			TypeDocument:          "carte_identite",
			NumeroDocument:        "CD2345678901",
			DateEmissionDoc:       &[]time.Time{time.Date(2021, 8, 20, 0, 0, 0, 0, time.UTC)}[0],
			DateExpirationDoc:     &[]time.Time{time.Date(2031, 8, 20, 0, 0, 0, 0, time.UTC)}[0],
			AutoriteEmission:      "République Démocratique du Congo",
			Telephone:             "+243812345678",
			Email:                 "jeanbaptiste.mbuyi@email.cd",
			AdresseActuelle:       "Commune de Kalamu, Kinshasa",
			VilleActuelle:         "Kinshasa",
			PaysActuel:            "République Démocratique du Congo",
			SituationMatrimoniale: "marie",
			NombreEnfants:         4,
			PersonneContact:       "MBUYI Marie",
			TelephoneContact:      "+243898765432",
			StatutMigratoire:      "deplace_interne",
			DateEntree:            &[]time.Time{time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)}[0],
			PointEntree:           "Migration interne depuis Kananga",
			PaysOrigine:           "République Démocratique du Congo",
			CreatedAt:             time.Date(2025, 6, 15, 11, 45, 0, 0, time.UTC),
			UpdatedAt:             time.Date(2025, 6, 15, 11, 45, 0, 0, time.UTC),
		},
		{
			UUID:                  utils.GenerateUUID(),
			NumeroIdentifiant:     "DPI2025003",
			Nom:                   "TSHIMANGA",
			Prenom:                "Grace",
			DateNaissance:         time.Date(1993, 8, 3, 0, 0, 0, 0, time.UTC),
			LieuNaissance:         "Lubumbashi",
			Sexe:                  "F",
			Nationalite:           "Congolaise (RDC)",
			TypeDocument:          "carte_identite",
			NumeroDocument:        "CD7890123456",
			DateEmissionDoc:       &[]time.Time{time.Date(2023, 1, 12, 0, 0, 0, 0, time.UTC)}[0],
			DateExpirationDoc:     &[]time.Time{time.Date(2033, 1, 12, 0, 0, 0, 0, time.UTC)}[0],
			AutoriteEmission:      "République Démocratique du Congo",
			Telephone:             "+243973456789",
			Email:                 "grace.tshimanga@email.cd",
			AdresseActuelle:       "Commune de Masina, Avenue des Martyrs",
			VilleActuelle:         "Kinshasa",
			PaysActuel:            "République Démocratique du Congo",
			SituationMatrimoniale: "celibataire",
			NombreEnfants:         2,
			PersonneContact:       "TSHIMANGA Joseph",
			TelephoneContact:      "+243854123789",
			StatutMigratoire:      "deplace_interne",
			DateEntree:            &[]time.Time{time.Date(2025, 6, 22, 0, 0, 0, 0, time.UTC)}[0],
			PointEntree:           "Migration interne depuis Lubumbashi",
			PaysOrigine:           "République Démocratique du Congo",
			CreatedAt:             time.Date(2025, 6, 22, 15, 20, 0, 0, time.UTC),
			UpdatedAt:             time.Date(2025, 6, 22, 15, 20, 0, 0, time.UTC),
		},
		{
			UUID:                  utils.GenerateUUID(),
			NumeroIdentifiant:     "DPI2025004",
			Nom:                   "LUNDA",
			Prenom:                "Prosper",
			DateNaissance:         time.Date(1988, 4, 17, 0, 0, 0, 0, time.UTC),
			LieuNaissance:         "Mbuji-Mayi",
			Sexe:                  "M",
			Nationalite:           "Congolaise (RDC)",
			TypeDocument:          "carte_identite",
			NumeroDocument:        "CD4567890123",
			DateEmissionDoc:       &[]time.Time{time.Date(2022, 3, 25, 0, 0, 0, 0, time.UTC)}[0],
			DateExpirationDoc:     &[]time.Time{time.Date(2032, 3, 25, 0, 0, 0, 0, time.UTC)}[0],
			AutoriteEmission:      "République Démocratique du Congo",
			Telephone:             "+243981234567",
			Email:                 "prosper.lunda@email.cd",
			AdresseActuelle:       "Commune de Ngaba, Rue 42",
			VilleActuelle:         "Kinshasa",
			PaysActuel:            "République Démocratique du Congo",
			SituationMatrimoniale: "marie",
			NombreEnfants:         3,
			PersonneContact:       "LUNDA Esperance",
			TelephoneContact:      "+243876234567",
			StatutMigratoire:      "deplace_interne",
			DateEntree:            &[]time.Time{time.Date(2025, 6, 28, 0, 0, 0, 0, time.UTC)}[0],
			PointEntree:           "Migration interne depuis Mbuji-Mayi",
			PaysOrigine:           "République Démocratique du Congo",
			CreatedAt:             time.Date(2025, 6, 28, 9, 10, 0, 0, time.UTC),
			UpdatedAt:             time.Date(2025, 6, 28, 9, 10, 0, 0, time.UTC),
		},

		// === DÉPLACÉS INTERNES RDC - JUILLET 2025 ===
		{
			UUID:                  utils.GenerateUUID(),
			NumeroIdentifiant:     "DPI2025005",
			Nom:                   "NGOY",
			Prenom:                "Espérance",
			DateNaissance:         time.Date(1998, 7, 8, 0, 0, 0, 0, time.UTC),
			LieuNaissance:         "Bunia",
			Sexe:                  "F",
			Nationalite:           "Congolaise (RDC)",
			TypeDocument:          "carte_identite",
			NumeroDocument:        "CD3456789012",
			DateEmissionDoc:       &[]time.Time{time.Date(2023, 3, 10, 0, 0, 0, 0, time.UTC)}[0],
			DateExpirationDoc:     &[]time.Time{time.Date(2033, 3, 10, 0, 0, 0, 0, time.UTC)}[0],
			AutoriteEmission:      "République Démocratique du Congo",
			Telephone:             "+243976543210",
			Email:                 "esperance.ngoy@email.cd",
			AdresseActuelle:       "Site de déplacés de Rhoe, Bunia",
			VilleActuelle:         "Bunia",
			PaysActuel:            "République Démocratique du Congo",
			SituationMatrimoniale: "celibataire",
			NombreEnfants:         1,
			PersonneContact:       "NGOY Paulin",
			TelephoneContact:      "+243854321098",
			StatutMigratoire:      "deplace_interne",
			DateEntree:            &[]time.Time{time.Date(2025, 7, 5, 0, 0, 0, 0, time.UTC)}[0],
			PointEntree:           "Fuite depuis Djugu",
			PaysOrigine:           "République Démocratique du Congo",
			CreatedAt:             time.Date(2025, 7, 5, 14, 20, 0, 0, time.UTC),
			UpdatedAt:             time.Date(2025, 7, 5, 14, 20, 0, 0, time.UTC),
		},
		{
			UUID:                  utils.GenerateUUID(),
			NumeroIdentifiant:     "DPI2025006",
			Nom:                   "KASONGO",
			Prenom:                "Patient",
			DateNaissance:         time.Date(1975, 4, 25, 0, 0, 0, 0, time.UTC),
			LieuNaissance:         "Beni",
			Sexe:                  "M",
			Nationalite:           "Congolaise (RDC)",
			TypeDocument:          "carte_identite",
			NumeroDocument:        "CD4567890123",
			DateEmissionDoc:       &[]time.Time{time.Date(2020, 12, 8, 0, 0, 0, 0, time.UTC)}[0],
			DateExpirationDoc:     &[]time.Time{time.Date(2030, 12, 8, 0, 0, 0, 0, time.UTC)}[0],
			AutoriteEmission:      "République Démocratique du Congo",
			Telephone:             "+243823456789",
			Email:                 "patient.kasongo@email.cd",
			AdresseActuelle:       "Commune de Lemba, Kinshasa",
			VilleActuelle:         "Kinshasa",
			PaysActuel:            "République Démocratique du Congo",
			SituationMatrimoniale: "marie",
			NombreEnfants:         5,
			PersonneContact:       "KASONGO Alphonsine",
			TelephoneContact:      "+243876543219",
			StatutMigratoire:      "deplace_interne",
			DateEntree:            &[]time.Time{time.Date(2025, 7, 12, 0, 0, 0, 0, time.UTC)}[0],
			PointEntree:           "Évacuation depuis Beni",
			PaysOrigine:           "République Démocratique du Congo",
			CreatedAt:             time.Date(2025, 7, 12, 6, 45, 0, 0, time.UTC),
			UpdatedAt:             time.Date(2025, 7, 12, 6, 45, 0, 0, time.UTC),
		},
		{
			UUID:                  utils.GenerateUUID(),
			NumeroIdentifiant:     "DPI2025007",
			Nom:                   "WAMBA",
			Prenom:                "Christine",
			DateNaissance:         time.Date(1991, 9, 16, 0, 0, 0, 0, time.UTC),
			LieuNaissance:         "Kisangani",
			Sexe:                  "F",
			Nationalite:           "Congolaise (RDC)",
			TypeDocument:          "carte_identite",
			NumeroDocument:        "CD6789012345",
			DateEmissionDoc:       &[]time.Time{time.Date(2022, 6, 30, 0, 0, 0, 0, time.UTC)}[0],
			DateExpirationDoc:     &[]time.Time{time.Date(2032, 6, 30, 0, 0, 0, 0, time.UTC)}[0],
			AutoriteEmission:      "République Démocratique du Congo",
			Telephone:             "+243984567123",
			Email:                 "christine.wamba@email.cd",
			AdresseActuelle:       "Commune de Kintambo, Avenue Mpolo",
			VilleActuelle:         "Kinshasa",
			PaysActuel:            "République Démocratique du Congo",
			SituationMatrimoniale: "divorce",
			NombreEnfants:         2,
			PersonneContact:       "WAMBA Sylvain",
			TelephoneContact:      "+243867891234",
			StatutMigratoire:      "deplace_interne",
			DateEntree:            &[]time.Time{time.Date(2025, 7, 19, 0, 0, 0, 0, time.UTC)}[0],
			PointEntree:           "Migration interne depuis Kisangani",
			PaysOrigine:           "République Démocratique du Congo",
			CreatedAt:             time.Date(2025, 7, 19, 10, 30, 0, 0, time.UTC),
			UpdatedAt:             time.Date(2025, 7, 19, 10, 30, 0, 0, time.UTC),
		},
		{
			UUID:                  utils.GenerateUUID(),
			NumeroIdentifiant:     "DPI2025008",
			Nom:                   "MUKENDI",
			Prenom:                "Serge",
			DateNaissance:         time.Date(1986, 12, 1, 0, 0, 0, 0, time.UTC),
			LieuNaissance:         "Kolwezi",
			Sexe:                  "M",
			Nationalite:           "Congolaise (RDC)",
			TypeDocument:          "carte_identite",
			NumeroDocument:        "CD9012345678",
			DateEmissionDoc:       &[]time.Time{time.Date(2021, 11, 18, 0, 0, 0, 0, time.UTC)}[0],
			DateExpirationDoc:     &[]time.Time{time.Date(2031, 11, 18, 0, 0, 0, 0, time.UTC)}[0],
			AutoriteEmission:      "République Démocratique du Congo",
			Telephone:             "+243978123456",
			Email:                 "serge.mukendi@email.cd",
			AdresseActuelle:       "Commune de Selembao, Rue 18",
			VilleActuelle:         "Kinshasa",
			PaysActuel:            "République Démocratique du Congo",
			SituationMatrimoniale: "marie",
			NombreEnfants:         3,
			PersonneContact:       "MUKENDI Ange",
			TelephoneContact:      "+243889012345",
			StatutMigratoire:      "deplace_interne",
			DateEntree:            &[]time.Time{time.Date(2025, 7, 26, 0, 0, 0, 0, time.UTC)}[0],
			PointEntree:           "Migration interne depuis Kolwezi",
			PaysOrigine:           "République Démocratique du Congo",
			CreatedAt:             time.Date(2025, 7, 26, 13, 15, 0, 0, time.UTC),
			UpdatedAt:             time.Date(2025, 7, 26, 13, 15, 0, 0, time.UTC),
		},

		// === DÉPLACÉS INTERNES RDC - AOÛT 2025 ===
		{
			UUID:                  utils.GenerateUUID(),
			NumeroIdentifiant:     "DPI2025009",
			Nom:                   "ILUNGA",
			Prenom:                "Honoré",
			DateNaissance:         time.Date(1980, 6, 12, 0, 0, 0, 0, time.UTC),
			LieuNaissance:         "Kamina",
			Sexe:                  "M",
			Nationalite:           "Congolaise (RDC)",
			TypeDocument:          "carte_identite",
			NumeroDocument:        "CD1357902468",
			DateEmissionDoc:       &[]time.Time{time.Date(2020, 4, 22, 0, 0, 0, 0, time.UTC)}[0],
			DateExpirationDoc:     &[]time.Time{time.Date(2030, 4, 22, 0, 0, 0, 0, time.UTC)}[0],
			AutoriteEmission:      "République Démocratique du Congo",
			Telephone:             "+243985123467",
			Email:                 "honore.ilunga@email.cd",
			AdresseActuelle:       "Commune de Bumbu, Avenue des Usines",
			VilleActuelle:         "Kinshasa",
			PaysActuel:            "République Démocratique du Congo",
			SituationMatrimoniale: "marie",
			NombreEnfants:         6,
			PersonneContact:       "ILUNGA Beatrice",
			TelephoneContact:      "+243876543098",
			StatutMigratoire:      "deplace_interne",
			DateEntree:            &[]time.Time{time.Date(2025, 8, 3, 0, 0, 0, 0, time.UTC)}[0],
			PointEntree:           "Migration interne depuis Kamina",
			PaysOrigine:           "République Démocratique du Congo",
			CreatedAt:             time.Date(2025, 8, 3, 8, 0, 0, 0, time.UTC),
			UpdatedAt:             time.Date(2025, 8, 3, 8, 0, 0, 0, time.UTC),
		},
		{
			UUID:                  utils.GenerateUUID(),
			NumeroIdentifiant:     "DPI2025010",
			Nom:                   "KALALA",
			Prenom:                "Noella",
			DateNaissance:         time.Date(1994, 3, 7, 0, 0, 0, 0, time.UTC),
			LieuNaissance:         "Uvira",
			Sexe:                  "F",
			Nationalite:           "Congolaise (RDC)",
			TypeDocument:          "carte_identite",
			NumeroDocument:        "CD2468013579",
			DateEmissionDoc:       &[]time.Time{time.Date(2023, 9, 14, 0, 0, 0, 0, time.UTC)}[0],
			DateExpirationDoc:     &[]time.Time{time.Date(2033, 9, 14, 0, 0, 0, 0, time.UTC)}[0],
			AutoriteEmission:      "République Démocratique du Congo",
			Telephone:             "+243979234561",
			Email:                 "noella.kalala@email.cd",
			AdresseActuelle:       "Commune de Mont-Ngafula, Quartier Kimwenza",
			VilleActuelle:         "Kinshasa",
			PaysActuel:            "République Démocratique du Congo",
			SituationMatrimoniale: "celibataire",
			NombreEnfants:         1,
			PersonneContact:       "KALALA Emmanuel",
			TelephoneContact:      "+243863214567",
			StatutMigratoire:      "deplace_interne",
			DateEntree:            &[]time.Time{time.Date(2025, 8, 10, 0, 0, 0, 0, time.UTC)}[0],
			PointEntree:           "Fuite depuis Uvira",
			PaysOrigine:           "République Démocratique du Congo",
			CreatedAt:             time.Date(2025, 8, 10, 12, 30, 0, 0, time.UTC),
			UpdatedAt:             time.Date(2025, 8, 10, 12, 30, 0, 0, time.UTC),
		},
		{
			UUID:                  utils.GenerateUUID(),
			NumeroIdentifiant:     "DPI2025011",
			Nom:                   "NZUZI",
			Prenom:                "Cedric",
			DateNaissance:         time.Date(1987, 11, 28, 0, 0, 0, 0, time.UTC),
			LieuNaissance:         "Bukavu",
			Sexe:                  "M",
			Nationalite:           "Congolaise (RDC)",
			TypeDocument:          "carte_identite",
			NumeroDocument:        "CD3691470258",
			DateEmissionDoc:       &[]time.Time{time.Date(2021, 12, 5, 0, 0, 0, 0, time.UTC)}[0],
			DateExpirationDoc:     &[]time.Time{time.Date(2031, 12, 5, 0, 0, 0, 0, time.UTC)}[0],
			AutoriteEmission:      "République Démocratique du Congo",
			Telephone:             "+243986547321",
			Email:                 "cedric.nzuzi@email.cd",
			AdresseActuelle:       "Commune de Makala, Avenue de la Libération",
			VilleActuelle:         "Kinshasa",
			PaysActuel:            "République Démocratique du Congo",
			SituationMatrimoniale: "marie",
			NombreEnfants:         2,
			PersonneContact:       "NZUZI Claudine",
			TelephoneContact:      "+243875432109",
			StatutMigratoire:      "deplace_interne",
			DateEntree:            &[]time.Time{time.Date(2025, 8, 17, 0, 0, 0, 0, time.UTC)}[0],
			PointEntree:           "Évacuation depuis Bukavu",
			PaysOrigine:           "République Démocratique du Congo",
			CreatedAt:             time.Date(2025, 8, 17, 16, 45, 0, 0, time.UTC),
			UpdatedAt:             time.Date(2025, 8, 17, 16, 45, 0, 0, time.UTC),
		},
		{
			UUID:                  utils.GenerateUUID(),
			NumeroIdentifiant:     "DPI2025012",
			Nom:                   "MATONGO",
			Prenom:                "Francine",
			DateNaissance:         time.Date(1992, 1, 19, 0, 0, 0, 0, time.UTC),
			LieuNaissance:         "Mbandaka",
			Sexe:                  "F",
			Nationalite:           "Congolaise (RDC)",
			TypeDocument:          "carte_identite",
			NumeroDocument:        "CD8520741963",
			DateEmissionDoc:       &[]time.Time{time.Date(2022, 10, 7, 0, 0, 0, 0, time.UTC)}[0],
			DateExpirationDoc:     &[]time.Time{time.Date(2032, 10, 7, 0, 0, 0, 0, time.UTC)}[0],
			AutoriteEmission:      "République Démocratique du Congo",
			Telephone:             "+243974185269",
			Email:                 "francine.matongo@email.cd",
			AdresseActuelle:       "Commune de Ngiri-Ngiri, Rue des Palmiers",
			VilleActuelle:         "Kinshasa",
			PaysActuel:            "République Démocratique du Congo",
			SituationMatrimoniale: "marie",
			NombreEnfants:         4,
			PersonneContact:       "MATONGO Pascal",
			TelephoneContact:      "+243891234567",
			StatutMigratoire:      "deplace_interne",
			DateEntree:            &[]time.Time{time.Date(2025, 8, 24, 0, 0, 0, 0, time.UTC)}[0],
			PointEntree:           "Migration interne depuis Mbandaka",
			PaysOrigine:           "République Démocratique du Congo",
			CreatedAt:             time.Date(2025, 8, 24, 11, 20, 0, 0, time.UTC),
			UpdatedAt:             time.Date(2025, 8, 24, 11, 20, 0, 0, time.UTC),
		},
	}

	// Insérer en base
	for _, migrant := range migrants {
		if err := db.Create(&migrant).Error; err != nil {
			log.Printf("Erreur lors de la création du migrant %s: %v", migrant.NumeroIdentifiant, err)
			continue
		}
	}

	log.Printf("✅ %d migrants créés", len(migrants))
	return nil
}

// simulateGeolocalisations crée des géolocalisations simulées
func simulateGeolocalisations(db *gorm.DB) error {
	// Récupérer les migrants existants
	var migrants []models.Migrant
	if err := db.Find(&migrants).Error; err != nil {
		return err
	}

	if len(migrants) == 0 {
		return nil
	}

	geolocalisations := []models.Geolocalisation{
		{
			UUID:             utils.GenerateUUID(),
			MigrantUUID:      migrants[0].UUID,
			Latitude:         -4.3317,
			Longitude:        15.3139,
			TypeLocalisation: "residence_actuelle",
			Description:      "Résidence principale à Kinshasa",
			Adresse:          "Avenue Kasavubu, N°45, Commune de Gombe",
			Ville:            "Kinshasa",
			Pays:             "République Démocratique du Congo",
			TypeMouvement:    "residence_permanente",
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
		{
			UUID:             utils.GenerateUUID(),
			MigrantUUID:      migrants[1].UUID,
			Latitude:         -4.3728,
			Longitude:        15.2663,
			TypeLocalisation: "centre_accueil",
			Description:      "Centre d'accueil pour demandeurs d'asile",
			Adresse:          "Centre CARITAS, Boulevard Triomphal",
			Ville:            "Kinshasa",
			Pays:             "République Démocratique du Congo",
			TypeMouvement:    "residence_temporaire",
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
		// === GÉOLOCALISATIONS POUR DÉPLACÉS INTERNES RDC ===
		{
			UUID:             utils.GenerateUUID(),
			MigrantUUID:      migrants[4].UUID, // Jeanne KABILA
			Latitude:         -1.6792,
			Longitude:        29.2228,
			TypeLocalisation: "centre_accueil",
			Description:      "Camp de déplacés de Mugunga - Nord-Kivu",
			Adresse:          "Camp de déplacés de Mugunga, Route de Sake",
			Ville:            "Goma",
			Pays:             "République Démocratique du Congo",
			TypeMouvement:    "residence_temporaire",
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
		{
			UUID:             utils.GenerateUUID(),
			MigrantUUID:      migrants[4].UUID, // Jeanne KABILA - Lieu d'origine
			Latitude:         -1.1853,
			Longitude:        29.2441,
			TypeLocalisation: "point_passage",
			Description:      "Village d'origine à Rutshuru (avant déplacement)",
			Adresse:          "Village de Kiwanja, Territoire de Rutshuru",
			Ville:            "Rutshuru",
			Pays:             "République Démocratique du Congo",
			TypeMouvement:    "depart",
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
		{
			UUID:             utils.GenerateUUID(),
			MigrantUUID:      migrants[5].UUID, // Jean-Baptiste MBUYI
			Latitude:         -4.3317,
			Longitude:        15.3139,
			TypeLocalisation: "residence_actuelle",
			Description:      "Logement temporaire à Kalamu, Kinshasa",
			Adresse:          "Commune de Kalamu, Avenue de la Paix, N°234",
			Ville:            "Kinshasa",
			Pays:             "République Démocratique du Congo",
			TypeMouvement:    "residence_temporaire",
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
		{
			UUID:             utils.GenerateUUID(),
			MigrantUUID:      migrants[6].UUID, // Espérance NGOY
			Latitude:         1.5593,
			Longitude:        30.0944,
			TypeLocalisation: "centre_accueil",
			Description:      "Site de déplacés de Rhoe, Bunia",
			Adresse:          "Site de déplacés de Rhoe, Commune Kindia",
			Ville:            "Bunia",
			Pays:             "République Démocratique du Congo",
			TypeMouvement:    "residence_temporaire",
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
		{
			UUID:             utils.GenerateUUID(),
			MigrantUUID:      migrants[7].UUID, // Patient KASONGO
			Latitude:         -4.3728,
			Longitude:        15.2663,
			TypeLocalisation: "residence_actuelle",
			Description:      "Hébergement familial à Lemba, Kinshasa",
			Adresse:          "Commune de Lemba, Quartier Righini, Avenue Lukusa",
			Ville:            "Kinshasa",
			Pays:             "République Démocratique du Congo",
			TypeMouvement:    "residence_temporaire",
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
		{
			UUID:             utils.GenerateUUID(),
			MigrantUUID:      migrants[7].UUID, // Patient KASONGO - Lieu d'origine
			Latitude:         0.4951,
			Longitude:        29.4721,
			TypeLocalisation: "point_passage",
			Description:      "Ville d'origine Beni (avant évacuation)",
			Adresse:          "Ville de Beni, Quartier Mulekera",
			Ville:            "Beni",
			Pays:             "République Démocratique du Congo",
			TypeMouvement:    "depart",
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
	}

	// Insérer en base
	for _, geo := range geolocalisations {
		if err := db.Create(&geo).Error; err != nil {
			log.Printf("Erreur lors de la création de la géolocalisation: %v", err)
			continue
		}
	}

	log.Printf("✅ %d géolocalisations créées", len(geolocalisations))
	return nil
}

// simulateMotifDeplacements crée des motifs de déplacement simulés
func simulateMotifDeplacements(db *gorm.DB) error {
	// Récupérer les migrants existants
	var migrants []models.Migrant
	if err := db.Find(&migrants).Error; err != nil {
		return err
	}

	if len(migrants) == 0 {
		return nil
	}

	motifDeplacements := []models.MotifDeplacement{
		{
			UUID:                 utils.GenerateUUID(),
			MigrantUUID:          migrants[0].UUID,
			TypeMotif:            "economique",
			MotifPrincipal:       "Recherche d'opportunités d'emploi mieux rémunérées",
			MotifSecondaire:      "Diversification des activités commerciales",
			Description:          "Commerçant burkinabè cherchant à développer son commerce de produits artisanaux et textiles au Congo.",
			CaractereVolontaire:  true,
			Urgence:              "faible",
			DateDeclenchement:    time.Date(2023, 12, 1, 0, 0, 0, 0, time.UTC),
			DureeEstimee:         365,
			ConflitArme:          false,
			CatastropheNaturelle: false,
			Persecution:          false,
			ViolenceGeneralisee:  false,
			CreatedAt:            time.Now(),
			UpdatedAt:            time.Now(),
		},
		{
			UUID:                 utils.GenerateUUID(),
			MigrantUUID:          migrants[1].UUID,
			TypeMotif:            "politique",
			MotifPrincipal:       "Instabilité politique et menaces sécuritaires au Mali",
			MotifSecondaire:      "Protection de la famille",
			Description:          "Fuit l'instabilité politique au Mali suite aux coups d'État successifs.",
			CaractereVolontaire:  false,
			Urgence:              "elevee",
			DateDeclenchement:    time.Date(2023, 10, 15, 0, 0, 0, 0, time.UTC),
			DureeEstimee:         730,
			ConflitArme:          true,
			CatastropheNaturelle: false,
			Persecution:          true,
			ViolenceGeneralisee:  true,
			CreatedAt:            time.Now(),
			UpdatedAt:            time.Now(),
		},
		// === MOTIFS POUR DÉPLACÉS INTERNES RDC ===
		{
			UUID:                 utils.GenerateUUID(),
			MigrantUUID:          migrants[4].UUID, // Jeanne KABILA
			TypeMotif:            "politique",
			MotifPrincipal:       "Violences intercommunautaires dans le Nord-Kivu",
			MotifSecondaire:      "Protection de la famille et des enfants",
			Description:          "Conflits armés entre groupes rebelles dans la région de Rutshuru. Violences contre les civils, pillages et menaces directes contre la famille.",
			CaractereVolontaire:  false,
			Urgence:              "critique",
			DateDeclenchement:    time.Date(2023, 9, 1, 0, 0, 0, 0, time.UTC),
			DureeEstimee:         1095, // 3 ans
			ConflitArme:          true,
			CatastropheNaturelle: false,
			Persecution:          true,
			ViolenceGeneralisee:  true,
			CreatedAt:            time.Now(),
			UpdatedAt:            time.Now(),
		},
		{
			UUID:                 utils.GenerateUUID(),
			MigrantUUID:          migrants[5].UUID, // Jean-Baptiste MBUYI
			TypeMotif:            "economique",
			MotifPrincipal:       "Effondrement de l'activité minière artisanale",
			MotifSecondaire:      "Recherche d'opportunités d'emploi à Kinshasa",
			Description:          "Fermeture des sites miniers artisanaux dans la région de Kananga due à l'épuisement des ressources et aux conflits. Migration vers Kinshasa pour chercher du travail.",
			CaractereVolontaire:  true,
			Urgence:              "moyenne",
			DateDeclenchement:    time.Date(2024, 1, 10, 0, 0, 0, 0, time.UTC),
			DureeEstimee:         730, // 2 ans
			ConflitArme:          false,
			CatastropheNaturelle: false,
			Persecution:          false,
			ViolenceGeneralisee:  false,
			CreatedAt:            time.Now(),
			UpdatedAt:            time.Now(),
		},
		{
			UUID:                 utils.GenerateUUID(),
			MigrantUUID:          migrants[6].UUID, // Espérance NGOY
			TypeMotif:            "politique",
			MotifPrincipal:       "Violences ethniques dans l'Ituri",
			MotifSecondaire:      "Menaces et intimidations",
			Description:          "Conflits ethniques entre communautés Hema et Lendu dans la région de Djugu. Massacres, destructions de villages et ciblage des jeunes femmes.",
			CaractereVolontaire:  false,
			Urgence:              "critique",
			DateDeclenchement:    time.Date(2023, 11, 28, 0, 0, 0, 0, time.UTC),
			DureeEstimee:         1460, // 4 ans
			ConflitArme:          true,
			CatastropheNaturelle: false,
			Persecution:          true,
			ViolenceGeneralisee:  true,
			CreatedAt:            time.Now(),
			UpdatedAt:            time.Now(),
		},
		{
			UUID:                 utils.GenerateUUID(),
			MigrantUUID:          migrants[7].UUID, // Patient KASONGO
			TypeMotif:            "securite",
			MotifPrincipal:       "Attaques des groupes armés ADF dans la région de Beni",
			MotifSecondaire:      "Protection de la famille nombreuse",
			Description:          "Attaques répétées des Forces Démocratiques Alliées (ADF) dans la région de Beni. Massacres de civils, enlèvements et destructions de biens. Fuite urgente avec toute la famille.",
			CaractereVolontaire:  false,
			Urgence:              "critique",
			DateDeclenchement:    time.Date(2023, 6, 10, 0, 0, 0, 0, time.UTC),
			DureeEstimee:         1825, // 5 ans
			ConflitArme:          true,
			CatastropheNaturelle: false,
			Persecution:          true,
			ViolenceGeneralisee:  true,
			CreatedAt:            time.Now(),
			UpdatedAt:            time.Now(),
		},
	}

	// Insérer en base
	for _, motif := range motifDeplacements {
		if err := db.Create(&motif).Error; err != nil {
			log.Printf("Erreur lors de la création du motif de déplacement: %v", err)
			continue
		}
	}

	log.Printf("✅ %d motifs de déplacement créés", len(motifDeplacements))
	return nil
}

// simulateBiometries crée des données biométriques simulées
func simulateBiometries(db *gorm.DB) error {
	// Récupérer les migrants existants
	var migrants []models.Migrant
	if err := db.Find(&migrants).Error; err != nil {
		return err
	}

	if len(migrants) == 0 {
		return nil
	}

	// Fonction pour générer des données biométriques simulées
	generateBiometricData := func(dataType string, index int) string {
		var data string
		switch dataType {
		case "empreinte_digitale":
			data = fmt.Sprintf("FINGERPRINT_DATA_%d_%d", index, rand.Intn(10000))
		case "reconnaissance_faciale":
			data = fmt.Sprintf("FACIAL_RECOGNITION_DATA_%d", rand.Intn(10000))
		}
		return base64.StdEncoding.EncodeToString([]byte(data))
	}

	var biometries []models.Biometrie

	// Créer des données biométriques pour chaque migrant
	for i, migrant := range migrants {
		// Empreinte digitale
		bio := models.Biometrie{
			UUID:                utils.GenerateUUID(),
			MigrantUUID:         migrant.UUID,
			TypeBiometrie:       "empreinte_digitale",
			IndexDoigt:          &[]int{1}[0],
			QualiteDonnee:       "excellente",
			DonneesBiometriques: generateBiometricData("empreinte_digitale", 1),
			AlgorithmeEncodage:  "SHA-256",
			TailleFichier:       rand.Intn(5000) + 1000,
			DateCapture:         time.Now().Add(-time.Hour * 24 * time.Duration(rand.Intn(60))),
			DisposifCapture:     "Scanner biométrique SecuGen",
			ResolutionCapture:   "500 DPI",
			OperateurCapture:    fmt.Sprintf("Agent DGM00%d", (i%3)+1),
			Verifie:             true,
			ScoreConfiance:      &[]float64{0.95}[0],
			Chiffre:             true,
			CleChiffrement:      fmt.Sprintf("AES256_KEY_%s", utils.GenerateUUID()[:8]),
			CreatedAt:           time.Now(),
			UpdatedAt:           time.Now(),
		}
		bio.DateVerification = &[]time.Time{bio.DateCapture.Add(time.Hour * 2)}[0] 

		biometries = append(biometries, bio)

		// Reconnaissance faciale
		bio2 := models.Biometrie{
			UUID:                utils.GenerateUUID(),
			MigrantUUID:         migrant.UUID,
			TypeBiometrie:       "reconnaissance_faciale",
			QualiteDonnee:       "excellente",
			DonneesBiometriques: generateBiometricData("reconnaissance_faciale", 0),
			AlgorithmeEncodage:  "CNN-DeepFace",
			TailleFichier:       rand.Intn(15000) + 5000,
			DateCapture:         time.Now().Add(-time.Hour * 24 * time.Duration(rand.Intn(30))),
			DisposifCapture:     "Caméra HD avec capteur infrarouge",
			ResolutionCapture:   "1920x1080",
			OperateurCapture:    fmt.Sprintf("Agent DGM00%d", (i%3)+1),
			Verifie:             true,
			ScoreConfiance:      &[]float64{0.92}[0],
			Chiffre:             true,
			CleChiffrement:      fmt.Sprintf("AES256_KEY_%s", utils.GenerateUUID()[:8]),
			CreatedAt:           time.Now(),
			UpdatedAt:           time.Now(),
		}
		bio2.DateVerification = &[]time.Time{bio2.DateCapture.Add(time.Hour * 1)}[0] 

		biometries = append(biometries, bio2)
	}

	// Insérer en base
	for _, bio := range biometries {
		if err := db.Create(&bio).Error; err != nil {
			log.Printf("Erreur lors de la création des données biométriques: %v", err)
			continue
		}
	}

	log.Printf("✅ %d données biométriques créées", len(biometries))
	return nil
}

// simulateAlerts crée des alertes simulées
func simulateAlerts(db *gorm.DB) error {
	// Récupérer les migrants existants
	var migrants []models.Migrant
	if err := db.Find(&migrants).Error; err != nil {
		return err
	}

	if len(migrants) == 0 {
		return nil
	}

	alerts := []models.Alert{
		{
			UUID:                utils.GenerateUUID(),
			MigrantUUID:         migrants[0].UUID,
			TypeAlerte:          "securite",
			NiveauGravite:       "warning",
			Titre:               "Document d'identité expirant bientôt",
			Description:         "Le passeport de M. KEMBO expire dans 45 jours. Il est urgent de procéder au renouvellement.",
			Statut:              "active",
			DateExpiration:      &[]time.Time{time.Now().Add(time.Hour * 24 * 45)}[0],
			ActionRequise:       "Contacter l'ambassade du Burkina Faso pour renouvellement",
			PersonneResponsable: "Agent DGM002", 
			CreatedAt:           time.Now().Add(-time.Hour * 24 * 5),
			UpdatedAt:           time.Now().Add(-time.Hour * 24 * 1),
		},
		{
			UUID:                utils.GenerateUUID(),
			MigrantUUID:         migrants[1].UUID,
			TypeAlerte:          "sante",
			NiveauGravite:       "danger",
			Titre:               "Suivi médical urgent requis",
			Description:         "Mme TRAORE présente des symptômes de stress post-traumatique. Un suivi médical urgent est nécessaire.",
			Statut:              "active",
			DateExpiration:      &[]time.Time{time.Now().Add(time.Hour * 24 * 15)}[0],
			ActionRequise:       "Orientation vers le centre médical MSF",
			PersonneResponsable: "Agent DGM003", 
			CreatedAt:           time.Now().Add(-time.Hour * 24 * 10),
			UpdatedAt:           time.Now().Add(-time.Hour * 24 * 2),
		},
		{
			UUID:                utils.GenerateUUID(),
			MigrantUUID:         migrants[2].UUID,
			TypeAlerte:          "administrative",
			NiveauGravite:       "info",
			Titre:               "Renouvellement de permis de séjour",
			Description:         "Le permis de séjour de M. KONE expire dans 60 jours. Procédure de renouvellement à entamer.",
			Statut:              "active",
			DateExpiration:      &[]time.Time{time.Now().Add(time.Hour * 24 * 60)}[0],
			ActionRequise:       "Accompagner dans les démarches de renouvellement",
			PersonneResponsable: "Agent DGM001", 
			CreatedAt:           time.Now().Add(-time.Hour * 24 * 2),
			UpdatedAt:           time.Now().Add(-time.Hour * 24 * 1),
		},
		// Alertes spécifiques pour les déplacés internes de la RDC
		{
			UUID:                utils.GenerateUUID(),
			MigrantUUID:         migrants[len(migrants)-4].UUID, // Jeanne KABILA (déplacée interne)
			TypeAlerte:          "securite",
			NiveauGravite:       "danger",
			Titre:               "Zone d'origine toujours instable",
			Description:         "La zone de Rutshuru reste instable avec des combats sporadiques. Retour non recommandé pour le moment.",
			Statut:              "active",
			DateExpiration:      &[]time.Time{time.Now().Add(time.Hour * 24 * 30)}[0],
			ActionRequise:       "Maintenir en zone sécurisée, surveiller évolution sécuritaire",
			PersonneResponsable: "Coordinateur Camp Mugunga", 
			CreatedAt:           time.Now().Add(-time.Hour * 24 * 7),
			UpdatedAt:           time.Now().Add(-time.Hour * 24 * 1),
		},
		{
			UUID:                utils.GenerateUUID(),
			MigrantUUID:         migrants[len(migrants)-3].UUID, // Jean-Baptiste MBUYI
			TypeAlerte:          "social",
			NiveauGravite:       "warning",
			Titre:               "Recherche d'opportunités d'emploi",
			Description:         "Déplacé interne cherche formation professionnelle ou opportunité d'emploi pour intégration économique.",
			Statut:              "active",
			DateExpiration:      &[]time.Time{time.Now().Add(time.Hour * 24 * 90)}[0],
			ActionRequise:       "Orientation vers programmes de formation professionnelle",
			PersonneResponsable: "Agent DGM004", 
			CreatedAt:           time.Now().Add(-time.Hour * 24 * 14),
			UpdatedAt:           time.Now().Add(-time.Hour * 24 * 3),
		},
		{
			UUID:                utils.GenerateUUID(),
			MigrantUUID:         migrants[len(migrants)-2].UUID, // Espérance NGOY
			TypeAlerte:          "sante",
			NiveauGravite:       "warning",
			Titre:               "Suivi psychologique traumatisme",
			Description:         "Victime de violences ethniques, nécessite un suivi psychologique régulier pour traiter le traumatisme.",
			Statut:              "active",
			DateExpiration:      &[]time.Time{time.Now().Add(time.Hour * 24 * 180)}[0],
			ActionRequise:       "Sessions thérapeutiques hebdomadaires avec psychologue",
			PersonneResponsable: "Dr. MUKENDI - Centre médical", 
			CreatedAt:           time.Now().Add(-time.Hour * 24 * 21),
			UpdatedAt:           time.Now().Add(-time.Hour * 24 * 5),
		},
		{
			UUID:                utils.GenerateUUID(),
			MigrantUUID:         migrants[len(migrants)-1].UUID, // Patient KASONGO
			TypeAlerte:          "administrative",
			NiveauGravite:       "info",
			Titre:               "Demande de carte d'identité nationale",
			Description:         "Documents d'identité perdus lors de la fuite. Procédure de renouvellement de carte d'identité en cours.",
			Statut:              "active",
			DateExpiration:      &[]time.Time{time.Now().Add(time.Hour * 24 * 60)}[0],
			ActionRequise:       "Accompagner aux services de l'état civil pour reconstitution dossier",
			PersonneResponsable: "Agent DGM005", 
			CreatedAt:           time.Now().Add(-time.Hour * 24 * 12),
			UpdatedAt:           time.Now().Add(-time.Hour * 24 * 2),
		},
		{
			UUID:                utils.GenerateUUID(),
			MigrantUUID:         migrants[len(migrants)-4].UUID, // Alerte supplémentaire pour Jeanne KABILA
			TypeAlerte:          "social",
			NiveauGravite:       "info",
			Titre:               "Recherche de membres de famille",
			Description:         "Recherche active de membres de famille séparés lors du déplacement forcé depuis Rutshuru.",
			Statut:              "active",
			DateExpiration:      &[]time.Time{time.Now().Add(time.Hour * 24 * 120)}[0],
			ActionRequise:       "Inscription au programme de recherche familiale de la Croix-Rouge",
			PersonneResponsable: "CICR Goma", 
			CreatedAt:           time.Now().Add(-time.Hour * 24 * 18),
			UpdatedAt:           time.Now().Add(-time.Hour * 24 * 4),
		},
	}

	// Insérer en base
	for _, alert := range alerts {
		if err := db.Create(&alert).Error; err != nil {
			log.Printf("Erreur lors de la création de l'alerte: %v", err)
			continue
		}
	}

	log.Printf("✅ %d alertes créées", len(alerts))
	return nil
}
