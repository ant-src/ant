package tests

import (
	"strings"
	"testing"
)

// TestDocsHarness runs pageProblems on small pages, so the harness that holds the docs to their
// contract is itself tested: every page that should fail names the problem, and every page that
// should pass does.
func TestDocsHarness(t *testing.T) {
	const f = "```"
	draft := "<!-- status: draft -->\n\n"
	cases := []struct {
		name string
		rel  string // default probe/p.md
		page string // after the status line, unless raw
		raw  bool
		want []string // each must be in some problem; none means the page passes
	}{
		// Pages that pass.
		{name: "program with output", page: f + "ant\nprintln(\"hi\")\n" + f + "\n\n" + f + "text\nhi\n" + f + "\n"},
		{name: "output without a final line break", page: f + "ant\nprint(\"x\")\n" + f + "\n\n" + f + "text\nx\n" + f + "\n"},
		{name: "cases", page: f + "ant\nArray.map([1, 2], { x * 2 }) // [2, 4]\nvar xs = [3, 1]\nArray.sort(xs) // [1, 3]\n" + f + "\n"},
		{name: "a // in a string", page: f + "ant\nprintln(\"a // b\")\n" + f + "\n\n" + f + "text\na // b\n" + f + "\n"},
		{name: "a case whose value holds //", page: f + "ant\n\"a // b\" // \"a // b\"\n" + f + "\n"},
		{name: "comments after statements", page: f + "ant\nvar xs = [1, 2]\nxs[0 + 1] = 5 // set\nfor x in xs { // each\n  println(x)\n}\n" + f + "\n"},
		{name: "files, stdin, args", page: "<!-- test: file people.csv\nname\nAnn\n-->\n<!-- test: stdin\nhello\n-->\n<!-- test: args a b -->\n\n" + f + "ant\nprintln(Os.args())\nprintln(Result.unwrap(File.read(\"people.csv\")))\nprintln(Result.unwrapOr(File.readLine(Os.stdin()), null))\n" + f + "\n\n" + f + "text\n[\"a\", \"b\"]\nname\nAnn\n\nhello\n" + f + "\n"},
		{name: "panics and error", page: "<!-- test: panics index 5 out of range -->\n\n" + f + "ant\nvar ys = [1]\nprintln(ys[5])\n" + f + "\n\n<!-- test: error cannot use -->\n\n" + f + "ant\nvar x: int = \"s\"\n" + f + "\n"},
		{name: "modules and tests", page: f + "ant [lib/geo.ant]\nexport func double(n: int): int { n * 2 }\n" + f + "\n\n" + f + "ant\nimport \"./lib/geo.ant\" { double }\n\nprintln(double(21))\n" + f + "\n\n" + f + "text\n42\n" + f + "\n\n<!-- test: cmd ant test . -->\n\n" + f + "ant [geo_test.ant]\nimport \"./lib/geo.ant\" { double }\n\ntest \"double\" {\n  assert(double(2) == 4)\n}\n" + f + "\n"},
		{name: "check-only", page: "<!-- test: check-only prints the time -->\n\n" + f + "ant\nprintln(Time.now())\n" + f + "\n"},
		{name: "indented fence", page: "1. Run it:\n\n   " + f + "ant\n   println(1)\n   " + f + "\n\n   " + f + "text\n   1\n   " + f + "\n"},
		{name: "odd but local file name", page: "<!-- test: file a..b.txt\nx\n-->\n\n" + f + "ant\nprintln(Result.unwrap(File.read(\"a..b.txt\")))\n" + f + "\n"},
		{name: "links", page: "[same](p.md) [other](other.md \"title\") [site](/vscode) [web](https://ant-lang.org) `[code](nowhere.md)`\n"},
		{name: "readme needs no status", rel: "README.md", raw: true, page: "# Docs\n"},

		// Pages that fail.
		{name: "no status", raw: true, page: "# p\n", want: []string{"p.md:1: the first line must be"}},
		{name: "wrong output", page: f + "ant\nprintln(\"hi\")\n" + f + "\n\n" + f + "text\nbye\n" + f + "\n", want: []string{"the output differs"}},
		{name: "wrong case", page: f + "ant\nArray.map([1, 2], { x * 2 }) // [2, 5]\n" + f + "\n", want: []string{"p.md:4:", "want: [2, 5]"}},
		{name: "wrong multi-line case", page: f + "ant\nArray.map([1, 2], {\n  x * 2\n}) // [999]\n" + f + "\n", want: []string{"[999]"}},
		{name: "wrong type expression case", page: f + "ant\ntype(int) // nonsense\n" + f + "\n", want: []string{"nonsense"}},
		{name: "case line after hidden setup", page: "<!-- test: before\nvar a = 1\nvar b = 2\nvar c = 3\n-->\n\n" + f + "ant\na + b + c // 7\n" + f + "\n", want: []string{"p.md:10:"}},
		{name: "check error at its page line", page: f + "ant\nprintln(1)\nvar x: int = \"s\"\n" + f + "\n", want: []string{"p.md:5:14: cannot use string"}},
		{name: "unformatted", page: f + "ant\nvar  x = 1\n" + f + "\n", want: []string{"not formatted"}},
		{name: "broken indented fence", page: "1. Run it:\n\n   " + f + "ant\n   this is (not ant\n   " + f + "\n", want: []string{"p.md:"}},
		{name: "signature outside the generated part", page: f + "ant signature\nthis is (not ant\n" + f + "\n", want: []string{"generated part only"}},
		{name: "output after a module", page: f + "ant [lib.ant]\nexport func f(): int { 1 }\n" + f + "\n\n" + f + "text\nWRONG\n" + f + "\n", want: []string{"does not run has no output"}},
		{name: "run directives on a module", page: "<!-- test: exit 7 -->\n\n" + f + "ant [lib.ant]\nexport func f(): int { 1 }\n" + f + "\n", want: []string{"test: exit does not apply"}},
		{name: "test file is checked", page: f + "ant [x_test.ant]\nvar x: int = \"no\"\n" + f + "\n", want: []string{"ant check"}},
		{name: "error on a data file", page: "<!-- test: error x -->\n\n" + f + "ant [data.txt]\nhello\n" + f + "\n", want: []string{"test: error does not apply"}},
		{name: "exit needs a number", page: "<!-- test: exit one -->\n\n" + f + "ant\nprintln(1)\n" + f + "\n", want: []string{"needs a number"}},
		{name: "wrong exit", page: "<!-- test: exit 3 -->\n\n" + f + "ant\nprintln(1)\n" + f + "\n", want: []string{"exits with 0, not 3"}},
		{name: "no panic", page: "<!-- test: panics boom -->\n\n" + f + "ant\nprintln(1)\n" + f + "\n", want: []string{"does not panic"}},
		{name: "no error", page: "<!-- test: error cannot use -->\n\n" + f + "ant\nprintln(1)\n" + f + "\n", want: []string{"ant check passes"}},
		{name: "links that name nothing", page: "[a](nope.md \"title\") [b](P.md) [c](../probe) [d](../../outside.md)\n", want: []string{"nope.md names no file", "P.md names no file", "../probe names no file", "outside.md names no file"}},
		{name: "mistyped directives", page: "<!-- tests: exit 3 -->\n<!-- Test: exit 3 -->\n\n" + f + "ant\nprintln(1)\n" + f + "\n", want: []string{"p.md:3: a directive starts with test: exactly", "p.md:4: a directive starts"}},
		{name: "unknown directive", page: "<!-- test: frobnicate -->\n\n" + f + "ant\nprintln(1)\n" + f + "\n", want: []string{"unknown directive"}},
		{name: "body on the head line", page: "<!-- test: stdin hello -->\n\n" + f + "ant\nprintln(1)\n" + f + "\n", want: []string{"takes no argument"}},
		{name: "duplicate directive", page: "<!-- test: before\nvar a = 1\n-->\n<!-- test: before\nvar b = 2\n-->\n\n" + f + "ant\nprintln(1)\n" + f + "\n", want: []string{"a second test: before"}},
		{name: "directive before no ant block", page: "<!-- test: stdin\nx\n-->\n\n" + f + "sh\nls\n" + f + "\n\n<!-- test: args a -->\n", want: []string{"not an ant block", "before no ant block"}},
		{name: "file outside the directory", page: "<!-- test: file ../x\nx\n-->\n\n" + f + "ant\nprintln(1)\n" + f + "\n", want: []string{"relative path inside"}},
		{name: "unclosed comment", page: "<!-- note\n\n" + f + "ant\nthis is (not ant\n" + f + "\n", want: []string{"comment is not closed"}},
		{name: "unclosed block", page: f + "ant\nprintln(1)\n", want: []string{"block is not closed"}},
		{name: "output after cases", page: f + "ant\n1 + 1 // 2\n" + f + "\n\n" + f + "text\n2\n" + f + "\n", want: []string{"cases has no output block"}},
		{name: "sections out of order", rel: "lib/X/y.md", page: "## Scope\n\n## Basic examples\n\n" + f + "ant\n1 + 1 // 2\n" + f + "\n\n## Extras\n", want: []string{"## Basic examples is out of order", "## Extras is not a section"}},
		{name: "draft without examples", rel: "lib/X/y.md", page: "## Basic examples\n\n## See also\n", want: []string{"needs ## Basic examples with an example"}},
	}
	bin := buildAnt(t)
	files := map[string]bool{"probe/p.md": true, "probe/other.md": true, "README.md": true, "lib/X/y.md": true}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			rel := c.rel
			if rel == "" {
				rel = "probe/p.md"
			}
			page := c.page
			if !c.raw {
				page = draft + page
			}
			problems := pageProblems(t, bin, rel, page, files)
			all := strings.Join(problems, "\n")
			if len(c.want) == 0 && len(problems) > 0 {
				t.Errorf("want no problems, got:\n%s", all)
			}
			for _, w := range c.want {
				if !strings.Contains(all, w) {
					t.Errorf("want a problem containing %q, got:\n%s", w, all)
				}
			}
		})
	}
}
