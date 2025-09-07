# Base de Données SysMobembo

## Migration Automatique avec Données Simulées

La fonction `Connect()` dans `connection.go` effectue automatiquement :

1. **Migration des modèles** - Création/mise à jour des tables
2. **Initialisation des données simulées** - Si la base est vide

### 📊 Données Créées Automatiquement

Au premier démarrage, le système crée automatiquement :

#### 👥 Utilisateurs (3) :
- **Jean-Claude MUKENDI** - Directeur (Administrator) - `jean.mukendi@dgm.cd`
- **Marie-Claire KASONGO** - Agent (Manager) - `marie.kasongo@dgm.cd`  
- **Joseph TSHISEKEDI** - Superviseur (Supervisor) - `joseph.tshisekedi@dgm.cd`

*Mot de passe pour tous : `password123`*

#### 🌍 Migrants (4) :
- **Amadou OUEDRAOGO** (Burkina Faso) - Commerçant régulier
- **Aïssata TRAORE** (Mali) - Demandeur d'asile
- **Ibrahim KONE** (Côte d'Ivoire) - Entrepreneur
- **Fatima DIALLO** (Guinée) - Réfugiée

#### 📍 Géolocalisations (2) :
- Résidence principale à Gombe
- Centre d'accueil CARITAS

#### 🎯 Motifs de Déplacement (2) :
- Économique (recherche d'emploi)
- Politique (instabilité au Mali)

#### 🔒 Données Biométriques (8) :
- Empreintes digitales et reconnaissance faciale pour chaque migrant

#### 🚨 Alertes (3) :
- Document expirant
- Suivi médical urgent  
- Renouvellement administratif

### 🔧 Configuration

La création des données se fait automatiquement lors de la migration si :
- Aucun utilisateur n'existe
- Aucun migrant n'existe

### 🎯 Utilisation

1. **Premier démarrage** : Les données sont créées automatiquement
2. **Démarrages suivants** : Aucune donnée n'est ajoutée (protection contre les doublons)

### 🛡️ Sécurité

- Mots de passe hashés avec bcrypt
- Données biométriques encodées en base64
- Clés de chiffrement générées aléatoirement
- Aucune donnée personnelle réelle

Cette approche garantit que votre application a toujours des données de démonstration cohérentes et réalistes pour vos présentations, sans intervention manuelle.
