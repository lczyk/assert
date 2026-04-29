#!/usr/bin/env bash
# Run demo tests. Demos intentionally fail to show off assert's failure
# output (message + file:line + source snippet). Paired TestVanilla*/
# TestDemo* tests are run back-to-back to compare stdlib testing vs assert.
# Exits 0 regardless of test failures — the failures ARE the demo.

set -u
cd "$(dirname "$0")"

# ANSI colors when stdout is a tty.
if [[ -t 1 ]]; then
	BOLD=$'\e[1m'; DIM=$'\e[2m'; BLU=$'\e[34m'; GRN=$'\e[32m'
	CYN=$'\e[36m'; YLW=$'\e[33m'; RST=$'\e[0m'
else
	BOLD=''; DIM=''; BLU=''; GRN=''; CYN=''; YLW=''; RST=''
fi

# Strip go test trailer noise ("FAIL", "FAIL\tpkg\t...", "ok\tpkg\t...",
# "PASS") and indent the rest for visual grouping.
strip_and_indent() {
	awk -v ind='  ' '
		/^(FAIL|PASS)$/                 { next }
		/^(FAIL|ok)[ \t]+[^[:space:]]/  { next }
		{ printed=1; print ind $0 }
		END { if (!printed) print ind "(passed — no output)" }
	'
}

run_one() {
	local name=$1 color=$2 label=$3
	printf '%s%s%s%s%s\n' "$BOLD" "$color" "$label" " ($name)" "$RST"
	go test -tags demo -count=1 -run "^${name}\$" . 2>&1 | strip_and_indent || true
}

run_one_verbose() {
	local name=$1 color=$2 label=$3
	printf '%s%s%s%s%s\n' "$BOLD" "$color" "$label" " ($name)" "$RST"
	go test -v -tags demo -count=1 -run "^${name}\$" . 2>&1 \
		| awk -v ind='  ' '
			/^(FAIL|PASS)$/                 { next }
			/^(FAIL|ok)[ \t]+[^[:space:]]/  { next }
			{ print ind $0 }
		' || true
}

# Collect suffixes after the TestDemo / TestVanilla prefix from both files.
suffixes=$(
	grep -hEo '^func Test(Demo|Vanilla)[A-Za-z0-9_]+' demos_test.go vanilla_test.go \
		| sed -E 's/^func Test(Demo|Vanilla)//' \
		| sort -u
)

demo_only=()
paired=()
for s in $suffixes; do
	has_vanilla=$(grep -cE "^func TestVanilla${s}\b" vanilla_test.go || true)
	has_demo=$(grep -cE "^func TestDemo${s}\b" demos_test.go || true)
	if [[ $has_vanilla -gt 0 && $has_demo -gt 0 ]]; then
		paired+=("$s")
	elif [[ $has_demo -gt 0 ]]; then
		demo_only+=("$s")
	fi
done

for s in "${paired[@]}"; do
	printf '\n%s════════ %s ════════%s\n\n' "$BOLD" "$s" "$RST"
	run_one "TestVanilla${s}" "$YLW" "VANILLA"
	echo
	run_one "TestDemo${s}"    "$BLU" "ASSERT "
done

# Pull "Passing" out so it always runs last as a positive note.
filtered=()
has_passing=0
for s in "${demo_only[@]}"; do
	if [[ $s == "Passing" ]]; then
		has_passing=1
	else
		filtered+=("$s")
	fi
done

if [[ ${#filtered[@]} -gt 0 ]]; then
	printf '\n%s════════ assert-only demos ════════%s\n' "$BOLD" "$RST"
	for s in "${filtered[@]}"; do
		echo
		run_one "TestDemo${s}" "$BLU" "ASSERT "
	done
fi

if [[ $has_passing -eq 1 ]]; then
	printf '\n%s════════ passing ════════%s\n\n' "$BOLD" "$RST"
	run_one_verbose "TestDemoPassing" "$GRN" "ASSERT "
fi
