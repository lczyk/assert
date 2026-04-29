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

This is exactly the wrapper we wrote in `integration/helpers.go`. We've now
seen the same shape across multiple projects — it should live in `assert`.

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

Add two thin helpers, one each for the two domains we hit:

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

`EqualBytes` is the byte-slice counterpart most projects need; `EqualText` is
the natural form when callers already have strings (config files, generated
output, snapshot fixtures). `EqualBytes` can delegate to `EqualText` when
both sides are valid UTF-8, otherwise fall back to a fixed-width hex diff.

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

## Companion: `assert.Fail`

Smaller but related: during the same refactor we found ourselves writing

```go
assert.That(t, false, "require %s not found", mod)
```

to flag an unreachable branch. `assert.That(t, false, ...)` reads as a riddle
on first encounter. A two-line helper would clarify intent:

```go
// Fail unconditionally records a test failure with the given message. The
// args are forwarded to the failure message in printf style, matching
// That/NoError.
func Fail(t testing.TB, args ...any)
```

This is independent of `EqualBytes` and could land separately, but it falls
out of the same observation: callers reach for `That(false, ...)` because
there's no direct way to spell "I've already decided this is a failure."

## Naming notes

- `EqualBytes` rather than `BytesEqual` follows the existing convention
  (`Equal`, `EqualArrays`, `EqualMaps`, `EqualLineByLine` — verb second is
  not the house style; the noun describes *what kind* of equality).
- `EqualText` rather than `EqualString` — emphasises the *diff-friendly*
  semantics. `EqualString` would imply identical behaviour to `Equal` for the
  string type, which would be misleading since the failure rendering is
  fundamentally different.

## Scope / non-goals

- No structural diff for arbitrary types. `EqualBytes`/`EqualText` are
  text/binary specific. A general `Diff` helper for structs is a much larger
  proposal and out of scope here.
- No coloured output. Tests run in many environments (CI, IDE panels,
  plain stdout); colour is an opinion and we shouldn't bake it in. If
  someone wants colour, that's a wrapper above this layer.
- No third-party diff dependency. A hand-rolled Myers diff is ~100 lines and
  has no external imports. Keeping `assert` zero-dep is part of why people
  pick it; a diff helper that pulls in a 4 MB module for one test convenience
  would betray that.

## Implementation notes

- Existing `EqualLineByLine` already does the line split + first-mismatch
  walk; `EqualText` extends that to emit a windowed diff with context.
- Suggested placement: extend `interface.go` with the two new functions; put
  the diff implementation in a new `internal/textdiff/textdiff.go` so the
  public API surface stays narrow.
- The optional `args ...any` "prelude" formatting is identical to what
  `NoError(t, err, args...)` already does — reuse `argsToMessage`.

## Migration

Once these helpers land, downstream wrappers can drop:

```go
// before
func bytesEqual(t *testing.T, got, want []byte, msg string) { ... }
bytesEqual(t, gotMod, wantMod, "go.mod after roundtrip")

// after
assert.EqualBytes(t, gotMod, wantMod, "go.mod after roundtrip")
```

No breaking changes. Both helpers are additive.
