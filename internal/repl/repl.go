// Package repl is the read-eval-print loop a bare ant command runs.
package repl

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/ant-src/ant"
	"github.com/ant-src/ant/value"
)

// Run reads inputs from stdin until it ends, evaluating each on one session, printing every
// non-void result with String.from, and reporting errors on stderr. On a terminal it edits each
// line with the keys of readline, highlights it as it is typed, and completes names, members,
// and arguments with Tab, suggesting the first after the cursor. It returns the exit status:
// 130 when ctx ends first, which is an interrupt, as it ends a run, or when Ctrl-C is pressed
// twice in a row at the prompt.
func Run(ctx context.Context, stdin io.Reader, stdout, stderr io.Writer) int {
	var r lineReader
	var e *editor
	if in, out, ok := terminal(stdin, stdout); ok {
		// what the session and the loop write goes through the editor, which keeps the
		// prompt whole when a goroutine prints while it waits
		e = newTerminalReader(ctx, in, out)
		r = e
		stdout, stderr = e.writer(stdout), e.writer(stderr)
	} else {
		r = newPlainReader(ctx, stdin, stdout)
	}
	// The REPL owns standard input, so a program's Os.stdin() is at end of file.
	s, err := ant.NewSession(ctx, ant.WithStdin(strings.NewReader("")), ant.WithStdout(stdout),
		ant.WithStderr(stderr))
	if err != nil {
		fmt.Fprintln(stderr, err)
		return 2
	}
	defer s.Close()
	if e != nil {
		e.complete = s.Complete
	}
	return loop(ctx, s, r, stdout, stderr)
}

// loop is the read-eval-print loop over the lines r reads.
func loop(ctx context.Context, s *ant.Session, r lineReader, stdout, stderr io.Writer) int {
	var pending strings.Builder
	// armed is set by Ctrl-C: another right after it ends the REPL
	armed := false
	for {
		prompt := "> "
		if pending.Len() > 0 {
			prompt = ". "
		}
		line, err := r.readLine(prompt, pending.String())
		if errors.Is(err, errInterrupt) {
			if line == "" && armed {
				return 130
			}
			fmt.Fprintln(stdout, "(press ^C again or ^D to exit)")
			armed = true
			pending.Reset()
			continue
		}
		armed = false
		if ctx.Err() != nil {
			fmt.Fprintln(stdout)
			return 130
		}
		if err != nil && !errors.Is(err, io.EOF) {
			// the terminal cannot be read, or put into raw mode
			fmt.Fprintln(stderr, err)
			return 2
		}
		// at the end of input an unfinished input is evaluated as it is, so its error is
		// reported rather than the input dropped
		eof := err != nil
		if eof && pending.Len() == 0 {
			break
		}
		if !eof {
			pending.WriteString(line)
			pending.WriteString("\n")
		}
		src := pending.String()
		if strings.TrimSpace(src) == "" {
			pending.Reset()
			continue
		}
		if !eof && ant.Incomplete(src) {
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
		case *ant.PanicError, *ant.ResultError:
			// reported by the runtime already
		case *ant.ExitError:
			return ant.ExitStatus(err)
		default:
			fmt.Fprintln(stderr, err)
		}
		if eof {
			break
		}
	}
	if _, ok := r.(*plainReader); ok {
		fmt.Fprintln(stdout)
	}
	return 0
}

// A lineReader reads the lines of the inputs. readLine returns io.EOF at the end of the input,
// errInterrupt with the line Ctrl-C dropped, the context's error when it ends, and another
// error when the input cannot be read; pending is the input so far, for the continuation
// prompt's context.
type lineReader interface {
	readLine(prompt, pending string) (string, error)
}

// terminal reports whether the REPL is on a terminal it can edit lines on: standard input and
// output are the same kind of terminal, and not a dumb one, which has no cursor to move.
func terminal(stdin io.Reader, stdout io.Writer) (in, out *os.File, ok bool) {
	in, ok = stdin.(*os.File)
	if !ok || !isTerminal(in) {
		return nil, nil, false
	}
	out, ok = stdout.(*os.File)
	if !ok || !isTerminal(out) || os.Getenv("TERM") == "dumb" {
		return nil, nil, false
	}
	return in, out, true
}

// plainReader reads lines as they come, for a pipe or a file, printing the prompt before each.
type plainReader struct {
	ctx   context.Context
	out   io.Writer
	lines <-chan string
}

func newPlainReader(ctx context.Context, stdin io.Reader, stdout io.Writer) *plainReader {
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
	return &plainReader{ctx: ctx, out: stdout, lines: lines}
}

func (r *plainReader) readLine(prompt, pending string) (string, error) {
	fmt.Fprint(r.out, prompt)
	select {
	case line, ok := <-r.lines:
		if !ok {
			return "", io.EOF
		}
		return line, nil
	case <-r.ctx.Done():
		return "", r.ctx.Err()
	}
}

// newTerminalReader is the line editor on the terminal in and out, with its history in the
// user's home directory; it completes once given the session's Complete.
func newTerminalReader(ctx context.Context, in, out *os.File) *editor {
	chunks := make(chan []byte, 64)
	go func() {
		defer close(chunks)
		buf := make([]byte, 4096) // a paste arrives in few chunks, each drawn once
		for {
			n, err := in.Read(buf)
			if n > 0 {
				chunks <- append([]byte(nil), buf[:n]...)
			}
			if err != nil {
				return
			}
		}
	}()
	th := &ansiTheme
	if _, noColor := os.LookupEnv("NO_COLOR"); noColor {
		th = &plainTheme
	}
	path := ""
	if home, err := os.UserHomeDir(); err == nil {
		path = filepath.Join(home, ".ant_history")
	}
	return &editor{
		ctx:       ctx,
		in:        chunks,
		out:       out,
		size:      func() (int, int) { cols, rows, _ := windowSize(out); return cols, rows },
		th:        th,
		hist:      loadHistory(path),
		raw:       func() (func(), error) { return makeRaw(in) },
		bracketed: true,
	}
}
