package upstream

import (
	"testing"
)

const sampleAdvisoriesJSON = `[
  {
    "ghsa_id": "GHSA-xxxx-xxxx-xxxx",
    "severity": "high",
    "published_at": "2026-07-01T00:00:00Z",
    "summary": "Buffer overflow in gentle-ai parser",
    "vulnerabilities": [
      {
        "package": {
          "name": "gentle-ai",
          "ecosystem": "NUGET"
        },
        "vulnerable_version_range": "<= 0.4.0",
        "patched_versions": ">= 0.4.1"
      }
    ]
  },
  {
    "ghsa_id": "GHSA-yyyy-yyyy-yyyy",
    "severity": "critical",
    "published_at": "2026-07-15T00:00:00Z",
    "summary": "RCE in gentle-ai auth",
    "vulnerabilities": [
      {
        "package": {
          "name": "gentle-ai",
          "ecosystem": "NUGET"
        },
        "vulnerable_version_range": "<= 0.5.0",
        "patched_versions": ">= 0.5.1"
      }
    ]
  }
]`

func TestParseAdvisories(t *testing.T) {
	ads, err := parseAdvisories([]byte(sampleAdvisoriesJSON))
	if err != nil {
		t.Fatalf("parseAdvisories: %v", err)
	}
	if len(ads) != 2 {
		t.Fatalf("got %d advisories, want 2", len(ads))
	}
	if ads[0].GHSAID != "GHSA-xxxx-xxxx-xxxx" {
		t.Errorf("GHSAID = %q", ads[0].GHSAID)
	}
	if ads[0].Severity != "high" {
		t.Errorf("Severity = %q, want high", ads[0].Severity)
	}
}

func TestParseAdvisoriesEmpty(t *testing.T) {
	ads, err := parseAdvisories([]byte(`[]`))
	if err != nil {
		t.Fatalf("parseAdvisories empty: %v", err)
	}
	if len(ads) != 0 {
		t.Errorf("got %d, want 0", len(ads))
	}
}

func TestParseAdvisoriesInvalidJSON(t *testing.T) {
	_, err := parseAdvisories([]byte(`not-json`))
	if err == nil {
		t.Fatal("esperaba error por JSON inválido")
	}
}

func TestAdvisoryPatchedVersion(t *testing.T) {
	ads, _ := parseAdvisories([]byte(sampleAdvisoriesJSON))
	if got := advisoryPatchedVersion(ads[0]); got != "0.4.1" {
		t.Errorf("advisoryPatchedVersion = %q, want 0.4.1", got)
	}
	if got := advisoryPatchedVersion(ads[1]); got != "0.5.1" {
		t.Errorf("advisoryPatchedVersion = %q, want 0.5.1", got)
	}
}

func TestMarkSecurityReleases(t *testing.T) {
	ads, _ := parseAdvisories([]byte(sampleAdvisoriesJSON))
	releases := []Release{
		{Version: v("0.3.0")},
		{Version: v("0.4.0")},
		{Version: v("0.4.1")}, // security fix
		{Version: v("0.5.0")},
		{Version: v("0.5.1")}, // security fix
		{Version: v("0.6.0")},
	}

	marked := MarkSecurityReleases(releases, ads)
	if len(marked) != len(releases) {
		t.Fatalf("len = %d, want %d", len(marked), len(releases))
	}

	checks := []struct {
		idx  int
		want bool
	}{
		{0, false}, // 0.3.0
		{1, false}, // 0.4.0
		{2, true},  // 0.4.1 → fix for first advisory
		{3, false}, // 0.5.0
		{4, true},  // 0.5.1 → fix for second advisory
		{5, false}, // 0.6.0
	}
	for _, c := range checks {
		if marked[c.idx].Security != c.want {
			t.Errorf("releases[%d].Security = %v, want %v", c.idx, marked[c.idx].Security, c.want)
		}
	}
}

func TestMarkSecurityReleasesEmptyAdvisories(t *testing.T) {
	releases := []Release{{Version: v("0.4.0")}, {Version: v("0.4.1")}}
	marked := MarkSecurityReleases(releases, []ghAdvisory{})
	for i, r := range marked {
		if r.Security {
			t.Errorf("marked[%d].Security = true sin advisories", i)
		}
	}
}

func TestMarkSecurityReleasesNoMatch(t *testing.T) {
	ads, _ := parseAdvisories([]byte(sampleAdvisoriesJSON))
	releases := []Release{{Version: v("0.3.0")}, {Version: v("0.6.0")}}
	marked := MarkSecurityReleases(releases, ads)
	for i, r := range marked {
		if r.Security {
			t.Errorf("marked[%d] = true, esperaba false (no es fix)", i)
		}
	}
}

// TestAdvisoryNotModifiesOriginal asegura que MarkSecurityReleases no muta el slice
// original.
func TestMarkSecurityReleasesImmutability(t *testing.T) {
	ads, _ := parseAdvisories([]byte(sampleAdvisoriesJSON))
	orig := []Release{{Version: v("0.4.0")}, {Version: v("0.4.1")}}
	_ = MarkSecurityReleases(orig, ads)
	if orig[1].Security {
		t.Error("MarkSecurityReleases mutó el slice original")
	}
}

// TestFetchAdvisoriesReal verifica parseo contra datos reales de la API (usa cache
// o red si está disponible). Sin autenticación, el rate-limit anónimo suele ser
// suficiente para esta consulta.
func TestFetchAdvisoriesReal(t *testing.T) {
	if testing.Short() {
		t.Skip("red — short mode")
	}
	t.Log("Consulta real a GitHub Security Advisories API")
	ads, err := FetchAdvisories(t.Context())
	if err != nil {
		t.Fatalf("FetchAdvisories: %v", err)
	}
	t.Logf("Se obtuvieron %d advisories para gentle-ai", len(ads))
	for _, a := range ads {
		pv := advisoryPatchedVersion(a)
		t.Logf("  %s [%s] → %s: %s", a.GHSAID, a.Severity, pv, a.Summary)
	}
}

// helpers
func v(s string) Version {
	v, err := ParseVersion(s)
	if err != nil {
		panic(err)
	}
	return v
}
