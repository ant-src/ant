package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const (
	example = "../../tests/run/example.ant"
	broken  = "../../tests/check/types.ant"
	panics  = "../../tests/run/panic_explicit.ant"
	missing = "../../tests/run/missing_main.ant"
)

func TestRunCommand(t *testing.T) {
	cases := []struct {
		name   string
		args   []string
		stdin  string
		status int
		out    string
		err    string
	}{
		{"run", []string{"run", example}, "", 0, "ada: 50.0\n", ""},
		{"run with args", []string{"run", "../../tests/run/os_args.ant", "a", "b"}, "", 0, `["a", "b"]`, ""},
		{"run short form", []string{"../../tests/run/os_args.ant", "a", "b"}, "", 0, `["a", "b"]`, ""},
		{"run panic", []string{"run", panics}, "", 1, "", "panic:"},
		{"run cannot load", []string{"run", missing}, "", 2, "", "main"},
		{"run no file", []string{"run"}, "", 2, "", "usage"},
		{"run nonexistent", []string{"run", "/nonexistent/x.ant"}, "", 2, "", ""},
		{"check clean", []string{"check", example}, "", 0, "", ""},
		{"check errors", []string{"check", broken}, "", 1, "", "types.ant:"},
		{"check shows the code", []string{"check", broken}, "", 1, "", " │ "},
		{"check counts the errors", []string{"check", broken}, "", 1, "", " errors\n"},
		{"check unknown flag", []string{"check", "-x"}, "", 2, "", "usage"},
		{"check cannot load", []string{"check", missing}, "", 2, "", "main"},
		{"help", []string{"help"}, "", 0, "usage", ""},
		{"help flag", []string{"--help"}, "", 0, "usage", ""},
		{"version", []string{"version"}, "", 0, "ant ", ""},
		{"version flag", []string{"--version"}, "", 0, "ant ", ""},
		{"unknown", []string{"frobnicate"}, "", 2, "", "usage"},
		{"repl", nil, "1 + 1\nvar s = \"x\"\ns + s\n", 0, "2\n", ""},
		{"repl owns stdin", nil, "File.readLine(Os.stdin())\n", 0, "Ok{ value: null }\n", ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var out, errOut strings.Builder
			status := run(c.args, strings.NewReader(c.stdin), &out, &errOut)
			if status != c.status {
				t.Errorf("status %d, want %d; stderr: %s", status, c.status, errOut.String())
			}
			if !strings.Contains(out.String(), c.out) {
				t.Errorf("stdout %q lacks %q", out.String(), c.out)
			}
			if !strings.Contains(errOut.String(), c.err) {
				t.Errorf("stderr %q lacks %q", errOut.String(), c.err)
			}
		})
	}
}

// TestHelpAfterCommand checks that -h or --help right after a command prints that command's
// usage to standard output and exits 0, and that version takes no arguments.
func TestHelpAfterCommand(t *testing.T) {
	for _, c := range []struct{ args, want string }{
		{"run --help", "usage: ant run <file.ant> [args...]"},
		{"check -h", "usage: ant check [-json] [path...]"},
		{"fmt --help", "usage: ant fmt"},
		{"doc --help", "usage: ant doc"},
		{"lsp -h", "usage: ant lsp"},
	} {
		var out, errOut strings.Builder
		if status := run(strings.Fields(c.args), strings.NewReader(""), &out, &errOut); status != 0 || !strings.HasPrefix(out.String(), c.want) {
			t.Errorf("ant %s: status %d, output %q, want %q", c.args, status, out.String(), c.want)
		}
	}
	var out, errOut strings.Builder
	if status := run([]string{"version", "x"}, strings.NewReader(""), &out, &errOut); status != 2 {
		t.Errorf("ant version x: status %d, want 2", status)
	}
}

// TestRunWithoutSuffix checks that ant runs an existing file whose name lacks .ant, as a #!
// script installed without its suffix is.
func TestRunWithoutSuffix(t *testing.T) {
	path := filepath.Join(t.TempDir(), "hello")
	if err := os.WriteFile(path, []byte("#!/usr/bin/env ant\nprintln(Os.args())\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	var out, errOut strings.Builder
	if status := run([]string{path, "a"}, strings.NewReader(""), &out, &errOut); status != 0 || out.String() != "[\"a\"]\n" {
		t.Errorf("status %d, output %q, errors %q", status, out.String(), errOut.String())
	}
}
