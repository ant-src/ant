package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ant-src/ant/ast"
)

// corpus lists the programs the span tests run over: every .ant file of the test corpus, the
// examples, and the formatter's golden files, skipping those that do not parse.
func corpus(t *testing.T) map[string]string {
	t.Helper()
	files := map[string]string{}
	for _, pattern := range []string{"../tests/run/*.ant", "../tests/check/*.ant", "../tests/check/lib/*.ant",
		"../tests/unchecked/*.ant", "../tests/stdlib/*.ant", "../examples/*.ant", "../examples/lib/*.ant",
		"../internal/format/testdata/*.ant"} {
		paths, _ := filepath.Glob(pattern)
		for _, path := range paths {
			src, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			files[path] = string(src)
		}
	}
	if len(files) < 100 {
		t.Fatalf("found only %d corpus files", len(files))
	}
	return files
}

// lineTable maps a position to its byte offset in src.
type lineTable []int

func newLineTable(src string) lineTable {
	lines := lineTable{0}
	for i := 0; i < len(src); i++ {
		if src[i] == '\n' || src[i] == '\r' && (i+1 == len(src) || src[i+1] != '\n') {
			lines = append(lines, i+1)
		}
	}
	return lines
}

func (lt lineTable) offset(p ast.Pos) int { return lt[p.Line-1] + p.Col - 1 }

// textAt is the source between two positions.
func textAt(src string, lt lineTable, from, to ast.Pos) string {
	return src[lt.offset(from):lt.offset(to)]
}

// TestSpans checks that every node has an end after its start, that each child lies within its
// parent, and that the name positions point at the names.
func TestSpans(t *testing.T) {
	files := corpus(t)
	for path, text := range files {
		f, err := Parse(path, text, Program)
		if err != nil {
			continue // the syntax tests
		}
		checkSpans(t, path, text, f)
	}
	preludeSrc, _ := os.ReadFile("../internal/prelude/prelude.d.ant")
	pf, err := Parse("prelude.d.ant", string(preludeSrc), Declarations)
	if err != nil {
		t.Fatal(err)
	}
	checkSpans(t, "prelude.d.ant", string(preludeSrc), pf)
}

func checkSpans(t *testing.T, path, src string, f *ast.File) {
	t.Helper()
	lt := newLineTable(src)
	nodes := 0
	var check func(n ast.Node)
	check = func(n ast.Node) {
		ast.Inspect(n, func(c ast.Node) bool {
			if c == n {
				return true
			}
			nodes++
			start, end := ast.Start(c), c.EndPosition()
			if end == (ast.Pos{}) {
				t.Errorf("%s: %T at %v has no end", path, c, start)
				return false
			}
			if end.Before(start) {
				t.Errorf("%s: %T at %v ends at %v, before it starts", path, c, start, end)
			}
			if ps, pe := ast.Start(n), n.EndPosition(); start.Before(ps) || pe.Before(end) {
				t.Errorf("%s: %T %v-%v is not inside its parent %T %v-%v", path, c, start, end, n, ps, pe)
			}
			checkNames(t, path, src, lt, c)
			check(c)
			return false
		})
	}
	check(f)
	if nodes == 0 {
		t.Errorf("%s: no nodes", path)
	}
}

// checkNames checks that a name position holds the name.
func checkNames(t *testing.T, path, src string, lt lineTable, n ast.Node) {
	t.Helper()
	want := func(p ast.Pos, name string) {
		if p == (ast.Pos{}) {
			t.Errorf("%s: %T has no position for %q", path, n, name)
			return
		}
		off := lt.offset(p)
		if off+len(name) > len(src) || src[off:off+len(name)] != name {
			t.Errorf("%s: %T: expected %q at %v, found %q", path, n, name, p, src[off:min(off+len(name), len(src))])
		}
	}
	switch x := n.(type) {
	case *ast.Ident:
		want(x.Pos, x.Name)
		if textAt(src, lt, x.Pos, x.End) != x.Name {
			t.Errorf("%s: identifier %s at %v spans %q", path, x.Name, x.Pos, textAt(src, lt, x.Pos, x.End))
		}
	case *ast.FuncDecl:
		want(x.NamePos, x.Name)
		if x.Owner != "" {
			want(x.OwnerPos, x.Owner)
		}
	case *ast.TypeDecl:
		want(x.NamePos, x.Name)
		if x.Owner != "" {
			want(x.OwnerPos, x.Owner)
		}
	case *ast.Field:
		want(x.NamePos, x.Name)
	case *ast.TypeName:
		if len(x.PathPos) != len(x.Path) {
			t.Errorf("%s: type name %s has %d positions", path, strings.Join(x.Path, "."), len(x.PathPos))
			return
		}
		for i, part := range x.Path {
			want(x.PathPos[i], part)
		}
	case *ast.Import:
		want(x.PathPos, `"`)
		if textAt(src, lt, x.PathPos, x.PathEnd) != `"`+x.Path+`"` {
			t.Errorf("%s: import path spans %q", path, textAt(src, lt, x.PathPos, x.PathEnd))
		}
	case *ast.ImportName:
		want(x.Pos, x.Name)
		if x.Alias != x.Name {
			want(x.AliasPos, x.Alias)
		}
	case *ast.NameTarget:
		want(x.Pos, x.Name)
	case *ast.Param:
		want(x.Pos, x.Name)
	case *ast.TypeParam:
		want(x.Pos, x.Name)
	case *ast.FieldType:
		want(x.Pos, x.Name)
	case *ast.HtmlElement:
		want(x.NamePos, x.Name)
		if x.Close != (ast.Pos{}) {
			want(x.Close, "</"+x.Name)
		}
	case *ast.HtmlAttr:
		if !x.Spread {
			want(x.Pos, x.Name)
		}
	case *ast.Is:
		if x.Kind != "" {
			want(x.KindPos, x.Kind)
		}
	case *ast.Html:
		if x.Close != (ast.Pos{}) {
			want(x.Close, "</>")
		}
	}
}

// TestPath checks that the path to a position runs from the file down to the innermost node.
func TestPath(t *testing.T) {
	src := "type Point = { x: float64, y: float64 }\nfunc norm(p: Point): float64 { p.x * p.x + p.y * p.y }\n"
	f, err := Parse("t.ant", src, Program)
	if err != nil {
		t.Fatal(err)
	}
	// the x of the second p.x
	path := ast.Path(f, ast.Pos{Line: 2, Col: 40})
	var kinds []string
	for _, n := range path {
		kinds = append(kinds, strings.TrimPrefix(fmt.Sprintf("%T", n), "*ast."))
	}
	want := "File FuncDecl Block Binary Binary Field"
	if got := strings.Join(kinds, " "); got != want {
		t.Errorf("path = %s, want %s", got, want)
	}
	if fld, ok := path[len(path)-1].(*ast.Field); !ok || fld.Name != "x" || fld.NamePos != (ast.Pos{Line: 2, Col: 40}) {
		t.Errorf("innermost = %#v", path[len(path)-1])
	}
}
