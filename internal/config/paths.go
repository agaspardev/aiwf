package config

import (
	"os"
	"path/filepath"
)

// InstallRoot es la raíz donde se despliega el overlay de aiwf. Por defecto el
// directorio de configuración del agente (~/.claude). Se puede sobreescribir con la
// variable de entorno AIWF_INSTALL_ROOT.
//
// TODO: resolver el target por-agente (gentle-ai configura múltiples agentes); por
// ahora asumimos Claude Code.
func InstallRoot() (string, error) {
	if v := os.Getenv("AIWF_INSTALL_ROOT"); v != "" {
		return v, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".claude"), nil
}

// ConfigFilePath es la ubicación del archivo de configuración de aiwf
// (formato JSON). Se puede sobreescribir con AIWF_CONFIG.
func ConfigFilePath() string {
	if v := os.Getenv("AIWF_CONFIG"); v != "" {
		return v
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".aiwf", "config.json")
}

// ManifestPath es la ubicación del manifiesto de aiwf (registro de lo aplicado).
// Se puede sobreescribir con la variable de entorno AIWF_MANIFEST.
func ManifestPath() (string, error) {
	if v := os.Getenv("AIWF_MANIFEST"); v != "" {
		return v, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".aiwf", "manifest.json"), nil
}
