package lib

import (
	"testing"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/internal/prelude"
	"github.com/ant-src/ant/value"
)

// TestCmdDefaults keeps the defaults the library fills in a Cmd literal and in Cmd.new in step
// with the ones the prelude declares, which the runtime never evaluates.
func TestCmdDefaults(t *testing.T) {
	f, err := prelude.Decls()
	if err != nil {
		t.Fatal(err)
	}
	const want = "type Cmd = { program: string, args: Array[string] = [], dir: string | null = null, env: Map[string, string] | null = null }"
	found := false
	for _, d := range f.Decls {
		if td, ok := d.(*ast.TypeDecl); ok && td.Owner == "" && td.Name == "Cmd" {
			found = true
			if td.Sig != want {
				t.Errorf("the prelude declares\n%s\nbut the library fills the defaults of\n%s", td.Sig, want)
			}
		}
	}
	if !found {
		t.Fatal("no type Cmd in the prelude")
	}
	native := CmdTag.Defaults.(interp.NativeDefaults)
	if len(native) != 4 || native[0] != nil {
		t.Fatalf("defaults %v", native)
	}
	if a, ok := native[1]().(*value.Array); !ok || a.Len() != 0 || native[1]() == native[1]() {
		t.Error("args should default to a fresh empty array")
	}
	if native[2]() != value.Null || native[3]() != value.Null {
		t.Error("dir and env should default to null")
	}
}
