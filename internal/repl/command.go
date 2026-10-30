package repl

import (
	"fmt"
	"os"
	"strings"

	"github.com/ant-src/ant"
	"github.com/ant-src/ant/internal/doc"
	"github.com/ant-src/ant/internal/prelude"
)

// A command is a line of the REPL that begins with a colon.
type command struct {
	name, args, help string
}

var commands = []command{
	{"help", "", "list the commands and the keys"},
	{"type", "expr", "print the type of an expression"},
	{"doc", "name", "print a declaration of the standard library with its doc"},
	{"load", "file.ant", "evaluate a file in the session"},
	{"reset", "", "start a new session, forgetting every name"},
	{"quit", "", "end the REPL"},
}

const keysHelp = `Keys:
  Tab, Shift-Tab        complete; cycle through the completions
  Right, End            take the completion shown after the cursor
  Up, Down              browse the history
  Ctrl-A, Ctrl-E        start, end of the line
  Alt-B, Alt-F          a word left, right
  Ctrl-W, Ctrl-K, Ctrl-U  delete a word, to the end, to the start
  Ctrl-L                clear the screen
  Ctrl-C                drop the line; twice, end the REPL
  Ctrl-D                end the REPL on an empty line`

// isCommand reports whether a line is a command.
func isCommand(line string) bool {
	return strings.HasPrefix(strings.TrimSpace(line), ":")
}

// command runs a command line. done reports that the REPL ends, with status.
func (l *looper) command(line string) (status int, done bool) {
	name, arg, _ := strings.Cut(strings.TrimSpace(line)[1:], " ")
	arg = strings.TrimSpace(arg)
	var found *command
	for i := range commands {
		if commands[i].name == name {
			found = &commands[i]
		}
	}
	if name == "exit" || name == "q" {
		found = &commands[len(commands)-1]
	}
	if found == nil {
		fmt.Fprintf(l.stderr, "unknown command :%s; :help lists them\n", name)
		return 0, false
	}
	if found.args != "" && arg == "" {
		fmt.Fprintf(l.stderr, "usage: :%s %s\n", found.name, found.args)
		return 0, false
	}
	switch found.name {
	case "help":
		fmt.Fprintln(l.stdout, "Commands:")
		for _, c := range commands {
			usage := ":" + c.name
			if c.args != "" {
				usage += " " + c.args
			}
			fmt.Fprintf(l.stdout, "  %-20s  %s\n", usage, c.help)
		}
		fmt.Fprintln(l.stdout, keysHelp)
	case "type":
		t, err := l.s.TypeOf(arg)
		if err != nil {
			fmt.Fprintln(l.stderr, err)
			return 0, false
		}
		fmt.Fprintln(l.stdout, t)
	case "doc":
		f, err := prelude.Decls()
		if err != nil {
			fmt.Fprintln(l.stderr, err)
			return 0, false
		}
		if !doc.Print(l.stdout, f.Decls, arg, true) {
			fmt.Fprintf(l.stderr, "no declaration %s\n", arg)
		}
	case "load":
		src, err := os.ReadFile(arg)
		if err != nil {
			fmt.Fprintln(l.stderr, err)
			return 0, false
		}
		return l.eval(string(src))
	case "reset":
		s, err := l.open()
		if err != nil {
			fmt.Fprintln(l.stderr, err)
			return 0, false
		}
		l.s.Close()
		l.s = s
	case "quit":
		return 0, true
	}
	return 0, false
}

// commandCompletions completes a command line: the command's name, and the expression or the
// name after :type and :doc. src and off are what to complete instead, when ok.
func commandCompletions(line string, off int) (items []ant.Completion, src string, srcOff int, ok bool) {
	name, _, spaced := strings.Cut(line, " ")
	if !spaced || off <= len(name) {
		for _, c := range commands {
			items = append(items, ant.Completion{Label: c.name, Insert: c.name, Kind: "keyword", Detail: c.help})
		}
		return items, "", 0, false
	}
	switch name {
	case ":type", ":doc":
		return nil, line[len(name)+1:], off - len(name) - 1, true
	}
	return nil, "", 0, false
}
