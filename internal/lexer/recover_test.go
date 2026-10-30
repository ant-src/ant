package lexer

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/ant-src/ant/internal/token"
)

// corpusFiles lists every .ant file of the corpus and the examples.
func corpusFiles(t *testing.T) map[string]string {
	t.Helper()
	files := map[string]string{}
	for _, pattern := range []string{"../../tests/run/*.ant", "../../tests/check/*.ant", "../../tests/check/lib/*.ant",
		"../../tests/unchecked/*.ant", "../../tests/stdlib/*.ant", "../../tests/syntax/*.ant", "../../examples/*.ant",
		"../../examples/lib/*.ant", "../../internal/format/testdata/*.ant", "../prelude/prelude.d.ant"} {
		paths, _ := filepath.Glob(pattern)
		for _, path := range paths {
			src, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			files[path] = string(src)
		}
	}
	if len(files) < 100 {
		t.Fatalf("found only %d corpus files", len(files))
	}
	return files
}

// TestRecoverAgrees checks that on a source without lexical errors the recovering lexer produces
// the tokens and comments the strict one does, and no errors.
func TestRecoverAgrees(t *testing.T) {
	for path, src := range corpusFiles(t) {
		toks, comments, err := LexLayout(src, true)
		if err != nil {
			continue
		}
		rtoks, rcomments, errs := LexRecover(src, true)
		if len(errs) > 0 {
			t.Errorf("%s: recovering lexer reports %v", path, errs[0])
			continue
		}
		if !reflect.DeepEqual(toks, rtoks) {
			t.Errorf("%s: tokens differ", path)
		}
		if !reflect.DeepEqual(comments, rcomments) {
			t.Errorf("%s: comments differ", path)
		}
	}
}

// TestRecoverStrictErrors checks that on a source with a lexical error the recovering lexer
// reports the strict lexer's error first, and goes on.
func TestRecoverStrictErrors(t *testing.T) {
	for path, src := range corpusFiles(t) {
		_, _, err := LexLayout(src, true)
		if err == nil {
			continue
		}
		_, _, errs := LexRecover(src, true)
		if len(errs) == 0 {
			t.Errorf("%s: strict lexer fails with %v, recovering one reports nothing", path, err)
			continue
		}
		if errs[0].Error() != err.Error() {
			t.Errorf("%s: first error %v, strict %v", path, errs[0], err)
		}
	}
}

// TestRecoverTruncated checks that the recovering lexer never panics and always ends its stream
// with an EOF, on every prefix of every corpus file cut at a token boundary and at every byte of
// the smaller files.
func TestRecoverTruncated(t *testing.T) {
	for path, src := range corpusFiles(t) {
		toks, _, _ := LexRecover(src, true)
		cuts := map[int]bool{}
		stride := 1
		if len(src) > 8000 {
			stride = 10 // the prelude: a sample of its tokens
		}
		for i, tok := range toks {
			if i%stride != 0 {
				continue
			}
			cuts[tok.Off] = true
			cuts[tok.End] = true
			if tok.Off+1 < len(src) {
				cuts[tok.Off+1] = true
			}
		}
		if len(src) < 1200 {
			for i := range src {
				cuts[i] = true
			}
		}
		for cut := range cuts {
			if cut > len(src) {
				continue
			}
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("%s cut at %d: panic: %v", path, cut, r)
					}
				}()
				toks, _, _ := LexRecover(src[:cut], true)
				if len(toks) == 0 || toks[len(toks)-1].Kind != token.EOF {
					t.Errorf("%s cut at %d: no EOF", path, cut)
				}
			}()
		}
	}
}

// TestRecoverCases checks what the recovering lexer makes of particular errors.
// TestRecoverNestedOpen checks that strings left open inside the interpolations of strings left
// open are each read again once, not once for every level around them.
func TestRecoverNestedOpen(t *testing.T) {
	src := strings.Repeat(`"${`, 200) + `"x` + strings.Repeat("\n}", 200)
	start := time.Now()
	LexRecover(src, false)
	if d := time.Since(start); d > 2*time.Second {
		t.Errorf("200 open strings, each in the last one's interpolation, took %v", d)
	}
}

func TestRecoverCases(t *testing.T) {
	cases := []struct {
		name, src string
		errs      []string // substrings of the errors, in order
		kinds     string   // the token kinds and texts that come out, Newline and EOF left out
	}{
		{"unterminated string", "var s = \"abc\nvar t = 1", []string{"unterminated string"},
			`var s = "abc" var t = 1`},
		{"bad escape kept", `var s = "a\qb"`, []string{"bad escape"}, `var s = "a\qb"`},
		{"backslash ends the first line", "var s = \"abc\\\nvar y = 1\nvar z = 2", []string{"unterminated string"},
			`var s = "abc" var y = 1 var z = 2`},
		{"backslash ends the source", "var s = \"abc\\", []string{"unterminated string"}, `var s = "abc"`},
		{"backslash before a line break", "var s = \"a\\\nb\"", []string{"bad escape: a \\ before a line break"},
			"var s = \"a\\\nb\""},
		{"unterminated interpolating string", "var s = \"a ${x} b", []string{"unterminated string"}, "var s = `…`"},
		{"unterminated interpolation", "var s = \"a ${x", []string{"unterminated ${ in string"}, "var s = `…`"},
		{"unterminated string ends its line", "var s = \"a ${x} b\nvar y = 1", []string{"unterminated string"}, "var s = `…` var y = 1"},
		{"unterminated raw string ends its line", "var s = `a\\\nvar y = 1", []string{"unterminated raw string"}, `var s = "a\" var y = 1`},
		{"escaped backtick", "var s = \"a\\`b\"", []string{"a backtick needs no escape"}, "var s = \"a`b\""},
		{"unterminated comment", "var x = 1 /* open\nvar y", []string{"unterminated block comment"}, `var x = 1`},
		{"unexpected char", "var x = 1 @ 2", []string{"unexpected character"}, `var x = 1 @ 2`},
		{"malformed number", "var x = 0x", []string{"malformed hexadecimal literal"}, `var x = 0x`},
		{"regex flag", "var r = /a/g", []string{"there is no g flag"}, `var r = /a/`},
		{"regex unterminated", "var r = /ab\nvar y = 1", []string{"unterminated regular expression"}, `var r = /ab var y = 1`},
		{"markup unterminated start tag", "var h = <div class=\"a\"", []string{"unterminated start tag <div>"}, `var h = <div>`},
		{"markup never closed", "var h = <div>hi", []string{"<div> is never closed"}, `var h = <div>`},
		{"markup mismatched end tag closes outer", "var h = <div><span>a</div>", []string{"</div> does not close <span>"}, `var h = <div>`},
		{"markup mismatched end tag closes inner", "var h = <div><span>a</b></div>", []string{"</b> does not close <span>"}, `var h = <div>`},
		{"markup void with content", "var h = <br></br>", []string{"<br> is a void element"}, `var h = <br>`},
		{"markup brace in text", "var h = <p>}</p>", []string{"a } in text"}, `var h = <p>`},
		{"markup comment", "var h = <p><!-- c -->x</p>", []string{"an HTML comment is written"}, `var h = <p>`},
		{"markup unknown reference", "var h = <p>&copy;</p>", []string{"not one of the named references"}, `var h = <p>`},
		{"markup single quotes", "var h = <p class='a'>x</p>", []string{"double quotes"}, `var h = <p>`},
		{"markup code attribute", "var h = <p onclick={f}>x</p>", []string{"takes only literal text"}, `var h = <p>`},
		{"markup unterminated start then tag", "var h = <div\n<p>x</p>", []string{"unexpected '<' in the start tag <div>", "<div> is never closed"}, `var h = <div>`},
		{"two errors", "var s = `a\nvar t = \"b\nvar u = 1", []string{"unterminated raw string", "unterminated string"}, `var s = "a" var t = "b" var u = 1`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			toks, _, errs := LexRecover(c.src, true)
			var got []string
			for _, e := range errs {
				got = append(got, e.Msg)
			}
			if len(got) != len(c.errs) {
				t.Fatalf("errors %q, want %q", got, c.errs)
			}
			for i, want := range c.errs {
				if !strings.Contains(got[i], want) {
					t.Errorf("error %d = %q, want one containing %q", i, got[i], want)
				}
			}
			if kinds := describe(toks); kinds != c.kinds {
				t.Errorf("tokens %s, want %s", kinds, c.kinds)
			}
		})
	}
}

// describe renders tokens for the recovery cases: words and operators as written, strings
// quoted, strings with interpolations as `…`, regexes between slashes, markup as its root tag, Bad as the text.
func describe(toks []token.Token) string {
	var parts []string
	for _, t := range toks {
		switch t.Kind {
		case token.Newline, token.EOF:
		case token.String:
			parts = append(parts, `"`+t.Text+`"`)
		case token.Template:
			parts = append(parts, "`…`")
		case token.Regex:
			parts = append(parts, "/"+t.Text+"/")
		case token.Html:
			if len(t.Markup.Nodes) == 1 && t.Markup.Nodes[0].Kind == token.MarkupElement {
				parts = append(parts, "<"+t.Markup.Nodes[0].Name+">")
			} else {
				parts = append(parts, "<>")
			}
		default:
			parts = append(parts, t.Text)
		}
	}
	return strings.Join(parts, " ")
}
