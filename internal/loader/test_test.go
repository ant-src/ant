package loader

import (
	"strings"
	"testing"
)

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
