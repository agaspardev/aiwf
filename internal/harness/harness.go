// Package harness porta el lanzador de sesiones del workflow (originalmente
// harness.ps1): resuelve un modo desde modes.json y construye la invocación de
// `claude` (modelo/combo, permission-mode, add-dir, mcp-config, contract prompt).
// El paquete es puro (no ejecuta procesos ni toca env); el cmd se encarga de eso.
package harness

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// Mode es la configuración de un modo de trabajo (entrada de modes.json).
// F1: el modo es un CATÁLOGO de infra negociable ({combo, aux, capabilityGate}).
// La metodología/gobernanza (permissionMode, gateSet, risk, contract, directives)
// NO vive aquí — es invariante (governanceCore + harness). Quitar estos campos
// hace la invariancia estructural: el compilador impide que un modo la declare.
type Mode struct {
	Description     string   `json:"description"`
	Combo           string   `json:"combo"`
	AuxiliaryCombos []string `json:"auxiliaryCombos,omitempty"`
	CapabilityGate  bool     `json:"capabilityGate,omitempty"`
}

// forbiddenModeKeys son claves de gobernanza que un modo NO puede declarar (F1).
var forbiddenModeKeys = []string{"permissionMode", "risk", "contract", "gateSet", "directives"}

// Modes es el contenido de modes.json.
type Modes struct {
	SchemaVersion int             `json:"schemaVersion"`
	DefaultMode   string          `json:"defaultMode"`
	Modes         map[string]Mode `json:"modes"`
}

// LoadModes parsea el JSON de modes.json y aplica el guard de invariancia (F1):
// rechaza esquemas obsoletos (v1) y cualquier modo que declare gobernanza.
func LoadModes(data []byte) (*Modes, error) {
	var m Modes
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("modes.json inválido: %w", err)
	}
	if len(m.Modes) == 0 {
		return nil, fmt.Errorf("modes.json no define modos")
	}
	if m.SchemaVersion < 2 {
		return nil, fmt.Errorf("modes.json schemaVersion %d es obsoleto: la gobernanza por-modo se eliminó en F1; re-aplicá el overlay aiwf para regenerarlo (schemaVersion 2)", m.SchemaVersion)
	}
	// Guard estructural: ningún modo puede declarar claves de gobernanza.
	var probe struct {
		Modes map[string]map[string]json.RawMessage `json:"modes"`
	}
	if err := json.Unmarshal(data, &probe); err != nil {
		return nil, fmt.Errorf("modes.json inválido: %w", err)
	}
	for name, fields := range probe.Modes {
		for _, k := range forbiddenModeKeys {
			if _, present := fields[k]; present {
				return nil, fmt.Errorf("modo %q declara %q: la gobernanza es invariante y no puede definirse por-modo (F1); eliminá esa clave", name, k)
			}
		}
	}
	return &m, nil
}

// Names devuelve los nombres de modo ordenados (para mensajes de ayuda/error).
func (m *Modes) Names() []string {
	names := make([]string, 0, len(m.Modes))
	for k := range m.Modes {
		names = append(names, k)
	}
	sort.Strings(names)
	return names
}

// Resolve devuelve el modo por nombre, o un error listando los disponibles.
func (m *Modes) Resolve(name string) (Mode, error) {
	if mode, ok := m.Modes[name]; ok {
		return mode, nil
	}
	return Mode{}, fmt.Errorf("modo %q no existe; disponibles: %s", name, strings.Join(m.Names(), ", "))
}

// LaunchOptions parametriza la construcción de la invocación de claude.
type LaunchOptions struct {
	// OmniActive indica si OmniRoute está disponible y se debe enrutar por combo.
	OmniActive bool
	// SkipPermissions usa --dangerously-skip-permissions en lugar del permissionMode.
	SkipPermissions bool
	// VaultDir se pasa como --add-dir (si no está vacío).
	VaultDir string
	// AdditionalDirs son --add-dir extra (ya resueltos y existentes).
	AdditionalDirs []string
	// MCPConfig es la ruta al mcp-config (si no está vacía).
	MCPConfig string
	// Contract es el system prompt a anexar (--append-system-prompt).
	Contract string
	// PermissionMode es el permiso DERIVADO de la certificación de capacidad
	// (F1/1B): "acceptEdits" solo si el combo está certificado como agent con
	// escritura local; "default" (supervisado) en caso contrario. El modo ya no
	// lo declara — lo deriva el harness.
	PermissionMode string
}

// ValidateLaunchRequirements evita degradar silenciosamente un modo con garantías
// de capacidad cuando OmniRoute no está disponible.
func ValidateLaunchRequirements(mode Mode, omniActive bool) error {
	if mode.CapabilityGate && !omniActive {
		return fmt.Errorf("el modo con capabilityGate requiere OmniRoute disponible")
	}
	return nil
}

// BuildClaudeArgs construye los argumentos de `claude` para un modo, replicando la
// lógica de harness.ps1. Función pura y determinista.
func BuildClaudeArgs(mode Mode, opts LaunchOptions) []string {
	args := make([]string, 0, 12)

	if opts.OmniActive && mode.Combo != "" {
		args = append(args, "--model", mode.Combo)
	}

	if opts.SkipPermissions {
		args = append(args, "--dangerously-skip-permissions")
	} else {
		permission := opts.PermissionMode
		if permission == "" {
			permission = "default" // fail-closed a supervisado si no se derivó
		}
		args = append(args, "--permission-mode", permission)
	}

	if opts.VaultDir != "" {
		args = append(args, "--add-dir", opts.VaultDir)
	}
	for _, d := range opts.AdditionalDirs {
		args = append(args, "--add-dir", d)
	}

	if opts.MCPConfig != "" {
		args = append(args, "--mcp-config", opts.MCPConfig)
	}

	if opts.Contract != "" {
		args = append(args, "--append-system-prompt", opts.Contract)
	}
	return args
}
