package ant_test

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/ant-src/ant"
	"github.com/ant-src/ant/value"
)

func loadSrc(t *testing.T, src string, opts ...ant.Option) *ant.Program {
	t.Helper()
	opts = append([]ant.Option{ant.WithSource("main.ant", src), ant.WithFS(fstest.MapFS{})}, opts...)
	prog, err := ant.Load("main.ant", opts...)
	if err != nil {
		t.Fatal(err)
	}
	return prog
}

func TestWithEnv(t *testing.T) {
	prog := loadSrc(t, `println(Os.env("GREETING")); println(Os.env("MISSING"))`)
	var out strings.Builder
	lookup := func(name string) (string, bool) {
		if name == "GREETING" {
			return "hi", true
		}
		return "", false
	}
	if err := prog.Run(context.Background(), ant.WithStdout(&out), ant.WithEnv(lookup)); err != nil {
		t.Fatal(err)
	}
	if out.String() != "hi\nnull\n" {
		t.Errorf("stdout %q", out.String())
	}
}

func TestInstanceValues(t *testing.T) {
	prog := loadSrc(t, `
type Point = { x: int }
var total = 0
func add(n: int): int { total += n; total }
func apply(f: func(int): int, n: int): int { f(n) }
func fails(): void { panic("nope") }
func leaves(): void { Os.exit(4) }
`, ant.WithStderr(&strings.Builder{}))
	in, err := prog.Start(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	add, ok := in.Func("add")
	if !ok {
		t.Fatal("add not found")
	}
	if v, err := add.Call(int64(5)); err != nil || v != int64(5) {
		t.Errorf("add(5) = %v, %v", v, err)
	}
	apply, _ := in.Func("apply")
	if v, err := apply.Call(add.Value(), int64(2)); err != nil || v != int64(7) {
		t.Errorf("apply(add, 2) = %v, %v", v, err)
	}
	if total, ok := in.Value("total"); !ok || total != int64(7) {
		t.Errorf("total = %v, %v", total, ok)
	}
	if tag, ok := in.Value("Point"); !ok || value.KindOf(tag) != value.KindTag {
		t.Errorf("Point = %v, %v", tag, ok)
	}
	if _, ok := in.Func("total"); ok {
		t.Error("a var is not a Func")
	}
	if _, ok := in.Func("nothing"); ok {
		t.Error("a missing name is not a Func")
	}
	fails, _ := in.Func("fails")
	if _, err := fails.Call(); err == nil {
		t.Error("expected a panic")
	} else {
		var pe *ant.PanicError
		if !errors.As(err, &pe) || pe.Message != "nope" || len(pe.Trace) == 0 || pe.Error() != "panic: nope" {
			t.Errorf("panic error: %#v", err)
		}
	}
	leaves, _ := in.Func("leaves")
	if _, err := leaves.Call(); err == nil {
		t.Error("expected an exit")
	} else {
		var ee *ant.ExitError
		if !errors.As(err, &ee) || ee.Status != 4 || ee.Error() != "exit status 4" || ant.ExitStatus(err) != 4 {
			t.Errorf("exit error: %#v", err)
		}
	}
}

func TestWithoutNamespaces(t *testing.T) {
	prog := loadSrc(t, `println(File.exists("x"))`, ant.WithoutNamespaces("File"))
	if diags := prog.Check(); len(diags) == 0 || !strings.Contains(diags[0].Message, "File") {
		t.Errorf("check: %v", diags)
	}
	var errOut strings.Builder
	err := prog.Run(context.Background(), ant.WithStdout(&strings.Builder{}), ant.WithStderr(&errOut))
	if _, ok := err.(*ant.PanicError); !ok {
		t.Errorf("run: %v", err)
	}
}

func TestErrorsAndStatuses(t *testing.T) {
	_, err := ant.Load("main.ant", ant.WithSource("main.ant", "var x ="), ant.WithFS(fstest.MapFS{}))
	var le *ant.LoadError
	if !errors.As(err, &le) || !strings.Contains(le.Error(), "main.ant:1:") || ant.ExitStatus(err) != 2 {
		t.Errorf("load error: %v", err)
	}
	if ant.ExitStatus(nil) != 0 || ant.ExitStatus(errors.New("other")) != 1 {
		t.Error("ExitStatus of nil or an unknown error")
	}
	d := ant.Diagnostic{Message: "bare"}
	if d.Error() != "bare" {
		t.Errorf("bare diagnostic: %q", d.Error())
	}
	d.Pos = ant.Position{Path: "p.ant", Line: 2, Col: 3}
	if d.Error() != "p.ant:2:3: bare" {
		t.Errorf("positioned diagnostic: %q", d.Error())
	}
	if _, err := ant.Load("/nonexistent/program.ant"); err == nil {
		t.Error("loading a missing file")
	}
}

func TestSessionValue(t *testing.T) {
	s, err := ant.NewSession(context.Background(), ant.WithStdout(&strings.Builder{}))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if _, err := s.Eval("var answer = 42"); err != nil {
		t.Fatal(err)
	}
	if v, ok := s.Value("answer"); !ok || v != int64(42) {
		t.Errorf("answer = %v, %v", v, ok)
	}
	if _, ok := s.Value("nothing"); ok {
		t.Error("nothing should be unbound")
	}
	if _, err := s.Eval("var y = "); err == nil {
		t.Error("expected a load error")
	} else if _, ok := err.(*ant.LoadError); !ok {
		t.Errorf("error type %T", err)
	}
	if !ant.Incomplete("func f() {") || ant.Incomplete("1 +") || ant.Incomplete("1") {
		t.Error("Incomplete")
	}
}

// TestNamespaceDefaults: a host function declared with defaults may be called with fewer
// arguments or by name, and the omitted ones arrive as nil.
func TestNamespaceDefaults(t *testing.T) {
	greet := ant.Namespace{
		Name:       "Greet",
		Signatures: `func Greet.hello(name: string, greeting: string = "hi", punct: string = "!"): string`,
		Funcs: map[string]ant.HostFunc{
			"hello": func(rt ant.Runtime, args []value.Value) (value.Value, error) {
				greeting, punct := "hi", "!"
				if args[1] != nil {
					greeting = args[1].(string)
				}
				if args[2] != nil {
					punct = args[2].(string)
				}
				return greeting + " " + args[0].(string) + punct, nil
			},
		},
	}
	prog := loadSrc(t, `println(Greet.hello("ada"))
println(Greet.hello("ada", "yo"))
println(Greet.hello("ada", punct: "?"))`, ant.WithNamespace(greet))
	if diags := prog.Check(); len(diags) != 0 {
		t.Errorf("check: %v", diags)
	}
	var out strings.Builder
	if err := prog.Run(context.Background(), ant.WithStdout(&out)); err != nil {
		t.Fatal(err)
	}
	if out.String() != "hi ada!\nyo ada!\nhi ada?\n" {
		t.Errorf("output %q", out.String())
	}
}

// TestWithStdin: Os.stdin reads what the option gives, and a read blocked on input that never
// comes ends when the run is cancelled.
func TestWithStdin(t *testing.T) {
	prog := loadSrc(t, `println(File.readAll(Os.stdin()))`)
	var out strings.Builder
	if err := prog.Run(context.Background(), ant.WithStdin(strings.NewReader("a\nb\n")), ant.WithStdout(&out)); err != nil {
		t.Fatal(err)
	}
	if out.String() != "Ok{ value: \"a\\nb\\n\" }\n" {
		t.Errorf("output %q", out.String())
	}
	pr, pw := io.Pipe()
	defer pw.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	start := time.Now()
	err := prog.Run(ctx, ant.WithStdin(pr), ant.WithStdout(&strings.Builder{}))
	if !errors.Is(err, context.DeadlineExceeded) || time.Since(start) > 5*time.Second {
		t.Errorf("blocked read: %v after %v", err, time.Since(start))
	}
}

// TestExecCancelled: cancelling the run kills the child, and Run returns promptly.
func TestExecCancelled(t *testing.T) {
	prog := loadSrc(t, `println(Os.exec("sleep", ["30"]))`)
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	start := time.Now()
	err := prog.Run(ctx, ant.WithStdout(&strings.Builder{}))
	if !errors.Is(err, context.DeadlineExceeded) || time.Since(start) > 5*time.Second {
		t.Errorf("cancelled exec: %v after %v", err, time.Since(start))
	}
}

// TestHostTimeNamespace: a host namespace named Time declares sync types, not time types, so
// the checker rejects is on them as it does for the prelude's sync types.
func TestHostTimeNamespace(t *testing.T) {
	ns := ant.Namespace{
		Name:       "Time",
		Signatures: "type Time.Stamp\nfunc Time.stamp(): Time.Stamp | null",
		Funcs: map[string]ant.HostFunc{
			"stamp": func(rt ant.Runtime, args []value.Value) (value.Value, error) { return value.Null, nil },
		},
	}
	prog := loadSrc(t, `var s = Time.stamp(); println(s is Time.Stamp)`, ant.WithNamespace(ns), ant.WithoutNamespaces("Time"))
	diags := prog.Check()
	if len(diags) != 1 || !strings.Contains(diags[0].Message, "is cannot test Time.Stamp") {
		t.Errorf("check: %v", diags)
	}
}

// TestWithEnvironment: the whole environment, for Os.env, Os.environ and a child of Os.exec.
func TestWithEnvironment(t *testing.T) {
	prog := loadSrc(t, `println(Os.env("ANT_ONLY"))
println(Os.env("PATH"))
println(Map.has(Os.environ(), "ANT_ONLY") && !Map.has(Os.environ(), "PATH"))
println(Os.exec("/usr/bin/env") | Result.map({ x.stdout }))`)
	var out strings.Builder
	err := prog.Run(context.Background(), ant.WithStdout(&out), ant.WithEnvironment(map[string]string{"ANT_ONLY": "yes"}))
	if err != nil {
		t.Fatal(err)
	}
	if want := "yes\nnull\ntrue\nOk{ value: \"ANT_ONLY=yes\\n\" }\n"; out.String() != want {
		t.Errorf("output %q, want %q", out.String(), want)
	}
}

// TestStartOptions: a started instance takes the same standard input and environment as a run.
func TestStartOptions(t *testing.T) {
	prog := loadSrc(t, `
func input(): string { Result.unwrap(File.readAll(Os.stdin())) }
func vars(): bool { Map.has(Os.environ(), "ANT_ONLY") && !Map.has(Os.environ(), "PATH") }
`)
	in, err := prog.Start(context.Background(), ant.WithStdin(strings.NewReader("typed")),
		ant.WithEnvironment(map[string]string{"ANT_ONLY": "yes"}))
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	for name, want := range map[string]value.Value{"input": "typed", "vars": true} {
		f, _ := in.Func(name)
		if got, err := f.Call(); err != nil || got != want {
			t.Errorf("%s() = %v, %v; want %v", name, got, err, want)
		}
	}
}

// TestStartContext checks that the context given to Start ends a call that never waits, as a
// session's context ends such an input.
func TestStartContext(t *testing.T) {
	src := "func spin(): int {\n  var n = 0\n  while true { n += 1 }\n  n\n}\n"
	prog, err := ant.Load("spin.ant", ant.WithSource("spin.ant", src))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	in, err := prog.Start(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	spin, _ := in.Func("spin")
	if _, err := spin.Call(); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("an endless call ended with %v, want the context's error", err)
	}
}

// TestCallSite: a panic of the call itself, such as a missing argument, names the function's
// position instead of 0:0.
func TestCallSite(t *testing.T) {
	prog, err := ant.Load("id.ant", ant.WithSource("id.ant", "func id(x: int): int { x }\n"))
	if err != nil {
		t.Fatal(err)
	}
	in, err := prog.Start(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	id, _ := in.Func("id")
	_, err = id.Call()
	var p *ant.PanicError
	if !errors.As(err, &p) || len(p.Trace) == 0 {
		t.Fatalf("id() ended with %v, want a panic with a trace", err)
	}
	if f := p.Trace[0]; f.Pos.Line != 1 {
		t.Errorf("the call's frame is at %d:%d, want line 1", f.Pos.Line, f.Pos.Col)
	}
}

// TestStartInterrupts: an instance honors WithInterrupts as a run does. An interrupt the
// program never asked for ends the instance, and the call in progress returns ErrInterrupted.
func TestStartInterrupts(t *testing.T) {
	src := "func spin(): int {\n  var n = 0\n  while true { n += 1 }\n  n\n}\n"
	interrupts := make(chan struct{}, 1)
	// The deadline only bounds the test: without the interrupt the call would never end.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	in, err := loadSrc(t, src).Start(ctx, ant.WithInterrupts(interrupts))
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	time.AfterFunc(50*time.Millisecond, func() { interrupts <- struct{}{} })
	spin, _ := in.Func("spin")
	if _, err := spin.Call(); !errors.Is(err, ant.ErrInterrupted) {
		t.Fatalf("an interrupted call ended with %v, want ErrInterrupted", err)
	}
}

// TestHostTextIsDecoded checks that a string a host hands to a program, in a host function's
// result or in a call's arguments, holds only scalar values, as outside text does
// (section 11, String): each ill-formed sequence becomes U+FFFD.
func TestHostTextIsDecoded(t *testing.T) {
	db := ant.Namespace{
		Name:       "Db",
		Signatures: "func Db.row(): Map[string, Array[string]]\n",
		Funcs: map[string]ant.HostFunc{
			"row": func(rt ant.Runtime, args []value.Value) (value.Value, error) {
				m := value.NewMap()
				m.Set("caf\xe9", value.NewArray([]value.Value{"ok", "\xff\xfe"}))
				return m, nil
			},
		},
	}
	prog := loadSrc(t, `
func bytes(s: string): int { len(String.toBytes(s)) }
func show(s: string): string { String.from([s, bytes(s)]) }
for { key, value } in Db.row() {
  println(show(key))
  for v in value { println(show(v)) }
}
`, ant.WithNamespace(db))
	var out strings.Builder
	if err := prog.Run(context.Background(), ant.WithStdout(&out)); err != nil {
		t.Fatal(err)
	}
	want := "[\"caf�\", 6]\n[\"ok\", 2]\n[\"��\", 6]\n"
	if out.String() != want {
		t.Errorf("stdout %q, want %q", out.String(), want)
	}

	in, err := loadSrc(t, `
func bytes(s: string): int { len(String.toBytes(s)) }
`).Start(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	f, _ := in.Func("bytes")
	if v, err := f.Call("a\xffb"); err != nil || v != int64(5) {
		t.Errorf("bytes(\"a\\xffb\") = %v, %v; want 5", v, err)
	}
}

// TestHostBytesAreKept: Bytes from a host, in a result or an argument, reach the program as they
// are, since only strings are outside text.
func TestHostBytesAreKept(t *testing.T) {
	blob := ant.Namespace{
		Name:       "Blob",
		Signatures: "func Blob.raw(): Array[Bytes]\n",
		Funcs: map[string]ant.HostFunc{
			"raw": func(rt ant.Runtime, args []value.Value) (value.Value, error) {
				return value.NewArray([]value.Value{value.Bytes("\xff\xfe"), "\xff"}), nil
			},
		},
	}
	prog := loadSrc(t, `
println(Blob.raw())
func size(b: Bytes): int { len(b) }
`, ant.WithNamespace(blob))
	var out strings.Builder
	if err := prog.Run(context.Background(), ant.WithStdout(&out)); err != nil {
		t.Fatal(err)
	}
	if want := "[Bytes(ff fe), \"�\"]\n"; out.String() != want {
		t.Errorf("stdout %q, want %q", out.String(), want)
	}
	in, err := prog.Start(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	f, _ := in.Func("size")
	if v, err := f.Call(value.Bytes("\xff\xfe\xfd")); err != nil || v != int64(3) {
		t.Errorf("size(Bytes) = %v, %v; want 3", v, err)
	}
}

// TestWithoutBytesNamespace: removing the Bytes namespace removes its functions and keeps the
// type, which is built in: is, len, for, b[i] and annotations still check and run.
func TestWithoutBytesNamespace(t *testing.T) {
	prog := loadSrc(t, `var b: Bytes = String.toBytes("ab")
var u: string | Bytes = b
var total = 0
for x in b { total = total + int(x) }
println([u is Bytes, len(b) == 2, int(b[1]) == 98, total == 195])`, ant.WithoutNamespaces("Bytes"))
	if diags := prog.Check(); len(diags) != 0 {
		t.Fatalf("check: %v", diags)
	}
	var out strings.Builder
	if err := prog.Run(context.Background(), ant.WithStdout(&out)); err != nil {
		t.Fatal(err)
	}
	if out.String() != "[true, true, true, true]\n" {
		t.Errorf("stdout %q", out.String())
	}
	gone := loadSrc(t, `println(Bytes.length(String.toBytes("a")))`, ant.WithoutNamespaces("Bytes"))
	if diags := gone.Check(); len(diags) != 1 || !strings.Contains(diags[0].Message, "length") {
		t.Errorf("check: %v", diags)
	}
}

// TestArityPanicFrames: a call given the wrong arguments panics at the call site, in the caller,
// with no frame of the callee it never entered.
func TestArityPanicFrames(t *testing.T) {
	src := "func f(a: int): int { a }\n\nprintln(f(1, 2))\n"
	prog, err := ant.Load("arity.ant", ant.WithSource("arity.ant", src))
	if err != nil {
		t.Fatal(err)
	}
	err = prog.Run(context.Background(), ant.WithStdout(io.Discard), ant.WithStderr(io.Discard))
	var p *ant.PanicError
	if !errors.As(err, &p) {
		t.Fatalf("the run ended with %v, want a panic", err)
	}
	if len(p.Trace) != 1 || p.Trace[0].Function != "module arity.ant" || p.Trace[0].Pos.Line != 3 || p.Trace[0].Pos.Col != 10 {
		t.Errorf("trace %+v, want the one frame of the script at the call, 3:10", p.Trace)
	}
}
