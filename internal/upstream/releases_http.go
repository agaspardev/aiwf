package upstream

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// releasesURL es el endpoint de releases de gentle-ai en la API de GitHub.
const releasesURL = "https://api.github.com/repos/Gentleman-Programming/gentle-ai/releases"

// ghRelease es el subconjunto de campos que consumimos de la API de GitHub.
type ghRelease struct {
	TagName     string    `json:"tag_name"`
	PublishedAt time.Time `json:"published_at"`
	Prerelease  bool      `json:"prerelease"`
	Draft       bool      `json:"draft"`
}

// parseReleases convierte el JSON de la API en []Release, descartando drafts,
// prereleases y tags sin semver. Función pura (testeable sin red).
//
// Para releases con Security=true, usar FetchReleasesWithSecurity (combina con
// FetchAdvisories + MarkSecurityReleases).
func parseReleases(data []byte) ([]Release, error) {
	var raw []ghRelease
	if err := json.Unmarshal(data, &raw); err != nil {
		return nil, err
	}
	out := make([]Release, 0, len(raw))
	for _, r := range raw {
		if r.Draft || r.Prerelease {
			continue
		}
		v, err := ParseVersion(r.TagName)
		if err != nil {
			continue
		}
		out = append(out, Release{Version: v, PublishedAt: r.PublishedAt})
	}
	return out, nil
}

// FetchReleases obtiene las releases de gentle-ai desde GitHub.
func FetchReleases(ctx context.Context) ([]Release, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, releasesURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github releases: status %d", resp.StatusCode)
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 5<<20))
	if err != nil {
		return nil, err
	}
	return parseReleases(body)
}
