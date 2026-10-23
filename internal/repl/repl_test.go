package repl

import (
	"context"
	"io"
	"path/filepath"
	"strings"
	"testing"
	"time"
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
		"print(\"side\")",
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
	status := Run(context.Background(), strings.NewReader("print(1)\nOs.exit(3)\nprint(2)\n"), &out, &errOut)
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
