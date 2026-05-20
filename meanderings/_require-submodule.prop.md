---
status: implemented
date: 2026-05-20
description: require sub-package mirrors assert.* with Fatal semantics
---

# Proposal: `require` sub-package

> Resolves #1 from the godep_differ migration proposal (`Must*`
> hard-fail variants). Supersedes `must-variants.prop.md` -- that
> framing assumed in-place `Must*` siblings; we shipped a sibling
> package instead.

## Gap

`assert.*` only calls `t.Errorf`. Test continues after failure. Every
site that asserts then dereferences needs a manual guard:

```go
assert.NoError(t, err)
if err != nil { t.FailNow() }

assert.NotNil(t, client)
if client == nil { return }
```

In a downstream migration (~46 test files / ~27k lines onto
`lczyk/assert@v0.5.0`) this boilerplate appeared at ~200+ sites,
dwarfing the assertion itself. Same pattern surfaced in a separate
goruby test refactor (~700 sites).

## Shape options considered

### A. `Must*` family in `assert` (rejected)

`assert.MustNoError`, `assert.MustNotNil`, `assert.MustEqual`, ...

Doubles surface in-place. Naming matches Go stdlib (`regexp.MustCompile`,
`template.Must`). Caller mixes hard and soft freely in the same file.

Downside: doubles cognitive load on package surface count; every new
primitive has to ship its `Must*` twin at the same time. Single-flag
inconsistency risk.

### B. Per-call sentinel flag (rejected)

`assert.NoError(t, err, assert.Fatal)` -- flag arg flips failure mode.

Conflicts with the uniform `args ...any` custom-message convention; either
the flag has to be magic-detected from args, or every signature gains a
mode parameter. Both bad.

### C. Sibling sub-package (shipped)

`github.com/lczyk/assert/require` -- mirrors `assert.*` API one-for-one,
calls `t.Fatalf` instead of `t.Errorf`. Matches testify's
`assert` / `require` split, which has wide mental-model transfer in the
Go test ecosystem.

Caller picks per-file (or per-call) which to import. Common pattern:
`require.*` for preconditions, `assert.*` for the behaviour under test.

## Naming

Picked `require` over `must` / `ensure` / `expect` / `verify`:

- `require.NoError(t, err)` reads as "require no error" -- verb-noun,
  natural prose. Fits the "reads like a sentence" tenet.
- `must.NoError(t, err)` reads as "must no error" -- modal-fragment,
  awkward.
- `ensure` reads but has no precedent.
- `expect` has soft-assertion connotation (jest, pytest's `assert`-like
  flow); confusing.
- testify familiarity is a real signal -- callers transferring from
  testify recognise the split immediately.

## Implementation

Extracted all primitive impls into `internal/core/`. Each takes a
`Failer` (a method-value of either `t.Errorf` or `t.Fatalf`). The
public `assert.*` and `require.*` packages are thin wrappers passing
the appropriate one.

```
internal/core/
    core.go        -- Failer, Numeric, AnyError, shared internals
    primitives.go  -- That, Equal, NoError, ...

assert/interface.go   -- thin wrappers, pass t.Errorf
require/require.go    -- thin wrappers, pass t.Fatalf
```

`AnyError` is defined once in `core` and re-exported from both
packages, so `assert.AnyError == require.AnyError` (identity-based
sentinel match works across both).

Wrappers call `t.Helper()` before the core call so that `t.Errorf` /
`t.Fatalf` attributes the failure to the test source line, not to the
wrapper.

## Breaking change: `assert.Type`

Previous behaviour: `assert.Type[T]` called `t.Fatalf` on mismatch (the
sole Fatal-by-default in the package; rationale was that the returned
zero value is a nil-deref footgun if the test keeps running). With the
`require` split, that asymmetry no longer makes sense: callers wanting
the safe-return-on-success behaviour move to `require.Type`.

`assert.Type` is now soft (`t.Errorf`), uniform with the rest of
`assert.*`. The zero-value footgun still exists -- the godoc warns
about it and points at `require.Type`. This is the only behaviour
change for existing `assert.*` callers; flagged as `refactor!:` at
commit time.

## Goroutine hazard

`t.Fatalf` calls `runtime.Goexit`, which only unwinds the goroutine it
runs on. A failed `require.*` call from a background goroutine does
not stop the test and is not reported via the goroutine that ran it
-- the test goroutine sees nothing and may pass silently.

Documented prominently in the `require` package godoc and the README.
The hazard is opt-in: callers who pick `require` accept the contract
(call from the test goroutine).

## What this does NOT do

- No `assert.MustEqual` / `assert.Must*` in-place siblings. Use
  `require.Equal` etc.
- No `assert.TypeOK[T](t, obj) (T, bool)` no-fail probe (goruby
  finding #5). May add later if demand surfaces; soft `assert.Type` +
  zero-value-guard covers the same use case at the call site.
- No `assert.Fatal` sentinel flag.

## Outcome

Shipped. See:

- `internal/core/` -- shared primitives.
- `require/` -- hard-fail wrappers.
- `assert/` -- soft wrappers (rewritten as thin pass-through).

Tests: `require/require_test.go` covers each primitive with happy +
fail paths, verifying via a `myT` mock that the fail path routes to
`Fatalf` (not `Errorf`).
