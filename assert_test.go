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
	t.Run("matching errors", func(t *testing.T) {
		tt := &myT{}
		a := fmt.Errorf("boom")
		b := fmt.Errorf("boom")
		assert.Error(tt, a, b)
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
	assert.ContainsString(t, tt.message, "expected no error, got 'boom'")
}

func TestErrorStringEmptyExpectedNilErr(t *testing.T) {
	// Empty expected string + nil err is a no-op per assert_error.
	tt := &myT{}
	assert.Error(tt, nil, "")
	assert.That(t, !tt.Failed(), "expected pass, got: %s", tt.message)
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

func TestTypeCustomMessage(t *testing.T) {
	tt := &myT{}
	var x int = 1
	_ = assert.Type[myThing](tt, x, "want myThing got %d", 1)
	assert.That(t, tt.Failed(), "expected fail")
	assert.ContainsString(t, tt.message, "want myThing got 1")
}
