package upstream

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ─── Tipos de la API ──────────────────────────────────────────────────────────

// advisoriesURL es el endpoint de security advisories de gentle-ai.
const advisoriesURL = "https://api.github.com/repos/Gentleman-Programming/gentle-ai/security-advisories"

// ghAdvisory es el subconjunto de campos que consumimos del REST API.
type ghAdvisory struct {
	GHSAID      string    `json:"ghsa_id"`
	Severity    string    `json:"severity"`
	PublishedAt time.Time `json:"published_at"`
	Summary     string    `json:"summary"`
	Vulns       []ghVuln  `json:"vulnerabilities"`
}

// ghVuln describe el paquete afectado y las versiones.
type ghVuln struct {
	Package             ghPkg  `json:"package"`
	VulnerableVersionRange string `json:"vulnerable_version_range"`
	PatchedVersions     string `json:"patched_versions"`
}

type ghPkg struct {
	Name      string `json:"name"`
	Ecosystem string `json:"ecosystem"`
}

// ─── Parsing ──────────────────────────────────────────────────────────────────

// parseAdvisories convierte el JSON de la API REST en []ghAdvisory.
func parseAdvisories(data []byte) ([]ghAdvisory, error) {
	var raw []ghAdvisory
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, fmt.Errorf("parseando advisories: %w", err)
	}
	// Filtramos advisories que no son para el paquete gentle-ai.
	var out []ghAdvisory
	for _, a := range raw {
		hasGentleAI := false
		for _, v := range a.Vulns {
			if strings.EqualFold(v.Package.Name, "gentle-ai") ||
				strings.Contains(v.Package.Name, "gentle") {
				hasGentleAI = true
				break
			}
		}
		if hasGentleAI || len(a.Vulns) == 0 {
			out = append(out, a)
		}
	}
	return out, nil
}

// advisoryPatchedVersion extrae la versión ">= X.Y.Z" de patched_versions.
// Devuelve "" si no se puede extraer.
func advisoryPatchedVersion(av ghAdvisory) string {
	for _, v := range av.Vulns {
		pv := strings.TrimSpace(v.PatchedVersions)
		// Formatos típicos: ">= 0.4.1", ">=0.4.1", "0.4.1"
		pv = strings.TrimPrefix(pv, ">=")
		pv = strings.TrimSpace(pv)
		if pv != "" {
			return pv
		}
	}
	return ""
}

// ─── Integración con releases ─────────────────────────────────────────────────

// MarkSecurityReleases toma releases y una lista de advisories y marca como
// Security=true aquellas releases cuya versión coincide con un advisory fix.
// Devuelve una copia de releases con security marcado.
func MarkSecurityReleases(releases []Release, advisories []ghAdvisory) []Release {
	if len(advisories) == 0 {
		return releases
	}
	// Construir conjunto de versiones con advisory.
	patched := make(map[string]bool)
	for _, a := range advisories {
		if v := advisoryPatchedVersion(a); v != "" {
			patched[v] = true
		}
	}
	if len(patched) == 0 {
		return releases
	}

	out := make([]Release, len(releases))
	for i, r := range releases {
		vs := r.Version.String()
		if patched[vs] {
			r.Security = true
		}
		out[i] = r
	}
	return out
}

// ─── HTTP ────────────────────────────────────────────────────────────────────

// FetchAdvisories obtiene los security advisories de gentle-ai desde la API de
// GitHub. Requiere autenticación (token) para mayor rate-limit; sin token funciona
// pero con límite bajo. Devuelve la lista de advisories parseados.
func FetchAdvisories(ctx context.Context) ([]ghAdvisory, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, advisoriesURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("github advisories: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github advisories: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2<<20))
	if err != nil {
		return nil, err
	}
	return parseAdvisories(body)
}

// FetchReleasesWithSecurity obtiene releases desde GitHub y los enrich con datos
// de security advisories. Es un wrapper que llama FetchReleases + FetchAdvisories
// + MarkSecurityReleases. Si FetchAdvisories falla, devuelve los releases sin
// marcar (degradación graceful).
func FetchReleasesWithSecurity(ctx context.Context) ([]Release, error) {
	releases, err := FetchReleases(ctx)
	if err != nil {
		return nil, err
	}
	advisories, advErr := FetchAdvisories(ctx)
	if advErr != nil {
		// Degradación graceful: advisories no disponibles → releases sin marcar.
		return releases, nil
	}
	return MarkSecurityReleases(releases, advisories), nil
}
