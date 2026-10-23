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
// non-void result with String.from, and reporting errors on stderr. It returns the exit status:
// 130 when ctx ends first, which is an interrupt, as it ends a run.
func Run(ctx context.Context, stdin io.Reader, stdout, stderr io.Writer) int {
	// The REPL owns standard input, so a program's Os.stdin() is at end of file.
	s, err := ant.NewSession(ctx, ant.WithStdin(strings.NewReader("")), ant.WithStdout(stdout),
		ant.WithStderr(stderr))
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	defer s.Close()
	// The lines come from a reader of their own, so an interrupt at the prompt ends the loop
	// without waiting for another line.
	lines := make(chan string)
	go func() {
		defer close(lines)
		// a reader, not a scanner, so a line of any length is read whole
		in := bufio.NewReader(stdin)
		for {
			line, err := in.ReadString('\n')
			if line == "" && err != nil {
				return
			}
			line = strings.TrimSuffix(strings.TrimSuffix(line, "\n"), "\r")
			select {
			case lines <- line:
			case <-ctx.Done():
				return
			}
		}
	}()
	var pending strings.Builder
	prompt := func() {
		if pending.Len() == 0 {
			fmt.Fprint(stdout, "> ")
		} else {
			fmt.Fprint(stdout, ". ")
		}
	}
	prompt()
	for {
		var line string
		var ok bool
		select {
		case line, ok = <-lines:
		case <-ctx.Done():
		}
		if ctx.Err() != nil {
			fmt.Fprintln(stdout)
			return 130
		}
		if !ok {
			break
		}
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
		if ctx.Err() != nil {
			fmt.Fprintln(stdout)
			return 130
		}
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
