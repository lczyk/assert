// Package assert is a thin wrapper over Go's standard testing framework that
// makes test bodies a little terser -- closer in feel to pytest -- without
// introducing a custom runner, parallel reporting layer, or DSL.
//
// Every assertion takes a [testing.TB] as its first argument and calls
// [testing.TB.Errorf] on failure. Failures are soft; the test continues
// past a failed assertion. For hard-fail (test aborts on first failure)
// use the sibling sub-package [github.com/lczyk/assert/require], which
// mirrors the same API but calls [testing.TB.Fatalf] instead.
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
//
// Failure messages render values with %v, truncated beyond 2 KB. The
// call-site snippet needs the source file to be readable at test time;
// under -trimpath, or when the test binary runs on another machine, it
// degrades to plain file:line.
//
// Every assertion takes a trailing args ...any custom message: a lone
// string verbatim, a format string plus arguments, or anything else
// printed with %v. See [That]. go vet's printf analysis does not cover
// these calls.
//
// The equality assertions accept interface types for their type
// parameter ([Equal] with any, for instance) and, like ==, panic on two
// values of the same non-comparable dynamic type. [EqualCmp] with
// reflect.DeepEqual covers those.
package assert
