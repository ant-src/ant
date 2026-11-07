package lsp

import "testing"

// TestSignatureActiveParam counts the arguments by the commas of the call, not those in a
// comment, a string, an interpolation, a regular expression, or markup.
func TestSignatureActiveParam(t *testing.T) {
	const decls = "func add(a: int, b: int, c: string): int { a + b }\n" +
		"func g(r: Regex, n: int): int { n }\n" +
		"func h(x: Html, n: int): int { n }\n"
	for _, tc := range []struct {
		src, needle string
		want        int
	}{
		{"var v = add(1, // one, two\n  2, \"\")", "2, \"", 1},
		{"var v = add(1, 2, /* a, b */ \"\")", "\"\")", 2},
		{"var v = add(1, \"${\"a,b\"}\", 3)", "3)", 2},
		{"var v = g(/a\\(,/, 2)", "2)", 1},
		{"var v = h(<p>a, b</p>, 3)", "3)", 1},
		{"var v = add(1, [2, 3].length, \"\")", "\"\")", 2},
		{"var v = \"${add(1, 2, \"\")}\"", "\"\")", 2},
		{"var v = <p>{add(1, 2, \"\")}</p>", "\"\")", 2},
	} {
		c := start(t, map[string]string{})
		src := decls + tc.src
		uri := c.open("main.ant", src)
		var sh SignatureHelp
		c.call("textDocument/signatureHelp", SignatureHelpParams{TextDocumentPositionParams: c.pos(uri, at(src, tc.needle, 1, 0))}, &sh)
		if len(sh.Signatures) != 1 || sh.ActiveParameter != tc.want {
			t.Errorf("%q: active parameter %d, want %d: %+v", tc.src, sh.ActiveParameter, tc.want, sh)
		}
	}
}
