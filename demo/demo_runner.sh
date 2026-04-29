#!/usr/bin/env bash
# Run each demo test individually. Demos intentionally fail to show
# off assert's failure output (message + file:line + source snippet).
# Exits 0 regardless of test failures — the failures ARE the demo.

set -u
cd "$(dirname "$0")"

if command -v gotest >/dev/null 2>&1; then
	GOTEST=gotest
else
	GOTEST="go test"
fi

demos=($(grep -Eo '^func (TestDemo[A-Za-z0-9_]+)' demos_test.go | awk '{print $2}'))

for d in "${demos[@]}"; do
	printf '\n=== %s ===\n' "$d"
	$GOTEST -tags demo -run "^${d}\$" . || true
done

