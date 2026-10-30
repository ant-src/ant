package repl

import (
	"os"
	"strings"
)

// historySize is how many lines the history keeps, in memory and in its file.
const historySize = 1000

// A history is the lines entered, oldest first, kept in a file across sessions when it has one.
type history struct {
	lines []string
	path  string
}

// loadHistory reads the history file at path, or starts an empty history when there is none;
// "" keeps the history in memory only.
func loadHistory(path string) *history {
	h := &history{path: path}
	if path == "" {
		return h
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return h
	}
	for _, line := range strings.Split(strings.TrimRight(string(data), "\n"), "\n") {
		if line != "" {
			h.lines = append(h.lines, line)
		}
	}
	if len(h.lines) > historySize {
		h.lines = h.lines[len(h.lines)-historySize:]
	}
	return h
}

// add appends a line, unless it is blank or repeats the last one, and writes it to the file.
func (h *history) add(line string) {
	if strings.TrimSpace(line) == "" || len(h.lines) > 0 && h.lines[len(h.lines)-1] == line {
		return
	}
	h.lines = append(h.lines, line)
	if len(h.lines) > historySize {
		h.lines = h.lines[1:]
	}
	if h.path == "" {
		return
	}
	f, err := os.OpenFile(h.path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0o600)
	if err != nil {
		return
	}
	f.WriteString(line + "\n")
	f.Close()
}
