package ant

import (
	"context"
	"io"
	"testing"
)

// TestSessionDecls checks that the transcript Complete analyzes keeps one text per name: an
// input declared again replaces the earlier one, while an input that also declared something
// else stays.
func TestSessionDecls(t *testing.T) {
	s, err := NewSession(context.Background(), WithStdout(io.Discard), WithStderr(io.Discard))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	for _, src := range []string{"var a = 1", "var b = 2", "var a = \"s\"", "1 + 1", "type CD = { c: int, d: int }", "var { c, d } = CD{ c: 1, d: 2 }", "var c = 3", "func f(): int { 1 }", "func f(): string { \"\" }", "func f(): string { \"x\" }", "type T = { n: int }", "func T.get(t: T): int { t.n }"} {
		if _, err := s.Eval(src); err != nil {
			t.Fatalf("%s: %v", src, err)
		}
	}
	var got []string
	for _, d := range s.decls {
		got = append(got, d.src)
	}
	want := []string{"var b = 2", "var a = \"s\"", "type CD = { c: int, d: int }", "var { c, d } = CD{ c: 1, d: 2 }", "var c = 3", "func f(): string { \"x\" }", "type T = { n: int }", "func T.get(t: T): int { t.n }"}
	if len(got) != len(want) {
		t.Fatalf("decls %q, want %q", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("decls[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}
