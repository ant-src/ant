package ant

import (
	"errors"
	"fmt"
	"strings"
)

// Position is a location in a source file. Line and Col are 1-based; Col counts bytes.
type Position struct {
	Path string
	Line int
	Col  int
}

func (p Position) String() string {
	if p.Line == 0 {
		return p.Path
	}
	return fmt.Sprintf("%s:%d:%d", p.Path, p.Line, p.Col)
}

// Diagnostic is one error found while loading or checking a program.
type Diagnostic struct {
	Pos     Position
	Message string
}

func (d Diagnostic) Error() string {
	if d.Pos.Path == "" {
		return d.Message
	}
	return d.Pos.String() + ": " + d.Message
}

// LoadError reports why a program could not be loaded: a syntax error, a missing module, an
// import cycle, or a missing main. The ant command exits with status 2.
type LoadError struct {
	Diagnostics []Diagnostic
}

func (e *LoadError) Error() string {
	lines := make([]string, len(e.Diagnostics))
	for i, d := range e.Diagnostics {
		lines[i] = d.Error()
	}
	return strings.Join(lines, "\n")
}

// Frame is one entry of the stack trace a panic carries, innermost first.
type Frame struct {
	Function string
	Pos      Position
}

// PanicError is returned by Run when a panic ended the main goroutine. The runtime has already
// written the message and trace to the run's stderr writer. The ant command exits with status 1.
type PanicError struct {
	Message string
	Trace   []Frame
}

func (e *PanicError) Error() string { return "panic: " + e.Message }

// ExitError is returned by Run when the program called Os.exit.
type ExitError struct {
	Status int
}

func (e *ExitError) Error() string { return fmt.Sprintf("exit status %d", e.Status) }

// ExitStatus maps the error returned by Load or Run to the exit status the ant command uses:
// 0 for nil, 2 for a LoadError, 1 for a PanicError, the status of an ExitError, and 1 for any
// other error.
func ExitStatus(err error) int {
	var le *LoadError
	var pe *PanicError
	var ee *ExitError
	switch {
	case err == nil:
		return 0
	case errors.As(err, &le):
		return 2
	case errors.As(err, &pe):
		return 1
	case errors.As(err, &ee):
		return ee.Status
	case errors.Is(err, ErrInterrupted):
		return 130
	default:
		return 1
	}
}
