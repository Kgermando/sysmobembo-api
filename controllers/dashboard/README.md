# 📊 Dashboard API - Système de Monitoring des Migrants

Ce module fournit des analyses avancées et un système de monitoring en temps réel pour le suivi des migrants.

## 🗂️ Structure des Dashboards

### 1. 🔮 **Analyse Prédictive** (`predictive_analysis.controller.go`)
Analyse les tendances et prédit les flux migratoires futurs.

**Endpoints:**
- `GET /api/dashboard/predictive/migration-flow` - Prédiction des flux migratoires
- `GET /api/dashboard/predictive/risk-analysis` - Analyse des risques prédictifs  
- `GET /api/dashboard/predictive/demographic-prediction` - Prédiction démographique
- `GET /api/dashboard/predictive/movement-patterns` - Patterns de mouvement prédictifs

**Paramètres:**
- `periode_days` : Période d'analyse en jours (défaut: 30)
- `pays_origine` : Filtrer par pays d'origine
- `pays_destination` : Filtrer par pays de destination

### 2. ⏱️ **Suivi Temps Réel** (`realtime_monitoring.controller.go`)
Monitoring en temps réel des activités et alertes.

**Endpoints:**
- `GET /api/dashboard/realtime/dashboard` - Vue d'ensemble temps réel
- `GET /api/dashboard/realtime/alerts` - Monitoring des alertes actives
- `GET /api/dashboard/realtime/movements` - Suivi des mouvements récents
- `GET /api/dashboard/realtime/status` - Statuts et évolutions
- `GET /api/dashboard/realtime/updates` - Dernières mises à jour

### 3. 🛤️ **Analyse de Trajectoire** (`trajectory_analysis.controller.go`)
Analyse détaillée des trajectoires et détection d'anomalies.

**Endpoints:**
- `GET /api/dashboard/trajectory/individual?migrant_uuid={uuid}` - Trajectoire individuelle
- `GET /api/dashboard/trajectory/group` - Trajectoires groupées
- `GET /api/dashboard/trajectory/patterns` - Patterns de mouvement
- `GET /api/dashboard/trajectory/anomalies` - Détection d'anomalies

**Paramètres:**
- `migrant_uuid` : UUID du migrant (requis pour individual)
- `days` : Période d'analyse (défaut: 30)
- `pays_origine` : Filtrer par pays d'origine
- `pays_destination` : Filtrer par pays de destination
- `limit` : Limite de résultats (défaut: 20)

### 4. 🗺️ **Analyse Spatiale** (`spatial_analysis.controller.go`)
Analyse géographique avancée avec clustering et proximité.

**Endpoints:**
- `GET /api/dashboard/spatial/density` - Analyse de densité spatiale
- `GET /api/dashboard/spatial/corridors` - Corridors de migration
- `GET /api/dashboard/spatial/proximity` - Analyse de proximité
- `GET /api/dashboard/spatial/areas-of-interest` - Zones d'intérêt

**Paramètres pour proximité:**
- `latitude` : Latitude du point central (requis)
- `longitude` : Longitude du point central (requis) 
- `radius` : Rayon de recherche en km (défaut: 50)
- `days` : Période d'analyse (défaut: 30)

**Paramètres pour densité:**
- `radius` : Rayon de clustering en km (défaut: 10)
- `min_migrants` : Nombre minimum de migrants (défaut: 3)

### 5. 🌍 **Système d'Information Géographique** (`gis_system.controller.go`)
SIG complet avec c
uches cartographiques et export de données.

**Endpoints:**
- `GET /api/dashboard/gis/config` - Configuration de la carte
- `GET /api/dashboard/gis/layers/migrants` - Couche des migrants
- `GET /api/dashboard/gis/layers/routes` - Couche des routes
- `GET /api/dashboard/gis/layers/alerts` - Couche des alertes
- `GET /api/dashboard/gis/layers/infrastructure` - Couche infrastructure
- `GET /api/dashboard/gis/layers/heatmap` - Couche heatmap
- `GET /api/dashboard/gis/export` - Export des données

**Paramètres pour infrastructure:**
- `type` : Type d'infrastructure ("border_points", "reception_centers", "all")

**Paramètres pour heatmap:**
- `intensity` : Type d'intensité ("migrant_count", "alert_count", "movement_count")
- `days` : Période d'analyse (défaut: 30)

## 🚀 **Intégration dans l'API principale**

Pour intégrer ces dashboards dans votre API principale, ajoutez dans votre fichier de routes principal :

```go
package routes

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kgermando/sysmobembo-api/controllers/dashboard"
	// ... autres imports
)

func SetupRoutes(app *fiber.App) {
	// ... autres routes
	
	// Intégrer les routes dashboard
	dashboard.SetupDashboardRoutes(app)
}
```

## 📊 **Formats de Données**

### **Réponses Standard**
```json
{
  "status": "success",
  "data": {
    // Données spécifiques à l'endpoint
  },
  "timestamp": "2025-09-07T00:00:00Z"
}
```

### **Données GeoJSON (SIG)**
```json
{
  "type": "FeatureCollection",
  "features": [
    {
      "id": "unique_id",
      "type": "Feature", 
      "geometry": {
        "type": "Point",
        "coordinates": [longitude, latitude]
      },
      "properties": {
        "nom": "Propriétés spécifiques",
        "popup_content": "Contenu du popup"
      }
    }
  ]
}
```

### **Données Heatmap**
```json
[
  {
    "latitude": -4.038333,
    "longitude": 21.758664,
    "weight": 0.85,
    "intensity": 0.75
  }
]
```

## 🔧 **Fonctionnalités Avancées**

### **Calculs Géographiques**
- Distance Haversine entre coordonnées GPS
- Calculs de vitesse et trajectoires  
- Clustering spatial avec densité
- Zones de proximité géographique

### **Analyses Temporelles**
- Prédictions basées sur l'historique
- Détection d'anomalies temporelles
- Patterns saisonniers et hebdomadaires
- Suivi en temps réel

### **Détection d'Anomalies**
- Vitesses de déplacement anormales
- Trajets inhabituels ou suspects
- Concentrations anormales de migrants
- Alertes géographiques automatiques

## 🎯 **Cas d'Usage**

1. **Centre de Contrôle** : Monitoring temps réel des flux migratoires
2. **Analyse Stratégique** : Prédictions et planification des ressources
3. **Sécurité Frontalière** : Détection d'anomalies et zones à risque
4. **Aide Humanitaire** : Identification des zones nécessitant une intervention
5. **Cartographie** : Visualisation complète sur SIG avec couches multiples

## ⚡ **Performance**

- Requêtes SQL optimisées avec agrégations
- Pagination automatique pour les gros volumes
- Cache possible sur les données prédictives  
- Format GeoJSON standard pour l'interopérabilité
- Export multi-format (GeoJSON, KML, Shapefile)

## 🔐 **Sécurité**

Les endpoints peuvent être sécurisés avec les middlewares d'authentification existants :

```go
// Exemple d'ajout d'authentification
protectedGroup := dashboardGroup.Group("/", middleware.AuthRequired())
protectedGroup.Get("/sensitive-data", GetSensitiveData)
```
