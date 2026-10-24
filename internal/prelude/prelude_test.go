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
			if d.Owner == "" || d.Init != nil || d.Type == nil {
				t.Errorf("constant %v must be var Owner.NAME: T", d.Target)
			}
		}
	}
	if funcs < 150 || types < 15 || consts != 9 {
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
