package interp

import (
	"errors"
	"syscall"
)

// errorNoData is ERROR_NO_DATA, which a write gives when the pipe is being closed; syscall
// does not name it.
const errorNoData = syscall.Errno(232)

// isBrokenPipe reports whether err is a write to a pipe whose reader is gone: Go's EPIPE, or
// either error Windows gives for it.
func isBrokenPipe(err error) bool {
	return errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ERROR_BROKEN_PIPE) ||
		errors.Is(err, errorNoData)
}
