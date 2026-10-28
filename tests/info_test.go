package tests

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/check"
	"github.com/ant-src/ant/internal/loader"
	"github.com/ant-src/ant/internal/resolve"
	"github.com/ant-src/ant/internal/types"
	"github.com/ant-src/ant/parser"
)

const infoMain = `import "./lib.ant" { Point, distance: dist }
import "./lib.ant" Geo

// Shape is a named area.
type Shape = { name: string, area: float64 }

// describe names a shape.
func describe(s: Shape, n: int = 1): string {
  var label = s.name
  var p = Point{ x: 1.0, y: 2.0 }
  var q: Geo.Point = p
  var d = dist(p, Geo.origin())
  var xs = [1, 2, 3] | Array.map({ x * n })
  var m = Array.map(xs, { x })
  if s.area > 0.0 { label } else { "none" }
}

func first(v: int | null): int {
  if v != null { v } else { 0 }
}
`

const infoLib = `export type Point = { x: float64, y: float64 }

// distance is the straight-line distance from a to b.
export func distance(a: Point, b: Point): float64 { 0.0 }

export func origin(): Point { Point{ x: 0.0, y: 0.0 } }
`

// TestInfo checks what the resolver and the checker record for an editor.
func TestInfo(t *testing.T) {
	fsys := fstest.MapFS{"main.ant": {Data: []byte(infoMain)}, "lib.ant": {Data: []byte(infoLib)}}
	prog, diags := loader.Load("main.ant", loader.Config{FS: fsys, Mode: parser.Program, Recover: true})
	if len(diags) > 0 {
		t.Fatalf("load: %v", diags)
	}
	info := check.NewInfo()
	if cdiags := check.Program(prog, check.Config{Info: info}); len(cdiags) > 0 {
		t.Fatalf("check: %v", cdiags)
	}
	f := prog.Root.File
	rinfo := prog.Root.Resolve

	// an imported function under its alias
	dist := ident(f, "dist", 1)
	if o := info.Objects[dist]; o == nil || o.Kind != check.Func || o.Name != "distance" || o.Path != "lib.ant" || !o.Exported {
		t.Errorf("dist refers to %+v", o)
	} else if _, ok := o.Decl.(*ast.FuncDecl); !ok || !strings.HasPrefix(o.Doc(), "distance is") || o.Sig() != "func distance(a: Point, b: Point): float64" {
		t.Errorf("dist's declaration: %T, doc %q, sig %q", o.Decl, o.Doc(), o.Sig())
	}
	if d, ok := rinfo.Defs[dist].(*ast.ImportName); !ok || d.Name != "distance" {
		t.Errorf("the resolver says dist is declared by %T", rinfo.Defs[dist])
	}

	// a member of a module namespace, and a record field
	origin := field(f, "origin")
	if o := info.Members[origin]; o == nil || o.Kind != check.Func || o.Name != "origin" || o.Path != "lib.ant" {
		t.Errorf("Geo.origin refers to %+v", o)
	}
	if o := info.Objects[ident(f, "Geo", 1)]; o == nil || o.Kind != check.Module || o.Path != "lib.ant" {
		t.Errorf("Geo refers to %+v", o)
	}
	name := field(f, "name")
	if o := info.Members[name]; o == nil || o.Kind != check.Field || o.Owner != "Shape" || types.String(o.Type()) != "string" {
		t.Errorf("s.name refers to %+v", o)
	} else if ft, ok := o.Decl.(*ast.FieldType); !ok || ft.Name != "name" {
		t.Errorf("s.name's declaration: %T", o.Decl)
	}

	// types at uses, narrowed
	if ty := info.Types[ident(f, "label", 1)]; types.String(ty) != "string" {
		t.Errorf("label at its use: %v", ty)
	}
	v := ident(f, "v", 2) // in the then branch
	if ty := info.Types[v]; types.String(ty) != "int" {
		t.Errorf("v in the then branch: %v, want int", ty)
	}
	if o := info.Objects[v]; o == nil || o.Kind != check.Param || types.String(o.Type()) != "int | null" {
		t.Errorf("v's declaration: %+v", o)
	}

	// calls
	var pipe, plain, distCall *ast.Call
	ast.Inspect(f, func(n ast.Node) bool {
		if c, ok := n.(*ast.Call); ok {
			switch fn := c.Fn.(type) {
			case *ast.Field:
				if fn.Name == "map" && c.Pipe {
					pipe = c
				} else if fn.Name == "map" {
					plain = c
				}
			case *ast.Ident:
				if fn.Name == "dist" {
					distCall = c
				}
			}
		}
		return true
	})
	if ci := info.Calls[distCall]; ci == nil || ci.Callee == nil || ci.Callee.Name != "distance" || strings.Join(ci.Names, ",") != "a,b" || !equalInts(ci.Slots, 0, 1) {
		t.Errorf("dist(...): %+v", ci)
	}
	if ci := info.Calls[pipe]; ci == nil || !ci.Pipe || ci.Callee == nil || ci.Callee.Owner != "Array" || ci.Callee.Path != "prelude.ant" || !equalInts(ci.Slots, 0, 1) {
		t.Errorf("xs | Array.map(...): %+v", ci)
	}
	if ci := info.Calls[plain]; ci == nil || ci.Pipe || len(ci.Defaults) != 2 {
		t.Errorf("Array.map(xs, ...): %+v", ci)
	}

	// lambdas and record entries
	var lambdas []*ast.Lambda
	var entries []*ast.RecordEntry
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.Lambda:
			lambdas = append(lambdas, x)
		case *ast.RecordEntry:
			entries = append(entries, x)
		}
		return true
	})
	if len(lambdas) != 2 || types.String(info.Lambdas[lambdas[0]]) != "int" {
		t.Errorf("lambda x: %v", info.Lambdas)
	}
	if o := info.Entries[entries[0]]; o == nil || o.Name != "x" || o.Owner != "Point" || o.Path != "lib.ant" {
		t.Errorf("entry x: %+v", o)
	}

	// type names
	var typeNames []*ast.TypeName
	ast.Inspect(f, func(n ast.Node) bool {
		if tn, ok := n.(*ast.TypeName); ok {
			typeNames = append(typeNames, tn)
		}
		return true
	})
	for _, tn := range typeNames {
		objs := info.TypeNames[tn]
		switch strings.Join(tn.Path, ".") {
		case "Geo.Point":
			if len(objs) != 2 || objs[0].Kind != check.Module || objs[1].Kind != check.TypeKind || objs[1].Name != "Point" {
				t.Errorf("Geo.Point refers to %+v", objs)
			}
		case "Shape":
			if len(objs) != 1 || objs[0].Kind != check.TypeKind || objs[0].TypeDecl() == nil || len(objs[0].TypeDecl().Fields) != 2 {
				t.Errorf("Shape refers to %+v", objs)
			}
		case "int", "string", "float64":
			if len(objs) != 1 || objs[0].Kind != check.TypeKind || objs[0].Decl != nil {
				t.Errorf("%s refers to %+v", tn.Path[0], objs)
			}
		}
	}

	// declarations and listings
	describe := f.Decls[1].(*ast.FuncDecl)
	if o := info.Decls[describe]; o == nil || o.Kind != check.Func || types.String(o.Type()) != "func(Shape, int = ...): string" {
		t.Errorf("describe: %+v, type %v", o, o.Type())
	} else if !strings.HasPrefix(o.Doc(), "describe names") || len(o.Defaults()) != 2 || !o.Defaults()[1] {
		t.Errorf("describe's doc %q, defaults %v", o.Doc(), o.Defaults())
	}
	var array *check.Object
	for _, o := range info.Prelude {
		if o.Name == "Array" {
			array = o
		}
	}
	if array == nil || array.Kind != check.Namespace {
		t.Fatalf("Array in the prelude: %+v", array)
	}
	found := false
	for _, m := range array.Members() {
		if m.Name == "map" && m.Kind == check.Func && strings.HasPrefix(m.Doc(), "Array.map returns") {
			found = true
		}
	}
	if !found {
		t.Error("Array.map is not among Array's members with its doc")
	}
	if m := info.Modules["lib.ant"]; m == nil || len(m.Exports) != 3 || m.Exports[0].Name != "Point" {
		t.Errorf("lib.ant: %+v", m)
	}
	shape := info.Decls[f.Decls[0]]
	var fields []string
	for _, m := range shape.Members() {
		fields = append(fields, m.Kind.String()+" "+m.Name)
	}
	if strings.Join(fields, ", ") != "field area, field name" {
		t.Errorf("Shape's members: %v", fields)
	}

	// the resolver's scopes
	body := rinfo.Scopes[describe.Body]
	if body == nil {
		t.Fatal("no scope for describe's body")
	}
	var names []string
	for _, b := range body.Names {
		names = append(names, b.Name)
	}
	if strings.Join(names, " ") != "s n label p q d xs m" || body.Parent != rinfo.Scopes[f] {
		t.Errorf("describe's scope: %v, parent %v", names, body.Parent)
	}
	label := ident(f, "label", 1)
	if d, ok := rinfo.Defs[label].(*ast.NameTarget); !ok || d.Name != "label" || len(rinfo.Uses[d]) != 1 {
		t.Errorf("label is declared by %T with %d uses", rinfo.Defs[label], len(rinfo.Uses[rinfo.Defs[label]]))
	}
	if _, ok := rinfo.Defs[ident(f, "x", 1)].(*ast.Lambda); !ok {
		t.Errorf("x is declared by %T", rinfo.Defs[ident(f, "x", 1)])
	}
}

// ident finds the nth identifier named name in source order.
func ident(f *ast.File, name string, n int) *ast.Ident {
	var found *ast.Ident
	ast.Inspect(f, func(node ast.Node) bool {
		if id, ok := node.(*ast.Ident); ok && id.Name == name && found == nil {
			n--
			if n == 0 {
				found = id
			}
		}
		return found == nil
	})
	return found
}

// field finds the first field access named name.
func field(f *ast.File, name string) *ast.Field {
	var found *ast.Field
	ast.Inspect(f, func(node ast.Node) bool {
		if fl, ok := node.(*ast.Field); ok && fl.Name == name && found == nil {
			found = fl
		}
		return found == nil
	})
	return found
}

func equalInts(got []int, want ...int) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}
	return true
}

var _ = resolve.NewInfo
