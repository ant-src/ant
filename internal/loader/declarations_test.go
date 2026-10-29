package loader

import (
	"strings"
	"testing"
	"testing/fstest"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/parser"
)

// TestDeclarationFile: a root whose name ends in .d.ant is a declaration file (section 10,
// Declaration files), parsed as one whatever mode was asked for and not resolved, and an import
// of one is an error.
func TestDeclarationFile(t *testing.T) {
	fsys := fstest.MapFS{
		"db.d.ant": {Data: []byte("type Db.Row = Map[string, string]\nfunc Db.query(sql: string): Array[Db.Row]\n")},
		"main.ant": {Data: []byte("import \"./db.d.ant\" Db\nprintln(1)\n")},
	}
	prog, diags := Load("db.d.ant", Config{FS: fsys, Mode: parser.Program})
	if prog == nil || len(diags) != 0 {
		t.Fatalf("loading a declaration file: %v", diags)
	}
	f := prog.Root.File
	if !f.Declarations || len(f.Decls) != 2 || prog.Root.Resolve != nil {
		t.Errorf("declaration file: Declarations %v, %d decls", f.Declarations, len(f.Decls))
	}
	if td, ok := f.Decls[0].(*ast.TypeDecl); !ok || td.Owner != "Db" || td.Name != "Row" {
		t.Errorf("first declaration: %+v", f.Decls[0])
	}
	if fd, ok := f.Decls[1].(*ast.FuncDecl); !ok || fd.Owner != "Db" || fd.Body != nil {
		t.Errorf("second declaration: %+v", f.Decls[1])
	}
	prog, diags = Load("main.ant", Config{FS: fsys, Mode: parser.Program})
	if prog != nil || len(diags) != 1 || !strings.HasPrefix(diags[0].Msg, `cannot import "./db.d.ant": a declaration file`) || diags[0].Pos.Line != 1 {
		t.Errorf("importing a declaration file: %v", diags)
	}
	prog, diags = Load("main.ant", Config{FS: fsys, Mode: parser.Program, Recover: true})
	if prog == nil || len(diags) != 1 || prog.Root.File.Imports[0].Module != -1 {
		t.Errorf("importing a declaration file with Recover: %v", diags)
	}
	// The import is refused by its name, before the file is looked for.
	fsys["main.ant"] = &fstest.MapFile{Data: []byte("import \"./missing.d.ant\" Db\nprintln(1)\n")}
	prog, diags = Load("main.ant", Config{FS: fsys, Mode: parser.Program})
	if prog != nil || len(diags) != 1 || !strings.HasPrefix(diags[0].Msg, `cannot import "./missing.d.ant": a declaration file`) {
		t.Errorf("importing a missing declaration file: %v", diags)
	}
}
