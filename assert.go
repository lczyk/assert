package assert

import (
	"fmt"
	"reflect"
	"regexp"
	"runtime"
	"sync"
	"testing"

	"github.com/lczyk/assert/compare"
)

// regexCache memoizes compiled regex patterns supplied as strings to Error,
// so the success path of repeated assertions avoids recompiling.
var regexCache sync.Map // map[string]*regexp.Regexp

func compile_cached(pattern string) *regexp.Regexp {
	if v, ok := regexCache.Load(pattern); ok {
		return v.(*regexp.Regexp)
	}
	re := regexp.MustCompile(pattern)
	regexCache.Store(pattern, re)
	return re
}

// fail_here reports msg with the source location of the caller N frames up.
// Use this from the success-path-fast helpers so we only build the message
// (Sprintf, slice literal, etc.) on actual failure.
func fail_here(t testing.TB, N int, msg string) {
	t.Helper()
	file, line := get_parent_info(N + 1)
	if loc, err := loc_str(file, line); err != nil {
		t.Errorf(msg+" in %s:%d", file, line)
	} else {
		t.Errorf("%s in %s", msg, loc)
	}
}

func get_parent_info(N int) (string, int) {
	parent, _, _, _ := runtime.Caller(1 + N)
	return runtime.FuncForPC(parent).FileLine(parent)
}

// convert 'args ...any' to the assertion message
// internal utility so we don't use variadics to make the calls a bit more consistent
func args_to_message(default_func func() string, args []any) string {
	var msg string
	if len(args) == 0 {
		msg = default_func()
	} else {
		switch args[0].(type) {
		case string:
			msg = args[0].(string)
			msg = fmt.Sprintf(msg, args[1:]...)
		default:
			msg = fmt.Sprintf("%v", args)
		}
	}
	return msg
}

type anyErr struct{}

func (anyErr) Error() string { return "<any error>" }

const nested_assert_parent = 2

func assert(t testing.TB, N int, predicate bool, args []any) {
	t.Helper()
	if !predicate {
		file, line := get_parent_info(N)
		msg := args_to_message(func() string { return "assertion failed" }, args)
		if loc, err := loc_str(file, line); err != nil {
			t.Errorf(msg+" in %s:%d", file, line)
		} else {
			t.Errorf("%s in %s", msg, loc)
		}
	}
}

func assert_error(t testing.TB, N int, err error, expected any, args []any) {
	t.Helper()
	var msg_fun func() string

	// AnyError sentinel: any non-nil err passes; nil err fails.
	if e, ok := expected.(error); ok && e == AnyError {
		if err == nil {
			msg_fun = func() string { return "expected an error, got nil" }
		}
		if msg_fun != nil {
			msg := args_to_message(msg_fun, args)
			file, line := get_parent_info(N)
			if loc, err := loc_str(file, line); err != nil {
				t.Errorf(msg+" in %s:%d", file, line)
			} else {
				t.Errorf("%s in %s", msg, loc)
			}
		}
		return
	}

	switch expected := expected.(type) {
	case string:
		if err == nil {
			msg_fun = func() string {
				return fmt.Sprintf("expected error to match '%s', got no error (nil)", expected)
			}
		} else {
			// Regex pattern matched as substring against err.Error().
			re := compile_cached(expected)
			if !re.MatchString(err.Error()) {
				msg_fun = func() string {
					return fmt.Sprintf("expected error to match '%s', got %s", expected, describe_err(err))
				}
			}
		}

	case error:
		if expected == nil {
			if err != nil {
				msg_fun = func() string {
					return fmt.Sprintf("expected no error, got %s", describe_err(err))
				}
			}
		} else {
			if err == nil {
				msg_fun = func() string {
					return fmt.Sprintf("expected error %s, got no error (nil)", describe_err(expected))
				}
			} else {
				if !compare.Errors(err, expected) && !compare.ErrorsIs(err, expected) {
					msg_fun = func() string {
						return fmt.Sprintf("expected error %s, got %s", describe_err(expected), describe_err(err))
					}
				}
			}
		}
	case nil:
		if err != nil {
			msg_fun = func() string {
				return fmt.Sprintf("expected no error, got %s", describe_err(err))
			}
		}
	case *regexp.Regexp:
		if err == nil {
			msg_fun = func() string {
				return fmt.Sprintf("expected error '%v' (%T), got no error (nil)", expected, expected)
			}
		} else {
			if !expected.MatchString(err.Error()) {
				msg_fun = func() string {
					return fmt.Sprintf("expected error to match '%s', got %s", expected, describe_err(err))
				}
			}
		}
	default:
		panic("expected type is not an error or string")

	}

	if msg_fun != nil {
		msg := args_to_message(msg_fun, args)
		file, line := get_parent_info(N)
		if loc, err := loc_str(file, line); err != nil {
			t.Errorf(msg+" in %s:%d", file, line)
		} else {
			t.Errorf("%s in %s", msg, loc)
		}
	}
}

func equal_cmp[T any](t testing.TB, N int, a T, b T, comparator func(T, T) bool, args []any) {
	t.Helper()
	if comparator(a, b) {
		return
	}
	file, line := get_parent_info(N)
	msg := args_to_message(func() string {
		return fmt.Sprintf("expected '%v' (%T) == '%v' (%T)", a, a, b, b)
	}, args)
	if loc, err := loc_str(file, line); err != nil {
		t.Errorf(msg+" in %s:%d", file, line)
	} else {
		t.Errorf("%s in %s", msg, loc)
	}
}

func equal_cmp_any(t testing.TB, N int, a any, b any, comparator func(any, any) bool, args []any) {
	defer func() {
		if r := recover(); r != nil {
			// If the comparator panics, we want to catch it and report it as a test failure.
			file, line := get_parent_info(4)
			if loc, err := loc_str(file, line); err != nil {
				t.Errorf("Comparator panicked: %v in %s:%d", r, file, line)
			} else {
				t.Errorf("Comparator panicked: %v in %s", r, loc)
			}
		}
	}()
	t.Helper()
	if comparator(a, b) {
		return
	}
	file, line := get_parent_info(N)
	msg := args_to_message(func() string {
		return fmt.Sprintf("expected '%v' (%T) == '%v' (%T)", a, a, b, b)
	}, args)
	if loc, err := loc_str(file, line); err != nil {
		t.Errorf(msg+" in %s:%d", file, line)
	} else {
		t.Errorf("%s in %s", msg, loc)
	}
}

// describe_err formats an error for failure messages. Suppresses the
// universal *errors.errorString / *fmt.wrapError type tags as noise;
// keeps the type for custom error types where it's informative.
func describe_err(e error) string {
	t := fmt.Sprintf("%T", e)
	if t == "*errors.errorString" || t == "*fmt.wrapError" {
		return fmt.Sprintf("'%v'", e)
	}
	return fmt.Sprintf("'%v' (%s)", e, t)
}

// describe_non_nil formats a non-nil value for failure messages.
// For pointers it shows the pointed-to value rather than the address,
// which is more useful when debugging assertions.
func describe_non_nil(x any) string {
	v := reflect.ValueOf(x)
	if v.Kind() == reflect.Pointer && !v.IsNil() {
		return fmt.Sprintf("'%v' (%T)", v.Elem().Interface(), x)
	}
	return fmt.Sprintf("'%v' (%T)", x, x)
}

// is_nil handles the typed-nil-in-interface case: var p *T = nil; var i any = p
// — `i != nil` is true but the underlying value is nil.
func is_nil(x any) bool {
	if x == nil {
		return true
	}
	v := reflect.ValueOf(x)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice, reflect.UnsafePointer:
		return v.IsNil()
	}
	return false
}

// Check that the type of obj is T.
func assert_type[T any](t testing.TB, N int, obj any, args ...any) T {
	t.Helper()
	if obj_T, ok := obj.(T); ok {
		return obj_T
	} else {
		file, line := get_parent_info(N)
		msg := args_to_message(func() string {
			return fmt.Sprintf("expected type %s, got %T", reflect.TypeOf((*T)(nil)).Elem(), obj)
		}, args)
		if loc, err := loc_str(file, line); err != nil {
			t.Errorf(msg+" in %s:%d", file, line)
		} else {
			t.Errorf("%s in %s", msg, loc)
		}
	}
	return *new(T)
}
