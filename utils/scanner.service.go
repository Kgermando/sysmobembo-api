package utils

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

// ScannerService gère le scan de documents via TWAIN/WIA (Windows) ou SANE (Linux)
type ScannerService struct {
	OutputDir string
}

// NewScannerService crée une nouvelle instance du service de scan
func NewScannerService(outputDir string) *ScannerService {
	// Créer le dossier de sortie s'il n'existe pas
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Printf("Erreur lors de la création du répertoire de sortie: %v\n", err)
	}
	return &ScannerService{
		OutputDir: outputDir,
	}
}

// ScanDocument déclenche un scan et retourne le chemin du fichier scanné
func (s *ScannerService) ScanDocument() (string, error) {
	timestamp := time.Now().Format("20060102_150405")
	outputFileName := fmt.Sprintf("scan_%s.jpg", timestamp)
	outputPath := filepath.Join(s.OutputDir, outputFileName)

	var err error
	switch runtime.GOOS {
	case "windows":
		err = s.scanWindows(outputPath)
	case "linux":
		err = s.scanLinux(outputPath)
	default:
		return "", fmt.Errorf("système d'exploitation non supporté: %s", runtime.GOOS)
	}

	if err != nil {
		return "", err
	}

	// Vérifier que le fichier a été créé
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		return "", fmt.Errorf("le fichier scanné n'a pas été créé")
	}

	return outputPath, nil
}

// scanWindows utilise WIA (Windows Image Acquisition) pour scanner
func (s *ScannerService) scanWindows(outputPath string) error {
	// Script PowerShell pour utiliser WIA (Windows Image Acquisition)
	psScript := fmt.Sprintf(`
$deviceManager = New-Object -ComObject WIA.DeviceManager
$device = $deviceManager.DeviceInfos.Item(1)

if ($device -eq $null) {
    Write-Error "Aucun scanner détecté"
    exit 1
}

$scanner = $device.Connect()
$item = $scanner.Items.Item(1)

# Configuration du scan
$item.Properties("6146").Value = 300  # Résolution DPI (haute qualité pour OCR)
$item.Properties("6147").Value = 300
$item.Properties("6148").Value = 0    # Format: Couleur (0=Couleur, 1=Gris, 2=N&B)

# Exécuter le scan
$image = $item.Transfer("{B96B3CAE-0728-11D3-9D7B-0000F81EF32E}")  # Format JPEG

# Sauvegarder l'image
if ($image -ne $null) {
    $image.SaveFile("%s")
    Write-Output "Scan terminé avec succès"
} else {
    Write-Error "Erreur lors du scan"
    exit 1
}
`, outputPath)

	// Créer un fichier temporaire pour le script PowerShell
	tmpFile, err := os.CreateTemp("", "scan_*.ps1")
	if err != nil {
		return fmt.Errorf("erreur lors de la création du fichier temporaire: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(psScript); err != nil {
		return fmt.Errorf("erreur lors de l'écriture du script: %v", err)
	}
	tmpFile.Close()

	// Exécuter le script PowerShell
	cmd := exec.Command("powershell.exe", "-ExecutionPolicy", "Bypass", "-File", tmpFile.Name())
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("erreur lors de l'exécution du scan: %v\nSortie: %s", err, string(output))
	}

	return nil
}

// scanLinux utilise SANE pour scanner sous Linux
func (s *ScannerService) scanLinux(outputPath string) error {
	// Vérifier que scanimage (SANE) est installé
	if _, err := exec.LookPath("scanimage"); err != nil {
		return fmt.Errorf("scanimage (SANE) n'est pas installé. Installez-le avec: sudo apt-get install sane-utils")
	}

	// Commande scanimage avec paramètres optimisés pour OCR
	cmd := exec.Command("scanimage",
		"--format=jpeg",
		"--resolution=300", // 300 DPI pour une bonne qualité OCR
		"--mode=Color",
		fmt.Sprintf("--output=%s", outputPath),
	)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("erreur lors de l'exécution du scan: %v\nSortie: %s", err, string(output))
	}

	return nil
}

// ScanToPDF scanne un document et le sauvegarde en PDF
func (s *ScannerService) ScanToPDF() (string, error) {
	timestamp := time.Now().Format("20060102_150405")
	outputFileName := fmt.Sprintf("scan_%s.pdf", timestamp)
	outputPath := filepath.Join(s.OutputDir, outputFileName)

	var err error
	switch runtime.GOOS {
	case "windows":
		err = s.scanWindowsPDF(outputPath)
	case "linux":
		err = s.scanLinuxPDF(outputPath)
	default:
		return "", fmt.Errorf("système d'exploitation non supporté: %s", runtime.GOOS)
	}

	if err != nil {
		return "", err
	}

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		return "", fmt.Errorf("le fichier PDF n'a pas été créé")
	}

	return outputPath, nil
}

// scanWindowsPDF scanne en PDF sous Windows
func (s *ScannerService) scanWindowsPDF(outputPath string) error {
	// Scanner d'abord en image puis convertir en PDF
	jpgPath := outputPath[:len(outputPath)-4] + ".jpg"
	if err := s.scanWindows(jpgPath); err != nil {
		return err
	}
	defer os.Remove(jpgPath)

	// Ici, vous pourriez utiliser une bibliothèque pour convertir JPG en PDF
	// Pour l'instant, on renomme simplement (à améliorer avec une vraie conversion)
	return os.Rename(jpgPath, outputPath)
}

// scanLinuxPDF scanne directement en PDF sous Linux
func (s *ScannerService) scanLinuxPDF(outputPath string) error {
	// Avec SANE, il faut d'abord scanner en image puis convertir
	// On peut utiliser convert (ImageMagick) pour la conversion
	jpgPath := outputPath[:len(outputPath)-4] + ".jpg"

	if err := s.scanLinux(jpgPath); err != nil {
		return err
	}
	defer os.Remove(jpgPath)

	// Convertir en PDF avec ImageMagick
	cmd := exec.Command("convert", jpgPath, outputPath)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("erreur lors de la conversion en PDF: %v\nSortie: %s", err, string(output))
	}

	return nil
}

// ListScanners retourne la liste des scanners disponibles
func (s *ScannerService) ListScanners() ([]string, error) {
	var scanners []string
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		psScript := `
$deviceManager = New-Object -ComObject WIA.DeviceManager
$devices = $deviceManager.DeviceInfos
foreach ($device in $devices) {
    if ($device.Type -eq 1) {  # Type 1 = Scanner
        Write-Output $device.Properties("Name").Value
    }
}
`
		tmpFile, err := os.CreateTemp("", "list_scanners_*.ps1")
		if err != nil {
			return nil, err
		}
		defer os.Remove(tmpFile.Name())

		tmpFile.WriteString(psScript)
		tmpFile.Close()

		cmd = exec.Command("powershell.exe", "-ExecutionPolicy", "Bypass", "-File", tmpFile.Name())

	case "linux":
		// Utiliser scanimage -L pour lister les scanners
		cmd = exec.Command("scanimage", "-L")

	default:
		return nil, fmt.Errorf("système d'exploitation non supporté: %s", runtime.GOOS)
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("erreur lors de la récupération des scanners: %v", err)
	}

	// Parser la sortie (à adapter selon le format exact)
	if len(output) > 0 {
		scanners = append(scanners, string(output))
	}

	return scanners, nil
}
