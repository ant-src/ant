//go:build unix && !(darwin || dragonfly || freebsd || netbsd || openbsd || linux)

package lib

import "os"

func isTerminal(f *os.File) bool { return false }
