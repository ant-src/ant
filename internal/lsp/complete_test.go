package lsp

import (
	"strings"
	"testing"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/loader"
	"github.com/ant-src/ant/parser"
)

// TestComplete drives Complete as a REPL does: the text is the declarations of earlier inputs
// followed by the input being typed, resolved as a session, so a name declared earlier, or
// declared again, completes by its latest type.
func TestComplete(t *testing.T) {
	labels := func(text string) []string {
		cfg := loader.Config{Mode: parser.Session, Session: &ast.ModuleScope{Index: map[string]int{}}}
		var out []string
		for _, it := range Complete(cfg, "<input>", text, len(text)) {
			out = append(out, it.Label)
		}
		return out
	}
	has := func(text string, want ...string) {
		t.Helper()
		got := labels(text)
		for _, w := range want {
			found := false
			for _, l := range got {
				if l == w {
					found = true
				}
			}
			if !found {
				t.Errorf("%q: %q is not offered among %d items: %q", text, w, len(got), got)
			}
		}
	}
	lacks := func(text string, unwanted ...string) {
		t.Helper()
		got := labels(text)
		for _, u := range unwanted {
			for _, l := range got {
				if l == u {
					t.Errorf("%q: %q is offered: %q", text, u, got)
				}
			}
		}
	}
	has("type Point = { x: int, y: int }\nvar p = Point{ x: 1, y: 2 }\np.", "x", "y")
	has("var xs = [1, 2]\nxs | Array.", "map", "filter")
	has("var xs = [1, 2]\nxs | ", "Array.map")
	has("func twice(n: int): int { n * 2 }\ntw", "twice")
	has("var x = 1\nx + ", "x", "println")
	// declared again: the latest declaration decides
	has("var s = 1\nvar s = \"a\"\ns | String.", "length")
	lacks("var s = 1\nvar s = \"a\"\ns | String.", "nonsense")
	has("var n = 1\nvar n = \"a\"\nvar n = [1]\nn | Array.", "map")
	has("import \"../../tests/run/lib/geometry.ant\" { distance, Point }\nvar q = Point{ x: 0.0, y: 0.0 }\nq.", "x", "y")
	// an empty input offers the session's names, the prelude, and keywords
	got := labels("var answer = 42\n")
	if len(got) == 0 || !strings.Contains(strings.Join(got, " "), "answer") {
		t.Errorf("empty input: %q", got)
	}
}
