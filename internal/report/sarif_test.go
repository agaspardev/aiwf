package report

import (
	"strings"
	"testing"
)

const sarifSemgrep = `{
  "runs": [{
    "tool": {"driver": {"name": "semgrep"}},
    "results": [
      {"ruleId": "rules.sqli", "level": "error",
       "message": {"text": "SQL injection"},
       "locations": [{"physicalLocation": {
         "artifactLocation": {"uri": "app/db.go"},
         "region": {"startLine": 42}}}]},
      {"ruleId": "rules.weak", "level": "warning",
       "message": {"text": "weak hash"},
       "locations": [{"physicalLocation": {
         "artifactLocation": {"uri": "app/hash.go"},
         "region": {"startLine": 7}}}]}
    ]
  }]
}`

func TestParseSARIFMapsFindings(t *testing.T) {
	fs, err := ParseSARIF([]byte(sarifSemgrep))
	if err != nil {
		t.Fatalf("ParseSARIF: %v", err)
	}
	if len(fs) != 2 {
		t.Fatalf("esperaba 2 findings, got %d", len(fs))
	}
	f := fs[0]
	if f.Tool != "semgrep" || f.RuleID != "rules.sqli" || f.Severity != SeverityHigh ||
		f.File != "app/db.go" || f.Line != 42 || f.Category != "sast" {
		t.Fatalf("finding mal mapeado: %+v", f)
	}
	if fs[1].Severity != SeverityMedium {
		t.Errorf("warning debería ser medium, got %s", fs[1].Severity)
	}
}

func TestParseSARIFEmptyIsZeroFindings(t *testing.T) {
	fs, err := ParseSARIF([]byte(`{"runs":[]}`))
	if err != nil {
		t.Fatalf("SARIF vacío no debería dar error: %v", err)
	}
	if len(fs) != 0 {
		t.Fatalf("esperaba 0 findings, got %d", len(fs))
	}
}

func TestParseSARIFMissingLocation(t *testing.T) {
	data := `{"runs":[{"tool":{"driver":{"name":"osv-scanner"}},
	  "results":[{"ruleId":"CVE-1","level":"error","message":{"text":"vuln"}}]}]}`
	fs, err := ParseSARIF([]byte(data))
	if err != nil {
		t.Fatalf("ParseSARIF: %v", err)
	}
	if len(fs) != 1 || fs[0].File != "" || fs[0].Line != 0 {
		t.Fatalf("ubicación ausente debería dar File=''/Line=0: %+v", fs)
	}
	if fs[0].Category != "sca" {
		t.Errorf("osv-scanner debería ser sca, got %s", fs[0].Category)
	}
}

func TestParseSARIFTruncatedErrors(t *testing.T) {
	if _, err := ParseSARIF([]byte(`{"runs":[{"tool"`)); err == nil {
		t.Error("JSON truncado debería dar error contextualizado")
	}
}

func TestParseSARIFRedactsSecretMessages(t *testing.T) {
	// gitleaks (category=secret) con un secreto en el message → nunca crudo.
	data := `{"runs":[{"tool":{"driver":{"name":"gitleaks"}},"results":[{"ruleId":"aws-key","level":"error","message":{"text":"found AKIAIOSFODNN7EXAMPLE in config"},"locations":[{"physicalLocation":{"artifactLocation":{"uri":"c.env"},"region":{"startLine":3}}}]}]}]}`
	fs, err := ParseSARIF([]byte(data))
	if err != nil {
		t.Fatalf("ParseSARIF: %v", err)
	}
	if len(fs) != 1 {
		t.Fatalf("esperaba 1 finding")
	}
	if strings.Contains(fs[0].Message, "AKIAIOSFODNN7EXAMPLE") {
		t.Fatalf("el secreto NO debe aparecer crudo en el message: %q", fs[0].Message)
	}
	if !strings.Contains(fs[0].Message, "[REDACTED]") {
		t.Errorf("esperaba marca [REDACTED] en el message redactado: %q", fs[0].Message)
	}
}
