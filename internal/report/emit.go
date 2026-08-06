package report

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// SchemaVersion es la versión del contrato ai-report.json. Cambiarla es un cambio
// consciente (rompe consumidores); ver internal/assets/files/schemas/ai-report.schema.json.
const SchemaVersion = "1.0"

// Report es el reporte agregado emitido como ai-report.json.
type Report struct {
	SchemaVersion string           `json:"schemaVersion"`
	Project       string           `json:"project"`
	Tools         []string         `json:"tools"`
	Counts        map[Severity]int `json:"counts"`
	Findings      []Finding        `json:"findings"`
}

// severityOrder define el orden de severidad de mayor a menor para el summary.
var severityOrder = []Severity{SeverityCritical, SeverityHigh, SeverityMedium, SeverityLow, SeverityInfo}

// BuildReport agrega findings en un Report con counts por severidad y tools únicos.
func BuildReport(project string, findings []Finding) Report {
	counts := map[Severity]int{}
	toolSet := map[string]bool{}
	for _, f := range findings {
		counts[f.Severity]++
		if f.Tool != "" {
			toolSet[f.Tool] = true
		}
	}
	tools := make([]string, 0, len(toolSet))
	for t := range toolSet {
		tools = append(tools, t)
	}
	sort.Strings(tools)
	if findings == nil {
		findings = []Finding{}
	}
	return Report{
		SchemaVersion: SchemaVersion,
		Project:       project,
		Tools:         tools,
		Counts:        counts,
		Findings:      findings,
	}
}

// validate comprueba el contrato mínimo antes de emitir (dependency-free, sin
// librería de JSON-schema: el schema embebido es el contrato documentado).
func (r Report) validate() error {
	if r.SchemaVersion == "" {
		return fmt.Errorf("report inválido: falta schemaVersion")
	}
	if r.Project == "" {
		return fmt.Errorf("report inválido: falta project")
	}
	return nil
}

// WriteReport valida y escribe ai-report.json + ai-summary.md en dir.
func WriteReport(dir string, r Report) error {
	if err := r.validate(); err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return fmt.Errorf("serializar ai-report: %w", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "ai-report.json"), data, 0o644); err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, "ai-summary.md"), []byte(r.summaryMarkdown()), 0o644)
}

// summaryMarkdown produce el resumen humano legible.
func (r Report) summaryMarkdown() string {
	var b strings.Builder
	fmt.Fprintf(&b, "# AI Report — %s\n\n", r.Project)
	fmt.Fprintf(&b, "schema: %s | tools: %s\n\n", r.SchemaVersion, strings.Join(r.Tools, ", "))
	b.WriteString("## Counts por severidad\n\n")
	for _, sev := range severityOrder {
		if n := r.Counts[sev]; n > 0 {
			fmt.Fprintf(&b, "- %s: %d\n", sev, n)
		}
	}
	b.WriteString("\n## Findings\n\n")
	for _, f := range r.Findings {
		loc := f.File
		if f.Line > 0 {
			loc = fmt.Sprintf("%s:%d", f.File, f.Line)
		}
		fmt.Fprintf(&b, "- [%s] %s (%s) %s — %s\n", f.Severity, f.RuleID, f.Tool, loc, f.Message)
	}
	return b.String()
}
