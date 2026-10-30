package ant_test

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/ant-src/ant"
	"github.com/ant-src/ant/value"
)

func TestSession(t *testing.T) {
	var out, errOut strings.Builder
	s, err := ant.NewSession(context.Background(), ant.WithStdout(&out), ant.WithStderr(&errOut),
		ant.WithSource("lib.ant", "export var counter = 0\nexport func bump(): void { counter += 1 }\nexport type P = { x: int }"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	eval := func(src string) value.Value {
		t.Helper()
		v, err := s.Eval(src)
		if err != nil {
			t.Fatalf("%s: %v", src, err)
		}
		return v
	}
	if v := eval("Os.program()"); v != "" {
		t.Errorf("Os.program() in a session = %v", v)
	}
	if v := eval("1 + 2"); v != int64(3) {
		t.Errorf("1 + 2 = %v", v)
	}
	if v := eval("var x = 10"); v != value.Void {
		t.Errorf("a declaration is void, got %v", v)
	}
	if v := eval("x * 2"); v != int64(20) {
		t.Errorf("x * 2 = %v", v)
	}
	eval("func double(n: int): int { n * 2 }")
	eval("type Point = { x: int, y: int }")
	if v := eval("double(x)"); v != int64(20) {
		t.Errorf("double(x) = %v", v)
	}
	if v := value.Render(eval("Point{ x: 1, y: 2 }")); v != "Point{ x: 1, y: 2 }" {
		t.Errorf("record: %s", v)
	}
	eval("var x = \"again\"")
	if v := eval("x"); v != "again" {
		t.Errorf("redeclared x = %v", v)
	}
	eval(`println("hello")`)
	if !strings.Contains(out.String(), "hello\n") {
		t.Errorf("stdout: %q", out.String())
	}
	// multi-line input
	if v := eval("var f = func(a: int): int {\n  a + 1\n}\nf(1)"); v != int64(2) {
		t.Errorf("multi-line: %v", v)
	}
	// a panic is reported and the session continues
	if _, err := s.Eval("panic(\"boom\")"); err == nil {
		t.Error("expected a panic error")
	} else if _, ok := err.(*ant.PanicError); !ok {
		t.Errorf("expected *PanicError, got %T", err)
	}
	if !strings.Contains(errOut.String(), "boom") {
		t.Errorf("stderr lacks the report: %q", errOut.String())
	}
	if v := eval("x"); v != "again" {
		t.Errorf("after a panic x = %v", v)
	}
	// a syntax error is a load error
	if _, err := s.Eval("var = 1"); err == nil {
		t.Error("expected a syntax error")
	} else if _, ok := err.(*ant.LoadError); !ok {
		t.Errorf("expected *LoadError, got %T", err)
	}
	// imports stay in force and modules initialize once
	eval(`import "lib.ant" { counter, bump, P }`)
	eval("bump()")
	eval("bump()")
	if v := eval("counter"); v != int64(2) {
		t.Errorf("counter = %v", v)
	}
	if v := value.Render(eval("P{ x: 7 }")); v != "P{ x: 7 }" {
		t.Errorf("imported type: %s", v)
	}
	// an input after an import keeps its own positions, in a panic's trace and in a load error
	eval("func deep(): int {\n  [1][3]\n}")
	if _, err := s.Eval("deep()"); err == nil {
		t.Error("expected a panic error")
	} else if pe, ok := err.(*ant.PanicError); !ok || len(pe.Trace) == 0 || pe.Trace[0].Pos.Line != 2 {
		t.Errorf("trace after an import: %v", err)
	}
	if _, err := s.Eval("var a = 1\nvar = 2"); err == nil {
		t.Error("expected a load error")
	} else if le, ok := err.(*ant.LoadError); !ok || le.Diagnostics[0].Pos.Line != 2 {
		t.Errorf("load error after an import: %v", err)
	}
	// incomplete inputs
	for _, src := range []string{"func f() {", "var s = \"open", "[1,", "\"${",
		"f(1,\n", "/* open\n", "`a\n", "var s = \"open\n"} {
		if !ant.Incomplete(src) {
			t.Errorf("%q should be incomplete", src)
		}
	}
	for _, src := range []string{"1 + 2", "func f() {}", "var x = = 1",
		"var r = /ab\n", "if true\n", "var x = = 1\n"} {
		if ant.Incomplete(src) {
			t.Errorf("%q should not be incomplete", src)
		}
	}
}

// TestSessionContext checks that a session's context ends an input that never waits.
func TestSessionContext(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	s, err := ant.NewSession(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if _, err := s.Eval("var n = 0\nwhile true { n += 1 }"); err == nil {
		t.Fatal("an endless loop returned without an error")
	}
}

// TestSessionGrowsUnderGoroutines: an input that declares new names grows the session's module
// while a goroutine of an earlier input stores into it, reads an import and calls a function;
// no store is lost and nothing races.
func TestSessionGrowsUnderGoroutines(t *testing.T) {
	s, err := ant.NewSession(context.Background(), ant.WithStdout(&strings.Builder{}),
		ant.WithSource("lib.ant", "export var step = 1"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if _, err := s.Eval(`import "./lib.ant" L
var n = 0
var done: Channel[bool] = Channel.new(1)
func add(): void { n += L.step }
go { for i in Iter.range(0, 100000) { add() }; Channel.send(done, true) }`); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 200; i++ {
		if _, err := s.Eval(fmt.Sprintf("var v%d = %d", i, i)); err != nil {
			t.Fatal(err)
		}
	}
	v, err := s.Eval("Channel.receive(done)\nn")
	if err != nil || v != value.Value(int64(100000)) {
		t.Errorf("n = %v, %v; want 100000", v, err)
	}
}

// TestSessionImportInFailedInput checks that an import stays in force when the rest of its
// input panics, as the input's declarations do.
func TestSessionImportInFailedInput(t *testing.T) {
	var errOut strings.Builder
	s, err := ant.NewSession(context.Background(), ant.WithStderr(&errOut),
		ant.WithSource("lib.ant", "export var counter = 0\nexport func bump(): int { counter += 1; counter }"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if _, err := s.Eval("import \"lib.ant\" L\nvar a = L.bump()\npanic(\"boom\")"); err == nil {
		t.Fatal("expected a panic error")
	}
	if v, err := s.Eval("L.bump()"); err != nil || v != int64(2) {
		t.Errorf("L.bump() = %v, %v", v, err)
	}
}

// TestSessionTypeFuncLater checks that a function may be declared on a type an earlier input
// declared, and not on a name an earlier input redeclared as a variable.
func TestSessionTypeFuncLater(t *testing.T) {
	s, err := ant.NewSession(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	for _, src := range []string{"type P = { x: int }", "func P.get(p: P): int { p.x * 2 }"} {
		if _, err := s.Eval(src); err != nil {
			t.Fatalf("%s: %v", src, err)
		}
	}
	if v, err := s.Eval("P.get(P{ x: 4 })"); err != nil || v != int64(8) {
		t.Errorf("P.get = %v, %v", v, err)
	}
	if _, err := s.Eval("var P = 1"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Eval("func P.other(): int { 1 }"); err == nil {
		t.Error("expected a load error for a function on a variable")
	}
	// a type whose input failed to load is not a type of the session
	if _, err := s.Eval("type Z = { x: int }; var print = 1"); err == nil {
		t.Fatal("expected a load error")
	}
	if _, err := s.Eval("func Z.get(): int { 1 }"); err == nil {
		t.Error("expected a load error for a function on a type that failed to load")
	}
}

// TestSessionFieldDefaults: a default reads the session's variables at each literal, a later
// input's literal fills in a type an earlier input declared, and a redeclared type takes its new
// defaults.
func TestSessionFieldDefaults(t *testing.T) {
	s, err := ant.NewSession(context.Background(), ant.WithSource("lib.ant", "var hidden = 3\nexport type L = { n: int = hidden }"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	render := func(src string) string {
		t.Helper()
		v, err := s.Eval(src)
		if err != nil {
			t.Fatalf("%s: %v", src, err)
		}
		return value.Render(v)
	}
	render("var base = 5")
	render("type C = { n: int = base, tags: Array[string] = [] }")
	if got := render("C{}"); got != "C{ n: 5, tags: [] }" {
		t.Errorf("C{} = %s", got)
	}
	render("base = 7")
	if got := render("C{ tags: [\"a\"] }"); got != `C{ n: 7, tags: ["a"] }` {
		t.Errorf("after base = 7: %s", got)
	}
	render("var one = C{}")
	render("Array.push(one.tags, \"x\")")
	if got := render("C{}.tags"); got != "[]" {
		t.Errorf("a default shared its array: %s", got)
	}
	render("type C = { n: int = 1, m: int = 2 }")
	if got := render("C{ m: 3 }"); got != "C{ n: 1, m: 3 }" {
		t.Errorf("redeclared C = %s", got)
	}
	render(`import "lib.ant" { L }`)
	if got := render("L{}"); got != "L{ n: 3 }" {
		t.Errorf("imported L{} = %s", got)
	}
	if got := render("C{ n: 4 }"); got != "C{ n: 4, m: 2 }" {
		t.Errorf("C{ n: 4 } = %s", got)
	}
	if _, err := s.Eval("type D = { a: int, b: int = 1 }\nD{ b: 2 }"); err == nil || !strings.Contains(err.Error(), "missing the field a") {
		t.Errorf("D{ b: 2 } = %v, want missing the field a", err)
	}
}

// TestSessionInterrupts: a session honors WithInterrupts as a run does, so an interrupt ends
// an input that never waits.
func TestSessionInterrupts(t *testing.T) {
	interrupts := make(chan struct{}, 1)
	// The deadline only bounds the test: without the interrupt the input would never end.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s, err := ant.NewSession(ctx, ant.WithInterrupts(interrupts))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	time.AfterFunc(50*time.Millisecond, func() { interrupts <- struct{}{} })
	if _, err := s.Eval("var n = 0\nwhile true { n += 1 }"); !errors.Is(err, ant.ErrInterrupted) {
		t.Fatalf("an interrupted input ended with %v, want ErrInterrupted", err)
	}
}

// TestSessionComplete checks that completion sees what earlier inputs declared, by their latest
// types, and what the position allows.
func TestSessionComplete(t *testing.T) {
	s, err := ant.NewSession(context.Background(), ant.WithStdout(io.Discard), ant.WithStderr(io.Discard))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	for _, src := range []string{"type Point = { x: int, y: int }", "var p = Point{ x: 1, y: 2 }", "var n = 1", "var n = \"a\"", "func twice(k: int): int { k * 2 }", "p.x + 1"} {
		if _, err := s.Eval(src); err != nil {
			t.Fatalf("%s: %v", src, err)
		}
	}
	labels := func(src string) map[string]ant.Completion {
		out := map[string]ant.Completion{}
		for _, c := range s.Complete(src, len(src)) {
			out[c.Label] = c
		}
		return out
	}
	if got := labels("p."); got["x"].Kind != "field" || got["y"].Detail != "int" {
		t.Errorf("p.: %v", got)
	}
	if got := labels("n | String."); got["length"].Kind != "function" {
		t.Errorf("n | String.: %v", got)
	}
	if got := labels("tw"); got["twice"].Kind != "function" || !strings.Contains(got["twice"].Detail, "k: int") {
		t.Errorf("tw: %v", got["twice"])
	}
	if got := labels("twice("); !got["k:"].Preferred || got["k:"].Insert != "k: " {
		t.Errorf("twice(: %v", got["k:"])
	}
	// the parameters of a call come in their order
	var params []string
	for _, c := range s.Complete("Http.get(", len("Http.get(")) {
		if c.Preferred {
			params = append(params, c.Label)
		}
	}
	if strings.Join(params, " ") != "url: headers: timeout: ctx:" {
		t.Errorf("Http.get( parameters: %q", params)
	}
	// after a named argument, the parameters not given yet, in their order
	for _, src := range []string{`Http.get(url: "x", `, `var a = try Http.get(url: "x", `, `Http.get(url: "x", ti`, `Http.get("x", timeout: null, `} {
		params = nil
		for _, c := range s.Complete(src, len(src)) {
			if c.Preferred {
				params = append(params, c.Label)
			}
		}
		want := "headers: timeout: ctx:"
		if strings.Contains(src, "timeout: null") {
			want = "headers: ctx:"
		}
		if strings.Join(params, " ") != want {
			t.Errorf("%s: parameters %q, want %q", src, params, want)
		}
	}
	if got := labels(""); got["Array"].Kind != "namespace" || got["var"].Kind != "keyword" || got["n"].Detail != "string" {
		t.Errorf("empty: Array %v, var %v, n %v", got["Array"], got["var"], got["n"])
	}
	if got := s.Complete("x", 5); got != nil {
		t.Errorf("offset past the end: %v", got)
	}
	// a Result is an Ok or an Err, which share no field, so the type promises none; the
	// value at runtime is an Ok, whose value is read through
	if _, err := s.Eval("var r: Result[Point, Error] = Ok{ value: p }"); err != nil {
		t.Fatal(err)
	}
	if got := labels("r."); got["value"].Kind != "field" || got["value"].Detail != "Point" || len(got) != 1 {
		t.Errorf("r.: %v", got)
	}
	if got := labels("r.va"); got["value"].Label != "value" {
		t.Errorf("r.va: %v", got)
	}
	if got := labels("r.value."); got["x"].Detail != "int" || got["y"].Detail != "int" {
		t.Errorf("r.value.: %v", got)
	}
	if got := labels("r?.value?."); got["x"].Detail != "int" {
		t.Errorf("r?.value?.: %v", got)
	}
	// not through an expression, and not a field that is not there
	if got := labels("twice(1).r."); len(got) != 0 {
		t.Errorf("twice(1).r.: %v", got)
	}
	if got := labels("r.nope."); len(got) != 0 {
		t.Errorf("r.nope.: %v", got)
	}
}

// BenchmarkSessionComplete measures one completion on a session with an import and many
// declarations: what a keystroke costs.
func BenchmarkSessionComplete(b *testing.B) {
	s, err := ant.NewSession(context.Background(), ant.WithStdout(io.Discard), ant.WithStderr(io.Discard))
	if err != nil {
		b.Fatal(err)
	}
	defer s.Close()
	if _, err := s.Eval(`import "tests/run/lib/geometry.ant" { distance, Point }`); err != nil {
		b.Fatal(err)
	}
	for i := 0; i < 200; i++ {
		if _, err := s.Eval(fmt.Sprintf("func f%d(n: int): int { n + %d }", i, i)); err != nil {
			b.Fatal(err)
		}
	}
	s.Eval("var xs = [1, 2, 3]")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.Complete("xs | Array.", len("xs | Array."))
	}
}

// TestSessionTry checks that an input is a script body of its own: try unwraps an Ok and
// returns an Err from the input, which is reported and leaves the session going; return gives
// the input's result; defer runs when the input ends.
func TestSessionTry(t *testing.T) {
	var out, errOut strings.Builder
	s, err := ant.NewSession(context.Background(), ant.WithStdout(&out), ant.WithStderr(&errOut))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	eval := func(src string) (value.Value, error) {
		t.Helper()
		v, err := s.Eval(src)
		if _, ok := err.(*ant.LoadError); ok {
			t.Fatalf("%s: %v", src, err)
		}
		return v, err
	}
	if _, err := eval("var a = try Ok{ value: 1 }"); err != nil {
		t.Fatal(err)
	}
	if v, _ := eval("a + 1"); v != int64(2) {
		t.Errorf("a + 1 = %v", v)
	}
	_, err = eval("var b = try Err{ error: Error{ message: \"no page\" } }\nprintln(\"unreached\")")
	if _, ok := err.(*ant.ResultError); !ok || err.Error() != "no page" {
		t.Errorf("try on Err: %v", err)
	}
	if !strings.Contains(errOut.String(), "no page\n") || strings.Contains(out.String(), "unreached") {
		t.Errorf("stdout %q, stderr %q", out.String(), errOut.String())
	}
	if v, err := eval("return a * 10\n99"); err != nil || v != int64(10) {
		t.Errorf("return: %v, %v", v, err)
	}
	if v, err := eval("var x: int | null = null\nx ?? return \"none\""); err != nil || v != "none" {
		t.Errorf("?? return: %v, %v", v, err)
	}
	out.Reset()
	if _, err := eval("defer { println(\"deferred\") }\nprintln(\"body\")"); err != nil {
		t.Fatal(err)
	}
	if out.String() != "body\ndeferred\n" {
		t.Errorf("defer: %q", out.String())
	}
	// the session goes on after an Err
	if v, _ := eval("a"); v != int64(1) {
		t.Errorf("a = %v", v)
	}
}

// TestSessionWriteThrough checks that a goroutine's output reaches stdout while the session
// waits for its next input.
func TestSessionWriteThrough(t *testing.T) {
	out := &syncBuffer{}
	s, err := ant.NewSession(context.Background(), ant.WithStdout(out), ant.WithStderr(io.Discard))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if _, err := s.Eval("go { Time.sleep(Time.duration(milliseconds: 20)); print(\"late\") }"); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(5 * time.Second)
	for out.String() != "late" {
		if time.Now().After(deadline) {
			t.Fatalf("stdout %q", out.String())
		}
		time.Sleep(5 * time.Millisecond)
	}
}

type syncBuffer struct {
	mu sync.Mutex
	sb strings.Builder
}

func (b *syncBuffer) Write(p []byte) (int, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.sb.Write(p)
}

func (b *syncBuffer) String() string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.sb.String()
}

// TestSessionTypeOf checks the types :type prints, over what the session declared.
func TestSessionTypeOf(t *testing.T) {
	s, err := ant.NewSession(context.Background(), ant.WithStdout(io.Discard), ant.WithStderr(io.Discard))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	for _, src := range []string{"type Point = { x: int, y: int }", "var p = Point{ x: 1, y: 2 }", "func twice(k: int): int { k * 2 }", "var xs = [1, 2]"} {
		if _, err := s.Eval(src); err != nil {
			t.Fatal(err)
		}
	}
	for src, want := range map[string]string{
		"p":                        "Point",
		"p.x + 1":                  "int",
		"twice":                    "func(int): int",
		"xs | Array.map({ \"\" })": "Array[string]",
		"Http.get":                 "func(string, Map[string, string] = ..., Time.Duration | null = ..., Context | null = ...): Result[Http.Response, Error]",
		"Http.get(\"u\")":          "Result[Http.Response, Error]",
	} {
		got, err := s.TypeOf(src)
		if err != nil || got != want {
			t.Errorf("TypeOf(%q) = %q, %v; want %q", src, got, err, want)
		}
	}
	for _, src := range []string{"p.z", "nope", "1 +", ""} {
		if got, err := s.TypeOf(src); err == nil {
			t.Errorf("TypeOf(%q) = %q, want an error", src, got)
		} else if le, ok := err.(*ant.LoadError); ok && len(le.Diagnostics) > 0 && le.Diagnostics[0].Pos.Line > 1 {
			t.Errorf("TypeOf(%q): %v is not positioned in the expression", src, err)
		}
	}
}

// TestSessionLast checks that _ reads the result of the last input that had one, and is still
// written to as a discard.
func TestSessionLast(t *testing.T) {
	s, err := ant.NewSession(context.Background(), ant.WithStdout(io.Discard), ant.WithStderr(io.Discard))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if _, err := s.Eval("_"); err == nil {
		t.Error("_ before any result reads")
	}
	for _, step := range []struct {
		src  string
		want value.Value
	}{
		{"40 + 1", int64(41)},
		{"_ + 1", int64(42)},
		{"var x = 5", value.Void}, // void leaves _ as it was
		{"_", int64(42)},
		{"_ = 7", value.Void}, // a discard still
		{"_", int64(42)},
		{"\"s\"", "s"},
		{"_ + \"t\"", "st"},
	} {
		v, err := s.Eval(step.src)
		if err != nil || v != step.want {
			t.Errorf("%s = %v, %v; want %v", step.src, v, err, step.want)
		}
	}
	if v, ok := s.Value("_"); !ok || v != "st" {
		t.Errorf("Value(_) = %v, %v", v, ok)
	}
}
