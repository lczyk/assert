package core

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/lczyk/assert/compare"
)

// Each primitive below assumes the call chain:
//
//   test -> public-wrapper -> core.Primitive -> GetParentInfo
//
// so depth N=2 attributes failures to the test source line.

func That(t testing.TB, fail Failer, predicate bool, args []any) {
	t.Helper()
	if predicate {
		return
	}
	file, line := GetParentInfo(2)
	msg := ArgsToMessage(func() string { return "assertion failed" }, args)
	emit(fail, file, line, msg)
}

func Equal[T comparable](t testing.TB, fail Failer, a, b T, args []any) {
	t.Helper()
	if a == b {
		return
	}
	file, line := GetParentInfo(2)
	msg := ArgsToMessage(func() string {
		return fmt.Sprintf("expected '%v' (%T) == '%v' (%T)", a, a, b, b)
	}, args)
	emit(fail, file, line, msg)
}

func NotEqual[T comparable](t testing.TB, fail Failer, a, b T, args []any) {
	t.Helper()
	if a != b {
		return
	}
	file, line := GetParentInfo(2)
	msg := ArgsToMessage(func() string {
		return fmt.Sprintf("expected '%v' (%T) != '%v' (%T)", a, a, b, b)
	}, args)
	emit(fail, file, line, msg)
}

func NearlyEqual[T Numeric](t testing.TB, fail Failer, got, want, tolerance T, args []any) {
	t.Helper()
	var diff T
	if got >= want {
		diff = got - want
	} else {
		diff = want - got
	}
	// diff is larger-minus-smaller, so it is negative only if the signed
	// subtraction wrapped; that is a failure, not a pass.
	if diff >= 0 && diff <= tolerance {
		return
	}
	file, line := GetParentInfo(2)
	msg := ArgsToMessage(func() string {
		if diff < 0 {
			return fmt.Sprintf("expected '%v' nearly equal to '%v' (tolerance %v), but their difference overflows %T", got, want, tolerance, diff)
		}
		return fmt.Sprintf("expected '%v' nearly equal to '%v' (tolerance %v), got diff %v", got, want, tolerance, diff)
	}, args)
	emit(fail, file, line, msg)
}

func NoError(t testing.TB, fail Failer, err error, args []any) {
	t.Helper()
	if err == nil {
		return
	}
	file, line := GetParentInfo(2)
	msg := ArgsToMessage(func() string {
		return fmt.Sprintf("expected no error, got %s", DescribeErr(err))
	}, args)
	emit(fail, file, line, msg)
}

// Error matches err against expected. See assert.Error / require.Error
// godoc for the full type-switch on expected.
func Error(t testing.TB, fail Failer, err error, expected any, args []any) {
	t.Helper()
	var msg_fun func() string

	if e, ok := expected.(error); ok && e == AnyError {
		if err == nil {
			msg_fun = func() string { return "expected an error, got nil" }
		}
		if msg_fun != nil {
			file, line := GetParentInfo(2)
			emit(fail, file, line, ArgsToMessage(msg_fun, args))
		}
		return
	}

	switch expected := expected.(type) {
	case string:
		if err == nil {
			msg_fun = func() string {
				return fmt.Sprintf("expected error to contain '%s', got no error (nil)", expected)
			}
		} else if IsNil(err) || !strings.Contains(err.Error(), expected) {
			msg_fun = func() string {
				return fmt.Sprintf("expected error to contain '%s', got %s", expected, DescribeErr(err))
			}
		}
	case error:
		// IsNil, not == nil: a typed-nil error (e.g. (*MyErr)(nil) in the
		// interface) must take the no-error branch rather than fall through
		// to DescribeErr, which would call Error() on a nil receiver.
		if IsNil(expected) {
			if err != nil {
				msg_fun = func() string {
					return fmt.Sprintf("expected no error, got %s", DescribeErr(err))
				}
			}
		} else {
			if err == nil {
				msg_fun = func() string {
					return fmt.Sprintf("expected error %s, got no error (nil)", DescribeErr(expected))
				}
			} else {
				if !compare.Errors(err, expected) && !compare.ErrorsIs(err, expected) {
					msg_fun = func() string {
						return fmt.Sprintf("expected error %s, got %s", DescribeErr(expected), DescribeErr(err))
					}
				}
			}
		}
	case nil:
		if err != nil {
			msg_fun = func() string {
				return fmt.Sprintf("expected no error, got %s", DescribeErr(err))
			}
		}
	case *regexp.Regexp:
		if expected == nil {
			panic("Error: expected is a nil *regexp.Regexp")
		}
		if err == nil {
			msg_fun = func() string {
				return fmt.Sprintf("expected error '%v' (%T), got no error (nil)", expected, expected)
			}
		} else if IsNil(err) || !expected.MatchString(err.Error()) {
			msg_fun = func() string {
				return fmt.Sprintf("expected error to match '%s', got %s", expected, DescribeErr(err))
			}
		}
	default:
		panic(fmt.Sprintf("Error: expected must be nil, error, string, or *regexp.Regexp, got %T", expected))
	}

	if msg_fun != nil {
		file, line := GetParentInfo(2)
		emit(fail, file, line, ArgsToMessage(msg_fun, args))
	}
}

func ErrorIs(t testing.TB, fail Failer, err, expected error, args []any) {
	t.Helper()
	if compare.ErrorsIs(err, expected) {
		return
	}
	file, line := GetParentInfo(2)
	msg := ArgsToMessage(func() string {
		if err == nil {
			return fmt.Sprintf("expected error %s, got no error (nil)", DescribeErr(expected))
		}
		if expected == nil {
			return fmt.Sprintf("expected no error, got %s", DescribeErr(err))
		}
		return fmt.Sprintf("expected errors.Is('%v', '%v') to be true, got %s", err, expected, DescribeErr(err))
	}, args)
	emit(fail, file, line, msg)
}

func EqualCmp[T any](t testing.TB, fail Failer, a, b T, comparator func(T, T) bool, args []any) {
	// Comparator panic protection: recover and report. Capture file/line
	// outside the defer so depth stays consistent. Mirrors EqualCmpAny.
	file, line := GetParentInfo(2)
	defer func() {
		if r := recover(); r != nil {
			emit(fail, file, line, fmt.Sprintf("Comparator panicked: %v", r))
		}
	}()
	t.Helper()
	if comparator(a, b) {
		return
	}
	msg := ArgsToMessage(func() string {
		return fmt.Sprintf("expected '%v' (%T) == '%v' (%T)", a, a, b, b)
	}, args)
	emit(fail, file, line, msg)
}

func EqualCmpAny(t testing.TB, fail Failer, a, b any, comparator func(any, any) bool, args []any) {
	// Comparator panic protection: recover and report. Capture file/line
	// outside the defer so depth stays consistent.
	file, line := GetParentInfo(2)
	defer func() {
		if r := recover(); r != nil {
			emit(fail, file, line, fmt.Sprintf("Comparator panicked: %v", r))
		}
	}()
	t.Helper()
	if comparator(a, b) {
		return
	}
	msg := ArgsToMessage(func() string {
		return fmt.Sprintf("expected '%v' (%T) == '%v' (%T)", a, a, b, b)
	}, args)
	emit(fail, file, line, msg)
}

func EqualArrays[T comparable](t testing.TB, fail Failer, a, b []T, args []any) {
	t.Helper()
	if compare.Arrays(a, b) {
		return
	}
	file, line := GetParentInfo(2)
	msg := ArgsToMessage(func() string {
		return fmt.Sprintf("expected '%v' (%T) == '%v' (%T)", a, a, b, b)
	}, args)
	emit(fail, file, line, msg)
}

func EqualMaps[K, V comparable](t testing.TB, fail Failer, a, b map[K]V, args []any) {
	t.Helper()
	if compare.Maps(a, b) {
		return
	}
	file, line := GetParentInfo(2)
	msg := ArgsToMessage(func() string {
		return fmt.Sprintf("expected '%v' (%T) == '%v' (%T)", a, a, b, b)
	}, args)
	emit(fail, file, line, msg)
}

func EqualArraysUnordered[T comparable](t testing.TB, fail Failer, a, b []T, args []any) {
	t.Helper()
	if compare.ArraysUnordered(a, b) {
		return
	}
	file, line := GetParentInfo(2)
	msg := ArgsToMessage(func() string {
		return fmt.Sprintf("expected '%v' (%T) == '%v' (%T)", a, a, b, b)
	}, args)
	emit(fail, file, line, msg)
}

func Nil(t testing.TB, fail Failer, x any, args []any) {
	t.Helper()
	if IsNil(x) {
		return
	}
	file, line := GetParentInfo(2)
	msg := ArgsToMessage(func() string { return fmt.Sprintf("expected nil, got %s", DescribeNonNil(x)) }, args)
	emit(fail, file, line, msg)
}

func NotNil(t testing.TB, fail Failer, x any, args []any) {
	t.Helper()
	if !IsNil(x) {
		return
	}
	file, line := GetParentInfo(2)
	msg := ArgsToMessage(func() string { return fmt.Sprintf("expected non-nil, got nil (%T)", x) }, args)
	emit(fail, file, line, msg)
}

func Len(t testing.TB, fail Failer, x any, n int, args []any) {
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
	file, line := GetParentInfo(2)
	msg := ArgsToMessage(func() string { return fmt.Sprintf("expected len %d, got len %d: %v", n, got, x) }, args)
	emit(fail, file, line, msg)
}

// Type asserts that obj is of type T and returns the asserted value.
// On mismatch, fail is called and the zero value of T is returned.
// Callers that need a guaranteed-non-zero return (i.e. cannot tolerate
// a nil-deref after a soft failure) should pass t.Fatalf as the fail
// argument -- this is what require.Type does.
func Type[T any](t testing.TB, fail Failer, obj any, args []any) T {
	t.Helper()
	if obj_T, ok := obj.(T); ok {
		return obj_T
	}
	file, line := GetParentInfo(2)
	msg := ArgsToMessage(func() string {
		return fmt.Sprintf("expected type %s, got %T", reflect.TypeOf((*T)(nil)).Elem(), obj)
	}, args)
	emit(fail, file, line, msg)
	return *new(T)
}

func EqualLineByLine(t testing.TB, fail Failer, a, b string, args []any) {
	t.Helper()
	a = strings.TrimSuffix(a, "\n")
	b = strings.TrimSuffix(b, "\n")
	// PERF: walk both strings line-by-line without strings.Split, which
	// would alloc two []string. strings.Count + IndexByte both alloc-free.
	aLines := strings.Count(a, "\n") + 1
	bLines := strings.Count(b, "\n") + 1
	if aLines != bLines {
		file, line := GetParentInfo(2)
		emit(fail, file, line, ArgsToMessage(func() string {
			return fmt.Sprintf("expected '%d' lines, got '%d'", aLines, bLines)
		}, args))
		return
	}
	var mismatches []string
	ai, bi := 0, 0
	for n := 1; n <= aLines; n++ {
		aOff := strings.IndexByte(a[ai:], '\n')
		if aOff < 0 {
			aOff = len(a) - ai
		}
		bOff := strings.IndexByte(b[bi:], '\n')
		if bOff < 0 {
			bOff = len(b) - bi
		}
		aLine := a[ai : ai+aOff]
		bLine := b[bi : bi+bOff]
		if aLine != bLine {
			mismatches = append(mismatches, fmt.Sprintf("line %d: expected '%s', got '%s'", n, aLine, bLine))
		}
		ai += aOff + 1
		bi += bOff + 1
	}
	if len(mismatches) == 0 {
		return
	}
	file, line := GetParentInfo(2)
	emit(fail, file, line, ArgsToMessage(func() string { return strings.Join(mismatches, "; ") }, args))
}

func HasKey[K comparable, V any](t testing.TB, fail Failer, m map[K]V, k K, args []any) {
	t.Helper()
	if _, ok := m[k]; ok {
		return
	}
	file, line := GetParentInfo(2)
	msg := ArgsToMessage(func() string {
		keys := make([]K, 0, len(m))
		for kk := range m {
			keys = append(keys, kk)
		}
		return fmt.Sprintf("expected key '%v' (%T) in map, got keys %v", k, k, keys)
	}, args)
	emit(fail, file, line, msg)
}

func ContainsString(t testing.TB, fail Failer, haystack, needle string, args []any) {
	t.Helper()
	if strings.Contains(haystack, needle) {
		return
	}
	file, line := GetParentInfo(2)
	msg := ArgsToMessage(func() string {
		return fmt.Sprintf("expected needle string '%s' to be in a haystack string '%s'", needle, haystack)
	}, args)
	emit(fail, file, line, msg)
}

// Panic runs f through call so a panic (including panic(nil), which
// recovers as nil on pre-1.21 main modules) and a normal return are
// told apart from f leaving via runtime.Goexit (t.FailNow, t.Skip): in
// that last case call never returns and nothing is reported here.
func Panic(t testing.TB, fail Failer, f func(), f_recover func(t testing.TB, rec any), args []any) {
	t.Helper()
	panicked, rec := call(f)
	if panicked {
		if f_recover != nil {
			f_recover(t, rec)
		}
		return
	}
	file, line := GetParentInfo(2)
	msg := ArgsToMessage(func() string { return "expected panic, but no panic occurred" }, args)
	emit(fail, file, line, msg)
}

// call runs f and reports whether it panicked, with the recovered value.
// If f exits via runtime.Goexit, call does not return.
func call(f func()) (panicked bool, rec any) {
	completed := false
	defer func() {
		if !completed {
			panicked = true
			rec = recover()
		}
	}()
	f()
	completed = true
	return false, nil
}

func Eventually(t testing.TB, fail Failer, predicate func() bool, timeout, interval time.Duration, args []any) {
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
	file, line := GetParentInfo(2)
	msg := ArgsToMessage(func() string {
		return fmt.Sprintf("predicate did not become true within %v (poll interval %v)", timeout, interval)
	}, args)
	emit(fail, file, line, msg)
}
