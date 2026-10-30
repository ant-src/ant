package repl

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/ant-src/ant"
)

// TestRun feeds a scripted session through the loop: declarations and expressions across
// lines, continuation prompts for incomplete input, blank lines, a syntax error, a panic, and
// an import that stays in force.
func TestRun(t *testing.T) {
	geometry, _ := filepath.Abs("../../tests/run/lib/geometry.ant")
	input := strings.Join([]string{
		"var x = 1",
		"x + 1",
		"func f(a: int): int {",
		"  a * 10",
		"}",
		"f(x)",
		"",
		"   ",
		"var y = ",
		"println(\"side\")",
		"y)",
		"panic(\"boom\")",
		"x",
		`import "` + geometry + `" { distance, Point }`,
		"distance(Point{ x: 0.0, y: 0.0 }, Point{ x: 3.0, y: 4.0 })",
		"if x > 0 { \"pos\" } else { \"neg\" }",
		"[1, 2] | Array.map({ x * 2 })",
		"_ = 1",
	}, "\n") + "\n"
	var out, errOut strings.Builder
	status := Run(context.Background(), strings.NewReader(input), &out, &errOut)
	if status != 0 {
		t.Errorf("status %d", status)
	}
	got := out.String()
	for _, want := range []string{"> 2\n", ". . > 10\n", "side\n", "> 1\n", "geometry init\n", "5.0\n", "pos\n", "[2, 4]\n"} {
		if !strings.Contains(got, want) {
			t.Errorf("stdout lacks %q:\n%s", want, got)
		}
	}
	if !strings.Contains(errOut.String(), "boom") {
		t.Errorf("stderr lacks the panic:\n%s", errOut.String())
	}
	if !strings.Contains(errOut.String(), "unexpected") && !strings.Contains(errOut.String(), "expected") {
		t.Errorf("stderr lacks the syntax error:\n%s", errOut.String())
	}
}

func TestRunExit(t *testing.T) {
	var out, errOut strings.Builder
	status := Run(context.Background(), strings.NewReader("println(1)\nOs.exit(3)\nprintln(2)\n"), &out, &errOut)
	if status != 3 {
		t.Errorf("status %d, want 3", status)
	}
	if !strings.Contains(out.String(), "1\n") || strings.Contains(out.String(), "2\n") {
		t.Errorf("stdout: %q", out.String())
	}
}

// TestRunInterrupted: an interrupt ends the loop with status 130, at the prompt or during an
// input, rather than leaving a session whose every later input is cancelled.
func TestRunInterrupted(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	pr, pw := io.Pipe()
	defer pw.Close()
	done := make(chan int)
	var out, errOut strings.Builder
	go func() { done <- Run(ctx, pr, &out, &errOut) }()
	io.WriteString(pw, "var c: Channel[int] = Channel.new()\nChannel.receive(c)\n")
	time.Sleep(50 * time.Millisecond)
	cancel()
	select {
	case status := <-done:
		if status != 130 {
			t.Errorf("status %d, want 130", status)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("the loop did not end")
	}
	if errOut.String() != "" {
		t.Errorf("stderr: %q", errOut.String())
	}
}

// TestRunLongLine checks that a line longer than any fixed buffer is read whole, and the
// inputs after it still run.
func TestRunLongLine(t *testing.T) {
	input := `var s = "` + strings.Repeat("a", 2<<20) + "\"\nString.length(s)\n7"
	var out, errOut strings.Builder
	if status := Run(context.Background(), strings.NewReader(input), &out, &errOut); status != 0 {
		t.Errorf("status %d", status)
	}
	if got := out.String(); !strings.Contains(got, "2097152\n") || !strings.Contains(got, "7\n") {
		t.Errorf("stdout %q, stderr %q", got, errOut.String())
	}
}

// TestRunUnfinishedAtEnd checks that an input the end of stdin leaves unfinished is reported,
// not dropped.
func TestRunUnfinishedAtEnd(t *testing.T) {
	var out, errOut strings.Builder
	if status := Run(context.Background(), strings.NewReader("1\nfunc f() {\n  1\n"), &out, &errOut); status != 0 {
		t.Errorf("status %d", status)
	}
	if !strings.Contains(errOut.String(), "<input>:3:1:") {
		t.Errorf("stderr lacks the syntax error: %q", errOut.String())
	}
}

// scriptedReader answers readLine from a list, as an editor would: a line, or an error with
// what Ctrl-C dropped.
type scriptedReader struct {
	steps []struct {
		line string
		err  error
	}
}

func (r *scriptedReader) readLine(prompt, pending string) (string, error) {
	if len(r.steps) == 0 {
		return "", io.EOF
	}
	st := r.steps[0]
	r.steps = r.steps[1:]
	return st.line, st.err
}

// TestLoopInterrupts checks what Ctrl-C does at the prompt: it drops the line, or the input
// so far, and says how to exit; twice in a row, it ends the REPL.
func TestLoopInterrupts(t *testing.T) {
	type step = struct {
		line string
		err  error
	}
	run := func(steps ...step) (int, string) {
		var out, errOut strings.Builder
		s, err := ant.NewSession(context.Background(), ant.WithStdout(&out), ant.WithStderr(&errOut))
		if err != nil {
			t.Fatal(err)
		}
		defer s.Close()
		l := &looper{ctx: context.Background(), s: s, r: &scriptedReader{steps: steps}, stdout: &out, stderr: &errOut}
		status := l.loop()
		return status, out.String()
	}
	// a line dropped, then a line entered: the REPL goes on
	const hint = "(press ^C again or ^D to exit)\n"
	status, out := run(step{"abc", errInterrupt}, step{"1 + 1", nil})
	if status != 0 || !strings.Contains(out, "2\n") || strings.Count(out, hint) != 1 {
		t.Errorf("drop then enter: status %d, out %q", status, out)
	}
	// Ctrl-C on an empty prompt tells how to exit; again right after, it exits
	status, out = run(step{"", errInterrupt}, step{"", errInterrupt}, step{"1 + 1", nil})
	if status != 130 || strings.Count(out, hint) != 1 || strings.Contains(out, "2\n") {
		t.Errorf("twice: status %d, out %q", status, out)
	}
	// a line dropped, then Ctrl-C again: exits too
	status, out = run(step{"abc", errInterrupt}, step{"", errInterrupt})
	if status != 130 || strings.Count(out, hint) != 1 {
		t.Errorf("drop then again: status %d, out %q", status, out)
	}
	// a line typed between two drops: no exit until the third
	status, out = run(step{"", errInterrupt}, step{"xy", errInterrupt}, step{"", errInterrupt})
	if status != 130 || strings.Count(out, hint) != 2 {
		t.Errorf("typed between: status %d, out %q", status, out)
	}
	status, _ = run(step{"", errInterrupt}, step{"1", nil}, step{"", errInterrupt})
	if status != 0 {
		t.Errorf("entered between: status %d", status)
	}
	// Ctrl-C in an unfinished input drops it
	status, out = run(step{"func f() {", nil}, step{"", errInterrupt}, step{"2", nil})
	if status != 0 || strings.Count(out, hint) != 1 || !strings.Contains(out, "2\n") {
		t.Errorf("unfinished: status %d, out %q", status, out)
	}
}

// TestCommands drives the commands through the plain reader.
func TestCommands(t *testing.T) {
	dir := t.TempDir()
	file := filepath.Join(dir, "lib.ant")
	os.WriteFile(file, []byte("var loaded = 41\nloaded + 1\n"), 0o600)
	input := strings.Join([]string{
		":help",
		"var p = { name: \"x\" }",
		"type Point = { x: int, y: int }",
		":type Point{ x: 1, y: 2 }.x",
		":type nope",
		":doc Array.map",
		":load " + file,
		"loaded",
		":reset",
		"loaded",
		":nope",
		":type",
		":help x",
		":  type 1 + 1",
		":quit",
		"println(\"after quit\")",
	}, "\n") + "\n"
	var out, errOut strings.Builder
	status := Run(context.Background(), strings.NewReader(input), &out, &errOut)
	if status != 0 {
		t.Errorf("status %d", status)
	}
	o, e := out.String(), errOut.String()
	for _, want := range []string{":type expr", "Ctrl-C", "> int\n", "func Array.map", "> 42\n", "> 41\n"} {
		if !strings.Contains(o, want) {
			t.Errorf("stdout lacks %q:\n%s", want, o)
		}
	}
	if strings.Count(o, "Commands:") != 1 || strings.Count(o, "> int\n") != 2 {
		t.Errorf("stdout:\n%s", o)
	}
	for _, want := range []string{"nope", "unknown command :nope", "usage: :type expr", "usage: :help\n", "loaded"} {
		if !strings.Contains(e, want) {
			t.Errorf("stderr lacks %q:\n%s", want, e)
		}
	}
	if strings.Contains(o, "after quit") {
		t.Errorf("ran after :quit:\n%s", o)
	}
}

// TestCompletePaths checks that an import's path completes as a file or a folder, whole, and
// is not taken for a qualified name.
func TestCompletePaths(t *testing.T) {
	complete := func(src string, off int) []ant.Completion {
		return []ant.Completion{
			{Label: "m.ant", Insert: "m.ant", Kind: "file"},
			{Label: "../", Insert: "../", Kind: "folder"},
		}
	}
	e, _, in := fakeTerminal(t, 80, complete)
	typeKeys(in, "import \"m\t\r")
	if line, _ := e.readLine("> ", ""); line != "import \"m.ant" {
		t.Errorf("line %q", line)
	}
}
