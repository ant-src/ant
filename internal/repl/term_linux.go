//go:build linux

package repl

import "syscall"

const (
	getTermios    = syscall.TCGETS
	setTermiosReq = syscall.TCSETS
)
