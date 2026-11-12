package lib

import (
	"strings"
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
			if d.Owner == "" && d.Name == "assert" {
				continue // not a value: the evaluator runs each call of it (section 8, assert)
			}
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
			if d.Owner == "" && (d.Name == "println" || d.Name == "eprintln") {
				optional = 1 // the checker, not a default, lets println() and eprintln() omit their value
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

// TestPreludeTypesMatchLibrary checks that every type the prelude declares is bound at run
// time by a value of the matching shape, so that a type expression reaches a tag whose kept
// body is the prelude's declaration (section 4, Types as values): a record type by a record tag
// with the declared fields, a primitive type by a constructor, an alias by an alias tag, and
// an opaque type by a tag that `is` cannot test. Bytes and Html are kinds, not tags.
func TestPreludeTypesMatchLibrary(t *testing.T) {
	f, err := prelude.Decls()
	if err != nil {
		t.Fatal(err)
	}
	table := Prelude()
	lookup := func(owner, name string) (value.Value, bool) {
		if owner == "" {
			v, ok := table[name]
			return v, ok
		}
		switch x := table[owner].(type) {
		case *value.Record:
			return x.Field(name)
		case *value.Tag:
			return x.Func(name)
		}
		return nil, false
	}
	for _, d := range f.Decls {
		td, ok := d.(*ast.TypeDecl)
		if !ok {
			continue
		}
		name := td.Name
		if td.Owner != "" {
			name = td.Owner + "." + td.Name
		}
		if name == "Bytes" || name == "Html" {
			continue
		}
		v, ok := lookup(td.Owner, td.Name)
		if !ok {
			t.Errorf("%s is declared but not bound", name)
			continue
		}
		var tag *value.Tag
		switch x := v.(type) {
		case *value.Tag:
			tag = x
		case *value.Constructor:
			tag = x.Tag
		default:
			t.Errorf("%s is bound to %s, not a type", name, value.Describe(v))
			continue
		}
		if tag.Name != name {
			t.Errorf("%s is bound to a tag named %s", name, tag.Name)
		}
		switch body := td.Body.(type) {
		case nil:
			if tag.Kind != value.SyncTag && tag.Kind != value.TimeTag && tag.Kind != value.RegexTag {
				t.Errorf("%s is opaque but its tag is not", name)
			}
		case *ast.RecordType:
			if tag.Kind != value.RecordTag {
				t.Errorf("%s is a record type but its tag is not a record tag", name)
				continue
			}
			var fields []string
			for _, fl := range body.Fields {
				fields = append(fields, fl.Name)
			}
			if strings.Join(fields, ",") != strings.Join(tag.Fields, ",") {
				t.Errorf("%s declares the fields %v, the tag has %v", name, fields, tag.Fields)
				continue
			}
			for i, fl := range body.Fields {
				if (fl.Default != nil) != interp.HasDefault(tag, i) {
					t.Errorf("%s.%s: default declared %v, implemented %v", name, fl.Name, fl.Default != nil, interp.HasDefault(tag, i))
				}
			}
		case *ast.TypeName:
			if k, ok := value.KindByName(body.Path[0]); ok && len(body.Path) == 1 && len(body.Args) == 0 && k.IsPrimitive() {
				if tag.Kind != value.PrimitiveTag || tag.Base != k {
					t.Errorf("%s is a named primitive type over %s but its tag is not", name, body.Path[0])
				}
			} else if tag.Kind != value.AliasTag {
				t.Errorf("%s is an alias but its tag is not an alias tag", name)
			}
		default:
			if tag.Kind != value.AliasTag {
				t.Errorf("%s is an alias but its tag is not an alias tag", name)
			}
		}
	}
}
