package assert

import (
	"testing"
	"time"

	"github.com/lczyk/assert/internal/core"
)

// AnyError matches any non-nil error when passed as the expected arg to Error.
// Use this sentinel rather than the empty string for clarity at call sites.
var AnyError = core.AnyError

// That asserts that predicate is true.
//
// args is an optional custom failure message: if args[0] is a string,
// it is treated as a format string (Sprintf semantics) and the rest
// of args are its arguments. Otherwise args are formatted as %v.
//
// Hard-fail variant: [github.com/lczyk/assert/require.That].
func That(t testing.TB, predicate bool, args ...any) {
	t.Helper()
	core.That(t, t.Errorf, predicate, args)
}

// Equal asserts that a == b. Argument order is (got, want) by
// convention, but the failure message names both so either reading
// is recoverable.
//
// Hard-fail variant: [github.com/lczyk/assert/require.Equal].
func Equal[T comparable](t testing.TB, a T, b T, args ...any) {
	t.Helper()
	core.Equal(t, t.Errorf, a, b, args)
}

// NotEqual asserts that a != b.
//
// Hard-fail variant: [github.com/lczyk/assert/require.NotEqual].
func NotEqual[T comparable](t testing.TB, a T, b T, args ...any) {
	t.Helper()
	core.NotEqual(t, t.Errorf, a, b, args)
}

// NearlyEqual asserts that |got - want| <= tolerance. Generic over
// numeric types. NaN comparisons always fail (NaN is not nearly equal
// to anything, including itself).
//
// Hard-fail variant: [github.com/lczyk/assert/require.NearlyEqual].
func NearlyEqual[T core.Numeric](t testing.TB, got T, want T, tolerance T, args ...any) {
	t.Helper()
	core.NearlyEqual(t, t.Errorf, got, want, tolerance, args)
}

// NoError asserts that err is nil.
//
// Hard-fail variant: [github.com/lczyk/assert/require.NoError].
func NoError(t testing.TB, err error, args ...any) {
	t.Helper()
	core.NoError(t, t.Errorf, err, args)
}

// Error asserts that err is non-nil and matches expected.
//
// expected may be:
//   - nil: passes only if err is nil (equivalent to NoError)
//   - AnyError: passes for any non-nil err
//   - error: structural match (same dynamic type and Error() string),
//     OR errors.Is wrap-chain match. ErrorIs is the strict-wrap-chain variant.
//   - string: literal substring match against err.Error()
//     (strings.Contains). Empty string matches any non-nil err
//     (equivalent to AnyError). For regex matching, pass *regexp.Regexp.
//   - *regexp.Regexp: regex pattern, matched as a substring against err.Error()
//
// Hard-fail variant: [github.com/lczyk/assert/require.Error].
func Error(t testing.TB, err error, expected any, args ...any) {
	t.Helper()
	core.Error(t, t.Errorf, err, expected, args)
}

// ErrorIs asserts that err matches expected via errors.Is semantics
// (identity or wrap chain). Use Error for structural type+message match.
//
// Hard-fail variant: [github.com/lczyk/assert/require.ErrorIs].
func ErrorIs(t testing.TB, err error, expected error, args ...any) {
	t.Helper()
	core.ErrorIs(t, t.Errorf, err, expected, args)
}

// EqualCmp compares two values using a custom comparator function.
//
// Hard-fail variant: [github.com/lczyk/assert/require.EqualCmp].
func EqualCmp[T any](t testing.TB, a T, b T, comparator func(T, T) bool, args ...any) {
	t.Helper()
	core.EqualCmp(t, t.Errorf, a, b, comparator, args)
}

// EqualCmpAny compares two values of any type using a custom
// comparator function. More generic than EqualCmp, less type-safe;
// the comparator is responsible for type assertions.
//
// Hard-fail variant: [github.com/lczyk/assert/require.EqualCmpAny].
func EqualCmpAny(t testing.TB, a any, b any, comparator func(any, any) bool, args ...any) {
	t.Helper()
	core.EqualCmpAny(t, t.Errorf, a, b, comparator, args)
}

// EqualArrays compares two slices for element-wise equality.
//
// Hard-fail variant: [github.com/lczyk/assert/require.EqualArrays].
func EqualArrays[T comparable](t testing.TB, a []T, b []T, args ...any) {
	t.Helper()
	core.EqualArrays(t, t.Errorf, a, b, args)
}

// EqualMaps compares two maps for key/value equality.
//
// Hard-fail variant: [github.com/lczyk/assert/require.EqualMaps].
func EqualMaps[K comparable, V comparable](t testing.TB, a map[K]V, b map[K]V, args ...any) {
	t.Helper()
	core.EqualMaps(t, t.Errorf, a, b, args)
}

// EqualArraysUnordered compares two slices for element-wise equality
// ignoring order (multiset equality).
//
// Hard-fail variant: [github.com/lczyk/assert/require.EqualArraysUnordered].
func EqualArraysUnordered[T comparable](t testing.TB, a []T, b []T, args ...any) {
	t.Helper()
	core.EqualArraysUnordered(t, t.Errorf, a, b, args)
}

// Nil asserts that x is nil. Handles typed-nil-in-interface (e.g.
// (*T)(nil) inside any).
//
// Hard-fail variant: [github.com/lczyk/assert/require.Nil].
func Nil(t testing.TB, x any, args ...any) {
	t.Helper()
	core.Nil(t, t.Errorf, x, args)
}

// NotNil asserts that x is not nil. Handles typed-nil-in-interface.
//
// Hard-fail variant: [github.com/lczyk/assert/require.NotNil].
func NotNil(t testing.TB, x any, args ...any) {
	t.Helper()
	core.NotNil(t, t.Errorf, x, args)
}

// Len asserts that len(x) == n. x must be array, chan, map, slice, or string.
//
// Hard-fail variant: [github.com/lczyk/assert/require.Len].
func Len(t testing.TB, x any, n int, args ...any) {
	t.Helper()
	core.Len(t, t.Errorf, x, n, args)
}

// Type asserts that obj is of type T and returns the asserted value.
// Fails softly (t.Errorf) on mismatch and returns the zero value of T,
// which the caller must guard against. The hard-fail variant aborts on
// mismatch, so its return value is safe to use unconditionally.
//
// Hard-fail variant: [github.com/lczyk/assert/require.Type].
func Type[T any](t testing.TB, obj any, args ...any) T {
	t.Helper()
	return core.Type[T](t, t.Errorf, obj, args)
}

// EqualLineByLine compares two strings line by line. Ignores a single
// trailing newline on either side. Mismatches reported in one message,
// semicolon-joined.
//
// Hard-fail variant: [github.com/lczyk/assert/require.EqualLineByLine].
func EqualLineByLine(t testing.TB, a string, b string, args ...any) {
	t.Helper()
	core.EqualLineByLine(t, t.Errorf, a, b, args)
}

// HasKey asserts that m contains key k.
//
// Hard-fail variant: [github.com/lczyk/assert/require.HasKey].
func HasKey[K comparable, V any](t testing.TB, m map[K]V, k K, args ...any) {
	t.Helper()
	core.HasKey(t, t.Errorf, m, k, args)
}

// ContainsString asserts that haystack contains needle as a substring.
//
// Hard-fail variant: [github.com/lczyk/assert/require.ContainsString].
func ContainsString(t testing.TB, haystack string, needle string, args ...any) {
	t.Helper()
	core.ContainsString(t, t.Errorf, haystack, needle, args)
}

// Panic asserts that f panics. f_recover, if non-nil, is called with
// the recovered panic value for further inspection.
//
// Hard-fail variant: [github.com/lczyk/assert/require.Panic].
func Panic(t testing.TB, f func(), f_recover func(t testing.TB, rec any), args ...any) {
	t.Helper()
	core.Panic(t, t.Errorf, f, f_recover, args)
}

// Eventually polls predicate until it returns true or timeout elapses.
// Sleeps interval between checks. Predicate is called at least once
// (even with zero timeout). Fails the test if predicate never returns
// true within the deadline.
//
// Hard-fail variant: [github.com/lczyk/assert/require.Eventually].
func Eventually(t testing.TB, predicate func() bool, timeout time.Duration, interval time.Duration, args ...any) {
	t.Helper()
	core.Eventually(t, t.Errorf, predicate, timeout, interval, args)
}
