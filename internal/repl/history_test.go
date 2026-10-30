package repl

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHistory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "history")
	h := loadHistory(path)
	for _, l := range []string{"one", "", "  ", "two", "two", "three"} {
		h.add(l)
	}
	if strings.Join(h.lines, ",") != "one,two,three" {
		t.Errorf("lines %q", h.lines)
	}
	if data, _ := os.ReadFile(path); string(data) != "one\ntwo\nthree\n" {
		t.Errorf("file %q", data)
	}
	// a file grown past historySize is cut back when it is loaded
	var sb strings.Builder
	for i := 0; i < historySize+50; i++ {
		sb.WriteString("line\n")
	}
	sb.WriteString("last\n")
	os.WriteFile(path, []byte(sb.String()), 0o600)
	h = loadHistory(path)
	if len(h.lines) != historySize || h.lines[historySize-1] != "last" {
		t.Errorf("%d lines, last %q", len(h.lines), h.lines[len(h.lines)-1])
	}
	if data, _ := os.ReadFile(path); strings.Count(string(data), "\n") != historySize {
		t.Errorf("file holds %d lines", strings.Count(string(data), "\n"))
	}
	// no file: in memory only
	h = loadHistory("")
	h.add("x")
	if len(h.lines) != 1 {
		t.Errorf("lines %q", h.lines)
	}
}
