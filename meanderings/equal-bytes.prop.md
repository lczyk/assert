---
status: open
date: 2026-04-27
description: `EqualBytes` / `EqualText` with diff-friendly failure rendering
---

# Proposal: `assert.EqualBytes` / `assert.EqualText` for `github.com/lczyk/assert`

## Motivation

While building the integration test suite for `change-go-version` we kept
hitting the same gap: comparing two byte slices (or two text blobs) where, on
mismatch, we want to see *what differs*, not just the raw `%v` output of two
slices.

The existing toolbox is close but not quite enough:

- `assert.Equal[T comparable]` — `[]byte` is not comparable, so it doesn't
  apply to byte slices at all.
- `assert.EqualArrays[T comparable]` — works for `[]byte`, but its failure
  rendering treats the slice as opaque elements. For text it produces a wall
  of characters; for binary it's even less readable.
- `assert.EqualLineByLine(string, string)` — useful, but stops at the first
  mismatched line and doesn't give surrounding context.

So every consumer ends up writing the same wrapper:

```go
func bytesEqual(t *testing.T, got, want []byte, msg string) {
    t.Helper()
    assert.That(t, bytes.Equal(got, want),
        "%s: differ\n--- got ---\n%s\n--- want ---\n%s", msg, got, want)
}
```

This is exactly the wrapper written in `integration/helpers.go`. The same
shape has surfaced across multiple projects.

## Concrete pain example

The most acute call site is a round-trip test: apply two opposing changes to
a module's `go.mod` / `go.sum` and assert the bytes are identical to a
single-step reference. When this fails, we want a diff. Today we either get a
single-line `%v` dump (truncated, unreadable) or we hand-roll the wrapper
above (which doesn't show *which lines* differ).

A unified-diff output would immediately point at the offending line; the
current rendering is fine for tiny inputs (two strings of three characters)
and useless for anything realistic (a 30-line `go.mod`).

## Proposal

Two helpers, one each for the two domains:

```go
// EqualBytes asserts that got and want are byte-equal. On failure it renders
// a unified diff if both sides are valid UTF-8, or a side-by-side hex dump
// otherwise. The optional args are forwarded to the failure message as a
// printf prelude (matching the existing NoError/Error signature style).
func EqualBytes(t testing.TB, got, want []byte, args ...any)

// EqualText asserts that two strings are equal. On failure it renders a
// unified diff (`--- got` / `+++ want` plus `@@` hunks). args is the same
// optional printf prelude.
func EqualText(t testing.TB, got, want string, args ...any)
```

`EqualBytes` is the byte-slice form; `EqualText` is for callers that
already have strings (config files, generated output, snapshot
fixtures). `EqualBytes` could delegate to `EqualText` when both sides
are valid UTF-8, and fall back to a fixed-width hex diff otherwise.

## Failure output sketch

For

```go
got := []byte("module fixture\n\ngo 1.21.0\n\nrequire foo v1.2.3\n")
want := []byte("module fixture\n\ngo 1.21\n\nrequire foo v1.2.4\n")
assert.EqualBytes(t, got, want, "after roundtrip")
```

the message would be something like:

```
EqualBytes failed: after roundtrip
--- got
+++ want
@@ -1,5 +1,5 @@
 module fixture

-go 1.21.0
+go 1.21

-require foo v1.2.3
+require foo v1.2.4
```

For non-UTF-8 inputs:

```
EqualBytes failed: after roundtrip
length: got=18, want=18; first diff at offset 4
got : 89 50 4e 47 0d 0a 1a 0a  00 00 00 0d 49 48 44 52  ...
want: 89 50 4e 47 0d 0a 1a 0a  00 00 00 0d 49 48 44 52  ...
                              ^^
```

The exact format is bikesheddable; the contract is "show me where they
differ, not just *that* they differ".

## Side note: `assert.Fail`

Independent but surfaced alongside this proposal. To flag an
unreachable branch the current spelling is:

```go
assert.That(t, false, "require %s not found", mod)
```

A direct form:

```go
// Fail unconditionally records a test failure with the given message.
// args are forwarded printf-style, matching That/NoError.
func Fail(t testing.TB, args ...any)
```

Independent of `EqualBytes`; could land separately.

## Naming notes

- `EqualBytes` rather than `BytesEqual` follows the existing convention
  (`Equal`, `EqualArrays`, `EqualMaps`, `EqualLineByLine` — verb second is
  not the house style; the noun describes *what kind* of equality).
- `EqualText` rather than `EqualString` — emphasises the *diff-friendly*
  semantics. `EqualString` would imply identical behaviour to `Equal` for the
  string type, which would be misleading since the failure rendering is
  fundamentally different.

## Scope notes

- Structural diff for arbitrary types is out of scope here —
  `EqualBytes` / `EqualText` are text/binary specific.
- Coloured output: tests run in many environments (CI, IDE panels,
  plain stdout). Whether to bake colour in is an open question.
- Diff dependency: a hand-rolled Myers diff is ~100 lines and has no
  external imports. `go-cmp` would pull in a ~4 MB module. The
  zero-dep status of `assert` is a design tension here.

## Implementation notes

- Existing `EqualLineByLine` already does the line split +
  first-mismatch walk; `EqualText` could extend that to emit a windowed
  diff with context.
- Possible placement: extend `interface.go` with the new functions; put
  the diff implementation in `internal/textdiff/textdiff.go` to keep
  the public API narrow.
- The optional `args ...any` "prelude" formatting is identical to what
  `NoError(t, err, args...)` already does — `argsToMessage` is reusable.

## Migration

Downstream wrappers like:

```go
func bytesEqual(t *testing.T, got, want []byte, msg string) { ... }
bytesEqual(t, gotMod, wantMod, "go.mod after roundtrip")
```

would become:

```go
assert.EqualBytes(t, gotMod, wantMod, "go.mod after roundtrip")
```

Both helpers are additive — no breaking changes to the existing API.
