package main

import (
	"bufio"
	"fmt"
	"io"
)

// Loop stands in for a library that owns the loop a program runs in, as a web server or a
// terminal UI framework does: it calls Handle with each line it reads, and recovers a panic of
// Handle, reports it, and goes on, so that one bad event does not end the program.
type Loop struct {
	Handle  func(line string)
	Log     io.Writer
	stopped bool
}

// Stop ends Run after the current line.
func (l *Loop) Stop() { l.stopped = true }

// Run reads r to its end, or until Stop.
func (l *Loop) Run(r io.Reader) error {
	lines := bufio.NewScanner(r)
	for !l.stopped && lines.Scan() {
		l.handle(lines.Text())
	}
	return lines.Err()
}

func (l *Loop) handle(line string) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintln(l.Log, "loop: recovered from a panic")
		}
	}()
	l.Handle(line)
}
