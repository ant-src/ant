package ant_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"syscall"
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

// TestWithoutOs: removing Os removes Cmd with it, so a host that forbids the process forbids
// running programs too.
func TestWithoutOs(t *testing.T) {
	prog := loadSrc(t, `println(Cmd.read(Cmd.new("echo", ["ran"])))`, ant.WithoutNamespaces("Os"))
	if diags := prog.Check(); len(diags) == 0 || !strings.Contains(diags[0].Message, "Cmd") {
		t.Errorf("check: %v", diags)
	}
	var out strings.Builder
	err := prog.Run(context.Background(), ant.WithStdout(&out), ant.WithStderr(&strings.Builder{}))
	if _, ok := err.(*ant.PanicError); !ok || strings.Contains(out.String(), "ran") {
		t.Errorf("run: %v, printed %q", err, out.String())
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
	prog := loadSrc(t, `println(Cmd.capture(Cmd.new("sleep", ["30"])))`)
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

// TestWithEnvironment: the whole environment, for Os.env, Os.environ and a child of a Cmd.
func TestWithEnvironment(t *testing.T) {
	prog := loadSrc(t, `println(Os.env("ANT_ONLY"))
println(Os.env("PATH"))
println(Map.has(Os.environ(), "ANT_ONLY") && !Map.has(Os.environ(), "PATH"))
println(Cmd.capture(Cmd.new("/usr/bin/env")) | Result.map({ x.stdout }))`)
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
				return value.NewArray([]value.Value{value.CopyBytes([]byte("\xff\xfe")), "\xff"}), nil
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
	if v, err := f.Call(value.CopyBytes([]byte("\xff\xfe\xfd"))); err != nil || v != int64(3) {
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

// TestHostTypeOwner: a host namespace's types are qualified by the namespace, as in a declaration
// file, like its functions.
func TestHostTypeOwner(t *testing.T) {
	ns := ant.Namespace{Name: "Db", Signatures: "type Row = Map[string, string]\nfunc Db.row(): Db.Row", Funcs: map[string]ant.HostFunc{
		"row": func(rt ant.Runtime, args []value.Value) (value.Value, error) { return value.Null, nil },
	}}
	_, err := ant.Load("main.ant", ant.WithSource("main.ant", "println(1)\n"), ant.WithFS(fstest.MapFS{}), ant.WithNamespace(ns))
	if err == nil || !strings.Contains(err.Error(), "namespace Db: type Row must be declared as Db.Row") {
		t.Errorf("unqualified host type: %v", err)
	}
}

// TestRecordTags: Runtime.Tag and Program.Tag give the tag the program holds for a record type
// of the prelude, a host namespace's or the standard library's, so a record the host builds is
// of that type; a tag the host makes itself is of none.
func TestRecordTags(t *testing.T) {
	var sawPoint value.Value
	ns := ant.Namespace{
		Name: "Geo",
		Signatures: `
type Geo.Point = { x: int, y: int }
type Geo.Id = int
type Geo.Shape = Geo.Point | null
type Geo.Handle
func Geo.origin(): Geo.Shape
func Geo.forged(): Geo.Shape
func Geo.output(): Cmd.Output | null
func Geo.take(p: Geo.Point): void
`,
		Funcs: map[string]ant.HostFunc{
			"origin": func(rt ant.Runtime, args []value.Value) (value.Value, error) {
				tag, ok := rt.Tag("Geo.Point")
				if !ok {
					return nil, errors.New("no tag for Geo.Point")
				}
				return value.NewRecord(tag, []value.Value{int64(0), int64(0)}), nil
			},
			"forged": func(rt ant.Runtime, args []value.Value) (value.Value, error) {
				tag := value.NewTag("Geo.Point", value.RecordTag, []string{"x", "y"})
				return value.NewRecord(tag, []value.Value{int64(0), int64(0)}), nil
			},
			"output": func(rt ant.Runtime, args []value.Value) (value.Value, error) {
				tag, ok := rt.Tag("Cmd.Output")
				if !ok {
					return nil, errors.New("no tag for Cmd.Output")
				}
				return value.NewRecord(tag, []value.Value{int64(3), value.NewArray(nil), "out", "err"}), nil
			},
			"take": func(rt ant.Runtime, args []value.Value) (value.Value, error) {
				sawPoint = args[0]
				for _, name := range []string{"Geo.Id", "Geo.Shape", "Geo.Handle", "Geo.origin", "Geo.Missing", "Missing.Point", "Geo", "String", "println", "Geo.Point.x", ""} {
					if tag, ok := rt.Tag(name); ok {
						return nil, fmt.Errorf("Tag(%q) = %v, want none", name, tag.Name)
					}
				}
				for _, name := range []string{"Error", "Ok", "Cmd", "File.Info", "Panic.Frame"} {
					if tag, ok := rt.Tag(name); !ok || tag.Name != name {
						return nil, fmt.Errorf("Tag(%q) = %v, %v", name, tag, ok)
					}
				}
				return nil, nil
			},
		},
	}
	src := `
let p = Geo.origin()
println(p is Geo.Point)
println(Geo.forged() is Geo.Point)
println(Geo.output()?.status)
println(Geo.output() is Cmd.Output)
Geo.take(Geo.Point{ x: 1, y: 2 })
func same(p: Geo.Shape): bool { p is Geo.Point && p.x == 5 }
`
	prog := loadSrc(t, src, ant.WithNamespace(ns))
	if diags := prog.Check(); len(diags) != 0 {
		t.Fatalf("check: %v", diags)
	}
	var out strings.Builder
	in, err := prog.Start(context.Background(), ant.WithStdout(&out))
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	if out.String() != "true\nfalse\n3\ntrue\n" {
		t.Errorf("output %q", out.String())
	}
	// The tag a program gives is the one its records carry, in both directions.
	tag, ok := prog.Tag("Geo.Point")
	if !ok {
		t.Fatal("Program.Tag: no Geo.Point")
	}
	if r, ok := sawPoint.(*value.Record); !ok || r.Tag != tag {
		t.Errorf("the record Ant built has tag %v, the program gives %v", sawPoint, tag)
	}
	same, _ := in.Func("same")
	if v, err := same.Call(value.NewRecord(tag, []value.Value{int64(5), int64(0)})); err != nil || v != true {
		t.Errorf("same(Point from Program.Tag) = %v, %v", v, err)
	}
	// Each Load makes the namespace's tags anew.
	other := loadSrc(t, src, ant.WithNamespace(ns))
	if tag2, _ := other.Tag("Geo.Point"); tag2 == nil || tag2 == tag {
		t.Errorf("a second program's tag: %v, the first's: %v", tag2, tag)
	}
}

// TestDeclarationFileProgram: a declaration file loads, checks, and cannot run (section 10,
// Declaration files).
func TestDeclarationFileProgram(t *testing.T) {
	fsys := fstest.MapFS{"db.d.ant": {Data: []byte("type Db.Row = Map[string, string]\nfunc Db.query(sql: string): Array[Db.Row]\nfunc Db.count(): Count\n")}}
	prog, err := ant.Load("db.d.ant", ant.WithFS(fsys))
	if err != nil {
		t.Fatal(err)
	}
	diags := prog.Check()
	if len(diags) != 1 || diags[0].Error() != "db.d.ant:3:18: unknown type Count" {
		t.Errorf("check: %v", diags)
	}
	err = prog.Run(context.Background())
	var le *ant.LoadError
	if !errors.As(err, &le) || ant.ExitStatus(err) != 2 || err.Error() != "db.d.ant: a declaration file declares names for the prelude and cannot be run" {
		t.Errorf("run: %v", err)
	}
	if _, err := prog.Start(context.Background()); !errors.As(err, &le) {
		t.Errorf("start: %v", err)
	}
	// A name a host namespace binds is a prelude name to the file; one the host removed is free.
	ns := ant.Namespace{Name: "Db", Signatures: "func Db.row(): int", Funcs: map[string]ant.HostFunc{
		"row": func(rt ant.Runtime, args []value.Value) (value.Value, error) { return value.Null, nil },
	}}
	fsys = fstest.MapFS{"x.d.ant": {Data: []byte("func Db.query(sql: string): int\nlet Os.LIMIT: int\ntype Cmd.Line = string\n")}}
	prog, err = ant.Load("x.d.ant", ant.WithFS(fsys), ant.WithNamespace(ns), ant.WithoutNamespaces("Os"))
	if err != nil {
		t.Fatal(err)
	}
	if diags := prog.Check(); len(diags) != 1 || diags[0].Error() != "x.d.ant:1:6: Db is a prelude name and cannot be redeclared" {
		t.Errorf("check against a host namespace: %v", diags)
	}
	// Defaults are checked as a module's are.
	fsys = fstest.MapFS{"d.d.ant": {Data: []byte("type Db.Limits = { rows: int = \"many\" }\nfunc Db.log(to: File.Handle = Os.stdout(), rows: int = 1.5): void\n")}}
	prog, err = ant.Load("d.d.ant", ant.WithFS(fsys))
	if err != nil {
		t.Fatal(err)
	}
	if diags := prog.Check(); len(diags) != 2 || diags[0].Error() != "d.d.ant:1:32: cannot use string where int is expected" ||
		diags[1].Error() != "d.d.ant:2:56: cannot use float64 where int is expected" {
		t.Errorf("defaults of a declaration file: %v", diags)
	}
}

// TestHostSignatureLimits: a host namespace's signatures take no field defaults and no
// constants, which only the standard library fills.
func TestHostSignatureLimits(t *testing.T) {
	row := map[string]ant.HostFunc{"run": func(rt ant.Runtime, args []value.Value) (value.Value, error) { return value.Null, nil }}
	for sigs, want := range map[string]string{
		"type Db.Opts = { limit: int = 1 }\nfunc Db.run(o: Db.Opts): string": "namespace Db: type Db.Opts gives a field a default, which a host namespace cannot fill",
		"let Db.VERSION: string\nfunc Db.run(): string":                      "namespace Db: a host namespace declares no constants",
	} {
		ns := ant.Namespace{Name: "Db", Signatures: sigs, Funcs: row}
		_, err := ant.Load("main.ant", ant.WithSource("main.ant", "println(1)\n"), ant.WithFS(fstest.MapFS{}), ant.WithNamespace(ns))
		if err == nil || !strings.Contains(err.Error(), want) {
			t.Errorf("%q: %v", sigs, err)
		}
	}
}

// TestStreamsInProgramOrder: neither stream is buffered, so where the two reach one writer they
// interleave in program order (stdlib.md, Os, Streams).
func TestStreamsInProgramOrder(t *testing.T) {
	prog := loadSrc(t, `print("a"); eprint("b"); println("c"); File.writeText(Os.stderr(), "d"); eprintln(); println("e")`)
	var both strings.Builder
	if err := prog.Run(context.Background(), ant.WithStdout(&both), ant.WithStderr(&both)); err != nil {
		t.Fatal(err)
	}
	if both.String() != "abc\nd\ne\n" {
		t.Errorf("got %q", both.String())
	}
}

// brokenPipe is an output whose reader is gone, as a host's writer or a pipe on Windows gives
// it; on Unix the Go runtime kills the process instead.
type brokenPipe struct{}

func (brokenPipe) Write([]byte) (int, error) { return 0, syscall.EPIPE }

// TestBrokenPipeEndsRun: print and eprint ignore a failed write, but a broken pipe ends the run
// with status 141, as SIGPIPE ends a shell's process, so a print loop does not spin.
func TestBrokenPipeEndsRun(t *testing.T) {
	for _, src := range []string{`while !false { println("y") }`, `while !false { eprintln("y") }`} {
		prog := loadSrc(t, src)
		err := prog.Run(context.Background(), ant.WithStdout(brokenPipe{}), ant.WithStderr(brokenPipe{}))
		var ee *ant.ExitError
		if !errors.As(err, &ee) || ee.Status != 141 {
			t.Errorf("%s: %v, want exit status 141", src, err)
		}
	}
}

// stuckWriter is an output nobody reads: a write blocks until release is closed.
type stuckWriter struct{ release chan struct{} }

func (w stuckWriter) Write(p []byte) (int, error) {
	<-w.release
	return len(p), nil
}

// TestStuckWriteDoesNotHoldRun: a goroutine blocked writing to a full pipe does not keep the
// run from ending once the script is done.
func TestStuckWriteDoesNotHoldRun(t *testing.T) {
	w := stuckWriter{make(chan struct{})}
	defer close(w.release)
	prog := loadSrc(t, `go { println("y") }
Time.sleep(Time.duration(milliseconds: 20))`)
	done := make(chan error, 1)
	go func() { done <- prog.Run(context.Background(), ant.WithStdout(w)) }()
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("the run waited behind the stuck write")
	}
}

// TestProgram: Os.program is the path given to Load.
func TestProgram(t *testing.T) {
	prog := loadSrc(t, `println(Os.program())`)
	var out strings.Builder
	if err := prog.Run(context.Background(), ant.WithStdout(&out)); err != nil || out.String() != "main.ant\n" {
		t.Errorf("got %q, %v", out.String(), err)
	}
}

// TestHostDeclarationFiles: a declaration file beside the root joins the program's prelude for
// the checker; the names a host provides are the host's; and a program that reads a name no
// host provides does not run (section 10, Declaration files).
func TestHostDeclarationFiles(t *testing.T) {
	decl := "func Tui.width(s: string): int\nfunc Tui.height(): int\ntype Tui.Key = { key: string }\n"
	fsys := fstest.MapFS{
		"app/tui.d.ant":     {Data: []byte(decl)},
		"app/main.ant":      {Data: []byte("import \"../lib/text.ant\" { wide }\nprintln(wide(\"ab\"))\nprintln(Tui.height())\n")},
		"app/plain.ant":     {Data: []byte("println(\"plain\")\n")},
		"app/wide_test.ant": {Data: []byte("test \"width\" { assert(Tui.width(\"ab\") == 2) }\n")},
		"app/prelude.d.ant": {Data: []byte("func Nope.nope(): int\n")},
		"lib/text.ant":      {Data: []byte("export func wide(s: string): int { Tui.width(s) }\n")},
		"lib/own.ant":       {Data: []byte("println(Tui.width(\"x\"))\n")},
		"bad/bad.d.ant":     {Data: []byte("func Bad.f(): Missing\nfunc File.extra(): int\n")},
		"bad/main.ant":      {Data: []byte("println(Bad.f())\nprintln(Nope.nope())\n")},
	}
	messages := func(diags []ant.Diagnostic) string {
		var lines []string
		for _, d := range diags {
			lines = append(lines, d.Error())
		}
		return strings.Join(lines, "\n")
	}

	// The checker reads the file for the root's whole program, a module in another directory
	// included, and not for a root elsewhere.
	prog, err := ant.Load("app/main.ant", ant.WithFS(fsys))
	if err != nil {
		t.Fatal(err)
	}
	if diags := prog.Check(); len(diags) != 0 {
		t.Errorf("check with a declaration file: %s", messages(diags))
	}
	own, err := ant.Load("lib/own.ant", ant.WithFS(fsys))
	if err != nil {
		t.Fatal(err)
	}
	if got := messages(own.Check()); !strings.Contains(got, "Tui is not defined") {
		t.Errorf("a root in another directory: %s", got)
	}

	// Its own errors are reported with the program's, and prelude.d.ant is never read.
	diags, _ := ant.Check([]string{"bad/main.ant", "bad/bad.d.ant", "app/main.ant", "lib/text.ant"}, ant.WithFS(fsys))
	got := messages(diags)
	for _, want := range []string{"bad/bad.d.ant:1:15: unknown type Missing", "bad/bad.d.ant:2:6: File is a prelude name", "bad/main.ant:2:9: Nope is not defined", "lib/text.ant:1:36: Tui is not defined"} {
		if strings.Count(got, want) != 1 {
			t.Errorf("ant.Check: want %q once in:\n%s", want, got)
		}
	}
	if strings.Contains(got, "app/") {
		t.Errorf("ant.Check reports app, which is well typed:\n%s", got)
	}

	// No host provides Tui: the run is a load error at the first read, in initialization order,
	// and so is a test; a program beside the file that reads none of its names runs.
	var out strings.Builder
	err = prog.Run(context.Background(), ant.WithStdout(&out))
	var le *ant.LoadError
	if !errors.As(err, &le) || ant.ExitStatus(err) != 2 || len(le.Diagnostics) != 1 ||
		!strings.HasPrefix(le.Diagnostics[0].Error(), "lib/text.ant:1:36: Tui is declared in app/tui.d.ant, which does not implement it") || out.Len() != 0 {
		t.Errorf("run without a host: %v, output %q", err, out.String())
	}
	if _, err := prog.Start(context.Background()); !errors.As(err, &le) {
		t.Errorf("start without a host: %v", err)
	}
	test, err := ant.Load("app/wide_test.ant", ant.WithFS(fsys))
	if err != nil {
		t.Fatal(err)
	}
	if err := test.Test(context.Background(), ant.TestConfig{}); !errors.As(err, &le) {
		t.Errorf("test without a host: %v", err)
	}
	plain, err := ant.Load("app/plain.ant", ant.WithFS(fsys))
	if err != nil {
		t.Fatal(err)
	}
	if err := plain.Run(context.Background(), ant.WithStdout(&out)); err != nil || out.String() != "plain\n" {
		t.Errorf("a program that reads no declared name: %v, output %q", err, out.String())
	}

	// A host that provides Tui has its own signatures read, and the file's are not: the file
	// declares Tui.height, the host does not, and the host gives width another type.
	host := ant.Namespace{
		Name:       "Tui",
		Signatures: "func Tui.width(s: string): string",
		Funcs: map[string]ant.HostFunc{
			"width": func(rt ant.Runtime, args []value.Value) (value.Value, error) { return "wide", nil },
		},
	}
	hosted, err := ant.Load("app/main.ant", ant.WithFS(fsys), ant.WithNamespace(host))
	if err != nil {
		t.Fatal(err)
	}
	if got := messages(hosted.Check()); !strings.Contains(got, "Tui has no member height") || !strings.Contains(got, "lib/text.ant:1:36") {
		t.Errorf("check with the host's own signatures: %s", got)
	}
	fsys["app/main.ant"] = &fstest.MapFile{Data: []byte("println(Tui.width(\"ab\"))\n")}
	hosted, err = ant.Load("app/main.ant", ant.WithFS(fsys), ant.WithNamespace(host))
	if err != nil {
		t.Fatal(err)
	}
	out.Reset()
	if err := hosted.Run(context.Background(), ant.WithStdout(&out)); err != nil || out.String() != "wide\n" {
		t.Errorf("run in a host that provides Tui: %v, output %q", err, out.String())
	}
}
