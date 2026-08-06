package report

import (
	"encoding/json"
	"fmt"
)

type cyclonedxDoc struct {
	Metadata struct {
		Tools []struct {
			Name string `json:"name"`
		} `json:"tools"`
	} `json:"metadata"`
	Vulnerabilities []struct {
		ID          string `json:"id"`
		Description string `json:"description"`
		Ratings     []struct {
			Severity string `json:"severity"`
		} `json:"ratings"`
		Affects []struct {
			Ref string `json:"ref"`
		} `json:"affects"`
	} `json:"vulnerabilities"`
}

// cyclonedxSeverity mapea la severidad CycloneDX a Severity.
func cyclonedxSeverity(s string) Severity {
	switch s {
	case "critical":
		return SeverityCritical
	case "high":
		return SeverityHigh
	case "medium":
		return SeverityMedium
	case "low":
		return SeverityLow
	default:
		return SeverityInfo
	}
}

// ParseCycloneDX extrae los vulnerabilities[] de un documento CycloneDX como
// Finding[] (category=sca). Un SBOM puro (solo components) no produce findings.
func ParseCycloneDX(data []byte) ([]Finding, error) {
	var doc cyclonedxDoc
	if err := json.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("CycloneDX inválido: %w", err)
	}
	tool := "cyclonedx"
	if len(doc.Metadata.Tools) > 0 && doc.Metadata.Tools[0].Name != "" {
		tool = doc.Metadata.Tools[0].Name
	}
	var findings []Finding
	for _, v := range doc.Vulnerabilities {
		f := Finding{
			Tool:     tool,
			RuleID:   v.ID,
			Message:  v.Description,
			Category: "sca",
			Severity: SeverityInfo,
		}
		if len(v.Ratings) > 0 {
			f.Severity = cyclonedxSeverity(v.Ratings[0].Severity)
		}
		if len(v.Affects) > 0 {
			f.File = v.Affects[0].Ref
		}
		findings = append(findings, f)
	}
	return findings, nil
}
