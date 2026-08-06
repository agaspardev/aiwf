#!/usr/bin/env sh
# mutation-gate.sh — mutation testing con gremlins sobre los paquetes de lógica
# donde más importa que los tests atrapen bugs. Verifica la calidad REAL de los
# tests (matan mutantes), la señal que coverage-gate (proxy) no puede dar.
#
# Requiere gremlins en PATH (go install github.com/go-gremlins/gremlins/cmd/gremlins@latest).
# Lento: pensado para el workflow programado, no para pre-commit.
#
# Usage: scripts/mutation-gate.sh
set -eu
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

if ! command -v gremlins >/dev/null 2>&1; then
  echo "mutation-gate: gremlins no está en PATH — instalá con:" >&2
  echo "  go install github.com/go-gremlins/gremlins/cmd/gremlins@latest" >&2
  exit 1
fi

# Paquetes de lógica de alto valor (parsers/agregador/gates). Se puede ampliar.
exec gremlins unleash \
  ./internal/report/ \
  ./internal/govuln/ \
  ./internal/containment/ \
  ./internal/state/
