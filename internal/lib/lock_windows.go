//go:build windows

package lib

import (
	"os"
	"syscall"
	"unsafe"
)

var procLockFileEx = syscall.NewLazyDLL("kernel32.dll").NewProc("LockFileEx")

const (
	lockfileFailImmediately = 0x1
	lockfileExclusiveLock   = 0x2
	errorLockViolation      = syscall.Errno(33)
)

// tryLockFile takes an exclusive LockFileEx lock on the whole of f without waiting, and reports
// held when another handle holds it. The lock belongs to the handle, and closing it releases it.
// Windows locks are mandatory: other processes cannot read or write the locked file.
func tryLockFile(f *os.File) (held bool, err error) {
	var ol syscall.Overlapped
	r, _, e := procLockFileEx.Call(f.Fd(), lockfileExclusiveLock|lockfileFailImmediately, 0,
		0xFFFFFFFF, 0xFFFFFFFF, uintptr(unsafe.Pointer(&ol)))
	if r != 0 {
		return false, nil
	}
	if e == errorLockViolation {
		return true, nil
	}
	return false, e
}

// fileIsTerminal reports whether f is a console.
func fileIsTerminal(f *os.File) bool {
	var mode uint32
	return syscall.GetConsoleMode(syscall.Handle(f.Fd()), &mode) == nil
}
