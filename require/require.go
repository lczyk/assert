package require

import (
	"testing"
	"time"

	"github.com/lczyk/assert/internal/core"
)

// AnyError matches any non-nil error when passed as the expected arg
// to Error. Aliased to the same sentinel exported by the assert
// package -- both forms share identity.
var AnyError = core.AnyError

// That requires that predicate is true. See assert.That for arg
// semantics. Failure aborts the test (t.Fatalf).
func That(t testing.TB, predicate bool, args ...any) {
	t.Helper()
	if msg, failed := core.That(predicate, args); failed {
		t.Fatalf("%s", msg)
	}
}

// Equal requires that a == b. See assert.Equal for the caveat on
// interface-typed T. Failure aborts the test.
func Equal[T comparable](t testing.TB, a T, b T, args ...any) {
	t.Helper()
	if msg, failed := core.Equal(a, b, args); failed {
		t.Fatalf("%s", msg)
	}
}

// NotEqual requires that a != b. See assert.Equal for the caveat on
// interface-typed T. Failure aborts the test.
func NotEqual[T comparable](t testing.TB, a T, b T, args ...any) {
	t.Helper()
	if msg, failed := core.NotEqual(a, b, args); failed {
		t.Fatalf("%s", msg)
	}
}

// NearlyEqual requires that |got - want| <= tolerance. NaN and two
// infinities of the same sign never compare nearly equal; see
// assert.NearlyEqual. Failure aborts the test.
func NearlyEqual[T core.Numeric](t testing.TB, got T, want T, tolerance T, args ...any) {
	t.Helper()
	if msg, failed := core.NearlyEqual(got, want, tolerance, args); failed {
		t.Fatalf("%s", msg)
	}
}

// NoError requires that err is nil. A typed-nil error is not nil and
// fails. Failure aborts the test.
func NoError(t testing.TB, err error, args ...any) {
	t.Helper()
	if msg, failed := core.NoError(err, args); failed {
		t.Fatalf("%s", msg)
	}
}

// Error requires that err matches expected. See assert.Error for the
// full type-switch on expected (and the panic on unsupported expected
// types). Failure aborts the test.
func Error(t testing.TB, err error, expected any, args ...any) {
	t.Helper()
	if msg, failed := core.Error(err, expected, args); failed {
		t.Fatalf("%s", msg)
	}
}

// ErrorIs requires that err matches expected via errors.Is semantics
// (identity or wrap chain). Failure aborts the test.
func ErrorIs(t testing.TB, err error, expected error, args ...any) {
	t.Helper()
	if msg, failed := core.ErrorIs(err, expected, args); failed {
		t.Fatalf("%s", msg)
	}
}

// EqualCmp requires that the comparator returns true for a, b.
// Failure aborts the test.
func EqualCmp[T any](t testing.TB, a T, b T, comparator func(T, T) bool, args ...any) {
	t.Helper()
	if msg, failed := core.EqualCmp(a, b, comparator, args); failed {
		t.Fatalf("%s", msg)
	}
}

// EqualCmpAny is the any-typed companion of EqualCmp. Failure aborts
// the test.
func EqualCmpAny(t testing.TB, a any, b any, comparator func(any, any) bool, args ...any) {
	t.Helper()
	if msg, failed := core.EqualCmpAny(a, b, comparator, args); failed {
		t.Fatalf("%s", msg)
	}
}

// EqualArrays requires element-wise equality of two slices (not
// arrays; pass a[:]). Nil equals empty. Failure aborts the test.
func EqualArrays[T comparable](t testing.TB, a []T, b []T, args ...any) {
	t.Helper()
	if msg, failed := core.EqualArrays(a, b, args); failed {
		t.Fatalf("%s", msg)
	}
}

// EqualMaps requires key/value equality of two maps. Nil equals
// empty. Failure aborts the test.
func EqualMaps[K comparable, V comparable](t testing.TB, a map[K]V, b map[K]V, args ...any) {
	t.Helper()
	if msg, failed := core.EqualMaps(a, b, args); failed {
		t.Fatalf("%s", msg)
	}
}

// EqualArraysUnordered requires multiset equality of two slices. Nil
// equals empty. Failure aborts the test.
func EqualArraysUnordered[T comparable](t testing.TB, a []T, b []T, args ...any) {
	t.Helper()
	if msg, failed := core.EqualArraysUnordered(a, b, args); failed {
		t.Fatalf("%s", msg)
	}
}

// Nil requires that x is nil. Handles typed-nil-in-interface; see
// assert.Nil. Failure aborts the test.
func Nil(t testing.TB, x any, args ...any) {
	t.Helper()
	if msg, failed := core.Nil(x, args); failed {
		t.Fatalf("%s", msg)
	}
}

// NotNil requires that x is not nil. Handles typed-nil-in-interface;
// see assert.Nil. Failure aborts the test.
func NotNil(t testing.TB, x any, args ...any) {
	t.Helper()
	if msg, failed := core.NotNil(x, args); failed {
		t.Fatalf("%s", msg)
	}
}

// Len requires that len(x) == n. See assert.Len for the panic on
// unsupported types of x. Failure aborts the test.
func Len(t testing.TB, x any, n int, args ...any) {
	t.Helper()
	if msg, failed := core.Len(x, n, args); failed {
		t.Fatalf("%s", msg)
	}
}

// Type requires that obj is of type T and returns the asserted value.
// Failure aborts the test, so the returned value is safe to use
// unconditionally on the success path. This is the canonical use case
// for require over assert. A nil obj always fails, even for T = any.
func Type[T any](t testing.TB, obj any, args ...any) T {
	t.Helper()
	v, msg, failed := core.Type[T](obj, args)
	if failed {
		t.Fatalf("%s", msg)
	}
	return v
}

// EqualLineByLine requires line-by-line equality of two strings,
// ignoring a single trailing newline on either side. Failure aborts
// the test.
func EqualLineByLine(t testing.TB, a string, b string, args ...any) {
	t.Helper()
	if msg, failed := core.EqualLineByLine(a, b, args); failed {
		t.Fatalf("%s", msg)
	}
}

// HasKey requires that m contains key k. Failure aborts the test.
func HasKey[K comparable, V any](t testing.TB, m map[K]V, k K, args ...any) {
	t.Helper()
	if msg, failed := core.HasKey(m, k, args); failed {
		t.Fatalf("%s", msg)
	}
}

// ContainsString requires that haystack contains needle as a
// substring. Failure aborts the test.
func ContainsString(t testing.TB, haystack string, needle string, args ...any) {
	t.Helper()
	if msg, failed := core.ContainsString(haystack, needle, args); failed {
		t.Fatalf("%s", msg)
	}
}

// Panic requires that f panics; f_recover, if non-nil, is called with
// the recovered value. See assert.Panic for panic(nil) and for f
// leaving via runtime.Goexit. Failure aborts the test.
func Panic(t testing.TB, f func(), f_recover func(t testing.TB, rec any), args ...any) {
	t.Helper()
	if msg, failed := core.Panic(t, f, f_recover, args); failed {
		t.Fatalf("%s", msg)
	}
}

// Eventually requires that predicate becomes true within timeout,
// polling every interval; see assert.Eventually for the exact call
// schedule and why predicate must not call require.* itself. Failure
// aborts the test.
func Eventually(t testing.TB, predicate func() bool, timeout time.Duration, interval time.Duration, args ...any) {
	t.Helper()
	if msg, failed := core.Eventually(predicate, timeout, interval, args); failed {
		t.Fatalf("%s", msg)
	}
}
