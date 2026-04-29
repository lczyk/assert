# assert

[![lint_and_test](https://github.com/lczyk/assert/actions/workflows/lint_and_test.yml/badge.svg)](https://github.com/lczyk/assert/actions/workflows/lint_and_test.yml)

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
- Failures are soft (`t.Errorf`); the test continues. Use `t.Fatal` yourself
  if you need fail-fast semantics.
- No state hidden in package globals (besides a tiny source-line cache for
  failure rendering). Each assertion stands alone.
- Drop-in: you can mix `assert.Equal(t, ...)` and raw `if a != b { t.Errorf(...) }`
  in the same test without conflict.

## Demos

`make demo` runs a suite of intentionally-failing tests under `demo/` to show
off failure output (message + `file:line` + source-line snippet, including
multi-line calls). Output is the point — the tests are tag-gated
(`//go:build demo`) so a normal `go test ./...` stays clean.

The runner is [`demo/demo_runner.sh`](demo/demo_runner.sh); demos are
auto-discovered by grepping `^func TestDemo` from `demos_test.go`.

## dev

There is a bunch of design meanderings in [meanderings/](meanderings/); some
are implemented, some shelved, some rejected. **NOT ALL SHOULD ship** —
these are just meanderings after all. See [meanderings/README.md](meanderings/README.md)
for the index with statuses.