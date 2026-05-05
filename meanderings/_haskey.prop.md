---
status: implemented
date: 2026-05-05
description: HasKey for map presence — only the prose-fitting third of #5
---

# Proposal: HasKey

> Partial resolution of #5 from the godep_differ migration proposal.
> Of the three proposed containment primitives (`Contains`,
> `HasKey`, `MapContains`), only `HasKey` shipped. The other two
> were left out because their prose-fit and demand signal didn't
> earn the surface.

## Gap

Map-key-presence assertions today fall back to one of:

```go
_, ok := m[k]
assert.That(t, ok)

assert.That(t, len(m) > 0 && m[k] != zero) // unsafe — zero may be valid
```

Both work but lose the structured failure msg ("expected key %q in
map, got keys %v"). The first is a two-line dance every site.

## Why HasKey but not Contains / MapContains

Reads-as-prose check across the three:

- `assert.HasKey(t, m, k)` -> "assert has key m k" -- natural.
- `assert.Contains(t, s, v)` -> "assert contains s v" -- borderline,
  reads sideways. `In(t, v, s)` would read better but collides with
  no Go idiom.
- `assert.MapContains(t, m, k, v)` -> "assert map contains m k v"
  -- awkward. `Equal(t, m[k], v)` covers the common case (modulo
  zero-value-on-miss footgun, which is real but per-callsite-fixable).

Empirical signal: zero usage of `slices.Contains` or `_, ok := m[k]`
patterns across the user's ~85-file corpus. Demand for `Contains` /
`MapContains` does not exist in this codebase. `HasKey` was shipped on
prose-fit alone; the others wait for evidence.

## Sig

```go
func HasKey[K comparable, V any](t testing.TB, m map[K]V, k K, args ...any)
```

`V` is `any` -- caller cares about presence, not value. Internally
uses comma-ok lookup, not value comparison, so zero-V at present
key still passes.

## Failure message shape

```
expected key '<k>' (<KType>) in map, got keys [<k1> <k2> ...]
```

All map keys included. Iteration order is map-randomised; test
assertions on this msg should match by substring of the missing key
('expected key 'X'') rather than the full keys list.

For very large maps this msg is verbose. Acceptable cost: the
message only renders on failure (lazy via msg_fun closure), and
debugging "where did my key go" is exactly when a full key dump
helps.

## What this does NOT do

- No `Contains[T comparable](t, s []T, v T)`. Holding the line on
  prose-fit + zero downstream demand. Caller writes
  `assert.That(t, slices.Contains(s, v))`.
- No `MapContains[K, V comparable](t, m, k, v)`. Caller writes
  `assert.Equal(t, m[k], v)` if zero-value-on-miss is acceptable, or
  guards with `HasKey` first if it isn't.
- No deterministic ordering of the failure-msg keys list. `K` is
  just `comparable`, not `cmp.Ordered`, so generic sort isn't free.
  Caller should assert on substring, not full list.

## Outcome

Shipped in v0.6.x.

- `interface.go`: `HasKey[K comparable, V any]` added.
- `assert_test.go`: `TestHasKey` covering string-key / int-key /
  empty-map / custom-msg / present-but-zero-value-V (regression
  for comma-ok vs value-equality bug).

## Reads-as-prose check

`assert.HasKey(t, results, "user-id")` -> "assert has key results
user-id". Reads.
