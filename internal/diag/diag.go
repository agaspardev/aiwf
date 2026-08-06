// Package diag implementa el diagnóstico del toolchain operativo (`aiwf check`),
// portado del modo -Doctor de harness.ps1: verifica herramientas con degradación
// graceful (ausente = DEGRADED, no error), salvo las críticas.
package diag

import "os/exec"

// Check es una herramienta a verificar.
type Check struct {
	Name     string
	Command  string
	Critical bool
	Present  bool
}

// DefaultChecks son las herramientas del workflow. Solo Claude Code es crítica; el
// resto degrada de forma graceful si falta.
func DefaultChecks() []Check {
	return []Check{
		{Name: "Claude Code", Command: "claude", Critical: true},
		{Name: "OmniRoute", Command: "omniroute"},
		{Name: "Gitleaks", Command: "gitleaks"},
		{Name: "OSV-Scanner", Command: "osv-scanner"},
		{Name: "Semgrep", Command: "semgrep"},
		{Name: "Trivy", Command: "trivy"},
		{Name: "Docker", Command: "docker"},
		{Name: "Node.js", Command: "node"},
		{Name: "Python", Command: "python"},
		{Name: "Engram", Command: "engram"},
		{Name: "CodeGraph", Command: "codegraph"},
	}
}

// lookPath se puede sustituir en tests.
var lookPath = exec.LookPath

// Run verifica la presencia en PATH de cada check y devuelve una copia con Present set.
func Run(checks []Check) []Check {
	out := make([]Check, len(checks))
	for i, c := range checks {
		_, err := lookPath(c.Command)
		c.Present = err == nil
		out[i] = c
	}
	return out
}

// CriticalMissing devuelve los checks críticos ausentes.
func CriticalMissing(checks []Check) []Check {
	var m []Check
	for _, c := range checks {
		if c.Critical && !c.Present {
			m = append(m, c)
		}
	}
	return m
}
