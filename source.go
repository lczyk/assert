package assert

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"
)

// sourceCache memoises file contents split into lines, keyed by path.
// Never evicted — fine for test processes (bounded set of source files,
// short-lived), would grow unbounded in a long-running daemon.
// Concurrent first-read of same file may ReadFile twice; last Store wins.
// Harmless — both reads produce identical content.
var sourceCache sync.Map // map[string][]string

var (
	errNoLocation     = errors.New("no file/line")
	errLineOutOfRange = errors.New("line out of range")
)

func loadSource(file string) ([]string, error) {
	if v, ok := sourceCache.Load(file); ok {
		// Safe on typed-nil: a prior ReadFile error stores []string(nil),
		// and v.([]string) succeeds on that — yields a nil slice, len 0.
		return v.([]string), nil
	}
	data, err := os.ReadFile(file)
	if err != nil {
		sourceCache.Store(file, []string(nil))
		return nil, err
	}
	lines := strings.Split(string(data), "\n")
	sourceCache.Store(file, lines)
	return lines, nil
}

// getSourceSnippet returns the source line at `line`, plus any
// continuation lines, until brace/paren/bracket depth returns to zero.
// Best-effort — handles strings and runes; not aware of comments.
func getSourceSnippet(file string, line int) (string, error) {
	if file == "" || line <= 0 {
		return "", errNoLocation
	}
	lines, err := loadSource(file)
	if err != nil {
		return "", err
	}
	if line > len(lines) {
		return "", errLineOutOfRange
	}
	depth := 0
	inStr := false
	var quote byte
	var out []string
	for i := line - 1; i < len(lines); i++ {
		s := lines[i]
		out = append(out, strings.TrimSpace(s))
		for j := 0; j < len(s); j++ {
			c := s[j]
			if inStr {
				if c == '\\' {
					j++
					continue
				}
				if c == quote {
					inStr = false
				}
				continue
			}
			switch c {
			case '"', '\'', '`':
				inStr = true
				quote = c
			case '(', '[', '{':
				depth++
			case ')', ']', '}':
				depth--
			}
		}
		if depth <= 0 {
			break
		}
	}
	return strings.Join(out, "\n"), nil
}

// loc_str formats "file:line" with the source snippet on following
// indented lines. Returns an error if the source could not be read;
// callers should fall back to a plain "file:line" format.
func loc_str(file string, line int) (string, error) {
	src, err := getSourceSnippet(file, line)
	if err != nil {
		return "", err
	}
	indented := strings.ReplaceAll(src, "\n", "\n  > ")
	return fmt.Sprintf("%s:%d\n  > %s", file, line, indented), nil
}
