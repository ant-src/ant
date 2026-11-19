package main

import (
	"strings"
	"testing"
)

// warned is a project whose main.ant has one of each warning, a library it uses, and a module
// with an error whose warning is not reported.
var warned = map[string]string{
	"lib.ant": "export func save(path: string): Result[void, Error] { File.write(path, \"x\") }\nexport func other(): int { 1 }\nexport type Count = { n: int }\n",
	"main.ant": "import \"./lib.ant\" { save, other, Count }\ntype UserId = string\n" +
		"func run(): void {\n  let unused = 5\n  save(\"a.txt\")\n  defer { save(\"b.txt\") }\n}\n" +
		"let c: Count = Count{ n: 1 }\nprintln(\"${UserId(\"u\")} ${c.n}\")\nrun()\n",
	"broken.ant": "import \"./lib.ant\" { save }\nlet n: int = \"x\"\n",
}

// TestCheckWarnings checks that warnings print with their code, leave the status 0 unless
// -strict, are counted apart from errors, and are not reported for a module with an error.
func TestCheckWarnings(t *testing.T) {
	dir := tree(t, warned)
	t.Chdir(dir)
	status, _, errOut := checkRun("main.ant")
	for _, want := range []string{
		"main.ant:1:28: warning: other is imported but never used [unused-import]",
		"main.ant:4:7: warning: unused is declared with let but never used [unused-variable]",
		"main.ant:5:3: warning: the Result of save is discarded, so its Err is lost [unused-result]",
		"main.ant:9:12: warning: the value is a UserId, which a string shows as UserId(...), not as its value [named-interpolation]",
		"\n4 warnings\n",
	} {
		if !strings.Contains(errOut, want) {
			t.Errorf("stderr lacks %q:\n%s", want, errOut)
		}
	}
	if status != 0 || strings.Contains(errOut, "b.txt") && strings.Contains(errOut, "6:11") {
		t.Errorf("status %d, stderr:\n%s", status, errOut)
	}
	if status, _, _ := checkRun("-strict", "main.ant"); status != 1 {
		t.Errorf("-strict: status %d, want 1", status)
	}
	status, _, errOut = checkRun("broken.ant")
	if status != 1 || strings.Contains(errOut, "warning") || !strings.HasSuffix(errOut, "\n1 error\n") {
		t.Errorf("broken.ant: status %d, stderr:\n%s", status, errOut)
	}
	_, out, _ := checkRun("-json", "main.ant")
	if !strings.Contains(out, `"severity":"warning","code":"unused-import"`) {
		t.Errorf("-json:\n%s", out)
	}
}

// TestTestIgnoresWarnings checks that ant test runs files with warnings and does not report them.
func TestTestIgnoresWarnings(t *testing.T) {
	dir := tree(t, map[string]string{"w_test.ant": "test \"t\" {\n  let unused = 1\n  assert(true)\n}\n"})
	t.Chdir(dir)
	var out, errOut strings.Builder
	status := run([]string{"test", "."}, strings.NewReader(""), &out, &errOut)
	if status != 0 || errOut.String() != "" || !strings.Contains(out.String(), "ok    1 passed") {
		t.Errorf("status %d, stdout %q, stderr %q", status, out.String(), errOut.String())
	}
}

// TestRenameWithWarnings checks that a warning does not stop a rename, as an error does.
func TestRenameWithWarnings(t *testing.T) {
	dir := tree(t, map[string]string{"a.ant": "func f(): int {\n  let unused = 1\n  2\n}\n\nprintln(f())\n"})
	t.Chdir(dir)
	if status, out, errOut := codeRun("rename", "a.ant:1:6", "g"); status != 0 || out != "a.ant: 2 edits\n" {
		t.Errorf("status %d, stdout %q, stderr %q", status, out, errOut)
	}
}

// TestRedeclaredKeepsFirst checks that a name declared twice is checked as its first declaration,
// so that the only error is the redeclaration.
func TestRedeclaredKeepsFirst(t *testing.T) {
	dir := tree(t, map[string]string{"r.ant": "func f(): int { 1 }\nfunc f(): string { \"a\" }\nlet n: int = f()\n" +
		"type Point = { x: int }\ntype Point = { y: int }\nlet p = Point{ x: 1 }\nprintln(\"${n} ${p.x}\")\n"})
	t.Chdir(dir)
	_, _, errOut := checkRun("r.ant")
	if strings.Count(errOut, "[redeclared]") != 2 || !strings.HasSuffix(errOut, "\n2 errors\n") {
		t.Errorf("stderr:\n%s", errOut)
	}
}

// TestUnusedImportAllNames checks that when no name of an import is used, the warnings offer no
// auto fix, whose edits would share commas and leave an empty import.
func TestUnusedImportAllNames(t *testing.T) {
	dir := tree(t, map[string]string{"lib.ant": "export func mk(): int { 1 }\nexport func g(): int { 2 }\n", "main.ant": "import \"./lib.ant\" { mk, g }\n\nprintln(1)\n"})
	t.Chdir(dir)
	_, out, _ := checkRun("-json", "main.ant")
	if strings.Count(out, `"code":"unused-import"`) != 2 || strings.Contains(out, `"fixes"`) || !strings.Contains(out, `"help":"remove the import"`) {
		t.Errorf("-json:\n%s", out)
	}
}
