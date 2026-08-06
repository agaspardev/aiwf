package report

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func sampleFindings() []Finding {
	return []Finding{
		{Tool: "semgrep", RuleID: "sqli", Severity: SeverityHigh, Message: "SQLi", File: "a.go", Line: 1, Category: "sast"},
		{Tool: "trivy", RuleID: "CVE-1", Severity: SeverityCritical, Message: "rce", File: "pkg", Category: "sca"},
		{Tool: "semgrep", RuleID: "weak", Severity: SeverityMedium, Message: "weak", File: "b.go", Line: 2, Category: "sast"},
	}
}

func TestBuildReportComputesCountsAndTools(t *testing.T) {
	r := BuildReport("demo", sampleFindings())
	if r.Project != "demo" || r.SchemaVersion == "" {
		t.Fatalf("report base incompleto: %+v", r)
	}
	if r.Counts[SeverityHigh] != 1 || r.Counts[SeverityCritical] != 1 || r.Counts[SeverityMedium] != 1 {
		t.Fatalf("counts mal computados: %+v", r.Counts)
	}
	// tools únicos y ordenados
	if len(r.Tools) != 2 || r.Tools[0] != "semgrep" || r.Tools[1] != "trivy" {
		t.Fatalf("tools = %v, esperaba [semgrep trivy]", r.Tools)
	}
}

func TestWriteReportEmitsJSONAndSummary(t *testing.T) {
	dir := t.TempDir()
	r := BuildReport("demo", sampleFindings())
	if err := WriteReport(dir, r); err != nil {
		t.Fatalf("WriteReport: %v", err)
	}

	jsonData, err := os.ReadFile(filepath.Join(dir, "ai-report.json"))
	if err != nil {
		t.Fatalf("ai-report.json: %v", err)
	}
	var back Report
	if err := json.Unmarshal(jsonData, &back); err != nil {
		t.Fatalf("ai-report.json no es JSON válido: %v", err)
	}
	if back.Project != "demo" || len(back.Findings) != 3 {
		t.Fatalf("ai-report.json mal serializado: %+v", back)
	}

	md, err := os.ReadFile(filepath.Join(dir, "ai-summary.md"))
	if err != nil {
		t.Fatalf("ai-summary.md: %v", err)
	}
	summary := string(md)
	for _, want := range []string{"demo", "critical", "high", "CVE-1"} {
		if !strings.Contains(summary, want) {
			t.Errorf("ai-summary.md no contiene %q", want)
		}
	}
}

func TestWriteReportRejectsInvalidReport(t *testing.T) {
	dir := t.TempDir()
	if err := WriteReport(dir, Report{Findings: sampleFindings()}); err == nil {
		t.Error("esperaba error: report sin SchemaVersion/Project es inválido")
	}
}
