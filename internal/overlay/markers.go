// Package overlay aplica la capa propia de aiwf sobre una instalación de gentle-ai
// SIN sobrescribir sus archivos: escribe archivos propios (OWNED) e inyecta contenido
// en archivos compartidos (SHARED) mediante bloques marcados o merge de JSON. Todas
// las operaciones son idempotentes para poder re-aplicarse tras un update de gentle-ai
// (reconciliación). Ver design.md §Coexistencia.
package overlay

import "strings"

const (
	markerStart = "# >>> aiwf (managed — no editar) >>>"
	markerEnd   = "# <<< aiwf <<<"
)

// HasBlock indica si content ya contiene un bloque gestionado por aiwf.
func HasBlock(content string) bool {
	s := strings.Index(content, markerStart)
	e := strings.Index(content, markerEnd)
	return s != -1 && e != -1 && e > s
}

// InjectBlock inserta o reemplaza el bloque gestionado con payload. Es idempotente:
// InjectBlock(InjectBlock(c, p), p) == InjectBlock(c, p). Nunca toca el contenido
// fuera del bloque (lo de gentle-ai queda intacto).
func InjectBlock(content, payload string) string {
	block := markerStart + "\n" + payload + "\n" + markerEnd

	s := strings.Index(content, markerStart)
	e := strings.Index(content, markerEnd)
	if s != -1 && e != -1 && e > s {
		return content[:s] + block + content[e+len(markerEnd):]
	}
	if strings.TrimSpace(content) == "" {
		return block + "\n"
	}
	sep := "\n\n"
	if strings.HasSuffix(content, "\n") {
		sep = "\n"
	}
	return content + sep + block + "\n"
}

// RemoveBlock elimina el bloque gestionado (para uninstall/reconcile). Si no hay
// bloque, devuelve content sin cambios.
func RemoveBlock(content string) string {
	s := strings.Index(content, markerStart)
	e := strings.Index(content, markerEnd)
	if s == -1 || e == -1 || e < s {
		return content
	}
	before := strings.TrimRight(content[:s], "\n")
	after := strings.TrimLeft(content[e+len(markerEnd):], "\n")
	switch {
	case before == "":
		return after
	case after == "":
		return before + "\n"
	default:
		return before + "\n" + after
	}
}
