---
status: implemented
date: 2026-05-05
description: uniform args ...any across all public assertions; finishes the job started in _equal-cmp-args
---

# Proposal: uniform `args ...any` across public assertions

> Resolves #7 from the godep_differ migration proposal. Sibling to
> [_equal-cmp-args.prop.md](_equal-cmp-args.prop.md), which shipped
> the variadic msg tail on `EqualCmp` / `EqualCmpAny` /
> `EqualArrays` / `EqualMaps` / `EqualArraysUnordered` in v0.5. This
> closes out the four remaining funcs that still lacked the tail.

## Gap

After `_equal-cmp-args` shipped, four public assertions still lacked
`args ...any`:

- `Equal[T comparable](t, a, b T)`
- `NotEqual[T comparable](t, a, b T)`
- `EqualLineByLine(t, a, b string)`
- `ContainsString(t, haystack, needle string)`

Caller wanting custom context ("for case %s") on equality / line /
substring assertions had to drop back to `t.Errorf`. Asymmetry made
the API less predictable -- "does this one take args?" was a per-fn
lookup, not a uniform rule.

Empirical signal across ~85 downstream files using `lczyk/assert`:

- 729 calls across the four funcs.
- ~71+ inside table-test or loop scope (where custom msg is most
  valuable). Conservative lower bound; real number higher.

## Decision

Added `args ...any` to all four. Threading via the existing
`args_to_message` helper. Non-breaking (variadic addition).

## EqualLineByLine: emit caller msg once, not per line

Pre-flip behaviour: per-line mismatch -> one `t.Errorf` per mismatched
line. With args threading naively added, every line would carry a
copy of the caller's msg -- noisy.

Post-flip: collect all line mismatches, emit a single `t.Errorf`
with the joined mismatches and one copy of the caller's msg.

```go
// before
line 2: expected 'b', got 'X' in foo.go:42
line 5: expected 'e', got 'Y' in foo.go:42  // duplicated caller msg too

// after
line 2: expected 'b', got 'X'; line 5: expected 'e', got 'Y' in foo.go:42
```

Side benefit: simpler test runner output for big-string compares.
Format change is mildly breaking for test consumers grepping the
exact pre-flip wording (`"expected line N to be 'X', got 'Y'"` ->
`"line N: expected 'X', got 'Y'"`). Single repo affected (this one);
internal tests updated.

## fail_here removed

Pre-flip, `fail_here(t, N, msg)` was a small helper used only by the
four funcs above (the rest of the package used the inline
`get_parent_info + args_to_message + t.Errorf` pattern). Switching to
the inline pattern in all four sites left `fail_here` with zero
callers. Removed.

## Outcome

Shipped in v0.6.x.

- `interface.go`: signatures of `Equal`, `NotEqual`,
  `EqualLineByLine`, `ContainsString` extended with `args ...any`.
  Bodies switched to the inline `args_to_message + t.Errorf`
  pattern.
- `assert.go`: `fail_here` removed (dead).
- `assert_test.go`: existing failure-msg expectations updated for
  the new `EqualLineByLine` wording. Custom-message subtest added
  to each of the four `Test*` functions to lock in args threading.
  New `multiple mismatches reported once` subtest covers the
  consolidated-failure path.

## What this does NOT do

- No `Equalf`-style separate variants. Variadic covers the same
  ground at zero extra surface.
- No retroactive deprecation -- additive change, no caller breaks.
- Does not touch `Type[T]` (already had args), `That` (already had
  args), nor any other variadic-bearing fn.

## Reads-as-prose check

`assert.Equal(t, got, want, "for case %s", name)` ->
"assert equal got want for case <name>". Reads. Same shape across
all four.
