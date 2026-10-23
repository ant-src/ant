package lexer

import (
	"strings"
	"testing"

	"github.com/ant-src/ant/internal/token"
)

// TestSpans checks that every token's span slices back to the token as written, for each kind,
// including the tokens of an interpolation and a template's parts.
func TestSpans(t *testing.T) {
	src := "var x = 0xFF + 1_000 * 1e9 // trailing\n" +
		"/* block\n   comment */ f(\"a\\n\", /^\\d+$/i, `t ${ a+b } u`)\n" +
		"  e.y ?. z ... !q\n"
	toks, comments, err := LexLayout(src)
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
		"f", "(", `"a\n"`, ",", `/^\d+$/i`, ",", "`t ${ a+b } u`",
		"part:t ", "part:${ a+b }", "a", "+", "b", "part: u", ")",
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
