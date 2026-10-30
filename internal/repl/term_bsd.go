//go:build darwin || dragonfly || freebsd || netbsd || openbsd

package repl

import "syscall"

const (
	getTermios    = syscall.TIOCGETA
	setTermiosReq = syscall.TIOCSETA
)
