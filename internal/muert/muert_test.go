// Smoke tests for muert. muert is designed to be copy/pasted into other
// packages' internal/ directories as a single-file assertion helper, so
// downstream copies do NOT need this test file — it only exists here to
// guard against drift within this repo. Imports muert as `assert` so the
// call-site style matches github.com/lczyk/assert.
package muert_test

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	assert "github.com/lczyk/assert/internal/muert"
)

func TestLicenseEmbedded(t *testing.T) {
	// muert is intended to be copy/pasted. The License constant must carry
	// the project's LICENCE text verbatim so downstream copies stay compliant.
	path := filepath.Join("..", "..", "LICENCE")
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("cannot read %s: %v", path, err)
	}
	want := string(b)
	if !strings.Contains(assert.License, strings.TrimRight(want, "\n")) {
		t.Errorf("muert.License does not contain LICENCE file verbatim")
	}
}

type myT struct {
	testing.T
	message string
}

func (t *myT) Errorf(format string, args ...any) {
	t.message = fmt.Sprintf(format, args...)
	t.Fail()
}

func TestThat(t *testing.T) {
	t.Run("true", func(t *testing.T) {
		tt := &myT{}
		assert.That(tt, true)
		if tt.Failed() {
			t.Errorf("expected pass, got fail: %s", tt.message)
		}
	})
	t.Run("false", func(t *testing.T) {
		tt := &myT{}
		assert.That(tt, false)
		if !tt.Failed() {
			t.Errorf("expected fail, got pass")
		}
	})
	t.Run("false with message", func(t *testing.T) {
		tt := &myT{}
		assert.That(tt, false, "boom %d", 42)
		if !tt.Failed() {
			t.Errorf("expected fail")
		}
		if want := "boom 42"; !contains(tt.message, want) {
			t.Errorf("expected message to contain %q, got %q", want, tt.message)
		}
	})
}

func TestEqual(t *testing.T) {
	t.Run("equal", func(t *testing.T) {
		tt := &myT{}
		assert.Equal(tt, 1, 1)
		if tt.Failed() {
			t.Errorf("expected pass: %s", tt.message)
		}
	})
	t.Run("not equal", func(t *testing.T) {
		tt := &myT{}
		assert.Equal(tt, 1, 2)
		if !tt.Failed() {
			t.Errorf("expected fail")
		}
	})
}

func TestNotEqual(t *testing.T) {
	t.Run("not equal", func(t *testing.T) {
		tt := &myT{}
		assert.NotEqual(tt, 1, 2)
		if tt.Failed() {
			t.Errorf("expected pass: %s", tt.message)
		}
	})
	t.Run("equal", func(t *testing.T) {
		tt := &myT{}
		assert.NotEqual(tt, 1, 1)
		if !tt.Failed() {
			t.Errorf("expected fail")
		}
	})
}

func TestNoError(t *testing.T) {
	t.Run("nil", func(t *testing.T) {
		tt := &myT{}
		assert.NoError(tt, nil)
		if tt.Failed() {
			t.Errorf("expected pass: %s", tt.message)
		}
	})
	t.Run("non-nil", func(t *testing.T) {
		tt := &myT{}
		assert.NoError(tt, fmt.Errorf("boom"))
		if !tt.Failed() {
			t.Errorf("expected fail")
		}
	})
}

func TestError(t *testing.T) {
	t.Run("contains", func(t *testing.T) {
		tt := &myT{}
		assert.Error(tt, fmt.Errorf("boom lemons"), "lemons")
		if tt.Failed() {
			t.Errorf("expected pass: %s", tt.message)
		}
	})
	t.Run("does not contain", func(t *testing.T) {
		tt := &myT{}
		assert.Error(tt, fmt.Errorf("boom lemons"), "oranges")
		if !tt.Failed() {
			t.Errorf("expected fail")
		}
	})
	t.Run("nil err", func(t *testing.T) {
		tt := &myT{}
		assert.Error(tt, nil, "anything")
		if !tt.Failed() {
			t.Errorf("expected fail on nil error")
		}
		if want := "got nil"; !contains(tt.message, want) {
			t.Errorf("expected message to contain %q, got %q", want, tt.message)
		}
	})
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
