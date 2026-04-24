package assert_test

import (
	"fmt"
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

func TestExample(t *testing.T) {
	a := 1
	b := 2
	if a == b {
		t.Errorf("Expected %d to not equal %d", a, b)
	}
}
