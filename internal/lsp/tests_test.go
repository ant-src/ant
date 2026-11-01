package lsp

import (
	"strings"
	"testing"
)

func TestTestFile(t *testing.T) {
	lib := "type Stack = { items: Array[int] }\n\nfunc secret(): int { 1 }\n"
	src := `import "./helpers_test.ant" { h }
import "./stack.ant" { secret }

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

test "x/y" {
  test "c" {}
}
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
	if got := strings.Join(names, " "); got != "./helpers_test.ant ./stack.ant basics helper second x/y" {
		t.Errorf("symbols: %s", got)
	}
	if len(syms) > 2 && (len(syms[2].Children) != 1 || syms[2].Children[0].Name != "a.b" || syms[2].Detail != "test") {
		t.Errorf("basics: %+v", syms[2])
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
		`run test ant.test.run ^basics$/^a\.b( #[0-9]+)?$`,
		`debug test ant.test.debug ^basics$/^a\.b( #[0-9]+)?$`,
		`run test ant.test.run ^second$`,
		`debug test ant.test.debug ^second$`,
		`run test ant.test.run ^x\/y$`,
		`debug test ant.test.debug ^x\/y$`,
		`run test ant.test.run ^x\/y$/^c( #[0-9]+)?$`,
		`debug test ant.test.debug ^x\/y$/^c( #[0-9]+)?$`,
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

// TestRenameReachesTestFile: a test file may import any top-level declaration of a module
// (section 10, Tests), so renaming one in the library renames it in the test file, private or not.
func TestRenameReachesTestFile(t *testing.T) {
	lib := "func secret(n: int): int { n }\n\nexport func open(): int { secret(n: 1) }\n"
	test := "import \"./stack.ant\" { secret, open }\n\ntest \"a\" {\n  assert(secret(n: 2) == open())\n}\n"
	c := start(t, map[string]string{"stack.ant": lib, "stack_test.ant": test, "other.ant": "func secret(): int { 3 }\n"})
	libURI := c.open("stack.ant", lib)
	testURI := c.uri("stack_test.ant")
	cases := []struct {
		needle, name string
		edits        int // in the test file
	}{
		{"secret(n:", "hidden", 2}, // a private function: its import and its call
		{"n: int", "m", 1},         // a parameter of one, which a named argument spells
		{"open()", "opened", 2},    // an export
	}
	for _, k := range cases {
		var edit WorkspaceEdit
		c.call("textDocument/rename", RenameParams{TextDocumentPositionParams: c.pos(libURI, at(lib, k.needle, 1, 0)), NewName: k.name}, &edit)
		if len(edit.Changes[testURI]) != k.edits {
			t.Errorf("renaming %s: %+v", k.needle, edit.Changes)
		}
		if _, touched := edit.Changes[c.uri("other.ant")]; touched {
			t.Errorf("renaming %s touched an unrelated module", k.needle)
		}
	}
}
