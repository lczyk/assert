// Package assert is a thin wrapper over Go's standard testing framework that
// makes test bodies a little terser -- closer in feel to pytest -- without
// introducing a custom runner, parallel reporting layer, or DSL.
//
// Every assertion takes a [testing.TB] as its first argument and ultimately
// calls [testing.TB.Errorf] (or [testing.TB.Fatalf] for assertions that would
// otherwise leave the caller with a zero value, like [Type]). Failures are
// soft by default; the test continues. Use t.Fatal yourself if you need
// fail-fast semantics.
//
// # Example
//
// Instead of:
//
//	func TestExample(t *testing.T) {
//		a, b := 1, 2
//		if a == b {
//			t.Errorf("expected %d != %d", a, b)
//		}
//	}
//
// write:
//
//	func TestExample(t *testing.T) {
//		a, b := 1, 2
//		assert.That(t, a != b)
//	}
//
// # Assertion catalogue
//
// Boolean: [That].
//
// Equality: [Equal], [NotEqual], [NearlyEqual], [EqualCmp], [EqualCmpAny],
// [EqualArrays], [EqualArraysUnordered], [EqualMaps], [EqualLineByLine].
//
// Errors: [NoError], [Error], [ErrorIs]. The [AnyError] sentinel matches any
// non-nil error when passed to [Error].
//
// Nil / type / shape: [Nil], [NotNil], [Type], [Len], [HasKey],
// [ContainsString].
//
// Control flow: [Panic], [Eventually].
//
// # Design notes
//
// Plays nicely with go test, -run, -v, -race, t.Run subtests, table tests,
// and parallel tests -- all unchanged. No state hidden in package globals
// beyond a small source-line cache used to render failures with a snippet of
// the failing call site. Drop-in: assert.Equal(t, ...) and raw
// if a != b { t.Errorf(...) } coexist freely in the same test.
package assert
