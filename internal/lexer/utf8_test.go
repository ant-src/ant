package lexer

import "testing"

// TestInvalidUTF8 checks that a source that is not UTF-8 is a syntax error at its first bad byte,
// so no string literal carries such bytes into a program.
func TestInvalidUTF8(t *testing.T) {
	for _, tc := range []struct{ src, want string }{
		{"var s = \"a\xffb\"", "1:11: the source is not UTF-8 text at the byte 0xFF"},
		{"var s = 1\n// größe \xe2\x82\n", "2:12: the source is not UTF-8 text at the byte 0xE2"},
		{"\xed\xa0\x80", "1:1: the source is not UTF-8 text at the byte 0xED"},
	} {
		_, err := Lex(tc.src)
		if err == nil || err.Error() != tc.want {
			t.Errorf("Lex(%+q) = %v, want %s", tc.src, err, tc.want)
		}
		if _, _, err := LexLayout(tc.src); err == nil {
			t.Errorf("LexLayout(%+q) accepted it", tc.src)
		}
	}
	if _, err := Lex("var s = \"größe 🐜\"\n"); err != nil {
		t.Errorf("Lex rejected UTF-8: %v", err)
	}
}
