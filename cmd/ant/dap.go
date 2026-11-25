package main

import (
	"context"
	"fmt"
	"io"

	"ant-lang.org/internal/dap"
)

// dapCommand is ant dap: it serves the Debug Adapter Protocol on
// standard input and standard output until the client disconnects, logging to standard error.
func dapCommand(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) != 0 {
		fmt.Fprintln(stderr, "usage: ant dap")
		return 2
	}
	return dap.Serve(context.Background(), stdin, stdout, stderr)
}
