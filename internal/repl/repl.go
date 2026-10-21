// Package repl is the read-eval-print loop a bare ant command runs.
package repl

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/ant-src/ant"
	"github.com/ant-src/ant/value"
)

// Run reads inputs from stdin until it ends, evaluating each on one session, printing every
// non-void result with String.from, and reporting errors on stderr. It returns the exit status.
func Run(ctx context.Context, stdin io.Reader, stdout, stderr io.Writer) int {
	s, err := ant.NewSession(ctx, ant.WithStdout(stdout), ant.WithStderr(stderr))
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	defer s.Close()
	in := bufio.NewScanner(stdin)
	in.Buffer(make([]byte, 1024*1024), 1024*1024)
	var pending strings.Builder
	prompt := func() {
		if pending.Len() == 0 {
			fmt.Fprint(stdout, "> ")
		} else {
			fmt.Fprint(stdout, ". ")
		}
	}
	prompt()
	for in.Scan() {
		line := in.Text()
		pending.WriteString(line)
		pending.WriteString("\n")
		src := pending.String()
		if strings.TrimSpace(src) == "" {
			pending.Reset()
			prompt()
			continue
		}
		if ant.Incomplete(src) {
			prompt()
			continue
		}
		pending.Reset()
		v, err := s.Eval(src)
		switch err.(type) {
		case nil:
			if v != value.Void {
				fmt.Fprintln(stdout, value.Render(v))
			}
		case *ant.PanicError:
			// reported by the runtime already
		case *ant.ExitError:
			return ant.ExitStatus(err)
		default:
			fmt.Fprintln(stderr, err)
		}
		prompt()
	}
	fmt.Fprintln(stdout)
	return 0
}
