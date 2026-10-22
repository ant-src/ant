// Command ant runs, checks, and evaluates Ant programs.
//
//	ant run file.ant [args...]   run a program
//	ant run --cpuprofile cpu.out --memprofile mem.out file.ant [args...]
//	                             run a program and write Go pprof profiles of the interpreter
//	ant check file.ant           type-check a program and every module it imports
//	ant                          start a REPL
//
// Exit status: 0 when main returns or the check passes, 1 when the main goroutine panics or the
// check finds errors, 2 when the program cannot be loaded or the usage is wrong. Os.exit sets
// its own status.
package main

import (
	_ "time/tzdata" // the zone database, for hosts that have none

	"context"
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime"
	"runtime/pprof"
	"strings"

	"github.com/ant-src/ant"
	replpkg "github.com/ant-src/ant/internal/repl"
)

const usage = `usage:
  ant run [--cpuprofile <file>] [--memprofile <file>] <file.ant> [args...]
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
		args = args[1:]
		var cpuProfile, memProfile string
		for len(args) >= 2 && strings.HasPrefix(args[0], "--") {
			switch args[0] {
			case "--cpuprofile":
				cpuProfile = args[1]
			case "--memprofile":
				memProfile = args[1]
			default:
				fmt.Fprintln(stderr, usage)
				return 2
			}
			args = args[2:]
		}
		if len(args) < 1 {
			fmt.Fprintln(stderr, usage)
			return 2
		}
		prog, err := ant.Load(args[0])
		if err != nil {
			fmt.Fprintln(stderr, err)
			return ant.ExitStatus(err)
		}
		if cpuProfile != "" {
			f, err := os.Create(cpuProfile)
			if err != nil {
				fmt.Fprintln(stderr, err)
				return 2
			}
			defer f.Close()
			if err := pprof.StartCPUProfile(f); err != nil {
				fmt.Fprintln(stderr, err)
				return 2
			}
			defer pprof.StopCPUProfile()
		}
		if memProfile != "" {
			runtime.MemProfileRate = 4096
			defer writeHeapProfile(memProfile, stderr)
		}
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
		defer stop()
		err = prog.Run(ctx, ant.WithStdin(stdin), ant.WithStdout(stdout), ant.WithStderr(stderr),
			ant.WithArgs(args[1:]))
		switch err.(type) {
		case nil, *ant.PanicError, *ant.ExitError:
			// A panic has already been reported by the runtime; an exit says nothing.
		default:
			fmt.Fprintln(stderr, err)
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

// writeHeapProfile writes the allocation profile of the run; read it with
// go tool pprof -sample_index=alloc_objects.
func writeHeapProfile(path string, stderr io.Writer) {
	f, err := os.Create(path)
	if err != nil {
		fmt.Fprintln(stderr, err)
		return
	}
	defer f.Close()
	if err := pprof.Lookup("allocs").WriteTo(f, 0); err != nil {
		fmt.Fprintln(stderr, err)
	}
}

func repl(stdin io.Reader, stdout, stderr io.Writer) int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	return replpkg.Run(ctx, stdin, stdout, stderr)
}
