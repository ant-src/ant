// Command ant runs, checks, and evaluates Ant programs.
//
//	ant run file.ant [args...]   run a program
//	ant check file.ant           type-check a program and every module it imports
//	ant                          start a REPL
//
// Exit status: 0 when main returns or the check passes, 1 when the main goroutine panics or the
// check finds errors, 2 when the program cannot be loaded or the usage is wrong, 130 when an
// interrupt the program did not ask for ends the run, with nothing printed, as a shell does.
// Os.exit sets its own status.
package main

import (
	_ "time/tzdata" // the zone database, for hosts that have none

	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/ant-src/ant"
	replpkg "github.com/ant-src/ant/internal/repl"
)

const usage = `usage:
  ant run <file.ant> [args...]
  ant check <file.ant>
  ant`

func main() {
	os.Exit(run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		return repl(stdin, stdout, stderr)
	}
	switch args[0] {
	case "run":
		if len(args) < 2 {
			fmt.Fprintln(stderr, usage)
			return 2
		}
		prog, err := ant.Load(args[1])
		if err != nil {
			fmt.Fprintln(stderr, err)
			return ant.ExitStatus(err)
		}
		// An interrupt goes to the program when it asked for one (Os.interrupt), and ends the
		// run otherwise, or at the second; the run decides, the command only relays.
		sigs := make(chan os.Signal, 1)
		signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)
		defer signal.Stop(sigs)
		interrupts := make(chan struct{}, 1)
		go func() {
			for range sigs {
				select {
				case interrupts <- struct{}{}:
				default:
				}
			}
		}()
		err = prog.Run(context.Background(), ant.WithStdin(stdin), ant.WithStdout(stdout), ant.WithStderr(stderr),
			ant.WithArgs(args[2:]), ant.WithInterrupts(interrupts))
		switch {
		case err == nil, errors.Is(err, ant.ErrInterrupted):
			// An interrupt that ends the run says nothing, as a shell says nothing.
		default:
			var pe *ant.PanicError
			var ee *ant.ExitError
			if !errors.As(err, &pe) && !errors.As(err, &ee) {
				// A panic has already been reported by the runtime; an exit says nothing.
				fmt.Fprintln(stderr, err)
			}
		}
		return ant.ExitStatus(err)
	case "check":
		if len(args) != 2 {
			fmt.Fprintln(stderr, usage)
			return 2
		}
		prog, err := ant.Load(args[1])
		if err != nil {
			fmt.Fprintln(stderr, err)
			return ant.ExitStatus(err)
		}
		diags := prog.Check()
		for _, d := range diags {
			fmt.Fprintln(stderr, d)
		}
		if len(diags) > 0 {
			return 1
		}
		return 0
	case "help", "-h", "--help":
		fmt.Fprintln(stdout, usage)
		return 0
	default:
		fmt.Fprintln(stderr, usage)
		return 2
	}
}

func repl(stdin io.Reader, stdout, stderr io.Writer) int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	return replpkg.Run(ctx, stdin, stdout, stderr)
}
