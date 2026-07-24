// Package bootstrap detecta el sistema operativo y el gestor de paquetes, y verifica
// los prerequisitos necesarios antes de instalar gentle-ai y la capa de aiwf.
package bootstrap

import (
	"os/exec"
	"runtime"
)

// PackageManager identifica el gestor de paquetes del sistema.
type PackageManager string

const (
	Winget   PackageManager = "winget"
	Scoop    PackageManager = "scoop"
	Homebrew PackageManager = "brew"
	Apt      PackageManager = "apt"
	Dnf      PackageManager = "dnf"
	Pacman   PackageManager = "pacman"
	Unknown  PackageManager = ""
)

// packageManagerCandidates devuelve, por orden de preferencia, los gestores de
// paquetes plausibles para el SO indicado.
func packageManagerCandidates(goos string) []PackageManager {
	switch goos {
	case "windows":
		return []PackageManager{Winget, Scoop}
	case "darwin":
		return []PackageManager{Homebrew}
	default: // linux y otros unix
		return []PackageManager{Apt, Dnf, Pacman, Homebrew}
	}
}

// DetectPackageManager devuelve el primer gestor de paquetes disponible en PATH para
// el SO actual, o Unknown si no encuentra ninguno.
func DetectPackageManager() PackageManager {
	for _, pm := range packageManagerCandidates(runtime.GOOS) {
		if _, err := exec.LookPath(string(pm)); err == nil {
			return pm
		}
	}
	return Unknown
}

// Prereq describe una herramienta requerida y si está presente en el sistema.
type Prereq struct {
	Name     string // nombre legible
	Resolved string // ejecutable encontrado en PATH (vacío si ausente)
	Present  bool
}

// requiredCommands son los prerequisitos obligatorios. Cada uno admite ejecutables
// alternativos (p. ej. python3/python). Go es obligatorio porque gentle-ai se instala
// vía `go install` en Windows.
var requiredCommands = []struct {
	Name     string
	Commands []string
}{
	{"Git", []string{"git"}},
	{"Node.js", []string{"node"}},
	{"Docker", []string{"docker"}},
	{"Python 3", []string{"python3", "python"}},
	{"Go", []string{"go"}},
}

// lookAny devuelve el primer comando de la lista presente en PATH.
func lookAny(commands []string) (string, bool) {
	for _, c := range commands {
		if _, err := exec.LookPath(c); err == nil {
			return c, true
		}
	}
	return "", false
}

// CheckPrereqs verifica cada prerequisito y devuelve su estado, preservando el orden
// de declaración.
func CheckPrereqs() []Prereq {
	out := make([]Prereq, 0, len(requiredCommands))
	for _, c := range requiredCommands {
		resolved, present := lookAny(c.Commands)
		out = append(out, Prereq{Name: c.Name, Resolved: resolved, Present: present})
	}
	return out
}

// Missing devuelve solo los prerequisitos ausentes.
func Missing(ps []Prereq) []Prereq {
	var m []Prereq
	for _, p := range ps {
		if !p.Present {
			m = append(m, p)
		}
	}
	return m
}
