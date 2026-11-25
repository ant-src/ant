package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"ant-lang.org"
)

func checkRun(args ...string) (int, string, string) {
	var out, errOut strings.Builder
	status := run(append([]string{"check"}, args...), strings.NewReader(""), &out, &errOut)
	return status, out.String(), errOut.String()
}

// project is a script, a library it imports, and a test file of the library, which reaches its
// private names.
var project = map[string]string{
	"main.ant":         "import \"./lib/geo.ant\" { area }\n\nprintln(area(2, 3))\n",
	"lib/geo.ant":      "export func area(w: int, h: int): int {\n  w * h\n}\n\nfunc half(n: int): int {\n  n / 2\n}\n",
	"lib/geo_test.ant": "import \"./geo.ant\" { area, half }\n\ntest \"area\" {\n  assert(area(2, 3) == 6)\n  assert(half(4) == 2)\n}\n",
}

func with(base map[string]string, extra map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range base {
		out[k] = v
	}
	for k, v := range extra {
		out[k] = v
	}
	return out
}

// TestCheckDirectory checks that a directory is checked whole, test files included, and that
// a clean one prints nothing.
func TestCheckDirectory(t *testing.T) {
	dir := tree(t, project)
	if status, out, errOut := checkRun(dir); status != 0 || out != "" || errOut != "" {
		t.Fatalf("clean: status %d, stdout %q, stderr %q", status, out, errOut)
	}
	dir = tree(t, with(project, map[string]string{
		"lib/geo_test.ant": "import \"./geo.ant\" { half }\n\ntest \"half\" {\n  assert(half(4) == \"2\")\n}\n",
	}))
	status, _, errOut := checkRun(dir)
	if status != 1 || !strings.Contains(errOut, "geo_test.ant:4:") || !strings.HasSuffix(errOut, "1 error\n") {
		t.Errorf("test file: status %d, stderr:\n%s", status, errOut)
	}
}

// TestCheckSharedModule checks that a module several roots import is checked once.
func TestCheckSharedModule(t *testing.T) {
	dir := tree(t, with(project, map[string]string{
		"lib/geo.ant": "export func area(w: int, h: int): int {\n  \"oops\"\n}\n\nfunc half(n: int): int {\n  n / 2\n}\n",
	}))
	status, _, errOut := checkRun(dir)
	if status != 1 || strings.Count(errOut, "geo.ant:2:3:") != 1 || !strings.HasSuffix(errOut, "1 error\n") {
		t.Errorf("status %d, stderr:\n%s", status, errOut)
	}
}

// TestCheckLoadError checks that a program that does not load is reported and checked past its
// syntax error, as are the programs that import it, with no error about a name the error hid,
// and that the status is then 2.
func TestCheckLoadError(t *testing.T) {
	dir := tree(t, map[string]string{
		"broken.ant":      "let x = (1 +\n",
		"uses_broken.ant": "import \"./broken.ant\" { x }\n\nlet y: int = \"unchecked\"\n",
		"fine.ant":        "let z: int = \"checked\"\n",
	})
	status, _, errOut := checkRun(dir)
	if status != 2 {
		t.Errorf("status %d, want 2", status)
	}
	for _, want := range []string{"broken.ant:2:1: unexpected end of file", "fine.ant:1:14:", "uses_broken.ant:3:14:", "3 errors\n"} {
		if !strings.Contains(errOut, want) {
			t.Errorf("stderr lacks %q:\n%s", want, errOut)
		}
	}
	if strings.Count(errOut, "unexpected end of file") != 1 || strings.Contains(errOut, "uses_broken.ant:1") {
		t.Errorf("stderr:\n%s", errOut)
	}
}

// TestCheckJSON checks the JSON lines: the position as the text prints it, and the span it
// underlines, which starts before the position for an operator's error.
func TestCheckJSON(t *testing.T) {
	dir := tree(t, map[string]string{
		"a.ant": "let n = 1\nlet ok = n == \"1\"\n",
		"b.ant": "let s: string = 2\n",
	})
	status, out, errOut := checkRun("-json", dir)
	a, b := filepath.Join(dir, "a.ant"), filepath.Join(dir, "b.ant")
	want := `{"path":"` + a + `","line":2,"column":12,"severity":"error","code":"operands","message":"int and string cannot be compared with ==: neither is assignable to the other","start":{"line":2,"column":10},"end":{"line":2,"column":18}}
{"path":"` + b + `","line":1,"column":17,"severity":"error","code":"type-mismatch","message":"cannot use int where string is expected","start":{"line":1,"column":17},"end":{"line":1,"column":18}}
`
	if status != 1 || out != want || errOut != "" {
		t.Errorf("status %d, stderr %q, stdout:\n%s\nwant:\n%s", status, errOut, out, want)
	}
}

// TestCheckPaths checks that with no path the current directory is checked, that directories
// whose names begin with a dot are left out, that a file named directly is checked whatever
// its name, and that a declaration file is checked on its own.
func TestCheckPaths(t *testing.T) {
	dir := tree(t, with(project, map[string]string{
		".hidden/x.ant": "let q: int = \"hidden\"\n",
		"script":        "let r: int = \"named\"\n",
		"shout.d.ant":   "// Shout.loud returns text in capitals.\nfunc Shout.loud(text: string): Strin\n",
	}))
	t.Chdir(dir)
	status, _, errOut := checkRun()
	if status != 1 || !strings.Contains(errOut, "shout.d.ant:2:32: unknown type Strin") || strings.Contains(errOut, "hidden") ||
		strings.Contains(errOut, "named") {
		t.Errorf("no path: status %d, stderr:\n%s", status, errOut)
	}
	if status, _, errOut := checkRun("script", "lib"); status != 1 || !strings.Contains(errOut, "script:1:14:") {
		t.Errorf("named file: status %d, stderr:\n%s", status, errOut)
	}
	if status, _, errOut := checkRun("nonexistent"); status != 2 || !strings.Contains(errOut, "nonexistent") {
		t.Errorf("missing path: status %d, stderr:\n%s", status, errOut)
	}
	if status, _, errOut := checkRun("-x"); status != 2 || !strings.Contains(errOut, "usage: ant check") {
		t.Errorf("unknown flag: status %d, stderr:\n%s", status, errOut)
	}
}

// TestCheckSymlinkDirectory checks that a directory named through a symbolic link is walked
// through its target, by ant check and ant test alike.
func TestCheckSymlinkDirectory(t *testing.T) {
	dir := tree(t, map[string]string{
		"real/a.ant":      "let x: int = \"s\"\n",
		"real/a_test.ant": "test \"t\" {\n  assert(1 == 2)\n}\n",
	})
	link := filepath.Join(dir, "link")
	if err := os.Symlink(filepath.Join(dir, "real"), link); err != nil {
		t.Skip(err)
	}
	if status, _, errOut := checkRun(link); status != 1 || !strings.Contains(errOut, filepath.Join(link, "a.ant")+":1:14:") {
		t.Errorf("check: status %d, stderr:\n%s", status, errOut)
	}
	var out, errOut strings.Builder
	if status := run([]string{"test", link}, strings.NewReader(""), &out, &errOut); status != 1 {
		t.Errorf("test: status %d, stdout:\n%s", status, out.String())
	}
}

// TestCheckDeclarationOnce checks that a declaration file two paths reach is checked once.
func TestCheckDeclarationOnce(t *testing.T) {
	dir := tree(t, map[string]string{"sub/x.d.ant": "// Shout.loud is loud.\nfunc Shout.loud(text: string): Strin\n"})
	t.Chdir(dir)
	status, _, errOut := checkRun(".", "sub", "sub/x.d.ant", "./sub/x.d.ant")
	if status != 1 || strings.Count(errOut, "unknown type Strin") != 1 || !strings.HasSuffix(errOut, "1 error\n") {
		t.Errorf("status %d, stderr:\n%s", status, errOut)
	}
}

// TestCheckEmptyDirectory checks that a directory with no .ant file under it is an error, so
// that a check of the wrong directory does not pass.
func TestCheckEmptyDirectory(t *testing.T) {
	dir := tree(t, map[string]string{"notes.txt": "", ".hidden/x.ant": "let x = 1\n"})
	status, _, errOut := checkRun(dir)
	if status != 2 || errOut != "ant check: no .ant files under "+dir+"\n" {
		t.Errorf("status %d, stderr %q", status, errOut)
	}
}

// TestCheckJSONWholeFile checks that an error with no position has line and column 0 and no
// span.
func TestCheckJSONWholeFile(t *testing.T) {
	var out strings.Builder
	writeJSONDiagnostics(&out, []ant.Diagnostic{{Pos: ant.Position{Path: "gone.ant"}, Message: "cannot read the module"}})
	if want := `{"path":"gone.ant","line":0,"column":0,"message":"cannot read the module"}` + "\n"; out.String() != want {
		t.Errorf("got %s, want %s", out.String(), want)
	}
}

// TestCheckEveryError checks that a file with syntax errors is checked past them, so that one run
// reports every error of it, and that an error that only follows from another is not reported:
// a type the checker could not determine, the parameters a lambda rejected for its shape, a name
// imported without export, a declaration a syntax error hid, and a line a syntax error broke.
func TestCheckEveryError(t *testing.T) {
	dir := tree(t, map[string]string{
		"multi.ant":   "let a = 1 as int\nlet b: int = \"x\"\nlet match = 2\nlet c: string = 5\n",
		"chain.ant":   "type Op = { op: string }\nfunc f(k: Op | int): string {\n  let m = k.foo\n  m?.bar ?? \"x\"\n}\n",
		"lambda.ant":  "func top(xs: Array[int]): Array[int] {\n  xs.sortWith({ b - a })\n}\n",
		"temp.ant":    "export func celsius(f: float64): float64 { (f - 32.0) * ratio() }\nfunc ratio(): float64 { 5.0 / 9.0 }\n",
		"private.ant": "import \"./temp.ant\" { celsius, ratio }\nprintln(ratio())\n",
		"broken.ant":  "let x = (1 +\nexport func f(): int { 1 }\n",
		"hidden.ant":  "import \"./broken.ant\" B\nimport \"./broken.ant\" { x }\nprintln(B.x + x)\nlet q: int = \"s\"\n",
		"line.ant":    "export func f( {\n",
	})
	status, _, errOut := checkRun(dir)
	if status != 2 {
		t.Errorf("status %d, want 2", status)
	}
	var got []string
	for _, l := range strings.Split(errOut, "\n") {
		if strings.HasPrefix(l, dir) {
			got = append(got, strings.TrimPrefix(l, dir+string(filepath.Separator)))
		}
	}
	want := []string{
		"broken.ant:2:1: unexpected 'export' [syntax]",
		"chain.ant:3:12: Op | int has no field foo on every member: narrow the value first [needs-narrowing]",
		"hidden.ant:4:14: cannot use string where int is expected [type-mismatch]",
		"lambda.ant:2:15: a lambda has at most one parameter, x; func(int, int): int expects 2: write func with named parameters [lambda]",
		"line.ant:1:16: expected a parameter name, got '{' [syntax]",
		"multi.ant:1:11: 'as' is a reserved word [syntax]",
		"multi.ant:2:14: cannot use string where int is expected [type-mismatch]",
		"multi.ant:3:5: 'match' cannot be a variable name [syntax]",
		"multi.ant:4:17: cannot use int where string is expected [type-mismatch]",
		"private.ant:1:32: ratio is not exported by ./temp.ant [not-exported]",
	}
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Errorf("got\n%s\nwant\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}
