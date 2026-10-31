package repl

import (
	"strings"
	"testing"
)

func TestPaint(t *testing.T) {
	src := "var x = 1"
	var sb strings.Builder
	ansiTheme.paint(&sb, src, classes(src), 0)
	if got := sb.String(); got != "\x1b[35mvar\x1b[0m x = \x1b[33m1\x1b[0m" {
		t.Errorf("paint: %q", got)
	}
	sb.Reset()
	plainTheme.paint(&sb, src, classes(src), 0)
	if sb.String() != src {
		t.Errorf("plain paint: %q", sb.String())
	}
	// a window of the source keeps the classes of its bytes
	sb.Reset()
	ansiTheme.paint(&sb, "x = 1", classes(src), 4)
	if got := sb.String(); got != "x = \x1b[33m1\x1b[0m" {
		t.Errorf("paint from 4: %q", got)
	}
}
