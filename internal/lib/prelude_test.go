package lib

import (
	"testing"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/internal/prelude"
	"github.com/ant-src/ant/value"
)

// TestPreludeMatchesLibrary checks that every function and constant the prelude declares has
// a run-time binding with the same parameter names, and that the library binds nothing the
// prelude does not declare.
func TestPreludeMatchesLibrary(t *testing.T) {
	f, err := prelude.Decls()
	if err != nil {
		t.Fatal(err)
	}
	table := Prelude()
	declared := map[string]bool{}
	member := func(owner, name string) (value.Value, bool) {
		ent, ok := table[owner]
		if !ok {
			return nil, false
		}
		switch x := ent.(type) {
		case *value.Record:
			return x.Field(name)
		case *value.Tag:
			return x.Func(name)
		}
		return nil, false
	}
	for _, d := range f.Decls {
		switch d := d.(type) {
		case *ast.FuncDecl:
			var v value.Value
			var ok bool
			if d.Owner == "" {
				v, ok = table[d.Name]
				declared[d.Name] = true
			} else {
				v, ok = member(d.Owner, d.Name)
				declared[d.Owner+"."+d.Name] = true
			}
			if !ok {
				t.Errorf("%s.%s is declared but not implemented", d.Owner, d.Name)
				continue
			}
			b, ok := v.(*interp.Builtin)
			if !ok {
				t.Errorf("%s.%s is not a builtin", d.Owner, d.Name)
				continue
			}
			if len(b.Params) != len(d.Params) {
				t.Errorf("%s.%s: %d parameters implemented, %d declared", d.Owner, d.Name, len(b.Params), len(d.Params))
				continue
			}
			for i, p := range d.Params {
				if b.Params[i] != p.Name {
					t.Errorf("%s.%s: parameter %d is %s, declared %s", d.Owner, d.Name, i, b.Params[i], p.Name)
				}
			}
			optional := 0
			for i := len(d.Params) - 1; i >= 0 && d.Params[i].Default != nil; i-- {
				optional++
			}
			if d.Owner == "" && d.Name == "println" {
				optional = 1 // the checker, not a default, lets println() omit its value
			}
			if b.Optional != optional {
				t.Errorf("%s.%s: %d optional parameters implemented, %d declared with defaults", d.Owner, d.Name, b.Optional, optional)
			}
		case *ast.VarDecl:
			name := d.Target.(*ast.NameTarget).Name
			declared[d.Owner+"."+name] = true
			if _, ok := member(d.Owner, name); !ok {
				t.Errorf("%s.%s is declared but not bound", d.Owner, name)
			}
		case *ast.TypeDecl:
			if d.Owner == "" {
				declared[d.Name] = true
			} else {
				declared[d.Owner+"."+d.Name] = true
			}
		}
	}
	for name, ent := range table {
		switch x := ent.(type) {
		case *value.Record:
			for _, field := range x.Tag.Fields {
				if !declared[name+"."+field] {
					t.Errorf("%s.%s is implemented but not declared", name, field)
				}
			}
		case *value.Tag:
			for fn := range x.Funcs() {
				if !declared[name+"."+fn] {
					t.Errorf("%s.%s is implemented but not declared", name, fn)
				}
			}
			if !declared[name] {
				t.Errorf("%s is bound but not declared", name)
			}
		default:
			if !declared[name] {
				t.Errorf("%s is bound but not declared", name)
			}
		}
	}
}
