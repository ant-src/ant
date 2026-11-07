package lsp

import (
	"strings"
	"testing"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/loader"
)

// TestComplete drives Complete as a REPL does: the text is the declarations of earlier inputs
// followed by the input being typed, resolved as a session, so a name declared earlier, or
// declared again, completes by its latest type.
func TestComplete(t *testing.T) {
	labels := func(text string) []string {
		cfg := loader.Config{Session: &ast.ModuleScope{Index: map[string]int{}}}
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
	// an empty literal takes its type from the function: an array or a map of anything
	has("[] | Array.", "filter", "map")
	lacks("[] | Array.", "filled", "range")
	has("[] | ", "Array.filter", "println")
	lacks("[] | ", "Array.filled", "String.trim")
	has("Map{} | Map.", "keys", "filter")
	// a left side with no type filters nothing
	has("nope | Array.", "filter", "filled")
	// a function's name finds its qualified name
	has("var xs = [1, 2]\nxs | filt", "Array.filter")
	// the prelude's own functions too, println taking any type
	has("var xs = [1, 2]\nxs | ", "println", "print")
	lacks("var xs = [1, 2]\nxs | ", "panic", "int")
	has("var n = 1\nn | ", "println", "int", "float64")
	has("func twice(n: int): int { n * 2 }\ntw", "twice")
	has("var x = 1\nx + ", "x", "println")
	// a type expression takes a type (section 4, Types as values)
	has("type Point = { x: int }\nvar t = type(", "Point", "int", "Array")
	lacks("type Point = { x: int }\nvar t = type(", "println", "t")
	has("type Point = { x: int }\nvar t = type(Array[", "Point", "int")
	has("type Point = { x: int }\nvar t = type(Map[string, ", "Point", "int")
	has("type Point = { x: int }\nvar t = type(Map[string, Array[", "Point", "int")
	has("type Point = { x: int }\nvar xs = [type(int), ", "println")
	has("type Point = { x: int }\nvar t = type(Po", "Point")
	has("type Point = { x: int }\nvar t = type(Point | ", "int")
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

// TestCompleteParens checks what a function's item inserts: the parentheses of a call on the
// right side of a pipe always, and elsewhere with snippets, unless a function is expected there
// or a parenthesis follows already.
func TestCompleteParens(t *testing.T) {
	var snippets ClientCapabilities
	snippets.TextDocument.Completion.CompletionItem.SnippetSupport = true
	insert := func(caps ClientCapabilities, text string, off int, label string) (string, int) {
		t.Helper()
		cfg := loader.Config{Session: &ast.ModuleScope{Index: map[string]int{}}}
		for _, it := range completeText(cfg, "<input>", text, off, caps) {
			if it.Label == label {
				return it.InsertText, it.InsertTextFormat
			}
		}
		t.Fatalf("%q: %q is not offered", text, label)
		return "", 0
	}
	end := func(s string) int { return len(s) }
	const decls = "func twice(n: int): int { n * 2 }\nfunc zero(): int { 0 }\nfunc add(a: int, b: int): int { a + b }\nvar xs = [1, 2]\n"
	for _, tc := range []struct {
		caps   ClientCapabilities
		text   string
		off    int
		label  string
		want   string
		format int
	}{
		// the right side of a pipe is a call, with or without snippets
		{ClientCapabilities{}, decls + "1 | tw", -1, "twice", "twice()", 0},
		{snippets, decls + "1 | tw", -1, "twice", "twice()", 0},
		{snippets, decls + "1 | ad", -1, "add", "add($0)", 2},
		{ClientCapabilities{}, decls + "1 | ad", -1, "add", "add()", 0},
		{ClientCapabilities{}, decls + "xs | Array.", -1, "map", "map()", 0},
		{snippets, decls + "xs | Array.", -1, "map", "map($0)", 2},
		{snippets, decls + "xs | ", -1, "Array.map", "Array.map($0)", 2},
		// elsewhere only with snippets
		{ClientCapabilities{}, decls + "var y = tw", -1, "twice", "", 0},
		{snippets, decls + "var y = tw", -1, "twice", "twice($0)", 2},
		{snippets, decls + "var y = ze", -1, "zero", "zero()", 0},
		{snippets, decls + "var y = Array.", -1, "map", "map($0)", 2},
		// a function passed as a function argument stays a name
		{snippets, decls + "var ys = Array.map(xs, tw", -1, "twice", "", 0},
		{snippets, decls + "var ys = xs | Array.map(tw", -1, "twice", "", 0},
		{snippets, decls + "var ys = Array.map(tw", -1, "twice", "twice($0)", 2},
		// a parenthesis follows the word
		{snippets, decls + "var y = tw(1)", len(decls + "var y = tw"), "twice", "", 0},
		{snippets, decls + "1 | tw()", len(decls + "1 | tw"), "twice", "", 0},
	} {
		off := tc.off
		if off < 0 {
			off = end(tc.text)
		}
		got, format := insert(tc.caps, tc.text, off, tc.label)
		if got != tc.want || format != tc.format {
			t.Errorf("%q, snippets %v: %s inserts %q (format %d), want %q (format %d)", tc.text[len(decls):], tc.caps.TextDocument.Completion.CompletionItem.SnippetSupport, tc.label, got, format, tc.want, tc.format)
		}
	}
}

// completeItems is completeText for a session input with the cursor at the first occurrence of
// mark, which is removed from text.
func completeItems(t *testing.T, caps ClientCapabilities, text string) []CompletionItem {
	t.Helper()
	off := strings.Index(text, "|>")
	if off < 0 {
		t.Fatalf("no cursor in %q", text)
	}
	text = text[:off] + text[off+2:]
	cfg := loader.Config{Session: &ast.ModuleScope{Index: map[string]int{}}}
	return completeText(cfg, "<input>", text, off, caps)
}

func itemLabels(items []CompletionItem) []string {
	var out []string
	for _, it := range items {
		out = append(out, it.Label)
	}
	return out
}

func contains(xs []string, x string) bool {
	for _, y := range xs {
		if y == x {
			return true
		}
	}
	return false
}

// TestCompletePipeInCall: in a pipe's call whose arguments are written, the functions are still
// those that take the left side.
func TestCompletePipeInCall(t *testing.T) {
	const decls = "var xs = [1, 2]\n"
	for _, tc := range []struct {
		text       string
		has, lacks string
	}{
		{"xs | Array.fil|>ter({ x > 1 })", "filter", "filled"},
		{"xs | Array.|>filter({ x > 1 })", "filter", "filled"},
		{"xs | fil|>({ x > 1 })", "Array.filter", "Array.filled"},
	} {
		got := itemLabels(completeItems(t, ClientCapabilities{}, decls+tc.text))
		if !contains(got, tc.has) || contains(got, tc.lacks) {
			t.Errorf("%q: want %s and not %s among %q", tc.text, tc.has, tc.lacks, got)
		}
	}
}
