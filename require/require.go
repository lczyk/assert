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
	core.That(t, t.Fatalf, predicate, args)
}

// Equal requires that a == b. Failure aborts the test.
func Equal[T comparable](t testing.TB, a T, b T, args ...any) {
	t.Helper()
	core.Equal(t, t.Fatalf, a, b, args)
}

// NotEqual requires that a != b. Failure aborts the test.
func NotEqual[T comparable](t testing.TB, a T, b T, args ...any) {
	t.Helper()
	core.NotEqual(t, t.Fatalf, a, b, args)
}

// NearlyEqual requires that |got - want| <= tolerance. Failure aborts
// the test.
func NearlyEqual[T core.Numeric](t testing.TB, got T, want T, tolerance T, args ...any) {
	t.Helper()
	core.NearlyEqual(t, t.Fatalf, got, want, tolerance, args)
}

// NoError requires that err is nil. Failure aborts the test.
func NoError(t testing.TB, err error, args ...any) {
	t.Helper()
	core.NoError(t, t.Fatalf, err, args)
}

// Error requires that err is non-nil and matches expected. See
// assert.Error for the full type-switch on expected. Failure aborts
// the test.
func Error(t testing.TB, err error, expected any, args ...any) {
	t.Helper()
	core.Error(t, t.Fatalf, err, expected, args)
}

// ErrorIs requires that err matches expected via errors.Is semantics
// (identity or wrap chain). Failure aborts the test.
func ErrorIs(t testing.TB, err error, expected error, args ...any) {
	t.Helper()
	core.ErrorIs(t, t.Fatalf, err, expected, args)
}

// EqualCmp requires that the comparator returns true for a, b.
// Failure aborts the test.
func EqualCmp[T any](t testing.TB, a T, b T, comparator func(T, T) bool, args ...any) {
	t.Helper()
	core.EqualCmp(t, t.Fatalf, a, b, comparator, args)
}

// EqualCmpAny is the any-typed companion of EqualCmp. Failure aborts
// the test.
func EqualCmpAny(t testing.TB, a any, b any, comparator func(any, any) bool, args ...any) {
	t.Helper()
	core.EqualCmpAny(t, t.Fatalf, a, b, comparator, args)
}

// EqualArrays requires element-wise equality of two slices. Failure
// aborts the test.
func EqualArrays[T comparable](t testing.TB, a []T, b []T, args ...any) {
	t.Helper()
	core.EqualArrays(t, t.Fatalf, a, b, args)
}

// EqualMaps requires key/value equality of two maps. Failure aborts
// the test.
func EqualMaps[K comparable, V comparable](t testing.TB, a map[K]V, b map[K]V, args ...any) {
	t.Helper()
	core.EqualMaps(t, t.Fatalf, a, b, args)
}

// EqualArraysUnordered requires multiset equality of two slices.
// Failure aborts the test.
func EqualArraysUnordered[T comparable](t testing.TB, a []T, b []T, args ...any) {
	t.Helper()
	core.EqualArraysUnordered(t, t.Fatalf, a, b, args)
}

// Nil requires that x is nil. Failure aborts the test.
func Nil(t testing.TB, x any, args ...any) {
	t.Helper()
	core.Nil(t, t.Fatalf, x, args)
}

// NotNil requires that x is not nil. Failure aborts the test.
func NotNil(t testing.TB, x any, args ...any) {
	t.Helper()
	core.NotNil(t, t.Fatalf, x, args)
}

// Len requires that len(x) == n. Failure aborts the test.
func Len(t testing.TB, x any, n int, args ...any) {
	t.Helper()
	core.Len(t, t.Fatalf, x, n, args)
}

// Type requires that obj is of type T and returns the asserted value.
// Failure aborts the test, so the returned value is safe to use
// unconditionally on the success path. This is the canonical use case
// for require over assert.
func Type[T any](t testing.TB, obj any, args ...any) T {
	t.Helper()
	return core.Type[T](t, t.Fatalf, obj, args)
}

// EqualLineByLine requires line-by-line equality of two strings.
// Failure aborts the test.
func EqualLineByLine(t testing.TB, a string, b string, args ...any) {
	t.Helper()
	core.EqualLineByLine(t, t.Fatalf, a, b, args)
}

// HasKey requires that m contains key k. Failure aborts the test.
func HasKey[K comparable, V any](t testing.TB, m map[K]V, k K, args ...any) {
	t.Helper()
	core.HasKey(t, t.Fatalf, m, k, args)
}

// ContainsString requires that haystack contains needle as a
// substring. Failure aborts the test.
func ContainsString(t testing.TB, haystack string, needle string, args ...any) {
	t.Helper()
	core.ContainsString(t, t.Fatalf, haystack, needle, args)
}

// Panic requires that f panics. Failure aborts the test.
func Panic(t testing.TB, f func(), f_recover func(t testing.TB, rec any), args ...any) {
	t.Helper()
	core.Panic(t, t.Fatalf, f, f_recover, args)
}

// Eventually requires that predicate becomes true within timeout.
// Failure aborts the test.
func Eventually(t testing.TB, predicate func() bool, timeout time.Duration, interval time.Duration, args ...any) {
	t.Helper()
	core.Eventually(t, t.Fatalf, predicate, timeout, interval, args)
}
