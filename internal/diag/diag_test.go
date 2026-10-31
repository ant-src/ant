package diag

import (
	"errors"
	"strings"
	"testing"

	"github.com/ant-src/ant/ast"
)

func printed(t *testing.T, src string, pos ast.Pos, msg string, color bool) string {
	t.Helper()
	p := &Printer{Color: color, Read: func(path string) (string, error) {
		if path != "a.ant" {
			return "", errors.New("no such file")
		}
		return src, nil
	}}
	var sb strings.Builder
	p.Print(&sb, "a.ant", pos, msg)
	return sb.String()
}

func TestPrint(t *testing.T) {
	tests := []struct {
		name string
		src  string
		pos  ast.Pos
		want string
	}{
		{"binary covers the expression", "var total = balance + \"x\"\n", ast.Pos{Line: 1, Col: 21}, `
a.ant:1:21: msg
   │
 1 │ var total = balance + "x"
   │             ^^^^^^^^^^^^^
`},
		{"name", "var a = 1\nprint(totl)\n", ast.Pos{Line: 2, Col: 7}, `
a.ant:2:7: msg
   │
 2 │ print(totl)
   │       ^^^^
`},
		{"keyword token", "var x = if\n", ast.Pos{Line: 1, Col: 9}, `
a.ant:1:9: msg
   │
 1 │ var x = if
   │         ^^
`},
		{"end of file shows the line before", "var x = (1 +\n", ast.Pos{Line: 2, Col: 1}, `
a.ant:2:1: msg
   │
 1 │ var x = (1 +
 2 │
   │ ^
`},
		{"tabs and wide characters", "func f() {\n\tvar s = \"語\" + 1\n}\n", ast.Pos{Line: 2, Col: 16}, `
a.ant:2:16: msg
   │
 2 │     var s = "語" + 1
   │             ^^^^^^^^
`},
		{"lines of a span", "var y = 1 +\n  \"a\"\n", ast.Pos{Line: 1, Col: 11}, `
a.ant:1:11: msg
   │
 1 │ var y = 1 +
   │         ^^^
 2 │   "a"
   │   ^^^
`},
		{"a long span elides its middle", "var y = 1 +\n  2 +\n  3 +\n  4 +\n  \"a\"\n", ast.Pos{Line: 4, Col: 5}, `
a.ant:4:5: msg
   │
 1 │ var y = 1 +
   │         ^^^
 2 │   2 +
   │   ^^^
 ⋮ │
 5 │   "a"
   │   ^^^
`},
		{"no position", "var a = 1\n", ast.Pos{}, `
a.ant: msg
`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := printed(t, tc.src, tc.pos, "msg", false)
			want := strings.TrimPrefix(tc.want, "\n") + "\n"
			if got != want {
				t.Errorf("got:\n%s\nwant:\n%s", got, want)
			}
		})
	}
}

func TestPrintUnreadable(t *testing.T) {
	p := &Printer{Read: func(string) (string, error) { return "", errors.New("gone") }}
	var sb strings.Builder
	p.Print(&sb, "b.ant", ast.Pos{Line: 3, Col: 4}, "msg")
	if got := sb.String(); got != "b.ant:3:4: msg\n\n" {
		t.Errorf("got %q", got)
	}
}

func TestPrintColor(t *testing.T) {
	got := printed(t, "var x = 1 + \"a\"\n", ast.Pos{Line: 1, Col: 11}, "msg", true)
	for _, part := range []string{
		"\x1b[1ma.ant:1:11:\x1b[0m \x1b[1mmsg\x1b[0m\n",
		"\x1b[35mvar\x1b[0m x = \x1b[33m1\x1b[0m + \x1b[32m\"a\"\x1b[0m\n",
		"\x1b[1;31m^^^^^^^\x1b[0m\n",
	} {
		if !strings.Contains(got, part) {
			t.Errorf("%q lacks %q", got, part)
		}
	}
}

func TestSummary(t *testing.T) {
	var sb strings.Builder
	p := &Printer{}
	p.Summary(&sb, 1)
	p.Summary(&sb, 3)
	if got := sb.String(); got != "1 error\n3 errors\n" {
		t.Errorf("got %q", got)
	}
}
