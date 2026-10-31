package loader

import (
	"strings"
	"testing"
)

// TestCompanion covers the companion scope of section 10, Tests: a test file beside a library
// imports every top-level name of it and the names it imports; beside a script, or with nothing
// beside it, nothing.
func TestCompanion(t *testing.T) {
	sources := map[string]string{
		"stack.ant":        "import \"./lib/geometry.ant\" { Point }\ntype Stack = {}\nfunc secret(): int { 1 }",
		"lib/geometry.ant": "export type Point = {}",
		"stack_test.ant":   "test \"a\" { assert(secret() == 1) }",
	}
	prog, diags := Load("stack_test.ant", Config{Sources: sources})
	if len(diags) > 0 {
		t.Fatal(diags[0])
	}
	imps := prog.Root.File.Imports
	if len(imps) != 2 || !imps[0].Implicit || imps[0].Names[0].Name != "Point" || !imps[1].Companion || !imps[1].Implicit {
		t.Fatalf("imports: %+v", imps)
	}
	var names []string
	for _, n := range imps[1].Names {
		names = append(names, n.Name)
	}
	if got := strings.Join(names, " "); got != "Stack secret" {
		t.Errorf("companion names %q", got)
	}
	if prog.Modules[imps[1].Module].Path != "stack.ant" || prog.Root.Path != "stack_test.ant" {
		t.Errorf("modules: %v", prog.Modules)
	}

	// beside a script, a test file is black box
	prog, diags = Load("tool_test.ant", Config{Sources: map[string]string{
		"tool.ant":      "println(1)",
		"tool_test.ant": "test \"a\" {}",
	}})
	if len(diags) > 0 || len(prog.Root.File.Imports) != 0 {
		t.Fatalf("beside a script: %v %v", diags, prog)
	}
	// and with nothing beside it
	if prog, diags = Load("api_test.ant", Config{Sources: map[string]string{"api_test.ant": "test \"a\" {}"}}); len(diags) > 0 || len(prog.Root.File.Imports) != 0 {
		t.Fatalf("alone: %v", diags)
	}
}

func TestTestFileErrors(t *testing.T) {
	cases := []struct {
		name, root string
		sources    map[string]string
		want       string
	}{
		{"import of a test file", "main.ant", map[string]string{
			"main.ant":   "import \"./h_test.ant\" { f }",
			"h_test.ant": "export func f(): int { 1 }",
		}, "it is a test file, which only another test file can import"},
		{"shadowing a companion's name", "a_test.ant", map[string]string{
			"a.ant":      "func f(): int { 1 }",
			"a_test.ant": "func f(): int { 2 }",
		}, "f is already bound by an import"},
		{"importing a companion's import again", "a_test.ant", map[string]string{
			"a.ant":      "import \"./b.ant\" { g }",
			"b.ant":      "export func g(): int { 1 }",
			"a_test.ant": "import \"./b.ant\" { g }",
		}, "g is already bound by an import"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, diags := Load(c.root, Config{Sources: c.sources})
			if len(diags) == 0 || !strings.Contains(diags[0].Msg, c.want) {
				t.Fatalf("got %v, want %q", diags, c.want)
			}
		})
	}
	// a test file importing another is fine
	if _, diags := Load("a_test.ant", Config{Sources: map[string]string{
		"a_test.ant": "import \"./h_test.ant\" { f }\ntest \"a\" { assert(f() == 1) }",
		"h_test.ant": "export func f(): int { 1 }",
	}}); len(diags) > 0 {
		t.Fatal(diags[0])
	}
}
