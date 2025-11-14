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
