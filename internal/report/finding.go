// Package report agrega los reportes nativos de los scanners (SARIF/CycloneDX)
// en un modelo Finding unificado, redactado y validado. Determinista, sin red.
package report

// Severity es la severidad normalizada de un hallazgo.
type Severity string

const (
	SeverityCritical Severity = "critical"
	SeverityHigh     Severity = "high"
	SeverityMedium   Severity = "medium"
	SeverityLow      Severity = "low"
	SeverityInfo     Severity = "info"
)

// Finding es un hallazgo normalizado, agnóstico de la herramienta de origen.
// Message SIEMPRE está redactado: nunca contiene el secreto o código crudo.
type Finding struct {
	Tool     string   `json:"tool"`
	RuleID   string   `json:"ruleId"`
	Severity Severity `json:"severity"`
	Message  string   `json:"message"`
	File     string   `json:"file"`
	Line     int      `json:"line"`
	Category string   `json:"category"` // secret|sast|sca|sbom|vuln
}

// categoryForTool mapea el nombre del driver a la categoría del dominio.
func categoryForTool(tool string) string {
	switch tool {
	case "gitleaks":
		return "secret"
	case "semgrep":
		return "sast"
	case "trivy", "osv-scanner", "osv":
		return "sca"
	case "govulncheck":
		return "vuln"
	case "syft":
		return "sbom"
	default:
		return "sast"
	}
}
