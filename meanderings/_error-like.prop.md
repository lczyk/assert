---
status: implemented
date: 2026-04-29
description: structural error comparison (type + Error()) — surfaced during goruby migration
---

# Proposal: `assert.Error` structural error comparison (ErrorLike)

> Surfaced while migrating goruby's `utils/testing.go` to `assert@v0.3.1`.

## Gap (original)

`compare.Errors` (used by `assert.Error` when `expected` is an `error`) was:

```go
func Errors(err error, target error) bool {
    if err == nil && target == nil { return true }
    return errors.Is(err, target)
}
```

`errors.Is` matches by identity (or unwrapped chain). It does **not** match
two distinct error instances that share the same dynamic type and `Error()`
string.

The old `utils.CompareErrors` did:

```go
reflect.TypeOf(a) == reflect.TypeOf(b) && a.Error() == b.Error()
```

Many goruby tests construct an expected value (e.g. `NewTypeError("...")`)
and compare against a freshly-built error returned by the code under test.
Those two values are never `errors.Is`-equal but **are** structurally
equivalent — the migration broke ~15 assertions in `object/` and `parser/`.

## Options considered

- **(a)** New `compare.ErrorsLike` comparator (reflect-type + `Error()`).
- **(b)** New top-level `assert.ErrorLike(t, err, expected)`.
- **(c)** Have `assert.Error(t, err, errVal)` fall back to type+message
  equality when `errors.Is` returns false.

Original preference: (a) + (b). Behaviour-changing fallback was deemed
"probably not worth it".

## Outcome

Shipped option **(c)** after all. `assert_error` in `assert.go` now does:

```go
if !compare.Errors(err, expected) && !compare.ErrorsIs(err, expected) {
    // fail
}
```

`compare.Errors` matches structurally (type + message) and `compare.ErrorsIs`
covers the wrap-chain case. Behaviour change was small and removed the need
for a parallel `ErrorLike` API. `ErrorIs` exists as a strict-wrap-only
sibling for callers that want the original `errors.Is` semantics.
