package prelude

import (
	"strings"
	"testing"

	"github.com/ant-src/ant/ast"
)

func TestPreludeParses(t *testing.T) {
	f, err := Decls()
	if err != nil {
		t.Fatal(err)
	}
	funcs, types, consts := 0, 0, 0
	for _, d := range f.Decls {
		switch d := d.(type) {
		case *ast.FuncDecl:
			funcs++
			if d.Body != nil {
				t.Errorf("%s has a body", d.Name)
			}
		case *ast.TypeDecl:
			types++
		case *ast.VarDecl:
			consts++
			if d.Owner == "" || d.Init != nil || d.Type == nil || !d.Let {
				t.Errorf("constant %v must be let Owner.NAME: T", d.Target)
			}
		}
	}
	if funcs < 150 || types < 15 || consts != 10 {
		t.Errorf("unexpected counts: %d funcs, %d types, %d constants", funcs, types, consts)
	}
}

// TestPreludeDocs requires every declaration of the prelude to have a doc comment whose first
// sentence starts with its name as called (spec doc.md section 2).
func TestPreludeDocs(t *testing.T) {
	f, err := Decls()
	if err != nil {
		t.Fatal(err)
	}
	for _, d := range f.Decls {
		var owner, name, doc string
		switch d := d.(type) {
		case *ast.FuncDecl:
			owner, name, doc = d.Owner, d.Name, d.Doc
		case *ast.TypeDecl:
			owner, name, doc = d.Owner, d.Name, d.Doc
		case *ast.VarDecl:
			owner, name, doc = d.Owner, d.Target.(*ast.NameTarget).Name, d.Doc
		}
		if owner != "" {
			name = owner + "." + name
		}
		if !strings.HasPrefix(doc, name+" ") {
			t.Errorf("%s: the doc comment must start with %q, got %q", name, name+" ", doc)
		}
	}
}

// TestNamespaceDocs requires every namespace without a type of its own name, such as Array or
// Time, to have a doc that starts with its name, and no namespace to have two.
func TestNamespaceDocs(t *testing.T) {
	f, err := Decls()
	if err != nil {
		t.Fatal(err)
	}
	types, owners := map[string]bool{}, map[string]bool{}
	for _, d := range f.Decls {
		switch d := d.(type) {
		case *ast.TypeDecl:
			if d.Owner == "" {
				types[d.Name] = true
			} else {
				owners[d.Owner] = true
			}
		case *ast.FuncDecl:
			owners[d.Owner] = true
		case *ast.VarDecl:
			owners[d.Owner] = true
		}
	}
	for ns := range owners {
		_, doc := NamespaceDoc(ns)
		switch {
		case ns == "":
		case types[ns] && doc != "":
			t.Errorf("%s has a type, whose doc is the namespace's; its section label must not have one", ns)
		case !types[ns] && !strings.HasPrefix(doc, ns+" "):
			t.Errorf("namespace %s: the doc under its section label must start with %q, got %q", ns, ns+" ", doc)
		}
	}
}

// TestNoFieldShadowsFunction requires that no record type of the prelude has a field named like
// a function its namespace offers on it, since x.f() reads the field where the author meant the
// function (section 5, Calls through .). A namespace offers a function on a type when the
// function's first parameter is that type, a type parameter, or a union naming the type.
func TestNoFieldShadowsFunction(t *testing.T) {
	f, err := Decls()
	if err != nil {
		t.Fatal(err)
	}
	// the record types by full name, with their namespace: Owner for Owner.T, else the type's own name
	records := map[string]*ast.TypeDecl{}
	for _, d := range f.Decls {
		if td, ok := d.(*ast.TypeDecl); ok {
			if _, isRecord := td.Body.(*ast.RecordType); isRecord {
				name := td.Name
				if td.Owner != "" {
					name = td.Owner + "." + td.Name
				}
				records[name] = td
			}
		}
	}
	for _, d := range f.Decls {
		fd, ok := d.(*ast.FuncDecl)
		if !ok || fd.Owner == "" || len(fd.Params) == 0 {
			continue
		}
		tparams := map[string]bool{}
		for _, tp := range fd.TypeParams {
			tparams[tp.Name] = true
		}
		for name, td := range records {
			ns := td.Owner
			if ns == "" {
				ns = td.Name
			}
			if ns == "Ok" || ns == "Err" {
				ns = "Result"
			}
			if ns != fd.Owner || !firstParamTakes(fd.Params[0].Type, name, tparams) {
				continue
			}
			for _, fld := range td.Body.(*ast.RecordType).Fields {
				if fld.Name == fd.Name {
					t.Errorf("%s.%s is a field of %s and a function %s.%s offers on it", name, fld.Name, name, fd.Owner, fd.Name)
				}
			}
		}
	}
}

// firstParamTakes reports whether a parameter type takes a value of the named type: it names
// it, is a type parameter, or is a union with such a member.
func firstParamTakes(t ast.Type, name string, tparams map[string]bool) bool {
	switch x := t.(type) {
	case *ast.TypeName:
		path := strings.Join(x.Path, ".")
		return path == name || len(x.Path) == 1 && tparams[x.Path[0]]
	case *ast.UnionType:
		for _, m := range x.Members {
			if firstParamTakes(m, name, tparams) {
				return true
			}
		}
	}
	return false
}
