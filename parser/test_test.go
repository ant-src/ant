package parser

import (
	"strings"
	"testing"

	"github.com/ant-src/ant/ast"
)

// TestTestFiles covers the syntax of tests (section 10, Tests): where test declares a test or
// runs a subtest, what its name may be, and test as a keyword in test files only.
func TestTestFiles(t *testing.T) {
	errs := []struct{ name, path, src, want string }{
		{"top-level statement", "a_test.ant", "var x = 1\nx = 2", "cannot have top-level statements"},
		{"interpolated top-level name", "a_test.ant", "var n = 1\ntest \"${n}\" {}", "without interpolation"},
		{"empty name", "a_test.ant", "test \"\" {}", "cannot be empty"},
		{"duplicate name", "a_test.ant", "test \"a\" {}\ntest \"a\" {}", `test "a" is already declared`},
		{"name not a string", "a_test.ant", "test x {}", "expected the test's name"},
		{"subtest name not a string", "a_test.ant", "test \"a\" { test 1 {} }", "expected the test's name"},
		{"test in a function", "a_test.ant", "func f() { test \"a\" {} }", "runs as a subtest in a test's body"},
		{"subtest in a lambda", "a_test.ant", "test \"a\" { var f = { test \"b\" {} } }", "runs as a subtest in a test's body"},
		{"subtest in a go block", "a_test.ant", "test \"a\" { go { test \"b\" {} } }", "runs as a subtest in a test's body"},
		{"test is a keyword in a test file", "a_test.ant", "var test = 1", "'test' cannot be a variable name"},
		{"test outside a test file", "a.ant", "test \"a\" {}", "expected a newline"},
	}
	for _, c := range errs {
		t.Run(c.name, func(t *testing.T) {
			_, err := Parse(c.path, c.src, Program)
			if err == nil || !strings.Contains(err.Error(), c.want) {
				t.Fatalf("got %v, want an error containing %q", err, c.want)
			}
		})
	}
	ok := []struct{ name, path, src string }{
		{"test is a name elsewhere", "a.ant", "var test = 1\nprintln(test)"},
		{"a field named test", "a_test.ant", "var t = Regex.test\nvar r = Map{ test: 1 }"},
		{"a named argument and an import named test", "a_test.ant", "import \"./lib.ant\" { test: check }\nvar x = f(test: 3)"},
		{"exports for other test files", "helpers_test.ant", "export func f(): int { 1 }"},
		{"try, return, and defer in a body", "a_test.ant", "test \"a\" {\n  defer { println(1) }\n  var x = try Result.ok(1)\n  if x > 0 { return }\n}"},
		{"subtests in loops and blocks", "a_test.ant", "test \"a\" {\n  for i in [1] {\n    if true { test \"${i}\" { test \"deeper\" {} } }\n  }\n}"},
	}
	for _, c := range ok {
		t.Run(c.name, func(t *testing.T) {
			if _, err := Parse(c.path, c.src, Program); err != nil {
				t.Fatal(err)
			}
		})
	}
	f, err := Parse("a_test.ant", "// Doc.\ntest \"a\" { test \"b${1}\" {} }\nfunc f() {}", Program)
	if err != nil {
		t.Fatal(err)
	}
	if len(f.Tests) != 1 || f.Tests[0].Doc != "Doc." || f.Script || len(f.Decls) != 1 {
		t.Fatalf("file: %s", ast.Sprint(f))
	}
	if _, ok := f.Tests[0].Body.Stmts[0].(*ast.Test).Name.(*ast.Template); !ok {
		t.Errorf("subtest: %s", ast.Sprint(f.Tests[0]))
	}
}
