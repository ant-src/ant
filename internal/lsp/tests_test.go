package lsp

import (
	"strings"
	"testing"
)

func TestTestFile(t *testing.T) {
	lib := "type Stack = { items: Array[int] }\n\nfunc secret(): int { 1 }\n"
	src := `import "./helpers_test.ant" { h }

// The basics.
test "basics" {
  assert(secret() == h())
  for i in [1, 2] {
    test "row ${i}" {}
  }
  test "a.b" {
    go { test "hidden" {} }
  }
}

func helper(): int { 1 }

test "second" {}
`
	c := start(t, map[string]string{"stack.ant": lib, "helpers_test.ant": "export func h(): int { 1 }\n"})
	uri := c.open("stack_test.ant", src)
	// the go block's test is a syntax error; everything else is clean
	diags := c.diagnostics(uri)
	if len(diags) != 1 || !strings.Contains(diags[0].Message, "runs as a subtest") {
		t.Fatalf("diagnostics: %+v", diags)
	}
	src = strings.Replace(src, `go { test "hidden" {} }`, `go { println(1) }`, 1)
	c.change(uri, 2, src)
	if diags := c.diagnostics(uri); len(diags) != 0 {
		t.Fatalf("diagnostics: %+v", diags)
	}

	var syms []DocumentSymbol
	c.call("textDocument/documentSymbol", DocumentSymbolParams{TextDocument: TextDocumentIdentifier{URI: uri}}, &syms)
	var names []string
	for _, s := range syms {
		names = append(names, s.Name)
	}
	if got := strings.Join(names, " "); got != "./helpers_test.ant basics helper second" {
		t.Errorf("symbols: %s", got)
	}
	if len(syms) > 1 && (len(syms[1].Children) != 1 || syms[1].Children[0].Name != "a.b" || syms[1].Detail != "test") {
		t.Errorf("basics: %+v", syms[1])
	}

	var lenses []CodeLens
	c.call("textDocument/codeLens", CodeLensParams{TextDocument: TextDocumentIdentifier{URI: uri}}, &lenses)
	var got []string
	for _, l := range lenses {
		got = append(got, l.Command.Title+" "+l.Command.Command+" "+l.Command.Arguments[1].(string))
	}
	want := []string{
		`run test ant.test.run ^basics$`,
		`debug test ant.test.debug ^basics$`,
		`run test ant.test.run ^basics$/^a\.b$`,
		`debug test ant.test.debug ^basics$/^a\.b$`,
		`run test ant.test.run ^second$`,
		`debug test ant.test.debug ^second$`,
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("code lenses:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
	if len(lenses) > 0 && lenses[0].Range.Start != at(src, `test "basics"`, 1, 0) {
		t.Errorf("lens range: %+v", lenses[0].Range)
	}
	// a file that is not a test file has none
	other := c.open("stack.ant", lib)
	lenses = nil
	c.call("textDocument/codeLens", CodeLensParams{TextDocument: TextDocumentIdentifier{URI: other}}, &lenses)
	if len(lenses) != 0 {
		t.Errorf("lenses in a library: %+v", lenses)
	}
}
