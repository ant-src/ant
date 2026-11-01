//go:build darwin || dragonfly || freebsd || netbsd || openbsd || linux

package lib

import (
	"errors"
	"os"
	"syscall"
)

// tryLockFile takes f's exclusive flock without waiting, and reports held when another open file
// holds it. A flock belongs to the open file, not the process, so a second open of the same path
// conflicts with the first even in this process, and closing f releases it.
func tryLockFile(f *os.File) (held bool, err error) {
	c, err := f.SyscallConn()
	if err != nil {
		return false, err
	}
	var lockErr error
	if err := c.Control(func(fd uintptr) {
		for {
			lockErr = syscall.Flock(int(fd), syscall.LOCK_EX|syscall.LOCK_NB)
			if !errors.Is(lockErr, syscall.EINTR) {
				return
			}
		}
	}); err != nil {
		return false, err
	}
	if errors.Is(lockErr, syscall.EWOULDBLOCK) {
		return true, nil
	}
	return false, lockErr
}
