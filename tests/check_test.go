package tests

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/ant-src/ant"
)

// TestCheckRunCorpus checks that every program under run/ and examples/ passes ant check:
// those programs reach only value-dependent panics, which the checker allows.
func TestCheckRunCorpus(t *testing.T) {
	paths, _ := filepath.Glob("run/*.ant")
	examples, _ := filepath.Glob(filepath.Join("..", "examples", "*.ant"))
	paths = append(paths, examples...)
	for _, path := range paths {
		t.Run(testName(path), func(t *testing.T) {
			t.Parallel()
			src, _ := os.ReadFile(path)
			if strings.Contains(string(src), "// exit: 2") {
				t.Skip("cannot load")
			}
			abs, _ := filepath.Abs(path)
			prog, err := ant.Load(abs)
			if err != nil {
				t.Fatal(err)
			}
			for _, d := range prog.Check() {
				t.Error(d.Error())
			}
		})
	}
}

var errorLine = regexp.MustCompile(`^// error: (\d+): (.*)$`)

// TestCheckCorpus runs the checker over check/*.ant. A line that ends in "// ERR: <text>"
// expects an error on that line whose message contains the text; a file may also list
// "// error: <line>: <text>" lines. A file without any must check clean. Every expected error
// must be reported, and nothing else may be reported.
func TestCheckCorpus(t *testing.T) {
	paths, _ := filepath.Glob("check/*.ant")
	if len(paths) == 0 {
		t.Skip("no check programs yet")
	}
	for _, path := range paths {
		name := strings.TrimSuffix(filepath.Base(path), ".ant")
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			src, _ := os.ReadFile(path)
			type want struct {
				line int
				text string
			}
			var wants []want
			for i, line := range strings.Split(string(src), "\n") {
				if m := errorLine.FindStringSubmatch(line); m != nil {
					n, _ := strconv.Atoi(m[1])
					wants = append(wants, want{n, m[2]})
					continue
				}
				if _, rest, ok := strings.Cut(line, "// ERR: "); ok {
					for _, text := range strings.Split(rest, " ;; ") {
						wants = append(wants, want{i + 1, strings.TrimSpace(text)})
					}
				}
			}
			abs, _ := filepath.Abs(path)
			prog, err := ant.Load(abs)
			if err != nil {
				t.Fatal(err)
			}
			diags := prog.Check()
			used := make([]bool, len(diags))
			for _, w := range wants {
				found := false
				for i, d := range diags {
					if !used[i] && d.Pos.Line == w.line && strings.Contains(d.Message, w.text) {
						used[i] = true
						found = true
						break
					}
				}
				if !found {
					t.Errorf("expected an error on line %d containing %q", w.line, w.text)
				}
			}
			for i, d := range diags {
				if !used[i] {
					t.Errorf("unexpected: %s", d.Error())
				}
			}
		})
	}
}
