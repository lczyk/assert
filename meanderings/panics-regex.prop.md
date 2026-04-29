---
status: open
date: 2026-04-29
description: `Panics(t, f, "regex")` shorthand for panic-message regex match
---

# Proposal: `Panics(t, f, "regex")` shorthand

## Gap

Current `Panic(t, f, f_recover)` requires a callback to inspect the
recovered value. Common case: assert that `f` panics and the panic
message matches a pattern. That's a 5-line callback today:

```go
assert.Panic(t, f, func(t testing.TB, rec any) {
    s, ok := rec.(string)
    assert.That(t, ok)
    assert.ContainsString(t, s, "boom")
})
```

Mirror Rust's `#[should_panic(expected = "...")]` with:

```go
// Panics asserts that f panics with a value whose stringified form
// matches the regex pattern.
func Panics(t testing.TB, f func(), pattern string)
```

Wrapper over existing `Panic`.

## Open questions

- Match against `fmt.Sprint(rec)` or require panic value to be `string`
  / `error`? Sprint is more permissive; type-restrict is stricter but
  may surprise.
- Regex-substring match (like `assert.Error` does) or exact match?
