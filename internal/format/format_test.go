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
		{"crlf in a raw string goes", "func main() {\r\n  println(`a\r\nb`)\r\n}\r\n", "func main() {\n  println(`a\nb`)\n}\n"},
		{"crlf in a string goes", "func main() {\r\n  println(\"a\r\nb${c}\r\nd\")\r\n}\r\n", "func main() {\n  println(\"a\nb${c}\nd\")\n}\n"},
		{"crlf in a comment goes", "/* a\r\n b */\r\nfunc main() {}\r\n", "/* a\n b */\nfunc main() {}\n"},
		{"lone cr is a line break", "func main() {\r println(1)\r}\r", "func main() {\n  println(1)\n}\n"},
		{"lone cr ends a line comment", "// a\rfunc main() {}\n", "// a\nfunc main() {}\n"},
		{"lone cr in a block comment", "/* a\rb */\nfunc main() {}\n", "/* a\nb */\nfunc main() {}\n"},
		{"lone cr in a raw string is a line break", "func main() {\r  println(`a\rb`)\r}\r", "func main() {\n  println(`a\nb`)\n}\n"},
		{"tabs", "func main() {\n\tvar x =\t1\n\tprintln(x)\t// c\n}\n", "func main() {\n  var x = 1\n  println(x) // c\n}\n"},
		{"tab in a string stays", "func main() {\n\tprintln(\"a\tb\", /a\tb/)\n}\n", "func main() {\n  println(\"a\tb\", /a\tb/)\n}\n"},
		{"no final newline", "func main() {}", "func main() {}\n"},
		{"an int keeps its space before a dot", "func main() { println(0 .x, 1.5 .y) }\n", "func main() { println(0 .x, 1.5.y) }\n"},
		{"no trailing spaces in an interpolation", "func main() {\n  println(\"a ${x +   \n    y} b\")\n}\n", "func main() {\n  println(\"a ${x +\n    y} b\")\n}\n"},
		{"crlf in an interpolation goes", "func main() {\r\n  println(\"a ${x +\r\n    y}\")\r\n}\r\n", "func main() {\n  println(\"a ${x +\n    y}\")\n}\n"},
		{"crlf and trailing spaces in markup go", "var h = <p>\r\n  a   \r\n</p>\r\n", "var h = <p>\n  a\n</p>\n"},
		{"crlf and trailing spaces in a pre stay", "var h = <pre>a  \r\nb</pre>\r\n", "var h = <pre>a  \r\nb</pre>\n"},
		{"crlf in a quoted attribute stays", "var h = <p title=\"a\r\nb\"></p>\n", "var h = <p title=\"a\r\nb\"></p>\n"},
		{"tabs in markup", "var h = <ul>\n\t<li>a\tb</li>\n</ul>\n", "var h = <ul>\n  <li>a\tb</li>\n</ul>\n"},
		{"lone cr in markup is a line break", "var h = <p>\ra\r</p>\n", "var h = <p>\n  a\n</p>\n"},
		{"a comment after the { of code in markup takes no space", "var k = <p>{ /* c */\n}{ // d\n  x}</p>\n", "var k = <p>{/* c */\n}{// d\n  x\n}</p>\n"},
		{"markup in a string's interpolation stays as written", "var s = \"x ${<p>\n         a</p>} y\"\n", "var s = \"x ${<p>\n         a</p>} y\"\n"},
		{"markup in the code of a pre stays as written", "var h = <pre>{<b>\n       x</b>}</pre>\n", "var h = <pre>{<b>\n       x</b>}</pre>\n"},
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

// TestDeclarationFile: a file whose name ends in .d.ant is formatted as a declaration file
// (section 10, Declaration files), which a module's parse would reject.
func TestDeclarationFile(t *testing.T) {
	src := "type  Db.Row =  { a: int,b: int }\nfunc Db.query( sql: string ) : Array[Db.Row]\nvar Db.VERSION : string\ntype Db.Handle\n"
	want := "type Db.Row = { a: int, b: int }\nfunc Db.query(sql: string): Array[Db.Row]\nvar Db.VERSION: string\ntype Db.Handle\n"
	out, err := Source("db.d.ant", []byte(src))
	if err != nil || string(out) != want {
		t.Errorf("formatting a declaration file: %v\n%s", err, out)
	}
	if _, err := Source("db.ant", []byte(src)); err == nil {
		t.Error("a module with a func without a body was formatted")
	}
}

// TestLayoutAt checks the helper that exempts text kept as written from the layout checks of
// checkProperties: it must not exempt markup that ant fmt reindents, and must find verbatim
// text nested in the code of markup and strings.
func TestLayoutAt(t *testing.T) {
	for _, c := range []struct {
		src, at      string // at marks the byte checked: the first occurrence of it in src
		kept, interp bool
	}{
		{"var h = <div>   \n  <p /></div>\n", "   \n", false, false},
		{"var h = <pre class={\n\tx}>a</pre>\n", "\tx", false, false},
		{"var h = <pre>a  \nb</pre>\n", "  \n", true, false},
		{"var h = <p title=\"a  \nb\"></p>\n", "  \n", true, false},
		{"var h = <div>{<pre>a  \nb</pre>}</div>\n", "  \n", true, false},
		{"var h = <p>{\"a ${f(\n\tx)}\"}</p>\n", "\tx", false, true},
		{"var s = \"a ${<p>\n\tx</p>}\"\n", "\tx", true, true},
	} {
		off := strings.Index(c.src, c.at)
		if kept, interp := layoutAt(c.src, off); kept != c.kept || interp != c.interp {
			t.Errorf("%q at %q: kept %v, interp %v; want %v, %v", c.src, c.at, kept, interp, c.kept, c.interp)
		}
	}
}
