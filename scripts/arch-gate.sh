#!/usr/bin/env sh
# arch-gate.sh — enforcement determinista de arquitectura limpia (clean arch) con
# solo el toolchain Go, sin librerías. Verifica la DIRECCIÓN de dependencias: el
# núcleo puro no puede tocar filesystem/red/procesos.
#
# Invariante F0: internal/workspace debe ser I/O-free (resolución de rutas/IDs/
# contexto, sin I/O). Este gate lo hace cumplir; el prompt ya no alcanza.
#
# Usage: scripts/arch-gate.sh
set -eu
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

# Paquetes que DEBEN ser puros (sin I/O directo).
PURE_PKGS="./internal/workspace/"
FORBIDDEN='^(os|os/exec|net|net/http|io/ioutil|bufio)$'

fail=0
for pkg in $PURE_PKGS; do
  bad=$(go list -f '{{ range .Imports }}{{ . }}{{ "\n" }}{{ end }}' "$pkg" 2>/dev/null | grep -E "$FORBIDDEN" || true)
  if [ -n "$bad" ]; then
    echo "arch-gate: $pkg importa I/O prohibido (debe ser núcleo puro):" >&2
    printf '  %s\n' "$bad" >&2
    fail=1
  fi
done

if [ "$fail" -ne 0 ]; then
  echo "arch-gate: FAILED — el núcleo puro no puede depender de filesystem/red/procesos." >&2
  exit 1
fi
echo "arch-gate: PASS — núcleos puros libres de I/O (dirección de dependencias correcta)."
