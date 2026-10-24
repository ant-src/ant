package prelude

import (
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
