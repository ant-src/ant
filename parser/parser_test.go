package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ant-src/ant/ast"
)

// specPath finds the specification, a sibling checkout of github.com/ant-src/spec or the file
// ANT_SPEC names.
func specPath(t *testing.T) string {
	t.Helper()
	if p := os.Getenv("ANT_SPEC"); p != "" {
		return p
	}
	p := filepath.Join("..", "..", "spec", "ant-lang.md")
	if _, err := os.Stat(p); err != nil {
		t.Skip("the specification is not checked out next to this repository; set ANT_SPEC")
	}
	return p
}

// TestSpecBlocks parses every ```ant block of the specification in spec mode.
func TestSpecBlocks(t *testing.T) {
	data, err := os.ReadFile(specPath(t))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(data), "\n")
	blocks := 0
	for i := 0; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) != "```ant" {
			continue
		}
		start := i + 1
		j := start
		for j < len(lines) && strings.TrimSpace(lines[j]) != "```" {
			j++
		}
		blocks++
		src := strings.Join(lines[start:j], "\n")
		if _, err := Parse("spec", src, Spec); err != nil {
			e := err.(*Error)
			t.Errorf("block at spec line %d: %s (spec line %d)\n%s", start+1, e.Msg, start+e.Pos.Line, src)
		}
		i = j
	}
	if blocks < 40 {
		t.Errorf("found only %d ant blocks in the spec", blocks)
	}
}

// TestSyntaxErrors covers the syntax errors the specification names, one case per sentence.
func TestSyntaxErrors(t *testing.T) {
	cases := []struct {
		name, src, want string
		mode            Mode
	}{
		{"float leading dot", "var x = .5", "digit before the point", Program},
		{"float trailing dot", "var x = 1.", "digit after the point", Program},
		{"int overflow", "var x = 9223372036854775808", "does not fit an int", Program},
		{"octal-looking", "var x = 0o17", "malformed number", Program},
		{"underscore at end", "var x = 1_", "malformed number", Program},
		{"bad escape", `var s = "\q"`, "bad escape", Program},
		{"escape out of range", `var s = "\u{110000}"`, "not a Unicode scalar value", Program},
		{"escape surrogate", `var s = "\u{D800}"`, "not a Unicode scalar value", Program},
		{"newline in string", "var s = \"a\nb\"", "unterminated string", Program},
		{"unterminated comment", "/* open", "unterminated block comment", Program},
		{"non-ascii identifier", "var ñ = 1", "unexpected character", Program},
		{"reserved word", "var let = 1", "cannot be a variable name", Program},
		{"keyword variable", "var type = 1", "cannot be a variable name", Program},
		{"js function", "function f() {}", "'function' is a reserved word", Program},
		{"js let", "func main() { let x = 1 }", "'let' is a reserved word", Program},
		{"break without loop", "func f() { break }", "'break' outside a loop", Program},
		{"continue in lambda", "func f() { for x in xs { var g = { continue } } }", "'continue' outside a loop", Program},
		{"break in go", "func f() { for x in xs { go { break } } }", "'break' outside a loop", Program},
		{"return at top level", "return 1", "'return' outside a function", Session},
		{"return value in defer", "func f() { defer { return 1 } }", "'return' takes no value in a defer block", Program},
		{"return value in go", "func f() { go { return 1 } }", "'return' takes no value in a go block", Program},
		{"return value in go template", "func f() { go { print(`${return 1}`) } }", "'return' takes no value in a go block", Program},
		{"chained assignment", "func f() { a = b = 1 }", "assignment cannot be chained", Program},
		{"assign to call", "func f() { g() = 1 }", "left side of '='", Program},
		{"compound destructure", "func f() { [a, b] += [1, 2] }", "allows only '='", Program},
		{"rest in the middle", "func f() { [a, ...r, b] = xs }", "rest must be the last", Program},
		{"rest in the middle var", "var [a, ...r, b] = xs", "rest must be the last", Program},
		{"destructure literal position", "func f() { [a, 1] = xs }", "assignable places", Program},
		{"record destructuring assignment", "func f() { { a, b } = p }", "expected", Program},
		{"mix nullish", "var x = a ?? b && c", "cannot mix", Program},
		{"mix nullish after", "var x = a && b ?? c", "cannot mix", Program},
		{"pipe to non-call", "var x = a | b", "must be a call", Program},
		{"positional after named", "var x = f(a: 1, 2)", "cannot follow a named argument", Program},
		{"spread in map", "var m = Map{ ...other }", "no meaning in a map literal", Program},
		{"keyword shorthand map", "var m = Map{ type }", "needs ':'", Program},
		{"keyword shorthand destructure", "var { type } = e", "cannot be bound", Program},
		{"record string key", `var p = Point{ "x": 1 }`, "must be a name", Program},
		{"type in block", "func f() { type T = int }", "only allowed at the top level", Program},
		{"export in block", "func f() { export var x = 1 }", "only allowed at the top level", Program},
		{"named func in block", "func f() { func g() {} }", "only allowed at the top level", Program},
		{"import after declaration", "var x = 1\nimport \"./a.ant\" A", "imports must come before", Program},
		{"top-level expression", "print(1)", "expected a declaration", Program},
		{"two statements one line", "func f() { a b }", "expected a newline or ';'", Program},
		{"else alone", "func f() { else { 1 } }", "unexpected 'else'", Program},
		{"header missing brace", "func f() { if a b }", "expected '{' after the if header", Program},
		{"header literal", "func f() { if u == User{ name: \"a\" } { 1 } }", "must be parenthesized", Program},
		{"header map literal", "func f() { for e in Map{ \"k\": 1 } { 1 } }", "in the for header must be parenthesized", Program},
		{"trailing operator", "var x = 1 +", "unexpected", Program},
		{"func type missing return", "type F = func(int)", "expected ':'", Program},
		{"func with two dots", "func A.b.c() {}", "belongs to one type", Program},
		{"brace after literal", "func f() { 1{ } }", "expected a newline or ';'", Program},
		{"defer at top level", "defer { 1 }", "defer needs an enclosing block", Session},
		{"empty source with garbage", "@", "unexpected character", Program},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := Parse("test", c.src, c.mode)
			if err == nil {
				t.Fatalf("parsed, want an error containing %q", c.want)
			}
			if !strings.Contains(err.Error(), c.want) {
				t.Fatalf("error %q, want one containing %q", err.Error(), c.want)
			}
		})
	}
}

// TestAST checks the precedence and desugaring rules of section 5 on printed trees.
func TestAST(t *testing.T) {
	exprs := []struct{ src, want string }{
		{"a | f(b)", "(pipe f a b)"},
		{"a | f(b) | g()", "(pipe g (pipe f a b))"},
		{"try f(x) | g()", "(pipe g (try (call f x)))"},
		{"a && b || c && d", "(|| (&& a b) (&& c d))"},
		{"a || b && c", "(|| a (&& b c))"},
		{"x is int == true", "(== (is x int) true)"},
		{"!(x is Circle)", "(! (is x Circle))"},
		{"x is Geometry.Point", "(is x (. Geometry Point))"},
		{"x is func", "(is x func)"},
		{"a ?? b ?? c", "(?? (?? a b) c)"},
		{"-a * b + c < d", "(< (+ (* (- a) b) c) d)"},
		{"a - b - c", "(- (- a b) c)"},
		{"a % b / c", "(/ (% a b) c)"},
		{"Point{ x, y: 2 }", "(Point x:x y:2)"},
		{"Point{ ...p, x: 0.0 }", "(Point ...p x:0)"},
		{"Geometry.Point{ x: 1.0 }", "((. Geometry Point) x:1)"},
		{"Map{ type: 1, null: 2 }", `(Map "type":1 "null":2)`},
		{"Map{ name: 1, (name): 2, (1 + 2): 3, age }", `(Map "name":1 name:2 (+ 1 2):3 "age":age)`},
		{`Map{ "a": 1, 7: "seven" }`, `(Map "a":1 7:"seven")`},
		{"Map{}", "(Map)"},
		{"[1, ...xs, 2]", "[1 ...xs 2]"},
		{"{ x * 2 }", "(lambda (* x 2) )"},
		{"{}", "(lambda)"},
		{"{ { x } }", "(lambda (lambda x ) )"},
		{"func(a: int, b: int = 2): int { a + b }", "(func(a:int b:int=2):int { (+ a b) })"},
		{"a?.b?.[0]?.(1)", "(?.call (?.[] (?. a b) 0) 1)"},
		{"f(x)(y)[0].z", "(. ([] (call (call f x) y) 0) z)"},
		{"connect(port: 80, host: \"x\")", `(call connect port:80 host:"x")`},
		{"`a ${x + 1} b`", `(template "a " (+ x 1) " b")`},
		{"`${ Point{ x: 1.0 }.x }`", "(template (. (Point x:1) x))"},
		{"x ?? return y + 1", "(?? x (return (+ y 1)))"},
		{"if a { 1 } else if b { 2 } else { 3 }", "(if a { 1 } (if b { 2 } { 3 }))"},
		{"if c { Point{ x: 1.0, y: 0.0 } } else { origin }", "(if c { (Point x:1 y:0) } { origin })"},
		{"if u == (User{ name: \"ada\" }) { true }", `(if (== u (User name:"ada")) { true })`},
		{"if Array.some(xs, { x > 0 }) { 1 }", "(if (call (. Array some) xs (lambda (> x 0) )) { 1 })"},
		{"while !done { done = step() }", "(while (! done) { (= done (call step)) })"},
		{"for { key, value } in ages { key }", "(for {key value} ages { key })"},
		{"for [a, b] in pairs { a }", "(for [a b] pairs { a })"},
		{"for n in Iter.range(0, 9) | Iter.take(5) { n }", "(for n (pipe (. Iter take) (call (. Iter range) 0 9) 5) { n })"},
		{"[a, b] = [b, a]", "(= [a b] [b a])"},
		{"[xs[0], p.x, ...rest] = f()", "(= [([] xs 0) (. p x) ...rest] (call f))"},
		{"count += 1", "(+= count 1)"},
		{"go { work() }", "(go { (call work) })"},
	}
	for _, c := range exprs {
		e, err := ParseExpr(c.src)
		if err != nil {
			t.Errorf("%s: %v", c.src, err)
			continue
		}
		if got := ast.Sprint(e); got != c.want {
			t.Errorf("%s:\n got %s\nwant %s", c.src, got, c.want)
		}
	}

	files := []struct{ src, want string }{
		{"func f() {\n  a\n  (b)\n}", "(file (func f() { a; b }))"},
		{"func f() {\n  var x = 1; var y = 2; x + y\n}", "(file (func f() { (var x 1); (var y 2); (+ x y) }))"},
		{"type S = Circle\n  | Rect\n  | Empty", "(file (type S (| Circle Rect Empty)))"},
		{"type H = func(int): A | B", "(file (type H (func (int):(| A B))))"},
		{"type H = (func(int): A) | B", "(file (type H (| (func (int):A) B)))"},
		{"type P[A, B: Ordered] = { first: A, second: B }", "(file (type P[A B:Ordered] {first:A second:B}))"},
		{"type F = func(request: Request): Response", "(file (type F (func (request:Request):Response)))"},
		{"func f(a: bool): int {\n  if a { 1 }\n  else { 2 }\n}", "(file (func f(a:bool):int { (if a { 1 } { 2 }) }))"},
		{"func f() {\n  return\n  5\n}", "(file (func f() { (return); 5 }))"},
		{"func f() {\n  x = 1 +\n    2\n}", "(file (func f() { (= x (+ 1 2)) }))"},
		{"var t = orders\n  | Array.filter({ x.paid })\n  | Array.sum()", "(file (var t (pipe (. Array sum) (pipe (. Array filter) orders (lambda (. x paid) )))))"},
		{"var c = builder\n  .name\n  ?.first", "(file (var c (?. (. builder name) first)))"},
		{"var l = [\n  1,\n  2,  // a comment\n]", "(file (var l [1 2]))"},
		{"func P.make(): P { P{} }", "(file (func P.make():P { (P) }))"},
		{"export func Point.distance(a: Point, b: Point): float64 { 1.0 }", "(file (func export Point.distance(a:Point b:Point):float64 { 1 }))"},
		{"import \"./g.ant\" { distance: dist, Point }\nimport \"./g.ant\" G", "(file (import \"./g.ant\" {distance:dist Point}) (import \"./g.ant\" G))"},
		{"var [a, { b: [c, _] }] = v", "(file (var [a {b:[c _]}] v))"},
		{"var { address: { city } } = user", "(file (var {address:{city}} user))"},
		{"var [head, ...tail] = xs", "(file (var [head ...tail] xs))"},
		{"func f() {\n  defer { close() }\n  1\n}", "(file (func f() { (defer { (call close) }); 1 }))"},
		{"func f() {\n  a /* comment\n  spanning */ + b\n}", "(file (func f() { (+ a b) }))"},
		{"func f(): string {\n  `value: ${s ?? return \"missing\"}`\n}", "(file (func f():string { (template \"value: \" (?? s (return \"missing\"))) }))"},
		{"func f() {\n  for x in xs {\n    print(`${x ?? continue}`)\n  }\n}", "(file (func f() { (for x xs { (call print (template (?? x (continue)))) }) }))"},
	}
	for _, c := range files {
		f, err := Parse("test", c.src, Program)
		if err != nil {
			t.Errorf("%q: %v", c.src, err)
			continue
		}
		if got := ast.Sprint(f); got != c.want {
			t.Errorf("%q:\n got %s\nwant %s", c.src, got, c.want)
		}
	}
}

// TestHasDefer checks that defer marks the block it registers with.
func TestHasDefer(t *testing.T) {
	f, err := Parse("test", "func f() {\n  if c {\n    defer { x() }\n  }\n  y()\n}", Program)
	if err != nil {
		t.Fatal(err)
	}
	body := f.Decls[0].(*ast.FuncDecl).Body
	if body.HasDefer {
		t.Error("the function body should not have the defer")
	}
	inner := body.Stmts[0].(*ast.If).Then
	if !inner.HasDefer {
		t.Error("the if block should have the defer")
	}
}

// TestNesting checks the nesting limit: a thousand levels parse, more are a syntax error rather
// than a stack overflow, in expressions, types and blocks alike.
func TestNesting(t *testing.T) {
	deep := func(open, close string, n int) string {
		return strings.Repeat(open, n) + "1" + strings.Repeat(close, n)
	}
	if _, err := Parse("t", "var x = "+deep("(", ")", 900), Program); err != nil {
		t.Errorf("900 levels: %v", err)
	}
	cases := []string{
		"var x = " + deep("(", ")", 1100),
		"var x = " + deep("[", "]", 1100),
		"var x = " + strings.Repeat("-", 1100) + "1",
		"var x: " + strings.Repeat("Array[", 1100) + "int" + strings.Repeat("]", 1100) + " = 1",
		"func f(): void " + strings.Repeat("{", 1100) + strings.Repeat("}", 1100),
		"var s = " + strings.Repeat("`${", 1100) + "1" + strings.Repeat("}`", 1100),
	}
	for _, src := range cases {
		_, err := Parse("t", src, Program)
		if err == nil || !strings.Contains(err.Error(), "nesting deeper than") {
			t.Errorf("%.30q: got %v, want a nesting error", src, err)
		}
	}
}
