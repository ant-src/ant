//go:build !windows

package interp

import (
	"errors"
	"syscall"
)

// isBrokenPipe reports whether err is a write to a pipe whose reader is gone.
func isBrokenPipe(err error) bool { return errors.Is(err, syscall.EPIPE) }
