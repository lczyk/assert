# assert

[![lint_and_test](https://github.com/lczyk/assert/actions/workflows/lint_and_test.yml/badge.svg)](https://github.com/lczyk/assert/actions/workflows/lint_and_test.yml)
[![Go Reference](https://pkg.go.dev/badge/github.com/lczyk/assert.svg)](https://pkg.go.dev/github.com/lczyk/assert)

Mini package to make writing tests in golang a bit neater -- a little bit more like `pytest`.

For example this:

```go
func TestExample(t *testing.T) {
	a := 1
	b := 2
	if a == b {
		t.Errorf("Expected %d to not equal %d", a, b)
	}
}
```

becomes:

```go
func TestExample(t *testing.T) {
	a := 1
	b := 2
	assert.That(t, a != b)
}
```

Not a big difference but over the course of a large test suite it adds up.

## Design

This package is a **thin wrapper** over Go's standard `testing` framework — no
custom runner, no parallel reporting layer, no DSL. Every assertion ultimately
calls `t.Errorf` (or `t.Helper`) on the `testing.TB` you pass in. That's why
`t` is the first argument to every assertion: nothing here works without it,
because there is nothing else to fall back on.

Consequences:

- Plays nicely with `go test`, `-run`, `-v`, `-race`, `t.Run` subtests, table
  tests, parallel tests — all unchanged.
- Failures in `assert.*` are soft (`t.Errorf`); the test continues. The
  sibling [`require`](#soft-vs-hard-assert-vs-require) package provides the
  hard (`t.Fatalf`) variant.
- No state hidden in package globals (besides a tiny source-line cache for
  failure rendering). Each assertion stands alone.
- Drop-in: you can mix `assert.Equal(t, ...)` and raw `if a != b { t.Errorf(...) }`
  in the same test without conflict.

## Soft vs hard: `assert` vs `require`

Two packages, same surface, different failure mode:

- [`github.com/lczyk/assert`](https://pkg.go.dev/github.com/lczyk/assert) -- soft. `t.Errorf` on failure. Test continues.
- [`github.com/lczyk/assert/require`](https://pkg.go.dev/github.com/lczyk/assert/require) -- hard. `t.Fatalf` on failure. Test aborts.

Pattern: `require.*` for preconditions, `assert.*` for the behaviour under test.

```go
func TestParse(t *testing.T) {
    raw, err := os.ReadFile("fixture.json")
    require.NoError(t, err) // setup; no point continuing on failure

    cfg, err := Parse(raw)
    require.NoError(t, err)

    assert.Equal(t, cfg.Port, 8080)        // behaviour
    assert.Equal(t, cfg.Host, "localhost") // both checked even if Port fails
}
```

### Goroutine hazard

`t.Fatalf` calls `runtime.Goexit`, which only unwinds the goroutine it runs on.
A failed `require.*` call from a background goroutine still records the
failure, but only that goroutine stops: whatever it was going to do next,
including signalling the test that it is done, never happens, so a test
waiting on it hangs unless the signal was deferred. Once the test function
has returned, any `assert.*` or `require.*` call from a leftover goroutine
panics inside `testing` and takes the whole test binary down.
**Always call `require.*` from the test goroutine.** From background
goroutines either use `assert.*` and make the test wait for them, or send
the error back to the test goroutine through a channel and check it there.

## Demos

`make demo` runs a suite of intentionally-failing tests under `demo/` to show
off failure output (message + `file:line` + source-line snippet, including
multi-line calls). Output is the point — the tests are tag-gated
(`//go:build demo`) so a normal `go test ./...` stays clean.

The runner is [`demo/demo_runner.sh`](demo/demo_runner.sh); demos are
auto-discovered by grepping `^func Test(Demo|Vanilla)` from `demos_test.go`
and `vanilla_test.go`. A `TestDemoX` with a matching `TestVanillaX` runs
right after it, so the stdlib and assert renderings of the same failure
sit next to each other.

## dev

There is a bunch of design meanderings in [meanderings/](meanderings/); some
are implemented, some superseded, some still open. **NOT ALL SHOULD ship** --
these are just meanderings after all. See [meanderings/README.md](meanderings/README.md)
for the index with statuses.