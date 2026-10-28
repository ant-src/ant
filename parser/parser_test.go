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

// pendingBlocks lists the ```ant blocks of the specification the parser cannot read yet, each by
// its first line. A pending block that parses is an error, so the list shrinks as the parser
// grows.
var pendingBlocks = []string{}

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
		pending := false
		for _, first := range pendingBlocks {
			if start < len(lines) && strings.TrimSpace(lines[start]) == first {
				pending = true
			}
		}
		src := strings.Join(lines[start:j], "\n")
		_, err := Parse("spec", src, Spec)
		switch {
		case err != nil && !pending:
			e := err.(*Error)
			t.Errorf("block at spec line %d: %s (spec line %d)\n%s", start+1, e.Msg, start+e.Pos.Line, src)
		case err == nil && pending:
			t.Errorf("block at spec line %d parses now; remove it from pendingBlocks", start+1)
		}
		i = j
	}
	if blocks < 40 {
		t.Errorf("found only %d ant blocks in the spec", blocks)
	}
}

// TestSyntaxErrors covers the syntax errors the specification names, one case per sentence.
type syntaxCase struct {
	name, src, want string
	mode            Mode
}

func TestSyntaxErrors(t *testing.T) {
	for _, c := range syntaxCases() {
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

// syntaxCases are the syntax errors the specification names, one case per sentence.
func syntaxCases() []syntaxCase {
	return []syntaxCase{
		{"float leading dot", "var x = .5", "digit before the point", Program},
		{"module doc after code", "var x = 1\n//! late", "before the first import or declaration", Program},
		{"module doc in a body", "func f() {\n  //! inside\n}", "before the first import or declaration", Program},
		{"module doc after a blank line", "//! a\n\n//! b\nvar x = 1", "must be consecutive", Program},
		{"module doc after a comment", "//! a\n// b\n//! c\nvar x = 1", "must be consecutive", Program},
		{"module doc not alone", "/* c */ //! a\nvar x = 1", "alone on its line", Program},
		{"regex unterminated", "var r = /abc", "unterminated regular expression", Program},
		{"regex newline", "var r = /ab\ncd/", "unterminated regular expression", Program},
		{"regex unknown flag", "var r = /a/x", "unknown flag", Program},
		{"regex repeated flag", "var r = /a/ii", "the flag i is repeated", Program},
		{"regex flag g", "var r = /a/g", "there is no g flag", Program},
		{"regex flag y", "var r = /a/y", "there is no y flag", Program},
		{"regex flag u", "var r = /a/u", "there is no u flag", Program},
		{"regex lookahead", "var r = /(?=a)/", "lookahead is not supported", Program},
		{"regex lookbehind", "var r = /(?<!a)/", "lookbehind is not supported", Program},
		{"regex backreference", "var r = /(a)\\1/", "backreferences are not supported", Program},
		{"regex unicode escape", "var r = /\\u0041/", "write a code point as", Program},
		{"regex unbalanced", "var r = /(/", "missing closing )", Program},
		{"select empty", "func f() { select {} }", "at least one clause with a head", Program},
		{"select else only", "func f() { select { else { 1 } } }", "at least one clause with a head", Program},
		{"select two else", "func f() { select { Channel.receive(c) { 1 }\n else { 2 }\n else { 3 } } }", "at most one else clause", Program},
		{"select else not last", "func f() { select { else { 2 }\n Channel.receive(c) { 1 } } }", "must be the last clause", Program},
		{"select bind else", "func f() { select { var x = else { 2 } } }", "an else clause binds nothing", Program},
		{"select bad head", "func f() { select { g(c) { 1 } } }", "head must be a call of", Program},
		{"select pipe head", "func f() { select { c | Channel.receive() { 1 } } }", "head must be a call of", Program},
		{"select brace on next line", "func f() { select { Channel.receive(c)\n { 1 } } }", "expected '{' after the select clause header", Program},
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
		{"reserved parameter", "func f(class: int) {}", "'class' is a reserved word and cannot be a parameter name", Program},
		{"keyword type name", "type if = int", "'if' is a keyword and cannot be a type name", Program},
		{"js function", "function f() {}", "'function' is a reserved word", Program},
		{"js let", "func main() { let x = 1 }", "'let' is a reserved word", Program},
		{"break without loop", "func f() { break }", "'break' outside a loop", Program},
		{"continue in lambda", "func f() { for x in xs { var g = { continue } } }", "'continue' outside a loop", Program},
		{"break in go", "func f() { for x in xs { go { break } } }", "'break' outside a loop", Program},
		{"return at top level", "return 1", "'return' outside a function", Session},
		{"return value in defer", "func f() { defer { return 1 } }", "'return' takes no value in a defer block", Program},
		{"return value in go", "func f() { go { return 1 } }", "'return' takes no value in a go block", Program},
		{"return value in go template", "func f() { go { println(`${return 1}`) } }", "'return' takes no value in a go block", Program},
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
		{"reserved shorthand destructure", "var { class } = e", "'class' is a reserved word and cannot be bound", Program},
		{"reserved shorthand map", "var m = Map{ class }", "'class' is a reserved word: as a key", Program},
		{"record string key", `var p = Point{ "x": 1 }`, "must be a name", Program},
		{"export in block", "func f() { export var x = 1 }", "only allowed at the top level", Program},
		{"named func in block", "func f() { func g() {} }", "only allowed at the top level", Program},
		{"import after declaration", "var x = 1\nimport \"./a.ant\" A", "imports must come before", Program},
		{"export in a script", "export func f() {}\nprintln(1)", "a module with exports cannot have top-level statements", Program},
		{"try in a library initializer", "var x = try f()", "'try' outside a function", Program},
		{"return in a library initializer", "var x = 1\nvar y = x ?? return 2", "'return' outside a function", Program},
		{"top-level expression in declarations", "println(1)", "expected a declaration", Declarations},
		{"two statements one line", "func f() { a b }", "expected a newline or ';'", Program},
		{"leading or", "func f() { var t = a\n  || b }", "unexpected '||'", Program},
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
		{"field default in an annotation", "var a: { port: int = 80 } = b", "belongs in the record body of a type declaration", Program},
		{"field default in a nested record type", "type A = Array[{ n: int = 1 }]", "belongs in the record body of a type declaration", Program},
		{"field default in a union member", "type U = { n: int = 1 } | null", "not in a union member", Program},
		{"field default in a function type", "func f(g: func({ n: int = 1 }): int) {}", "belongs in the record body of a type declaration", Program},
		{"field default in declarations", "type Opts = { n: int = 1 }", "a declarations file takes no field defaults", Declarations},
		{"return in a field default", "type T = { n: int = return 1 }", "'return' in a field default", Program},
		{"try in a field default", "func f(): Result[int, Error] {\n  type T = { n: int = try g() }\n  Result.ok(1)\n}", "'try' in a field default", Program},
		{"try in a field default in a script", "type T = { n: int = try g() }\nprintln(1)", "'try' in a field default", Program},
		{"break in a field default inside a loop", "for i in xs {\n  type T = { n: int = if c { break } else { 1 } }\n}", "'break' outside a loop", Program},
		{"continue in a field default inside a loop", "while c {\n  type T = { n: int = if c { continue } else { 1 } }\n}", "'continue' outside a loop", Program},
		{"defer in a field default", "func f() {\n  type T = { n: int = defer { 1 } }\n}", "defer needs an enclosing block", Program},
		{"field default needs an expression", "type T = { n: int = }", "unexpected", Program},
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
		{"func f() {\n  type T = { v: int }\n  T{ v: 1 }\n}", "(file (func f() { (type T {v:int}); (T v:1) }))"},
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
		{"func f() {\n  for x in xs {\n    println(`${x ?? continue}`)\n  }\n}", "(file (func f() { (for x xs { (call println (template (?? x (continue)))) }) }))"},
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
// TestFieldDefaults checks what the parser accepts in a field default: any expression, with
// loops and blocks of its own, and a lambda whose return and try belong to the lambda.
func TestFieldDefaults(t *testing.T) {
	ok := []string{
		"type T = { a: int, b: int = 1, c: Array[string] = [], d: string | null = null }",
		"type T = {\n  a: int = 1,\n  b: int = 2,\n}",
		"type Box[T] = { items: Array[T] = [] }",
		"type T = { n: int = if c { 1 } else { 2 } }",
		"type T = { n: int = { var s = 0\n  for i in xs { if i > 2 { break }\n s += i }\n s }() }",
		"type T = { f: func(): int = { return 1 } }",
		"type T = { f: func(): Result[int, Error] = { Result.ok(try g()) } }",
		"type T = { n: int = if c { defer { 1 }\n 2 } else { 3 } }",
		"type U = { n: int } | null",
		"func f() {\n  type T = { n: int = 1 }\n}",
	}
	for _, src := range ok {
		if _, err := Parse("test", src, Program); err != nil {
			t.Errorf("%q: %v", src, err)
		}
	}
	f, err := Parse("test", "type T = { a: int, b: int = 1 + 2 }", Program)
	if err != nil {
		t.Fatal(err)
	}
	rt := f.Decls[0].(*ast.TypeDecl).Body.(*ast.RecordType)
	if rt.Fields[0].Default != nil {
		t.Errorf("a has a default: %s", ast.Sprint(rt.Fields[0].Default))
	}
	if got := ast.Sprint(rt.Fields[1].Default); got != "(+ 1 2)" {
		t.Errorf("b's default = %s", got)
	}
	if got := ast.Sprint(rt); got != "{a:int b:int=(+ 1 2)}" {
		t.Errorf("record type = %s", got)
	}
}

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

// TestBlockTypes checks that a block, a lambda body included, lists the types declared directly
// in it and no others.
func TestBlockTypes(t *testing.T) {
	f, err := Parse("test", "func f() {\n  type A = int\n  if c { type B = int }\n  g({ type C = int; C(1) })\n}", Program)
	if err != nil {
		t.Fatal(err)
	}
	body := f.Decls[0].(*ast.FuncDecl).Body
	if len(body.Types) != 1 || body.Types[0].Name != "A" {
		t.Errorf("the function body should list A alone, got %v", body.Types)
	}
	if then := body.Stmts[1].(*ast.If).Then; len(then.Types) != 1 || then.Types[0].Name != "B" {
		t.Errorf("the if block should list B alone, got %v", then.Types)
	}
	lam := body.Stmts[2].(*ast.Call).Args[0].Value.(*ast.Lambda)
	if len(lam.Body.Types) != 1 || lam.Body.Types[0].Name != "C" {
		t.Errorf("the lambda body should list C alone, got %v", lam.Body.Types)
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
	if _, err := Parse("t", "var x = 1"+strings.Repeat(" + 1", 900), Program); err != nil {
		t.Errorf("900 links: %v", err)
	}
	cases := []string{
		"var x = " + deep("(", ")", 1100),
		"var x = " + deep("[", "]", 1100),
		"var x = " + strings.Repeat("-", 1100) + "1",
		"var x: " + strings.Repeat("Array[", 1100) + "int" + strings.Repeat("]", 1100) + " = 1",
		"func f(): void " + strings.Repeat("{", 1100) + strings.Repeat("}", 1100),
		"var s = " + strings.Repeat("`${", 1100) + "1" + strings.Repeat("}`", 1100),
		// the depth carries into an interpolation instead of starting again at zero
		"var s = " + strings.Repeat("`${"+strings.Repeat("(", 600), 2) + "1" + strings.Repeat(strings.Repeat(")", 600)+"}`", 2),
		"var x = if a { 1 }" + strings.Repeat(" else if a { 1 }", 1100),
		"var x = 1" + strings.Repeat(" + 1", 1100),
		"var x = a" + strings.Repeat(" ?? a", 1100),
		"var x = a" + strings.Repeat(" < a", 1100),
		"var x = a" + strings.Repeat(" | f()", 1100),
		"var x = f" + strings.Repeat("()", 1100),
		"var x = a" + strings.Repeat(".b", 1100),
		"var x = a" + strings.Repeat("[0]", 1100),
		"var x = a is b" + strings.Repeat(".c", 1100),
		"var " + strings.Repeat("[", 1100) + "a" + strings.Repeat("]", 1100) + " = xs",
		"var " + strings.Repeat("{ a: ", 1100) + "b" + strings.Repeat(" }", 1100) + " = r",
		"func f() { for " + strings.Repeat("[", 1100) + "a" + strings.Repeat("]", 1100) + " in xs {} }",
	}
	for _, src := range cases {
		_, err := Parse("t", src, Program)
		if err == nil || !strings.Contains(err.Error(), "nesting deeper than") {
			t.Errorf("%.30q: got %v, want a nesting error", src, err)
		}
	}
}

// TestDocComments attaches doc comments by the rule of section 2.
func TestDocComments(t *testing.T) {
	src := `// Package text before a blank line documents nothing.

// Point is a position.
//
//     Point{ x: 1.0 }
export type Point = {
  x: float64,
}

var a = 1 // trailing, not a doc
func f(): int { 1 }

/* block */
// only the line comment after the block
func g(x: int) { x }

// Origin is where the axes meet.
var origin: Point = Point{ x: 0.0 }

//no space
//   kept indent
func Point.norm(p: Point): float64 { p.x }

// inside
func h() {
  // a comment inside h
}
func k() {}

// only a
var a2 = 1; var b2 = 2
`
	f, err := Parse("t.ant", src, Program)
	if err != nil {
		t.Fatal(err)
	}
	want := []struct{ doc, sig string }{
		{"Point is a position.\n\n    Point{ x: 1.0 }", "type Point = {\n  x: float64,\n}"},
		{"", "var a"},
		{"", "func f(): int"},
		{"only the line comment after the block", "func g(x: int)"},
		{"Origin is where the axes meet.", "var origin: Point"},
		{"no space\n  kept indent", "func Point.norm(p: Point): float64"},
		{"inside", "func h()"},
		{"", "func k()"},
		{"only a", "var a2"},
		{"", "var b2"},
	}
	if len(f.Decls) != len(want) {
		t.Fatalf("got %d declarations", len(f.Decls))
	}
	for i, d := range f.Decls {
		var doc, sig string
		switch d := d.(type) {
		case *ast.VarDecl:
			doc, sig = d.Doc, d.Sig
		case *ast.FuncDecl:
			doc, sig = d.Doc, d.Sig
		case *ast.TypeDecl:
			doc, sig = d.Doc, d.Sig
		}
		if doc != want[i].doc || sig != want[i].sig {
			t.Errorf("declaration %d: doc %q sig %q, want %q %q", i, doc, sig, want[i].doc, want[i].sig)
		}
	}
}

// TestModuleDoc reads the //! lines before the first import as the module doc (section 2), with
// other comments around them, and keeps them out of the first declaration's doc.
func TestModuleDoc(t *testing.T) {
	src := "// Copyright.\n\n//! Geometry computes distances.\n//!\n//!     code\n// Point is a point.\nexport type Point = { x: int }\n"
	f, err := Parse("t.ant", src, Program)
	if err != nil {
		t.Fatal(err)
	}
	if want := "Geometry computes distances.\n\n    code"; f.Doc != want {
		t.Errorf("module doc %q, want %q", f.Doc, want)
	}
	if d := f.Decls[0].(*ast.TypeDecl); d.Doc != "Point is a point." {
		t.Errorf("declaration doc %q", d.Doc)
	}
	f, err = Parse("t.ant", "//! Only a doc.\nvar x = 1\n", Program)
	if err != nil {
		t.Fatal(err)
	}
	if f.Doc != "Only a doc." || f.Decls[0].(*ast.VarDecl).Doc != "" {
		t.Errorf("module doc %q, declaration doc %q", f.Doc, f.Decls[0].(*ast.VarDecl).Doc)
	}
}

// TestRestComma rejects a comma after the rest position of a destructuring assignment, as it is
// rejected in a var, while an array literal ending in a spread may take one.
func TestRestComma(t *testing.T) {
	for _, src := range []string{"func f() { [a, ...b,] = xs }", "func f() { var [a, ...b,] = xs }"} {
		if _, err := Parse("t", src, Program); err == nil || !strings.Contains(err.Error(), "rest") {
			t.Errorf("%s: got %v, want an error about the rest", src, err)
		}
	}
	if _, err := Parse("t", "func f() { var c = [a, ...b,] }", Program); err != nil {
		t.Errorf("a spread with a trailing comma: %v", err)
	}
}
