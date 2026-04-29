package assert_test

import (
	"errors"
	"fmt"
	"regexp"
	"testing"

	"github.com/lczyk/assert"
)

// Benchmarks measure the passing-path overhead of each assertion versus the
// equivalent vanilla `if !cond { t.Errorf }` form. The point is to confirm
// the helper layer doesn't measurably slow real test runs (where assertions
// almost always pass).

// benchNilErr / benchNilAny defeat the nilness analyzer for vanilla baselines
// without changing what the comparison measures.
var benchNilErr error
var benchNilAny any

func BenchmarkVanillaEqual(b *testing.B) {
	x, y := 42, 42
	for i := 0; i < b.N; i++ {
		if x != y {
			b.Errorf("expected %d == %d", x, y)
		}
	}
}

func BenchmarkEqual(b *testing.B) {
	x, y := 42, 42
	for i := 0; i < b.N; i++ {
		assert.Equal(b, x, y)
	}
}

func BenchmarkVanillaThat(b *testing.B) {
	for i := 0; i < b.N; i++ {
		if !(1 < 2) {
			b.Errorf("expected 1 < 2")
		}
	}
}

func BenchmarkThat(b *testing.B) {
	for i := 0; i < b.N; i++ {
		assert.That(b, 1 < 2)
	}
}

func BenchmarkThatWithMsg(b *testing.B) {
	for i := 0; i < b.N; i++ {
		assert.That(b, 1 < 2, "expected %d < %d", 1, 2)
	}
}

func BenchmarkVanillaNoError(b *testing.B) {
	err := benchNilErr
	for i := 0; i < b.N; i++ {
		if err != nil {
			b.Errorf("expected no error, got %v", err)
		}
	}
}

func BenchmarkNoError(b *testing.B) {
	var err error
	for i := 0; i < b.N; i++ {
		assert.NoError(b, err)
	}
}

func BenchmarkErrorSentinel(b *testing.B) {
	err := errors.New("boom")
	for i := 0; i < b.N; i++ {
		assert.Error(b, err, assert.AnyError)
	}
}

func BenchmarkErrorIs(b *testing.B) {
	sentinel := errors.New("sentinel")
	wrapped := fmt.Errorf("ctx: %w", sentinel)
	for i := 0; i < b.N; i++ {
		assert.ErrorIs(b, wrapped, sentinel)
	}
}

func BenchmarkErrorRegexString(b *testing.B) {
	err := errors.New("connection refused: 127.0.0.1:5432")
	for i := 0; i < b.N; i++ {
		assert.Error(b, err, "connection refused")
	}
}

func BenchmarkErrorRegexCompiled(b *testing.B) {
	err := errors.New("connection refused: 127.0.0.1:5432")
	re := regexp.MustCompile("connection refused")
	for i := 0; i < b.N; i++ {
		assert.Error(b, err, re)
	}
}

func BenchmarkVanillaNil(b *testing.B) {
	x := benchNilAny
	for i := 0; i < b.N; i++ {
		if x != nil {
			b.Errorf("expected nil")
		}
	}
}

func BenchmarkNil(b *testing.B) {
	var x any
	for i := 0; i < b.N; i++ {
		assert.Nil(b, x)
	}
}

func BenchmarkNotNil(b *testing.B) {
	x := any(&struct{}{})
	for i := 0; i < b.N; i++ {
		assert.NotNil(b, x)
	}
}

func BenchmarkVanillaLen(b *testing.B) {
	xs := []int{1, 2, 3, 4, 5}
	for i := 0; i < b.N; i++ {
		if len(xs) != 5 {
			b.Errorf("expected len 5")
		}
	}
}

func BenchmarkLen(b *testing.B) {
	xs := []int{1, 2, 3, 4, 5}
	for i := 0; i < b.N; i++ {
		assert.Len(b, xs, 5)
	}
}

func BenchmarkEqualArraysSmall(b *testing.B) {
	x := []int{1, 2, 3, 4, 5}
	y := []int{1, 2, 3, 4, 5}
	for i := 0; i < b.N; i++ {
		assert.EqualArrays(b, x, y)
	}
}

func BenchmarkEqualArraysLarge(b *testing.B) {
	x := make([]int, 1024)
	y := make([]int, 1024)
	for i := range x {
		x[i] = i
		y[i] = i
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		assert.EqualArrays(b, x, y)
	}
}

func BenchmarkEqualArraysUnordered(b *testing.B) {
	x := []int{5, 4, 3, 2, 1}
	y := []int{1, 2, 3, 4, 5}
	for i := 0; i < b.N; i++ {
		assert.EqualArraysUnordered(b, x, y)
	}
}

func BenchmarkEqualMaps(b *testing.B) {
	x := map[string]int{"a": 1, "b": 2, "c": 3}
	y := map[string]int{"a": 1, "b": 2, "c": 3}
	for i := 0; i < b.N; i++ {
		assert.EqualMaps(b, x, y)
	}
}

func BenchmarkType(b *testing.B) {
	var v any = "hello"
	for i := 0; i < b.N; i++ {
		_ = assert.Type[string](b, v)
	}
}

func BenchmarkContainsString(b *testing.B) {
	h := "the quick brown fox jumps over the lazy dog"
	for i := 0; i < b.N; i++ {
		assert.ContainsString(b, h, "fox")
	}
}

func BenchmarkEqualLineByLine(b *testing.B) {
	s := "line1\nline2\nline3\nline4\nline5"
	for i := 0; i < b.N; i++ {
		assert.EqualLineByLine(b, s, s)
	}
}

func BenchmarkPanic(b *testing.B) {
	f := func() { panic("expected") }
	for i := 0; i < b.N; i++ {
		assert.Panic(b, f, nil)
	}
}
