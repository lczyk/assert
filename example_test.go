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

func TestExample(t *testing.T) {
	a := 1
	b := 2
	if a == b {
		t.Errorf("Expected %d to not equal %d", a, b)
	}
}
