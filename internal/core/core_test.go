// White-box tests for the failure-rendering internals: the emit
// fallback, the source cache, and the snippet scanner. The public
// assert/require behaviour is covered by the wrapper packages' tests;
// these paths need unreadable or crafted source files, which is only
// reachable from inside the package.
package core

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"unicode/utf8"
)

func TestRenderFallbackPercentInMessage(t *testing.T) {
	// Regression: the fallback used to splice the pre-formatted message
	// into a format string, re-interpreting any literal % in it.
	missing := filepath.Join(t.TempDir(), "missing.go")
	got := render(missing, 3, "100% full")
	want := "100% full in " + missing + ":3"
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestLoadSourceFailureCached(t *testing.T) {
	// Regression: a failed read used to be cached as nil lines with a nil
	// error; the second (cached) call must also report the failure.
	missing := filepath.Join(t.TempDir(), "missing.go")
	if _, err := loadSource(missing); err == nil {
		t.Errorf("expected error on first read of missing file")
	}
	if _, err := loadSource(missing); err == nil {
		t.Errorf("expected error on second (cached) read of missing file")
	}
}

func TestSnippetQuoteStateResetsPerLine(t *testing.T) {
	// Regression: an apostrophe in a line comment used to leave the
	// scanner in in-string state across lines, swallowing the closing
	// bracket of a multi-line call so the snippet ran on past it.
	src := strings.Join([]string{
		"That(t, x, // don't",
		"\t1+1 == 2,",
		")",
		"trailing()",
	}, "\n")
	path := filepath.Join(t.TempDir(), "fixture.go")
	if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
		t.Fatalf("cannot write fixture: %v", err)
	}
	snippet, err := getSourceSnippet(path, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(snippet, ")") {
		t.Errorf("expected snippet to reach the closing bracket, got %q", snippet)
	}
	if strings.Contains(snippet, "trailing") {
		t.Errorf("expected snippet to stop at the closing bracket, got %q", snippet)
	}
}

func writeFixture(t *testing.T, lines ...string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "fixture.go")
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0o644); err != nil {
		t.Fatalf("cannot write fixture: %v", err)
	}
	return path
}

func TestSnippetIgnoresBracketsInComments(t *testing.T) {
	t.Run("line comment after a complete call", func(t *testing.T) {
		path := writeFixture(t,
			"That(t, x) // see foo(",
			"trailing()",
		)
		snippet, err := getSourceSnippet(path, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if strings.Contains(snippet, "trailing") {
			t.Errorf("expected snippet to stop after the call, got %q", snippet)
		}
	})
	t.Run("closing bracket inside a comment does not end the call", func(t *testing.T) {
		path := writeFixture(t,
			"That(t,",
			"\tx == y, // this ) is not real",
			")",
			"trailing()",
		)
		snippet, err := getSourceSnippet(path, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !strings.HasSuffix(snippet, "\n)") {
			t.Errorf("expected snippet to reach the real closing bracket, got %q", snippet)
		}
	})
	t.Run("block comment", func(t *testing.T) {
		path := writeFixture(t,
			"That(t, /* bounds ( */ x)",
			"trailing()",
		)
		snippet, err := getSourceSnippet(path, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if strings.Contains(snippet, "trailing") {
			t.Errorf("expected snippet to stop after the call, got %q", snippet)
		}
	})
}

func TestSnippetIsCapped(t *testing.T) {
	lines := []string{"That(t, ("}
	for i := 0; i < 2*maxSnippetLines; i++ {
		lines = append(lines, "\tx,")
	}
	path := writeFixture(t, lines...)
	snippet, err := getSourceSnippet(path, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := strings.Split(snippet, "\n")
	if len(got) != maxSnippetLines+1 || got[len(got)-1] != "..." {
		t.Errorf("expected %d lines plus an ellipsis, got %d lines ending %q", maxSnippetLines, len(got), got[len(got)-1])
	}
}

func TestTruncate(t *testing.T) {
	short := strings.Repeat("a", maxValueLen)
	if Truncate(short) != short {
		t.Errorf("expected a value at the cap to be untouched")
	}
	long := strings.Repeat("a", maxValueLen) + "b"
	got := Truncate(long)
	if !strings.HasPrefix(got, short) || !strings.HasSuffix(got, fmt.Sprintf("(truncated, %d bytes total)", len(long))) {
		t.Errorf("expected cap plus a size note, got %q", got[len(got)-60:])
	}
	// The cut must not split a multi-byte rune.
	runes := strings.Repeat("\u00e9", maxValueLen) // 2 bytes each
	if !utf8.ValidString(Truncate(runes)) {
		t.Errorf("expected truncation on a rune boundary")
	}
}

func TestDescribeNonNilPointerChain(t *testing.T) {
	x := 7
	p := &x
	pp := &p
	if got := DescribeNonNil(pp); !strings.Contains(got, "'7' (**int)") {
		t.Errorf("expected the value behind the pointer chain, got %q", got)
	}
	var nilPtr *int
	outer := &nilPtr
	if got := DescribeNonNil(outer); got != "non-nil **int pointing at a nil *int" {
		t.Errorf("expected the nil inner pointer to be named, got %q", got)
	}
}

func TestDescribeErrStdlibTypesUntagged(t *testing.T) {
	joined := errors.Join(errors.New("a"), errors.New("b"))
	if got := DescribeErr(joined); strings.Contains(got, "joinError") {
		t.Errorf("expected no type tag for errors.Join, got %q", got)
	}
	multi := fmt.Errorf("%w and %w", errors.New("a"), errors.New("b"))
	if got := DescribeErr(multi); strings.Contains(got, "wrapErrors") {
		t.Errorf("expected no type tag for multi-%%w, got %q", got)
	}
}
