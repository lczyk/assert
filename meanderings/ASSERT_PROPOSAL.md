# Proposal: gaps in `github.com/lczyk/assert` found while migrating goruby

Migrating goruby's `utils/testing.go` to `github.com/lczyk/assert@v0.3.1` was almost
1-to-1. One real semantic gap surfaced; a few minor ergonomic items are listed too.

## 1. `assert.Error` with an `error` value: type+message comparison

### Gap

`compare.Errors` (used by `assert.Error` when `expected` is an `error`) is:

```go
func Errors(err error, target error) bool {
    if err == nil && target == nil { return true }
    return errors.Is(err, target)
}
```

`errors.Is` matches by identity (or unwrapped chain). It does **not** match two
distinct error instances that share the same dynamic type and `Error()` string.

The old `utils.CompareErrors` did:

```go
reflect.TypeOf(a) == reflect.TypeOf(b) && a.Error() == b.Error()
```

Many goruby tests construct an expected value (e.g. `NewTypeError("...")`) and
compare against a freshly-built error returned by the code under test. Those
two values are never `errors.Is`-equal but **are** structurally equivalent, so
the migration broke ~15 assertions in `object/` and `parser/`.

Workaround used in this repo: a private `assertErrorLike` helper in
[object/asserterror_test.go](object/asserterror_test.go) and
[parser/asserterror_test.go](parser/asserterror_test.go).

### Proposed addition

Either:

**(a)** A new comparator in `compare/`:

```go
// ErrorsLike reports whether err and target have the same dynamic type and
// equal Error() strings. Useful when tests construct expected error values
// inline rather than referencing sentinels.
func ErrorsLike(err, target error) bool {
    if err == nil || target == nil { return err == target }
    return reflect.TypeOf(err) == reflect.TypeOf(target) && err.Error() == target.Error()
}
```

**(b)** A new top-level assertion that uses it:

```go
// ErrorLike asserts err matches expected by dynamic type and Error() string.
func ErrorLike(t testing.TB, err error, expected error, args ...any)
```

**(c)** Have `assert.Error(t, err, errVal)` fall back to type+message equality
when `errors.Is` returns false. Behaviour-changing — probably not worth it.

Preference: **(a) + (b)**. Keeps `Error` semantics tight and gives callers an
explicit opt-in for the looser comparison.

## 2. Minor: `AnyError = ""` is surprising

`assert.Error(t, err, "")` matches any non-nil error because the empty regex
matches everything. The exported `AnyError` constant documents this, but at
the call site `assert.Error(t, err, "")` reads as "expected error message is
the empty string", which is the opposite of what it does. Consider:

- A separate `AnyError` sentinel of distinct type, e.g. `var AnyError = errAny{}`.
- Or `assert.AnyError(t, err)` as a dedicated helper.

Low priority — `AnyError` works once you know.

## 3. Minor: `assert.Error` with `string` does substring/regex match

When `expected` is a `string`, the match is `regexp.MustCompile(expected).MatchString(err.Error())`.
That is a regex *substring* match, not an equality match. Two consequences:

- A test that expects exactly `"oops"` will pass on `"oops happened, then more oops"`.
- Special characters in messages (`(`, `)`, `.`, `?`, etc.) silently change semantics.

`utils.AssertError` had the same behaviour, so this is not a regression — but
the docstring on `Error` does not state it. Suggest documenting it explicitly,
and/or providing a sibling `assert.ErrorMessage(t, err, exact string)` for the
common "exact equality" case.

## 4. Nice-to-have: `EqualCmp` accepting a `string` message arg

`utils.AssertEqualCmp` and `assert.EqualCmp` both lack a variadic `args ...any`
message tail (unlike `Equal`, `That`, etc.). When a custom comparator fails, the
default `expected '%v' (%T) == '%v' (%T)` message is often unhelpful for
domain types. Adding `args ...any` would unify the API.

## 5. Nice-to-have: `Equal` for non-comparable types via `reflect.DeepEqual`

`assert.Equal` requires `comparable`, which excludes slices, maps, and structs
containing them. Today the workaround is `EqualArrays` / `EqualMaps` /
`EqualCmp`. A generic `assert.DeepEqual[T any](t, a, b T)` (or just letting
`Equal` fall back to `reflect.DeepEqual` for non-comparable T via constraints)
would remove the per-shape helpers.

The original `utils.AssertDeepEqual` was commented out, suggesting the same
itch.
