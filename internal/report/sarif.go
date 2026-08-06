package report

import (
	"encoding/json"
	"fmt"
)

// sarifDoc modela el subconjunto de SARIF 2.1.0 que consumimos.
type sarifDoc struct {
	Runs []struct {
		Tool struct {
			Driver struct {
				Name string `json:"name"`
			} `json:"driver"`
		} `json:"tool"`
		Results []struct {
			RuleID  string `json:"ruleId"`
			Level   string `json:"level"`
			Message struct {
				Text string `json:"text"`
			} `json:"message"`
			Locations []struct {
				PhysicalLocation struct {
					ArtifactLocation struct {
						URI string `json:"uri"`
					} `json:"artifactLocation"`
					Region struct {
						StartLine int `json:"startLine"`
					} `json:"region"`
				} `json:"physicalLocation"`
			} `json:"locations"`
		} `json:"results"`
	} `json:"runs"`
}

// levelToSeverity mapea el level SARIF (error/warning/note) a Severity.
func levelToSeverity(level string) Severity {
	switch level {
	case "error":
		return SeverityHigh
	case "warning":
		return SeverityMedium
	case "note":
		return SeverityLow
	default:
		return SeverityInfo
	}
}

// ParseSARIF convierte un documento SARIF en []Finding. Fail-soft: SARIF vacío
// devuelve 0 findings sin error; JSON inválido devuelve un error contextualizado.
// Los findings de categoría secret se redactan.
func ParseSARIF(data []byte) ([]Finding, error) {
	var doc sarifDoc
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("SARIF inválido: %w", err)
	}
	var findings []Finding
	for _, run := range doc.Runs {
		tool := run.Tool.Driver.Name
		category := categoryForTool(tool)
		for _, res := range run.Results {
			f := Finding{
				Tool:     tool,
				RuleID:   res.RuleID,
				Severity: levelToSeverity(res.Level),
				Message:  res.Message.Text,
				Category: category,
			}
			if len(res.Locations) > 0 {
				loc := res.Locations[0].PhysicalLocation
				f.File = loc.ArtifactLocation.URI
				f.Line = loc.Region.StartLine
			}
			if category == "secret" {
				f.Message = redact(f.Message)
			}
			findings = append(findings, f)
		}
	}
	return findings, nil
}
