package core

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/lczyk/assert/compare"
)

// Each primitive below assumes the call chain:
//
//   test -> public-wrapper -> core.Primitive -> GetParentInfo
//
// so depth N=2 attributes failures to the test source line. A primitive
// returns the rendered failure message and whether it failed; the
// wrapper, being the frame that calls t.Errorf / t.Fatalf, reports it
// and is the only frame that needs t.Helper().

func That(predicate bool, args []any) (string, bool) {
	if predicate {
		return "", false
	}
	file, line := GetParentInfo(2)
	return render(file, line, ArgsToMessage(func() string { return "assertion failed" }, args)), true
}

func Equal[T comparable](a, b T, args []any) (string, bool) {
	if a == b {
		return "", false
	}
	file, line := GetParentInfo(2)
	return render(file, line, ArgsToMessage(func() string {
		return fmt.Sprintf("expected '%s' (%T) == '%s' (%T)", Fmt(a), a, Fmt(b), b)
	}, args)), true
}

func NotEqual[T comparable](a, b T, args []any) (string, bool) {
	if a != b {
		return "", false
	}
	file, line := GetParentInfo(2)
	return render(file, line, ArgsToMessage(func() string {
		return fmt.Sprintf("expected '%s' (%T) != '%s' (%T)", Fmt(a), a, Fmt(b), b)
	}, args)), true
}

func NearlyEqual[T Numeric](got, want, tolerance T, args []any) (string, bool) {
	var diff T
	if got >= want {
		diff = got - want
	} else {
		diff = want - got
	}
	// diff is larger-minus-smaller, so it is negative only if the signed
	// subtraction wrapped; that is a failure, not a pass.
	if diff >= 0 && diff <= tolerance {
		return "", false
	}
	file, line := GetParentInfo(2)
	return render(file, line, ArgsToMessage(func() string {
		if diff < 0 {
			return fmt.Sprintf("expected '%v' nearly equal to '%v' (tolerance %v), but their difference overflows %T", got, want, tolerance, diff)
		}
		return fmt.Sprintf("expected '%v' nearly equal to '%v' (tolerance %v), got diff %v", got, want, tolerance, diff)
	}, args)), true
}

func NoError(err error, args []any) (string, bool) {
	if err == nil {
		return "", false
	}
	file, line := GetParentInfo(2)
	return render(file, line, ArgsToMessage(func() string {
		return fmt.Sprintf("expected no error, got %s", DescribeErr(err))
	}, args)), true
}

// Error matches err against expected. See assert.Error / require.Error
// godoc for the full type-switch on expected.
func Error(err error, expected any, args []any) (string, bool) {
	var msg_fun func() string

	if e, ok := expected.(error); ok && e == AnyError {
		if err != nil {
			return "", false
		}
		file, line := GetParentInfo(2)
		return render(file, line, ArgsToMessage(func() string { return "expected an error, got nil" }, args)), true
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
		// IsNil, not == nil: a typed-nil expected means no error expected.
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

	if msg_fun == nil {
		return "", false
	}
	file, line := GetParentInfo(2)
	return render(file, line, ArgsToMessage(msg_fun, args)), true
}

func ErrorIs(err, expected error, args []any) (string, bool) {
	if compare.ErrorsIs(err, expected) {
		return "", false
	}
	file, line := GetParentInfo(2)
	return render(file, line, ArgsToMessage(func() string {
		if err == nil {
			return fmt.Sprintf("expected error %s, got no error (nil)", DescribeErr(expected))
		}
		if expected == nil {
			return fmt.Sprintf("expected no error, got %s", DescribeErr(err))
		}
		return fmt.Sprintf("expected errors.Is('%s', '%s') to be true, got %s", Fmt(err), Fmt(expected), DescribeErr(err))
	}, args)), true
}

func EqualCmp[T any](a, b T, comparator func(T, T) bool, args []any) (string, bool) {
	ok, rec, panicked := compareSafely(comparator, a, b)
	if ok {
		return "", false
	}
	file, line := GetParentInfo(2)
	return render(file, line, cmpMessage(a, b, rec, panicked, args)), true
}

func EqualCmpAny(a, b any, comparator func(any, any) bool, args []any) (string, bool) {
	ok, rec, panicked := compareSafely(comparator, a, b)
	if ok {
		return "", false
	}
	file, line := GetParentInfo(2)
	return render(file, line, cmpMessage(a, b, rec, panicked, args)), true
}

// compareSafely calls comparator, turning a panic into a reported result
// so the caller fails the test instead of crashing it. Only the
// comparator runs under the recover, and it is fully unwound before the
// caller looks up its location. If comparator exits via runtime.Goexit,
// compareSafely does not return.
func compareSafely[T any](comparator func(T, T) bool, a, b T) (ok bool, rec any, panicked bool) {
	completed := false
	defer func() {
		if !completed {
			panicked = true
			rec = recover()
		}
	}()
	ok = comparator(a, b)
	completed = true
	return ok, nil, false
}

// cmpMessage builds the EqualCmp failure message. A comparator panic is
// always reported, appended to the custom message when one was given.
func cmpMessage(a, b any, rec any, panicked bool, args []any) string {
	if !panicked {
		return ArgsToMessage(func() string {
			return fmt.Sprintf("expected '%s' (%T) == '%s' (%T)", Fmt(a), a, Fmt(b), b)
		}, args)
	}
	p := fmt.Sprintf("Comparator panicked: %v", rec)
	msg := ArgsToMessage(func() string { return p }, args)
	if len(args) > 0 {
		msg += " (" + p + ")"
	}
	return msg
}

func EqualArrays[T comparable](a, b []T, args []any) (string, bool) {
	if compare.Arrays(a, b) {
		return "", false
	}
	file, line := GetParentInfo(2)
	return render(file, line, ArgsToMessage(func() string {
		return fmt.Sprintf("expected '%s' (%T) == '%s' (%T)", Fmt(a), a, Fmt(b), b)
	}, args)), true
}

func EqualMaps[K, V comparable](a, b map[K]V, args []any) (string, bool) {
	if compare.Maps(a, b) {
		return "", false
	}
	file, line := GetParentInfo(2)
	return render(file, line, ArgsToMessage(func() string {
		return fmt.Sprintf("expected '%s' (%T) == '%s' (%T)", Fmt(a), a, Fmt(b), b)
	}, args)), true
}

func EqualArraysUnordered[T comparable](a, b []T, args []any) (string, bool) {
	if compare.ArraysUnordered(a, b) {
		return "", false
	}
	file, line := GetParentInfo(2)
	return render(file, line, ArgsToMessage(func() string {
		return fmt.Sprintf("expected '%s' (%T) == '%s' (%T)", Fmt(a), a, Fmt(b), b)
	}, args)), true
}

func Nil(x any, args []any) (string, bool) {
	if IsNil(x) {
		return "", false
	}
	file, line := GetParentInfo(2)
	return render(file, line, ArgsToMessage(func() string { return fmt.Sprintf("expected nil, got %s", DescribeNonNil(x)) }, args)), true
}

func NotNil(x any, args []any) (string, bool) {
	if !IsNil(x) {
		return "", false
	}
	file, line := GetParentInfo(2)
	return render(file, line, ArgsToMessage(func() string { return fmt.Sprintf("expected non-nil, got nil (%T)", x) }, args)), true
}

func Len(x any, n int, args []any) (string, bool) {
	v := reflect.ValueOf(x)
	switch v.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
	default:
		panic(fmt.Sprintf("Len: unsupported kind %s", v.Kind()))
	}
	got := v.Len()
	if got == n {
		return "", false
	}
	file, line := GetParentInfo(2)
	return render(file, line, ArgsToMessage(func() string { return fmt.Sprintf("expected len %d, got len %d: %s", n, got, Fmt(x)) }, args)), true
}

// Type asserts that obj is of type T and returns the asserted value. On
// mismatch the zero value of T is returned along with the failure.
func Type[T any](obj any, args []any) (T, string, bool) {
	if obj_T, ok := obj.(T); ok {
		return obj_T, "", false
	}
	file, line := GetParentInfo(2)
	msg := ArgsToMessage(func() string {
		return fmt.Sprintf("expected type %s, got %T", reflect.TypeOf((*T)(nil)).Elem(), obj)
	}, args)
	return *new(T), render(file, line, msg), true
}

func EqualLineByLine(a, b string, args []any) (string, bool) {
	a = strings.TrimSuffix(a, "\n")
	b = strings.TrimSuffix(b, "\n")
	// PERF: walk both strings line-by-line without strings.Split, which
	// would alloc two []string. strings.Count + IndexByte both alloc-free.
	aLines := strings.Count(a, "\n") + 1
	bLines := strings.Count(b, "\n") + 1
	if aLines != bLines {
		file, line := GetParentInfo(2)
		return render(file, line, ArgsToMessage(func() string {
			return fmt.Sprintf("expected '%d' lines, got '%d'", aLines, bLines)
		}, args)), true
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
			mismatches = append(mismatches, fmt.Sprintf("line %d: expected '%s', got '%s'", n, Truncate(aLine), Truncate(bLine)))
		}
		ai += aOff + 1
		bi += bOff + 1
	}
	if len(mismatches) == 0 {
		return "", false
	}
	file, line := GetParentInfo(2)
	return render(file, line, ArgsToMessage(func() string { return strings.Join(mismatches, "; ") }, args)), true
}

func HasKey[K comparable, V any](m map[K]V, k K, args []any) (string, bool) {
	if _, ok := m[k]; ok {
		return "", false
	}
	file, line := GetParentInfo(2)
	return render(file, line, ArgsToMessage(func() string {
		// Sorted by their printed form so the message is stable across
		// runs; K is only comparable, not ordered.
		keys := make([]string, 0, len(m))
		for kk := range m {
			keys = append(keys, fmt.Sprint(kk))
		}
		sort.Strings(keys)
		return fmt.Sprintf("expected key '%v' (%T) in map, got keys %s", k, k, Fmt(keys))
	}, args)), true
}

func ContainsString(haystack, needle string, args []any) (string, bool) {
	if strings.Contains(haystack, needle) {
		return "", false
	}
	file, line := GetParentInfo(2)
	return render(file, line, ArgsToMessage(func() string {
		return fmt.Sprintf("expected needle string '%s' to be in a haystack string '%s'", Truncate(needle), Truncate(haystack))
	}, args)), true
}

// Panic runs f through call so a panic (including panic(nil), which
// recovers as nil on pre-1.21 main modules) and a normal return are
// told apart from f leaving via runtime.Goexit (t.FailNow, t.Skip): in
// that last case call never returns and nothing is reported. t is only
// handed on to f_recover.
func Panic(t testing.TB, f func(), f_recover func(t testing.TB, rec any), args []any) (string, bool) {
	panicked, rec := call(f)
	if panicked {
		if f_recover != nil {
			f_recover(t, rec)
		}
		return "", false
	}
	file, line := GetParentInfo(2)
	return render(file, line, ArgsToMessage(func() string { return "expected panic, but no panic occurred" }, args)), true
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

func Eventually(predicate func() bool, timeout, interval time.Duration, args []any) (string, bool) {
	deadline := time.Now().Add(timeout)
	for {
		if predicate() {
			return "", false
		}
		remaining := time.Until(deadline)
		if remaining <= 0 {
			break
		}
		if interval < remaining {
			time.Sleep(interval)
		} else {
			time.Sleep(remaining)
		}
	}
	file, line := GetParentInfo(2)
	return render(file, line, ArgsToMessage(func() string {
		return fmt.Sprintf("predicate did not become true within %v (poll interval %v)", timeout, interval)
	}, args)), true
}
