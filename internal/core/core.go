// Package core holds the shared implementation of the assertion
// primitives. Each primitive takes a [Failer] (a method-value of
// either [testing.TB.Errorf] or [testing.TB.Fatalf]); the public
// wrapper packages (assert, require) pass the appropriate one.
//
// Internal to the module; not for direct downstream use.
package core

import (
	"errors"
	"fmt"
	"os"
	"reflect"
	"runtime"
	"strings"
	"sync"
)

// Failer is the failure-reporting function passed in by the wrapper
// package. assert.* passes t.Errorf (soft); require.* passes t.Fatalf
// (hard, calls runtime.Goexit on the test goroutine).
type Failer func(format string, args ...any)

// Numeric covers built-in integer and float kinds (and named types
// based on them). Used as the constraint for NearlyEqual.
type Numeric interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr |
		~float32 | ~float64
}

type anyErr struct{}

func (anyErr) Error() string { return "<any error>" }

// AnyError sentinel: matches any non-nil error when passed as the
// expected arg to Error. Defined here so both assert and require
// re-export the same value -- identity-based sentinel works across
// both packages.
var AnyError error = anyErr{}

// GetParentInfo returns the file and line of the caller N frames above
// the function that called GetParentInfo. N=2 attributes to the test
// when the call chain is test -> wrapper -> core.Primitive ->
// GetParentInfo.
func GetParentInfo(N int) (string, int) {
	// Use Caller's own file/line: it adjusts the pc for us, whereas
	// FuncForPC(pc).FileLine(pc) on the raw pc can attribute to the
	// line after the call instruction.
	_, file, line, _ := runtime.Caller(1 + N)
	return file, line
}

// ArgsToMessage converts the variadic args ...any tail of an assertion
// call into a message. If args[0] is a string it's treated as a format
// string (Sprintf semantics); otherwise args are stringified as a
// whole. With no args, default_func is invoked to produce the message.
func ArgsToMessage(default_func func() string, args []any) string {
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

// DescribeErr formats an error for failure messages. Suppresses the
// universal *errors.errorString / *fmt.wrapError type tags as noise;
// keeps the type for custom error types where it's informative.
func DescribeErr(e error) string {
	t := fmt.Sprintf("%T", e)
	if t == "*errors.errorString" || t == "*fmt.wrapError" {
		return fmt.Sprintf("'%v'", e)
	}
	return fmt.Sprintf("'%v' (%s)", e, t)
}

// DescribeNonNil formats a non-nil value for failure messages.
// For pointers it shows the pointed-to value rather than the address.
func DescribeNonNil(x any) string {
	v := reflect.ValueOf(x)
	if v.Kind() == reflect.Pointer && !v.IsNil() {
		return fmt.Sprintf("'%v' (%T)", v.Elem().Interface(), x)
	}
	return fmt.Sprintf("'%v' (%T)", x, x)
}

// IsNil handles the typed-nil-in-interface case: var p *T = nil; var i any = p
// -- `i != nil` is true but the underlying value is nil.
func IsNil(x any) bool {
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

// sourceCache memoises file contents split into lines, keyed by path.
// Never evicted -- fine for test processes (bounded set of source files,
// short-lived), would grow unbounded in a long-running daemon.
// Concurrent first-read of same file may ReadFile twice; last Store wins.
// Harmless -- both reads produce identical content.
var sourceCache sync.Map // map[string]sourceEntry

// sourceEntry caches the outcome of a source read -- lines on success,
// the read error on failure -- so failed reads are cached explicitly
// rather than as nil lines that happen to trip the range check.
type sourceEntry struct {
	lines []string
	err   error
}

var (
	errNoLocation     = errors.New("no file/line")
	errLineOutOfRange = errors.New("line out of range")
)

func loadSource(file string) ([]string, error) {
	if v, ok := sourceCache.Load(file); ok {
		e := v.(sourceEntry)
		return e.lines, e.err
	}
	data, err := os.ReadFile(file)
	if err != nil {
		sourceCache.Store(file, sourceEntry{err: err})
		return nil, err
	}
	lines := strings.Split(string(data), "\n")
	sourceCache.Store(file, sourceEntry{lines: lines})
	return lines, nil
}

// getSourceSnippet returns the source line at `line`, plus any
// continuation lines, until brace/paren/bracket depth returns to zero.
func getSourceSnippet(file string, line int) (string, error) {
	if file == "" || line <= 0 {
		return "", errNoLocation
	}
	lines, err := loadSource(file)
	if err != nil {
		return "", err
	}
	if line > len(lines) {
		return "", errLineOutOfRange
	}
	depth := 0
	inStr := false
	var quote byte
	var out []string
	for i := line - 1; i < len(lines); i++ {
		s := lines[i]
		out = append(out, strings.TrimSpace(s))
		// Only raw (backtick) strings span lines in Go; a `"` or `'`
		// left open at end-of-line (e.g. an apostrophe in a comment)
		// must not leak into the next line.
		if inStr && quote != '`' {
			inStr = false
		}
		for j := 0; j < len(s); j++ {
			c := s[j]
			if inStr {
				if c == '\\' {
					j++
					continue
				}
				if c == quote {
					inStr = false
				}
				continue
			}
			switch c {
			case '"', '\'', '`':
				inStr = true
				quote = c
			case '(', '[', '{':
				depth++
			case ')', ']', '}':
				depth--
			}
		}
		if depth <= 0 {
			break
		}
	}
	return strings.Join(out, "\n"), nil
}

// LocStr formats "file:line" with the source snippet on following
// indented lines. Returns an error if the source could not be read;
// callers should fall back to a plain "file:line" format.
func LocStr(file string, line int) (string, error) {
	src, err := getSourceSnippet(file, line)
	if err != nil {
		return "", err
	}
	indented := strings.ReplaceAll(src, "\n", "\n  > ")
	return fmt.Sprintf("%s:%d\n  > %s", file, line, indented), nil
}

// emit is the shared tail used by every primitive: produce the
// formatted message, look up the call-site source snippet, and call
// fail with the assembled string.
func emit(fail Failer, file string, line int, msg string) {
	if loc, err := LocStr(file, line); err != nil {
		fail("%s in %s:%d", msg, file, line)
	} else {
		fail("%s in %s", msg, loc)
	}
}
