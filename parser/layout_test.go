package parser

import (
	"strings"
	"testing"

	"github.com/ant-src/ant/internal/token"
)

// roleString lists the opening brackets of src in order, each with its role: B for a block, T
// for a typed literal's brace, a trailing , when a comma may follow the last item, and - for a
// bracket the parser recorded no role for.
func roleString(t *testing.T, src string) string {
	t.Helper()
	_, l, err := ParseLayout("t.ant", src)
	if err != nil {
		t.Fatalf("%q: %v", src, err)
	}
	var out []string
	var walk func([]token.Token)
	walk = func(toks []token.Token) {
		for _, tok := range toks {
			for _, part := range tok.Parts {
				walk(part.Tokens)
			}
			if !tok.Is("(") && !tok.Is("[") && !tok.Is("{") {
				continue
			}
			r, ok := l.Roles[tok.Off]
			s := tok.Text
			switch {
			case !ok:
				s += "-"
			case r.Block:
				s += "B"
			case r.Typed:
				s += "T"
			}
			if r.Comma {
				s += ","
			}
			out = append(out, s)
		}
	}
	walk(l.Tokens)
	return strings.Join(out, " ")
}

func TestRoles(t *testing.T) {
	for _, c := range []struct{ src, want string }{
		// Section 5, Blocks: a brace after a type name is a literal, after func(...), else, go,
		// defer, a header, or a select clause's head it is a block, and anywhere else a lambda.
		{"var p = Point{ x: 1 }", "{T,"},
		{"var m = Map{}", "{T,"},
		{"var p = Geo.Point{ x: 1 }", "{T,"},
		{"func f() { if a { b } else { c } }", "(, {B {B {B"},
		{"func f() { go { a }; defer { b } }", "(, {B {B {B"},
		{"func f() { while ready { a } }", "(, {B {B"},
		{"func f() { for x in xs { a } }", "(, {B {B"},
		{"var f = func(a: int) { a }", "(, {B"},
		{"var g = Array.map(xs, { x * 2 })", "(, {B"},
		{"func f() { select { Context.wait(c) { a } else { b } } }", "(, {B {B (, {B {B"},
		{"func f() { if u == (User{ name: 1 }) { a } }", "(, {B ( {T, {B"},
		// Lists that take a comma after the last item, and those that do not.
		{"func f[T: Ordered](a: T, b: int = 1): Array[T] { [a] }", "[, (, [, {B [,"},
		{"type P = { x: int, f: func(int): int }", "{, (,"},
		{"type U = (A | B)", "("},
		{"var x = xs[i]", "["},
		{"var x = (a + b)", "("},
		{"var y = a?.[0]", "["},
		{"var z = f?.(1)", "(,"},
		{"var [a, b] = xs", "[,"},
		{"var [a, ...rest] = xs", "["},
		{"var { x, y: [p, ...q] } = r", "{, ["},
		{"func f() { [a, b] = [b, a] }", "(, {B [, [,"},
		{"var s = `${Point{ x: 1 }} ${f(a)}`", "{T, (,"},
	} {
		if got := roleString(t, c.src); got != c.want {
			t.Errorf("%s\n got %s\nwant %s", c.src, got, c.want)
		}
	}
	if got := roleString(t, "import \"./a.ant\" { a, b: c }\nfunc main() {}"); got != "{, (, {B" {
		t.Errorf("import list: got %s", got)
	}
}

// TestParseRecordsNoRoles checks that an ordinary parse leaves the role map off.
func TestParseRecordsNoRoles(t *testing.T) {
	p := &parser{mode: Program}
	p.mark(token.Token{Off: 1}, Role{Block: true})
	if p.roles != nil {
		t.Error("mark filled a nil role map")
	}
}

// TestPrefix checks that ParseLayout marks the - and ! it reads as prefix operators, and only
// them: a - after a line break inside a list or an interpolation still subtracts.
func TestPrefix(t *testing.T) {
	for _, c := range []struct{ src, want string }{
		{"var x = -a - -b", "- -"},
		{"var x = !a && !(b)", "! !"},
		{"func f() { a\n-1 }", "-"},
		{"var x = (a\n- 1)", ""},
		{"var x = [a,\n-1]", "-"},
		{"var x = `${a\n  - b} ${-c}`", "-"},
		{"var x = try f()", ""},
	} {
		_, l, err := ParseLayout("t.ant", c.src)
		if err != nil {
			t.Fatalf("%q: %v", c.src, err)
		}
		var got []string
		for off := range len(c.src) {
			if l.Prefix[off] {
				got = append(got, c.src[off:off+1])
			}
		}
		if strings.Join(got, " ") != c.want {
			t.Errorf("%q: prefix %q, want %q", c.src, strings.Join(got, " "), c.want)
		}
	}
}
