//go:build linux

package lib

import "syscall"

const getTermios = syscall.TCGETS
