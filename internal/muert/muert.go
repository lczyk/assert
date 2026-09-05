// micro version of assert to copy/paste into other packages `internal`.
//
// Intentional divergences from github.com/lczyk/assert:
//   - Subset API: only That, Equal, NotEqual, NoError, Error. No EqualCmp,
//     EqualArrays, EqualMaps, Type, Panic, etc.
//   - Error(err, expected) does substring match (strings.Contains), not regex.
//   - NoError uses the generic "assertion failed" default message.
//   - No dependency on the compare subpackage.
package muert

import (
	"fmt"
	"runtime"
	"strings"
	"testing"
)

// License of github.com/lczyk/assert, embedded verbatim so a copy/pasted
// muert.go carries its origin license with it.
const License = `MIT License

Copyright (c) 2025 Marcin Konowalczyk @lczyk

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
`

// Check that the predicate is true, otherwise it fail the test.
func That(t testing.TB, predicate bool, args ...any) {
	t.Helper()
	assert(t, 2, predicate, assertion_failed, args)
}

// Check that two comparable values are equal, otherwise it fail the test.
func Equal[T comparable](t testing.TB, a T, b T) {
	t.Helper()
	assert(t, 2, a == b, func() string {
		return fmt.Sprintf("expected '%v' (%T) == '%v' (%T)", a, a, b, b)
	}, nil)
}

// Check that two comparable values are not equal, otherwise it fail the test.
func NotEqual[T comparable](t testing.TB, a T, b T) {
	t.Helper()
	assert(t, 2, a != b, func() string {
		return fmt.Sprintf("expected '%v' (%T) != '%v' (%T)", a, a, b, b)
	}, nil)
}

// Check that an error is nil.
func NoError(t testing.TB, err error, args ...any) {
	t.Helper()
	assert(t, 2, err == nil, assertion_failed, args)
}

// Check that the error is not nil and contains the expected message.
func Error(t testing.TB, err error, expected string, args ...any) {
	t.Helper()
	if err == nil {
		assert(t, 2, false, func() string {
			return fmt.Sprintf("expected error containing '%s', got nil", expected)
		}, args)
		return
	}
	errs := err.Error()
	assert(t, 2, strings.Contains(errs, expected), func() string {
		return fmt.Sprintf("expected error to contain '%s', got '%s' (%T)", expected, errs, err)
	}, args)
}

func assertion_failed() string { return "assertion failed" }

func get_parent_info(N int) (string, int) {
	// Caller's own file/line is already pc-adjusted; FileLine on the raw
	// pc can point past the call instruction.
	_, file, line, _ := runtime.Caller(1 + N)
	return file, line
}

// convert 'args ...any' to the assertion message: a lone string is used
// verbatim, a string plus further args is a Sprintf format, anything
// else is printed with %v; with no args, default_msg is used.
func args_to_message(default_msg func() string, args []any) string {
	if len(args) == 0 {
		return default_msg()
	}
	if s, ok := args[0].(string); ok {
		if len(args) == 1 {
			return s
		}
		return fmt.Sprintf(s, args[1:]...)
	}
	return fmt.Sprintf("%v", args)
}

func assert(t testing.TB, N int, predicate bool, default_msg func() string, args []any) {
	t.Helper()
	if !predicate {
		file, line := get_parent_info(N)
		t.Errorf("%s in %s:%d", args_to_message(default_msg, args), file, line)
	}
}
