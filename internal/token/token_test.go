package token

import "testing"

func TestStrings(t *testing.T) {
	kinds := map[Kind]string{
		EOF: "end of file", Newline: "newline", Ident: "identifier", Keyword: "keyword",
		Reserved: "reserved word", Int: "integer", Float: "float", String: "string",
		Template: "string", Op: "operator", Kind(99): "token",
	}
	for k, want := range kinds {
		if got := k.String(); got != want {
			t.Errorf("%d: %q, want %q", k, got, want)
		}
	}
	tokens := map[string]Token{
		`identifier "x"`: {Kind: Ident, Text: "x"},
		"'func'":         {Kind: Keyword, Text: "func"},
		"'class'":        {Kind: Reserved, Text: "class"},
		"'+='":           {Kind: Op, Text: "+="},
		"42":             {Kind: Int, Text: "42", Value: 42},
		"1.5":            {Kind: Float, Text: "1.5", Float: 1.5},
		"string literal": {Kind: String, Text: "s"},
		"end of file":    {Kind: EOF},
		"newline":        {Kind: Newline},
	}
	for want, tok := range tokens {
		if got := tok.String(); got != want {
			t.Errorf("%v: %q, want %q", tok, got, want)
		}
	}
	// a string with interpolations reads as any other string literal
	if got := (Token{Kind: Template}).String(); got != "string literal" {
		t.Errorf("Template: %q, want %q", got, "string literal")
	}
	plus := Token{Kind: Op, Text: "+"}
	if !plus.Is("+") || plus.Is("-") || (Token{Kind: Ident, Text: "if"}).Is("if") || !(Token{Kind: Keyword, Text: "if"}).Is("if") {
		t.Error("Is")
	}
	if !IsKeyword("while") || IsKeyword("x") || !IsReserved("await") || IsReserved("while") {
		t.Error("IsKeyword or IsReserved")
	}
}
