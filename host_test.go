package ant_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"testing/fstest"

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
	prog := loadSrc(t, `func main(): void { print(Os.env("GREETING")); print(Os.env("MISSING")) }`)
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
func main(): void {}
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
	prog := loadSrc(t, `func main(): void { print(File.exists("x")) }`, ant.WithoutNamespaces("File"))
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
	_, err := ant.Load("main.ant", ant.WithSource("main.ant", "func main(): void { var x = }"), ant.WithFS(fstest.MapFS{}))
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
