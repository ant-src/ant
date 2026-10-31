package main

import (
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
		{"check wrong arity", []string{"check"}, "", 2, "", "usage"},
		{"check cannot load", []string{"check", missing}, "", 2, "", "main"},
		{"help", []string{"help"}, "", 0, "usage", ""},
		{"help flag", []string{"--help"}, "", 0, "usage", ""},
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
