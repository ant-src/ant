package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/ant-src/ant/internal/lsp"
)

// codeWorkspace is a lexer, a module that imports it under its name and an alias, a test file
// that imports a private function, a script, and a file with CRLF line endings in a directory
// whose name has a space. Each module of calc declares a private helper, so that the bare name is
// ambiguous, and eval.ant writes a file when it runs, which ant code must never do.
var codeWorkspace = map[string]string{
	"calc/lexer.ant": `// Pos is where a token starts.
export type Pos = { line: int, col: int }

export type Token = { text: string, pos: Pos }

func posOf(line: int, col: int): Pos {
  Pos{ line: line, col }
}

// tokenize splits src into tokens.
export func tokenize(src: string): Array[Token] {
  var out: Array[Token] = []
  var col = 1
  for word in src.split(" ") {
    out.push(Token{ text: word, pos: posOf(1, col) })
    col += word.length() + 1
  }
  out
}

func helper(): int { 1 }
`,
	"calc/eval.ant": `import "./lexer.ant" { Pos, Pos: Where, tokenize }

// ran is set when the module runs, which ant code never does.
let ran = File.write("ran.txt", "ran")

export type EvalError = { message: string, pos: Pos }

export func eval(src: string): Result[int, EvalError] {
  var total = 0
  for t in tokenize(src) {
    let n = t.text.toInt()
    if n == null {
      let here: Where = t.pos
      return Err{ error: EvalError{ message: "not a number: ${t.text}", pos: here } }
    }
    total = total + n
  }
  let doubled = ([total] | Array.map({ x * 2 }) | Array.sum()) ?? 0
  Ok{ value: doubled }
}

func helper(): int { 2 }
`,
	"calc/lexer_test.ant": `import "./lexer.ant" { tokenize, posOf }

test "tokenize" {
  assert(tokenize("1 2").length() == 2)
  test "positions" {
    assert(posOf(1, 2).col == 2)
  }
}
`,
	"main.ant": `import "./calc/eval.ant" { eval }

let label = "café"
let result = eval("1 2 3")
println("${label} ${result}")
`,
	"my dir/crlf.ant": "import \"../calc/lexer.ant\" { Pos }\r\n\r\nlet p: Pos = Pos{ line: 1, col: 2 }\r\nprintln(p)\r\n",
}

// codeTree writes the workspace, with extra files, and makes it the current directory.
func codeTree(t *testing.T, extra map[string]string) string {
	t.Helper()
	dir := tree(t, with(codeWorkspace, extra))
	t.Chdir(dir)
	return dir
}

func codeRun(args ...string) (int, string, string) {
	var out, errOut strings.Builder
	status := run(append([]string{"code"}, args...), strings.NewReader(""), &out, &errOut)
	return status, out.String(), errOut.String()
}

// TestCodeAnswers runs each subcommand over the workspace and compares what it prints.
func TestCodeAnswers(t *testing.T) {
	dir := codeTree(t, nil)
	posRefs := `calc/eval.ant:1:24: import "./lexer.ant" { Pos, Pos: Where, tokenize }
calc/eval.ant:1:29: import "./lexer.ant" { Pos, Pos: Where, tokenize }
calc/eval.ant:6:49: export type EvalError = { message: string, pos: Pos }
calc/eval.ant:13:17: let here: Where = t.pos (as Where)
calc/lexer.ant:4:42: export type Token = { text: string, pos: Pos }
calc/lexer.ant:6:34: func posOf(line: int, col: int): Pos {
calc/lexer.ant:7:3: Pos{ line: line, col }
my dir/crlf.ant:1:30: import "../calc/lexer.ant" { Pos }
my dir/crlf.ant:3:8: let p: Pos = Pos{ line: 1, col: 2 }
my dir/crlf.ant:3:14: let p: Pos = Pos{ line: 1, col: 2 }
`
	for _, c := range []struct {
		args           []string
		status         int
		stdout, stderr string
	}{
		// refs: the same answer from the declaration, by name, and from a use in another module
		{[]string{"refs", "calc/lexer.ant:Pos"}, 0, posRefs, "10 references in 3 files\n"},
		{[]string{"refs", "calc/lexer.ant:2:13"}, 0, posRefs, "10 references in 3 files\n"},
		{[]string{"refs", "my dir/crlf.ant:3:Pos"}, 0, posRefs, "10 references in 3 files\n"},
		{[]string{"refs", "-decl", "my dir/crlf.ant:3:5"}, 0, "my dir/crlf.ant:3:5: let p: Pos = Pos{ line: 1, col: 2 }\nmy dir/crlf.ant:4:9: println(p)\n", "2 references in 1 file\n"},
		// a private function a test file imports
		{[]string{"refs", "posOf"}, 0, `calc/lexer.ant:15:38: out.push(Token{ text: word, pos: posOf(1, col) })
calc/lexer_test.ant:1:34: import "./lexer.ant" { tokenize, posOf }
calc/lexer_test.ant:6:12: assert(posOf(1, 2).col == 2)
`, "3 references in 2 files\n"},
		// an alias is a binding of its own
		{[]string{"refs", "calc/eval.ant:13:Where"}, 0, "calc/eval.ant:13:17: let here: Where = t.pos\n", "1 reference in 1 file\n"},
		{[]string{"refs", "-json", "calc/eval.ant:13:Where"}, 0, `{"declaration":{"path":"calc/eval.ant","line":1,"column":34,"start":{"line":1,"column":34},"end":{"line":1,"column":39}},"references":[{"location":{"path":"calc/eval.ant","line":13,"column":17,"start":{"line":13,"column":17},"end":{"line":13,"column":22}},"alias":null}],"complete":true,"skipped":[]}` + "\n", ""},
		// a prelude function, searched for in every program
		{[]string{"refs", "String.toInt"}, 0, "calc/eval.ant:11:20: let n = t.text.toInt()\n", "1 reference in 1 file\n"},

		{[]string{"def", "main.ant:4:eval"}, 0, "calc/eval.ant:8:13: func eval(src: string): Result[int, EvalError]\n", ""},
		{[]string{"def", "calc/eval.ant:helper"}, 0, "calc/eval.ant:22:6: func helper(): int\n", ""},
		{[]string{"def", "main.ant:3:5"}, 0, "main.ant:3:5: let label: string\n", ""},
		{[]string{"def", "main.ant:1:9"}, 0, "calc/eval.ant:1:1: import \"./calc/eval.ant\"\n", ""},
		{[]string{"def", "-json", "Pos"}, 0, `{"location":{"path":"calc/lexer.ant","line":2,"column":13,"start":{"line":2,"column":13},"end":{"line":2,"column":16}},"signature":"type Pos = { line: int, col: int }"}` + "\n", ""},
		{[]string{"def", "String.toInt"}, 0, "prelude.d.ant:689:13: func String.toInt(s: string, base: int = 10): int | null\n", ""},

		// type: a name, narrowed; an operator; a step of a pipeline; and no answer
		{[]string{"type", "calc/eval.ant:18:doubled"}, 0, "let doubled: int\n", ""},
		{[]string{"type", "calc/lexer.ant:tokenize"}, 0, "func tokenize(src: string): Array[Token]\n    tokenize splits src into tokens.\n", ""},
		{[]string{"type", "calc/eval.ant:16:21"}, 0, "let n: int | null\n// here: int\n", ""},
		{[]string{"type", "calc/eval.ant:16:19"}, 0, "int\n", ""},
		{[]string{"type", "calc/eval.ant:12:10"}, 0, "bool\n", ""},
		{[]string{"type", "calc/eval.ant:18:26"}, 0, "Array[int]\n", ""},
		{[]string{"type", "calc/eval.ant:18:49"}, 0, "int | null\n", ""},
		{[]string{"type", "-json", "calc/eval.ant:18:49"}, 0, `{"kind":"expression","text":"int | null","doc":"","location":{"path":"calc/eval.ant","line":18,"column":18,"start":{"line":18,"column":18},"end":{"line":18,"column":62}}}` + "\n", ""},
		{[]string{"type", "calc/eval.ant:12:5"}, 1, "", "ant code: calc/eval.ant:12:5 is on \"if\", which has no type\n"},
		{[]string{"type", "calc/eval.ant:18:48"}, 1, "", "ant code: calc/eval.ant:18:48 is not on an expression\n"},
		{[]string{"type", "calc/eval.ant:3:5"}, 1, "", "ant code: calc/eval.ant:3:5 is not on an expression\n"},

		{[]string{"symbols"}, 0, `calc/eval.ant:4:5: let ran
calc/eval.ant:6:13: export type EvalError = { message: string, pos: Pos }
calc/eval.ant:8:13: export func eval(src: string): Result[int, EvalError]
calc/eval.ant:22:6: func helper(): int
calc/lexer.ant:2:13: export type Pos = { line: int, col: int }
calc/lexer.ant:4:13: export type Token = { text: string, pos: Pos }
calc/lexer.ant:6:6: func posOf(line: int, col: int): Pos
calc/lexer.ant:11:13: export func tokenize(src: string): Array[Token]
calc/lexer.ant:21:6: func helper(): int
calc/lexer_test.ant:3:6: test "tokenize"
calc/lexer_test.ant:5:8: test "tokenize/positions"
main.ant:3:5: let label
main.ant:4:5: let result
my dir/crlf.ant:3:5: let p: Pos
`, ""},
		{[]string{"symbols", "-all", "calc/eval.ant"}, 0, `calc/eval.ant:4:5: let ran
calc/eval.ant:6:13: export type EvalError = { message: string, pos: Pos }
calc/eval.ant:6:27: message: string
calc/eval.ant:6:44: pos: Pos
calc/eval.ant:8:13: export func eval(src: string): Result[int, EvalError]
calc/eval.ant:9:7: var total: int
calc/eval.ant:11:9: let n: int | null
calc/eval.ant:13:11: let here: Pos
calc/eval.ant:18:7: let doubled: int
calc/eval.ant:22:6: func helper(): int
`, ""},
		{[]string{"symbols", "-match", "POS", "-json", "calc/lexer.ant"}, 0, `{"symbols":[{"location":{"path":"calc/lexer.ant","line":2,"column":13,"start":{"line":2,"column":13},"end":{"line":2,"column":16}},"kind":"type","name":"Pos","signature":"export type Pos = { line: int, col: int }","exported":true,"container":null},{"location":{"path":"calc/lexer.ant","line":6,"column":6,"start":{"line":6,"column":6},"end":{"line":6,"column":11}},"kind":"function","name":"posOf","signature":"func posOf(line: int, col: int): Pos","exported":false,"container":null}]}` + "\n", ""},
		{[]string{"symbols", "-match", "nothing"}, 1, "", ""},

		{[]string{"rename", "-diff", "my dir/crlf.ant:3:p", "q"}, 0, "--- a/my dir/crlf.ant\n+++ b/my dir/crlf.ant\n@@ -1,4 +1,4 @@\n import \"../calc/lexer.ant\" { Pos }\r\n \r\n-let p: Pos = Pos{ line: 1, col: 2 }\r\n-println(p)\r\n+let q: Pos = Pos{ line: 1, col: 2 }\r\n+println(q)\r\n", ""},
		{[]string{"rename", "-diff", "calc/lexer.ant:6:line", "ln"}, 0, posOfDiff("func posOf(ln: int, col: int): Pos {", "  Pos{ line: ln, col }"), ""},
		{[]string{"rename", "-diff", "calc/lexer.ant:6:col", "c"}, 0, posOfDiff("func posOf(line: int, c: int): Pos {", "  Pos{ line: line, col: c }"), ""},
		{[]string{"rename", "Pos", "Pos"}, 0, "", ""},

		// errors
		{[]string{"refs", "calc/lexer.ant:7:line"}, 1, "", "ant code: line names 2 bindings on line 7\ncalc/lexer.ant:7:8: Pos{ line: line, col }\ncalc/lexer.ant:7:14: Pos{ line: line, col }\n"},
		{[]string{"refs", "helper"}, 1, "", "ant code: helper names 2 declarations\ncalc/eval.ant:22:6: func helper(): int { 2 }\ncalc/lexer.ant:21:6: func helper(): int { 1 }\n"},
		{[]string{"refs", "-json", "helper"}, 1, `{"error":{"code":"ambiguous","message":"helper names 2 declarations","candidates":[{"path":"calc/eval.ant","line":22,"column":6,"start":{"line":22,"column":6},"end":{"line":22,"column":12}},{"path":"calc/lexer.ant","line":21,"column":6,"start":{"line":21,"column":6},"end":{"line":21,"column":12}}]}}` + "\n", ""},
		{[]string{"def", "main.ant:3:14"}, 1, "", "ant code: main.ant:3:14 selects no name\n"},
		{[]string{"def", "main.ant:4:4"}, 1, "", "ant code: main.ant:4:4 selects no name\n"},
		{[]string{"def", "calc/eval.ant:Nope"}, 1, "", "ant code: calc/eval.ant declares no Nope\n"},
		{[]string{"def", "main.ant:9:1"}, 2, "", "ant code: main.ant has no line 9\n"},
		{[]string{"def", "main.ant:3:40"}, 2, "", "ant code: line 3 of main.ant has no column 40\n"},
		{[]string{"def", "main.ant:3:18"}, 2, "", "ant code: column 18 of line 3 of main.ant is inside a character\n"},
		{[]string{"def", "nope.ant:1:1"}, 2, "", "ant code: cannot read nope.ant: no such file or directory\n"},
		{[]string{"def", "main.ant"}, 2, "", "ant code: \"main.ant\" is not a location: file:line:col, file:line:name, file:Name, or Name\n"},
		{[]string{"def", "-json", "main.ant:3"}, 2, `{"error":{"code":"usage","message":"\"main.ant:3\" is not a location: file:line:col, file:line:name, file:Name, or Name"}}` + "\n", ""},
		{[]string{"def", "-decl", "Pos"}, 2, "", "ant code: flag provided but not defined: -decl\nusage: ant code def [-json] [-root dir] <loc>\n"},
		{[]string{"rename", "calc/eval.ant:6:pos", "where"}, 1, "", "ant code: a record field is not renamed\n"},
		{[]string{"rename", "Pos", "Token"}, 1, "", "ant code: Token is already declared where Pos is used\n"},
		{[]string{"rename", "String.toInt", "parse"}, 1, "", "ant code: a name of the prelude cannot be renamed\n"},
		{[]string{"rename", "Pos", "if"}, 1, "", "ant code: if is a keyword\n"},
		{[]string{"refs", "-h"}, 0, "usage: ant code refs [-json] [-root dir] [-decl] <loc>\n", ""},
	} {
		status, out, errOut := codeRun(c.args...)
		if status != c.status || out != c.stdout || errOut != c.stderr {
			t.Errorf("ant code %s: status %d\nstdout:\n%s\nstderr:\n%s\nwant status %d\nstdout:\n%s\nstderr:\n%s",
				strings.Join(c.args, " "), status, out, errOut, c.status, c.stdout, c.stderr)
		}
	}
	// nothing ran, and nothing was written
	if _, err := os.Stat(filepath.Join(dir, "ran.txt")); err == nil {
		t.Error("ant code ran a module")
	}
	for name, src := range codeWorkspace {
		if got := read(t, filepath.Join(dir, name)); got != src {
			t.Errorf("%s changed:\n%s", name, got)
		}
	}
}

// posOfDiff is the diff of a rename that changes the two lines of posOf.
func posOfDiff(sig, body string) string {
	return "--- a/calc/lexer.ant\n+++ b/calc/lexer.ant\n@@ -3,8 +3,8 @@\n \n export type Token = { text: string, pos: Pos }\n \n" +
		"-func posOf(line: int, col: int): Pos {\n-  Pos{ line: line, col }\n+" + sig + "\n+" + body + "\n }\n \n // tokenize splits src into tokens.\n"
}

// TestCodeRename renames a type across modules and checks that the workspace still checks and
// its tests pass, and that every reference was renamed.
func TestCodeRename(t *testing.T) {
	dir := codeTree(t, nil)
	status, out, errOut := codeRun("rename", "calc/lexer.ant:Pos", "Span")
	want := "calc/eval.ant: 3 edits\ncalc/lexer.ant: 4 edits\nmy dir/crlf.ant: 3 edits\n"
	if status != 0 || out != want || errOut != "" {
		t.Fatalf("status %d, stdout:\n%s\nstderr:\n%s", status, out, errOut)
	}
	if status, _, errOut := checkRun("."); status != 0 {
		t.Errorf("ant check after the rename: %d\n%s", status, errOut)
	}
	var tout, terr strings.Builder
	if status := run([]string{"test", "."}, strings.NewReader(""), &tout, &terr); status != 0 {
		t.Errorf("ant test after the rename: %d\n%s%s", status, tout.String(), terr.String())
	}
	if got := read(t, filepath.Join(dir, "my dir/crlf.ant")); got != strings.ReplaceAll(codeWorkspace["my dir/crlf.ant"], "Pos", "Span") {
		t.Errorf("crlf.ant:\n%q", got)
	}
	// the comment and the alias stay
	if got := read(t, filepath.Join(dir, "calc/lexer.ant")); !strings.HasPrefix(got, "// Pos is where a token starts.\nexport type Span =") {
		t.Errorf("lexer.ant:\n%s", got)
	}
	if got := read(t, filepath.Join(dir, "calc/eval.ant")); !strings.HasPrefix(got, "import \"./lexer.ant\" { Span, Span: Where, tokenize }") {
		t.Errorf("eval.ant:\n%s", got)
	}
	// the same references under the new name, from the declaration and from a use
	_, fromDecl, _ := codeRun("refs", "calc/lexer.ant:Span")
	_, fromUse, _ := codeRun("refs", "my dir/crlf.ant:3:Span")
	if strings.Count(fromDecl, "\n") != 10 || fromDecl != fromUse {
		t.Errorf("references of Span:\n%s\nfrom a use:\n%s", fromDecl, fromUse)
	}
	if _, err := os.Stat(filepath.Join(dir, "ran.txt")); err == nil {
		t.Error("ant code ran a module")
	}
}

// TestCodeRenameJSON checks the JSON of a rename that writes.
func TestCodeRenameJSON(t *testing.T) {
	codeTree(t, nil)
	status, out, _ := codeRun("rename", "-json", "calc/lexer.ant:6:col", "c")
	want := `{"files":[{"path":"calc/lexer.ant","edits":[{"location":{"path":"calc/lexer.ant","line":6,"column":23,"start":{"line":6,"column":23},"end":{"line":6,"column":26}},"newText":"c"},{"location":{"path":"calc/lexer.ant","line":7,"column":20,"start":{"line":7,"column":20},"end":{"line":7,"column":23}},"newText":"col: c"}]}],"written":true,"complete":true,"skipped":[]}` + "\n"
	if status != 0 || out != want {
		t.Errorf("status %d:\n%s", status, out)
	}
}

// TestCodeIncomplete checks refs and rename when a file of the workspace does not parse.
func TestCodeIncomplete(t *testing.T) {
	dir := codeTree(t, map[string]string{"broken.ant": "let x = 1\nlet y = (\n"})
	status, _, errOut := codeRun("refs", "tokenize")
	if status != 2 || !strings.Contains(errOut, "cannot read or parse broken.ant") {
		t.Errorf("refs: status %d, stderr:\n%s", status, errOut)
	}
	status, out, _ := codeRun("refs", "-json", "tokenize")
	var res struct {
		Complete bool
		Skipped  []string
	}
	if err := json.Unmarshal([]byte(out), &res); err != nil || status != 2 || res.Complete || len(res.Skipped) != 1 || res.Skipped[0] != "broken.ant" {
		t.Errorf("refs -json: status %d, %s", status, out)
	}
	// a local's search does not reach the workspace
	if status, _, errOut := codeRun("refs", "calc/eval.ant:9:total"); status != 0 {
		t.Errorf("refs of a local: status %d, %s", status, errOut)
	}
	status, out, _ = codeRun("rename", "-json", "tokenize", "split")
	if status != 2 || !strings.Contains(out, `"code":"incomplete"`) || !strings.Contains(out, `"skipped":["broken.ant"]`) {
		t.Errorf("rename: status %d, %s", status, out)
	}
	if got := read(t, filepath.Join(dir, "calc/lexer.ant")); got != codeWorkspace["calc/lexer.ant"] {
		t.Error("a refused rename wrote")
	}
	status, out, errOut = codeRun("rename", "-force", "tokenize", "split")
	if status != 0 || !strings.Contains(errOut, "warning: cannot read or parse broken.ant") || out != "calc/eval.ant: 2 edits\ncalc/lexer.ant: 1 edit\ncalc/lexer_test.ant: 2 edits\n" {
		t.Errorf("rename -force: status %d, stdout:\n%s\nstderr:\n%s", status, out, errOut)
	}
	// the location's own file must parse
	if status, _, errOut := codeRun("def", "broken.ant:1:5"); status != 2 || !strings.Contains(errOut, "broken.ant:3:1:") {
		t.Errorf("def in a file that does not parse: %d %s", status, errOut)
	}
	// symbols lists what it recovered
	status, out, errOut = codeRun("symbols", "broken.ant")
	if status != 2 || out != "broken.ant:1:5: let x\n" || !strings.Contains(errOut, "broken.ant has syntax errors") {
		t.Errorf("symbols of a file that does not parse: %d %q %q", status, out, errOut)
	}
}

// TestCodeRenameErrors checks that a rename refuses a program with errors, unless forced, and
// ignores errors in programs it does not change.
func TestCodeRenameErrors(t *testing.T) {
	dir := codeTree(t, map[string]string{"other.ant": "let x: int = \"s\"\n"})
	status, out, errOut := codeRun("rename", "posOf", "at")
	if status != 0 || out != "calc/lexer.ant: 2 edits\ncalc/lexer_test.ant: 2 edits\n" {
		t.Errorf("an error in another program: %d\n%s%s", status, out, errOut)
	}
	os.WriteFile(filepath.Join(dir, "calc/eval.ant"), []byte(codeWorkspace["calc/eval.ant"]+"let y: int = \"s\"\n"), 0o644)
	status, _, errOut = codeRun("rename", "tokenize", "split")
	if status != 1 || !strings.Contains(errOut, "-force") || !strings.Contains(errOut, "calc/eval.ant:23:14:") {
		t.Errorf("an error in a program the rename changes: %d\n%s", status, errOut)
	}
	status, out, _ = codeRun("rename", "-json", "tokenize", "split")
	if status != 1 || !strings.Contains(out, `"code":"check"`) || !strings.Contains(out, `"errors":[{"path":"calc/eval.ant","line":23,"column":14,`) {
		t.Errorf("-json: %d\n%s", status, out)
	}
	if status, _, errOut := codeRun("rename", "-force", "tokenize", "split"); status != 0 || !strings.Contains(errOut, "warning:") {
		t.Errorf("-force: %d\n%s", status, errOut)
	}
}

// TestCodeWrite checks that a rename writes nothing when a file changed since the analysis, and
// reports what it wrote when a write fails.
func TestCodeWrite(t *testing.T) {
	dir := codeTree(t, nil)
	a, b := filepath.Join(dir, "calc/eval.ant"), filepath.Join(dir, "calc/lexer.ant")
	files := []lsp.CodeFile{
		{Path: a, Old: codeWorkspace["calc/eval.ant"], New: "a\n"},
		{Path: b, Old: "not what it holds", New: "b\n"},
	}
	var out, errOut strings.Builder
	c := &coder{json: true, stdout: &out, stderr: &errOut}
	if status := c.write(files); status != 1 || !strings.Contains(out.String(), `"code":"changed","message":"files changed since they were analyzed; nothing was written","paths":["calc/lexer.ant"]`) {
		t.Errorf("changed: %d %s", status, out.String())
	}
	if read(t, a) != codeWorkspace["calc/eval.ant"] {
		t.Error("wrote although a file changed")
	}
	if runtime.GOOS == "windows" || os.Getuid() == 0 {
		return
	}
	files[1].Old = codeWorkspace["calc/lexer.ant"]
	if err := os.Chmod(b, 0o444); err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if status := c.write(files); status != 2 || !strings.Contains(out.String(), `"written":["calc/eval.ant"],"unwritten":["calc/lexer.ant"]`) {
		t.Errorf("a failed write: %d %s", status, out.String())
	}
	if read(t, a) != "a\n" || read(t, b) != codeWorkspace["calc/lexer.ant"] {
		t.Error("files after a failed write")
	}
}

// TestCodeCompletion checks that the shell completes ant code's subcommands and their flags.
func TestCodeCompletion(t *testing.T) {
	has := func(cands []candidate, want string) bool {
		for _, c := range cands {
			if c.value == want {
				return true
			}
		}
		return false
	}
	if got := complete("ant code re"); !has(got, "refs") || !has(got, "rename") {
		t.Errorf("subcommands: %v", got)
	}
	if got := complete("ant code refs -"); !has(got, "-decl") || !has(got, "-json") || has(got, "-diff") {
		t.Errorf("flags of refs: %v", got)
	}
	// the value of -root is not an operand
	if got := complete("ant code refs -root src -"); !has(got, "-decl") {
		t.Errorf("flags after -root: %v", got)
	}
	if got := complete("ant code symbols -match "); len(got) != 0 {
		t.Errorf("the value of -match: %v", got)
	}
}

// TestCodeRoots checks a workspace reached through a symbolic link, named as the root or entered
// as the current directory, and a root that is the file system's.
func TestCodeRoots(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symbolic links need privileges on Windows")
	}
	dir := tree(t, codeWorkspace)
	link := filepath.Join(t.TempDir(), "link")
	if err := os.Symlink(dir, link); err != nil {
		t.Fatal(err)
	}
	t.Chdir(link)
	status, out, errOut := codeRun("refs", "calc/lexer.ant:Pos")
	if status != 0 || strings.Count(out, "\n") != 10 || !strings.HasPrefix(out, "calc/eval.ant:1:24:") {
		t.Errorf("refs under a linked current directory: %d\n%s%s", status, out, errOut)
	}
	if status, out, errOut := codeRun("symbols", "-match", "posof", "."); status != 0 || out != "calc/lexer.ant:6:6: func posOf(line: int, col: int): Pos\n" {
		t.Errorf("symbols under a linked current directory: %d %q %q", status, out, errOut)
	}
	t.Chdir(t.TempDir())
	if status, out, errOut := codeRun("refs", "-root", link, filepath.Join(link, "calc/lexer.ant:Pos")); status != 0 || strings.Count(out, "\n") != 10 {
		t.Errorf("refs with a linked root: %d\n%s%s", status, out, errOut)
	}
	t.Chdir(dir)
	// a local's rename reads no file but its own, so the root / is not walked
	if status, out, errOut := codeRun("rename", "-diff", "-root", "/", "calc/eval.ant:9:total", "sum"); status != 0 || !strings.Contains(out, "+  var sum = 0") {
		t.Errorf("rename with the root /: %d\n%s%s", status, out, errOut)
	}
}

// TestCodeUsageJSON checks that a wrong subcommand is a JSON error under -json.
func TestCodeUsageJSON(t *testing.T) {
	status, out, errOut := codeRun("bogus", "-json")
	if status != 2 || !strings.HasPrefix(out, `{"error":{"code":"usage","message":"usage: ant code def`) || errOut != "" {
		t.Errorf("status %d, stdout %q, stderr %q", status, out, errOut)
	}
	if status, out, errOut := codeRun("bogus"); status != 2 || out != "" || !strings.HasPrefix(errOut, "usage: ant code def") {
		t.Errorf("text: status %d, stdout %q, stderr %q", status, out, errOut)
	}
}
