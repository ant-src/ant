package format

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestGolden formats each testdata/*.in.ant and compares it with the .out.ant beside it, which
// must itself be formatted. Each pair proves the section of fmt.md its tag names.
func TestGolden(t *testing.T) {
	ins, err := filepath.Glob("testdata/*.in.ant")
	if err != nil || len(ins) == 0 {
		t.Fatalf("no golden files: %v", err)
	}
	for _, in := range ins {
		name := strings.TrimSuffix(in, ".in.ant")
		t.Run(filepath.Base(name), func(t *testing.T) {
			src, err := os.ReadFile(in)
			if err != nil {
				t.Fatal(err)
			}
			want, err := os.ReadFile(name + ".out.ant")
			if err != nil {
				t.Fatal(err)
			}
			got, err := Source(in, src)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != string(want) {
				t.Errorf("got:\n%s\nwant:\n%s", got, want)
			}
			checkProperties(t, in, string(src))
			again, err := Source(in, want)
			if err != nil || string(again) != string(want) {
				t.Errorf("the .out.ant is not formatted (err %v):\n%s", err, again)
			}
		})
	}
}

// TestBytes checks the rules of fmt.md 4.1 that golden files cannot show: line endings, tabs,
// and the edges of the file.
func TestBytes(t *testing.T) {
	for _, c := range []struct{ name, src, want string }{
		{"empty", "", ""},
		{"only blank lines", "\n\n\n", ""},
		{"crlf", "func main() {\r\n  print(1)\r\n}\r\n", "func main() {\n  print(1)\n}\n"},
		{"crlf in a template stays", "func main() {\r\n  print(`a\r\nb`)\r\n}\r\n", "func main() {\n  print(`a\r\nb`)\n}\n"},
		{"crlf in a comment goes", "/* a\r\n b */\r\nfunc main() {}\r\n", "/* a\n b */\nfunc main() {}\n"},
		{"lone cr is a space", "func main() {\r print(1)\r}\n", "func main() { print(1) }\n"},
		{"lone cr in a comment stays", "// a\rb\nfunc main() {}\n", "// a\rb\nfunc main() {}\n"},
		{"tabs", "func main() {\n\tvar x =\t1\n\tprint(x)\t// c\n}\n", "func main() {\n  var x = 1\n  print(x) // c\n}\n"},
		{"tab in a string stays", "func main() {\n\tprint(\"a\tb\", /a\tb/)\n}\n", "func main() {\n  print(\"a\tb\", /a\tb/)\n}\n"},
		{"no final newline", "func main() {}", "func main() {}\n"},
		{"an int keeps its space before a dot", "func main() { print(0 .x, 1.5 .y) }\n", "func main() { print(0 .x, 1.5.y) }\n"},
		{"trailing spaces", "func main() {   \n  print(1)   \n}   \n", "func main() {\n  print(1)\n}\n"},
	} {
		got, err := Source("t.ant", []byte(c.src))
		if err != nil {
			t.Errorf("%s: %v", c.name, err)
			continue
		}
		if string(got) != c.want {
			t.Errorf("%s:\n got %q\nwant %q", c.name, got, c.want)
		}
	}
}

// TestSyntaxError checks that a source that does not parse gives the parser's error and no
// output.
func TestSyntaxError(t *testing.T) {
	out, err := Source("t.ant", []byte("func main() {\n  print(\n}\n"))
	if err == nil || out != nil {
		t.Fatalf("out %q, err %v", out, err)
	}
	if !strings.HasPrefix(err.Error(), "3:1: ") {
		t.Errorf("error %q", err)
	}
	if _, err := Source("t.ant", []byte("print(1)\n")); err == nil {
		t.Error("a top-level statement formatted; fmt.md section 1 says it is a syntax error")
	}
}
