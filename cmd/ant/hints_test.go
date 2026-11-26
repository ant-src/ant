package main

import (
	"encoding/json"
	"strings"
	"testing"
)

// TestCheckHelp checks the help and fixes ant check adds to the errors whose repair it can name
// (spec/check.md section 2, Help and fixes), as -json writes them.
func TestCheckHelp(t *testing.T) {
	cases := []struct {
		name, src string
		help      string // a part of the help
		fix       string // the title of the only fix, or ""
		auto      bool
		applied   string // the source with the fix applied, when there is one
	}{
		{"namespace function", "let s = [\"a\"].join(\",\")\n", "join is String.join, which takes the array first",
			"String.join(..., ...)", true, "let s = String.join([\"a\"], \",\")\n"},
		{"no arguments", "let b = \"x\".toBytes()\nlet h = b.toHex()\n", "toHex is Encoding.toHex, which takes the Bytes first",
			"Encoding.toHex(b)", true, "let b = \"x\".toBytes()\nlet h = Encoding.toHex(b)\n"},
		// the receiver's start is inside its parentheses, so no edit there can be right
		{"receiver in parentheses", "let xs = [\"a\"]\nlet s = (xs).join(\",\")\n", "join is String.join", "", false, ""},
		{"another type's function", "let xs = [3, 1]\nlet n = xs.fields()\n", "", "", false, ""},
		// Time.with takes a T bounded by Time.Wall, which a record is not
		{"outside the bound", "type P = { a: int }\nlet q = P{ a: 1 }.with(a: 2)\n", "", "", false, ""},
		{"misspelled function", "let n = \"ab\".lenght()\n", "did you mean length?", "", false, ""},
		{"misspelled field", "type U = { name: string }\nlet u = U{ name: \"a\" }\nprintln(u.nmae)\n", "did you mean name?", "", false, ""},
		{"misspelled literal field", "type P = { x: int }\nlet p = P{ x: 1, yy: 2 }\n", "", "", false, ""},
		{"misspelled name", "let counter = 1\nprintln(countr)\n", "did you mean counter?", "", false, ""},
		{"namespace member", "let t = Math.sqr(4.0)\n", "did you mean sqrt?", "", false, ""},
		{"namespace listing", "let m = Math.mod(5.0, 2.0)\n", "ant doc Math lists its members", "", false, ""},
		{"Result.ok of void", "let r: Result[void, Error] = Result.ok()\n", "the value of a Result[void, E] is void",
			"Result.ok(void)", false, "let r: Result[void, Error] = Result.ok(void)\n"},
		{"Result.ok of a value", "let r: Result[int, Error] = Result.ok()\n", "Result.ok takes the value", "Result.ok(void)", false, ""},
		{"cast", "let n = 1 as int\n", "Ant has no casts", "", false, ""},
		{"reserved name", "let match = 1\n", "choose another name", "", false, ""},
		{"array type", "func f(xs: int[]): int { 0 }\n", "an array of int is Array[int]", "", false, ""},
		{"empty type parameters", "func id[](x: int): int { x }\n", "leave out the brackets", "", false, ""},
		{"subtest name", "test \"t\" {\n  for c in [\"a\"] {\n    test c {}\n  }\n}\n", "test \"${c}\"", "", false, ""},
		{"field path", "type A = { n: int }\ntype B = { s: string }\ntype R = { x: A | B }\nfunc f(r: R): int {\n  if r.x is A { r.x.n } else { 0 }\n}\n",
			"bind it first, as let v = r.x", "", false, ""},
		{"union left out", "type X = {}\ntype Y = {}\ntype Z = {}\nfunc f(v: X | Y | Z): int {\n  if v is X { 1 } else if v is Y { 2 }\n}\n",
			"the tests leave out Z: add else if v is Z", "", false, ""},
		{"prelude name", "type Entry = { a: int }\n", "ant doc Entry shows the prelude's Entry", "", false, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			name := "main.ant"
			if strings.HasPrefix(c.src, "test ") {
				name = "main_test.ant"
			}
			dir := tree(t, map[string]string{name: c.src})
			_, out, _ := checkRun("-json", dir)
			var d jsonDiagnostic
			if err := json.Unmarshal([]byte(strings.SplitN(out, "\n", 2)[0]), &d); err != nil {
				t.Fatalf("%v: %q", err, out)
			}
			if !strings.Contains(d.Help, c.help) {
				t.Errorf("help %q, want one containing %q (%s)", d.Help, c.help, d.Message)
			}
			if c.fix == "" {
				if len(d.Fixes) > 0 && c.name != "Result.ok of a value" {
					t.Errorf("fixes %+v, want none", d.Fixes)
				}
				return
			}
			if len(d.Fixes) != 1 || d.Fixes[0].Auto != c.auto {
				t.Fatalf("fixes %+v, want one, auto %v", d.Fixes, c.auto)
			}
			if c.applied != "" {
				if got := applyEdits(c.src, d.Fixes[0].Edits); got != c.applied {
					t.Errorf("applied:\n%s\nwant:\n%s", got, c.applied)
				}
			}
		})
	}
}

// applyEdits applies edits of one file to src, the last first, so that each position still
// counts in the source as it was.
func applyEdits(src string, edits []jsonEdit) string {
	lines := strings.SplitAfter(src, "\n")
	offset := func(p jsonPos) int {
		n := 0
		for i := 0; i < p.Line-1; i++ {
			n += len(lines[i])
		}
		return n + p.Column - 1
	}
	for i := len(edits) - 1; i >= 0; i-- {
		e := edits[i]
		src = src[:offset(e.Start)] + e.Text + src[offset(e.End):]
	}
	return src
}
