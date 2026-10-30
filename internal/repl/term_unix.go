//go:build darwin || dragonfly || freebsd || netbsd || openbsd || linux

package repl

import (
	"errors"
	"os"
	"syscall"
	"unsafe"
)

// isTerminal reports whether f is a terminal: whether it has terminal attributes to read. It
// reads them through SyscallConn, since Fd would put a pipe into blocking mode.
func isTerminal(f *os.File) bool {
	_, err := termios(f)
	return err == nil
}

func termios(f *os.File) (syscall.Termios, error) {
	var t syscall.Termios
	c, err := f.SyscallConn()
	if err != nil {
		return t, err
	}
	var errno syscall.Errno
	c.Control(func(fd uintptr) {
		_, _, errno = syscall.Syscall(syscall.SYS_IOCTL, fd, getTermios, uintptr(unsafe.Pointer(&t)))
	})
	if errno != 0 {
		return t, errno
	}
	return t, nil
}

func setTermios(f *os.File, t *syscall.Termios) error {
	c, err := f.SyscallConn()
	if err != nil {
		return err
	}
	var errno syscall.Errno
	c.Control(func(fd uintptr) {
		_, _, errno = syscall.Syscall(syscall.SYS_IOCTL, fd, setTermiosReq, uintptr(unsafe.Pointer(t)))
	})
	if errno != 0 {
		return errno
	}
	return nil
}

// makeRaw puts the terminal f into raw mode: keys arrive one at a time, unechoed, and Ctrl-C is
// a key rather than a signal. Output processing stays on, so a line break still returns the
// carriage. It returns the function that restores the terminal.
func makeRaw(f *os.File) (restore func(), err error) {
	old, err := termios(f)
	if err != nil {
		return nil, err
	}
	raw := old
	raw.Iflag &^= syscall.BRKINT | syscall.ICRNL | syscall.INPCK | syscall.ISTRIP | syscall.IXON
	raw.Lflag &^= syscall.ECHO | syscall.ICANON | syscall.IEXTEN | syscall.ISIG
	raw.Cc[syscall.VMIN] = 1
	raw.Cc[syscall.VTIME] = 0
	if err := setTermios(f, &raw); err != nil {
		return nil, err
	}
	return func() { setTermios(f, &old) }, nil
}

// windowSize is the terminal's columns and rows, or an error when f is not a terminal that
// reports them.
func windowSize(f *os.File) (cols, rows int, err error) {
	var ws struct{ rows, cols, x, y uint16 }
	c, err := f.SyscallConn()
	if err != nil {
		return 0, 0, err
	}
	var errno syscall.Errno
	c.Control(func(fd uintptr) {
		_, _, errno = syscall.Syscall(syscall.SYS_IOCTL, fd, syscall.TIOCGWINSZ, uintptr(unsafe.Pointer(&ws)))
	})
	if errno != 0 {
		return 0, 0, errno
	}
	if ws.cols == 0 {
		return 0, 0, errors.New("no window size")
	}
	return int(ws.cols), int(ws.rows), nil
}
