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

// ValidateMutableModes falla de forma cerrada: cada modelo de un combo primario con
// autoridad de edición debe estar registrado y completamente certificado como agent.
func ValidateMutableModes(modes *Modes, combos []ComboDefinition, capabilities ModelCapabilities) []CapabilityViolation {
	combosByName := make(map[string]ComboDefinition, len(combos))
	for _, combo := range combos {
		combosByName[combo.Name] = combo
	}

	var violations []CapabilityViolation
	for modeName, mode := range modes.Modes {
		if mode.PermissionMode != "acceptEdits" {
			continue
		}
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

// ValidateAuxiliaryPolicy comprueba que un modo mutable que consulta auxiliares
// conserve explícitamente la mutación y la verificación final en el agente principal.
func ValidateAuxiliaryPolicy(modeName string, mode Mode) []CapabilityViolation {
	if mode.PermissionMode != "acceptEdits" || len(mode.AuxiliaryCombos) == 0 {
		return nil
	}

	directives := strings.ToLower(strings.Join(mode.Directives, "\n"))
	checks := []struct {
		needle string
		reason string
	}{
		{needle: "filesystem mutations", reason: "auxiliary mutation guard is missing"},
		{needle: "final verification", reason: "auxiliary final-verification guard is missing"},
	}

	var violations []CapabilityViolation
	for _, check := range checks {
		if !strings.Contains(directives, "never delegate "+check.needle) {
			violations = append(violations, CapabilityViolation{
				Mode: modeName, Combo: mode.Combo, Reason: check.reason,
			})
		}
	}
	return violations
}

// ValidateConfiguredModes combina las reglas de capacidad y delegación en un único gate.
func ValidateConfiguredModes(modes *Modes, combos []ComboDefinition, capabilities ModelCapabilities) []CapabilityViolation {
	violations := ValidateMutableModes(modes, combos, capabilities)
	for modeName, mode := range modes.Modes {
		violations = append(violations, ValidateAuxiliaryPolicy(modeName, mode)...)
	}
	sortCapabilityViolations(violations)
	return violations
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
