package assert_test

import (
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"testing"

	"github.com/lczyk/assert"
)

func ExampleThat() {

	t := &testing.T{}
	assert.That(t, true, "This should always pass")
	fmt.Println(t.Failed())

	assert.That(t, false, "This should always fail")
	fmt.Println(t.Failed())

	// output:
	//
	// false
	// true
}

func ExamplePanic() {

	t := &testing.T{}

	// Assert that f panics, and inspect the recovered value.
	assert.Panic(t,
		func() { panic("boom") },
		func(t testing.TB, rec any) {
			assert.Equal(t, rec, "boom")
		},
	)
	fmt.Println(t.Failed())

	// Pass nil as the recovery func if you only care that *something* panicked.
	assert.Panic(t, func() { panic("ignored") }, nil)
	fmt.Println(t.Failed())

	// Fails when f does not panic.
	assert.Panic(t, func() {}, nil)
	fmt.Println(t.Failed())

	// output:
	// false
	// false
	// true
}

func ExampleEqual() {
	t := &testing.T{}
	assert.Equal(t, 1, 1)
	fmt.Println(t.Failed())

	assert.Equal(t, 1, 2)
	fmt.Println(t.Failed())

	// output:
	// false
	// true
}

func ExampleNotEqual() {
	t := &testing.T{}
	assert.NotEqual(t, 1, 2)
	fmt.Println(t.Failed())

	assert.NotEqual(t, 1, 1)
	fmt.Println(t.Failed())

	// output:
	// false
	// true
}

func ExampleNoError() {
	t := &testing.T{}
	assert.NoError(t, nil)
	fmt.Println(t.Failed())

	assert.NoError(t, errors.New("boom"))
	fmt.Println(t.Failed())

	// output:
	// false
	// true
}

func ExampleError() {
	t := &testing.T{}

	// AnyError sentinel: any non-nil err passes.
	assert.Error(t, errors.New("boom"), assert.AnyError)
	fmt.Println(t.Failed())

	// String expected: regex substring match against err.Error().
	assert.Error(t, errors.New("file not found"), "not found")
	fmt.Println(t.Failed())

	// *regexp.Regexp expected.
	assert.Error(t, errors.New("code 42"), regexp.MustCompile(`code \d+`))
	fmt.Println(t.Failed())

	// error expected: errors.Is wrap-chain match.
	assert.Error(t, fmt.Errorf("wrap: %w", io.EOF), io.EOF)
	fmt.Println(t.Failed())

	// Mismatch fails.
	assert.Error(t, errors.New("boom"), "other")
	fmt.Println(t.Failed())

	// output:
	// false
	// false
	// false
	// false
	// true
}

func ExampleErrorIs() {
	t := &testing.T{}
	assert.ErrorIs(t, fmt.Errorf("wrap: %w", io.EOF), io.EOF)
	fmt.Println(t.Failed())

	assert.ErrorIs(t, errors.New("boom"), io.EOF)
	fmt.Println(t.Failed())

	// output:
	// false
	// true
}

func ExampleEqualCmp() {
	t := &testing.T{}
	caseInsensitive := func(a, b string) bool {
		if len(a) != len(b) {
			return false
		}
		for i := range a {
			ca, cb := a[i], b[i]
			if ca >= 'A' && ca <= 'Z' {
				ca += 32
			}
			if cb >= 'A' && cb <= 'Z' {
				cb += 32
			}
			if ca != cb {
				return false
			}
		}
		return true
	}
	assert.EqualCmp(t, "Foo", "foo", caseInsensitive)
	fmt.Println(t.Failed())

	assert.EqualCmp(t, "Foo", "bar", caseInsensitive)
	fmt.Println(t.Failed())

	// output:
	// false
	// true
}

func ExampleEqualArrays() {
	t := &testing.T{}
	assert.EqualArrays(t, []int{1, 2, 3}, []int{1, 2, 3})
	fmt.Println(t.Failed())

	assert.EqualArrays(t, []int{1, 2, 3}, []int{3, 2, 1})
	fmt.Println(t.Failed())

	// output:
	// false
	// true
}

func ExampleEqualArraysUnordered() {
	t := &testing.T{}
	assert.EqualArraysUnordered(t, []int{1, 2, 3}, []int{3, 2, 1})
	fmt.Println(t.Failed())

	assert.EqualArraysUnordered(t, []int{1, 2}, []int{1, 2, 3})
	fmt.Println(t.Failed())

	// output:
	// false
	// true
}

func ExampleEqualMaps() {
	t := &testing.T{}
	assert.EqualMaps(t, map[string]int{"a": 1}, map[string]int{"a": 1})
	fmt.Println(t.Failed())

	assert.EqualMaps(t, map[string]int{"a": 1}, map[string]int{"a": 2})
	fmt.Println(t.Failed())

	// output:
	// false
	// true
}

func ExampleNil() {
	t := &testing.T{}
	assert.Nil(t, nil)
	fmt.Println(t.Failed())

	var p *int
	// Typed-nil-in-interface still detected as nil.
	assert.Nil(t, p)
	fmt.Println(t.Failed())

	x := 1
	assert.Nil(t, &x)
	fmt.Println(t.Failed())

	// output:
	// false
	// false
	// true
}

func ExampleNotNil() {
	t := &testing.T{}
	x := 1
	assert.NotNil(t, &x)
	fmt.Println(t.Failed())

	assert.NotNil(t, nil)
	fmt.Println(t.Failed())

	// output:
	// false
	// true
}

func ExampleLen() {
	t := &testing.T{}
	assert.Len(t, []int{1, 2, 3}, 3)
	fmt.Println(t.Failed())

	assert.Len(t, "hello", 5)
	fmt.Println(t.Failed())

	assert.Len(t, []int{1, 2}, 5)
	fmt.Println(t.Failed())

	// output:
	// false
	// false
	// true
}

func ExampleType() {
	t := &testing.T{}
	var v any = "hello"
	s := assert.Type[string](t, v)
	fmt.Println(s, t.Failed())

	_ = assert.Type[int](t, v)
	fmt.Println(t.Failed())

	// output:
	// hello false
	// true
}

func ExampleEqualLineByLine() {
	t := &testing.T{}
	assert.EqualLineByLine(t, "a\nb\nc", "a\nb\nc")
	fmt.Println(t.Failed())

	// Trailing newline ignored.
	assert.EqualLineByLine(t, "a\nb\n", "a\nb")
	fmt.Println(t.Failed())

	assert.EqualLineByLine(t, "a\nb", "a\nx")
	fmt.Println(t.Failed())

	// output:
	// false
	// false
	// true
}

func ExampleContainsString() {
	t := &testing.T{}
	assert.ContainsString(t, "hello world", "world")
	fmt.Println(t.Failed())

	assert.ContainsString(t, "hello world", "xyz")
	fmt.Println(t.Failed())

	// output:
	// false
	// true
}

// captureT prints only the source-snippet portion of failure messages
// so Example output is stable (no file paths or line numbers).
type captureT struct{ *testing.T }

func (c *captureT) Errorf(format string, args ...any) {
	s := fmt.Sprintf(format, args...)
	if i := strings.Index(s, "\n  > "); i >= 0 {
		fmt.Println(strings.TrimSpace(s[i+1:]))
	}
}

func ExampleEqual_sourceSnippet() {
	t := &captureT{T: &testing.T{}}
	assert.Equal(t, 1, 2)

	// Output:
	// > assert.Equal(t, 1, 2)
}

func TestExample(t *testing.T) {
	a := 1
	b := 2
	if a == b {
		t.Errorf("Expected %d to not equal %d", a, b)
	}
}
