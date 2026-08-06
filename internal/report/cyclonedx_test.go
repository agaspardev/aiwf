package report

import "testing"

const cyclonedxVulns = `{
  "metadata": {"tools": [{"name": "trivy"}]},
  "vulnerabilities": [
    {"id": "CVE-2024-1", "description": "rce en libX",
     "ratings": [{"severity": "critical"}],
     "affects": [{"ref": "pkg:golang/libx@1.0"}]},
    {"id": "CVE-2024-2", "description": "dos en libY",
     "ratings": [{"severity": "medium"}],
     "affects": [{"ref": "pkg:golang/liby@2.0"}]}
  ]
}`

func TestParseCycloneDXMapsVulnerabilities(t *testing.T) {
	fs, err := ParseCycloneDX([]byte(cyclonedxVulns))
	if err != nil {
		t.Fatalf("ParseCycloneDX: %v", err)
	}
	if len(fs) != 2 {
		t.Fatalf("esperaba 2 findings, got %d", len(fs))
	}
	if fs[0].RuleID != "CVE-2024-1" || fs[0].Severity != SeverityCritical ||
		fs[0].Category != "sca" || fs[0].File != "pkg:golang/libx@1.0" {
		t.Fatalf("finding mal mapeado: %+v", fs[0])
	}
	if fs[1].Severity != SeverityMedium {
		t.Errorf("segunda severity=%s, esperaba medium", fs[1].Severity)
	}
}

func TestParseCycloneDXPureSBOMHasNoFindings(t *testing.T) {
	// SBOM sin vulnerabilities es inventario, no hallazgos.
	fs, err := ParseCycloneDX([]byte(`{"components":[{"name":"libz"}]}`))
	if err != nil {
		t.Fatalf("ParseCycloneDX: %v", err)
	}
	if len(fs) != 0 {
		t.Fatalf("SBOM puro esperaba 0 findings, got %d", len(fs))
	}
}

func TestParseCycloneDXTruncatedErrors(t *testing.T) {
	if _, err := ParseCycloneDX([]byte(`{"vulnerabilities":[`)); err == nil {
		t.Error("JSON truncado debería dar error")
	}
}
