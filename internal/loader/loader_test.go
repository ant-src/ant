package loader

import (
	"strings"
	"testing"
)

func load(t *testing.T, sources map[string]string) (*Program, []Diagnostic) {
	t.Helper()
	return Load("main.ant", Config{Sources: sources, RequireMain: true})
}

func TestOrderAndImports(t *testing.T) {
	prog, diags := load(t, map[string]string{
		"main.ant":         "import \"./lib/shapes.ant\" { unit }\nimport \"./lib/geometry.ant\" G\nfunc main() {}",
		"lib/shapes.ant":   "import \"./geometry.ant\" { Point }\nexport func unit(): Point { Point{} }",
		"lib/geometry.ant": "export type Point = {}",
	})
	if len(diags) > 0 {
		t.Fatal(diags[0])
	}
	var order []string
	for _, m := range prog.Modules {
		order = append(order, m.Path)
	}
	want := "lib/geometry.ant lib/shapes.ant main.ant"
	if got := strings.Join(order, " "); got != want {
		t.Errorf("order %q, want %q", got, want)
	}
	root := prog.Root.File
	if root.Imports[0].Module != 1 || root.Imports[1].Module != 0 {
		t.Errorf("import indices: %d %d", root.Imports[0].Module, root.Imports[1].Module)
	}
}

func TestErrors(t *testing.T) {
	cases := []struct {
		name    string
		sources map[string]string
		want    string
	}{
		{"missing module", map[string]string{"main.ant": "import \"./x.ant\" { a }\nfunc main() {}"}, `cannot find module "./x.ant"`},
		{"cycle", map[string]string{
			"main.ant": "import \"./a.ant\" A\nfunc main() {}",
			"a.ant":    "import \"./b.ant\" B",
			"b.ant":    "import \"./a.ant\" A",
		}, "import cycle through a.ant"},
		{"self cycle", map[string]string{"main.ant": "import \"./main.ant\" M\nfunc main() {}"}, "import cycle"},
		{"missing main", map[string]string{"main.ant": "func helper() {}"}, "no func main"},
		{"syntax error in import", map[string]string{
			"main.ant": "import \"./a.ant\" A\nfunc main() {}",
			"a.ant":    "var x = = 1",
		}, "a.ant:1:9"},
		{"unreadable root", map[string]string{}, "cannot read the module"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, diags := load(t, c.sources)
			if len(diags) == 0 {
				t.Fatalf("loaded, want an error containing %q", c.want)
			}
			if !strings.Contains(diags[0].String(), c.want) {
				t.Fatalf("diagnostic %q, want one containing %q", diags[0].String(), c.want)
			}
		})
	}
}
