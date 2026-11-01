//go:build !(darwin || dragonfly || freebsd || netbsd || openbsd || linux || windows)

package lib

import (
	"errors"
	"os"
)

// tryLockFile has no lock to take on this system.
func tryLockFile(f *os.File) (held bool, err error) {
	return false, errors.New("file locks are not supported on this system")
}
