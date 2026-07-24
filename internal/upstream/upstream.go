// Package upstream detecta gentle-ai en el sistema, resuelve su versión y decide si
// hay que instalar, actualizar o saltar (idempotencia). Ver spec.md §Idempotencia.
package upstream

import (
	"context"
	"os/exec"
	"strings"
)

// binaryName es el ejecutable de gentle-ai en PATH.
const binaryName = "gentle-ai"

// Action es la decisión del instalador respecto a gentle-ai.
type Action int

const (
	// Skip: gentle-ai presente y ≥ versión objetivo; no se hace nada.
	Skip Action = iota
	// Install: gentle-ai ausente; hay que instalarlo.
	Install
	// Update: gentle-ai presente pero por debajo de la versión objetivo.
	Update
)

func (a Action) String() string {
	switch a {
	case Install:
		return "install"
	case Update:
		return "update"
	case Skip:
		return "skip"
	default:
		return "unknown"
	}
}

// Decide aplica la matriz de idempotencia contra la versión objetivo (pin).
func Decide(installed bool, current, target Version) Action {
	switch {
	case !installed:
		return Install
	case current.Less(target):
		return Update
	default:
		return Skip
	}
}

// State es el estado detectado de gentle-ai en el sistema.
type State struct {
	Installed  bool
	Path       string
	Version    Version
	RawVersion string // salida cruda de `gentle-ai --version`
}

// commandContext se puede sustituir en tests. Por defecto usa exec.CommandContext.
var commandContext = exec.CommandContext

// Detect busca gentle-ai en PATH y, si está, resuelve su versión ejecutando
// `gentle-ai --version`.
func Detect(ctx context.Context) (State, error) {
	path, err := exec.LookPath(binaryName)
	if err != nil {
		// No presente en PATH: estado no instalado, sin error.
		return State{Installed: false}, nil
	}
	st := State{Installed: true, Path: path}

	out, err := commandContext(ctx, binaryName, "--version").Output()
	if err != nil {
		// Está instalado pero no pudimos leer la versión; lo reportamos como
		// instalado con versión cero para que la decisión tienda a Update.
		return st, err
	}
	st.RawVersion = strings.TrimSpace(string(out))
	if v, perr := ParseVersion(st.RawVersion); perr == nil {
		st.Version = v
	}
	return st, nil
}
