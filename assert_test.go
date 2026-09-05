package assert_test

import (
	"fmt"
	"math"
	"regexp"
	"runtime"
	"testing"
	"time"

	"github.com/lczyk/assert"
)

// get the file and line number of the line above the call to this function
func getAboveLineInfo(N int) (string, int) {
	parent, _, _, _ := runtime.Caller(1)
	info := runtime.FuncForPC(parent)
	file, line := info.FileLine(parent)
	return file, line - 1 - N
}

func TestThat(t *testing.T) {
	assert.That(t, true)
}

type myThing interface {
	SomeBehaviour()
}

type myThingImpl struct{}

func (m *myThingImpl) SomeBehaviour() {}

var _ myThing = &myThingImpl{}

func TestType(t *testing.T) {
	t.Run("fails on mismatch", func(t *testing.T) {
		tt := &myT{}
		assert.That(t, !tt.Failed())
		var x int = 1
		y := assert.Type[myThing](tt, x)
		_ = y
		assert.That(t, tt.Failed(), "Expected test to fail, but it did not")
		assert.ContainsString(t, tt.message, "expected type")
	})
	t.Run("succeeds", func(t *testing.T) {
		tt := &myT{}
		assert.That(t, !tt.Failed())
		x := &myThingImpl{}
		y := assert.Type[myThing](tt, x)
		_ = y
		assert.That(t, !tt.Failed(), "Expected test to not fail, but it did")
	})
}

type myT struct {
	testing.T
	message string // latest error message
}

func (t *myT) Errorf(format string, args ...any) {
	t.message = fmt.Sprintf(format, args...)
	t.Fail()
}

// Fatalf override: capture the message and mark Fail without invoking
// runtime.Goexit (real testing.T.Fatalf would unwind the goroutine,
// which breaks tests that assert on post-Fatalf state).
func (t *myT) Fatalf(format string, args ...any) {
	t.message = fmt.Sprintf(format, args...)
	t.Fail()
}

var _ testing.TB = &myT{}

func TestNoError(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		tt := &myT{}
		assert.NoError(tt, nil)
		assert.That(t, !tt.Failed(), "Expected no error, but got one")
		assert.That(t, tt.message == "", "Expected no error message, but got one: %s", tt.message)
	})

	t.Run("non-nil error", func(t *testing.T) {
		tt := &myT{}
		err := fmt.Errorf("this is an error")

		assert.NoError(tt, err)
		file, line := getAboveLineInfo(0)

		assert.That(t, tt.Failed(), "Expected test to fail, but it did not")
		assert.ContainsString(t, tt.message, "this is an error")
		assert.ContainsString(t, tt.message, "in "+file+":"+fmt.Sprint(line))
	})

	t.Run("non-nil error with args", func(t *testing.T) {
		tt := &myT{}
		err := fmt.Errorf("this is an error")

		assert.NoError(tt, err, "we expected no error, but got one: %d", 42)
		file, line := getAboveLineInfo(0)

		assert.That(t, tt.Failed(), "Expected test to fail, but it did not")
		assert.ContainsString(t, tt.message, "we expected no error, but got one: 42")
		assert.ContainsString(t, tt.message, "in "+file+":"+fmt.Sprint(line))
	})
}

func TestError(t *testing.T) {
	t.Run("nil error", func(t *testing.T) {
		tt := &myT{}
		assert.Error(tt, nil, "this is an error")
		file, line := getAboveLineInfo(0)

		assert.That(t, tt.Failed(), "Expected test to fail, but it did not")
		assert.ContainsString(t, tt.message, "this is an error")
		assert.ContainsString(t, tt.message, "in "+file+":"+fmt.Sprint(line))
	})

	t.Run("nil error with args", func(t *testing.T) {
		tt := &myT{}
		assert.Error(tt, nil, "this is an error", "this is an error with args: %d", 42)
		file, line := getAboveLineInfo(0)
		assert.That(t, tt.Failed(), "Expected test to fail, but it did not")
		assert.ContainsString(t, tt.message, "this is an error with args: 42")
		assert.ContainsString(t, tt.message, "in "+file+":"+fmt.Sprint(line))
	})

	t.Run("non-nil error", func(t *testing.T) {
		tt := &myT{}
		err := fmt.Errorf("this is an error")
		assert.Error(tt, err, "this is an error")
		assert.That(t, !tt.Failed(), "Expected test to not fail, but it did")
	})

	t.Run("non-nil error with args", func(t *testing.T) {
		tt := &myT{}
		err := fmt.Errorf("this is an error")
		assert.Error(tt, err, "this is an error")
		assert.That(t, !tt.Failed(), "Expected test to not fail, but it did")
	})

	t.Run("non-nil error with regexp passing", func(t *testing.T) {
		tt := &myT{}
		err := fmt.Errorf("this is an error, also lemons")
		assert.Error(tt, err, regexp.MustCompile("lemons"))
		assert.That(t, !tt.Failed(), "Expected test to not fail, but it did")
	})

	t.Run("non-nil error with regexp failing", func(t *testing.T) {
		tt := &myT{}
		err := fmt.Errorf("this is an error, also lemons")
		assert.Error(tt, err, regexp.MustCompile("oranges"))
		file, line := getAboveLineInfo(0)

		assert.That(t, tt.Failed(), "Expected test to fail, but it did not")
		assert.ContainsString(t, tt.message, "expected error to match 'oranges'")
		assert.ContainsString(t, tt.message, "in "+file+":"+fmt.Sprint(line))
	})
}

func TestEqual(t *testing.T) {
	t.Run("equal", func(t *testing.T) {
		tt := &myT{}
		assert.Equal(tt, 1, 1)
		assert.That(t, !tt.Failed())
	})
	t.Run("not equal", func(t *testing.T) {
		tt := &myT{}
		assert.Equal(tt, 1, 2)
		assert.That(t, tt.Failed())
		assert.ContainsString(t, tt.message, "expected '1'")
		assert.ContainsString(t, tt.message, "'2'")
	})
	t.Run("custom message", func(t *testing.T) {
		tt := &myT{}
		assert.Equal(tt, 1, 2, "case %s", "alpha")
		assert.That(t, tt.Failed())
		assert.ContainsString(t, tt.message, "case alpha")
	})
}

func TestNotEqual(t *testing.T) {
	t.Run("not equal", func(t *testing.T) {
		tt := &myT{}
		assert.NotEqual(tt, 1, 2)
		assert.That(t, !tt.Failed())
	})
	t.Run("equal", func(t *testing.T) {
		tt := &myT{}
		assert.NotEqual(tt, 1, 1)
		assert.That(t, tt.Failed())
	})
	t.Run("custom message", func(t *testing.T) {
		tt := &myT{}
		assert.NotEqual(tt, 1, 1, "case %s", "beta")
		assert.That(t, tt.Failed())
		assert.ContainsString(t, tt.message, "case beta")
	})
}

func TestEqualArrays(t *testing.T) {
	t.Run("equal", func(t *testing.T) {
		tt := &myT{}
		assert.EqualArrays(tt, []int{1, 2, 3}, []int{1, 2, 3})
		assert.That(t, !tt.Failed())
	})
	t.Run("different order", func(t *testing.T) {
		tt := &myT{}
		assert.EqualArrays(tt, []int{1, 2, 3}, []int{3, 2, 1})
		assert.That(t, tt.Failed())
	})
	t.Run("different length", func(t *testing.T) {
		tt := &myT{}
		assert.EqualArrays(tt, []int{1, 2, 3}, []int{1, 2})
		assert.That(t, tt.Failed())
	})
}

func TestEqualArraysUnordered(t *testing.T) {
	t.Run("equal", func(t *testing.T) {
		tt := &myT{}
		assert.EqualArraysUnordered(tt, []int{1, 2, 3}, []int{3, 2, 1})
		assert.That(t, !tt.Failed())
	})
	t.Run("different", func(t *testing.T) {
		tt := &myT{}
		assert.EqualArraysUnordered(tt, []int{1, 2, 3}, []int{1, 2, 4})
		assert.That(t, tt.Failed())
	})
}

func TestEqualMaps(t *testing.T) {
	t.Run("equal", func(t *testing.T) {
		tt := &myT{}
		assert.EqualMaps(tt, map[string]int{"a": 1, "b": 2}, map[string]int{"b": 2, "a": 1})
		assert.That(t, !tt.Failed())
	})
	t.Run("different value", func(t *testing.T) {
		tt := &myT{}
		assert.EqualMaps(tt, map[string]int{"a": 1}, map[string]int{"a": 2})
		assert.That(t, tt.Failed())
	})
	t.Run("different key", func(t *testing.T) {
		tt := &myT{}
		assert.EqualMaps(tt, map[string]int{"a": 1}, map[string]int{"b": 1})
		assert.That(t, tt.Failed())
	})
}

func TestContainsString(t *testing.T) {
	t.Run("contains", func(t *testing.T) {
		tt := &myT{}
		assert.ContainsString(tt, "hello world", "world")
		assert.That(t, !tt.Failed())
	})
	t.Run("does not contain", func(t *testing.T) {
		tt := &myT{}
		assert.ContainsString(tt, "hello world", "lemons")
		assert.That(t, tt.Failed())
		assert.ContainsString(t, tt.message, "lemons")
		assert.ContainsString(t, tt.message, "hello world")
	})
	t.Run("custom message", func(t *testing.T) {
		tt := &myT{}
		assert.ContainsString(tt, "hello world", "lemons", "case %s", "delta")
		assert.That(t, tt.Failed())
		assert.ContainsString(t, tt.message, "case delta")
	})
}

func TestEqualLineByLine(t *testing.T) {
	t.Run("equal single line", func(t *testing.T) {
		tt := &myT{}
		assert.EqualLineByLine(tt, "hello", "hello")
		assert.That(t, !tt.Failed(), "expected pass, got: %s", tt.message)
	})

	t.Run("equal multiline", func(t *testing.T) {
		tt := &myT{}
		assert.EqualLineByLine(tt, "a\nb\nc", "a\nb\nc")
		assert.That(t, !tt.Failed(), "expected pass, got: %s", tt.message)
	})

	t.Run("equal empty", func(t *testing.T) {
		tt := &myT{}
		assert.EqualLineByLine(tt, "", "")
		assert.That(t, !tt.Failed(), "expected pass, got: %s", tt.message)
	})

	t.Run("different line count", func(t *testing.T) {
		tt := &myT{}
		assert.EqualLineByLine(tt, "a\nb", "a\nb\nc")
		assert.That(t, tt.Failed(), "expected fail")
		assert.ContainsString(t, tt.message, "expected '2' lines, got '3'")
	})

	t.Run("trailing newline ignored on one side", func(t *testing.T) {
		// A trailing newline should not cause a spurious failure.
		tt := &myT{}
		assert.EqualLineByLine(tt, "a\n", "a")
		assert.That(t, !tt.Failed(), "trailing newline should be ignored, got: %s", tt.message)
	})

	t.Run("trailing newline ignored multiline", func(t *testing.T) {
		tt := &myT{}
		assert.EqualLineByLine(tt, "a\nb\nc\n", "a\nb\nc")
		assert.That(t, !tt.Failed(), "trailing newline should be ignored, got: %s", tt.message)
	})

	t.Run("trailing newlines on both sides", func(t *testing.T) {
		tt := &myT{}
		assert.EqualLineByLine(tt, "a\nb\n", "a\nb\n")
		assert.That(t, !tt.Failed(), "expected pass, got: %s", tt.message)
	})

	t.Run("genuine extra line not masked by trailing newline rule", func(t *testing.T) {
		// "a\nb\n" is equivalent to "a\nb" (2 lines). "a\nb\nc" is 3 lines.
		// Trailing-newline normalization must NOT also swallow a real extra line.
		tt := &myT{}
		assert.EqualLineByLine(tt, "a\nb\n", "a\nb\nc")
		assert.That(t, tt.Failed(), "expected fail — genuine line count mismatch")
		assert.ContainsString(t, tt.message, "expected '2' lines, got '3'")
	})

	t.Run("empty vs single newline are equal", func(t *testing.T) {
		// Under trailing-newline-ignored semantics, "\n" normalizes to "" —
		// both are zero lines of content.
		tt := &myT{}
		assert.EqualLineByLine(tt, "", "\n")
		assert.That(t, !tt.Failed(), "empty and '\\n' should be equal, got: %s", tt.message)
	})

	t.Run("differing middle line", func(t *testing.T) {
		tt := &myT{}
		assert.EqualLineByLine(tt, "a\nb\nc", "a\nX\nc")
		assert.That(t, tt.Failed(), "expected fail")
		assert.ContainsString(t, tt.message, "line 2: expected 'b', got 'X'")
	})

	t.Run("differing first line", func(t *testing.T) {
		tt := &myT{}
		assert.EqualLineByLine(tt, "x\ny", "a\ny")
		assert.That(t, tt.Failed(), "expected fail")
		assert.ContainsString(t, tt.message, "line 1: expected 'x', got 'a'")
	})

	t.Run("differing last line", func(t *testing.T) {
		tt := &myT{}
		assert.EqualLineByLine(tt, "a\nb\nc", "a\nb\nZ")
		assert.That(t, tt.Failed(), "expected fail")
		assert.ContainsString(t, tt.message, "line 3: expected 'c', got 'Z'")
	})

	t.Run("multiple mismatches reported once", func(t *testing.T) {
		tt := &myT{}
		assert.EqualLineByLine(tt, "a\nb\nc", "X\nb\nZ")
		assert.That(t, tt.Failed(), "expected fail")
		assert.ContainsString(t, tt.message, "line 1: expected 'a', got 'X'")
		assert.ContainsString(t, tt.message, "line 3: expected 'c', got 'Z'")
	})

	t.Run("custom message", func(t *testing.T) {
		tt := &myT{}
		assert.EqualLineByLine(tt, "a\nb", "a\nX", "case %s", "gamma")
		assert.That(t, tt.Failed(), "expected fail")
		assert.ContainsString(t, tt.message, "case gamma")
	})
}

func TestPanic(t *testing.T) {
	t.Run("no panic", func(t *testing.T) {
		tt := &myT{}
		assert.Panic(tt, func() {}, func(t testing.TB, rec any) {
			assert.That(t, false, "We should never call recovery function, because no panic should have happened")
		})
		file, line := getAboveLineInfo(2)

		assert.That(t, tt.Failed(), "Expected test to fail, but it did not")
		assert.ContainsString(t, tt.message, "expected panic, but no panic occurred")
		assert.ContainsString(t, tt.message, "in "+file+":"+fmt.Sprint(line))
	})

	t.Run("panic", func(t *testing.T) {
		tt := &myT{}
		assert.Panic(tt, func() { panic("this is a panic") }, func(t testing.TB, rec any) {
			assert.Equal(t, rec, "this is a panic")
		})

		assert.That(t, !tt.Failed(), "Expected test to not fail, but it did")
	})

	t.Run("panic but failed recovery", func(t *testing.T) {
		tt := &myT{}
		var file string
		var line int
		assert.Panic(tt, func() { panic("this is a panic") }, func(t testing.TB, rec any) {
			assert.Equal(t, rec, "this is not the panic we expected")
			file, line = getAboveLineInfo(0)
		})
		assert.That(t, tt.Failed(), "Expected test to fail, but it did not")
		assert.ContainsString(t, tt.message, "expected 'this is a panic'")
		assert.ContainsString(t, tt.message, "'this is not the panic we expected'")
		assert.ContainsString(t, tt.message, "in "+file+":"+fmt.Sprint(line))
	})

	t.Run("nil f", func(t *testing.T) {
		tt := &myT{}
		assert.Panic(tt, nil, func(t testing.TB, rec any) {
			assert.Equal(t, rec, "this is a panic")
		})
		assert.That(t, tt.Failed(), "Expected test to fail, but it did not")
	})

	t.Run("nil f_recover", func(t *testing.T) {
		tt := &myT{}
		assert.Panic(tt, func() { panic("this is a panic") }, nil)
		assert.That(t, !tt.Failed(), "Expected test to not fail, but it did")
	})

	t.Run("f leaving via Goexit reports nothing", func(t *testing.T) {
		// t.FailNow / t.Skip inside f unwind the goroutine without a panic;
		// Panic must not add a second 'no panic occurred' failure on top.
		tt := &myT{}
		done := make(chan struct{})
		go func() {
			defer close(done)
			assert.Panic(tt, func() { runtime.Goexit() }, nil)
			tt.Errorf("unreachable: Panic returned after Goexit")
		}()
		<-done
		assert.That(t, !tt.Failed(), "expected no failure, got: %s", tt.message)
	})

	t.Run("panic(nil) counts as a panic", func(t *testing.T) {
		tt := &myT{}
		called := false
		assert.Panic(tt, func() { panic(nil) }, func(t testing.TB, rec any) { called = true })
		assert.That(t, !tt.Failed(), "expected pass, got: %s", tt.message)
		assert.That(t, called, "expected the recovery func to be called")
	})
}

func TestEqualCmp(t *testing.T) {
	eqMod10 := func(a, b int) bool { return a%10 == b%10 }
	t.Run("equal", func(t *testing.T) {
		tt := &myT{}
		assert.EqualCmp(tt, 12, 22, eqMod10)
		assert.That(t, !tt.Failed(), "expected pass, got: %s", tt.message)
	})
	t.Run("not equal", func(t *testing.T) {
		tt := &myT{}
		assert.EqualCmp(tt, 12, 23, eqMod10)
		file, line := getAboveLineInfo(0)
		assert.That(t, tt.Failed(), "expected fail")
		assert.ContainsString(t, tt.message, "expected '12'")
		assert.ContainsString(t, tt.message, "'23'")
		assert.ContainsString(t, tt.message, "in "+file+":"+fmt.Sprint(line))
	})
	t.Run("custom message via args", func(t *testing.T) {
		tt := &myT{}
		assert.EqualCmp(tt, 12, 23, eqMod10, "domain mismatch: %d vs %d", 12, 23)
		assert.That(t, tt.Failed(), "expected fail")
		assert.ContainsString(t, tt.message, "domain mismatch: 12 vs 23")
	})
	t.Run("comparator panics", func(t *testing.T) {
		tt := &myT{}
		boom := func(a, b int) bool { panic("boom") }
		assert.EqualCmp(tt, 1, 2, boom)
		file, line := getAboveLineInfo(0)
		assert.That(t, tt.Failed(), "expected fail from panic")
		assert.ContainsString(t, tt.message, "Comparator panicked: boom")
		assert.ContainsString(t, tt.message, "in "+file+":"+fmt.Sprint(line))
	})
	t.Run("comparator panics with custom message keeps both", func(t *testing.T) {
		tt := &myT{}
		boom := func(a, b int) bool { panic("boom") }
		assert.EqualCmp(tt, 1, 2, boom, "widget %d", 7)
		assert.That(t, tt.Failed(), "expected fail from panic")
		assert.ContainsString(t, tt.message, "widget 7")
		assert.ContainsString(t, tt.message, "Comparator panicked: boom")
	})
}

func TestEqualCmpAny(t *testing.T) {
	strEq := func(a, b any) bool { return a.(string) == b.(string) }
	t.Run("equal", func(t *testing.T) {
		tt := &myT{}
		assert.EqualCmpAny(tt, "x", "x", strEq)
		assert.That(t, !tt.Failed(), "expected pass, got: %s", tt.message)
	})
	t.Run("not equal", func(t *testing.T) {
		tt := &myT{}
		assert.EqualCmpAny(tt, "x", "y", strEq)
		assert.That(t, tt.Failed(), "expected fail")
		assert.ContainsString(t, tt.message, "expected 'x'")
		assert.ContainsString(t, tt.message, "'y'")
	})
	t.Run("comparator panics", func(t *testing.T) {
		tt := &myT{}
		// Type-assert int to string - panics inside comparator.
		assert.EqualCmpAny(tt, 1, "y", strEq)
		assert.That(t, tt.Failed(), "expected fail from panic")
		assert.ContainsString(t, tt.message, "Comparator panicked")
	})
}

func TestErrorExpectedAsError(t *testing.T) {
	t.Run("both nil", func(t *testing.T) {
		tt := &myT{}
		var expected error = nil
		assert.Error(tt, nil, expected)
		assert.That(t, !tt.Failed(), "expected pass, got: %s", tt.message)
	})
	t.Run("nil expected, non-nil err", func(t *testing.T) {
		tt := &myT{}
		var expected error = nil
		assert.Error(tt, fmt.Errorf("boom"), expected)
		assert.That(t, tt.Failed(), "expected fail")
		assert.ContainsString(t, tt.message, "expected no error")
	})
	t.Run("non-nil expected, nil err", func(t *testing.T) {
		tt := &myT{}
		expected := fmt.Errorf("boom")
		assert.Error(tt, nil, expected)
		assert.That(t, tt.Failed(), "expected fail")
		assert.ContainsString(t, tt.message, "got no error")
	})
	t.Run("same sentinel", func(t *testing.T) {
		tt := &myT{}
		sentinel := fmt.Errorf("boom")
		assert.Error(tt, sentinel, sentinel)
		assert.That(t, !tt.Failed(), "expected pass, got: %s", tt.message)
	})
	t.Run("wrapped sentinel matches via errors.Is fallback", func(t *testing.T) {
		tt := &myT{}
		sentinel := fmt.Errorf("boom")
		wrapped := fmt.Errorf("context: %w", sentinel)
		assert.Error(tt, wrapped, sentinel)
		assert.That(t, !tt.Failed(), "expected pass via wrap-chain, got: %s", tt.message)
	})
	t.Run("distinct errors with same type and message match structurally", func(t *testing.T) {
		tt := &myT{}
		assert.Error(tt, fmt.Errorf("boom"), fmt.Errorf("boom"))
		assert.That(t, !tt.Failed(), "expected pass, got: %s", tt.message)
	})
	t.Run("mismatched errors", func(t *testing.T) {
		tt := &myT{}
		assert.Error(tt, fmt.Errorf("a"), fmt.Errorf("b"))
		assert.That(t, tt.Failed(), "expected fail")
		assert.ContainsString(t, tt.message, "expected error 'b'")
		assert.ContainsString(t, tt.message, "got 'a'")
	})
	t.Run("typed-nil expected treated as no-error", func(t *testing.T) {
		// (*nilRecvErr)(nil) inside the error interface is non-nil at the
		// interface level; must behave like nil expected, not panic on
		// Error() with a nil receiver.
		var expected *nilRecvErr // nil pointer, non-nil interface once passed
		t.Run("nil err passes", func(t *testing.T) {
			tt := &myT{}
			assert.Error(tt, nil, expected)
			assert.That(t, !tt.Failed(), "expected pass, got: %s", tt.message)
		})
		t.Run("non-nil err fails", func(t *testing.T) {
			tt := &myT{}
			assert.Error(tt, fmt.Errorf("boom"), expected)
			assert.That(t, tt.Failed(), "expected fail")
			assert.ContainsString(t, tt.message, "expected no error")
		})
	})
}

// nilRecvErr's Error() panics on a nil receiver -- used to prove the
// typed-nil expected path never formats the expected error.
type nilRecvErr struct{ msg string }

func (e *nilRecvErr) Error() string { return e.msg }

func TestErrorRegexpNilErr(t *testing.T) {
	tt := &myT{}
	assert.Error(tt, nil, regexp.MustCompile("boom"))
	assert.That(t, tt.Failed(), "expected fail")
	assert.ContainsString(t, tt.message, "got no error")
}

func TestErrorStringNilErrNonEmpty(t *testing.T) {
	tt := &myT{}
	assert.Error(tt, nil, "boom")
	assert.That(t, tt.Failed(), "expected fail")
	assert.ContainsString(t, tt.message, "expected error to contain 'boom'")
	assert.ContainsString(t, tt.message, "got no error (nil)")
}

func TestErrorAnyError(t *testing.T) {
	t.Run("non-nil err passes", func(t *testing.T) {
		tt := &myT{}
		assert.Error(tt, fmt.Errorf("anything"), assert.AnyError)
		assert.That(t, !tt.Failed(), "expected pass, got: %s", tt.message)
	})
	t.Run("empty string is equivalent to AnyError", func(t *testing.T) {
		tt := &myT{}
		assert.Error(tt, fmt.Errorf("anything"), "")
		assert.That(t, !tt.Failed(), "expected pass, got: %s", tt.message)
	})
	t.Run("nil err fails", func(t *testing.T) {
		tt := &myT{}
		assert.Error(tt, nil, assert.AnyError)
		assert.That(t, tt.Failed(), "expected fail")
		assert.ContainsString(t, tt.message, "expected an error, got nil")
	})
	t.Run("AnyError is a distinct error sentinel", func(t *testing.T) {
		// Sanity: AnyError satisfies the error interface but is not equal to
		// arbitrary errors with the same message.
		assert.That(t, assert.AnyError != nil, "AnyError must be non-nil")
		assert.That(t, assert.AnyError.Error() == "<any error>", "AnyError sentinel string")
	})
}

func TestErrorIs(t *testing.T) {
	t.Run("both nil", func(t *testing.T) {
		tt := &myT{}
		assert.ErrorIs(tt, nil, nil)
		assert.That(t, !tt.Failed(), "expected pass, got: %s", tt.message)
	})
	t.Run("same sentinel", func(t *testing.T) {
		tt := &myT{}
		sentinel := fmt.Errorf("boom")
		assert.ErrorIs(tt, sentinel, sentinel)
		assert.That(t, !tt.Failed(), "expected pass, got: %s", tt.message)
	})
	t.Run("wrapped sentinel matches", func(t *testing.T) {
		tt := &myT{}
		sentinel := fmt.Errorf("boom")
		wrapped := fmt.Errorf("ctx: %w", sentinel)
		assert.ErrorIs(tt, wrapped, sentinel)
		assert.That(t, !tt.Failed(), "expected pass, got: %s", tt.message)
	})
	t.Run("distinct errors with same message do not match", func(t *testing.T) {
		tt := &myT{}
		assert.ErrorIs(tt, fmt.Errorf("boom"), fmt.Errorf("boom"))
		assert.That(t, tt.Failed(), "expected fail")
	})
	t.Run("nil err with non-nil expected", func(t *testing.T) {
		tt := &myT{}
		assert.ErrorIs(tt, nil, fmt.Errorf("boom"))
		assert.That(t, tt.Failed(), "expected fail")
		assert.ContainsString(t, tt.message, "got no error")
	})
	t.Run("non-nil err with nil expected", func(t *testing.T) {
		tt := &myT{}
		assert.ErrorIs(tt, fmt.Errorf("boom"), nil)
		assert.That(t, tt.Failed(), "expected fail")
		assert.ContainsString(t, tt.message, "expected no error")
	})
}

func TestErrorInvalidExpectedTypePanics(t *testing.T) {
	tt := &myT{}
	assert.Panic(t, func() {
		assert.Error(tt, fmt.Errorf("x"), 42)
	}, func(t testing.TB, rec any) {
		s, ok := rec.(string)
		assert.That(t, ok, "expected string panic, got %T", rec)
		assert.ContainsString(t, s, "*regexp.Regexp")
		assert.ContainsString(t, s, "got int")
	})
}

func TestErrorNilRegexpPanics(t *testing.T) {
	var re *regexp.Regexp
	assert.Panic(t, func() {
		assert.Error(&myT{}, fmt.Errorf("x"), re)
	}, func(t testing.TB, rec any) {
		s, ok := rec.(string)
		assert.That(t, ok, "expected string panic, got %T", rec)
		assert.ContainsString(t, s, "nil *regexp.Regexp")
	})
}

// A typed-nil err must never have Error() called on it (nilRecvErr's Error()
// dereferences its receiver); it is reported as a typed nil instead.
func TestErrorTypedNilErr(t *testing.T) {
	var p *nilRecvErr
	var err error = p
	t.Run("string expected", func(t *testing.T) {
		tt := &myT{}
		assert.Error(tt, err, "boom")
		assert.That(t, tt.Failed(), "expected fail")
		assert.ContainsString(t, tt.message, "typed-nil error (*assert_test.nilRecvErr)(nil)")
	})
	t.Run("regexp expected", func(t *testing.T) {
		tt := &myT{}
		assert.Error(tt, err, regexp.MustCompile("boom"))
		assert.That(t, tt.Failed(), "expected fail")
		assert.ContainsString(t, tt.message, "typed-nil error")
	})
	t.Run("error expected of the same type", func(t *testing.T) {
		tt := &myT{}
		assert.Error(tt, err, error(&nilRecvErr{msg: "boom"}))
		assert.That(t, tt.Failed(), "expected fail")
		assert.ContainsString(t, tt.message, "typed-nil error")
	})
	t.Run("NoError names the typed nil", func(t *testing.T) {
		tt := &myT{}
		assert.NoError(tt, err)
		assert.That(t, tt.Failed(), "expected fail")
		assert.ContainsString(t, tt.message, "typed-nil error (*assert_test.nilRecvErr)(nil)")
	})
}

func TestThatNonStringFirstArg(t *testing.T) {
	// First arg not a string: args_to_message falls through to default %v path.
	tt := &myT{}
	assert.That(tt, false, 42)
	assert.That(t, tt.Failed(), "expected fail")
	assert.ContainsString(t, tt.message, "42")
}

func TestThatLoneMessageIsLiteral(t *testing.T) {
	// A single string arg is not a format string: a stray % must survive.
	tt := &myT{}
	assert.That(tt, false, "100% sure")
	assert.That(t, tt.Failed(), "expected fail")
	assert.ContainsString(t, tt.message, "100% sure in ")
	// With further args it is a format string, so %% is needed.
	tt2 := &myT{}
	assert.That(tt2, false, "%d%% done", 50)
	assert.ContainsString(t, tt2.message, "50% done in ")
}

func TestThatFailNoArgs(t *testing.T) {
	// Covers the "assertion failed" default-message lambda in assert().
	tt := &myT{}
	assert.That(tt, false)
	assert.That(t, tt.Failed(), "expected fail")
	assert.ContainsString(t, tt.message, "assertion failed")
}

func TestErrorStringLiteralNotRegex(t *testing.T) {
	// String arg is matched literally, not as a regex. Metacharacters are
	// not interpreted: pattern "a+" matches "a+b" but does NOT match "aaa".
	t.Run("metachars treated literally - match", func(t *testing.T) {
		tt := &myT{}
		assert.Error(tt, fmt.Errorf("got a+b in input"), "a+b")
		assert.That(t, !tt.Failed(), "expected pass, got: %s", tt.message)
	})
	t.Run("metachars treated literally - no superset match", func(t *testing.T) {
		tt := &myT{}
		assert.Error(tt, fmt.Errorf("got aaa in input"), "a+")
		assert.That(t, tt.Failed(), "expected fail (literal 'a+' not in 'got aaa in input')")
	})
	t.Run("dot is literal", func(t *testing.T) {
		tt := &myT{}
		assert.Error(tt, fmt.Errorf("X-Y mismatch"), "X.Y")
		assert.That(t, tt.Failed(), "expected fail (literal 'X.Y' not in 'X-Y mismatch')")
	})
}

func TestEventually(t *testing.T) {
	t.Run("predicate true immediately", func(t *testing.T) {
		tt := &myT{}
		assert.Eventually(tt, func() bool { return true }, 100*time.Millisecond, 10*time.Millisecond)
		assert.That(t, !tt.Failed(), "expected pass, got: %s", tt.message)
	})
	t.Run("predicate becomes true after delay", func(t *testing.T) {
		tt := &myT{}
		started := time.Now()
		assert.Eventually(tt, func() bool {
			return time.Since(started) >= 30*time.Millisecond
		}, 200*time.Millisecond, 10*time.Millisecond)
		assert.That(t, !tt.Failed(), "expected pass, got: %s", tt.message)
	})
	t.Run("predicate never true fails with timeout", func(t *testing.T) {
		tt := &myT{}
		assert.Eventually(tt, func() bool { return false }, 30*time.Millisecond, 10*time.Millisecond)
		assert.That(t, tt.Failed(), "expected fail")
		assert.ContainsString(t, tt.message, "predicate did not become true")
		assert.ContainsString(t, tt.message, "30ms")
	})
	t.Run("zero timeout, predicate true", func(t *testing.T) {
		tt := &myT{}
		assert.Eventually(tt, func() bool { return true }, 0, 10*time.Millisecond)
		assert.That(t, !tt.Failed(), "expected pass on zero-timeout single call")
	})
	t.Run("zero timeout, predicate false", func(t *testing.T) {
		tt := &myT{}
		assert.Eventually(tt, func() bool { return false }, 0, 10*time.Millisecond)
		assert.That(t, tt.Failed(), "expected fail on zero-timeout single call")
	})
	t.Run("interval longer than timeout does not overshoot", func(t *testing.T) {
		tt := &myT{}
		start := time.Now()
		assert.Eventually(tt, func() bool { return false }, 20*time.Millisecond, time.Second)
		elapsed := time.Since(start)
		assert.That(t, tt.Failed(), "expected fail")
		assert.That(t, elapsed < 500*time.Millisecond, "expected return near the 20ms timeout, took %v", elapsed)
	})
	t.Run("predicate called at least once", func(t *testing.T) {
		tt := &myT{}
		calls := 0
		assert.Eventually(tt, func() bool {
			calls++
			return false
		}, 0, 10*time.Millisecond)
		assert.That(t, calls >= 1, "predicate must be called at least once even with zero timeout")
	})
	t.Run("custom message", func(t *testing.T) {
		tt := &myT{}
		assert.Eventually(tt, func() bool { return false }, 10*time.Millisecond, 5*time.Millisecond, "case %s", "zeta")
		assert.That(t, tt.Failed())
		assert.ContainsString(t, tt.message, "case zeta")
	})
}

func TestHasKey(t *testing.T) {
	t.Run("string key present", func(t *testing.T) {
		tt := &myT{}
		m := map[string]int{"a": 1, "b": 2}
		assert.HasKey(tt, m, "a")
		assert.That(t, !tt.Failed(), "expected pass, got: %s", tt.message)
	})
	t.Run("string key absent", func(t *testing.T) {
		tt := &myT{}
		m := map[string]int{"a": 1, "b": 2}
		assert.HasKey(tt, m, "c")
		assert.That(t, tt.Failed(), "expected fail")
		assert.ContainsString(t, tt.message, "expected key 'c'")
	})
	t.Run("int key", func(t *testing.T) {
		tt := &myT{}
		m := map[int]string{1: "x", 2: "y"}
		assert.HasKey(tt, m, 1)
		assert.That(t, !tt.Failed())
	})
	t.Run("empty map", func(t *testing.T) {
		tt := &myT{}
		m := map[string]int{}
		assert.HasKey(tt, m, "any")
		assert.That(t, tt.Failed())
		assert.ContainsString(t, tt.message, "expected key 'any'")
	})
	t.Run("custom message", func(t *testing.T) {
		tt := &myT{}
		assert.HasKey(tt, map[string]int{}, "k", "case %s", "epsilon")
		assert.That(t, tt.Failed())
		assert.ContainsString(t, tt.message, "case epsilon")
	})
	t.Run("zero-value V (any) does not falsely pass on missing", func(t *testing.T) {
		// Regression: comma-ok lookup, not value comparison. A map with a
		// zero V at key K should pass HasKey(K); a map missing K should fail
		// even if V's zero would equal something.
		tt := &myT{}
		m := map[string]*int{"present": nil} // value is nil but key is present
		assert.HasKey(tt, m, "present")
		assert.That(t, !tt.Failed(), "expected pass on present-but-nil-value")
	})
}

func TestNearlyEqual(t *testing.T) {
	t.Run("float within tolerance", func(t *testing.T) {
		tt := &myT{}
		assert.NearlyEqual(tt, 1.0, 1.0+1e-9, 1e-6)
		assert.That(t, !tt.Failed(), "expected pass, got: %s", tt.message)
	})
	t.Run("float exactly at tolerance", func(t *testing.T) {
		tt := &myT{}
		assert.NearlyEqual(tt, 1.0, 1.5, 0.5)
		assert.That(t, !tt.Failed(), "expected pass, got: %s", tt.message)
	})
	t.Run("float exceeds tolerance", func(t *testing.T) {
		tt := &myT{}
		assert.NearlyEqual(tt, 1.0, 2.0, 0.5)
		assert.That(t, tt.Failed(), "expected fail")
		assert.ContainsString(t, tt.message, "nearly equal")
		assert.ContainsString(t, tt.message, "tolerance 0.5")
	})
	t.Run("got less than want", func(t *testing.T) {
		tt := &myT{}
		assert.NearlyEqual(tt, 5, 7, 3)
		assert.That(t, !tt.Failed(), "expected pass (|5-7|=2 <= 3)")
	})
	t.Run("integer types", func(t *testing.T) {
		tt := &myT{}
		assert.NearlyEqual(tt, 100, 105, 10)
		assert.That(t, !tt.Failed(), "expected pass")
	})
	t.Run("unsigned types", func(t *testing.T) {
		tt := &myT{}
		assert.NearlyEqual(tt, uint(3), uint(7), uint(5))
		assert.That(t, !tt.Failed(), "expected pass (unsigned, want > got)")
	})
	t.Run("signed overflow fails instead of wrapping to a pass", func(t *testing.T) {
		tt := &myT{}
		assert.NearlyEqual(tt, int64(math.MaxInt64), int64(-1), int64(10))
		assert.That(t, tt.Failed(), "expected fail")
		assert.ContainsString(t, tt.message, "overflows int64")
		tt2 := &myT{}
		assert.NearlyEqual(tt2, int8(127), int8(-128), int8(5))
		assert.That(t, tt2.Failed(), "expected fail")
	})
	t.Run("zero tolerance is exact equality", func(t *testing.T) {
		tt := &myT{}
		assert.NearlyEqual(tt, 1.0, 1.0, 0.0)
		assert.That(t, !tt.Failed(), "expected pass")
		tt2 := &myT{}
		assert.NearlyEqual(tt2, 1.0, 1.0001, 0.0)
		assert.That(t, tt2.Failed(), "expected fail")
	})
	t.Run("custom message", func(t *testing.T) {
		tt := &myT{}
		assert.NearlyEqual(tt, 1.0, 2.0, 0.1, "case %s", "alpha")
		assert.That(t, tt.Failed(), "expected fail")
		assert.ContainsString(t, tt.message, "case alpha")
	})
}

func TestErrorStringMismatch(t *testing.T) {
	// String arg = literal substring match (strings.Contains). When err's
	// message doesn't contain the substring, the assertion fails.
	tt := &myT{}
	assert.Error(tt, fmt.Errorf("boom"), "lemons")
	assert.That(t, tt.Failed(), "expected fail")
	assert.ContainsString(t, tt.message, "expected error to contain 'lemons'")
	assert.ContainsString(t, tt.message, "boom")
}

func TestNil(t *testing.T) {
	t.Run("untyped nil", func(t *testing.T) {
		tt := &myT{}
		assert.Nil(tt, nil)
		assert.That(t, !tt.Failed(), "expected pass, got: %s", tt.message)
	})
	t.Run("nil pointer", func(t *testing.T) {
		tt := &myT{}
		var p *myThingImpl
		assert.Nil(tt, p)
		assert.That(t, !tt.Failed(), "expected pass, got: %s", tt.message)
	})
	t.Run("typed-nil-in-interface", func(t *testing.T) {
		// Classic trap: var p *T = nil; var i any = p; i != nil but underlying is nil.
		tt := &myT{}
		var p *myThingImpl
		var i any = p
		assert.That(t, i != nil, "precondition: typed-nil interface != nil")
		assert.Nil(tt, i)
		assert.That(t, !tt.Failed(), "Nil should see through typed-nil interface, got: %s", tt.message)
	})
	t.Run("nil slice", func(t *testing.T) {
		tt := &myT{}
		var s []int
		assert.Nil(tt, s)
		assert.That(t, !tt.Failed(), "expected pass, got: %s", tt.message)
	})
	t.Run("nil map", func(t *testing.T) {
		tt := &myT{}
		var m map[string]int
		assert.Nil(tt, m)
		assert.That(t, !tt.Failed(), "expected pass, got: %s", tt.message)
	})
	t.Run("nil chan", func(t *testing.T) {
		tt := &myT{}
		var c chan int
		assert.Nil(tt, c)
		assert.That(t, !tt.Failed(), "expected pass, got: %s", tt.message)
	})
	t.Run("nil func", func(t *testing.T) {
		tt := &myT{}
		var f func()
		assert.Nil(tt, f)
		assert.That(t, !tt.Failed(), "expected pass, got: %s", tt.message)
	})
	t.Run("non-nil pointer", func(t *testing.T) {
		tt := &myT{}
		x := &myThingImpl{}
		assert.Nil(tt, x)
		file, line := getAboveLineInfo(0)
		assert.That(t, tt.Failed(), "expected fail")
		assert.ContainsString(t, tt.message, "expected nil")
		assert.ContainsString(t, tt.message, "*assert_test.myThingImpl")
		assert.ContainsString(t, tt.message, "in "+file+":"+fmt.Sprint(line))
	})
	t.Run("non-nil int", func(t *testing.T) {
		// Non-nilable kind: never nil.
		tt := &myT{}
		assert.Nil(tt, 42)
		assert.That(t, tt.Failed(), "expected fail")
	})
	t.Run("custom message", func(t *testing.T) {
		tt := &myT{}
		assert.Nil(tt, 42, "want nil for %s", "thing")
		assert.That(t, tt.Failed(), "expected fail")
		assert.ContainsString(t, tt.message, "want nil for thing")
	})
}

func TestNotNil(t *testing.T) {
	t.Run("non-nil pointer", func(t *testing.T) {
		tt := &myT{}
		x := &myThingImpl{}
		assert.NotNil(tt, x)
		assert.That(t, !tt.Failed(), "expected pass, got: %s", tt.message)
	})
	t.Run("non-nil int", func(t *testing.T) {
		tt := &myT{}
		assert.NotNil(tt, 0)
		assert.That(t, !tt.Failed(), "0 is not nil, got: %s", tt.message)
	})
	t.Run("untyped nil", func(t *testing.T) {
		tt := &myT{}
		assert.NotNil(tt, nil)
		assert.That(t, tt.Failed(), "expected fail")
		assert.ContainsString(t, tt.message, "expected non-nil")
	})
	t.Run("typed-nil-in-interface fails", func(t *testing.T) {
		// Mirror of the Nil case — NotNil must also see through.
		tt := &myT{}
		var p *myThingImpl
		var i any = p
		assert.NotNil(tt, i)
		assert.That(t, tt.Failed(), "NotNil should see through typed-nil interface")
	})
	t.Run("custom message", func(t *testing.T) {
		tt := &myT{}
		assert.NotNil(tt, nil, "want non-nil %d", 1)
		assert.That(t, tt.Failed(), "expected fail")
		assert.ContainsString(t, tt.message, "want non-nil 1")
	})
}

func TestLen(t *testing.T) {
	t.Run("slice match", func(t *testing.T) {
		tt := &myT{}
		assert.Len(tt, []int{1, 2, 3}, 3)
		assert.That(t, !tt.Failed(), "expected pass, got: %s", tt.message)
	})
	t.Run("slice mismatch shows contents", func(t *testing.T) {
		tt := &myT{}
		assert.Len(tt, []string{"a", "b"}, 3)
		file, line := getAboveLineInfo(0)
		assert.That(t, tt.Failed(), "expected fail")
		assert.ContainsString(t, tt.message, "expected len 3, got len 2")
		assert.ContainsString(t, tt.message, "[a b]")
		assert.ContainsString(t, tt.message, "in "+file+":"+fmt.Sprint(line))
	})
	t.Run("map", func(t *testing.T) {
		tt := &myT{}
		assert.Len(tt, map[string]int{"a": 1, "b": 2}, 2)
		assert.That(t, !tt.Failed(), "expected pass, got: %s", tt.message)
	})
	t.Run("string", func(t *testing.T) {
		tt := &myT{}
		assert.Len(tt, "hello", 5)
		assert.That(t, !tt.Failed(), "expected pass, got: %s", tt.message)
	})
	t.Run("array", func(t *testing.T) {
		tt := &myT{}
		assert.Len(tt, [3]int{1, 2, 3}, 3)
		assert.That(t, !tt.Failed(), "expected pass, got: %s", tt.message)
	})
	t.Run("chan", func(t *testing.T) {
		tt := &myT{}
		c := make(chan int, 4)
		c <- 1
		c <- 2
		assert.Len(tt, c, 2)
		assert.That(t, !tt.Failed(), "expected pass, got: %s", tt.message)
	})
	t.Run("nil slice has len 0", func(t *testing.T) {
		tt := &myT{}
		var s []int
		assert.Len(tt, s, 0)
		assert.That(t, !tt.Failed(), "expected pass, got: %s", tt.message)
	})
	t.Run("unsupported kind panics", func(t *testing.T) {
		assert.Panic(t, func() { assert.Len(&myT{}, 42, 1) }, func(t testing.TB, rec any) {
			s, ok := rec.(string)
			assert.That(t, ok, "expected string panic, got %T", rec)
			assert.ContainsString(t, s, "Len: unsupported kind")
		})
	})
	t.Run("custom message", func(t *testing.T) {
		tt := &myT{}
		assert.Len(tt, []int{1}, 2, "want %d items", 2)
		assert.That(t, tt.Failed(), "expected fail")
		assert.ContainsString(t, tt.message, "want 2 items")
	})
}

func TestTypeCustomMessage(t *testing.T) {
	tt := &myT{}
	var x int = 1
	_ = assert.Type[myThing](tt, x, "want myThing got %d", 1)
	assert.That(t, tt.Failed(), "expected fail")
	assert.ContainsString(t, tt.message, "want myThing got 1")
}
