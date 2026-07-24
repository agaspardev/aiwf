package security

import (
	_ "embed"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// ─── Structs del policy ────────────────────────────────────────────────────────

// Policy es la config de clasificación para el pipeline de seguridad.
// Se carga de block-warn-ignore.yaml (embebido en assets).
type Policy struct {
	Version    string              `yaml:"version"`
	Thresholds Thresholds          `yaml:"thresholds"`
	Gitleaks   ToolPolicy          `yaml:"gitleaks"`
	Semgrep    SemgrepPolicy       `yaml:"semgrep"`
	OSVScanner ToolSeverityPolicy  `yaml:"osv_scanner"`
	Trivy      TrivyPolicy         `yaml:"trivy"`
	Ignored    []IgnoredRule       `yaml:"ignored"`
}

type Thresholds struct {
	Block BlockThresholds `yaml:"block"`
	Warn  WarnThresholds  `yaml:"warn"`
}

type BlockThresholds struct {
	VerifiedSecrets      int `yaml:"verified_secrets"`
	CriticalVulns        int `yaml:"critical_vulnerabilities"`
	HighReachable        int `yaml:"high_reachable"`
}

type WarnThresholds struct {
	HighUnreachable string `yaml:"high_unreachable"`
	Medium          string `yaml:"medium"`
	Informational   string `yaml:"informational"`
}

type ToolPolicy struct {
	Default    string          `yaml:"default"`
	Exceptions []ToolException `yaml:"exceptions"`
}

type ToolException struct {
	RuleID string `yaml:"rule_id"`
	Reason string `yaml:"reason"`
}

type SemgrepPolicy struct {
	SeverityMap  map[string]string `yaml:"severity_map"`
	IgnoredRules []IgnoredRule     `yaml:"ignored_rules"`
}

type ToolSeverityPolicy struct {
	SeverityMap map[string]string `yaml:"severity_map"`
}

type TrivyPolicy struct {
	SeverityMap   map[string]string `yaml:"severity_map"`
	IgnoreUnfixed bool              `yaml:"ignore_unfixed"`
}

type IgnoredRule struct {
	Tool       string `yaml:"tool"`
	RuleID     string `yaml:"rule_id"`
	Reason     string `yaml:"reason"`
	Expires    string `yaml:"expires"`
	ApprovedBy string `yaml:"approved_by"`
}

// ─── Carga ─────────────────────────────────────────────────────────────────────

//go:generate echo "policy YAML loaded at build time via go:embed in assets package"

// LoadPolicy parsea el YAML de políticas y devuelve el Policy.
func LoadPolicy(data []byte) (*Policy, error) {
	var p Policy
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, fmt.Errorf("parseando policy: %w", err)
	}
	if p.Version == "" {
		return nil, fmt.Errorf("policy sin versión")
	}
	return &p, nil
}

// DefaultPolicy devuelve una política por defecto que replica el comportamiento
// histórico (clasificación por exit-code). Útil cuando no hay archivo de políticas.
func DefaultPolicy() *Policy {
	return &Policy{
		Version: "1.0-default",
		Gitleaks: ToolPolicy{
			Default: "BLOCK",
		},
		Semgrep: SemgrepPolicy{
			SeverityMap: map[string]string{
				"ERROR":   "BLOCK",
				"WARNING": "WARN",
				"INFO":    "IGNORE",
			},
		},
		OSVScanner: ToolSeverityPolicy{
			SeverityMap: map[string]string{
				"CRITICAL": "BLOCK",
				"HIGH":     "WARN",
				"MEDIUM":   "WARN",
				"LOW":      "IGNORE",
			},
		},
		Trivy: TrivyPolicy{
			SeverityMap: map[string]string{
				"CRITICAL": "BLOCK",
				"HIGH":     "WARN",
				"MEDIUM":   "WARN",
				"LOW":      "IGNORE",
			},
			IgnoreUnfixed: true,
		},
	}
}

// ─── Clasificación ─────────────────────────────────────────────────────────────

// ClassifyByExitCode asigna Status basado en exit code y la política por defecto
// de la herramienta. Es el modo legacy que usamos mientras no se parsean findings.
func (p *Policy) ClassifyByExitCode(tool string, exitCode int) string {
	if exitCode == 0 {
		return "OK"
	}
	switch tool {
	case "gitleaks":
		return pick(p.Gitleaks.Default, "BLOCK")
	case "semgrep":
		return pick(p.Semgrep.SeverityMap["WARNING"], "WARN")
	case "osv-scanner", "trivy":
		return pick(p.OSVScanner.SeverityMap["HIGH"], "WARN")
	default:
		return "WARN"
	}
}

// classifySeverity busca severity en la severity_map de la herramienta y devuelve
// el Status mapeado, o "" si no hay match.
func (p *Policy) classifySeverity(tool, severity string) string {
	severity = strings.ToUpper(severity)
	switch tool {
	case "semgrep":
		if p.Semgrep.SeverityMap != nil {
			if s, ok := p.Semgrep.SeverityMap[severity]; ok {
				return s
			}
		}
	case "osv-scanner":
		if p.OSVScanner.SeverityMap != nil {
			if s, ok := p.OSVScanner.SeverityMap[severity]; ok {
				return s
			}
		}
	case "trivy":
		if p.Trivy.SeverityMap != nil {
			if s, ok := p.Trivy.SeverityMap[severity]; ok {
				return s
			}
		}
	}
	return ""
}

// IsBlockedRule verifica si una rule_id específica está en la lista de ignorados
// (para overridear el BLOCK de una regla puntual).
func (p *Policy) IsBlockedRule(tool, ruleID string) bool {
	for _, ig := range p.Ignored {
		if ig.Tool == tool && ig.RuleID == ruleID {
			return false // está en la lista de ignorados → no bloquea
		}
	}
	switch tool {
	case "semgrep":
		for _, ir := range p.Semgrep.IgnoredRules {
			if ir.RuleID == ruleID {
				return false
			}
		}
	}
	return true
}

// pick es un pequeño helper que devuelve preferido, o fallback si preferido está vacío.
func pick(preferred, fallback string) string {
	if preferred != "" {
		return preferred
	}
	return fallback
}
