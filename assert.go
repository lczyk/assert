package assert

import (
	"fmt"
	"reflect"
	"regexp"
	"runtime"
	"testing"

	"github.com/lczyk/assert/compare"
)

func getParentInfo(N int) (string, int) {
	parent, _, _, _ := runtime.Caller(1 + N)
	return runtime.FuncForPC(parent).FileLine(parent)
}

// convert 'args ...any' to the assertion message
// internal utility so we don't use variadics to make the calls a bit more consistent
func argsToMessage(default_func func() string, args []any) string {
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

const nestedAssertParent = 2

func assert(t testing.TB, N int, predicate bool, args []any) {
	t.Helper()
	if !predicate {
		file, line := getParentInfo(N)
		msg := argsToMessage(func() string { return "assertion failed" }, args)
		if loc, err := locStr(file, line); err != nil {
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
			msg := argsToMessage(msg_fun, args)
			file, line := getParentInfo(N)
			if loc, err := locStr(file, line); err != nil {
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
			re := regexp.MustCompile(expected)
			if !re.MatchString(err.Error()) {
				msg_fun = func() string {
					return fmt.Sprintf("expected error to match '%s', got %s", expected, describeErr(err))
				}
			}
		}

	case error:
		if expected == nil {
			if err != nil {
				msg_fun = func() string {
					return fmt.Sprintf("expected no error, got %s", describeErr(err))
				}
			}
		} else {
			if err == nil {
				msg_fun = func() string {
					return fmt.Sprintf("expected error %s, got no error (nil)", describeErr(expected))
				}
			} else {
				if !compare.Errors(err, expected) && !compare.ErrorsIs(err, expected) {
					msg_fun = func() string {
						return fmt.Sprintf("expected error %s, got %s", describeErr(expected), describeErr(err))
					}
				}
			}
		}
	case nil:
		if err != nil {
			msg_fun = func() string {
				return fmt.Sprintf("expected no error, got %s", describeErr(err))
			}
		}
	case *regexp.Regexp:
		if err == nil {
			msg_fun = func() string {
				return fmt.Sprintf("expected error '%v' (%T), got no error (nil)", expected, expected)
			}
		} else {
			re := regexp.MustCompile(expected.String())
			if !re.MatchString(err.Error()) {
				msg_fun = func() string {
					return fmt.Sprintf("expected error to match '%s', got %s", expected, describeErr(err))
				}
			}
		}
	default:
		panic("expected type is not an error or string")

	}

	if msg_fun != nil {
		msg := argsToMessage(msg_fun, args)
		file, line := getParentInfo(N)
		if loc, err := locStr(file, line); err != nil {
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
	file, line := getParentInfo(N)
	msg := argsToMessage(func() string {
		return fmt.Sprintf("expected '%v' (%T) == '%v' (%T)", a, a, b, b)
	}, args)
	if loc, err := locStr(file, line); err != nil {
		t.Errorf(msg+" in %s:%d", file, line)
	} else {
		t.Errorf("%s in %s", msg, loc)
	}
}

func equal_cmp_any(t testing.TB, N int, a any, b any, comparator func(any, any) bool, args []any) {
	defer func() {
		if r := recover(); r != nil {
			// If the comparator panics, we want to catch it and report it as a test failure.
			file, line := getParentInfo(4)
			if loc, err := locStr(file, line); err != nil {
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
	file, line := getParentInfo(N)
	msg := argsToMessage(func() string {
		return fmt.Sprintf("expected '%v' (%T) == '%v' (%T)", a, a, b, b)
	}, args)
	if loc, err := locStr(file, line); err != nil {
		t.Errorf(msg+" in %s:%d", file, line)
	} else {
		t.Errorf("%s in %s", msg, loc)
	}
}

// describeErr formats an error for failure messages. Suppresses the
// universal *errors.errorString / *fmt.wrapError type tags as noise;
// keeps the type for custom error types where it's informative.
func describeErr(e error) string {
	t := fmt.Sprintf("%T", e)
	if t == "*errors.errorString" || t == "*fmt.wrapError" {
		return fmt.Sprintf("'%v'", e)
	}
	return fmt.Sprintf("'%v' (%s)", e, t)
}

// describeNonNil formats a non-nil value for failure messages.
// For pointers it shows the pointed-to value rather than the address,
// which is more useful when debugging assertions.
func describeNonNil(x any) string {
	v := reflect.ValueOf(x)
	if v.Kind() == reflect.Pointer && !v.IsNil() {
		return fmt.Sprintf("'%v' (%T)", v.Elem().Interface(), x)
	}
	return fmt.Sprintf("'%v' (%T)", x, x)
}

// isNil handles the typed-nil-in-interface case: var p *T = nil; var i any = p
// — `i != nil` is true but the underlying value is nil.
func isNil(x any) bool {
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
		file, line := getParentInfo(N)
		msg := argsToMessage(func() string {
			return fmt.Sprintf("expected type %s, got %T", reflect.TypeOf((*T)(nil)).Elem(), obj)
		}, args)
		if loc, err := locStr(file, line); err != nil {
			t.Errorf(msg+" in %s:%d", file, line)
		} else {
			t.Errorf("%s in %s", msg, loc)
		}
	}
	return *new(T)
}
