package assert_test

import (
	"fmt"
	"regexp"
	"runtime"
	"testing"

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
	t.Run("fails", func(t *testing.T) {
		tt := &testing.T{}
		assert.That(t, !tt.Failed())
		var x int = 1
		y := assert.Type[myThing](tt, x)
		_ = y
		assert.That(t, tt.Failed(), "Expected test to fail, but it did not")
	})
	t.Run("succeeds", func(t *testing.T) {
		tt := &testing.T{}
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
		assert.ContainsString(t, tt.message, "expected line 2 to be 'b', got 'X'")
	})

	t.Run("differing first line", func(t *testing.T) {
		tt := &myT{}
		assert.EqualLineByLine(tt, "x\ny", "a\ny")
		assert.That(t, tt.Failed(), "expected fail")
		assert.ContainsString(t, tt.message, "expected line 1 to be 'x', got 'a'")
	})

	t.Run("differing last line", func(t *testing.T) {
		tt := &myT{}
		assert.EqualLineByLine(tt, "a\nb\nc", "a\nb\nZ")
		assert.That(t, tt.Failed(), "expected fail")
		assert.ContainsString(t, tt.message, "expected line 3 to be 'c', got 'Z'")
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
}

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
	assert.ContainsString(t, tt.message, "expected error to match 'boom'")
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
		assert.Equal(t, rec, "expected type is not an error or string")
	})
}

func TestThatNonStringFirstArg(t *testing.T) {
	// First arg not a string: argsToMessage falls through to default %v path.
	tt := &myT{}
	assert.That(tt, false, 42)
	assert.That(t, tt.Failed(), "expected fail")
	assert.ContainsString(t, tt.message, "42")
}

func TestThatFailNoArgs(t *testing.T) {
	// Covers the "assertion failed" default-message lambda in assert().
	tt := &myT{}
	assert.That(tt, false)
	assert.That(t, tt.Failed(), "expected fail")
	assert.ContainsString(t, tt.message, "assertion failed")
}

func TestErrorStringRegexMismatch(t *testing.T) {
	// Covers the "expected error to match" path when expected is a string
	// (compiled as regex) and err's message doesn't match.
	tt := &myT{}
	assert.Error(tt, fmt.Errorf("boom"), "lemons")
	assert.That(t, tt.Failed(), "expected fail")
	assert.ContainsString(t, tt.message, "expected error to match 'lemons'")
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
