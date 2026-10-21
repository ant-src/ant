// Package bench holds the benchmark programs and the Go benchmark that runs them through the
// library. BASELINE.md records the numbers of the last baseline.
package bench

import (
	"context"
	"io"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ant-src/ant"
)

// BenchmarkRun runs each program under bench/ once per iteration, with its output discarded.
func BenchmarkRun(b *testing.B) {
	paths, err := filepath.Glob("*.ant")
	if err != nil || len(paths) == 0 {
		b.Fatal("no programs")
	}
	for _, path := range paths {
		prog, err := ant.Load(path)
		if err != nil {
			b.Fatal(err)
		}
		b.Run(strings.TrimSuffix(path, ".ant"), func(b *testing.B) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if err := prog.Run(context.Background(), ant.WithStdout(io.Discard), ant.WithStderr(io.Discard)); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

// TestPrograms checks that every benchmark program runs and checks clean, so the benchmark
// measures the interpreter and not a panic.
func TestPrograms(t *testing.T) {
	paths, _ := filepath.Glob("*.ant")
	for _, path := range paths {
		prog, err := ant.Load(path)
		if err != nil {
			t.Fatal(err)
		}
		if diags := prog.Check(); len(diags) > 0 {
			t.Errorf("%s: %v", path, diags)
		}
		if err := prog.Run(context.Background(), ant.WithStdout(io.Discard), ant.WithStderr(io.Discard)); err != nil {
			t.Errorf("%s: %v", path, err)
		}
	}
}
