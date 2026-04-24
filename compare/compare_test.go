package compare_test

import (
	"fmt"
	"testing"

	"github.com/lczyk/assert/compare"
)

type customErr struct{ msg string }

func (e customErr) Error() string { return e.msg }

func TestCompareErrors(t *testing.T) {
	t.Run("both nil", func(t *testing.T) {
		if !compare.Errors(nil, nil) {
			t.Errorf("expected both-nil errors to be equal")
		}
	})
	t.Run("one nil", func(t *testing.T) {
		if compare.Errors(nil, fmt.Errorf("x")) {
			t.Errorf("expected nil vs non-nil to be unequal")
		}
		if compare.Errors(fmt.Errorf("x"), nil) {
			t.Errorf("expected non-nil vs nil to be unequal")
		}
	})
	t.Run("same message same type", func(t *testing.T) {
		a := fmt.Errorf("boom")
		b := fmt.Errorf("boom")
		if !compare.Errors(a, b) {
			t.Errorf("expected equal errors to compare equal")
		}
	})
	t.Run("different message", func(t *testing.T) {
		if compare.Errors(fmt.Errorf("a"), fmt.Errorf("b")) {
			t.Errorf("expected different messages to be unequal")
		}
	})
	t.Run("same message different type", func(t *testing.T) {
		a := fmt.Errorf("boom")
		b := customErr{msg: "boom"}
		if compare.Errors(a, b) {
			t.Errorf("expected same message but different types to be unequal")
		}
	})
}

func TestCompareArrays(t *testing.T) {
	t.Run("equal", func(t *testing.T) {
		if !compare.Arrays([]int{1, 2, 3}, []int{1, 2, 3}) {
			t.Errorf("expected equal")
		}
	})
	t.Run("different order", func(t *testing.T) {
		if compare.Arrays([]int{1, 2, 3}, []int{3, 2, 1}) {
			t.Errorf("ordered compare should be unequal")
		}
	})
	t.Run("different length", func(t *testing.T) {
		if compare.Arrays([]int{1, 2}, []int{1, 2, 3}) {
			t.Errorf("different length should be unequal")
		}
	})
	t.Run("empty", func(t *testing.T) {
		if !compare.Arrays([]int{}, []int{}) {
			t.Errorf("empty arrays should be equal")
		}
	})
}

func TestCompareMaps(t *testing.T) {
	t.Run("equal", func(t *testing.T) {
		a := map[string]int{"x": 1, "y": 2}
		b := map[string]int{"y": 2, "x": 1}
		if !compare.Maps(a, b) {
			t.Errorf("expected equal maps")
		}
	})
	t.Run("different value", func(t *testing.T) {
		if compare.Maps(map[string]int{"x": 1}, map[string]int{"x": 2}) {
			t.Errorf("different value should be unequal")
		}
	})
	t.Run("different key", func(t *testing.T) {
		if compare.Maps(map[string]int{"x": 1}, map[string]int{"y": 1}) {
			t.Errorf("different key should be unequal")
		}
	})
	t.Run("different length", func(t *testing.T) {
		if compare.Maps(map[string]int{"x": 1}, map[string]int{"x": 1, "y": 2}) {
			t.Errorf("different length should be unequal")
		}
	})
	t.Run("empty", func(t *testing.T) {
		if !compare.Maps(map[string]int{}, map[string]int{}) {
			t.Errorf("empty maps should be equal")
		}
	})
}

func TestCompareArraysUnordered_SameLengthDifferentMultiset(t *testing.T) {
	a := []int{1, 1, 2}
	b := []int{1, 2, 2}
	if compare.ArraysUnordered(a, b) {
		t.Errorf("expected multisets to be unequal: %v vs %v", a, b)
	}
}

func TestCompareArraysUnordered_Empty(t *testing.T) {
	if !compare.ArraysUnordered([]int{}, []int{}) {
		t.Errorf("empty arrays should be equal")
	}
}

func TestCompareArraysUnordered(t *testing.T) {
	a := []int{1, 2, 3, 4, 5}
	b := []int{5, 4, 3, 2, 1}
	if !compare.ArraysUnordered(a, b) {
		t.Errorf("Expected arrays to be equal, but they are not: %v != %v", a, b)
	}
}

func TestCompareArraysUnordered_Duplicates(t *testing.T) {
	a := []int{1, 2, 3, 3, 5}
	b := []int{5, 3, 3, 2, 1}
	if !compare.ArraysUnordered(a, b) {
		t.Errorf("Expected arrays with duplicates to be equal, but they are not: %v != %v", a, b)
	}

}

func TestCompareArraysUnordered_DifferentLengths(t *testing.T) {
	a := []int{1, 2, 3, 4, 5}
	b := []int{5, 4, 3, 2}
	if compare.ArraysUnordered(a, b) {
		t.Errorf("Expected arrays of different lengths to not be equal, but they are: %v == %v", a, b)
	}
}
