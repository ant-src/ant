// Package tests runs the programs under run/ through the library and compares what they do
// with what their trailers expect.
//
// Each program under run/ is a complete Ant program followed by a trailer that starts with a
// line reading "// expect" and continues with lines of these forms:
//
//	// out: <line>     one expected line of standard output, in order; "// out:" alone is an
//	                   empty line
//	// exit: <n>       the expected exit status, as the ant command would report it; default 0
//	// err: <text>     standard error must contain this text; may repeat
//	// args: a b c     command-line arguments, split on spaces
//
// Files under run/lib are modules the programs import, not tests. Every program under run/ must
// also pass ant check, which is tested once the checker exists; programs under unchecked/ are
// the kind panics the checker would reject, and are only run.
package tests

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/ant-src/ant"
)

type expectation struct {
	out  []string
	exit int
	err  []string
	args []string
}

func parseTrailer(t *testing.T, path, src string) expectation {
	t.Helper()
	i := strings.LastIndex(src, "\n// expect\n")
	if i < 0 {
		t.Fatalf("%s: no \"// expect\" trailer", path)
	}
	var e expectation
	for _, line := range strings.Split(strings.TrimRight(src[i+len("\n// expect\n"):], "\n"), "\n") {
		key, rest, _ := strings.Cut(line, ":")
		rest = strings.TrimPrefix(rest, " ")
		switch key {
		case "// out":
			e.out = append(e.out, rest)
		case "// exit":
			n, err := strconv.Atoi(strings.TrimSpace(rest))
			if err != nil {
				t.Fatalf("%s: bad exit line %q", path, line)
			}
			e.exit = n
		case "// err":
			e.err = append(e.err, strings.TrimSpace(rest))
		case "// args":
			e.args = strings.Fields(rest)
		default:
			t.Fatalf("%s: unexpected trailer line %q", path, line)
		}
	}
	return e
}

func TestRun(t *testing.T) {
	paths, err := filepath.Glob("run/*.ant")
	if err != nil {
		t.Fatal(err)
	}
	unchecked, err := filepath.Glob("unchecked/*.ant")
	if err != nil {
		t.Fatal(err)
	}
	paths = append(paths, unchecked...)
	if len(paths) == 0 {
		t.Fatal("no programs under run/")
	}
	for _, path := range paths {
		t.Run(strings.TrimSuffix(strings.TrimPrefix(path, "run/"), ".ant"), func(t *testing.T) {
			t.Parallel()
			src, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			want := parseTrailer(t, path, string(src))
			abs, err := filepath.Abs(path)
			if err != nil {
				t.Fatal(err)
			}
			var stdout, stderr bytes.Buffer
			prog, err := ant.Load(abs)
			if err == nil {
				ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
				defer cancel()
				err = prog.Run(ctx,
					ant.WithStdout(&stdout), ant.WithStderr(&stderr), ant.WithArgs(want.args))
			}
			// The command prints load errors and unexpected errors itself; mirror that.
			switch err.(type) {
			case nil, *ant.PanicError, *ant.ExitError:
			default:
				stderr.WriteString(err.Error() + "\n")
			}
			status := ant.ExitStatus(err)
			if status != want.exit {
				t.Errorf("exit status %d, want %d\nstderr:\n%s", status, want.exit, stderr.String())
			}
			got := strings.Split(strings.TrimSuffix(stdout.String(), "\n"), "\n")
			if stdout.Len() == 0 {
				got = nil
			}
			if strings.Join(got, "\n") != strings.Join(want.out, "\n") {
				t.Errorf("stdout:\n%s\nwant:\n%s", stdout.String(), strings.Join(want.out, "\n"))
			}
			for _, text := range want.err {
				if !strings.Contains(stderr.String(), text) {
					t.Errorf("stderr lacks %q:\n%s", text, stderr.String())
				}
			}
		})
	}
}
