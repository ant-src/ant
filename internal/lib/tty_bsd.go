//go:build darwin || dragonfly || freebsd || netbsd || openbsd

package lib

import "syscall"

const getTermios = syscall.TIOCGETA
