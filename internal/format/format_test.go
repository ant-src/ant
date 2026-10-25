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
		{"crlf", "func main() {\r\n  println(1)\r\n}\r\n", "func main() {\n  println(1)\n}\n"},
		{"crlf in a template stays", "func main() {\r\n  println(`a\r\nb`)\r\n}\r\n", "func main() {\n  println(`a\r\nb`)\n}\n"},
		{"crlf in a comment goes", "/* a\r\n b */\r\nfunc main() {}\r\n", "/* a\n b */\nfunc main() {}\n"},
		{"lone cr is a line break", "func main() {\r println(1)\r}\r", "func main() {\n  println(1)\n}\n"},
		{"lone cr ends a line comment", "// a\rfunc main() {}\n", "// a\nfunc main() {}\n"},
		{"lone cr in a block comment", "/* a\rb */\nfunc main() {}\n", "/* a\nb */\nfunc main() {}\n"},
		{"lone cr in a template stays", "func main() {\r  println(`a\rb`)\r}\r", "func main() {\n  println(`a\rb`)\n}\n"},
		{"tabs", "func main() {\n\tvar x =\t1\n\tprintln(x)\t// c\n}\n", "func main() {\n  var x = 1\n  println(x) // c\n}\n"},
		{"tab in a string stays", "func main() {\n\tprintln(\"a\tb\", /a\tb/)\n}\n", "func main() {\n  println(\"a\tb\", /a\tb/)\n}\n"},
		{"no final newline", "func main() {}", "func main() {}\n"},
		{"an int keeps its space before a dot", "func main() { println(0 .x, 1.5 .y) }\n", "func main() { println(0 .x, 1.5.y) }\n"},
		{"no trailing spaces in an interpolation", "func main() {\n  println(`a ${x +   \n    y} b`)\n}\n", "func main() {\n  println(`a ${x +\n    y} b`)\n}\n"},
		{"crlf in an interpolation goes", "func main() {\r\n  println(`a ${x +\r\n    y}`)\r\n}\r\n", "func main() {\n  println(`a ${x +\n    y}`)\n}\n"},
		{"trailing spaces", "func main() {   \n  println(1)   \n}   \n", "func main() {\n  println(1)\n}\n"},
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
	out, err := Source("t.ant", []byte("func main() {\n  println(\n}\n"))
	if err == nil || out != nil {
		t.Fatalf("out %q, err %v", out, err)
	}
	if !strings.HasPrefix(err.Error(), "3:1: ") {
		t.Errorf("error %q", err)
	}
	if out, err := Source("t.ant", []byte("println( 1 );\n")); err != nil || string(out) != "println(1)\n" {
		t.Errorf("a script formats as a module (fmt.md section 1): %q, %v", out, err)
	}
}
