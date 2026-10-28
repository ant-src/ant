//go:build darwin || dragonfly || freebsd || netbsd || openbsd || linux

package lib

import (
	"os"
	"syscall"
	"unsafe"
)

// isTerminal reports whether f is a terminal: whether it has terminal attributes to read. It
// reads them through SyscallConn, since Fd would put a pipe into blocking mode.
func isTerminal(f *os.File) bool {
	c, err := f.SyscallConn()
	if err != nil {
		return false
	}
	var errno syscall.Errno
	c.Control(func(fd uintptr) {
		var t syscall.Termios
		_, _, errno = syscall.Syscall(syscall.SYS_IOCTL, fd, getTermios, uintptr(unsafe.Pointer(&t)))
	})
	return errno == 0
}
