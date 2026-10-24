package tests

import (
	"context"
	"regexp"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/ant-src/ant"
)

// TestPanicReport checks the format of a panic report: "panic: <message>" and then one "at"
// line per frame, innermost first, with the function and its position; a goroutine's report
// has the same shape, and the exit status is 1 only for the main goroutine.
func TestPanicReport(t *testing.T) {
	src := `func inner(n: int): int {
  if n == 0 { panic("boom") }
  inner(n - 1)
}

func outer(): int { inner(2) }

func main(): void {
  var wg = WaitGroup.new()
  WaitGroup.add(wg, 1)
  go {
    defer { WaitGroup.done(wg) }
    panic("in a goroutine")
  }
  WaitGroup.wait(wg)
  println(outer())
}
`
	prog, err := ant.Load("report.ant", ant.WithSource("report.ant", src), ant.WithFS(fstest.MapFS{}))
	if err != nil {
		t.Fatal(err)
	}
	var out, errOut lockedBuffer
	err = prog.Run(context.Background(), ant.WithStdout(&out), ant.WithStderr(&errOut))
	pe, ok := err.(*ant.PanicError)
	if !ok {
		t.Fatalf("expected a PanicError, got %v", err)
	}
	if pe.Message != "boom" || ant.ExitStatus(err) != 1 {
		t.Errorf("message %q, status %d", pe.Message, ant.ExitStatus(err))
	}
	if len(pe.Trace) < 5 || pe.Trace[0].Function != "inner" || pe.Trace[3].Function != "outer" || pe.Trace[4].Function != "main" {
		t.Errorf("trace: %+v", pe.Trace)
	}
	if pe.Trace[0].Pos.Path != "report.ant" || pe.Trace[0].Pos.Line != 2 {
		t.Errorf("innermost frame: %+v", pe.Trace[0])
	}
	got := errOut.String()
	mainReport := regexp.MustCompile(`(?m)^panic: boom\n  at inner \(report\.ant:2:\d+\)\n  at inner \(report\.ant:3:\d+\)\n  at inner \(report\.ant:3:\d+\)\n  at outer \(report\.ant:6:\d+\)\n  at main \(report\.ant:16:\d+\)\n`)
	if !mainReport.MatchString(got) {
		t.Errorf("the main goroutine's report does not match:\n%s", got)
	}
	goReport := regexp.MustCompile(`(?m)^panic: in a goroutine\n  at go block \(report\.ant:13:\d+\)\n`)
	if !goReport.MatchString(got) {
		t.Errorf("the goroutine's report does not match:\n%s", got)
	}
	if strings.Count(got, "panic:") != 2 {
		t.Errorf("expected exactly two reports:\n%s", got)
	}
}
