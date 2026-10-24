package tests

import (
	"strings"
	"testing"
	"time"
)

// TestStdlibHarness feeds the library-table harness tables that are wrong in one way each, and
// checks that it says so; and tables that are right in an unusual way, and checks that it says
// nothing. want is a text the one reported problem must contain, or "" for none.
func TestStdlibHarness(t *testing.T) {
	cases := []struct {
		name, src, want string
	}{
		{"wrong value", `String.count("aaaa", "aa")  // => 3`, "want: 3"},
		{"wrong panic", `String.repeat("x", -1)  // => panic: no`, "want: panic: no"},
		{"panic expected", `String.repeat("x", 1)  // => panic: no`, `got:  "x"`},
		{"marker without spaces", "String.count(\"aaaa\", \"aa\") //=> 999\nvar done = true", "want: 999"},
		{"marker with more spaces", `String.length("ab") //  => 7`, "want: 7"},
		{"marker not on the last line", "String.join([  // => \"x\"\n  \"a\"], \",\")  // => \"a\"", "nowhere else"},
		{"setup with a value", "String.length(\"ab\")\nvar done = true", "a value that nothing checks"},
		{"void case", "var xs = [1]\nArray.push(xs, 2)  // => null", "this one is void"},
		{"void case that panics", "var xs = [1]\nArray.insert(xs, 5, 2)  // => panic: Array.insert: the index 5 is not within 0..1", ""},
		{"setup panics", `var n = String.repeat("x", -1)`, "the input failed"},
		{"error is not an expectation", `Os.exit(3)  // => error: exit status 3`, "Os.exit"},
		{"goroutine", "go { }", "go"},
		{"select", "select { else { 1 } }", "select"},
		{"waits", "var rw = RwMutex.new()\nRwMutex.writeLock(rw)", "RwMutex.writeLock"},
		{"lock as a value", "var lock = Mutex.withLock", "Mutex.withLock"},
		{"go in a string", `String.length("go {")  // => 4`, ""},
		{"import", `import "../run/lib/geometry.ant" G`, "cannot import"},
		{"output", `println("hi")`, "writes nothing"},
		{"redeclared", "var s = 1\nvar s = 2", "already declared"},
		{"type error at its line", "var s = 1\n\nString.count(1, \"a\")  // => 0", "selftest.ant:3: cannot use int"},
		{"a second comment", `String.repeat("x", 2)  // note // => "xx"`, ""},
		{"multi-line case", "Array.map([1, 2], {\n  x * 2\n})  // => [2, 4]", ""},
		{"not closed", "Array.map([1, 2], {", "not closed"},
		{"endless", "var n = 0\nwhile true { n += 1 }", "did not finish in time"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			t.Parallel()
			problems := tableProblems(t.Context(), "stdlib/selftest.ant", c.src+"\n", time.Second)
			switch {
			case c.want == "" && len(problems) > 0:
				t.Errorf("want no problem, got %q", problems)
			case c.want != "" && len(problems) != 1:
				t.Errorf("want one problem containing %q, got %q", c.want, problems)
			case c.want != "" && !strings.Contains(problems[0], c.want):
				t.Errorf("want a problem containing %q, got %q", c.want, problems[0])
			}
		})
	}
}

// TestStdlibPanicCredit checks which function a panic case covers: the outermost call's, and
// only when the panic names it.
func TestStdlibPanicCredit(t *testing.T) {
	for src, want := range map[string]string{
		`String.repeat("x", -1)  // => panic: String.repeat needs a count of at least 0, got -1`:                    "String.repeat",
		`Array.chunk(Array.filled(-1, 0), 1)  // => panic: Array.filled needs a count of at least 0, got -1`:        "",
		`Iter.toArray(Iter.take(Iter.range(0, 3), -1))  // => panic: Iter.take needs a count of at least 0, got -1`: "",
		`String.repeat("x", 1)  // => "x"`: "",
	} {
		inputs, err := readTable("stdlib/selftest.ant", src)
		if err != nil {
			t.Fatal(err)
		}
		if got := coveredPanic(inputs[0]); got != want {
			t.Errorf("%s: covers %q, want %q", src, got, want)
		}
	}
}
