package tests

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ant-src/ant"
)

// TestCoverageScript runs scripts/coverage.ant, the coverage gate, on a small profile: a block
// that two test binaries report counts once, a package below the per-package minimum and a total
// below the overall minimum are marked, the status is 1 when anything is below and 2 for a bad
// invocation. The script must also check clean.
func TestCoverageScript(t *testing.T) {
	script, err := filepath.Abs("../scripts/coverage.ant")
	if err != nil {
		t.Fatal(err)
	}
	prog, err := ant.Load(script)
	if err != nil {
		t.Fatal(err)
	}
	if diags := prog.Check(); len(diags) != 0 {
		t.Fatalf("check: %v", diags)
	}
	profile := filepath.Join(t.TempDir(), "cover.out")
	const module = "github.com/ant-src/ant/"
	lines := []string{
		"mode: set",
		module + "value/a.go:1.1,2.2 4 1",
		module + "value/a.go:1.1,2.2 4 0", // the same block again, not hit this time
		module + "value/a.go:3.1,4.2 6 0",
		module + "value/a.go:3.1,4.2 6 1", // the same block again, hit this time
		module + "ant.go:1.1,2.2 10 1",
		module + "ant.go:3.1,4.2 10 0",
		module + "internal/lib/x.go:1.1,2.2 1 0",
	}
	if err := os.WriteFile(profile, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	row := func(name string, pct float64, mark string) string {
		return fmt.Sprintf("%-22s %5.1f%%%s", name, pct, mark)
	}
	cases := []struct {
		name   string
		args   []string
		status int
		out    []string
	}{
		{"below", []string{profile, "90", "80"}, 1, []string{
			row("(root)", 50, "  BELOW 80%"),
			row("internal/lib", 0, "  BELOW 80%"),
			row("value", 100, ""),
			row("total", 100*20.0/31, "  BELOW 90%"),
		}},
		{"within", []string{profile, "60", "0"}, 0, []string{
			row("(root)", 50, ""),
			row("internal/lib", 0, ""),
			row("value", 100, ""),
			row("total", 100*20.0/31, ""),
		}},
		{"defaults", []string{profile}, 1, []string{
			row("(root)", 50, "  BELOW 80%"),
			row("internal/lib", 0, "  BELOW 80%"),
			row("value", 100, ""),
			row("total", 100*20.0/31, "  BELOW 90%"),
		}},
		{"usage", nil, 2, []string{
			"usage: ant run scripts/coverage.ant <profile> [overall-min] [package-min]",
		}},
		{"bad threshold", []string{profile, "ninety"}, 2, []string{"not a number: ninety"}},
		{"missing profile", []string{filepath.Join(t.TempDir(), "none")}, 2, nil},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			err := prog.Run(context.Background(),
				ant.WithStdout(&stdout), ant.WithStderr(&stderr), ant.WithArgs(c.args))
			if status := ant.ExitStatus(err); status != c.status {
				t.Errorf("status %d, want %d\nstdout:\n%s\nstderr:\n%s", status, c.status, &stdout, &stderr)
			}
			got := strings.TrimSuffix(stdout.String(), "\n")
			if c.out == nil {
				if !strings.HasPrefix(got, "cannot read ") {
					t.Errorf("stdout:\n%s\nwant a cannot-read message", got)
				}
				return
			}
			if want := strings.Join(c.out, "\n"); got != want {
				t.Errorf("stdout:\n%s\nwant:\n%s", got, want)
			}
		})
	}
}
