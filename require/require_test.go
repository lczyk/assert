package require_test

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/lczyk/assert"
	"github.com/lczyk/assert/require"
)

// myT records which failure path was taken so each test can assert
// that require.* called Fatalf (not Errorf). Both override paths
// suppress runtime.Goexit so the test goroutine keeps running and
// can inspect the recorded state.
type myT struct {
	testing.T
	errorfCalled bool
	fatalfCalled bool
	message      string
}

func (t *myT) Errorf(format string, args ...any) {
	t.errorfCalled = true
	t.message = fmt.Sprintf(format, args...)
	t.Fail()
}

func (t *myT) Fatalf(format string, args ...any) {
	t.fatalfCalled = true
	t.message = fmt.Sprintf(format, args...)
	t.Fail()
}

var _ testing.TB = &myT{}

// requireUsedFatalf is the standard end-of-fail-path check: confirm
// the wrapper routed to t.Fatalf, not t.Errorf.
func requireUsedFatalf(t *testing.T, tt *myT) {
	t.Helper()
	assert.That(t, tt.fatalfCalled, "expected Fatalf to be called, but was not (Errorf=%v)", tt.errorfCalled)
	assert.That(t, !tt.errorfCalled, "expected Errorf NOT to be called, but it was")
}

func TestThat(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		tt := &myT{}
		require.That(tt, true)
		assert.That(t, !tt.Failed())
	})
	t.Run("fails fatally", func(t *testing.T) {
		tt := &myT{}
		require.That(tt, false, "boom")
		requireUsedFatalf(t, tt)
		assert.ContainsString(t, tt.message, "boom")
	})
}

func TestEqual(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		tt := &myT{}
		require.Equal(tt, 1, 1)
		assert.That(t, !tt.Failed())
	})
	t.Run("fails fatally", func(t *testing.T) {
		tt := &myT{}
		require.Equal(tt, 1, 2)
		requireUsedFatalf(t, tt)
	})
}

func TestNotEqual(t *testing.T) {
	tt := &myT{}
	require.NotEqual(tt, 1, 1)
	requireUsedFatalf(t, tt)
}

func TestNearlyEqual(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		tt := &myT{}
		require.NearlyEqual(tt, 1.0, 1.05, 0.1)
		assert.That(t, !tt.Failed())
	})
	t.Run("fails fatally", func(t *testing.T) {
		tt := &myT{}
		require.NearlyEqual(tt, 1.0, 2.0, 0.1)
		requireUsedFatalf(t, tt)
	})
}

func TestNoError(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		tt := &myT{}
		require.NoError(tt, nil)
		assert.That(t, !tt.Failed())
	})
	t.Run("fails fatally", func(t *testing.T) {
		tt := &myT{}
		require.NoError(tt, errors.New("boom"))
		requireUsedFatalf(t, tt)
	})
}

func TestError(t *testing.T) {
	t.Run("passes with AnyError", func(t *testing.T) {
		tt := &myT{}
		require.Error(tt, errors.New("x"), require.AnyError)
		assert.That(t, !tt.Failed())
	})
	t.Run("AnyError identity matches assert.AnyError", func(t *testing.T) {
		assert.That(t, require.AnyError == assert.AnyError, "sentinel must be shared")
	})
	t.Run("fails fatally on nil with AnyError", func(t *testing.T) {
		tt := &myT{}
		require.Error(tt, nil, require.AnyError)
		requireUsedFatalf(t, tt)
	})
	t.Run("passes with string substring", func(t *testing.T) {
		tt := &myT{}
		require.Error(tt, errors.New("connection refused"), "refused")
		assert.That(t, !tt.Failed())
	})
	t.Run("fails fatally on substring miss", func(t *testing.T) {
		tt := &myT{}
		require.Error(tt, errors.New("oops"), "nope")
		requireUsedFatalf(t, tt)
	})
}

func TestErrorIs(t *testing.T) {
	sentinel := errors.New("sentinel")
	t.Run("passes on wrap match", func(t *testing.T) {
		tt := &myT{}
		require.ErrorIs(tt, fmt.Errorf("wrap: %w", sentinel), sentinel)
		assert.That(t, !tt.Failed())
	})
	t.Run("fails fatally on mismatch", func(t *testing.T) {
		tt := &myT{}
		require.ErrorIs(tt, errors.New("other"), sentinel)
		requireUsedFatalf(t, tt)
	})
}

func TestEqualCmp(t *testing.T) {
	cmp := func(a, b []int) bool {
		if len(a) != len(b) {
			return false
		}
		for i := range a {
			if a[i] != b[i] {
				return false
			}
		}
		return true
	}
	t.Run("passes", func(t *testing.T) {
		tt := &myT{}
		require.EqualCmp(tt, []int{1, 2}, []int{1, 2}, cmp)
		assert.That(t, !tt.Failed())
	})
	t.Run("fails fatally", func(t *testing.T) {
		tt := &myT{}
		require.EqualCmp(tt, []int{1}, []int{2}, cmp)
		requireUsedFatalf(t, tt)
	})
}

func TestEqualCmpAny(t *testing.T) {
	tt := &myT{}
	require.EqualCmpAny(tt, 1, 2, func(a, b any) bool { return a == b })
	requireUsedFatalf(t, tt)
}

func TestEqualArrays(t *testing.T) {
	tt := &myT{}
	require.EqualArrays(tt, []int{1}, []int{2})
	requireUsedFatalf(t, tt)
}

func TestEqualMaps(t *testing.T) {
	tt := &myT{}
	require.EqualMaps(tt, map[string]int{"a": 1}, map[string]int{"a": 2})
	requireUsedFatalf(t, tt)
}

func TestEqualArraysUnordered(t *testing.T) {
	tt := &myT{}
	require.EqualArraysUnordered(tt, []int{1, 2}, []int{1, 3})
	requireUsedFatalf(t, tt)
}

func TestNil(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		tt := &myT{}
		require.Nil(tt, nil)
		assert.That(t, !tt.Failed())
	})
	t.Run("fails fatally", func(t *testing.T) {
		tt := &myT{}
		require.Nil(tt, 42)
		requireUsedFatalf(t, tt)
	})
}

func TestNotNil(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		tt := &myT{}
		require.NotNil(tt, 42)
		assert.That(t, !tt.Failed())
	})
	t.Run("fails fatally", func(t *testing.T) {
		tt := &myT{}
		require.NotNil(tt, nil)
		requireUsedFatalf(t, tt)
	})
}

func TestLen(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		tt := &myT{}
		require.Len(tt, []int{1, 2, 3}, 3)
		assert.That(t, !tt.Failed())
	})
	t.Run("fails fatally", func(t *testing.T) {
		tt := &myT{}
		require.Len(tt, []int{1}, 5)
		requireUsedFatalf(t, tt)
	})
}

type someType interface{ Quack() }

type duck struct{}

func (duck) Quack() {}

func TestType(t *testing.T) {
	t.Run("passes and returns the typed value", func(t *testing.T) {
		tt := &myT{}
		var v any = duck{}
		got := require.Type[someType](tt, v)
		assert.That(t, !tt.Failed())
		assert.That(t, got != nil)
	})
	t.Run("fails fatally on mismatch", func(t *testing.T) {
		tt := &myT{}
		var v any = 1
		_ = require.Type[someType](tt, v)
		requireUsedFatalf(t, tt)
	})
}

func TestEqualLineByLine(t *testing.T) {
	tt := &myT{}
	require.EqualLineByLine(tt, "a\nb", "a\nc")
	requireUsedFatalf(t, tt)
}

func TestHasKey(t *testing.T) {
	t.Run("passes", func(t *testing.T) {
		tt := &myT{}
		require.HasKey(tt, map[string]int{"a": 1}, "a")
		assert.That(t, !tt.Failed())
	})
	t.Run("fails fatally", func(t *testing.T) {
		tt := &myT{}
		require.HasKey(tt, map[string]int{"a": 1}, "b")
		requireUsedFatalf(t, tt)
	})
}

func TestContainsString(t *testing.T) {
	tt := &myT{}
	require.ContainsString(tt, "hello world", "nope")
	requireUsedFatalf(t, tt)
}

func TestPanic(t *testing.T) {
	t.Run("passes when f panics", func(t *testing.T) {
		tt := &myT{}
		require.Panic(tt, func() { panic("kapow") }, nil)
		assert.That(t, !tt.Failed())
	})
	t.Run("fails fatally when f does not panic", func(t *testing.T) {
		tt := &myT{}
		require.Panic(tt, func() {}, nil)
		requireUsedFatalf(t, tt)
	})
}

func TestEventually(t *testing.T) {
	t.Run("passes when predicate becomes true", func(t *testing.T) {
		tt := &myT{}
		require.Eventually(tt, func() bool { return true }, 50*time.Millisecond, 10*time.Millisecond)
		assert.That(t, !tt.Failed())
	})
	t.Run("fails fatally on timeout", func(t *testing.T) {
		tt := &myT{}
		require.Eventually(tt, func() bool { return false }, 20*time.Millisecond, 5*time.Millisecond)
		requireUsedFatalf(t, tt)
	})
}
