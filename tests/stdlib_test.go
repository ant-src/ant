package tests

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/ant-src/ant"
	"github.com/ant-src/ant/ast"
	"github.com/ant-src/ant/parser"
	"github.com/ant-src/ant/value"
)

// TestStdlib runs the library tables under stdlib/, as stdlib/README.md specifies: each file
// is checked as a program, then evaluated one input at a time in a fresh session, and every
// case's rendered value or panic is compared with what its "// =>" comment expects.
func TestStdlib(t *testing.T) {
	paths, _ := filepath.Glob("stdlib/*.ant")
	if len(paths) == 0 {
		t.Fatal("no library tables found")
	}
	for _, path := range paths {
		t.Run(strings.TrimSuffix(filepath.Base(path), ".ant"), func(t *testing.T) {
			t.Parallel()
			src, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			for _, p := range tableProblems(t.Context(), path, string(src), tableTimeout) {
				t.Error(p)
			}
		})
	}
}

// tableTimeout bounds one table's run, so a table that blocks fails alone instead of holding
// the suite.
const tableTimeout = 10 * time.Second

// tableProblems reads, checks and runs the table at path, whose text is src, and returns what
// is wrong with it: nothing when every case holds.
func tableProblems(ctx context.Context, path, src string, timeout time.Duration) []string {
	inputs, err := readTable(path, src)
	if err != nil {
		return []string{err.Error()}
	}
	if problems := checkTable(path, inputs); len(problems) > 0 {
		return problems
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return runTable(ctx, path, inputs)
}

// tableInput is one input of a library table: a setup, or a case with what it expects.
type tableInput struct {
	line   int    // the input's first line
	last   int    // its last line, where a case's comment is
	src    string // the text, without the expectation
	isCase bool
	want   string
	file   *ast.File
}

// caseMarker finds a case's comment. The spaces in it are optional, so that "//=> 2" is a case
// that fails rather than a setup whose value nobody looks at.
var caseMarker = regexp.MustCompile(`//\s*=>`)

// forbidden names what a library table may not reach: the namespaces of effects, and the
// functions that wait (stdlib/README.md).
var forbidden = map[string]bool{
	"Chan": true, "File": true, "Http": true, "Net": true, "Os": true,
	"Time.sleep": true, "Time.ticker": true, "Context.wait": true, "WaitGroup.wait": true,
	"Mutex.lock": true, "Mutex.withLock": true, "Cond.wait": true,
	"RwMutex.readLock": true, "RwMutex.writeLock": true, "RwMutex.withRead": true, "RwMutex.withWrite": true,
}

// readTable splits a table into its inputs as the REPL reads them: a line, or as many lines as
// it takes to close what one opens. Blank lines and comment lines between inputs are skipped.
func readTable(path, src string) ([]tableInput, error) {
	var inputs []tableInput
	var buf []string
	first := 0
	for i, line := range strings.Split(src, "\n") {
		if len(buf) == 0 {
			trimmed := strings.TrimSpace(line)
			if trimmed == "" || strings.HasPrefix(trimmed, "//") {
				continue
			}
			first = i + 1
		}
		buf = append(buf, line)
		if ant.Incomplete(strings.Join(buf, "\n")) {
			continue
		}
		in := tableInput{line: first, last: i + 1}
		lines := append([]string(nil), buf...)
		buf = nil
		end := lines[len(lines)-1]
		if locs := caseMarker.FindAllStringIndex(end, -1); locs != nil {
			at := locs[len(locs)-1]
			in.isCase = true
			in.want = strings.TrimSpace(end[at[1]:])
			lines[len(lines)-1] = strings.TrimRight(end[:at[0]], " \t")
		}
		in.src = strings.Join(lines, "\n")
		if caseMarker.MatchString(in.src) {
			return nil, fmt.Errorf("%s:%d: a // => belongs at the end of a case's last line, and nowhere else", path, in.line)
		}
		var err error
		in.file, err = parser.Parse(path, in.src, parser.Session)
		if err != nil {
			return nil, fmt.Errorf("%s:%d: %v", path, in.line, err)
		}
		if name := forbiddenName(in.file); name != "" {
			return nil, fmt.Errorf("%s:%d: a library table cannot wait, start a goroutine, or have an effect: %s", path, in.line, name)
		}
		f := in.file
		if len(f.Imports) > 0 {
			return nil, fmt.Errorf("%s:%d: a library table cannot import", path, in.line)
		}
		if in.isCase {
			if len(f.Decls) > 0 || len(f.Stmts) != 1 {
				return nil, fmt.Errorf("%s:%d: a case is one expression", path, in.last)
			}
		} else if len(f.Stmts) > 0 {
			// In a session a var is both a declaration and a statement; it goes into main.
			for _, d := range f.Decls {
				if _, ok := d.(*ast.VarDecl); !ok {
					return nil, fmt.Errorf("%s:%d: an input mixes declarations and statements", path, in.line)
				}
			}
		}
		inputs = append(inputs, in)
	}
	if len(buf) > 0 {
		return nil, fmt.Errorf("%s:%d: the input is not closed", path, first)
	}
	return inputs, nil
}

// forbiddenName returns the first thing in f a table may not contain, "go", "select", or a
// forbidden name, or "".
func forbiddenName(f *ast.File) string {
	found := ""
	var visit func(v reflect.Value)
	visit = func(v reflect.Value) {
		if found != "" || !v.IsValid() {
			return
		}
		switch v.Kind() {
		case reflect.Pointer, reflect.Interface:
			if v.IsNil() {
				return
			}
			switch n := v.Interface().(type) {
			case *ast.Go:
				found = "go"
				return
			case *ast.Select:
				found = "select"
				return
			case *ast.Field:
				if name := dottedName(n); name != "" {
					ns, _, _ := strings.Cut(name, ".")
					if forbidden[name] || forbidden[ns] {
						found = name
						return
					}
				}
			}
			visit(v.Elem())
		case reflect.Struct:
			for i := 0; i < v.NumField(); i++ {
				if v.Type().Field(i).IsExported() {
					visit(v.Field(i))
				}
			}
		case reflect.Slice:
			for i := 0; i < v.Len(); i++ {
				visit(v.Index(i))
			}
		}
	}
	visit(reflect.ValueOf(f))
	return found
}

// checkTable checks the table as a program: declarations at the top level, in order, and every
// other input in main, each case as var _ = <expression>. It returns each diagnostic at the
// table's own line.
func checkTable(path string, inputs []tableInput) []string {
	var top, body []string
	var topLines, bodyLines []int
	for _, in := range inputs {
		lines := strings.Split(in.src, "\n")
		if in.isCase {
			lines[0] = "var _ = " + lines[0]
		}
		for k, l := range lines {
			if len(in.file.Stmts) == 0 {
				top = append(top, l)
				topLines = append(topLines, in.line+k)
			} else {
				body = append(body, "  "+l)
				bodyLines = append(bodyLines, in.line+k)
			}
		}
	}
	lines := append(append(append(top, "func main(): void {"), body...), "}")
	src := strings.Join(lines, "\n") + "\n"
	lineOf := append(append(topLines, 0), bodyLines...) // the program's line n is lineOf[n-1]
	abs, _ := filepath.Abs(path)
	var problems []string
	report := func(d ant.Diagnostic) {
		line := 0
		if d.Pos.Path == abs && d.Pos.Line >= 1 && d.Pos.Line <= len(lineOf) {
			line = lineOf[d.Pos.Line-1]
		}
		problems = append(problems, fmt.Sprintf("%s:%d: %s", path, line, d.Message))
	}
	prog, err := ant.Load(abs, ant.WithSource(abs, src))
	if err != nil {
		var le *ant.LoadError
		if !errors.As(err, &le) {
			return []string{fmt.Sprintf("%s: %v", path, err)}
		}
		for _, d := range le.Diagnostics {
			report(d)
		}
		return problems
	}
	for _, d := range prog.Check() {
		report(d)
	}
	return problems
}

// runTable evaluates the inputs in order in a fresh session and compares each case. A panic's
// report on standard error is the session's, not the table's, and is dropped; anything else
// either stream receives fails the input that wrote it.
func runTable(ctx context.Context, path string, inputs []tableInput) []string {
	var stdout, stderr bytes.Buffer
	s, err := ant.NewSession(ctx, ant.WithStdin(strings.NewReader("")), ant.WithStdout(&stdout),
		ant.WithStderr(&stderr), ant.WithArgs(nil))
	if err != nil {
		return []string{fmt.Sprintf("%s: %v", path, err)}
	}
	defer s.Close()
	var problems []string
	fail := func(line int, format string, args ...any) {
		problems = append(problems, fmt.Sprintf("%s:%d: ", path, line)+fmt.Sprintf(format, args...))
	}
	for _, in := range inputs {
		before := stderr.Len()
		v, err := s.Eval(in.src)
		if ctx.Err() != nil {
			fail(in.line, "the table did not finish in time: %v", ctx.Err())
			break
		}
		var pe *ant.PanicError
		panicked := errors.As(err, &pe)
		if panicked {
			stderr.Truncate(before)
		}
		switch {
		case in.isCase && panicked:
			if got := "panic: " + pe.Message; got != in.want {
				fail(in.last, "%s\n    got:  %s\n    want: %s", in.src, got, in.want)
			}
		case err != nil:
			fail(in.line, "the input failed: %v", err)
		case in.isCase && v == value.Void:
			fail(in.last, "a case has a value, and this one is void; make it a setup")
		case in.isCase:
			if got := value.RenderQuoted(v); got != in.want {
				fail(in.last, "%s\n    got:  %s\n    want: %s", in.src, got, in.want)
			}
		case v != value.Void:
			fail(in.line, "this input has a value that nothing checks; give it a // => or bind it with var")
		}
		if stdout.Len() > 0 || stderr.Len() > 0 {
			fail(in.line, "a library table writes nothing, got %q", stdout.String()+stderr.String())
			stdout.Reset()
			stderr.Reset()
		}
	}
	return problems
}

// tablePanicCases returns, for every case of every table that expects a panic, the function it
// covers: its outermost call's, when the panic's message names that function.
func tablePanicCases(t *testing.T) []string {
	t.Helper()
	paths, _ := filepath.Glob("stdlib/*.ant")
	var names []string
	for _, path := range paths {
		src, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		inputs, err := readTable(path, string(src))
		if err != nil {
			t.Error(err)
			continue
		}
		for _, in := range inputs {
			if name := coveredPanic(in); name != "" {
				names = append(names, name)
			}
		}
	}
	return names
}

// coveredPanic is the function a case covers: for String.repeat("x", -1) expecting
// "panic: String.repeat needs ...", String.repeat; "" for a case of a value, a case that is not
// a call, and a case whose panic comes from a call inside the outermost one.
func coveredPanic(in tableInput) string {
	if !in.isCase || !strings.HasPrefix(in.want, "panic: ") {
		return ""
	}
	call, ok := in.file.Stmts[0].(*ast.Call)
	if !ok {
		return ""
	}
	name := dottedName(call.Fn)
	if name == "" || !strings.HasPrefix(in.want, "panic: "+name) {
		return ""
	}
	return name
}

// dottedName is the name a path of fields spells, "Time.sleep", or "".
func dottedName(e ast.Expr) string {
	switch e := e.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.Field:
		if x := dottedName(e.X); x != "" {
			return x + "." + e.Name
		}
	}
	return ""
}
