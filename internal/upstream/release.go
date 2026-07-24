package upstream

import "time"

// ProbationWindow es la ventana prudente durante la cual NO se adopta automáticamente
// una versión recién publicada (da tiempo a la comunidad a detectar paquetes
// comprometidos). Los parches de seguridad confirmados se exceptúan.
const ProbationWindow = 14 * 24 * time.Hour

// Release representa una publicación de gentle-ai (upstream).
type Release struct {
	Version     Version
	PublishedAt time.Time
	// Security indica si es un parche de seguridad confirmado (CVE/advisory).
	// Cuando es true, se salta la ProbationWindow.
	Security bool
}

// EligibleForAdoption indica si r puede adoptarse en el momento now, aplicando la
// ventana prudente: elegible si es parche de seguridad o si ya pasó ProbationWindow
// desde su publicación.
func EligibleForAdoption(r Release, now time.Time) bool {
	if r.Security {
		return true
	}
	return now.Sub(r.PublishedAt) >= ProbationWindow
}

// LatestEligible devuelve la mayor versión elegible para adopción en now.
// El segundo valor es false si ninguna es elegible.
func LatestEligible(rs []Release, now time.Time) (Release, bool) {
	var best Release
	found := false
	for _, r := range rs {
		if !EligibleForAdoption(r, now) {
			continue
		}
		if !found || best.Version.Less(r.Version) {
			best, found = r, true
		}
	}
	return best, found
}
