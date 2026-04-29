---
status: open
date: 2026-04-29
description: unified diff output for `EqualArrays` / `EqualMaps` / `EqualLineByLine`
---

# Proposal: diff output for collection mismatches

## Gap

`EqualLineByLine` currently prints lines one-by-one and stops at first
mismatch. `EqualArrays` / `EqualMaps` print `expected X got Y` with full
slice/map dump — useless for anything beyond ~5 elements.

Better: unified diff with `+` / `-` (color gated on `isatty`). Same
treatment for `EqualArrays` / `EqualMaps` — show a diff, not just
"expected X got Y".

## Options

- `github.com/pmezard/go-difflib` — line diff.
- `github.com/google/go-cmp` `cmp.Diff` — handles structs / maps with
  field-level diff. Could shape an `EqualDiff(t, a, b)`.
- Hand-rolled Myers diff (~100 LOC) for line/element diffs.

## Tradeoffs

- Adding `go-cmp` as a dep crosses a line: today this package has zero
  external deps. Module size: ~4 MB.
- A separate sub-package (`assert/diff`) would keep the core dep-free
  while exposing `cmp.Diff`-style output to callers that opt in.
- Hand-rolled Myers covers line/element diffs without any dep at all,
  but doesn't help with struct field-level output.
