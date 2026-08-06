package report

import "regexp"

// secretLike detecta tokens que parecen secretos/altas-entropía para redactarlos.
// Cubre claves tipo AWS (AKIA...), tokens largos alfanuméricos y valores base64-ish.
var secretLike = regexp.MustCompile(`\b(AKIA[0-9A-Z]{12,}|gh[pousr]_[A-Za-z0-9]{20,}|xox[baprs]-[A-Za-z0-9-]{10,}|[A-Za-z0-9+/_-]{24,})\b`)

// redact reemplaza cualquier substring con pinta de secreto por [REDACTED].
// Se aplica SIEMPRE al mapear findings de categoría secret: aiwf no confía en el
// --redact de la herramienta y nunca emite el match crudo.
func redact(message string) string {
	return secretLike.ReplaceAllString(message, "[REDACTED]")
}
