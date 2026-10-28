package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ant-src/ant/ast"
)

// TestRecoverAgrees checks that on a module without syntax errors the recovering parser builds
// the tree the strict one does, and reports nothing.
func TestRecoverAgrees(t *testing.T) {
	for path, src := range corpus(t) {
		f, err := Parse(path, src, Program)
		if err != nil {
			continue
		}
		rf, _, errs := ParseRecover(path, src)
		if len(errs) > 0 {
			t.Errorf("%s: recovering parser reports %v", path, errs[0])
			continue
		}
		if got, want := ast.Sprint(rf), ast.Sprint(f); got != want {
			t.Errorf("%s: trees differ:\n%s\n%s", path, got, want)
		}
		if rf.Script != f.Script || rf.Doc != f.Doc || len(rf.Decls) != len(f.Decls) {
			t.Errorf("%s: file facts differ", path)
		}
	}
}

// TestRecoverFirstError checks that on a module with a syntax error the recovering parser reports
// the strict parser's error, first when the lexer found nothing wrong.
func TestRecoverFirstError(t *testing.T) {
	paths, _ := filepath.Glob("../tests/syntax/*.ant")
	if len(paths) < 50 {
		t.Fatalf("found only %d syntax tests", len(paths))
	}
	for _, path := range paths {
		data, _ := os.ReadFile(path)
		src := string(data)
		_, err := Parse(path, src, Program)
		if err != nil {
			checkFirstError(t, path, src, err.(*Error))
		}
	}
}

func checkFirstError(t *testing.T, name, src string, want *Error) {
	t.Helper()
	_, _, errs := ParseRecover(name, src)
	if len(errs) == 0 {
		t.Errorf("%s: strict parser fails with %v, recovering one reports nothing", name, want)
		return
	}
	for _, e := range errs {
		if e.Pos == want.Pos && e.Msg == want.Msg {
			return
		}
	}
	t.Errorf("%s: strict error %v is not among the recovering parser's: %v", name, want, errs)
}

// TestRecoverSyntaxCases runs the syntax error cases through the recovering parser too.
func TestRecoverSyntaxCases(t *testing.T) {
	for _, c := range syntaxCases() {
		if c.mode != Program {
			continue
		}
		_, err := Parse("test", c.src, c.mode)
		if err == nil {
			continue
		}
		checkFirstError(t, c.name, c.src, err.(*Error))
	}
}

// TestRecoverTruncated checks that the recovering parser never panics on any prefix of a corpus
// file, and that the tree it builds has sound spans.
func TestRecoverTruncated(t *testing.T) {
	for path, src := range corpus(t) {
		_, l, _ := ParseRecover(path, src)
		cuts := map[int]bool{len(src): true}
		for _, tok := range l.Tokens {
			cuts[tok.Off] = true
			cuts[tok.End] = true
			if tok.Off+1 < len(src) {
				cuts[tok.Off+1] = true
			}
		}
		for cut := range cuts {
			func() {
				defer func() {
					if r := recover(); r != nil {
						t.Fatalf("%s cut at %d: panic: %v\n%s", path, cut, r, src[:cut])
					}
				}()
				f, _, _ := ParseRecover(path, src[:cut])
				checkRecoveredSpans(t, path, cut, f)
			}()
		}
	}
}

// checkRecoveredSpans checks the spans of a recovered tree: each node ends where or after it
// starts, inside its parent.
func checkRecoveredSpans(t *testing.T, path string, cut int, f *ast.File) {
	t.Helper()
	var check func(n ast.Node)
	check = func(n ast.Node) {
		ast.Inspect(n, func(c ast.Node) bool {
			if c == n {
				return true
			}
			start, end := ast.Start(c), c.EndPosition()
			if end.Before(start) {
				t.Errorf("%s cut at %d: %T at %v ends at %v, before it starts", path, cut, c, start, end)
			}
			if ps, pe := ast.Start(n), n.EndPosition(); start.Before(ps) || pe.Before(end) {
				t.Errorf("%s cut at %d: %T %v-%v is not inside its parent %T %v-%v", path, cut, c, start, end, n, ps, pe)
			}
			check(c)
			return false
		})
	}
	check(f)
}

// TestRecoverCases checks the trees the recovering parser builds for the states a module passes
// through while it is typed.
func TestRecoverCases(t *testing.T) {
	cases := []struct {
		name, src string
		errs      []string // substrings of the errors, in order
		tree      string
	}{
		// The lexer joins the lines after =, so the error names the keyword on the next line, as
		// Parse does.
		{"missing initializer", "var x = \nvar y = 1", []string{"unexpected 'var'"},
			"(file (var x (bad)) (var y 1))"},
		{"missing equals", "var x: int\nvar y = 1", []string{"expected '=', got newline"},
			"(file (var x:int (bad)) (var y 1))"},
		{"dangling dot", "func f() {\n  var p = g()\n  p.\n}", []string{"expected a field name, got '}'"},
			"(file (func f() { (var p (call g)); (. p ) }))"},
		{"dangling optional dot", "func f() { p?. }", []string{"expected a field name, got '}'"},
			"(file (func f() { (?. p ) }))"},
		{"call left open", "func f() {\n  foo(a, b\n}\nvar z = 1", []string{"expected ')', got '}'"},
			"(file (func f() { (call foo a b) }) (var z 1))"},
		{"blocks left open", "func f() {\n  if x {\n    y\n", []string{"expected '}', got end of file"},
			"(file (func f() { (if x { y }) }))"},
		{"record type left open", "type T = { a: int\nfunc g() {}", []string{"expected '}', got 'func'"},
			"(file (type T {a:int}) (func g() {}))"},
		{"dangling dot in markup", "var h = <div><p>{x.}</p></div>", []string{"expected a field name, got end of file"},
			`(file (var h (html (<div> (<p> (. x ))))))`},
		{"stray closer", "func f() {\n  )\n  var x = 1\n}", []string{"unexpected ')'"},
			"(file (func f() { (bad); (var x 1) }))"},
		{"garbage statement", "func f() {\n  var = 1\n  x\n}", []string{"expected a name or a destructuring after var"},
			"(file (func f() { (bad); x }))"},
		{"two statements one line", "func f() { a b }", []string{"expected a newline or ';' before identifier \"b\""},
			"(file (func f() { a; b }))"},
		{"function without body", "func f(a: int)\nfunc g() {}", []string{"expected '{', got newline"},
			"(file (func f(a:int)) (func g() {}))"},
		{"missing type", "func f(a: ) {}", []string{"expected a type, got ')'"},
			"(file (func f(a:(bad)) {}))"},
		{"header without brace", "func f() {\n  if x\n  y\n}", []string{"expected '{' after the if header, got newline"},
			"(file (func f() { (bad); y }))"},
		{"unterminated string", "var s = \"abc\nvar t = 1", []string{"unterminated string"},
			`(file (var s "abc") (var t 1))`},
		{"export in a script", "export func f() {}\nprintln(1)", []string{"a module with exports cannot have top-level statements"},
			"(file (func export f() {}) (call println 1))"},
		{"bad top-level item", "var x = 1\n)\nvar y = 2", []string{"unexpected ')'"},
			"(file (var x 1) (var y 2) (bad))"},
		// The lexer ends the source with a line break, which the lambda's body meets first.
		{"lambda left open", "var f: func(int): int = { x +", []string{"unexpected newline", "expected '}', got end of file"},
			"(file (var f:(func (int):int) (lambda (+ x (bad)) )))"},
		{"import left open", "import \"./a.ant\" {\nvar x = 1", []string{"expected '}', got 'var'"},
			`(file (import "./a.ant" {}) (var x 1))`},
		{"record literal left open", "func f() {\n  var p = Point{ x: 1,\n  var q = 2\n}", []string{"expected '}', got 'var'"},
			"(file (func f() { (var p (Point x:1)); (var q 2) }))"},
		{"named argument being typed", "func f() { g(a, na }", []string{"expected ')', got '}'"},
			"(file (func f() { (call g a na) }))"},
		// The lexer joins the lines after a dot, so r becomes the field and the pipe's right
		// side is not a call; an editor completes Array. from the tokens, not the tree.
		{"pipe left open", "func f() {\n  var r = xs\n    | Array.\n  r\n}", []string{"the right side of a pipe must be a call"},
			"(file (func f() { (bad) }))"},
		// An import being typed, whose line the lexer joins to the declaration below.
		{"import left incomplete", "import\nvar x = 1\n", []string{"expected a module path string, got 'var'"},
			"(file (var x 1) (bad))"},
		{"reserved word after a statement", "var x = 1 class\nvar y = 2", []string{"'class' is a reserved word"},
			"(file (var x 1) (var y 2))"},
		{"reserved word after a statement in a block", "func f() {\n  a class\n  b\n}", []string{"'class' is a reserved word"},
			"(file (func f() { a; b }))"},
		{"local type redeclares a prelude name", "func f() {\n  type Math = int\n  1\n}", nil,
			"(file (func f() { (type Math int); 1 }))"},
		// A Bad token is the lexer's error alone; for the parser it ends the statement.
		{"bad token", "var x = 1 @ 2\nvar y = 3", []string{"unexpected character '@'"},
			"(file (var x 1) (var y 3) 2)"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			f, _, errs := ParseRecover("test", c.src)
			var got []string
			for _, e := range errs {
				got = append(got, e.Msg)
			}
			if len(got) != len(c.errs) {
				t.Fatalf("errors %q, want %q\n%s", got, c.errs, ast.Sprint(f))
			}
			for i, want := range c.errs {
				if !strings.Contains(got[i], want) {
					t.Errorf("error %d = %q, want one containing %q", i, got[i], want)
				}
			}
			if tree := ast.Sprint(f); tree != c.tree {
				t.Errorf("tree %s, want %s", tree, c.tree)
			}
		})
	}
}
