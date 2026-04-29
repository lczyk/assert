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

// TestVanillaNilTypedNil shows the typed-nil-in-interface gotcha:
// `var p *int = nil; var i any = p` — `i != nil` is TRUE even though
// the underlying pointer is nil. Vanilla check silently passes here
// (false negative); the paired TestDemoNilTypedNil uses assert.Nil
// which correctly reports it via reflect.
func TestVanillaNilTypedNil(t *testing.T) {
	var p *int
	var i any = p
	if i != nil {
		t.Errorf("expected nil, got %v", i)
	}
}
