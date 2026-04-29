//go:build demo

package demo_test

// Tests in this file intentionally fail to showcase assert's failure
// output (message, file:line, source-line snippet). Run via demo_runner.sh
// or `make demo` — running `go test ./demo/...` directly will report
// failures, which is the point.

import (
	"errors"
	"io"
	"testing"

	"github.com/lczyk/assert"
)

func TestDemoThat(t *testing.T) {
	x, y := 1, 2
	assert.That(t, x == y, "x and y should match")
}

func TestDemoEqual(t *testing.T) {
	assert.Equal(t, 1, 2)
}

func TestDemoNotEqual(t *testing.T) {
	assert.NotEqual(t, 7, 7)
}

func TestDemoNoError(t *testing.T) {
	assert.NoError(t, errors.New("disk on fire"))
}

func TestDemoErrorRegex(t *testing.T) {
	assert.Error(t, errors.New("file not found"), "permission denied")
}

func TestDemoErrorIs(t *testing.T) {
	assert.ErrorIs(t, errors.New("plain error"), io.EOF)
}

func TestDemoEqualArrays(t *testing.T) {
	assert.EqualArrays(t, []int{1, 2, 3}, []int{1, 9, 3})
}

func TestDemoEqualMaps(t *testing.T) {
	assert.EqualMaps(t,
		map[string]int{"a": 1, "b": 2},
		map[string]int{"a": 1, "b": 99},
	)
}

func TestDemoNil(t *testing.T) {
	x := 42
	assert.Nil(t, &x)
}

func TestDemoNotNil(t *testing.T) {
	var p *int
	assert.NotNil(t, p)
}

func TestDemoLen(t *testing.T) {
	assert.Len(t, []int{1, 2, 3}, 5)
}

func TestDemoType(t *testing.T) {
	var v any = "hello"
	_ = assert.Type[int](t, v)
}

func TestDemoEqualLineByLine(t *testing.T) {
	assert.EqualLineByLine(t, "alpha\nbeta\ngamma", "alpha\nBETA\ngamma")
}

func TestDemoContainsString(t *testing.T) {
	assert.ContainsString(t, "the quick brown fox", "lazy dog")
}

func TestDemoPanic(t *testing.T) {
	assert.Panic(t, func() { /* no panic */ }, nil)
}
