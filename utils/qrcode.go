package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/skip2/go-qrcode"
)

// QRCodeData représente les données à encoder dans le QR code
type QRCodeData struct {
	UUID         string    `json:"uuid"`
	Matricule    string    `json:"matricule"`
	Nom          string    `json:"nom"`
	PostNom      string    `json:"postnom"`
	Prenom       string    `json:"prenom"`
	Grade        string    `json:"grade"`
	Fonction     string    `json:"fonction"`
	Service      string    `json:"service"`
	Direction    string    `json:"direction"`
	Ministere    string    `json:"ministere"`
	DateEmission time.Time `json:"date_emission"`
	ValidUntil   time.Time `json:"valid_until"`
}

// GenerateQRCode génère un QR code pour un agent public
func GenerateQRCode(data QRCodeData, outputDir string) (string, string, error) {
	// Créer le répertoire s'il n'existe pas
	err := os.MkdirAll(outputDir, 0755)
	if err != nil {
		return "", "", fmt.Errorf("erreur lors de la création du répertoire: %v", err)
	}

	// Encoder les données en JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", "", fmt.Errorf("erreur lors de l'encodage JSON: %v", err)
	}

	// Nom du fichier QR code
	filename := fmt.Sprintf("qr_%s_%d.png", data.Matricule, time.Now().Unix())
	filepath := filepath.Join(outputDir, filename)

	// Générer le QR code
	err = qrcode.WriteFile(string(jsonData), qrcode.Medium, 256, filepath)
	if err != nil {
		return "", "", fmt.Errorf("erreur lors de la génération du QR code: %v", err)
	}

	return filepath, string(jsonData), nil
}

// ValidateQRCode valide et décode un QR code
func ValidateQRCode(qrCodeData string) (*QRCodeData, error) {
	var data QRCodeData
	err := json.Unmarshal([]byte(qrCodeData), &data)
	if err != nil {
		return nil, fmt.Errorf("erreur lors du décodage du QR code: %v", err)
	}

	// Vérifier si le QR code est encore valide
	if time.Now().After(data.ValidUntil) {
		return nil, fmt.Errorf("QR code expiré")
	}

	return &data, nil
}

// UpdateQRCodeValidity met à jour la période de validité du QR code
func UpdateQRCodeValidity(qrCodeData string, newValidUntil time.Time) (string, error) {
	var data QRCodeData
	err := json.Unmarshal([]byte(qrCodeData), &data)
	if err != nil {
		return "", fmt.Errorf("erreur lors du décodage du QR code: %v", err)
	}

	data.ValidUntil = newValidUntil

	updatedJson, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("erreur lors de l'encodage JSON: %v", err)
	}

	return string(updatedJson), nil
}

// GenerateQRCodeURL génère une URL pour accéder aux informations de l'agent
func GenerateQRCodeURL(baseURL, uuid string) string {
	return fmt.Sprintf("%s/api/agents/verify/%s", baseURL, uuid)
}
