#!/usr/bin/env sh
# mutation-gate.sh — mutation testing con gremlins sobre los paquetes de lógica
# donde más importa que los tests atrapen bugs. Verifica la calidad REAL de los
# tests (matan mutantes), la señal que coverage-gate (proxy) no puede dar.
#
# Corre gremlins nativo si está en PATH; si no, cae al contenedor golang (Linux)
# vía Docker — útil en Windows/máquinas sin go install. Lento: pensado para el
# workflow programado, no para pre-commit.
#
# Usage: scripts/mutation-gate.sh
set -eu
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"

# Paquetes de lógica de alto valor (parsers/agregador/gates). Se puede ampliar.
PKGS="./internal/report/ ./internal/govuln/ ./internal/containment/ ./internal/state/"

if command -v gremlins >/dev/null 2>&1; then
  exec gremlins unleash $PKGS
fi

if command -v docker >/dev/null 2>&1; then
  # MSYS_NO_PATHCONV=1: Git Bash convierte -v /src → C:/Program Files/Git/src y
  # rompe el mount; con la var deshabilitada el path POSIX pasa intacto.
  export MSYS_NO_PATHCONV=1
  exec docker run --rm \
    -v "$ROOT:/src" \
    -w /src \
    golang:1.25 \
    /bin/sh -c "go install github.com/go-gremlins/gremlins/cmd/gremlins@latest && exec /go/bin/gremlins unleash $PKGS"
fi

echo "mutation-gate: gremlins no hallado (ni nativo ni docker) — instalá con:" >&2
echo "  go install github.com/go-gremlins/gremlins/cmd/gremlins@latest" >&2
exit 1
