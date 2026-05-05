---
status: implemented
date: 2026-05-05
description: Eventually polls predicate until true or timeout
---

# Proposal: Eventually

> Resolves the `Eventually` half of #12 from the godep_differ
> migration proposal. The `WithinDuration(t1, t2, delta)` half not
> shipped -- different shape (time.Time arithmetic returns
> time.Duration) and no demand signal.

## Gap

Integration tests with eventual-consistency reads, async state
changes, or "wait for state" loops fall back to:

```go
deadline := time.Now().Add(2 * time.Second)
for time.Now().Before(deadline) {
    if check() { break }
    time.Sleep(50 * time.Millisecond)
}
assert.That(t, check(), "state never reached")
```

Boilerplate dwarfs the assertion. Naming the assertion makes intent
explicit at the call site.

## Sig

```go
func Eventually(t testing.TB, predicate func() bool, timeout, interval time.Duration, args ...any)
```

Polls `predicate` at `interval` until it returns true or `timeout`
elapses. Predicate is called at least once (even with zero timeout).
Fails the test if predicate never returns true within the deadline.

## Tension with the design tenet

The README pitches the package as "a thin wrapper over Go's standard
testing framework -- no custom runner, no parallel reporting layer,
no DSL". `Eventually` introduces time-driven polling -- a different
shape than every other primitive in the package, all of which are
synchronous one-shot checks.

Counter: `time` is stdlib, not a custom runner. The fn doesn't hide
goroutines, channels, or background state -- it's a synchronous loop
that calls a caller-supplied predicate. Same level of "thin" as the
rest. Worth shipping.

If/when async / channel-driven shapes appear (e.g. select-on-channel
with timeout), reconsider whether they belong here or in a separate
package.

## Predicate-called-at-least-once guarantee

Loop shape:

```go
deadline := time.Now().Add(timeout)
for time.Now().Before(deadline) {
    if predicate() { return }
    time.Sleep(interval)
}
if predicate() { return }
fail
```

The trailing `predicate()` after the loop guarantees one final check
even with `timeout = 0` (where the for-loop body skips entirely).
Useful for parametrised tests where timeout may legitimately be zero
("check this is already true, no waiting").

## What this does NOT do

- No `WithinDuration(t, t1, t2 time.Time, delta time.Duration)` for
  timestamp comparisons. Different shape; if needed later, ship as
  a separate primitive.
- No fancy backoff strategies (linear/exponential). Fixed interval.
  Caller wanting backoff writes their own predicate that sleeps.
- No context.Context support. Caller can wire cancellation through
  the predicate body if needed (`if ctx.Err() != nil { return false
  }`). Adding a ctx param would clutter the most common case.
- No retry counter / "succeeded after N polls" reporting. Failure
  msg shows timeout + interval; that's enough for debugging.

## Reads-as-prose check

`assert.Eventually(t, ready, 2*time.Second, 50*time.Millisecond)` ->
"assert eventually ready 2-second 50-millisecond". Reads.

## Outcome

Shipped in v0.6.x.

- `interface.go`: `time` import added; `Eventually` fn appended at
  end of file.
- `assert_test.go`: `time` import added; `TestEventually` covering
  immediate-true / delayed-true / timeout-fail / zero-timeout (both
  branches) / called-at-least-once guarantee / custom-message
  threading. Short timeouts (10-200ms) keep the suite fast.
- No godoc example -- timing-sensitive output would make the example
  test flaky.
