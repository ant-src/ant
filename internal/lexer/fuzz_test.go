package lexer

import (
	"testing"

	"github.com/ant-src/ant/internal/token"
)

// FuzzLex checks that the lexer never panics, that it ends every token stream with EOF, and
// that positions never go backwards.
func FuzzLex(f *testing.F) {
	for _, seed := range []string{
		"", "var x = 1\n", "func f(a: int): int { return a + 1 }", "`a${1 + `${2}`}b`",
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
