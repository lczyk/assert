---
status: implemented
date: 2026-05-05
description: NearlyEqual for float-tolerance and integer-tolerance comparison
---

# Proposal: NearlyEqual

> Resolves #8 from the godep_differ migration proposal. Score / metric
> tests with epsilon comparisons fell back to
> `assert.That(t, math.Abs(got-want) < eps, "...")` -- one-line but
> sparse failure message and hand-formatted msg every site.

## Gap

```go
assert.That(t, math.Abs(got-want) < 1e-6, "score below threshold: got %v", got)
```

Loses structured "got %v want %v tolerance %v diff %v" diagnostic.
Caller has to construct the message every time. Naming the assertion
makes intent explicit at the call site.

## Decision

Shipped `NearlyEqual[T numeric](t, got, want, tolerance, args...)`.

```go
type numeric interface {
    ~int | ~int8 | ~int16 | ~int32 | ~int64 |
    ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
    ~float32 | ~float64
}

func NearlyEqual[T numeric](t testing.TB, got, want, tolerance T, args ...any)
```

Predicate: `|got - want| <= tolerance`. Symmetric branching for the
abs to keep unsigned types safe (no underflow on `got - want` when
`want > got`).

## Why this constraint

- `cmp.Ordered` is too wide -- includes `string`, which doesn't
  support subtraction. Compile would fail or surface confusing
  errors.
- `golang.org/x/exp/constraints.{Integer,Float}` would add a
  dependency on `x/exp` for one constraint definition. Inline union
  is cheaper.
- `~` on each prefix lets named types based on these (e.g. custom
  `type Score float64`, `time.Duration` which is `int64`) pass
  through. `time.Duration` example:
  ```go
  assert.NearlyEqual(t, elapsed, 100*time.Millisecond, 5*time.Millisecond)
  ```

## Reads-as-prose check

`assert.NearlyEqual(t, got, want, tol)` -> "assert nearly equal got
want tol". Reads. Alternative `Within(t, got, want, delta)` would also
read; chose `NearlyEqual` per proposal #8 wording -- "nearly" is more
informal and fits the package's pytest-ish positioning better than
the more formal "within".

## Edge cases

- **Zero tolerance:** degenerates to exact equality. Useful for
  parametrised tests where tolerance varies and may legitimately be
  zero. Passes iff `got == want`.
- **Negative tolerance:** never passes (since `diff` is always
  non-negative for the symmetric-branch abs, `diff <= negative` is
  always false). Caller's bug; surfaced as a test failure rather than
  a panic.
- **NaN:** every comparison and arithmetic op involving NaN is NaN /
  false. `NaN <= tolerance` is false, so NaN inputs always fail.
  Matches the IEEE intuition ("NaN is not nearly equal to anything,
  including itself").
- **Inf:** `+Inf` vs finite -> diff is `+Inf`, fails. `+Inf` vs
  `+Inf` -> `Inf - Inf = NaN`, fails. Arguable semantics; chose
  arithmetic-driven over special-cased. Document.

## What this does NOT do

- No `Within(t1, t2, delta)` for `time.Time` deltas. Different shape
  (subtraction returns `time.Duration`, not the input type). If
  needed later, ship as a separate primitive.
- No `Approximately`-style relative tolerance (`|got-want| / |want|
  <= eps`). Absolute tolerance only. Proposal #8 didn't specify.
  Add later iff demand emerges.
- Not exposed as `assert.Numeric` constraint -- the `numeric`
  interface is unexported. External callers don't need to use it as
  a bound; they only need to satisfy it.

## Outcome

Shipped in v0.6.x.

- `interface.go`: `numeric` constraint and `NearlyEqual` added.
- `assert_test.go`: `TestNearlyEqual` covering float/integer/unsigned
  paths, zero-tolerance, custom-message threading, exact-tolerance
  boundary.
