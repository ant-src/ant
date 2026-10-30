package lexer

import (
	"fmt"
	"strings"
	"testing"

	"github.com/ant-src/ant/internal/token"
)

// TestSpans checks that every token's span slices back to the token as written, for each kind,
// including the tokens of an interpolation and the parts of a string with interpolations.
func TestSpans(t *testing.T) {
	src := "var x = 0xFF + 1_000 * 1e9 // trailing\n" +
		"/* block\n   comment */ f(\"a\\n\", /^\\d+$/i, \"t ${ a+b } u\", `r\\`)\n" +
		"  e.y ?. z ... !q\n"
	toks, comments, err := LexLayout(src, false)
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	var walk func([]token.Token)
	walk = func(toks []token.Token) {
		for _, tok := range toks {
			if tok.Kind == token.Newline || tok.Kind == token.EOF {
				if tok.Off != tok.End {
					t.Errorf("%v at %d has a span of %d bytes", tok.Kind, tok.Off, tok.End-tok.Off)
				}
				continue
			}
			got = append(got, src[tok.Off:tok.End])
			for _, part := range tok.Parts {
				got = append(got, "part:"+src[part.Off:part.End])
				walk(part.Tokens)
			}
		}
	}
	walk(toks)
	want := []string{
		"var", "x", "=", "0xFF", "+", "1_000", "*", "1e9",
		"f", "(", `"a\n"`, ",", `/^\d+$/i`, ",", `"t ${ a+b } u"`,
		"part:t ", "part:${ a+b }", "a", "+", "b", "part: u", ",", "`r\\`", ")",
		"e", ".", "y", "?.", "z", "...", "!", "q",
	}
	if strings.Join(got, " | ") != strings.Join(want, " | ") {
		t.Errorf("spans:\n got %q\nwant %q", got, want)
	}
	var texts []string
	for _, c := range comments {
		if src[c.Off:c.End] != c.Text {
			t.Errorf("comment %q spans %q", c.Text, src[c.Off:c.End])
		}
		texts = append(texts, c.Text)
	}
	if want := []string{"// trailing", "/* block\n   comment */"}; strings.Join(texts, "|") != strings.Join(want, "|") {
		t.Errorf("comments: got %q, want %q", texts, want)
	}
	if comments[1].Pos != (token.Pos{Line: 2, Col: 1}) {
		t.Errorf("block comment at %v", comments[1].Pos)
	}
}

// TestLexKeepsNoComments checks that the ordinary lexer does not collect comments.
func TestLexKeepsNoComments(t *testing.T) {
	l := &lexer{src: "// a\nx /* b */\n"}
	if _, comments, err := lex(l); err != nil || comments != nil {
		t.Errorf("comments %v, err %v", comments, err)
	}
}

// TestLoneCR checks that a \r not followed by \n is a line break, as \n and \r\n are (section 2):
// it ends a statement, a line comment, and a plain string, and it starts a new line.
func TestLoneCR(t *testing.T) {
	toks, err := Lex("a\rb // c\rc\r\nd")
	if err != nil {
		t.Fatal(err)
	}
	var got []string
	for _, tok := range toks {
		got = append(got, fmt.Sprintf("%v %q %d:%d", tok.Kind, tok.Text, tok.Pos.Line, tok.Pos.Col))
	}
	want := []string{
		`identifier "a" 1:1`, `newline "" 1:2`, `identifier "b" 2:1`, `newline "" 2:7`,
		`identifier "c" 3:1`, `newline "" 3:3`, `identifier "d" 4:1`, `newline "" 4:2`,
		`end of file "" 4:2`,
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("got\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
	if _, err := Lex("/a\rb/"); err == nil || !strings.Contains(err.Error(), "unterminated") {
		t.Errorf("a lone \\r in a regular expression: got %v, want unterminated", err)
	}
	// In a string and a raw string, each line break is one \n.
	for _, src := range []string{"\"a\rb\r\nc\"", "`a\rb\r\nc`", "\"a\rb${1}\r\nc\""} {
		toks, err := Lex(src)
		if err != nil {
			t.Errorf("%q: %v", src, err)
			continue
		}
		var text string
		if toks[0].Kind == token.Template {
			text = toks[0].Parts[0].Text + toks[0].Parts[2].Text
		} else {
			text = toks[0].Text
		}
		if text != "a\nb\nc" {
			t.Errorf("%q: the text is %q, want %q", src, text, "a\nb\nc")
		}
	}
}

// TestShebang checks that a #! first line is skipped in a file and kept as its first comment,
// and that it is not Ant elsewhere.
func TestShebang(t *testing.T) {
	src := "#!/usr/bin/env ant\nx\n"
	toks, comments, err := LexLayout(src, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(comments) != 1 || comments[0].Text != "#!/usr/bin/env ant" {
		t.Errorf("comments %v, want the #! line", comments)
	}
	if toks[0].Text != "x" || toks[0].Pos != (token.Pos{Line: 2, Col: 1}) {
		t.Errorf("first token %v at %v, want x at 2:1", toks[0], toks[0].Pos)
	}
	for _, src := range []string{"#!x\n", " #!x\n", "x\n#!y\n"} {
		shebang := src != "#!x\n"
		if _, _, err := LexLayout(src, shebang); err == nil {
			t.Errorf("LexLayout(%q, %v) accepted the #! line", src, shebang)
		}
	}
}
