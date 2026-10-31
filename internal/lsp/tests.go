package lsp

import (
	"regexp"
	"sort"
	"strings"

	"github.com/ant-src/ant/ast"
)

// The tests of a test file as the editor sees them (spec/lsp.md section 4): document symbols, and
// a code lens on each test to run or debug it, which the client carries out with ant test and
// ant dap (spec/test.md).

// testItem is a test whose name is known without running anything: a top-level test, or a
// subtest whose name is a string literal inside one, with the -run pattern that selects it.
type testItem struct {
	node     *ast.Test
	name     string
	pattern  string
	children []*testItem
}

// fileTests lists f's tests, top-level first, each with its subtests whose names are literals.
func fileTests(f *ast.File) []*testItem {
	var out []*testItem
	for _, t := range f.Tests {
		out = append(out, testTree(t, ""))
	}
	return out
}

func testTree(t *ast.Test, parent string) *testItem {
	name := t.Name.(*ast.StringLit).Value
	pattern := "^" + regexp.QuoteMeta(name) + "$"
	if parent != "" {
		pattern = parent + "/" + pattern
	}
	it := &testItem{node: t, name: name, pattern: pattern}
	ast.Inspect(t.Body, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.Test:
			if _, ok := x.Name.(*ast.StringLit); ok {
				it.children = append(it.children, testTree(x, pattern))
			}
			return false
		case *ast.FuncLit, *ast.Lambda, *ast.Go, *ast.Defer:
			return false
		}
		return true
	})
	return it
}

// testSymbols are the document symbols of a test file's tests, a subtest under its test.
func (s *Server) testSymbols(d *document, items []*testItem) []DocumentSymbol {
	var out []DocumentSymbol
	for _, it := range items {
		out = append(out, DocumentSymbol{Name: it.name, Detail: "test", Kind: SymbolFunction,
			Range: s.rng(d, it.node.Pos, it.node.End), SelectionRange: s.rng(d, it.node.Name.Position(), it.node.Name.EndPosition()),
			Children: s.testSymbols(d, it.children)})
	}
	return out
}

// sortSymbols orders top-level symbols by where they start, so the tests take their places
// among the declarations.
func sortSymbols(syms []DocumentSymbol) {
	sort.SliceStable(syms, func(i, j int) bool {
		a, b := syms[i].Range.Start, syms[j].Range.Start
		return a.Line < b.Line || a.Line == b.Line && a.Character < b.Character
	})
}

// The commands a code lens names, which the client implements: each takes the file's URI and the
// -run pattern of the test.
const (
	commandRunTest   = "ant.test.run"
	commandDebugTest = "ant.test.debug"
)

// codeLenses answers textDocument/codeLens: "run test" and "debug test" above each test of a
// test file whose name is known.
func (s *Server) codeLenses(p *CodeLensParams) []CodeLens {
	d := s.doc(p.TextDocument.URI)
	out := []CodeLens{}
	if d == nil || !strings.HasSuffix(string(p.TextDocument.URI), "_test.ant") {
		return out
	}
	a := s.analyze(d)
	var walk func([]*testItem)
	walk = func(items []*testItem) {
		for _, it := range items {
			r := s.rng(d, it.node.Pos, it.node.Name.EndPosition())
			args := []any{p.TextDocument.URI, it.pattern}
			out = append(out,
				CodeLens{Range: r, Command: &Command{Title: "run test", Command: commandRunTest, Arguments: args}},
				CodeLens{Range: r, Command: &Command{Title: "debug test", Command: commandDebugTest, Arguments: args}})
			walk(it.children)
		}
	}
	walk(fileTests(a.module().File))
	return out
}
