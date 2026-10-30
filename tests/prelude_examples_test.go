package tests

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/ant-src/ant"
)

// TestPreludeExamples runs the examples in the doc comments of the prelude, as the header of
// internal/prelude/prelude.d.ant says they are written. An example whose lines include a case, a
// line ending in two spaces and "// value", is a library table (stdlib/README.md): every case
// must render as its comment says. An example with no case is only checked, as the body of a
// script, and may use names it does not define, placeholders such as handle or job.
func TestPreludeExamples(t *testing.T) {
	path := filepath.Join("..", "internal", "prelude", "prelude.d.ant")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(data), "\n")
	blocks := exampleBlocks(lines)
	if len(blocks) < 40 {
		t.Fatalf("found only %d examples", len(blocks))
	}
	for _, b := range blocks {
		t.Run(b.decl, func(t *testing.T) {
			t.Parallel()
			for _, p := range exampleProblems(t, path, len(lines), b) {
				t.Error(p)
			}
		})
	}
}

// exampleBlock is the example lines of one doc comment, keyed by their line in the file, without
// the comment prefix, and the name of the declaration or namespace the comment documents.
type exampleBlock struct {
	decl  string
	lines map[int]string
	cases int
}

var (
	codeLine     = regexp.MustCompile(`^//( {2,})(.*)$`)
	caseLine     = regexp.MustCompile(`^(.*\S) {2,}// (.*)$`)
	declName     = regexp.MustCompile(`^(?:func|type|var) ([\w.]+)`)
	labelComment = regexp.MustCompile(`^// ([\w.]+)$`)
)

// exampleBlocks finds every doc comment with example lines. A comment ends at the first line
// that is not a comment: the declaration it documents, or a blank line after a section label. A
// code line that is itself a comment, "// value", gives the value of the line before it.
func exampleBlocks(lines []string) []exampleBlock {
	var blocks []exampleBlock
	var b exampleBlock
	indent, start, prev := "", 0, 0
	for i, line := range lines {
		n := i + 1
		if strings.HasPrefix(line, "//") {
			if start == 0 {
				start = n
			}
			m := codeLine.FindStringSubmatch(line)
			if m == nil {
				continue
			}
			if b.lines == nil {
				b.lines = map[int]string{}
				indent = m[1]
			}
			code := strings.TrimPrefix(m[1], indent) + m[2]
			if value, ok := strings.CutPrefix(code, "// "); ok && prev != 0 {
				b.lines[prev] += "   // " + value
			} else {
				b.lines[n] = code
				prev = n
			}
			if caseLine.MatchString(b.lines[prev]) {
				b.cases++
			}
			continue
		}
		if b.lines != nil {
			if m := declName.FindStringSubmatch(line); m != nil {
				b.decl = m[1]
			} else if m := labelComment.FindStringSubmatch(lines[start-1]); m != nil {
				b.decl = m[1]
			} else {
				b.decl = fmt.Sprintf("line %d", start)
			}
			blocks = append(blocks, b)
		}
		b, start, prev = exampleBlock{}, 0, 0
	}
	return blocks
}

// exampleProblems checks or runs one example and returns what is wrong with it. The example's
// lines are placed at their own line numbers in a source of n lines, so a problem names the line
// of prelude.d.ant it is on. An example without a case gets a statement after its last line, so
// that it is a script and not a library, whose initializers may not try.
func exampleProblems(t *testing.T, path string, n int, b exampleBlock) []string {
	t.Helper()
	padded := make([]string, n)
	for i, l := range b.lines {
		if m := caseLine.FindStringSubmatch(l); m != nil && b.cases > 0 {
			l = m[1] + " // => " + m[2]
		}
		padded[i-1] = l
	}
	fake := filepath.Join(filepath.Dir(path), "example.ant")
	var problems []string
	if b.cases > 0 {
		problems = tableProblems(t.Context(), fake, strings.Join(padded, "\n")+"\n", tableTimeout)
	} else {
		problems = checkExample(fake, strings.Join(append(padded, "println()"), "\n")+"\n")
	}
	for i, p := range problems {
		problems[i] = strings.Replace(p, fake, path, 1)
	}
	return problems
}

// placeholder matches the diagnostics an example without a case may have: a name it does not
// define.
var placeholder = regexp.MustCompile(`^(\w+ is not defined|unknown type \w+)$`)

// checkExample checks src as a program and returns its diagnostics, less those placeholder names
// cause.
func checkExample(path, src string) []string {
	abs, _ := filepath.Abs(path)
	var problems []string
	report := func(d ant.Diagnostic) {
		if !placeholder.MatchString(d.Message) {
			problems = append(problems, fmt.Sprintf("%s:%d: %s", path, d.Pos.Line, d.Message))
		}
	}
	prog, err := ant.Load(abs, ant.WithSource(abs, src))
	if err != nil {
		var le *ant.LoadError
		if !errors.As(err, &le) {
			return []string{fmt.Sprintf("%s: %v", path, err)}
		}
		for _, d := range le.Diagnostics {
			report(d)
		}
		return problems
	}
	for _, d := range prog.Check() {
		report(d)
	}
	return problems
}
