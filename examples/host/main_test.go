package main

import (
	"os"
	"strings"
	"testing"
	"testing/fstest"
)

func host(t *testing.T, script, input string) (int, string, string) {
	t.Helper()
	var out, errOut strings.Builder
	status := run(fstest.MapFS{"main.ant": {Data: []byte(script)}}, "main.ant", strings.NewReader(input), &out, &errOut)
	return status, out.String(), errOut.String()
}

func TestHost(t *testing.T) {
	var out, errOut strings.Builder
	status := run(files, "shout.ant", strings.NewReader("hello\nant\nquit\nnot read\n"), &out, &errOut)
	if want := "1: HELLO\n2: ANT\n2 lines\n"; status != 0 || out.String() != want || errOut.String() != "" {
		t.Errorf("status %d, stdout %q, stderr %q", status, out.String(), errOut.String())
	}
}

// TestUnwind checks that what ends a script inside the handler passes the loop, which recovers
// panics: a panic, and a failed try at the top level after the loop.
func TestUnwind(t *testing.T) {
	status, out, errOut := host(t, "App.run({ [1][x.line.length()] == 1 })\nprintln(\"after\")\n", "\nab\nc\n")
	if status != 1 || out != "" || !strings.Contains(errOut, "panic: index 2 out of range") || strings.Contains(errOut, "loop:") {
		t.Errorf("a panic: status %d, stdout %q, stderr %q", status, out, errOut)
	}
	// a script that reads a name only a declaration file beside it declares is refused, and
	// the host says why
	fsys := fstest.MapFS{"main.ant": {Data: []byte("println(Other.N)\n")}, "other.d.ant": {Data: []byte("let Other.N: int\n")}}
	var out2, err2 strings.Builder
	if status := run(fsys, "main.ant", strings.NewReader(""), &out2, &err2); status != 2 || !strings.Contains(err2.String(), "Other is declared in other.d.ant") {
		t.Errorf("a refused program: status %d, stderr %q", status, err2.String())
	}
	for _, name := range []string{"File", "Http", "Net", "Os", "Cmd"} {
		if status, _, errOut := host(t, "println("+name+")\n", ""); status != 1 || !strings.Contains(errOut, name) {
			t.Errorf("removed %s: status %d, stderr %q", name, status, errOut)
		}
	}
	// without the script's own namespaces
	if status, _, errOut := host(t, "println(File.exists(\"x\"))\n", ""); status != 1 || !strings.Contains(errOut, "File") {
		t.Errorf("a removed namespace: status %d, stderr %q", status, errOut)
	}
}

// TestGuide checks that the Go the guide shows is the Go of this example, and its Ant the
// example's files.
func TestGuide(t *testing.T) {
	page, err := os.ReadFile("../../docs/guide/embedding.md")
	if err != nil {
		t.Fatal(err)
	}
	var source strings.Builder
	for _, name := range []string{"main.go", "loop.go", "app.d.ant", "shout.ant"} {
		b, err := os.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		source.Write(b)
	}
	flat := func(s string) string { return strings.Join(strings.Fields(s), " ") }
	have := flat(source.String())
	parts := strings.Split(string(page), "```")
	for i := 1; i < len(parts); i += 2 {
		lang, code, _ := strings.Cut(parts[i], "\n")
		if lang != "go" && lang != "ant [app.d.ant]" && lang != "ant [shout.ant]" {
			continue
		}
		if !strings.Contains(have, flat(code)) {
			t.Errorf("the guide shows code that is not in this example:\n%s", code)
		}
	}
}
