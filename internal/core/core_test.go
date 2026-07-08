// White-box tests for the failure-rendering internals: the emit
// fallback, the source cache, and the snippet scanner. The public
// assert/require behaviour is covered by the wrapper packages' tests;
// these paths need unreadable or crafted source files, which is only
// reachable from inside the package.
package core

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEmitFallbackPercentInMessage(t *testing.T) {
	// Regression: the fallback used to splice the pre-formatted message
	// into the Errorf format string, re-interpreting any literal % in it.
	var got string
	fail := func(format string, args ...any) { got = fmt.Sprintf(format, args...) }
	missing := filepath.Join(t.TempDir(), "missing.go")
	emit(fail, missing, 3, "100% full")
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
