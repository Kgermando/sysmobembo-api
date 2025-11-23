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

	// 4. Créer les géolocalisations (dépendent des migrants)
	log.Println("4. Création des géolocalisations...")
	if err := simulateGeolocalisations(db); err != nil {
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

// simulateIdentites crée les identités et retourne un map[NumeroIdentifiant]IdentiteUUID
func simulateIdentites(db *gorm.DB) (map[string]string, error) {
	identiteMap := make(map[string]string)

	identites := []struct {
		NumeroIdentifiant string
		Identite          models.Identite
	}{
		// === MIGRANTS INTERNATIONAUX ===
		{
			NumeroIdentifiant: "MIG2025001",
			Identite: models.Identite{
				UUID:             utils.GenerateUUID(),
				Nom:              "OUEDRAOGO",
				Prenom:           "Amadou",
				DateNaissance:    time.Date(1990, 5, 15, 0, 0, 0, 0, time.UTC),
				LieuNaissance:    "Ouagadougou",
				Sexe:             "M",
				Nationalite:      "Burkinabè",
				Adresse:          "Avenue Kasavubu, N°45, Kinshasa",
				Profession:       "Commerçant",
				PaysEmetteur:     "Burkina Faso",
				AutoriteEmetteur: "République du Burkina Faso",
				NumeroPasseport:  "BF1234567",
				CreatedAt:        time.Date(2025, 6, 5, 10, 30, 0, 0, time.UTC),
				UpdatedAt:        time.Date(2025, 6, 5, 10, 30, 0, 0, time.UTC),
			},
		},
		{
			NumeroIdentifiant: "MIG2025002",
			Identite: models.Identite{
				UUID:             utils.GenerateUUID(),
				Nom:              "SANKARA",
				Prenom:           "Fanta",
				DateNaissance:    time.Date(1987, 9, 12, 0, 0, 0, 0, time.UTC),
				LieuNaissance:    "Bobo-Dioulasso",
				Sexe:             "F",
				Nationalite:      "Burkinabè",
				Adresse:          "Avenue de la Justice, N°78, Kinshasa",
				Profession:       "Infirmière",
				PaysEmetteur:     "Burkina Faso",
				AutoriteEmetteur: "République du Burkina Faso",
				NumeroPasseport:  "BF9876543",
				CreatedAt:        time.Date(2025, 6, 12, 14, 15, 0, 0, time.UTC),
				UpdatedAt:        time.Date(2025, 6, 12, 14, 15, 0, 0, time.UTC),
			},
		},
		{
			NumeroIdentifiant: "MIG2025003",
			Identite: models.Identite{
				UUID:             utils.GenerateUUID(),
				Nom:              "ZONGO",
				Prenom:           "Rasmané",
				DateNaissance:    time.Date(1993, 2, 28, 0, 0, 0, 0, time.UTC),
				LieuNaissance:    "Ouahigouya",
				Sexe:             "M",
				Nationalite:      "Burkinabè",
				Adresse:          "Commune de Bandalungwa, Rue 15",
				Profession:       "Agriculteur",
				PaysEmetteur:     "Burkina Faso",
				AutoriteEmetteur: "République du Burkina Faso",
				NumeroPasseport:  "BF5432109",
				CreatedAt:        time.Date(2025, 6, 18, 9, 45, 0, 0, time.UTC),
				UpdatedAt:        time.Date(2025, 6, 18, 9, 45, 0, 0, time.UTC),
			},
		},
		{
			NumeroIdentifiant: "MIG2025004",
			Identite: models.Identite{
				UUID:             utils.GenerateUUID(),
				Nom:              "TRAORE",
				Prenom:           "Aïssata",
				DateNaissance:    time.Date(1985, 8, 22, 0, 0, 0, 0, time.UTC),
				LieuNaissance:    "Bamako",
				Sexe:             "F",
				Nationalite:      "Malienne",
				Adresse:          "Boulevard Triomphal, N°78",
				Profession:       "Enseignante",
				PaysEmetteur:     "Mali",
				AutoriteEmetteur: "République du Mali",
				NumeroPasseport:  "ML9876543",
				CreatedAt:        time.Date(2025, 6, 25, 16, 20, 0, 0, time.UTC),
				UpdatedAt:        time.Date(2025, 6, 25, 16, 20, 0, 0, time.UTC),
			},
		},
		{
			NumeroIdentifiant: "MIG2025005",
			Identite: models.Identite{
				UUID:             utils.GenerateUUID(),
				Nom:              "KEITA",
				Prenom:           "Moussa",
				DateNaissance:    time.Date(1991, 11, 7, 0, 0, 0, 0, time.UTC),
				LieuNaissance:    "Kayes",
				Sexe:             "M",
				Nationalite:      "Malienne",
				Adresse:          "Quartier Matete, Avenue des Usines",
				Profession:       "Mécanicien",
				PaysEmetteur:     "Mali",
				AutoriteEmetteur: "République du Mali",
				NumeroPasseport:  "ML7654321",
				CreatedAt:        time.Date(2025, 7, 2, 9, 15, 0, 0, time.UTC),
				UpdatedAt:        time.Date(2025, 7, 2, 9, 15, 0, 0, time.UTC),
			},
		},
		// === DÉPLACÉS INTERNES RDC ===
		{
			NumeroIdentifiant: "DPI2025001",
			Identite: models.Identite{
				UUID:             utils.GenerateUUID(),
				Nom:              "KABILA",
				Prenom:           "Jean-Pierre",
				DateNaissance:    time.Date(1982, 4, 15, 0, 0, 0, 0, time.UTC),
				LieuNaissance:    "Goma",
				Sexe:             "M",
				Nationalite:      "Congolaise (RDC)",
				Adresse:          "Camp de déplacés, Goma",
				Profession:       "Cultivateur",
				PaysEmetteur:     "République Démocratique du Congo",
				AutoriteEmetteur: "République Démocratique du Congo",
				NumeroPasseport:  "CD1234567890",
				CreatedAt:        time.Date(2025, 6, 8, 14, 0, 0, 0, time.UTC),
				UpdatedAt:        time.Date(2025, 6, 8, 14, 0, 0, 0, time.UTC),
			},
		},
		{
			NumeroIdentifiant: "DPI2025002",
			Identite: models.Identite{
				UUID:             utils.GenerateUUID(),
				Nom:              "MULAMBA",
				Prenom:           "Grace",
				DateNaissance:    time.Date(1990, 8, 23, 0, 0, 0, 0, time.UTC),
				LieuNaissance:    "Butembo",
				Sexe:             "F",
				Nationalite:      "Congolaise (RDC)",
				Adresse:          "Site de déplacement, Bunia",
				Profession:       "Commerçante",
				PaysEmetteur:     "République Démocratique du Congo",
				AutoriteEmetteur: "République Démocratique du Congo",
				NumeroPasseport:  "CD9876543210",
				CreatedAt:        time.Date(2025, 6, 15, 11, 30, 0, 0, time.UTC),
				UpdatedAt:        time.Date(2025, 6, 15, 11, 30, 0, 0, time.UTC),
			},
		},
	}

	// Créer les identités en base
	for _, item := range identites {
		if err := db.Create(&item.Identite).Error; err != nil {
			log.Printf("Erreur lors de la création de l'identité %s: %v", item.NumeroIdentifiant, err)
			continue
		}
		identiteMap[item.NumeroIdentifiant] = item.Identite.UUID
	}

	log.Printf("✅ %d identités créées", len(identiteMap))
	return identiteMap, nil
}

// createMigrantWithIdentite crée une identité et un migrant associé
func createMigrantWithIdentite(
	db *gorm.DB, identiteData models.Identite, migrantData models.Migrant) error {
	// Créer l'identité
	identiteData.UUID = utils.GenerateUUID()
	identiteData.CreatedAt = migrantData.CreatedAt
	identiteData.UpdatedAt = migrantData.UpdatedAt

	if err := db.Create(&identiteData).Error; err != nil {
		return fmt.Errorf("erreur création identité: %v", err)
	}

	// Créer le migrant avec la référence à l'identité
	migrantData.IdentiteUUID = identiteData.UUID
	if err := db.Create(&migrantData).Error; err != nil {
		return fmt.Errorf("erreur création migrant: %v", err)
	}

	return nil
}

// simulateMigrants crée des migrants simulés et les associe aux identités
func simulateMigrants(db *gorm.DB, identiteMap map[string]string) error {
	migrants := []models.Migrant{
		// === MIGRANTS INTERNATIONAUX - JUIN 2025 ===
		{
			UUID:                  utils.GenerateUUID(),
			NumeroIdentifiant:     "MIG2025001",
			IdentiteUUID:          identiteMap["MIG2025001"],
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
			PaysDestination:       "République Démocratique du Congo", 
			CreatedAt:             time.Date(2025, 6, 5, 10, 30, 0, 0, time.UTC),
			UpdatedAt:             time.Date(2025, 6, 5, 10, 30, 0, 0, time.UTC),
		},
		{
			UUID:                  utils.GenerateUUID(),
			NumeroIdentifiant:     "MIG2025002",
			IdentiteUUID:          identiteMap["MIG2025002"],
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
			PaysDestination:       "République Démocratique du Congo", 
			CreatedAt:             time.Date(2025, 6, 12, 14, 15, 0, 0, time.UTC),
			UpdatedAt:             time.Date(2025, 6, 12, 14, 15, 0, 0, time.UTC),
		},
		{
			UUID:                  utils.GenerateUUID(),
			NumeroIdentifiant:     "MIG2025003",
			IdentiteUUID:          identiteMap["MIG2025003"],
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
			PaysDestination:       "République Démocratique du Congo",
			CreatedAt:             time.Date(2025, 6, 18, 9, 45, 0, 0, time.UTC),
			UpdatedAt:             time.Date(2025, 6, 18, 9, 45, 0, 0, time.UTC),
		},
		{
			UUID:                  utils.GenerateUUID(),
			NumeroIdentifiant:     "MIG2025004",
			IdentiteUUID:          identiteMap["MIG2025004"],
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
			PaysDestination:       "République Démocratique du Congo", 
			CreatedAt:             time.Date(2025, 6, 25, 16, 20, 0, 0, time.UTC),
			UpdatedAt:             time.Date(2025, 6, 25, 16, 20, 0, 0, time.UTC),
		},
		{
			UUID:                  utils.GenerateUUID(),
			NumeroIdentifiant:     "MIG2025005",
			IdentiteUUID:          identiteMap["MIG2025005"],
			Telephone:             "+22376543210",
			Email:                 "moussa.keita@email.com",
			AdresseActuelle:       "Commune de Barumbu, Rue 24",
			VilleActuelle:         "Kinshasa",
			PaysActuel:            "République Démocratique du Congo",
			SituationMatrimoniale: "marie",
			NombreEnfants:         2,
			StatutMigratoire:      "regulier",
			DateEntree:            &[]time.Time{time.Date(2025, 7, 2, 0, 0, 0, 0, time.UTC)}[0],
			PointEntree:           "Aéroport de N'djili",
			PaysDestination:       "République Démocratique du Congo", 
			CreatedAt:             time.Date(2025, 7, 2, 9, 15, 0, 0, time.UTC),
			UpdatedAt:             time.Date(2025, 7, 2, 9, 15, 0, 0, time.UTC),
		},

		// === DÉPLACÉS INTERNES RDC ===
		{
			UUID:                  utils.GenerateUUID(),
			NumeroIdentifiant:     "DPI2025001",
			IdentiteUUID:          identiteMap["DPI2025001"],
			Telephone:             "+243998765432",
			Email:                 "jp.kabila@email.cd",
			AdresseActuelle:       "Commune de Kimbanseke, Avenue des Poids Lourds",
			VilleActuelle:         "Kinshasa",
			PaysActuel:            "République Démocratique du Congo",
			SituationMatrimoniale: "marie",
			NombreEnfants:         5,
			PersonneContact:       "KABILA Marie",
			TelephoneContact:      "+243812345678",
			StatutMigratoire:      "deplace_interne",
			DateEntree:            &[]time.Time{time.Date(2025, 6, 8, 0, 0, 0, 0, time.UTC)}[0],
			PointEntree:           "Déplacement depuis Goma suite aux conflits",
			PaysDestination:       "République Démocratique du Congo", 
			CreatedAt:             time.Date(2025, 6, 8, 14, 0, 0, 0, time.UTC),
			UpdatedAt:             time.Date(2025, 6, 8, 14, 0, 0, 0, time.UTC),
		},
		{
			UUID:                  utils.GenerateUUID(),
			NumeroIdentifiant:     "DPI2025002",
			IdentiteUUID:          identiteMap["DPI2025002"],
			Telephone:             "+243823456789",
			Email:                 "grace.mulamba@email.cd",
			AdresseActuelle:       "Commune de N'sele, Camp de déplacés UNHCR",
			VilleActuelle:         "Kinshasa",
			PaysActuel:            "République Démocratique du Congo",
			SituationMatrimoniale: "veuf",
			NombreEnfants:         3,
			PersonneContact:       "MULAMBA Joseph",
			TelephoneContact:      "+243897654321",
			StatutMigratoire:      "deplace_interne",
			DateEntree:            &[]time.Time{time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC)}[0],
			PointEntree:           "Fuite de Butembo suite aux violences",
			PaysDestination:       "République Démocratique du Congo", 
			CreatedAt:             time.Date(2025, 6, 15, 11, 30, 0, 0, time.UTC),
			UpdatedAt:             time.Date(2025, 6, 15, 11, 30, 0, 0, time.UTC),
		},
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
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
		{
			UUID:             utils.GenerateUUID(),
			MigrantUUID:      migrants[1].UUID,
			Latitude:         -4.3728,
			Longitude:        15.2663,  
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
		// === GÉOLOCALISATIONS POUR DÉPLACÉS INTERNES RDC ===
		{
			UUID:             utils.GenerateUUID(),
			MigrantUUID:      migrants[4].UUID, // Jeanne KABILA
			Latitude:         -1.6792,
			Longitude:        29.2228,  
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
		{
			UUID:             utils.GenerateUUID(),
			MigrantUUID:      migrants[4].UUID, // Jeanne KABILA - Lieu d'origine
			Latitude:         -1.1853,
			Longitude:        29.2441,  
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
		{
			UUID:             utils.GenerateUUID(),
			MigrantUUID:      migrants[5].UUID, // Jean-Baptiste MBUYI
			Latitude:         -4.3317,
			Longitude:        15.3139,  
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
		{
			UUID:             utils.GenerateUUID(),
			MigrantUUID:      migrants[6].UUID, // Espérance NGOY
			Latitude:         1.5593,
			Longitude:        30.0944,  
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
		{
			UUID:             utils.GenerateUUID(),
			MigrantUUID:      migrants[7].UUID, // Patient KASONGO
			Latitude:         -4.3728,
			Longitude:        15.2663,  
			CreatedAt:        time.Now(),
			UpdatedAt:        time.Now(),
		},
		{
			UUID:             utils.GenerateUUID(),
			MigrantUUID:      migrants[7].UUID, // Patient KASONGO - Lieu d'origine
			Latitude:         0.4951,
			Longitude:        29.4721,  
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
