//go:build demo

package demo_test

// Each TestVanilla* is paired with a TestDemo* of the same suffix to
// show the same failure rendered with stdlib testing vs assert.

import (
	"errors"
	"reflect"
	"testing"
)

func TestVanillaEqual(t *testing.T) {
	a, b := 1, 2
	if a != b {
		t.Errorf("expected %d == %d", a, b)
	}
}

func TestVanillaNoError(t *testing.T) {
	err := errors.New("disk on fire")
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestVanillaEqualArrays(t *testing.T) {
	got, want := []int{1, 2, 3}, []int{1, 9, 3}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

func TestVanillaEqualMaps(t *testing.T) {
	got := map[string]int{"a": 1, "b": 2}
	want := map[string]int{"a": 1, "b": 99}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
}

// TestVanillaNotNilTypedNil shows the typed-nil-in-interface gotcha:
// `var p *int = nil; var i any = p` makes `i == nil` false even though
// the pointer inside is nil, so this check passes silently (a false
// negative). The paired TestDemoNotNilTypedNil uses assert.NotNil,
// which looks inside the interface and fails.
func TestVanillaNotNilTypedNil(t *testing.T) {
	var p *int
	var i any = p
	if i == nil {
		t.Errorf("expected non-nil, got %v", i)
	}
}
