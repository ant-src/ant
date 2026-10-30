package repl

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/ant-src/ant"
	"github.com/ant-src/ant/value"
)

func TestPretty(t *testing.T) {
	s, err := ant.NewSession(context.Background(), ant.WithStdout(io.Discard), ant.WithStderr(io.Discard))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	eval := func(src string) value.Value {
		t.Helper()
		v, err := s.Eval(src)
		if err != nil {
			t.Fatalf("%s: %v", src, err)
		}
		return v
	}
	eval("type P = { name: string, tags: Array[string], n: int }")
	// what fits is String.from, plain
	for _, src := range []string{`P{ name: "a", tags: ["x"], n: 1 }`, `Map{ "k": [1, 2] }`, `Set.of([1])`, `[]`, `Map{}`, `"top"`, `1.5`, `null`} {
		v := eval(src)
		if got, want := plainTheme.pretty(v, 80), value.Render(v); got != want {
			t.Errorf("%s: %q, want %q", src, got, want)
		}
	}
	// what does not fit goes over lines, each element indented, nested ones that fit kept flat
	v := eval(`P{ name: "a long name", tags: ["first tag", "second tag"], n: 42 }`)
	want := "P{\n  name: \"a long name\",\n  tags: [\"first tag\", \"second tag\"],\n  n: 42,\n}"
	if got := plainTheme.pretty(v, 40); got != want {
		t.Errorf("wide record:\n%s\nwant\n%s", got, want)
	}
	v = eval(`Map{ "alpha": [1, 2, 3], "beta": [4, 5, 6] }`)
	want = "Map{\n  \"alpha\": [1, 2, 3],\n  \"beta\": [4, 5, 6],\n}"
	if got := plainTheme.pretty(v, 20); got != want {
		t.Errorf("wide map:\n%s\nwant\n%s", got, want)
	}
	// a long array is cut short
	v = eval(`Array.range(0, 1000)`)
	if got := plainTheme.pretty(v, 80); !strings.HasSuffix(got, "  99,\n  … 900 more\n]") {
		t.Errorf("long array ends %q", got[max(0, len(got)-40):])
	}
	// colors: a number, a string, a type
	v = eval(`P{ name: "a", tags: [], n: 1 }`)
	got := ansiTheme.pretty(v, 80)
	for _, want := range []string{"\x1b[36mP\x1b[0m{ ", "\x1b[32m\"a\"\x1b[0m", "\x1b[33m1\x1b[0m"} {
		if !strings.Contains(got, want) {
			t.Errorf("colored %q lacks %q", got, want)
		}
	}
}
