// Package config expone la configuración de aiwf: principalmente la versión de
// gentle-ai que el instalador toma como objetivo (pin). Se carga desde un archivo
// JSON en ~/.aiwf/config.json (o AIWF_CONFIG), con fallback a valores por defecto.
package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// PinnedGentleAIVersion es la versión objetivo por defecto de gentle-ai que aiwf
// instala o actualiza. Se sobreescribe desde config.json en runtime.
const PinnedGentleAIVersion = "0.0.0"

// Config contiene los parámetros de instalación de aiwf.
type Config struct {
	// GentleAIVersion es la versión de gentle-ai que se instala/actualiza.
	GentleAIVersion string `json:"gentle_ai_version,omitempty"`
}

// Default devuelve la configuración por defecto (usada mientras no exista
// config.json).
func Default() Config {
	return Config{GentleAIVersion: PinnedGentleAIVersion}
}

// LoadConfig lee y parsea un archivo JSON de configuración. Si el archivo no
// existe, devuelve Config vacío (usa Default posteriormente). Errores de parseo
// sí se propagan.
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Config{}, nil
		}
		return nil, fmt.Errorf("leyendo config %s: %w", path, err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parseando config %s: %w", path, err)
	}
	return &cfg, nil
}

// LoadOrDefault carga la config desde path con LoadConfig y la fusiona con
// Default: los campos vacíos en el archivo heredan el valor por defecto.
func LoadOrDefault(path string) Config {
	cfg := Default()
	fileCfg, err := LoadConfig(path)
	if err == nil && fileCfg != nil {
		if fileCfg.GentleAIVersion != "" {
			cfg.GentleAIVersion = fileCfg.GentleAIVersion
		}
	}
	return cfg
}
