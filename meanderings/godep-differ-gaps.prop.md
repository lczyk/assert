---
status: open
date: 2026-05-05
description: gaps surfaced during godep_differ migration (~46 files, ~27k lines)
---

# Proposal: gaps surfaced during godep_differ migration

> External report from a downstream user migrating ~46 test files / ~27k
> lines off `t.Errorf` patterns onto `assert@v0.5.0`. Bundled here as a
> single meandering; sub-items split into dedicated proposals if/when
> they grow legs.

Items already covered elsewhere are listed for completeness with a
pointer; the rest are spelled out below.

## Already on the pile

- **`Must*` / hard-fail variants** -- see [_require-submodule.prop.md](_require-submodule.prop.md)
  (which superseded the earlier `must-variants` note).
  External report reinforces the pain (~200+ `if err != nil { t.FailNow() }`
  sites across one repo) and explicitly calls out `Type[T]` as the
  primitive that almost always wants Fatal semantics, since the returned
  zero value is a deref footgun. New angle: naming. `Must*` doesn't
  read like prose (`assert.MustNoError` -> "assert must no error").
  A separate `require` sub-package (`require.NoError(t, err)` -> "require
  no error") fits the prose-reading design tenet better. Open question
  on top of the existing one: package-split vs prefix-on-same-package.
- **`DeepEqual`** -- see [deep-equal.prop.md](deep-equal.prop.md). External
  report adds a data point: in their codebase `EqualCmp(reflect.DeepEqual)`
  appears dozens of times. Strong evidence for the "single primitive
  collapses the typed variants" framing.
- **diff output for big-string compares** -- see [cmp-diff-output.prop.md](cmp-diff-output.prop.md).
  Same itch.
- **uniform `args ...any`** -- partially done (`EqualCmp` / `EqualCmpAny`
  in [_equal-cmp-args.prop.md](_equal-cmp-args.prop.md)). External report
  flags the remaining inconsistency: `Equal`, `NotEqual`, `EqualArrays`,
  `EqualMaps`, `EqualLineByLine`, `ContainsString` still lack the
  variadic message tail. Mechanical fix; no design call beyond "finish
  the job".

## New gaps

### 1. `Error` regex-vs-substring foot-gun

`assert.Error(t, err, "some pattern")` treats the string as a regex.
Metacharacters (`.()?+`) silently mismatch or panic.
[_error-string-regex-docs.prop.md](_error-string-regex-docs.prop.md)
documented this; downstream still tripped over it twice during one
migration. Documentation is not a load-bearing fix here -- the call
site looks like substring match in every other test framework.

Options:

- **(a)** split into two named primitives: `assert.ErrorContains(t, err, sub)`
  (literal `strings.Contains`) and `assert.ErrorMatches(t, err, pattern)`
  (explicit regex; accepts `string` or `*regexp.Regexp`). Drop the
  implicit-string-as-regex branch from `Error`. Breaking.
- **(b)** flip the default: `Error(t, err, "...")` becomes literal
  substring; regex stays via `*regexp.Regexp`. Less new API but still
  breaking, and silently changes behaviour for existing callers.
- **(c)** leave `Error` alone, add only `ErrorMatches` as the explicit
  regex form, recommend it in docs. Non-breaking but the foot-gun
  stays.

Tradeoffs:

- (a) is the cleanest mental model -- name signals semantics. Doubles
  the error-shape API count (`Error`, `ErrorIs`, `ErrorContains`,
  `ErrorMatches`) but each is unambiguous.
- (b) is the smallest API but the silent-behaviour-change is the worst
  failure mode for a test library: tests that previously passed for the
  wrong reason now keep passing for a different wrong reason.
- (c) preserves source compat but doesn't move callers off the
  foot-gun.

Reads-as-prose check: `assert.ErrorContains(t, err, s)` -> "assert
error contains s", `assert.ErrorMatches(t, err, p)` -> "assert error
matches p". Both fit.

Open questions:

- If (a), do existing callers passing literal strings to `Error` get a
  deprecation cycle, or a hard break at v0.6?
- Does `Error` keep the structural-match (`error` arg) and `AnyError`
  branches, dropping only the string branch?

### 2. ordering primitives (`Greater` / `Less`)

Threshold checks fall back to `That`:

```go
assert.That(t, score >= 0.4, "score below threshold: got %v", score)
assert.That(t, count > 4, "...")
```

`That` works but loses the structured "got %v, want > %v" message and
forces hand-formatting at every call site.

Options:

- **(a)** add `Greater` / `Less` only; skip `GreaterOrEqual` / `LessOrEqual`
  / `Between`. Caller writes `Greater(t, x, n-1)` for `>=` cases.
- **(b)** full set: `Greater`, `GreaterOrEqual`, `Less`, `LessOrEqual`,
  `Between`. Generic over `cmp.Ordered`.
- **(c)** none -- `That` is the boolean primitive by design; ordering
  is just a boolean predicate.

Tradeoffs:

- (a) two funcs, covers the structured-message win for the common
  cases. Edge: `>=` -> `Greater(t, x, n-1)` only works on integer
  types, not floats / durations.
- (b) five funcs. Bigger surface, but `cmp.Ordered` covers it
  uniformly. Matches testify shape.
- (c) holds the line on size. Loses the structured message.

Reads-as-prose check: `assert.Greater(t, x, y)` -> "assert greater x y"
reads sideways. `assert.GreaterThan(t, x, y)` -> "assert greater than
x y" reads. Naming matters more here than for most additions.

Open question: is the structured-message win worth 2-5 new funcs, or
does `That` + a one-line message stay the answer?

### 3. `True` / `False`

Trivial wrappers over `That`. Muscle-memory ask from other libs.

Tension with design tenet: `assert.That(t, ok)` -> "assert that ok"
already reads naturally. `assert.True(t, ok)` -> "assert true ok"
reads worse. README pitches `That` as the boolean primitive.

Open question: are there call sites where `True` / `False` genuinely
read better than `That`? If not, this is pure surface bloat; if yes,
worth listing them.

### 4. slice / map containment

Repeating shapes:

```go
assert.That(t, slices.Contains(s, v))
_, ok := m[k]; assert.That(t, ok)
```

Proposed: `assert.Contains[T comparable](t, []T, T)`,
`assert.HasKey[K comparable, V any](t, map[K]V, K)`,
`assert.MapContains(t, m, k, v)` for single-pair check.

Reads-as-prose check: `assert.HasKey(t, m, k)` -> "assert has key m k"
reads natural. `assert.Contains(t, s, v)` -> "assert contains s v"
borderline; `Includes` reads slightly better but collides with no
existing convention.

Tradeoffs:

- two-three new funcs vs one-line `That` + stdlib (`slices.Contains`,
  map `_, ok :=`).
- structured failure message ("expected key %v to be in map, got keys
  %v") is the win, same shape as ordering primitives. Without that,
  the wrappers are pure renaming.

Open question: is the structured-message + prose-fit combo enough to
justify, or wait for evidence of >Nx repetition in real downstream
code?

### 5. float-tolerance (`NearlyEqual` / `Within`)

```go
assert.That(t, math.Abs(got-want) < 1e-6, "...")
```

Proposed: `assert.NearlyEqual(t, got, want, tolerance)` over
`cmp.Ordered` numerics (covers floats and ints/durations).

Open questions:

- Frequency in real code -- one external repo had a handful of
  sites. Niche or genuine?
- API shape: `NearlyEqual(got, want, eps)` vs `Within(got, want, delta)`
  vs `Approximately(got, want, eps)`. Reads-as-prose: "assert nearly
  equal got want eps" works; "assert within got want delta" works too.
- Does this bundle with ordering primitives (#2) under a single
  numeric-comparison decision, or stay separate?

### 6. `EqualArraysUnordered` for non-comparable types (`ElementsMatch`)

Current `EqualArraysUnordered` requires `comparable` -- works for
`[]string` / `[]int`, doesn't work for `[]struct{...}` containing
slices/maps.

Options:

- **(a)** new `assert.ElementsMatch(t, a, b)`, reflect-based, any type.
  Mirrors testify naming.
- **(b)** make `EqualArraysUnordered` reflect-based and drop the
  `comparable` constraint. Slower for the common comparable case.
- **(c)** subsumed by `DeepEqual` (#deep-equal.prop.md) -- no, different
  semantics (unordered vs ordered).

Open question: real demand, or theoretical?

### 7. `Eventually` / `WithinDuration` (timing)

`assert.Eventually(t, predicate func() bool, timeout, interval)` polls
until the predicate holds or the timeout elapses. Cleans up "wait for
state" loops in integration tests.

Tension with design tenet: this package is a thin wrapper over
`testing.TB`. `Eventually` introduces time / polling -- a different
shape than every other primitive. Plausibly belongs in a separate
package (`assert/eventually` or downstream).

Open question: is timing-aware assertion in scope at all, or is the
right answer "use a test helper in your repo, not this package"?

## Cross-cutting open question

What's the size budget? Current public surface is ~15 funcs. The
external report's "high-impact" set alone (Must family, DeepEqual,
ordering, uniform args, Error split) lands ~5-10 new funcs depending
on shape. The "nice-to-have" set adds another ~5-8. Worth being
explicit about a target before merging individual proposals -- e.g.
"stay under 25 public funcs" forces tradeoff calls instead of
incremental drift.
