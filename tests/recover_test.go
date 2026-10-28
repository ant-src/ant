package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ant-src/ant/internal/check"
	"github.com/ant-src/ant/internal/loader"
	"github.com/ant-src/ant/parser"
)

// TestRecoverCheck checks that a program loaded past its syntax errors can be checked without a
// panic, on every corpus program cut short at a sample of its token boundaries, and that the
// checker's diagnostics on the whole program are the same whether it was loaded strictly or not.
func TestRecoverCheck(t *testing.T) {
	var paths []string
	for _, pattern := range []string{"run/*.ant", "check/*.ant", "unchecked/*.ant", "../examples/*.ant"} {
		found, _ := filepath.Glob(pattern)
		paths = append(paths, found...)
	}
	if len(paths) < 100 {
		t.Fatalf("found only %d programs", len(paths))
	}
	for _, path := range paths {
		t.Run(testName(path), func(t *testing.T) {
			t.Parallel()
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			src := string(data)
			abs, _ := filepath.Abs(path)
			if strings.Contains(src, "// exit: 2") {
				return // a program that cannot load on purpose
			}
			whole := recoverCheck(t, abs, src, len(src))
			strict, diags := loader.Load(abs, loader.Config{Mode: parser.Program})
			if strict != nil {
				sdiags := check.Program(strict, check.Config{})
				if got, want := render(whole), render(sdiags); got != want {
					t.Errorf("recovered diagnostics differ from strict ones:\n%s\n---\n%s", got, want)
				}
			} else if len(diags) > 0 && len(whole) == 0 {
				t.Errorf("strict load fails with %v, recovering check reports nothing", diags)
			}
			_, l, _ := parser.ParseRecover(abs, src)
			for i, tok := range l.Tokens {
				if i%7 != 0 {
					continue
				}
				recoverCheck(t, abs, src, tok.Off)
			}
		})
	}
}

// recoverCheck loads the program at path with its root's text cut at cut, recovering, and checks
// it, returning every diagnostic.
func recoverCheck(t *testing.T, path, src string, cut int) []loader.Diagnostic {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("cut at %d: panic: %v\n%s", cut, r, src[:cut])
		}
	}()
	prog, diags := loader.Load(path, loader.Config{Mode: parser.Program, Recover: true, Sources: map[string]string{path: src[:cut]}})
	if prog == nil {
		t.Fatalf("cut at %d: no program: %v", cut, diags)
	}
	out := append([]loader.Diagnostic{}, diags...)
	out = append(out, check.Program(prog, check.Config{})...)
	for _, d := range out {
		if d.Path == "" || strings.Contains(d.Msg, "internal error") {
			t.Errorf("cut at %d: a malformed diagnostic: %v", cut, d)
		}
	}
	return out
}

func render(diags []loader.Diagnostic) string {
	var lines []string
	for _, d := range diags {
		lines = append(lines, d.String())
	}
	return strings.Join(lines, "\n")
}
