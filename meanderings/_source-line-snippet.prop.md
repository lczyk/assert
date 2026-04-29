---
status: implemented
date: 2026-04-29
description: print source line at `file:line` in failure output
---

# Proposal: source-line snippet in failure output

## Gap

Rust's `assert2` / `googletest` print the failing expression source:
`assert!(x > 5)` → `assertion failed: x > 5`. Go can't get expression
text from `runtime.Caller`, but the source file is readable at
`file:line` — so print the calling line.

Big UX win for `That(t, complicated && expr)` which previously only
printed `assertion failed`.

## Outcome

Implemented in [source.go](../source.go). `locStr(file, line)` reads
the source file (cached in `sourceCache`) and returns
`file:line\n  > <source>`. For multi-line calls, reads continuation
lines until paren/brace/bracket depth returns to zero.

Failure modes (file unreadable, line out of range, etc.) return an
error and the caller falls back to the old `file:line`-only format.

See `make demo` for side-by-side examples vs vanilla `t.Errorf`.
