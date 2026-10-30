package ant

import (
	"context"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// TestSessionReview covers what review found: the transcript keeps every declaring input,
// since a declaration may read the one it declares again; an import after a declaration keeps
// the analysis whole; markup and _ have types; a loaded file imports from its own directory,
// and the inputs after it keep that import.
func TestSessionReview(t *testing.T) {
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "sub"), 0o755)
	os.WriteFile(filepath.Join(dir, "sub", "b.ant"), []byte("export func h(): int { 7 }\n"), 0o600)
	os.WriteFile(filepath.Join(dir, "sub", "a.ant"), []byte("import \"./b.ant\" { h }\nvar z = h()\n"), 0o600)
	os.WriteFile(filepath.Join(dir, "m.ant"), []byte("export var k = 1\n"), 0o600)
	s, err := NewSession(context.Background(), WithStdout(io.Discard), WithStderr(io.Discard))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	eval := func(src string) {
		t.Helper()
		if _, err := s.Eval(src); err != nil {
			t.Fatalf("%s: %v", src, err)
		}
	}
	typeOf := func(src, want string) {
		t.Helper()
		if got, err := s.TypeOf(src); err != nil || got != want {
			t.Errorf("TypeOf(%q) = %q, %v; want %q", src, got, err, want)
		}
	}
	eval(`var s = "a"`)
	eval(`var s = [s]`)
	typeOf("s", "Array[string]")
	eval("func g(): int { 1 }")
	eval("var x = g()")
	eval(`func g(): string { "" }`)
	typeOf("x", "int") // x keeps the type it had
	typeOf("g()", "string")
	eval(`import "` + filepath.Join(dir, "m.ant") + `" m`)
	typeOf("1 + 2", "int")
	typeOf("m.k", "int")
	typeOf("<div>hi</div>", "Html")
	eval("40 + 2")
	typeOf("_", "int")
	typeOf("_ * 2", "int")
	if _, err := s.EvalFile(filepath.Join(dir, "sub", "a.ant")); err != nil {
		t.Fatalf("EvalFile: %v", err)
	}
	if v, err := s.Eval("z + h()"); err != nil || v != int64(14) {
		t.Errorf("z + h() = %v, %v", v, err)
	}
	typeOf("h", "func(): int")
	if err := func() error { _, err := s.TypeOf("1 +"); return err }(); err == nil {
		t.Error("TypeOf(1 +): no error")
	} else if le := err.(*LoadError); le.Diagnostics[0].Pos.Line != 1 {
		t.Errorf("TypeOf(1 +): %v is not on the expression's line", err)
	}
}
