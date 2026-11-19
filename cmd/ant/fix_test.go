package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fixable is a project with two auto fixes in one file, an error no fix removes, and a file with
// a syntax error whose fix waits for the syntax error.
var fixable = map[string]string{
	"m.ant":      "let groups = [\"a\", \"b\"]\nlet s = groups.join(\",\")\nfunc save(): Result[void, Error] {\n  Result.ok()\n}\nlet t: int = \"x\"\n",
	"broken.ant": "let x = (1 +\nlet y = 2\nlet r: Result[void, Error] = Result.ok()\n",
}

const fixed = "let groups = [\"a\", \"b\"]\nlet s = String.join(groups, \",\")\nfunc save(): Result[void, Error] {\n  Result.ok(void)\n}\nlet t: int = \"x\"\n"

// TestCodeFix checks that fix applies the auto fixes and writes each file once, leaves a file
// with a syntax error and an error no fix removes, and exits with 1 for what is left.
func TestCodeFix(t *testing.T) {
	dir := tree(t, fixable)
	t.Chdir(dir)
	status, out, errOut := codeRun("fix")
	if status != 1 || out != "m.ant: 2 fixes\n" || !strings.Contains(errOut, "m.ant:6:14: cannot use string where int is expected") {
		t.Errorf("status %d, stdout %q, stderr:\n%s", status, out, errOut)
	}
	if got, _ := os.ReadFile("m.ant"); string(got) != fixed {
		t.Errorf("m.ant:\n%s\nwant:\n%s", got, fixed)
	}
	if got, _ := os.ReadFile("broken.ant"); string(got) != fixable["broken.ant"] {
		t.Errorf("broken.ant changed:\n%s", got)
	}
	// nothing is left to fix
	if status, out, _ := codeRun("fix", "m.ant"); status != 1 || out != "" {
		t.Errorf("again: status %d, stdout %q", status, out)
	}
}

// TestCodeFixDiff checks that -diff prints the change and writes nothing, and that -code fixes
// and reports only the codes it names.
func TestCodeFixDiff(t *testing.T) {
	dir := tree(t, fixable)
	t.Chdir(dir)
	status, out, _ := codeRun("fix", "-diff", "-code", "arguments", "m.ant")
	want := "--- a/m.ant\n+++ b/m.ant\n@@ -1,6 +1,6 @@\n let groups = [\"a\", \"b\"]\n let s = groups.join(\",\")\n func save(): Result[void, Error] {\n-  Result.ok()\n+  Result.ok(void)\n }\n let t: int = \"x\"\n"
	if status != 0 || out != want {
		t.Errorf("status %d, stdout:\n%s\nwant:\n%s", status, out, want)
	}
	if got, _ := os.ReadFile(filepath.Join(dir, "m.ant")); string(got) != fixable["m.ant"] {
		t.Errorf("-diff wrote m.ant:\n%s", got)
	}
}

// TestCodeFixJSON checks the JSON answer: the files fixed, whether they were written, and the
// diagnostics left.
func TestCodeFixJSON(t *testing.T) {
	dir := tree(t, fixable)
	t.Chdir(dir)
	status, out, _ := codeRun("fix", "-json", "m.ant")
	if status != 1 || !strings.HasPrefix(out, `{"files":[{"path":"m.ant","fixes":2}],"written":true,"diagnostics":[{"path":"m.ant","line":6,`) {
		t.Errorf("status %d, stdout %s", status, out)
	}
}
