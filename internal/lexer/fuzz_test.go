package lexer

import (
	"reflect"
	"testing"

	"github.com/ant-src/ant/internal/token"
)

// FuzzLex checks that the lexer never panics, that it ends every token stream with EOF, and
// that positions never go backwards.
func FuzzLex(f *testing.F) {
	for _, seed := range []string{
		"", "var x = 1\n", "func f(a: int): int { return a + 1 }", `"a${1 + "${2}"}b`,
		"0x1f 0b101 1_000 1.5e3 'a'", "// comment\n/* block */ x", "\"unterminated", "`",
		"x |> f(_) ?? y?.z", "type T = { a: int, b: Array[T] }",
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, src string) {
		toks, err := Lex(src)
		if err != nil {
			if _, ok := err.(*Error); !ok {
				t.Fatalf("error of type %T: %v", err, err)
			}
			return
		}
		if len(toks) == 0 || toks[len(toks)-1].Kind != token.EOF {
			t.Fatalf("token stream does not end with EOF: %v", toks)
		}
		var last token.Pos
		for _, tok := range toks {
			if tok.Pos.Line < last.Line || (tok.Pos.Line == last.Line && tok.Pos.Col < last.Col) {
				t.Fatalf("position %v after %v", tok.Pos, last)
			}
			last = tok.Pos
		}
	})
}

// FuzzRegexDivision checks the operand rule of section 2 against the lexer as it was before the
// regular expression literal existed: on every input in which no literal occurs, the two token
// streams are the same, so no program written before the literal changes meaning.
func FuzzRegexDivision(f *testing.F) {
	for _, seed := range []string{
		"a / b / c", "f(x) / 2", "{ x } / 2", "xs[0] / 2", "x /= 2", "(total\n  / count)",
		"[total\n  / count, 1]", "x = 1\n/a/", "return /x/", "f(/x/i)", `"${ /x/ }"`, "(/=/)",
		"/[/]/", "/a\\/b/", "a /\n b", "x // comment\n/ y", "x /* c */ / y", "true / false",
		"}\n/ 2", "null / 1", "x = /a/ / 2",
	} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, src string) {
		with, err := LexWith(src, true)
		if err != nil || hasRegex(with) {
			return
		}
		without, err := LexWith(src, false)
		if err != nil {
			t.Fatalf("the lexer without literals fails where the lexer with them succeeds: %v", err)
		}
		if !reflect.DeepEqual(with, without) {
			t.Fatalf("token streams differ on %q:\n%v\n%v", src, with, without)
		}
	})
}

// hasRegex reports whether a token stream, interpolations included, holds a literal.
func hasRegex(toks []token.Token) bool {
	for _, t := range toks {
		if t.Kind == token.Regex {
			return true
		}
		for _, p := range t.Parts {
			if hasRegex(p.Tokens) {
				return true
			}
		}
	}
	return false
}
