package harness

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// CapabilityClass expresa si un modelo puede controlar el loop local del agente.
type CapabilityClass string

const (
	CapabilityAgent     CapabilityClass = "agent"
	CapabilityAuxiliary CapabilityClass = "auxiliary"
	CapabilityBlocked   CapabilityClass = "blocked"
)

// ModelCapability describe las capacidades locales certificadas de un alias efectivo.
type ModelCapability struct {
	Class      CapabilityClass `json:"class"`
	ToolCall   bool            `json:"toolCall"`
	LocalRead  bool            `json:"localRead"`
	LocalWrite bool            `json:"localWrite"`
	Reason     string          `json:"reason,omitempty"`
}

// ModelCapabilities es el registro versionado de capacidades del harness.
type ModelCapabilities struct {
	SchemaVersion int                        `json:"schemaVersion"`
	LastValidated string                     `json:"lastValidated,omitempty"`
	Models        map[string]ModelCapability `json:"models"`
}

// LoadModelCapabilities parsea el registro de capacidades.
func LoadModelCapabilities(data []byte) (ModelCapabilities, error) {
	var capabilities ModelCapabilities
	if err := json.Unmarshal(data, &capabilities); err != nil {
		return ModelCapabilities{}, fmt.Errorf("model-capabilities.json inválido: %w", err)
	}
	if len(capabilities.Models) == 0 {
		return ModelCapabilities{}, fmt.Errorf("model-capabilities.json no define modelos")
	}
	return capabilities, nil
}

// ComboDefinition contiene solo los datos que necesita la política de capacidades.
type ComboDefinition struct {
	Name   string
	Models []string
}

// CapabilityViolation identifica una configuración insegura de manera accionable.
type CapabilityViolation struct {
	Mode   string
	Combo  string
	Model  string
	Reason string
}

// comboCertified indica si TODOS los modelos de un combo están certificados como
// agent con escritura local. Fail-closed: combo ausente/vacío o modelo sin
// registro cuenta como NO certificado.
func comboCertified(combo ComboDefinition, capabilities ModelCapabilities) bool {
	if len(combo.Models) == 0 {
		return false
	}
	for _, model := range combo.Models {
		capability, ok := capabilities.Models[model]
		if !ok {
			return false
		}
		if capability.Class != CapabilityAgent || !capability.ToolCall || !capability.LocalRead || !capability.LocalWrite {
			return false
		}
	}
	return true
}

// DerivePermissionMode (F1/1B) deriva el permission-mode de la certificación del
// combo primario: certificado -> "acceptEdits"; no certificado -> "default"
// (supervisado). El modo ya no lo declara; el harness lo decide de forma
// determinista, preservando la guarda de seguridad sin bloquear modos supervisados.
func DerivePermissionMode(mode Mode, combos []ComboDefinition, capabilities ModelCapabilities) string {
	combosByName := make(map[string]ComboDefinition, len(combos))
	for _, combo := range combos {
		combosByName[combo.Name] = combo
	}
	if comboCertified(combosByName[mode.Combo], capabilities) {
		return "acceptEdits"
	}
	return "default"
}

// ValidateMutableModes falla de forma cerrada: cada modelo del combo primario de
// un modo debe estar registrado y completamente certificado como agent. En runtime
// solo se invoca sobre el modo con capabilityGate (ver ValidateSelectedMode): un
// modo que EXIGE certificación pero cuyo combo no la tiene se rechaza.
func ValidateMutableModes(modes *Modes, combos []ComboDefinition, capabilities ModelCapabilities) []CapabilityViolation {
	combosByName := make(map[string]ComboDefinition, len(combos))
	for _, combo := range combos {
		combosByName[combo.Name] = combo
	}

	var violations []CapabilityViolation
	for modeName, mode := range modes.Modes {
		combo, ok := combosByName[mode.Combo]
		if !ok {
			violations = append(violations, CapabilityViolation{
				Mode: modeName, Combo: mode.Combo, Reason: "primary combo is not defined",
			})
			continue
		}
		if len(combo.Models) == 0 {
			violations = append(violations, CapabilityViolation{
				Mode: modeName, Combo: mode.Combo, Reason: "primary combo has no models",
			})
			continue
		}
		for _, model := range combo.Models {
			capability, ok := capabilities.Models[model]
			if !ok {
				violations = append(violations, CapabilityViolation{
					Mode: modeName, Combo: mode.Combo, Model: model, Reason: "model capability is not registered",
				})
				continue
			}
			if capability.Class != CapabilityAgent || !capability.ToolCall || !capability.LocalRead || !capability.LocalWrite {
				violations = append(violations, CapabilityViolation{
					Mode: modeName, Combo: mode.Combo, Model: model, Reason: "model is not certified for local agent writes",
				})
			}
		}
	}

	sortCapabilityViolations(violations)
	return violations
}

// ValidateConfiguredModes aplica la regla de capacidad. La política de auxiliares
// (nunca delegar mutaciones/verificación final) ya NO se valida por-modo: es
// invariante y vive en governanceCore (F1/Decisión 2), verificada por test.
func ValidateConfiguredModes(modes *Modes, combos []ComboDefinition, capabilities ModelCapabilities) []CapabilityViolation {
	return ValidateMutableModes(modes, combos, capabilities)
}

// ValidateSelectedMode aplica el gate al modo que se va a ejecutar. Otros modos
// pueden tener registros de capacidades incompletos sin bloquear esta sesión.
func ValidateSelectedMode(modeName string, modes *Modes, combos []ComboDefinition, capabilities ModelCapabilities) ([]CapabilityViolation, error) {
	mode, err := modes.Resolve(modeName)
	if err != nil {
		return nil, err
	}
	if !mode.CapabilityGate {
		return nil, nil
	}
	selected := &Modes{
		SchemaVersion: modes.SchemaVersion,
		DefaultMode:   modeName,
		Modes:         map[string]Mode{modeName: mode},
	}
	return ValidateConfiguredModes(selected, combos, capabilities), nil
}

// FormatCapabilityViolations produce evidencia estable para CLI, tests y reportes.
func FormatCapabilityViolations(violations []CapabilityViolation) string {
	ordered := append([]CapabilityViolation(nil), violations...)
	sortCapabilityViolations(ordered)
	lines := make([]string, 0, len(ordered))
	for _, violation := range ordered {
		location := violation.Mode + "/" + violation.Combo
		if violation.Model != "" {
			location += "/" + violation.Model
		}
		lines = append(lines, location+": "+violation.Reason)
	}
	return strings.Join(lines, "\n")
}

func sortCapabilityViolations(violations []CapabilityViolation) {
	sort.Slice(violations, func(i, j int) bool {
		left := violations[i].Mode + "\x00" + violations[i].Combo + "\x00" + violations[i].Model + "\x00" + violations[i].Reason
		right := violations[j].Mode + "\x00" + violations[j].Combo + "\x00" + violations[j].Model + "\x00" + violations[j].Reason
		return left < right
	})
}
