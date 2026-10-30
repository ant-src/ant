//go:build !(darwin || dragonfly || freebsd || netbsd || openbsd || linux)

package repl

import (
	"errors"
	"os"
)

// On a platform without termios the REPL reads lines as they come, without the line editor.

func isTerminal(f *os.File) bool { return false }

func makeRaw(f *os.File) (restore func(), err error) { return nil, errors.New("no terminal") }

func windowSize(f *os.File) (cols, rows int, err error) { return 0, 0, errors.New("no terminal") }
