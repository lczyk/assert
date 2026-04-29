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

func getSourceLine(file string, line int) (string, error) {
	if file == "" || line <= 0 {
		return "", errNoLocation
	}
	var lines []string
	if v, ok := sourceCache.Load(file); ok {
		// Safe on typed-nil: a prior ReadFile error stores []string(nil),
		// and v.([]string) succeeds on that — yields a nil slice, len 0.
		lines = v.([]string)
	} else {
		data, err := os.ReadFile(file)
		if err != nil {
			sourceCache.Store(file, []string(nil))
			return "", err
		}
		lines = strings.Split(string(data), "\n")
		sourceCache.Store(file, lines)
	}
	if line > len(lines) {
		return "", errLineOutOfRange
	}
	return strings.TrimSpace(lines[line-1]), nil
}

// locStr formats "file:line" with the source line snippet on a new
// line when available. Returns an error if the source line could not
// be read; callers should fall back to a plain "file:line" format.
func locStr(file string, line int) (string, error) {
	src, err := getSourceLine(file, line)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%s:%d\n  > %s", file, line, src), nil
}
