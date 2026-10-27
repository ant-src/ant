package main

import (
	_ "embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
)

// The shell scripts that `ant --comp <shell>` prints. Each asks ant itself what to complete,
// passing the command line in an environment variable, so a new command or flag needs no change
// here, only an entry in commands.
var (
	//go:embed complete.bash
	bashScript string
	//go:embed complete.zsh
	zshScript string
	//go:embed complete.fish
	fishScript string
)

// command is one of ant's commands as completion sees it.
type command struct {
	name  string
	flags []string
	// args lists the candidates for the nth argument after the flags, or nil for none.
	args func(n int, word string) []candidate
}

// commands are ant's commands, in the order the shell offers them. A flag a command accepts
// belongs in its flags here too.
var commands = []command{
	{name: "run", args: func(n int, word string) []candidate {
		if n == 0 {
			return files(word, ".ant")
		}
		return files(word, "") // the program's own arguments
	}},
	{name: "check", args: func(n int, word string) []candidate {
		if n == 0 {
			return files(word, ".ant")
		}
		return nil
	}},
	{name: "fmt", flags: []string{"--check"}, args: func(n int, word string) []candidate {
		return files(word, ".ant")
	}},
	{name: "doc", args: func(n int, word string) []candidate {
		if n == 0 {
			return files(word, ".ant")
		}
		return nil
	}},
	{name: "help"},
}

// flags are those ant takes in place of a command, as run's switch accepts them.
var flags = []string{"-h", "--help", "--comp"}

var shells = []string{"bash", "zsh", "fish"}

type candidate struct {
	value string
	kind  string // "file", "dir", or "word"
}

// completion answers a shell that asks what to complete, as the scripts in complete.* do: bash
// sets COMP_LINE and COMP_POINT and passes the command, the word, and the word before it as
// arguments, zsh sets COMP_ZSH and fish COMP_FISH to the command up to the cursor. It reports
// whether it answered.
func completion(args []string, stdout io.Writer) bool {
	if line, ok := os.LookupEnv("COMP_LINE"); ok && len(args) >= 2 {
		if point, err := strconv.Atoi(os.Getenv("COMP_POINT")); err == nil && point >= 0 && point <= len(line) {
			line = line[:point]
		}
		reply(stdout, "bash", complete(line))
		return true
	}
	if line, ok := os.LookupEnv("COMP_ZSH"); ok {
		reply(stdout, "zsh", complete(line))
		return true
	}
	if line, ok := os.LookupEnv("COMP_FISH"); ok {
		reply(stdout, "fish", complete(line))
		return true
	}
	return false
}

// compCommand is ant --comp <shell>: it prints the script that wires the shell's completion to
// ant.
func compCommand(args []string, stdout, stderr io.Writer) int {
	if len(args) != 1 {
		fmt.Fprintln(stderr, "usage: ant --comp bash|zsh|fish")
		return 2
	}
	switch args[0] {
	case "bash":
		fmt.Fprint(stdout, bashScript)
	case "zsh":
		fmt.Fprint(stdout, zshScript)
	case "fish":
		fmt.Fprint(stdout, fishScript)
	default:
		fmt.Fprintf(stderr, "unknown shell %q; want bash, zsh, or fish\n", args[0])
		return 2
	}
	return 0
}

// complete lists the candidates for the word at the end of line, which ends at the cursor.
// Words before ant, such as FOO=1 or sudo, are skipped.
func complete(line string) []candidate {
	before, word := splitWords(line)
	if len(before) == 0 {
		return nil // not even ant yet
	}
	i := slices.IndexFunc(before, func(w string) bool { return filepath.Base(w) == "ant" })
	before = before[max(i, 0)+1:]
	if len(before) == 0 {
		if strings.HasPrefix(word, "-") {
			return words(flags, word)
		}
		var names []string
		for _, c := range commands {
			names = append(names, c.name)
		}
		return append(words(names, word), files(word, ".ant")...)
	}
	if before[0] == "--comp" {
		if len(before) == 1 {
			return words(shells, word)
		}
		return nil
	}
	if strings.HasSuffix(before[0], ".ant") {
		return files(word, "") // ant file.ant args...: the program's own arguments
	}
	for _, c := range commands {
		if c.name != before[0] {
			continue
		}
		n := 0
		for _, w := range before[1:] {
			if !strings.HasPrefix(w, "-") || w == "-" {
				n++
			}
		}
		if n == 0 && strings.HasPrefix(word, "-") {
			return words(c.flags, word)
		}
		if c.args == nil {
			return nil
		}
		return c.args(n, word)
	}
	return nil
}

// words is those of list that start with prefix.
func words(list []string, prefix string) []candidate {
	var out []candidate
	for _, w := range list {
		if strings.HasPrefix(w, prefix) {
			out = append(out, candidate{w, "word"})
		}
	}
	return out
}

// files lists the directories and the files ending in ext (any file, when ext is empty) whose
// path starts with word, written as word writes its directory, ~/ included. A name starting with
// a dot is listed only when word asks for one; a name with a tab or a line break, which the
// replies cannot carry, is not listed.
func files(word, ext string) []candidate {
	if word == "~" {
		word = "~/"
	}
	dir, base := filepath.Split(word)
	read := dir
	if strings.HasPrefix(dir, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil
		}
		read = home + dir[1:]
	} else if strings.HasPrefix(word, "~") {
		return nil // ~user is not supported
	}
	if read == "" {
		read = "."
	}
	entries, err := os.ReadDir(read)
	if err != nil {
		return nil
	}
	var out []candidate
	for _, e := range entries {
		name := e.Name()
		if !strings.HasPrefix(name, base) || strings.HasPrefix(name, ".") && !strings.HasPrefix(base, ".") ||
			strings.ContainsAny(name, "\t\n") {
			continue
		}
		value := dir + name
		isDir := e.IsDir()
		if e.Type()&os.ModeSymlink != 0 {
			if info, err := os.Stat(filepath.Join(read, name)); err == nil {
				isDir = info.IsDir()
			}
		}
		switch {
		case isDir:
			out = append(out, candidate{value, "dir"})
		case ext == "" || strings.HasSuffix(name, ext):
			out = append(out, candidate{value, "file"})
		}
	}
	return out
}

// reply writes the candidates as the shell's script reads them: for zsh one line each of the
// name to show, the value to insert, and the kind, separated by tabs; for bash and fish the value
// alone, with fish given a directory's trailing slash, which bash adds itself.
func reply(w io.Writer, shell string, cands []candidate) {
	for _, c := range cands {
		switch shell {
		case "zsh":
			display := c.value
			if c.kind != "word" {
				display = filepath.Base(c.value)
			}
			if c.kind == "dir" {
				display += "/"
			}
			fmt.Fprintf(w, "%s\t%s\t%s\n", display, c.value, c.kind)
		case "fish":
			if c.kind == "dir" {
				fmt.Fprintln(w, c.value+"/")
			} else {
				fmt.Fprintln(w, c.value)
			}
		default:
			fmt.Fprintln(w, c.value)
		}
	}
}

// splitWords splits a command line into words as the shell does, removing quotes and
// backslashes, and returns apart the word at the end of the line, "" when the line ends between
// words. An unclosed quote runs to the end of the line.
func splitWords(line string) (before []string, last string) {
	var b strings.Builder
	inWord := false
	var quote rune
	escaped := false
	for _, r := range line {
		switch {
		case escaped:
			b.WriteRune(r)
			escaped = false
		case r == '\\' && quote != '\'':
			escaped, inWord = true, true
		case quote != 0 && r == quote:
			quote = 0
		case quote != 0:
			b.WriteRune(r)
		case r == '\'' || r == '"':
			quote, inWord = r, true
		case r == ' ' || r == '\t' || r == '\n':
			if inWord {
				before = append(before, b.String())
				b.Reset()
				inWord = false
			}
		default:
			b.WriteRune(r)
			inWord = true
		}
	}
	return before, b.String()
}
