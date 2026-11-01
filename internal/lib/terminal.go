//go:build !windows

package lib

import "os"

// fileIsTerminal reports whether f is a terminal, for File.isTerminal.
func fileIsTerminal(f *os.File) bool { return isTerminal(f) }
