package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	formatted   = "func main() {\n  print(1)\n}\n"
	unformatted = "func main() {\nprint( 1 );\n}"
	syntaxError = "func main() {\n  print(\n}\n"
)

// tree writes files under a fresh directory and returns it.
func tree(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := t.TempDir()
	for name, src := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(src), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func read(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func fmtRun(args []string, stdin string) (int, string, string) {
	var out, errOut strings.Builder
	status := run(append([]string{"fmt"}, args...), strings.NewReader(stdin), &out, &errOut)
	return status, out.String(), errOut.String()
}

// TestFmtStatus checks each row of the exit-status table of fmt.md section 1.
func TestFmtStatus(t *testing.T) {
	dir := tree(t, map[string]string{"ok.ant": formatted, "bad.ant": unformatted, "broken.ant": syntaxError})
	ok, bad, broken := filepath.Join(dir, "ok.ant"), filepath.Join(dir, "bad.ant"), filepath.Join(dir, "broken.ant")
	for _, c := range []struct {
		name   string
		args   []string
		status int
	}{
		{"formatted", []string{ok}, 0},
		{"check formatted", []string{"--check", ok}, 0},
		{"check unformatted", []string{"--check", ok, bad}, 1},
		{"check syntax error outranks", []string{"--check", bad, broken}, 2},
		{"no path", nil, 2},
		{"check and no path", []string{"--check"}, 2},
		{"unknown flag", []string{"-w", ok}, 2},
		{"missing file", []string{filepath.Join(dir, "none.ant")}, 2},
	} {
		if status, _, errOut := fmtRun(c.args, ""); status != c.status {
			t.Errorf("%s: status %d, want %d; stderr %q", c.name, status, c.status, errOut)
		}
	}
	if got := read(t, bad); got != unformatted {
		t.Errorf("--check wrote the file:\n%s", got)
	}
}

// TestFmtInPlace checks that files are rewritten, that a syntax error leaves its file as it was
// while the others are formatted, and that its error names the file and position.
func TestFmtInPlace(t *testing.T) {
	dir := tree(t, map[string]string{"a.ant": unformatted, "b.ant": syntaxError, "c.ant": unformatted})
	status, out, errOut := fmtRun([]string{filepath.Join(dir, "a.ant"), filepath.Join(dir, "b.ant"), filepath.Join(dir, "c.ant")}, "")
	if status != 2 || out != "" {
		t.Errorf("status %d, stdout %q", status, out)
	}
	if want := filepath.Join(dir, "b.ant") + ":3:1: "; !strings.HasPrefix(errOut, want) {
		t.Errorf("stderr %q, want it to start with %q", errOut, want)
	}
	for _, name := range []string{"a.ant", "c.ant"} {
		if got := read(t, filepath.Join(dir, name)); got != formatted {
			t.Errorf("%s:\n%s", name, got)
		}
	}
	if got := read(t, filepath.Join(dir, "b.ant")); got != syntaxError {
		t.Errorf("the file with a syntax error changed:\n%s", got)
	}
}

// TestFmtUnchangedNotWritten checks that a formatted file is not written at all.
func TestFmtUnchangedNotWritten(t *testing.T) {
	dir := tree(t, map[string]string{"a.ant": formatted})
	path := filepath.Join(dir, "a.ant")
	if err := os.Chmod(path, 0o444); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(dir, 0o555); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(dir, 0o755)
	if status, _, errOut := fmtRun([]string{path}, ""); status != 0 {
		t.Errorf("status %d; stderr %q", status, errOut)
	}
}

// TestFmtDirectory checks the walk: .ant files only, in lexical order, dot entries skipped, and
// paths printed as found.
func TestFmtDirectory(t *testing.T) {
	dir := tree(t, map[string]string{
		"b.ant":            unformatted,
		"a/z.ant":          unformatted,
		"a/ok.ant":         formatted,
		"notes.txt":        unformatted,
		".hidden/x.ant":    unformatted,
		"a/.dotfile.ant":   unformatted,
		"lib/deep/one.ant": unformatted,
	})
	status, out, _ := fmtRun([]string{"--check", dir}, "")
	want := []string{"a/z.ant", "b.ant", "lib/deep/one.ant"}
	for i := range want {
		want[i] = filepath.Join(dir, want[i])
	}
	if status != 1 || out != strings.Join(want, "\n")+"\n" {
		t.Errorf("status %d, stdout:\n%s\nwant:\n%s", status, out, strings.Join(want, "\n"))
	}
	if status, _, _ := fmtRun([]string{dir}, ""); status != 0 {
		t.Errorf("rewrite: status %d", status)
	}
	if got := read(t, filepath.Join(dir, "lib/deep/one.ant")); got != formatted {
		t.Errorf("not rewritten:\n%s", got)
	}
	if got := read(t, filepath.Join(dir, ".hidden/x.ant")); got != unformatted {
		t.Errorf("a dot directory was formatted")
	}
	if got := read(t, filepath.Join(dir, "notes.txt")); got != unformatted {
		t.Errorf("a .txt file in a directory was formatted")
	}
}

// TestFmtNamedFile checks that a file named on the command line is formatted whatever its
// extension, and that a named symbolic link is formatted through, staying a link.
func TestFmtNamedFile(t *testing.T) {
	dir := tree(t, map[string]string{"prog.txt": unformatted, "real.ant": unformatted})
	link := filepath.Join(dir, "link.ant")
	if err := os.Symlink("real.ant", link); err != nil {
		t.Skip(err)
	}
	if status, _, errOut := fmtRun([]string{filepath.Join(dir, "prog.txt"), link}, ""); status != 0 {
		t.Fatalf("status %d; stderr %q", status, errOut)
	}
	if got := read(t, filepath.Join(dir, "prog.txt")); got != formatted {
		t.Errorf("prog.txt:\n%s", got)
	}
	if got := read(t, filepath.Join(dir, "real.ant")); got != formatted {
		t.Errorf("the link's target:\n%s", got)
	}
	if info, err := os.Lstat(link); err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Errorf("the link was replaced: %v", err)
	}
	// A link met in a walk is not followed.
	other := tree(t, map[string]string{"x.ant": unformatted})
	if err := os.Symlink(filepath.Join(other, "x.ant"), filepath.Join(dir, "walked.ant")); err != nil {
		t.Fatal(err)
	}
	if status, out, _ := fmtRun([]string{"--check", dir}, ""); status != 0 || out != "" {
		t.Errorf("a link in the walk was followed: status %d, stdout %q", status, out)
	}
}

// TestFmtStdin checks - : the formatted source on standard output, <stdin> in errors and in
// --check.
func TestFmtStdin(t *testing.T) {
	if status, out, _ := fmtRun([]string{"-"}, unformatted); status != 0 || out != formatted {
		t.Errorf("status %d, stdout %q", status, out)
	}
	if status, out, _ := fmtRun([]string{"-"}, formatted); status != 0 || out != formatted {
		t.Errorf("formatted input: status %d, stdout %q", status, out)
	}
	if status, out, _ := fmtRun([]string{"--check", "-"}, unformatted); status != 1 || out != "<stdin>\n" {
		t.Errorf("check: status %d, stdout %q", status, out)
	}
	if status, out, _ := fmtRun([]string{"--check", "-"}, formatted); status != 0 || out != "" {
		t.Errorf("check formatted: status %d, stdout %q", status, out)
	}
	status, out, errOut := fmtRun([]string{"-"}, syntaxError)
	if status != 2 || out != "" || !strings.HasPrefix(errOut, "<stdin>:3:1: ") {
		t.Errorf("syntax error: status %d, stdout %q, stderr %q", status, out, errOut)
	}
}
