package tests

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

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
			inputs, err := readTable(path)
			if err != nil {
				t.Fatal(err)
			}
			if !checkTable(t, path, inputs) {
				return
			}
			runTable(t, path, inputs)
		})
	}
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

// waits names what a library table may not contain: a goroutine, a select, or a call that
// waits.
var waits = regexp.MustCompile(`\bgo\s*\{|\bselect\s*\{|\b(Time\.sleep|Time\.ticker|Chan\.send|Chan\.receive|Chan\.values|Context\.wait|WaitGroup\.wait|Mutex\.lock|RwMutex\.lock|RwMutex\.rlock|Cond\.wait|Os\.exec|Http\.|Net\.)\b`)

// readTable splits a table into its inputs as the REPL reads them: a line, or as many lines as
// it takes to close what one opens. Blank lines and comment lines between inputs are skipped.
func readTable(path string) ([]tableInput, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var inputs []tableInput
	var buf []string
	first := 0
	for i, line := range strings.Split(string(data), "\n") {
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
		if at := strings.LastIndex(end, "// =>"); at >= 0 {
			in.isCase = true
			in.want = strings.TrimSpace(end[at+len("// =>"):])
			lines[len(lines)-1] = strings.TrimRight(end[:at], " \t")
		}
		in.src = strings.Join(lines, "\n")
		if m := waits.FindString(in.src); m != "" {
			return nil, fmt.Errorf("%s:%d: a library table cannot wait or start a goroutine: %s", path, in.line, m)
		}
		in.file, err = parser.Parse(path, in.src, parser.Session)
		if err != nil {
			return nil, fmt.Errorf("%s:%d: %v", path, in.line, err)
		}
		f := in.file
		if in.isCase {
			if len(f.Imports) > 0 || len(f.Decls) > 0 || len(f.Stmts) != 1 {
				return nil, fmt.Errorf("%s:%d: a case is one expression", path, in.last)
			}
			if _, ok := f.Stmts[0].(ast.Expr); !ok {
				return nil, fmt.Errorf("%s:%d: a case is an expression, not a declaration", path, in.last)
			}
		} else if len(f.Stmts) > 0 {
			// In a session a var is both a declaration and a statement; it goes into main.
			for _, d := range f.Decls {
				if _, ok := d.(*ast.VarDecl); !ok || len(f.Imports) > 0 {
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

// checkTable checks the table as a program: declarations and imports at the top level, in
// order, and every other input in main, each case as var _ = (expression). It reports each
// diagnostic at the table's own line and returns whether there were none.
func checkTable(t *testing.T, path string, inputs []tableInput) bool {
	var top, body []string
	var topLines, bodyLines []int
	for _, in := range inputs {
		lines := strings.Split(in.src, "\n")
		if in.isCase {
			lines[0] = "var _ = (" + lines[0]
			lines[len(lines)-1] += ")"
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
	report := func(d ant.Diagnostic) {
		line := 0
		if d.Pos.Path == abs && d.Pos.Line >= 1 && d.Pos.Line <= len(lineOf) {
			line = lineOf[d.Pos.Line-1]
		}
		t.Errorf("tests/%s:%d: %s", path, line, d.Message)
	}
	prog, err := ant.Load(abs, ant.WithSource(abs, src))
	if err != nil {
		var le *ant.LoadError
		if !errors.As(err, &le) {
			t.Fatal(err)
		}
		for _, d := range le.Diagnostics {
			report(d)
		}
		return false
	}
	diags := prog.Check()
	for _, d := range diags {
		report(d)
	}
	return len(diags) == 0
}

// runTable evaluates the inputs in order in a fresh session and compares each case. A panic's
// report on standard error is the session's, not the table's, and is dropped; anything else
// either stream receives fails the input that wrote it.
func runTable(t *testing.T, path string, inputs []tableInput) {
	var stdout, stderr bytes.Buffer
	s, err := ant.NewSession(t.Context(), ant.WithStdin(strings.NewReader("")), ant.WithStdout(&stdout),
		ant.WithStderr(&stderr), ant.WithArgs(nil))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	for _, in := range inputs {
		before := stderr.Len()
		v, err := s.Eval(in.src)
		var pe *ant.PanicError
		panicked := errors.As(err, &pe)
		if panicked {
			stderr.Truncate(before)
		}
		switch {
		case in.isCase:
			got := ""
			switch {
			case panicked:
				got = "panic: " + pe.Message
			case err != nil:
				got = "error: " + err.Error()
			case v == value.Void:
				t.Errorf("tests/%s:%d: a case has a value, and this one is void; make it a setup", path, in.last)
				continue
			default:
				got = value.RenderQuoted(v)
			}
			if got != in.want {
				t.Errorf("tests/%s:%d: %s\n    got:  %s\n    want: %s", path, in.last, in.src, got, in.want)
			}
		case err != nil:
			t.Errorf("tests/%s:%d: the setup failed: %v", path, in.line, err)
		}
		if stdout.Len() > 0 || stderr.Len() > 0 {
			t.Errorf("tests/%s:%d: a library table writes nothing, got %q", path, in.line, stdout.String()+stderr.String())
			stdout.Reset()
			stderr.Reset()
		}
	}
}

// tablePanicCases returns, for every case of every table that expects a panic, the function
// of its outermost call: "String.repeat" for String.repeat("x", -1), "" when the case is not a
// call of a namespace's function.
func tablePanicCases(t *testing.T) []string {
	t.Helper()
	paths, _ := filepath.Glob("stdlib/*.ant")
	var names []string
	for _, path := range paths {
		inputs, err := readTable(path)
		if err != nil {
			t.Error(err)
			continue
		}
		for _, in := range inputs {
			if in.isCase && strings.HasPrefix(in.want, "panic:") {
				names = append(names, calledName(in.file.Stmts[0].(ast.Expr)))
			}
		}
	}
	return names
}

// calledName is the dotted name of the function an expression calls, or "".
func calledName(e ast.Expr) string {
	call, ok := e.(*ast.Call)
	if !ok {
		return ""
	}
	var name func(ast.Expr) string
	name = func(e ast.Expr) string {
		switch e := e.(type) {
		case *ast.Ident:
			return e.Name
		case *ast.Field:
			if x := name(e.X); x != "" {
				return x + "." + e.Name
			}
		}
		return ""
	}
	return name(call.Fn)
}
