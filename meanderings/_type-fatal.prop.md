---
status: superseded
date: 2026-05-05
description: Type[T] fails fatally (t.Fatalf) on mismatch -- the returned zero value is a nil-deref footgun
---

# Proposal: Type[T] Fatal-by-default

> **Superseded by [`_require-submodule.prop.md`](_require-submodule.prop.md)
> (2026-05-20).** The Fatal asymmetry described below was reverted when the
> `require` sub-package shipped: `assert.Type` is now soft (uniform with
> the rest of `assert.*`), and `require.Type` covers the hard-fail
> use case described here.
>
> Sub-item of #10 in the godep_differ migration proposal. Carved out as
> a focused decision because Type[T]'s shape -- assertion that returns a
> value -- makes soft-fail genuinely unsafe, not just inconvenient.

## Gap

```go
client := assert.Type[*GeminiClient](t, c)  // returns *GeminiClient
client.DoThing()                            // nil-deref if assert failed
```

Pre-flip, `Type[T]` called `t.Errorf` (soft fail). Caller code on the
next line ran with `client == nil` (zero value of `*T`). Asserting
then dereferencing was a footgun on every call site.

Unlike other assertions (Equal, NoError, etc.) that return nothing and
where soft-fail just means "test continues, more failures may pile
up", `Type[T]` returns a value the caller is *expected* to use. The
soft-fail contract is incompatible with the function's shape.

## Decision

Switched `Type[T]` to `t.Fatalf` on mismatch. Hard-fail by default.
Caller never observes the zero value -- test goroutine ends via
`runtime.Goexit` before the next line runs.

## Why this asymmetry is fine

The package's design tenet ("soft fail; let the test continue") was
written for assertions that don't influence subsequent code paths.
`Type[T]` does. Specifically:

- assertions that return `()`/`void`-shaped: soft-fail = caller's next
  line runs with no contamination.
- assertions that return a value: soft-fail = caller's next line runs
  with a poisoned value. Whatever the test does next is meaningless,
  and the test's later assertions about that value are noise.

`Type[T]` is the only public assertion in the current API that
returns a value, so this is the only asymmetry needed.

## Migration cost

Real downstream usage of `Type[T]`: small. Failing call sites all
followed the assert-then-deref pattern, which means callers were
already implicitly relying on hard-fail semantics (their next line
would crash on the zero value if the assertion didn't stop the test).
The flip moves them from "test crashes with nil-deref panic" to "test
fails cleanly via t.Fatalf" -- strict improvement.

Tests that mock `testing.TB` (the package's own test suite) needed a
`Fatalf` override on the mock that doesn't invoke `runtime.Goexit`,
so post-Fatalf assertions can still inspect the captured failure
message.

## Outcome

Shipped in v0.6.x.

- `assert.go`: `assert_type` switches `t.Errorf` to `t.Fatalf`. The
  unreachable `return *new(T)` after Fatalf stays for compiler
  satisfaction.
- `interface.go`: `Type[T]` doc updated to call out the asymmetry
  (Fatal-by-default, unlike other assertions in the package).
- `assert_test.go`: `myT.Fatalf` override added (captures message,
  marks Fail, no Goexit). `TestType` switched from bare `testing.T{}`
  to `myT` so the failing subtest can still assert post-Fatalf state.
- `example_test.go`: `ExampleType` dropped its failing-case
  demonstration (godoc examples are linear; `Goexit` mid-example
  breaks the runner). Comment in place explaining why.
- Demo: `TestDemoType` unchanged at the call site -- runs against
  real `*testing.T`, the Fatalf path renders correctly under the
  demo runner.

## What this does NOT do

- Does not introduce a `Must*` family or `require` sub-package. `Type[T]`
  was a one-off carve-out; the broader question (#1 in the godep_differ
  proposal) was later resolved by shipping the `require` sub-package
  -- see [`_require-submodule.prop.md`](_require-submodule.prop.md),
  which also reverts the Fatal asymmetry described in this proposal.
- Does not provide a soft-fail variant of `Type[T]`. If someone needs
  one (e.g. probing optional types), they can write `_, ok := obj.(T)`
  directly -- which is what they'd do today anyway.
