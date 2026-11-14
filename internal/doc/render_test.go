package doc

import (
	"regexp"
	"strings"
	"testing"

	"github.com/ant-src/ant/internal/prelude"
)

var ansi = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func TestDocLayout(t *testing.T) {
	tests := []struct {
		name string
		doc  string
		want string
	}{
		{"source lines", "One  two\nthree four.\n\n    f(1)  // example\n", "    One  two\n    three four.\n\n        f(1)  // example\n\n"},
		{"paragraphs", "One two\nthree four.\n\nNext paragraph.", "    One two\n    three four.\n\n    Next paragraph.\n"},
		{"long line", strings.Repeat("word ", 30), "    " + strings.Repeat("word ", 30) + "\n"},
		{"code without blank lines", "Before.\n    f(1)\nAfter.", "    Before.\n        f(1)\n    After.\n"},
		{"tab indentation", "Example:\n\tf(1)\n\tf(2)", "    Example:\n    \tf(1)\n    \tf(2)\n"},
		{"unicode", "café é 界界 👩‍💻 done.", "    café é 界界 👩‍💻 done.\n"},
		{"blank lines", "First.\n\n\nLast.\n", "    First.\n\n\n    Last.\n\n"},
		{"empty", "", "\n"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// Coloring must not affect line breaks, indentation, or the text itself.
			for _, color := range []bool{false, true} {
				p := &Printer{Color: color}
				var out strings.Builder
				p.PrintIndented(&out, tc.doc)
				got := out.String()
				if !color && strings.Contains(got, "\x1b") {
					t.Errorf("plain output contains escapes: %q", got)
				}
				if got := ansi.ReplaceAllString(got, ""); got != tc.want {
					t.Errorf("color %v:\n got %q\nwant %q", color, got, tc.want)
				}
			}
		})
	}
}

func TestHighlightDoc(t *testing.T) {
	const sig = "func greet(name: string): string"
	const text = "greet returns a message for name.\n\n" +
		"    let n = 42 // count\n" +
		"    let s = `first\n\n    last`\n" +
		"    /* comment\n    continued */\n" +
		"    println(s)"
	p := &Printer{Color: true}
	var out strings.Builder
	p.PrintDoc(&out, sig, text)
	got := out.String()
	for _, want := range []string{
		"\x1b[35mfunc\x1b[0m", "\x1b[34mgreet\x1b[0m", "\x1b[36mstring\x1b[0m",
		"\x1b[35mlet\x1b[0m", "\x1b[33m42\x1b[0m", "\x1b[90m// count\x1b[0m",
		"\x1b[32m`first\x1b[0m\n\n    \x1b[32m    last`\x1b[0m",
		"\x1b[90m/* comment\x1b[0m\n    \x1b[90m    continued */\x1b[0m",
		"\x1b[34mprintln\x1b[0m",
		"    greet returns a message for name.\n",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("output lacks %q:\n%q", want, got)
		}
	}
	p.Color = false
	out.Reset()
	p.PrintDoc(&out, sig, text)
	if plain := out.String(); ansi.ReplaceAllString(got, "") != plain {
		t.Errorf("highlighting changed the text:\ncolored %q\n  plain %q", got, plain)
	}
	// Signatures preserve their source layout.
	if !strings.HasPrefix(out.String(), sig+"\n\n") {
		t.Errorf("signature changed: %q", out.String())
	}
	out.Reset()
	p.PrintDoc(&out, sig, "")
	if out.String() != sig+"\n" {
		t.Errorf("undocumented declaration: %q", out.String())
	}
}

func TestNamespaceRendering(t *testing.T) {
	f, err := prelude.Decls()
	if err != nil {
		t.Fatal(err)
	}
	p := &Printer{Color: true}
	var out strings.Builder
	if !p.Print(&out, f.Decls, "array", f.NamespaceDocs) {
		t.Fatal("Array was not found")
	}
	got := out.String()
	if !strings.HasPrefix(got, "\x1b[1mArray\x1b[0m\n\n") || !strings.Contains(got, "\x1b[35mfunc\x1b[0m") {
		t.Errorf("namespace heading or members lack highlighting: %q", got)
	}
	p.Color = false
	out.Reset()
	p.Print(&out, f.Decls, "array", f.NamespaceDocs)
	if ansi.ReplaceAllString(got, "") != out.String() {
		t.Error("color changed the namespace's text or layout")
	}
}
