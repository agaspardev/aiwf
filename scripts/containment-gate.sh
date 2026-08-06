#!/usr/bin/env sh
# containment-gate.sh — enforcement determinista de la regla de contención (F2):
# los artefactos de trabajo (scratch/notes/evidence/reports/screenshots/coverage/
# handoffs, coverage.out, *.cov, *.log) deben vivir bajo .ai-workflow/.
# Falla si hay artefactos fuera. Respeta scripts/containment-allow.txt.
#
# Usage: scripts/containment-gate.sh
set -eu
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
exec go run ./cmd/aiwf gate --containment
