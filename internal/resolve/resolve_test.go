package resolve

import (
	"strings"
	"testing"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/parser"
)

func parse(t *testing.T, src string) *ast.File {
	t.Helper()
	f, err := parser.Parse("test", src, parser.Program)
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func TestScopeErrors(t *testing.T) {
	cases := []struct{ name, src, want string }{
		{"redeclare local", "func f() { var x = 1; var x = 2 }", "x is already declared"},
		{"redeclare param", "func f(a: int) { var a = 2 }", "a is already declared"},
		{"redeclare top", "var x = 1\nvar x = 2", "x is already declared"},
		{"shadow prelude", "func f() { var print = 1 }", "print is a prelude name"},
		{"shadow prelude param", "func f(int: int) { int }", "int is a prelude name"},
		{"shadow prelude loop", "func f() { for Array in xs { Array } }", "Array is a prelude name"},
		{"shadow prelude type param", "func f[int](x: int): int { x }", "int is a prelude name"},
		{"shadow prelude type decl param", "type P[string] = { a: string }", "string is a prelude name"},
		{"type param twice", "func f[T, T](x: T): T { x }", "T is already declared"},
		{"shadow import", "import \"./a.ant\" { thing }\nvar thing = 1", "thing is already bound by an import"},
		{"two imports", "import \"./a.ant\" { thing }\nimport \"./b.ant\" { thing }", "thing is already bound by an import"},
		{"import prelude name", "import \"./a.ant\" { print }", "print is a prelude name"},
		{"func on unknown type", "func Point.norm(p: Point): float64 { 1.0 }", "Point is not a type declared in this module"},
		{"func on type twice", "type P = {}\nfunc P.a() {}\nfunc P.a() {}", "P.a is already declared"},
		{"func shares field name", "type P = { a: int }\nfunc P.a() {}", "shares its name with a field"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			errs := Resolve(parse(t, c.src), Options{})
			if len(errs) == 0 {
				t.Fatalf("resolved, want an error containing %q", c.want)
			}
			if !strings.Contains(errs[0].Error(), c.want) {
				t.Fatalf("error %q, want one containing %q", errs[0].Error(), c.want)
			}
		})
	}
	ok := []string{
		"func f() { var _ = 1; var _ = 2 }",
		"func f() { var x = 1; if true { var x = 2 } }",
		"var x = 1\nfunc f() { var x = 2 }",
		"type P = { a: int }\nfunc P.b() {}\nfunc P.c(p: P): int { p.a }",
		"func f(x: int) { { x } }",
	}
	for _, src := range ok {
		if errs := Resolve(parse(t, src), Options{}); len(errs) > 0 {
			t.Errorf("%q: unexpected %v", src, errs[0])
		}
	}
}

// find returns the identifiers named name, in source order.
func find(n ast.Node, name string) []*ast.Ident {
	var out []*ast.Ident
	walk(n, func(x ast.Node) {
		if id, ok := x.(*ast.Ident); ok && id.Name == name {
			out = append(out, id)
		}
	})
	return out
}

func walk(n ast.Node, visit func(ast.Node)) {
	if n == nil {
		return
	}
	visit(n)
	switch x := n.(type) {
	case *ast.File:
		for _, d := range x.Decls {
			walk(d, visit)
		}
	case *ast.VarDecl:
		walk(x.Init, visit)
	case *ast.FuncDecl:
		for _, p := range x.Params {
			walk(p.Default, visit)
		}
		walk(x.Body, visit)
	case *ast.Block:
		for _, s := range x.Stmts {
			walk(s, visit)
		}
	case *ast.Lambda:
		walk(x.Body, visit)
	case *ast.FuncLit:
		walk(x.Body, visit)
	case *ast.Call:
		walk(x.Fn, visit)
		for _, a := range x.Args {
			walk(a.Value, visit)
		}
	case *ast.Binary:
		walk(x.X, visit)
		walk(x.Y, visit)
	case *ast.Assign:
		walk(x.Target, visit)
		walk(x.Value, visit)
	case *ast.If:
		walk(x.Cond, visit)
		walk(x.Then, visit)
		walk(x.Else, visit)
	case *ast.For:
		walk(x.Iter, visit)
		walk(x.Body, visit)
	case *ast.Go:
		walk(x.Body, visit)
	case *ast.Field:
		walk(x.X, visit)
	case *ast.Index:
		walk(x.X, visit)
		walk(x.Index, visit)
	case *ast.Return:
		walk(x.Value, visit)
	case *ast.RecordLit:
		walk(x.Head, visit)
		for _, e := range x.Entries {
			walk(e.Value, visit)
			walk(e.Spread, visit)
		}
	case *ast.MapLit:
		for _, e := range x.Entries {
			walk(e.Key, visit)
			walk(e.Value, visit)
		}
	case *ast.ArrayLit:
		for _, e := range x.Elems {
			walk(e.Value, visit)
		}
	case *ast.Template:
		for _, p := range x.Parts {
			walk(p.Expr, visit)
		}
	case *ast.Unary:
		walk(x.X, visit)
	case *ast.Is:
		walk(x.X, visit)
		walk(x.Path, visit)
	case *ast.While:
		walk(x.Cond, visit)
		walk(x.Body, visit)
	case *ast.Defer:
		walk(x.Body, visit)
	}
}

func TestRefs(t *testing.T) {
	src := `
var top = 1
func f(a: int): int {
  var b = a + top + unknown
  if true {
    var c = b
    c + _
  }
  var g = { x + a }
  go { b = 2 }
  for i in xs { var h = { i } }
  g(1)
}
`
	f := parse(t, src)
	if errs := Resolve(f, Options{}); len(errs) > 0 {
		t.Fatal(errs[0])
	}
	fn := f.Decls[1].(*ast.FuncDecl)
	// a is captured by the lambda: shared; b by the go block: shared; g is plain.
	if fn.Scope.Shared != 2 || fn.Scope.Plain != 1 {
		t.Errorf("function scope: plain %d shared %d, want 1 and 2", fn.Scope.Plain, fn.Scope.Shared)
	}
	a := find(fn, "a")
	if r := a[0].Ref; r.Kind != ast.RefLocal || r.Depth != 0 || !r.Shared {
		t.Errorf("a in body: %+v", r)
	}
	if r := a[1].Ref; r.Kind != ast.RefLocal || r.Depth != 1 || !r.Shared {
		t.Errorf("a in lambda: %+v", r)
	}
	if r := find(fn, "top")[0].Ref; r.Kind != ast.RefModule || r.Index != 0 {
		t.Errorf("top: %+v", r)
	}
	if r := find(fn, "unknown")[0].Ref; r.Kind != ast.RefUnbound {
		t.Errorf("unknown: %+v", r)
	}
	if r := find(fn, "_")[0].Ref; r.Kind != ast.RefDiscard {
		t.Errorf("_: %+v", r)
	}
	// c lives in the if block's own environment, one hop from the function's.
	ifBlock := fn.Body.Stmts[1].(*ast.If).Then
	if ifBlock.Scope == nil || ifBlock.Scope.Plain != 1 {
		t.Errorf("if block scope: %+v", ifBlock.Scope)
	}
	b := find(fn, "b")
	if r := b[0].Ref; r.Kind != ast.RefLocal || r.Depth != 1 || !r.Shared {
		t.Errorf("b in if block: %+v", r)
	}
	if r := b[1].Ref; r.Kind != ast.RefLocal || r.Depth != 1 || !r.Shared {
		t.Errorf("b in go block: %+v", r)
	}
	// The lambda's x is its slot 0.
	lam := fn.Body.Stmts[2].(*ast.VarDecl).Init.(*ast.Lambda)
	if r := find(lam, "x")[0].Ref; r.Kind != ast.RefLocal || r.Depth != 0 || r.Index != lam.Scope.X.Index {
		t.Errorf("x: %+v", r)
	}
	// The loop variable i is captured by the inner lambda, so the iteration scope has a shared slot.
	loop := fn.Body.Stmts[4].(*ast.For)
	if loop.Scope == nil || loop.Scope.Shared != 1 || loop.Scope.Plain != 1 {
		t.Errorf("loop scope: %+v", loop.Scope)
	}
	if r := find(fn, "g")[0].Ref; r.Kind != ast.RefLocal || r.Depth != 0 || r.Shared {
		t.Errorf("g: %+v", r)
	}
	if fields := f.Scope.Fields; len(fields) != 2 || fields[0] != "top" || fields[1] != "f" {
		t.Errorf("module fields: %v", fields)
	}
}

func TestImportsAndTypeFuncs(t *testing.T) {
	src := `
import "./g.ant" Geometry
import "./g.ant" { distance: dist, Point }
type Counter = { n: int }
func Counter.new(): Counter { Counter{ n: 0 } }
func main() { dist(Geometry.Point{ x: 1.0 }, Point{ x: 2.0 }); Counter.new() }
`
	f := parse(t, src)
	if errs := Resolve(f, Options{}); len(errs) > 0 {
		t.Fatal(errs[0])
	}
	main := f.Decls[2].(*ast.FuncDecl)
	if r := find(main, "dist")[0].Ref; r.Kind != ast.RefImport || r.Index != 1 || r.Name != "distance" {
		t.Errorf("dist: %+v", r)
	}
	if r := find(main, "Geometry")[0].Ref; r.Kind != ast.RefImport || r.Index != 0 || r.Name != "" {
		t.Errorf("Geometry: %+v", r)
	}
	if r := find(main, "Counter")[0].Ref; r.Kind != ast.RefModule {
		t.Errorf("Counter: %+v", r)
	}
	if fields := f.Scope.Fields; len(fields) != 2 || fields[0] != "Counter" || fields[1] != "main" {
		t.Errorf("module fields: %v", fields)
	}
}
