package ant_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
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
	for _, src := range []string{"func f() {", "var s = \"open", "[1,", "`${",
		"f(1,\n", "/* open\n", "`a\n"} {
		if !ant.Incomplete(src) {
			t.Errorf("%q should be incomplete", src)
		}
	}
	for _, src := range []string{"1 + 2", "func f() {}", "var x = = 1",
		"var s = \"open\n", "var r = /ab\n", "if true\n", "var x = = 1\n"} {
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
