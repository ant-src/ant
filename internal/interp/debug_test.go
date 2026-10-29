package interp_test

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/internal/interp"
	"github.com/ant-src/ant/internal/lib"
	"github.com/ant-src/ant/internal/loader"
	"github.com/ant-src/ant/parser"
	"github.com/ant-src/ant/value"
)

// debugRun is a program running under a debugger, driven by a test.
type debugRun struct {
	t     *testing.T
	d     *interp.Debugger
	prog  *loader.Program
	stops chan interp.Stop
	done  chan error
	out   *bytes.Buffer
	stop  interp.Stop
}

// debug loads main.ant from src, lets setup configure the debugger, and starts the run.
func debug(t *testing.T, src string, setup func(r *debugRun)) *debugRun {
	t.Helper()
	fsys := fstest.MapFS{"main.ant": {Data: []byte(src)}}
	prog, diags := loader.Load("main.ant", loader.Config{FS: fsys, Mode: parser.Program})
	if len(diags) > 0 {
		t.Fatalf("load: %v", diags)
	}
	r := &debugRun{t: t, d: interp.NewDebugger(), prog: prog, stops: make(chan interp.Stop, 1), done: make(chan error, 1), out: &bytes.Buffer{}}
	r.d.OnStop = func(s interp.Stop) { r.stops <- s }
	if setup != nil {
		setup(r)
	}
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	go func() {
		r.done <- interp.Run(ctx, prog, interp.Config{Prelude: lib.Prelude(), Stdout: r.out, Stderr: r.out, Debugger: r.d})
	}()
	return r
}

// breakAt sets breakpoints on the first stop point of each line.
func (r *debugRun) breakAt(lines ...int) {
	var bps []interp.Breakpoint
	for _, l := range lines {
		bps = append(bps, interp.Breakpoint{Pos: r.pointOn(l)})
	}
	r.set(bps...)
}

// set sets breakpoints in main.ant and fails on any that is not set.
func (r *debugRun) set(bps ...interp.Breakpoint) {
	r.t.Helper()
	for _, err := range r.d.SetBreakpoints(r.prog.Root.File.Path, bps) {
		if err != nil {
			r.t.Fatal(err)
		}
	}
}

// pointOn is the first stop point on line.
func (r *debugRun) pointOn(line int) ast.Pos {
	r.t.Helper()
	for _, sp := range interp.StopPoints(r.prog.Root.File) {
		if sp.Pos.Line == line {
			return sp.Pos
		}
	}
	r.t.Fatalf("no stop point on line %d", line)
	return ast.Pos{}
}

// wait waits for the next stop and checks its reason and line.
func (r *debugRun) wait(reason string, line int) {
	r.t.Helper()
	select {
	case r.stop = <-r.stops:
	case err := <-r.done:
		r.t.Fatalf("the run ended (%v) before a stop for %s at line %d; output:\n%s", err, reason, line, r.out)
	case <-time.After(5 * time.Second):
		r.t.Fatalf("no stop for %s at line %d", reason, line)
	}
	if r.stop.Reason != reason {
		r.t.Fatalf("stopped for %s, want %s", r.stop.Reason, reason)
	}
	if got := r.frames(r.stop.Goroutine)[0].Pos.Line; got != line {
		r.t.Fatalf("stopped at line %d, want %d", got, line)
	}
}

func (r *debugRun) frames(id int) []interp.StackFrame {
	r.t.Helper()
	fs, ok := r.d.Frames(id)
	if !ok || len(fs) == 0 {
		r.t.Fatalf("no frames for goroutine %d", id)
	}
	return fs
}

// locals renders the innermost frame's Locals as name=value pairs.
func (r *debugRun) locals() string {
	locals, _, _ := r.d.Scopes(r.frames(r.stop.Goroutine)[0])
	return render(locals)
}

func render(vs []interp.Var) string {
	var parts []string
	for _, v := range vs {
		parts = append(parts, v.Name+"="+value.RenderQuoted(v.Value))
	}
	return strings.Join(parts, " ")
}

func (r *debugRun) finish() {
	r.t.Helper()
	r.d.Continue()
	select {
	case err := <-r.done:
		if err != nil {
			r.t.Fatalf("run: %v\n%s", err, r.out)
		}
	case s := <-r.stops:
		r.t.Fatalf("unexpected stop: %+v", s)
	case <-time.After(5 * time.Second):
		r.t.Fatal("the run did not end")
	}
}

func TestDebugBreakpointAndLocals(t *testing.T) {
	r := debug(t, `func add(a: int, b: int): int {
  var sum = a + b
  return sum
}

var total = add(1, 2)
var later = 4
println(total)
`, func(r *debugRun) { r.breakAt(3) })
	r.wait("breakpoint", 3)
	if got, want := r.locals(), "a=1 b=2 sum=3"; got != want {
		t.Errorf("locals: %s, want %s", got, want)
	}
	fs := r.frames(r.stop.Goroutine)
	if len(fs) != 2 || fs[0].Function != "add" || fs[1].Pos.Line != 6 {
		t.Errorf("frames: %+v", fs)
	}
	_, _, module := r.d.Scopes(fs[1])
	if got := render(module); got != "" {
		t.Errorf("module vars before their initializers ran: %s", got)
	}
	r.finish()
	if r.out.String() != "3\n" {
		t.Errorf("output %q", r.out)
	}
}

func TestDebugStepping(t *testing.T) {
	r := debug(t, `func inc(n: int): int {
  var m = n + 1
  return m
}

var a = inc(1)
var b = inc(a)
println(b)
`, func(r *debugRun) { r.breakAt(6) })
	r.wait("breakpoint", 6)
	r.d.Step(r.stop.Goroutine, interp.StepIn)
	r.wait("step", 2)
	r.d.Step(r.stop.Goroutine, interp.StepOver)
	r.wait("step", 3)
	r.d.Step(r.stop.Goroutine, interp.StepOut)
	r.wait("step", 7)
	r.d.Step(r.stop.Goroutine, interp.StepOver)
	r.wait("step", 8)
	r.finish()
}

func TestDebugPipelineStages(t *testing.T) {
	r := debug(t, `var xs = [1, 2, 3, 4]
var total = xs
  | Array.filter({ x > 1 })
  | Array.map({ x * 10 })
  | Array.sum()
println(total)
`, func(r *debugRun) { r.breakAt(2) })
	r.wait("breakpoint", 2)
	r.d.Step(r.stop.Goroutine, interp.StepOver)
	r.wait("step", 3)
	if got, want := r.locals(), "(piped)=[1, 2, 3, 4]"; got != want {
		t.Errorf("at the filter: %s, want %s", got, want)
	}
	r.d.Step(r.stop.Goroutine, interp.StepOver)
	r.wait("step", 4)
	if got, want := r.locals(), "(piped)=[2, 3, 4]"; got != want {
		t.Errorf("at the map: %s, want %s", got, want)
	}
	r.d.Step(r.stop.Goroutine, interp.StepIn)
	r.wait("step", 4) // the lambda's body
	if got, want := r.locals(), "x=2"; got != want {
		t.Errorf("in the lambda: %s, want %s", got, want)
	}
	r.d.Step(r.stop.Goroutine, interp.StepOut)
	r.wait("step", 5)
	r.d.Step(r.stop.Goroutine, interp.StepOver)
	r.wait("step", 6)
	r.finish()
}

func TestDebugSequenceLoopStepsInItsFrame(t *testing.T) {
	r := debug(t, `func count(yield: func(int): bool) {
  for i in [1, 2] {
    if !yield(i) { return }
  }
}

var seen = 0
for n in count {
  seen += n
}
println(seen)
`, func(r *debugRun) { r.breakAt(9) })
	r.wait("breakpoint", 9)
	r.d.Step(r.stop.Goroutine, interp.StepOver)
	r.wait("step", 9) // the next iteration, not the sequence's statements
	r.d.Step(r.stop.Goroutine, interp.StepOver)
	r.wait("step", 11)
	r.finish()
}

func TestDebugPanics(t *testing.T) {
	src := `func boom() { panic("boom") }

var caught = Panic.catch({ boom() })
println("after catch")
boom()
`
	r := debug(t, src, nil)
	r.wait("exception", 1)
	if fs := r.frames(r.stop.Goroutine); len(fs) != 2 || fs[1].Pos.Line != 5 {
		t.Errorf("frames: %+v", fs)
	}
	if p := r.d.Panic(); p == nil || p.Message != "boom" {
		t.Errorf("panic: %+v", p)
	}
	if !strings.Contains(r.out.String(), "after catch") {
		t.Errorf("a caught panic stopped the run under uncaught: %q", r.out)
	}
	r.d.Continue()
	if err := <-r.done; err == nil {
		t.Error("the run did not end with the panic")
	}

	r = debug(t, src, func(r *debugRun) { r.d.SetPanicFilters(false, true) })
	r.wait("exception", 1)
	if fs := r.frames(r.stop.Goroutine); fs[len(fs)-1].Pos.Line != 3 {
		t.Errorf("the caught panic should stop first: %+v", fs)
	}
	r.d.Continue()
	r.wait("exception", 1)
	r.d.Continue()
	<-r.done
}

func TestDebugEntryAndPauseWithBlockedGoroutine(t *testing.T) {
	r := debug(t, `var ch = Channel.new()
go {
  Channel.receive(ch)
}
var i = 0
while true {
  i += 1
}
`, func(r *debugRun) { r.d.StopOnEntry = true })
	r.wait("entry", 1)
	r.d.Continue()
	time.Sleep(50 * time.Millisecond)
	r.d.Pause(1)
	select {
	case r.stop = <-r.stops:
	case <-time.After(5 * time.Second):
		t.Fatal("pause did not stop the run")
	}
	if r.stop.Reason != "pause" {
		t.Fatalf("reason %s", r.stop.Reason)
	}
	gs := r.d.Goroutines()
	if len(gs) != 2 || gs[0].Name != "main" || gs[1].Name != fmt.Sprintf("go at main.ant:2 #%d", gs[1].ID) {
		t.Fatalf("goroutines: %+v", gs)
	}
	if fs := r.frames(gs[1].ID); fs[0].Pos.Line != 3 {
		t.Errorf("the blocked goroutine's frames: %+v", fs)
	}
}

func TestDebugEvaluate(t *testing.T) {
	r := debug(t, `type User = { name: string, tags: Array[string] }

func show(u: User, n: int): string {
  var label = u.name + "#" + String.from(n)
  return label
}

var ch = Channel.new()
var u = User { name: "ada", tags: ["x", "y"] }
println(show(u, 2))
`, func(r *debugRun) { r.breakAt(5) })
	r.wait("breakpoint", 5)
	top := r.frames(r.stop.Goroutine)[0]
	for src, want := range map[string]string{
		"label":  `"ada#2"`,
		"n * 10": "20",
		"u.tags | Array.map({ x + String.from(n) })": `["x2", "y2"]`,
		"len(u.tags)": "2",
		"show(u, 3)":  `"ada#3"`,
	} {
		v, err := r.d.Evaluate(top, src)
		if err != nil {
			t.Errorf("%s: %v", src, err)
			continue
		}
		if got := value.RenderQuoted(v); got != want {
			t.Errorf("%s = %s, want %s", src, got, want)
		}
	}
	for src, want := range map[string]string{
		"nope + 1":            "nope is not defined here",
		"n = 3":               "cannot be assigned",
		"u.tags[9]":           "panic: index 9 out of range",
		"Channel.receive(ch)": "did not finish within a second",
		"n +":                 "",
	} {
		_, err := r.d.Evaluate(top, src)
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%s: error %v, want one with %q", src, err, want)
		}
	}
	// the caller's frame sees its own names
	if v, err := r.d.Evaluate(r.frames(r.stop.Goroutine)[1], "u.name"); err != nil || value.Render(v) != "ada" {
		t.Errorf("u.name in the caller: %v %v", v, err)
	}
	r.finish()
}

func TestDebugConditionsHitsAndLogpoints(t *testing.T) {
	var logs []string
	r := debug(t, `var total = 0
for i in [1, 2, 3, 4, 5, 6] {
  total += i
}
println(total)
`, func(r *debugRun) {
		r.d.OnLog = func(s string) { logs = append(logs, s) }
		r.set(interp.Breakpoint{Pos: r.pointOn(3), Condition: "i % 2 == 0", HitCondition: "2"})
	})
	r.wait("breakpoint", 3)
	if got := r.locals(); got != "i=4" {
		t.Errorf("the second even i: %s", got)
	}
	r.set(interp.Breakpoint{Pos: r.pointOn(3), LogMessage: "i={i} total={total} {{ok}}"})
	r.finish()
	want := []string{"i=5 total=10 {ok}\n", "i=6 total=15 {ok}\n"}
	if strings.Join(logs, "|") != strings.Join(want, "|") {
		t.Errorf("logs %q, want %q", logs, want)
	}
	for _, bad := range []interp.Breakpoint{{HitCondition: "often"}, {Condition: "i +"}, {LogMessage: "{i"}} {
		bad.Pos = r.pointOn(3)
		if errs := r.d.SetBreakpoints(r.prog.Root.File.Path, []interp.Breakpoint{bad}); errs[0] == nil {
			t.Errorf("%+v was set", bad)
		}
	}
}

func TestDebugConditionError(t *testing.T) {
	var logs []string
	r := debug(t, `var x = 1
println(x)
`, func(r *debugRun) {
		r.d.OnLog = func(s string) { logs = append(logs, s) }
		r.set(interp.Breakpoint{Pos: r.pointOn(2), Condition: "x + 1"})
	})
	r.wait("breakpoint", 2)
	if len(logs) != 1 || !strings.Contains(logs[0], "not a bool") {
		t.Errorf("logs: %q", logs)
	}
	r.finish()
}

func TestDebugSetLocals(t *testing.T) {
	r := debug(t, `func f(): int {
  var a = 1
  var b = 2
  var get = { b }
  return a + get()
}
println(f())
`, func(r *debugRun) { r.breakAt(5) })
	r.wait("breakpoint", 5)
	locals, _, _ := r.d.Scopes(r.frames(r.stop.Goroutine)[0])
	for _, v := range locals {
		switch v.Name {
		case "a":
			if err := r.d.SetVar(v, int64(10)); err != nil {
				t.Fatal(err)
			}
		case "b": // captured by the lambda, so a shared slot
			if err := r.d.SetVar(v, int64(20)); err != nil {
				t.Fatal(err)
			}
		}
	}
	r.finish()
	if r.out.String() != "30\n" {
		t.Errorf("output %q", r.out)
	}
}

func TestDebugEveryGoroutineReportsItsBreakpoint(t *testing.T) {
	for range 20 { // the goroutines reach the breakpoint together in some runs, one by one in others
		r := debug(t, `var ch = Channel.new(2)
func work(n: int) {
  Channel.send(ch, n)
}
go { work(1) }
go { work(2) }
var a = Channel.receive(ch)
var b = Channel.receive(ch)
println(a + b)
`, func(r *debugRun) { r.breakAt(3) })
		r.wait("breakpoint", 3)
		first := r.stop.Goroutine
		r.d.Continue()
		r.wait("breakpoint", 3)
		if r.stop.Goroutine == first {
			t.Fatalf("goroutine %d stopped twice", first)
		}
		r.finish()
	}
}
