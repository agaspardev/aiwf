#!/usr/bin/env sh
# coverage-gate.sh — deterministic per-package coverage floor gate.
#
# WHY this exists (not a filename check): "does foo_test.go exist" is a weak,
# falsifiable proxy for "is it tested". The truth is coverage, measured by the
# tool. This gate reads scripts/coverage-floors.tsv and fails if any package
# drops below its recorded floor — a ratchet that prevents regression without
# blocking correct code.
#
# Usage: scripts/coverage-gate.sh
# Exit:  0 = all packages meet their floor; 1 = at least one regressed.
set -eu

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
FLOORS="$ROOT/scripts/coverage-floors.tsv"
MODULE="$(cd "$ROOT" && go list -m)"
# Default floor for any package not listed in the tsv: new code must be tested.
DEFAULT_FLOOR=50

if [ ! -f "$FLOORS" ]; then
  echo "coverage-gate: missing $FLOORS" >&2
  exit 1
fi

floor_for() {
  # $1 = package path relative to module (e.g. internal/omniroute)
  awk -F '\t' -v pkg="$1" '
    /^[[:space:]]*#/ { next }
    NF < 2 { next }
    $1 == pkg { print $2; found=1; exit }
    END { if (!found) print "" }
  ' "$FLOORS"
}

echo "coverage-gate: running go test -cover across module..."
# Capture per-package coverage. Non-zero test failures propagate via set -e.
OUT="$(cd "$ROOT" && go test -cover ./... 2>&1)"
echo "$OUT" | grep -E "coverage:|no test files" || true

FAILED=0
# Parse lines like: ok  github.com/x/aiwf/internal/omniroute  1.1s  coverage: 45.5% of statements
echo "$OUT" | grep "coverage:" | while IFS= read -r line; do
  pkgfull=$(echo "$line" | awk '{print $2}')
  pkg=${pkgfull#"$MODULE/"}
  cov=$(echo "$line" | sed -n 's/.*coverage: \([0-9.]*\)%.*/\1/p')
  [ -z "$cov" ] && continue
  floor=$(floor_for "$pkg")
  [ -z "$floor" ] && floor=$DEFAULT_FLOOR
  # integer compare on the floored coverage value
  covint=${cov%.*}
  if [ "$covint" -lt "$floor" ]; then
    echo "  REGRESSION: $pkg at ${cov}% < floor ${floor}%" >&2
    echo "fail" >> "$ROOT/.coverage-gate.fail"
  fi
done

if [ -f "$ROOT/.coverage-gate.fail" ]; then
  rm -f "$ROOT/.coverage-gate.fail"
  echo "coverage-gate: FAILED — raise coverage or (if intentional) adjust the floor with justification." >&2
  exit 1
fi
echo "coverage-gate: PASS — every package meets its floor."
