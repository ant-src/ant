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

// parallel lists the programs that take a goroutine count as their argument, and the counts the
// benchmark runs them at.
var parallel = map[string][]string{"par_fib": {"1", "4", "10"}, "par_map": {"1", "4", "10"}}

// BenchmarkRun runs each program under bench/ once per iteration, with its output discarded.
// A parallel program runs once per goroutine count, as <name>/n=<count>.
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
		name := strings.TrimSuffix(path, ".ant")
		run := func(b *testing.B, args []string) {
			b.ReportAllocs()
			for i := 0; i < b.N; i++ {
				if err := prog.Run(context.Background(), ant.WithStdout(io.Discard), ant.WithStderr(io.Discard), ant.WithArgs(args)); err != nil {
					b.Fatal(err)
				}
			}
		}
		counts, ok := parallel[name]
		if !ok {
			b.Run(name, func(b *testing.B) { run(b, nil) })
			continue
		}
		for _, n := range counts {
			b.Run(name+"/n="+n, func(b *testing.B) { run(b, []string{n}) })
		}
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
