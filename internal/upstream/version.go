package upstream

import (
	"fmt"
	"regexp"
	"strconv"
)

// Version es un semver simplificado (major.minor.patch). Suficiente para comparar
// versiones de gentle-ai; ignora pre-release y build metadata.
type Version struct {
	Major, Minor, Patch int
}

// semverRe extrae el primer patrón X.Y.Z de una cadena (tolera prefijo "v" y texto
// alrededor, p. ej. "gentle-ai version v1.4.2").
var semverRe = regexp.MustCompile(`(\d+)\.(\d+)\.(\d+)`)

// ParseVersion extrae una Version del primer semver encontrado en s.
func ParseVersion(s string) (Version, error) {
	m := semverRe.FindStringSubmatch(s)
	if m == nil {
		return Version{}, fmt.Errorf("no se encontró semver en %q", s)
	}
	// Los grupos ya son \d+, atoi no puede fallar salvo overflow; lo ignoramos.
	major, _ := strconv.Atoi(m[1])
	minor, _ := strconv.Atoi(m[2])
	patch, _ := strconv.Atoi(m[3])
	return Version{Major: major, Minor: minor, Patch: patch}, nil
}

// Compare devuelve -1 si v < o, 0 si son iguales, +1 si v > o.
func (v Version) Compare(o Version) int {
	switch {
	case v.Major != o.Major:
		return sign(v.Major - o.Major)
	case v.Minor != o.Minor:
		return sign(v.Minor - o.Minor)
	case v.Patch != o.Patch:
		return sign(v.Patch - o.Patch)
	default:
		return 0
	}
}

// Less indica si v es estrictamente menor que o.
func (v Version) Less(o Version) bool { return v.Compare(o) < 0 }

func (v Version) String() string {
	return fmt.Sprintf("%d.%d.%d", v.Major, v.Minor, v.Patch)
}

func sign(n int) int {
	if n < 0 {
		return -1
	}
	return 1
}
