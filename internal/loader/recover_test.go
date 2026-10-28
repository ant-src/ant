package loader

import (
	"testing"
	"testing/fstest"

	"github.com/ant-src/ant/parser"
)

// TestRecover checks that a recovering load keeps every module it can read, parsed as far as
// it goes, and marks the imports it cannot follow.
func TestRecover(t *testing.T) {
	fsys := fstest.MapFS{
		"main.ant": {Data: []byte("import \"./lib.ant\" { f }\nimport \"./missing.ant\" M\nvar x = f(\n")},
		"lib.ant":  {Data: []byte("export func f(): int { 1 }\nvar y = \n")},
	}
	prog, diags := Load("main.ant", Config{FS: fsys, Mode: parser.Program, Recover: true})
	if prog == nil {
		t.Fatalf("no program; diagnostics %v", diags)
	}
	if len(prog.Modules) != 2 || prog.Root.File.Path != "main.ant" {
		t.Fatalf("modules %d, root %q", len(prog.Modules), prog.Root.File.Path)
	}
	imports := prog.Root.File.Imports
	if imports[0].Module != 0 || imports[1].Module != -1 {
		t.Errorf("import modules %d and %d, want 0 and -1", imports[0].Module, imports[1].Module)
	}
	if prog.Root.Src == "" || prog.Root.Layout == nil || len(prog.Root.Layout.Tokens) == 0 {
		t.Error("the root keeps no source or layout")
	}
	// The root's own errors come first, then each imported module's as it is reached, then the
	// import that leads nowhere, found when the imports are followed.
	want := []string{
		"main.ant:4:1: expected ')', got end of file",
		"lib.ant:3:1: unexpected newline",
		`main.ant:2:1: cannot find module "./missing.ant"`,
	}
	if len(diags) != len(want) {
		t.Fatalf("diagnostics %v, want %v", diags, want)
	}
	for i, d := range diags {
		if d.String() != want[i] {
			t.Errorf("diagnostic %d = %q, want %q", i, d.String(), want[i])
		}
	}
	// A strict load of the same program reports only the first syntax error of each module and
	// gives no program.
	strict, sdiags := Load("main.ant", Config{FS: fsys, Mode: parser.Program})
	if strict != nil || len(sdiags) == 0 {
		t.Errorf("strict load: program %v, diagnostics %v", strict, sdiags)
	}
}
