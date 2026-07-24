package security

import (
	"testing"
)

func TestLoadPolicy(t *testing.T) {
	p, err := LoadPolicy([]byte(`version: "1.0"
gitleaks:
  default: BLOCK
  exceptions: []
semgrep:
  severity_map:
    ERROR: BLOCK
    WARNING: WARN
    INFO: IGNORE
osv_scanner:
  severity_map:
    CRITICAL: BLOCK
trivy:
  severity_map:
    CRITICAL: BLOCK
  ignore_unfixed: true
`))
	if err != nil {
		t.Fatalf("LoadPolicy: %v", err)
	}
	if p.Version != "1.0" {
		t.Errorf("Version = %q, want 1.0", p.Version)
	}
	if p.Gitleaks.Default != "BLOCK" {
		t.Errorf("Gitleaks.Default = %q, want BLOCK", p.Gitleaks.Default)
	}
	if p.Trivy.IgnoreUnfixed != true {
		t.Error("Trivy.IgnoreUnfixed debería ser true")
	}
}

func TestLoadPolicyRejectsNoVersion(t *testing.T) {
	_, err := LoadPolicy([]byte(`gitleaks: {default: BLOCK}`))
	if err == nil {
		t.Fatal("esperaba error por policy sin versión")
	}
}

func TestDefaultPolicy(t *testing.T) {
	p := DefaultPolicy()
	if p.Version == "" {
		t.Error("DefaultPolicy sin version")
	}
	if p.Gitleaks.Default != "BLOCK" {
		t.Errorf("Gitleaks.Default = %q, want BLOCK", p.Gitleaks.Default)
	}
}

func TestClassifyByExitCode(t *testing.T) {
	p := DefaultPolicy()
	tests := []struct {
		tool     string
		exitCode int
		want     string
	}{
		{"gitleaks", 0, "OK"},
		{"gitleaks", 1, "BLOCK"},
		{"semgrep", 0, "OK"},
		{"semgrep", 1, "WARN"},
		{"osv-scanner", 0, "OK"},
		{"osv-scanner", 1, "WARN"},
	}
	for _, tt := range tests {
		got := p.ClassifyByExitCode(tt.tool, tt.exitCode)
		if got != tt.want {
			t.Errorf("ClassifyByExitCode(%q, %d) = %q, want %q", tt.tool, tt.exitCode, got, tt.want)
		}
	}
}

func TestClassifySeverity(t *testing.T) {
	p := DefaultPolicy()
	tests := []struct {
		tool     string
		severity string
		want     string
	}{
		{"semgrep", "ERROR", "BLOCK"},
		{"semgrep", "WARNING", "WARN"},
		{"semgrep", "INFO", "IGNORE"},
		{"semgrep", "UNKNOWN", ""},
		{"osv-scanner", "CRITICAL", "BLOCK"},
		{"osv-scanner", "HIGH", "WARN"},
		{"osv-scanner", "MEDIUM", "WARN"},
		{"osv-scanner", "LOW", "IGNORE"},
		{"trivy", "CRITICAL", "BLOCK"},
		{"trivy", "HIGH", "WARN"},
	}
	for _, tt := range tests {
		got := p.classifySeverity(tt.tool, tt.severity)
		if got != tt.want {
			t.Errorf("classifySeverity(%q, %q) = %q, want %q", tt.tool, tt.severity, got, tt.want)
		}
	}
}

func TestIsBlockedRule(t *testing.T) {
	p := DefaultPolicy()
	if !p.IsBlockedRule("semgrep", "some-rule") {
		t.Error("regla sin ignorar debería estar bloqueada")
	}

	p2, _ := LoadPolicy([]byte(`version: "1.0"
semgrep:
  severity_map:
    ERROR: BLOCK
  ignored_rules:
    - rule_id: "test-rule"
      reason: "falso positivo"
gitleaks:
  default: BLOCK
`))
	if p2.IsBlockedRule("semgrep", "test-rule") {
		t.Error("test-rule está en ignored_rules, no debería bloquear")
	}
	if !p2.IsBlockedRule("semgrep", "other-rule") {
		t.Error("other-rule no está ignorada, debería bloquear")
	}
}

func TestDefaultPolicyCoversExitCodeBehavior(t *testing.T) {
	// Verifica que DefaultPolicy clasifica igual que la lógica hardcodeada actual.
	runner := NewRunner(".", t.TempDir())
	results := make([]ToolResult, 0, 3)
	p := DefaultPolicy()

	// gitleaks exit 1 → BLOCK.
	gitleaks := runner.Secrets()
	gitleaks.ExitCode = 1
	gitleaks.Status = p.ClassifyByExitCode("gitleaks", 1)
	results = append(results, gitleaks)

	// semgrep exit 1 → WARN.
	semgrep := runner.Sast()
	semgrep.ExitCode = 1
	semgrep.Status = p.ClassifyByExitCode("semgrep", 1)
	results = append(results, semgrep)

	if ExitCode(results) != 1 {
		t.Error("gitleaks BLOCK + semgrep WARN debería dar exit=1")
	}
}
