// Package gatekeeper valida un phase-contract contra la REALIDAD del repo (determinista,
// cero tokens): ¿los artefactos declarados existen? ¿el estado es coherente? Portado de
// gatekeeper.ps1. P8: el código valida hechos; el modelo no se autocertifica.
package gatekeeper

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type artifact struct {
	Path string `json:"path"`
}
type risk struct {
	Severity string `json:"severity"`
}

// Result es el resultado de validar un contrato.
type Result struct {
	Phase         string
	NextPhase     string
	ArtifactCount int
	Problems      []string
}

// Passed indica si el gate pasa (sin problemas).
func (r Result) Passed() bool { return len(r.Problems) == 0 }

// requiredFields son los campos mínimos que todo contrato debe declarar.
var requiredFields = []string{"phase", "status", "executive_summary", "artifacts", "next_phase"}

// Validate valida el contrato en contractPath contra el repo en root.
func Validate(root, contractPath string) (Result, error) {
	data, err := os.ReadFile(contractPath)
	if err != nil {
		return Result{}, fmt.Errorf("contrato no encontrado: %w", err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return Result{}, fmt.Errorf("contrato no es JSON válido: %w", err)
	}

	var res Result
	for _, f := range requiredFields {
		if _, ok := raw[f]; !ok {
			res.Problems = append(res.Problems, "campo faltante: "+f)
		}
	}

	var status string
	_ = json.Unmarshal(raw["phase"], &res.Phase)
	_ = json.Unmarshal(raw["status"], &status)
	_ = json.Unmarshal(raw["next_phase"], &res.NextPhase)

	var arts []artifact
	_ = json.Unmarshal(raw["artifacts"], &arts)
	for _, a := range arts {
		if a.Path == "" {
			res.Problems = append(res.Problems, "artefacto sin path")
			continue
		}
		res.ArtifactCount++
		p := a.Path
		if !filepath.IsAbs(p) {
			p = filepath.Join(root, p)
		}
		if _, err := os.Stat(p); err != nil {
			res.Problems = append(res.Problems, "artefacto declarado NO existe: "+a.Path)
		}
	}

	if status == "passed" && res.ArtifactCount == 0 && res.Phase != "explore" {
		res.Problems = append(res.Problems, "status=passed sin ningún artefacto producido (posible alucinación)")
	}

	var risks []risk
	_ = json.Unmarshal(raw["risks"], &risks)
	crit := 0
	for _, rk := range risks {
		if rk.Severity == "critical" {
			crit++
		}
	}
	if status == "passed" && crit > 0 {
		res.Problems = append(res.Problems, fmt.Sprintf("status=passed con %d riesgo(s) critico(s) sin resolver", crit))
	}
	return res, nil
}
