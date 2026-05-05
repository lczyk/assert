package assert

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/lczyk/assert/compare"
)

func That(t testing.TB, predicate bool, args ...any) {
	t.Helper()
	assert(t, nested_assert_parent, predicate, args)
}

func Equal[T comparable](t testing.TB, a T, b T, args ...any) {
	t.Helper()
	if a == b {
		return
	}
	file, line := get_parent_info(1)
	msg := args_to_message(func() string {
		return fmt.Sprintf("expected '%v' (%T) == '%v' (%T)", a, a, b, b)
	}, args)
	if loc, err := loc_str(file, line); err != nil {
		t.Errorf(msg+" in %s:%d", file, line)
	} else {
		t.Errorf("%s in %s", msg, loc)
	}
}

func NotEqual[T comparable](t testing.TB, a T, b T, args ...any) {
	t.Helper()
	if a != b {
		return
	}
	file, line := get_parent_info(1)
	msg := args_to_message(func() string {
		return fmt.Sprintf("expected '%v' (%T) != '%v' (%T)", a, a, b, b)
	}, args)
	if loc, err := loc_str(file, line); err != nil {
		t.Errorf(msg+" in %s:%d", file, line)
	} else {
		t.Errorf("%s in %s", msg, loc)
	}
}

// NearlyEqual asserts that |got - want| <= tolerance. Generic over
// numeric types. NaN comparisons always fail (NaN is not nearly equal
// to anything, including itself).
func NearlyEqual[T numeric](t testing.TB, got T, want T, tolerance T, args ...any) {
	t.Helper()
	var diff T
	if got >= want {
		diff = got - want
	} else {
		diff = want - got
	}
	if diff <= tolerance {
		return
	}
	file, line := get_parent_info(1)
	msg := args_to_message(func() string {
		return fmt.Sprintf("expected '%v' nearly equal to '%v' (tolerance %v), got diff %v", got, want, tolerance, diff)
	}, args)
	if loc, err := loc_str(file, line); err != nil {
		t.Errorf(msg+" in %s:%d", file, line)
	} else {
		t.Errorf("%s in %s", msg, loc)
	}
}

// Assert that error is nil.
func NoError(t testing.TB, err error, args ...any) {
	t.Helper()
	assert_error(t, nested_assert_parent, err, nil, args)
}

// AnyError matches any non-nil error when passed as the expected arg to Error.
// Use this sentinel rather than the empty string for clarity at call sites.
var AnyError error = anyErr{}

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
func Error(t testing.TB, err error, expected any, args ...any) {
	t.Helper()
	assert_error(t, nested_assert_parent, err, expected, args)
}

// ErrorIs asserts that err matches expected via errors.Is semantics
// (identity or wrap chain). Use Error for structural type+message match.
func ErrorIs(t testing.TB, err error, expected error, args ...any) {
	t.Helper()
	if compare.ErrorsIs(err, expected) {
		return
	}
	file, line := get_parent_info(1)
	msg := args_to_message(func() string {
		if err == nil {
			return fmt.Sprintf("expected error %s, got no error (nil)", describe_err(expected))
		}
		if expected == nil {
			return fmt.Sprintf("expected no error, got %s", describe_err(err))
		}
		return fmt.Sprintf("expected errors.Is('%v', '%v') to be true, got %s", err, expected, describe_err(err))
	}, args)
	if loc, err := loc_str(file, line); err != nil {
		t.Errorf(msg+" in %s:%d", file, line)
	} else {
		t.Errorf("%s in %s", msg, loc)
	}
}

// Compare two values using a custom comparator function.
func EqualCmp[T any](t testing.TB, a T, b T, comparator func(T, T) bool, args ...any) {
	t.Helper()
	equal_cmp(t, nested_assert_parent, a, b, comparator, args)
}

// Compare two values of any type using a custom comparator function.
// This is a more generic version of EqualCmp, but it is less type-safe.
// The comparator function is responsible for type assertions.
func EqualCmpAny(t testing.TB, a any, b any, comparator func(any, any) bool, args ...any) {
	t.Helper()
	equal_cmp_any(t, nested_assert_parent, a, b, comparator, args)
}

// Compare two arrays for element-wise equality.
func EqualArrays[T comparable](t testing.TB, a []T, b []T, args ...any) {
	t.Helper()
	equal_cmp(t, nested_assert_parent, a, b, compare.Arrays, args)
}

// Compare two maps for key/value equality.
func EqualMaps[T comparable, V comparable](t testing.TB, a map[T]V, b map[T]V, args ...any) {
	t.Helper()
	equal_cmp(t, nested_assert_parent, a, b, compare.Maps, args)
}

func EqualArraysUnordered[T comparable](t testing.TB, a []T, b []T, args ...any) {
	t.Helper()
	equal_cmp(t, nested_assert_parent, a, b, compare.ArraysUnordered, args)
}

// Assert that x is nil. Handles typed-nil-in-interface (e.g. (*T)(nil) inside any).
func Nil(t testing.TB, x any, args ...any) {
	t.Helper()
	if is_nil(x) {
		return
	}
	file, line := get_parent_info(1)
	msg := args_to_message(func() string { return fmt.Sprintf("expected nil, got %s", describe_non_nil(x)) }, args)
	if loc, err := loc_str(file, line); err != nil {
		t.Errorf(msg+" in %s:%d", file, line)
	} else {
		t.Errorf("%s in %s", msg, loc)
	}
}

// Assert that x is not nil. Handles typed-nil-in-interface.
func NotNil(t testing.TB, x any, args ...any) {
	t.Helper()
	if !is_nil(x) {
		return
	}
	file, line := get_parent_info(1)
	msg := args_to_message(func() string { return fmt.Sprintf("expected non-nil, got nil (%T)", x) }, args)
	if loc, err := loc_str(file, line); err != nil {
		t.Errorf(msg+" in %s:%d", file, line)
	} else {
		t.Errorf("%s in %s", msg, loc)
	}
}

// Assert that len(x) == n. x must be array, chan, map, slice, or string.
func Len(t testing.TB, x any, n int, args ...any) {
	t.Helper()
	v := reflect.ValueOf(x)
	switch v.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
	default:
		panic(fmt.Sprintf("Len: unsupported kind %s", v.Kind()))
	}
	got := v.Len()
	if got == n {
		return
	}
	file, line := get_parent_info(1)
	msg := args_to_message(func() string { return fmt.Sprintf("expected len %d, got len %d: %v", n, got, x) }, args)
	if loc, err := loc_str(file, line); err != nil {
		t.Errorf(msg+" in %s:%d", file, line)
	} else {
		t.Errorf("%s in %s", msg, loc)
	}
}

// Type asserts that obj is of type T and returns the asserted value.
// Unlike other assertions in this package, Type fails the test fatally
// (t.Fatalf) on mismatch, since the returned zero value is a nil-deref
// footgun for the caller.
func Type[T any](t testing.TB, obj any, args ...any) T {
	t.Helper()
	return assert_type[T](t, nested_assert_parent, obj, args...)
}

func EqualLineByLine(t testing.TB, a string, b string, args ...any) {
	t.Helper()
	// Ignore a single trailing newline on either side so comparing file-like
	// content to a literal doesn't fail on an incidental EOF newline.
	a = strings.TrimSuffix(a, "\n")
	b = strings.TrimSuffix(b, "\n")
	a_lines := strings.Split(a, "\n")
	b_lines := strings.Split(b, "\n")
	var msg_fun func() string
	if len(a_lines) != len(b_lines) {
		msg_fun = func() string {
			return fmt.Sprintf("expected '%d' lines, got '%d'", len(a_lines), len(b_lines))
		}
	} else {
		var mismatches []string
		for i := range a_lines {
			if a_lines[i] != b_lines[i] {
				mismatches = append(mismatches, fmt.Sprintf("line %d: expected '%s', got '%s'", i+1, a_lines[i], b_lines[i]))
			}
		}
		if len(mismatches) > 0 {
			msg_fun = func() string { return strings.Join(mismatches, "; ") }
		}
	}
	if msg_fun == nil {
		return
	}
	file, line := get_parent_info(1)
	msg := args_to_message(msg_fun, args)
	if loc, err := loc_str(file, line); err != nil {
		t.Errorf(msg+" in %s:%d", file, line)
	} else {
		t.Errorf("%s in %s", msg, loc)
	}
}

// HasKey asserts that m contains key k.
func HasKey[K comparable, V any](t testing.TB, m map[K]V, k K, args ...any) {
	t.Helper()
	if _, ok := m[k]; ok {
		return
	}
	file, line := get_parent_info(1)
	msg := args_to_message(func() string {
		keys := make([]K, 0, len(m))
		for kk := range m {
			keys = append(keys, kk)
		}
		return fmt.Sprintf("expected key '%v' (%T) in map, got keys %v", k, k, keys)
	}, args)
	if loc, err := loc_str(file, line); err != nil {
		t.Errorf(msg+" in %s:%d", file, line)
	} else {
		t.Errorf("%s in %s", msg, loc)
	}
}

func ContainsString(t testing.TB, haystack string, needle string, args ...any) {
	t.Helper()
	if strings.Contains(haystack, needle) {
		return
	}
	file, line := get_parent_info(1)
	msg := args_to_message(func() string {
		return fmt.Sprintf("expected needle string '%s' to be in a haystack string '%s'", needle, haystack)
	}, args)
	if loc, err := loc_str(file, line); err != nil {
		t.Errorf(msg+" in %s:%d", file, line)
	} else {
		t.Errorf("%s in %s", msg, loc)
	}
}

func Panic(t testing.TB, f func(), f_recover func(t testing.TB, rec any), args ...any) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			if f_recover != nil {
				f_recover(t, r)
			}
			return
		}
		assert(t, nested_assert_parent+1, false, []any{"expected panic, but no panic occurred"})
	}()
	f()
}

// Eventually polls predicate until it returns true or timeout elapses.
// Sleeps interval between checks. Predicate is called at least once
// (even with zero timeout). Fails the test if predicate never returns
// true within the deadline.
func Eventually(t testing.TB, predicate func() bool, timeout time.Duration, interval time.Duration, args ...any) {
	t.Helper()
	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		if predicate() {
			return
		}
		time.Sleep(interval)
	}
	if predicate() {
		return
	}
	file, line := get_parent_info(1)
	msg := args_to_message(func() string {
		return fmt.Sprintf("predicate did not become true within %v (poll interval %v)", timeout, interval)
	}, args)
	if loc, err := loc_str(file, line); err != nil {
		t.Errorf(msg+" in %s:%d", file, line)
	} else {
		t.Errorf("%s in %s", msg, loc)
	}
}
