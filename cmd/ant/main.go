// Command ant runs, checks, and evaluates Ant programs.
//
//	ant run file.ant [args...]   run a program
//	ant file.ant [args...]       the same, for a #!/usr/bin/env ant line
//	ant check [-json] [-strict] [path...]  type-check the files under the paths and every module they import
//	ant test [flags] [path...]   run the tests of the test files under the paths (spec/test.md)
//	ant fmt [--check] path...    format files in place, or list those that are not formatted
//	ant doc [file.ant] [name]    print a declaration's signature and doc comment
//	ant code <subcommand> ...    find declarations, uses, and types, and rename
//	ant lsp                      serve the Language Server Protocol to an editor
//	ant dap                      serve the Debug Adapter Protocol to an editor
//	ant --comp bash|zsh|fish     print the script that completes ant's commands in the shell
//	ant version                  print the version of ant
//	ant                          start a REPL
//
// Exit status: 0 when the script ends or the check passes, 1 when the script returns an Err, the
// main goroutine panics, or the check finds errors, 2 when the program cannot be loaded or the usage is wrong. When an
// interrupt the program did not ask for ends the run, the command prints nothing and ends by that
// signal, SIGINT or SIGTERM, as a shell's child does, or exits with 130 where there are none.
// Os.exit sets its own status. ant fmt exits 1 when --check finds a file to change and 2 when a
// file cannot be read, parsed, or written. ant doc exits 1 when there is no such declaration.
// ant code exits 1 when nothing was found, a location is ambiguous, or a rename was refused, and
// 2 when the usage is wrong, a file cannot be read, parsed, or written, or an answer is incomplete.
// ant lsp exits 0 when the client sends shutdown and then exit, 1 when the connection ends
// otherwise. ant dap exits 0 when the client disconnects, 1 when the connection ends first.
//
// ant check, and ant run when it cannot load or run a program, print each error as path:line:col:
// message, then the lines of code it is about, numbered, with what it covers underlined (the
// range the language server reports), then how many errors there were. A span longer than four
// lines shows its first two and its last. On a terminal the code is highlighted as the REPL
// highlights it; NO_COLOR or a dumb terminal turns the colors off. ant check -json instead
// writes each error to standard output as a JSON object on a line of its own.
//
// ant doc and the REPL's :doc highlight signatures and examples on a terminal, preserving
// source line breaks and spacing. NO_COLOR, a dumb terminal, or redirected output disables
// highlighting.
//
// The REPL evaluates each input on one session and prints every non-void result with
// String.from, on a terminal colored and, when it does not fit, an element to a line; _ reads
// the last result. An input goes on over more lines while it is unfinished, which on a terminal
// are edited together, indented after an opening brace, and recalled whole from the history. On a terminal it edits
// each line with the keys of readline (arrows, Home and End or Ctrl-A and Ctrl-E, Alt-B and
// Alt-F or Ctrl-arrows by word, Ctrl-W, Ctrl-K, Ctrl-U, Ctrl-L), highlights it as it is typed,
// and completes what can be written at the cursor: the names declared so far and the prelude's,
// members after a dot, the functions that take the left side after a pipe, the parameters of a
// call, the fields of a record, by its type or by the value a variable holds, and import paths.
// The first completion is suggested after the
// cursor, where Right or End takes it; Tab writes the only completion or the common prefix,
// and otherwise lists them and cycles through them, Shift-Tab backwards. Up and Down browse
// the history, kept in ~/.ant_history. Ctrl-C drops the line, and twice in a row ends the REPL
// with 130, as Ctrl-D on an empty line ends it with 0; Ctrl-C while an input runs ends it with
// 130, as it ends a run. NO_COLOR turns the colors off; without a terminal, or on a dumb one,
// the REPL reads lines as they come. A line that begins with a colon is a command: :type expr
// prints a type, :doc name a declaration of the standard library, :load file.ant evaluates a
// file, :reset starts a new session, :quit ends it, and :help lists them with the keys.
package main

import (
	_ "time/tzdata" // the zone database, for hosts that have none

	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/signal"
	"strings"

	"github.com/ant-src/ant"
	replpkg "github.com/ant-src/ant/internal/repl"
)

const usage = `usage:
  ant run <file.ant> [args...]
  ant <file.ant> [args...]
  ant check [-json] [-strict] [path...]
  ant check -explain <code>
  ant test [flags] [path...]
  ant fmt [--check] <path>...   (- for standard input)
  ant doc <name> | <file.ant> [name]
  ant code def [flags] <loc>
  ant code refs [flags] <loc>
  ant code type [flags] <loc>
  ant code symbols [flags] [path...]
  ant code rename [flags] <loc> <name>
  ant lsp
  ant dap
  ant --comp bash|zsh|fish
  ant version | --version
  ant help | -h | --help
  ant`

func main() {
	if completion(os.Args[1:], os.Stdout) {
		return
	}
	status := run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr)
	if endSignal != nil {
		dieBy(endSignal)
	}
	os.Exit(status)
}

// endSignal is the signal an interrupt that ended the run came by, which the command ends by.
var endSignal os.Signal

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		return repl(stdin, stdout, stderr)
	}
	if len(args) > 1 && (args[1] == "-h" || args[1] == "--help") && args[0] != "test" {
		if lines := commandUsage(args[0]); lines != "" {
			// -h or --help right after a command prints its usage; ant test prints its flags too
			fmt.Fprintln(stdout, lines)
			return 0
		}
	}
	switch args[0] {
	case "run":
		if len(args) < 2 {
			fmt.Fprintln(stderr, usage)
			return 2
		}
		prog, err := ant.Load(args[1])
		if err != nil {
			printLoadError(stderr, err)
			return ant.ExitStatus(err)
		}
		// An interrupt goes to the program when it asked for one (Os.interrupt), and ends the
		// run otherwise, or at the second; the run decides, the command only relays.
		r := newRelay()
		defer r.stop()
		err = prog.Run(context.Background(), ant.WithStdin(stdin), ant.WithStdout(stdout), ant.WithStderr(stderr),
			ant.WithArgs(args[2:]), ant.WithInterrupts(r.interrupts), ant.WithForeground(r.hold))
		switch {
		case err == nil:
		case errors.Is(err, ant.ErrInterrupted):
			// An interrupt that ends the run says nothing, as a shell says nothing.
			endSignal = r.ending()
		default:
			var pe *ant.PanicError
			var ee *ant.ExitError
			var re *ant.ResultError
			var le *ant.LoadError
			if errors.As(err, &le) {
				// a declaration file, which Run refuses
				printLoadError(stderr, err)
			} else if !errors.As(err, &pe) && !errors.As(err, &ee) && !errors.As(err, &re) {
				// A panic or a returned Err has already been reported by the runtime; an exit
				// says nothing.
				fmt.Fprintln(stderr, err)
			}
		}
		return ant.ExitStatus(err)
	case "check":
		return checkCommand(args[1:], stdout, stderr)
	case "test":
		return testCommand(args[1:], stdin, stdout, stderr)
	case "fmt":
		return fmtCommand(args[1:], stdin, stdout, stderr)
	case "doc":
		return docCommand(args[1:], stdout, stderr)
	case "code":
		return codeCommand(args[1:], stdout, stderr)
	case "lsp":
		return lspCommand(args[1:], stdin, stdout, stderr)
	case "dap":
		return dapCommand(args[1:], stdin, stdout, stderr)
	case "--comp":
		return compCommand(args[1:], stdout, stderr)
	case "version", "--version":
		if len(args) > 1 {
			fmt.Fprintln(stderr, commandUsage("version"))
			return 2
		}
		fmt.Fprintln(stdout, "ant", antVersion())
		return 0
	case "help", "-h", "--help":
		fmt.Fprintln(stdout, usage)
		return 0
	default:
		// a script, also one installed without its .ant suffix and run through its #! line
		if info, err := os.Stat(args[0]); strings.HasSuffix(args[0], ".ant") || err == nil && info.Mode().IsRegular() {
			return run(append([]string{"run"}, args...), stdin, stdout, stderr)
		}
		fmt.Fprintln(stderr, usage)
		return 2
	}
}

// commandUsage is the usage line of one command, "usage: ant check <file.ant>", or "" for a name
// that is no command.
func commandUsage(name string) string {
	var lines []string
	for _, l := range strings.Split(usage, "\n")[1:] {
		l = strings.TrimSpace(l)
		if l == "ant "+name || strings.HasPrefix(l, "ant "+name+" ") {
			lines = append(lines, l)
		}
	}
	if len(lines) == 0 {
		return ""
	}
	return "usage: " + strings.Join(lines, "\n       ")
}

func repl(stdin io.Reader, stdout, stderr io.Writer) int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	return replpkg.Run(ctx, stdin, stdout, stderr)
}
